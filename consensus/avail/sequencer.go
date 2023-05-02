package avail

import (
	"bytes"
	"crypto/ecdsa"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/0xPolygon/polygon-edge/blockchain"
	"github.com/0xPolygon/polygon-edge/consensus"
	"github.com/0xPolygon/polygon-edge/state"
	"github.com/0xPolygon/polygon-edge/txpool"
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	avail_types "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/hashicorp/go-hclog"
	"github.com/maticnetwork/avail-settlement/consensus/avail/validator"
	"github.com/maticnetwork/avail-settlement/consensus/avail/watchtower"
	"github.com/maticnetwork/avail-settlement/pkg/avail"
	"github.com/maticnetwork/avail-settlement/pkg/block"
	"github.com/maticnetwork/avail-settlement/pkg/snapshot"
	"github.com/maticnetwork/avail-settlement/pkg/staking"
)

const availBlockWindowLen = 7

type transitionInterface interface {
	Write(txn *types.Transaction) error
}

type SequencerWorker struct {
	logger                     hclog.Logger
	blockchain                 *blockchain.Blockchain
	executor                   *state.Executor
	txpool                     *txpool.TxPool
	snapshotter                snapshot.Snapshotter
	snapshotDistributor        snapshot.Distributor
	apq                        staking.ActiveParticipants
	availAppID                 avail_types.UCompact
	availClient                avail.Client
	availAccount               signature.KeyringPair
	nodeSignKey                *ecdsa.PrivateKey
	nodeAddr                   types.Address
	nodeType                   MechanismType
	stakingNode                staking.Node
	availSender                avail.Sender
	closeCh                    chan struct{}
	blockTime                  time.Duration // Minimum block generation time in seconds
	blockProductionIntervalSec uint64

	// availBlockNumWhenStaked is a used to fence the sequencing logic until
	// this node is staked and there is a start of a fresh new Avail block window.
	// Point type is used intentionally. `nil` means that this node has not staked
	// yet or it's not visible in the blockchain yet. The value gets set when the
	// staking is visible and it must not be modified afterwards.
	availBlockNumWhenStaked *int64
}

func (sw *SequencerWorker) Run(account accounts.Account, key *keystore.Key) error {
	t := new(atomic.Int64)

	// Return same seed value for the period of  `availWindowLen`.
	randomSeedFn := func() int64 {
		return t.Load() / availBlockWindowLen
	}
	activeSequencersQuerier := staking.NewRandomizedActiveSequencersQuerier(randomSeedFn, sw.apq)
	validator := validator.New(sw.blockchain, sw.nodeAddr, sw.logger)
	watchTower := watchtower.New(sw.blockchain, sw.executor, sw.txpool, sw.logger, types.Address(account.Address), key.PrivateKey)

	enableBlockProductionCh := make(chan bool)
	fraudResolver := NewFraudResolver(sw.logger, sw.blockchain, sw.executor, sw.txpool, watchTower, enableBlockProductionCh, sw.nodeAddr, sw.nodeSignKey, sw.availSender, sw.nodeType)

	callIdx, err := avail.FindCallIndex(sw.availClient)
	if err != nil {
		return fmt.Errorf("failed to discover avail call index: %s", err)
	}

	// Will wait until contract is updated and there's a staking transaction written
	sw.waitForStakedSequencer(activeSequencersQuerier, sw.nodeAddr)

	// Check if block production should be stopped due to inbound dispute resolution tx found in txpool.
	go fraudResolver.ShouldStopProducingBlocks(sw.apq)

	// Write blocks to the local blockchain and avail in intervals uless block production is stopped.
	go sw.runWriteBlocksLoop(enableBlockProductionCh, activeSequencersQuerier, fraudResolver, account, key)

	// BlockStream watcher must be started after the staking is done. Otherwise
	// the stream is out-of-sync.
	availBlockStream := sw.availClient.BlockStream(0)
	defer availBlockStream.Close()

	sw.logger.Info("Block stream successfully started.", "node_type", sw.nodeType)

	for {
		var blk *avail_types.SignedBlock

		select {
		case blk = <-availBlockStream.Chan():
			// Process below.

		// nolint:gosimple
		case _ = <-sw.snapshotDistributor.Receive():
			// TODO: Process the state snapshot.

		case <-sw.closeCh:
			if err := sw.stakingNode.UnStake(sw.nodeSignKey); err != nil {
				sw.logger.Error("failed to unstake the node", "error", err)
				return err
			}
			return nil
		}

		// Time `t` is [mostly] monotonic clock, backed by Avail. It's used for all
		// time sensitive logic in sequencer, such as block generation timeouts.
		t.Store(int64(blk.Block.Header.Number))

		// So this is the situation...
		// Here we are not looking for if current node should be producing or not producing the block.
		// What we are interested, prior to fraud resolver, if block is containing fraud check request.
		edgeBlks, err := block.FromAvail(blk, sw.availAppID, callIdx, sw.logger)
		if len(edgeBlks) == 0 && err != nil {
			sw.logger.Error("cannot extract Edge block from Avail block", "block_number", blk.Block.Header.Number, "error", err)
			// It is expected that not all Avail blocks contain a SL block. On any other error,
			// log the error and wait for a next one.
			if err != block.ErrNoExtrinsicFound {
				sw.logger.Warn("unexpected error while extracting SL blocks from Avail block", "error", err)
				continue
			}
		}

		// Write down blocks received from avail to make sure we're synced before processing with the
		// fraud check or writing down new blocks...
		for _, edgeBlk := range edgeBlks {
			// In case that dispute resolution is ended, please make sure to set fraud resolution block
			// to nil so whole chain and corrupted node can continue making our day good!
			// Block does not have to be written into the chain as it's already written with syncer...
			if fraudResolver.IsDisputeResolutionEnded(edgeBlk.Header) {
				sw.logger.Warn(
					"Dispute resolution for fraud block has ended! Chain can now continue with new block production...",
					"edge_block_hash", edgeBlk.Hash(),
					"fraud_block_hash", fraudResolver.GetBlock().Hash(),
				)
				fraudResolver.EndDisputeResolution()
			}

			// We cannot write the fraud proof block to the blockchain at all due to following reasons:
			// - Block number already exists and block won't be written.
			// - Watchtower has syncer disabled and when writing block, next block can come in rejecting this block.
			// - BeginDisputeTx that is inside of the fraud block was already shipped into txpool and it will
			//   trigger failures when writing down block due to already existing tx in the store.
			if !fraudResolver.IsFraudProofBlock(edgeBlk) {
				if err := validator.Check(edgeBlk); err == nil {
					if err := sw.blockchain.WriteBlock(edgeBlk, sw.nodeType.String()); err != nil {
						sw.logger.Warn(
							"failed to write edge block received from avail",
							"edge_block_hash", edgeBlk.Hash(),
							"error", err,
						)
					}
				} else {
					sw.logger.Warn(
						"failed to validate edge block received from avail",
						"edge_block_hash", edgeBlk.Hash(),
						"error", err,
					)
				}
			}
		}

		// Go through the blocks from avail and make sure to set fraud block in case it was discovered...
		fraudResolver.CheckAndSetFraudBlock(edgeBlks)

		// Periodically verify that we are staked, before proceeding with sequencer
		// logic. In the unexpected case of being slashed and dropping below the
		// required sequencer staking threshold, we must stop processing, because
		// otherwise we just get slashed more.
		sequencerStaked, sequencerError := activeSequencersQuerier.Contains(sw.nodeAddr)
		if sequencerError != nil {
			sw.logger.Error("failed to check if my account is among active staked sequencers; cannot continue", "error", sequencerError)
			continue
		}

		if !sequencerStaked {
			sw.logger.Warn("my account is not among active staked sequencers; cannot continue", "address", sw.nodeAddr.String())
			continue
		} else {
			if sw.availBlockNumWhenStaked == nil {
				sw.availBlockNumWhenStaked = new(int64)
				*sw.availBlockNumWhenStaked = t.Load()
				sw.logger.Debug("staking observed in the blockchain; storing avail block number", "block_number", blk.Block.Header.Number)
			}

			// Only proceed with the sequencing logic after the "join window" changes to
			// next one. This logic is needed because on a new node, that joins in the
			// middle of the Avail block window, the ActiveSequencer cache is not
			// consistent with the other nodes in the network. When all the nodes renew
			// their active sequencer list on a start of the new block window, they gain
			// coherent view into who is the next "leader" (i.e. the active sequencer
			// allowed to produce a block).
			if (*sw.availBlockNumWhenStaked / availBlockWindowLen) == (t.Load() / availBlockWindowLen) {
				sw.logger.Debug("sequencer account staked, but waiting for a fresh Avail block window after joining the network")
				continue
			} else {
				sw.logger.Debug("past the point of sequencer ramp up window", "block_number", blk.Block.Header.Number)
			}
		}

		// Will check the block for fraudlent behaviour and slash parties accordingly.
		// WARN: Continue will hard stop whole network from producing blocks until dispute is resolved.
		// If we do not stop the network from processing, blocks will continue to be built from other sequencers
		// resulting in block number missmatches and fraud will potentially be corrupted.
		if _, err := fraudResolver.CheckAndSlash(); err != nil {
			continue
		}

		availBlockNum := blk.Block.Header.Number
		// Check if this node is the current sequencer.
		if sw.IsNextSequencer(activeSequencersQuerier) {
			// When availBlockNum is 0, 1, 2 ... (availBlockWindowLen - 1), enable the block production.
			if availBlockNum%availBlockWindowLen < availBlockWindowLen-1 {
				sw.logger.Debug("it's my turn; enable block producing", "t", availBlockNum)
				enableBlockProductionCh <- true
			} else {
				// This is the last block of `availBlockWindowLen` -> stop block production to allow nodes to synchronize.
				sw.logger.Debug("it's my turn; last block on availBlockWindowLen. disabling block production", "t", availBlockNum)
				enableBlockProductionCh <- false
			}
		} else {
			// Under no circumstances, blocks should be produced when the node is not an active sequencer.
			sw.logger.Debug("it's not my turn; disable block producing", "t", availBlockNum)
			enableBlockProductionCh <- false
		}
	}
}

func (sw *SequencerWorker) IsNextSequencer(activeSequencersQuerier staking.ActiveSequencers) bool {
	sequencers, err := activeSequencersQuerier.Get()
	if err != nil {
		sw.logger.Error("querying staked sequencers failed; quitting", "error", err)
		return false
	}

	return bytes.Equal(sequencers[0].Bytes(), sw.nodeAddr.Bytes())
}

// runWriteBlocksLoop produces blocks at an interval defined in the blockProductionIntervalSec config option
func (sw *SequencerWorker) runWriteBlocksLoop(enableBlockProductionCh chan bool, activeSequencersQuerier staking.ActiveSequencers, fraudResolver *Fraud, myAccount accounts.Account, signKey *keystore.Key) {
	t := time.NewTicker(time.Duration(sw.blockProductionIntervalSec) * time.Second)
	defer t.Stop()

	shouldProduce := false

	for {
		select {
		case <-t.C:
			if !shouldProduce {
				continue
			}

			// Means we are processing the disputed (fraud) block verification and should not create new
			// blocks anywhere...
			if fraudResolver.IsChainDisabled() {
				continue
			}

			if !sw.IsNextSequencer(activeSequencersQuerier) {
				sw.logger.Warn(
					"it is not my turn to produce the block",
					"sequencer_addr", myAccount.Address,
				)
				continue
			} else {
				sw.logger.Warn(
					"it is my turn to produce the block",
					"sequencer_addr", myAccount.Address,
				)
			}

			sw.logger.Debug("writing a new block", "sequencer_addr", myAccount.Address)

			if err := sw.writeBlock(myAccount, signKey); err != nil {
				sw.logger.Error("failed to mine block", "error", err)
			}
		case e := <-enableBlockProductionCh:
			sw.logger.Debug("sequencer block producing status", "should_produce", e)
			shouldProduce = e

		case <-sw.closeCh:
			sw.logger.Debug("received stop signal")
			return
		}
	}
}

// writeNewBLock generates a new block based on transactions from the pool,
// and writes them to the blockchain
func (sw *SequencerWorker) writeBlock(myAccount accounts.Account, signKey *keystore.Key) error {
	parent := sw.blockchain.Header()

	header := &types.Header{
		ParentHash: parent.Hash,
		Number:     parent.Number + 1,
		Miner:      myAccount.Address.Bytes(),
		Nonce:      types.Nonce{},
		GasLimit:   parent.GasLimit, // Inherit from parent for now, will need to adjust dynamically later.
		Timestamp:  uint64(time.Now().Unix()),
	}

	// calculate gas limit based on parent header
	gasLimit, err := sw.blockchain.CalculateGasLimit(header.Number)
	if err != nil {
		return err
	}

	header.GasLimit = gasLimit

	// set the timestamp
	parentTime := time.Unix(int64(parent.Timestamp), 0)
	headerTime := parentTime.Add(sw.blockTime)

	if headerTime.Before(time.Now()) {
		headerTime = time.Now()
	}

	header.Timestamp = uint64(headerTime.Unix())

	// we need to include in the extra field the current set of validators
	err = block.AssignExtraValidators(header, ValidatorSet{types.StringToAddress(myAccount.Address.Hex())})
	if err != nil {
		return err
	}

	transition, err := sw.executor.BeginTxn(parent.StateRoot, header, types.StringToAddress(myAccount.Address.Hex()))
	if err != nil {
		return err
	}

	txns := sw.writeTransactions(gasLimit, transition)

	/* 	// TRIGGER SEQUENCER SLASHING
	   	maliciousBlockWritten := false
	   	if sw.nodeType != Sequencer && !maliciousBlockWritten {
	   		if header.Number == 4 || header.Number == 5 {
	   			tx, _ := staking.BeginDisputeResolutionTx(types.ZeroAddress, types.BytesToAddress(types.ZeroAddress.Bytes()), 1_000_000)
	   			tx.Nonce = 1
	   			txSigner := &crypto.FrontierSigner{}
	   			dtx, err := txSigner.SignTx(tx, sw.nodeSignKey)
	   			if err != nil {
	   				sw.logger.Error("failed to sign fraud transaction", "error", err)
	   				return err
	   			}

	   			txns = append(txns, dtx)
	   			maliciousBlockWritten = true
	   		}
	   	} */

	// Commit the changes
	_, root := transition.Commit()

	// Update the header
	header.StateRoot = root
	header.GasUsed = transition.TotalGas()

	// Build the actual block
	// The header hash is computed inside buildBlock
	blk := consensus.BuildBlock(consensus.BuildBlockParams{
		Header:   header,
		Txns:     txns,
		Receipts: transition.Receipts(),
	})

	// write the seal of the block after all the fields are completed
	header, err = block.WriteSeal(signKey.PrivateKey, blk.Header)
	if err != nil {
		return err
	}

	blk.Header = header

	// compute the hash, this is only a provisional hash since the final one
	// is sealed after all the committed seals
	blk.Header.ComputeHash()

	sw.logger.Info(
		"Sending new block to avail",
		"sequencer_node_addr", sw.nodeAddr,
		"block_number", blk.Number(),
		"block_hash", blk.Hash(),
		"block_parent_hash", blk.ParentHash(),
	)

	err = sw.availSender.SendAndWaitForStatus(blk, avail_types.ExtrinsicStatus{IsInBlock: true})
	if err != nil {
		sw.logger.Error("Error while submitting data to avail", "error", err)
		return err
	}

	sw.logger.Info(
		"Block successfully sent to avail. Writing block to local chain...",
		"sequencer_node_addr", sw.nodeAddr,
		"block_number", blk.Number(),
		"block_hash", blk.Hash(),
		"block_parent_hash", blk.ParentHash(),
	)

	// Write the block to the blockchain
	if err := sw.blockchain.WriteBlock(blk, sw.nodeType.String()); err != nil {
		return err
	}

	sw.logger.Info(
		"Successfully wrote new sequencer block to the local chain",
		"sequencer_node_addr", sw.nodeAddr,
		"block_number", blk.Number(),
		"block_hash", blk.Hash(),
		"block_parent_hash", blk.ParentHash(),
	)

	// After the block has been written we reset the txpool to remove stale transactions.
	sw.txpool.ResetWithHeaders(blk.Header)

	return nil
}

func (sw *SequencerWorker) writeTransactions(gasLimit uint64, transition transitionInterface) []*types.Transaction {
	var successful []*types.Transaction

	sw.txpool.Prepare()

	for {
		tx := sw.txpool.Peek()
		if tx == nil {
			break
		}

		if tx.ExceedsBlockGasLimit(gasLimit) {
			continue
		}

		if err := transition.Write(tx); err != nil {
			if _, ok := err.(*state.GasLimitReachedTransitionApplicationError); ok { // nolint:errorlint
				sw.logger.Warn("transaction reached gas limit during excution", "hash", tx.Hash.String())
				break
			} else if appErr, ok := err.(*state.TransitionApplicationError); ok && appErr.IsRecoverable { // nolint:errorlint
				sw.logger.Warn("transaction caused application error", "hash", tx.Hash.String())
				sw.txpool.Demote(tx)
			} else {
				sw.logger.Error("transaction caused unknown error", "error", err)
				sw.txpool.Drop(tx)
			}

			continue
		}

		// no errors, pop the tx from the pool
		sw.txpool.Pop(tx)

		successful = append(successful, tx)
	}

	return successful
}

func NewSequencer(
	logger hclog.Logger, b *blockchain.Blockchain, e *state.Executor, txp *txpool.TxPool,
	snapshotter snapshot.Snapshotter, snapshotDistributor snapshot.Distributor,
	availClient avail.Client, availAccount signature.KeyringPair, availAppID avail_types.UCompact,
	nodeSignKey *ecdsa.PrivateKey, nodeAddr types.Address, nodeType MechanismType,
	apq staking.ActiveParticipants, stakingNode staking.Node, availSender avail.Sender, closeCh <-chan struct{},
	blockTime time.Duration, blockProductionIntervalSec uint64,
) (*SequencerWorker, error) {
	return &SequencerWorker{
		logger:                     logger,
		blockchain:                 b,
		executor:                   e,
		txpool:                     txp,
		snapshotter:                snapshotter,
		snapshotDistributor:        snapshotDistributor,
		apq:                        apq,
		availAppID:                 availAppID,
		availClient:                availClient,
		availAccount:               availAccount,
		nodeSignKey:                nodeSignKey,
		nodeAddr:                   nodeAddr,
		nodeType:                   nodeType,
		stakingNode:                stakingNode,
		availSender:                availSender,
		blockTime:                  blockTime,
		blockProductionIntervalSec: blockProductionIntervalSec,
	}, nil
}
