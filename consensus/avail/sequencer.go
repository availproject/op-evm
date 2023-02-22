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
	stypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/hashicorp/go-hclog"
	"github.com/maticnetwork/avail-settlement/consensus/avail/validator"
	"github.com/maticnetwork/avail-settlement/pkg/avail"
	"github.com/maticnetwork/avail-settlement/pkg/block"
	"github.com/maticnetwork/avail-settlement/pkg/staking"
)

type transitionInterface interface {
	Write(txn *types.Transaction) error
}

type SequencerWorker struct {
	logger                     hclog.Logger
	blockchain                 *blockchain.Blockchain
	executor                   *state.Executor
	validator                  validator.Validator
	txpool                     *txpool.TxPool
	apq                        staking.ActiveParticipants
	availAppID                 avail_types.U32
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
}

func (sw *SequencerWorker) IsNextSequencer(sequencer types.Address) bool {
	return bytes.Equal(sequencer.Bytes(), sw.nodeAddr.Bytes())
}

func (sw *SequencerWorker) Run(account accounts.Account, key *keystore.Key) error {
	t := new(atomic.Int64)
	activeSequencersQuerier := staking.NewRandomizedActiveSequencersQuerier(t.Load, sw.apq)
	validator := validator.New(sw.blockchain, sw.executor, sw.nodeAddr, sw.logger)
	fraudResolver := NewFraudResolver(sw.logger, sw.blockchain, sw.executor, sw.txpool, validator, sw.nodeAddr, sw.nodeSignKey, sw.availSender, sw.nodeType)

	accBalance, err := avail.GetBalance(sw.availClient, sw.availAccount)
	if err != nil {
		return fmt.Errorf("failed to discover account balance: %s", err)
	}

	sw.logger.Info("Current avail account", "balance", accBalance.Int64())

	callIdx, err := avail.FindCallIndex(sw.availClient)
	if err != nil {
		return fmt.Errorf("failed to discover avail call index: %s", err)
	}

	// Will wait until contract is updated and there's a staking transaction written
	sw.waitForStakedSequencer(activeSequencersQuerier, sw.nodeAddr)

	enableBlockProductionCh := make(chan bool)
	go sw.runWriteBlocksLoop(enableBlockProductionCh, sw.apq, account, key)

	// BlockStream watcher must be started after the staking is done. Otherwise
	// the stream is out-of-sync.
	availBlockStream := avail.NewBlockStream(sw.availClient, sw.logger, 0)
	defer availBlockStream.Close()

	sw.logger.Info("Block stream successfully started.", "node_type", sw.nodeType)

	for {
		select {
		case blk := <-availBlockStream.Chan():
			// Time `t` is [mostly] monotonic clock, backed by Avail. It's used for all
			// time sensitive logic in sequencer, such as block generation timeouts.
			t.Store(int64(blk.Block.Header.Number))

			// So this is the situation...
			// Here we are not looking for if current node should be producing or not producing the block.
			// What we are interested, prior to fraud resolver, if block is containing fraud check request.
			edgeBlks, err := block.FromAvail(blk, sw.availAppID, callIdx, sw.logger)
			if len(edgeBlks) == 0 && err != nil {
				sw.logger.Error("cannot extract Edge block from Avail block", "block_number", blk.Block.Header.Number, "error", err)
				// It is expected that each edge block should contain avail block, however,
				// this can block the entire production of the blocks later on.
				// Continue if the decompiling block resulted in any error other than not found.
				if err != block.ErrNoExtrinsicFound {
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
					fraudResolver.SetBlock(nil)
				}

				// We cannot write down disputed blocks to the blockchain as they would be rejected due to
				// numerous reasons. From block number missalignment to block already exist to skewing the
				// rest of the flow later on...
				if !fraudResolver.IsFraudProofBlock(edgeBlk) {
					if err := validator.Check(edgeBlk); err == nil {
						if err := sw.blockchain.WriteBlock(edgeBlk, sw.nodeType.String()); err != nil {
							sw.logger.Warn(
								"failed to write edge block received from avail",
								"edge_block_hash", edgeBlk.Hash(),
								"error", err,
							)
						} else {
							sw.txpool.ResetWithHeaders(edgeBlk.Header)
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

			// Periodically verify that we are staked, before proceeding with sequencer
			// logic. In the unexpected case of being slashed and dropping below the
			// required sequencer staking threshold, we must stop processing, because
			// otherwise we just get slashed more.
			sequencerStaked, sequencerError := activeSequencersQuerier.Contains(sw.nodeAddr)
			if sequencerError != nil {
				sw.logger.Error("failed to check if my account is among active staked sequencers; cannot continue", "err", sequencerError)
				continue
			}

			if !sequencerStaked {
				sw.logger.Error("my account is not among active staked sequencers; cannot continue", "address", sw.nodeAddr.String())
				continue
			}

			sequencers, err := activeSequencersQuerier.Get()
			if err != nil {
				sw.logger.Error("querying staked sequencers failed; quitting", "error", err)
				continue
			}

			if len(sequencers) == 0 {
				// This is something that should **never** happen.
				panic("no staked sequencers")
			}

			// Go through the blocks from avail and make sure to set fraud block in case it was discovered...
			fraudResolver.CheckAndSetFraudBlock(edgeBlks)

			isItMyTurn := sw.IsNextSequencer(sequencers[0])

			// Will check the block for fraudlent behaviour and slash parties accordingly.
			// WARN: Continue will hard stop whole network from producing blocks until dispute is resolved.
			// We can change this in the future but with it, understand that series of issues are going to
			// happen with syncing and publishing of the blocks that will need to be fixed.
			if _, err := fraudResolver.CheckAndSlash(false); err != nil {
				enableBlockProductionCh <- false
				continue
			}

			// Is it my turn to generate next block?
			if isItMyTurn {
				sw.logger.Debug("it's my turn; enable block producing", "t", blk.Block.Header.Number)
				enableBlockProductionCh <- true
				continue
			} else {
				sw.logger.Debug("it's not my turn; disable block producing", "t", blk.Block.Header.Number)
				enableBlockProductionCh <- false
			}

		case <-sw.closeCh:
			if err := sw.stakingNode.UnStake(sw.nodeSignKey); err != nil {
				sw.logger.Error("failed to unstake the node", "error", err)
				return err
			}
			return nil
		}
	}
}

// runWriteBlocksLoop produces blocks at an interval defined in the blockProductionIntervalSec config option
func (sw *SequencerWorker) runWriteBlocksLoop(enableBlockProductionCh chan bool, activeParticipantsQuerier staking.ActiveParticipants, myAccount accounts.Account, signKey *keystore.Key) {
	t := time.NewTicker(time.Duration(sw.blockProductionIntervalSec) * time.Second)
	defer t.Stop()

	shouldProduce := false
	for {
		select {
		case <-t.C:
			if !shouldProduce {
				continue
			}

			sw.logger.Debug("writing a new block", "myAccount.Address", myAccount.Address)
			if err := sw.writeBlock(enableBlockProductionCh, activeParticipantsQuerier, myAccount, signKey); err != nil {
				sw.logger.Error("failed to mine block", "err", err)
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
func (sw *SequencerWorker) writeBlock(enableBlockProductionCh chan bool, activeParticipantsQuerier staking.ActiveParticipants, myAccount accounts.Account, signKey *keystore.Key) error {
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

	txns := sw.writeTransactions(enableBlockProductionCh, activeParticipantsQuerier, gasLimit, transition)

	/* 	TRIGGER SEQUENCER SLASHING
		maliciousBlockWritten := false
	   	if sw.nodeType != Sequencer && !maliciousBlockWritten {
	   		if header.Number == 4 || header.Number == 5 {
	   			tx, _ := staking.BeginDisputeResolutionTx(types.ZeroAddress, types.BytesToAddress(types.ZeroAddress.Bytes()), 1_000_000)
	   			tx.Nonce = 1
	   			txSigner := &crypto.FrontierSigner{}
	   			dtx, err := txSigner.SignTx(tx, sw.nodeSignKey)
	   			if err != nil {
	   				sw.logger.Error("failed to sign fraud transaction", "err", err)
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

	err = sw.availSender.SendAndWaitForStatus(blk, stypes.ExtrinsicStatus{IsInBlock: true})
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
	// after the block has been written we reset the txpool so that
	// the old transactions are removed
	sw.txpool.ResetWithHeaders(blk.Header)

	return nil
}

func (sw *SequencerWorker) writeTransactions(enableBlockProductionCh chan bool, activeParticipantsQuerier staking.ActiveParticipants, gasLimit uint64, transition transitionInterface) []*types.Transaction {
	var successful []*types.Transaction

	sw.txpool.Prepare()

	for {
		tx := sw.txpool.Peek()
		if tx == nil {
			break
		}

		// We are not interested in errors at this moment. Only if watchtower is set to true or false
		isWatchtower, _ := activeParticipantsQuerier.Contains(tx.From, staking.WatchTower)

		sw.logger.Info(
			"New tx pool transaction discovered",
			"hash", tx.Hash,
			"value", tx.Value,
			"originating_addr", tx.From.String(),
			"recipient_addr", tx.To.String(),
			"submitted_via_watchtower", isWatchtower,
			"staking_contract_addr", staking.AddrStakingContract.String(),
			"submitted_towards_contract", bytes.Equal(tx.To.Bytes(), staking.AddrStakingContract.Bytes()),
		)

		// TODO: Figure out if the transaction is type of begin dispute resolution
		if isWatchtower && bytes.Equal(tx.To.Bytes(), staking.AddrStakingContract.Bytes()) {
			sw.logger.Warn(
				"Discovered begin dispute resolution tx while building block txns. Processing block txns until this tx and disabling node...",
				"originating_watchtower_addr", tx.From,
				"dispute_resolution_tx_hash", tx.Hash,
			)
			enableBlockProductionCh <- false
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
	logger hclog.Logger, b *blockchain.Blockchain, e *state.Executor, txp *txpool.TxPool, v validator.Validator, availClient avail.Client,
	availAccount signature.KeyringPair, availAppID avail_types.U32,
	nodeSignKey *ecdsa.PrivateKey, nodeAddr types.Address, nodeType MechanismType,
	apq staking.ActiveParticipants, stakingNode staking.Node, availSender avail.Sender, closeCh <-chan struct{},
	blockTime time.Duration, blockProductionIntervalSec uint64) (*SequencerWorker, error) {
	return &SequencerWorker{
		logger:                     logger,
		blockchain:                 b,
		executor:                   e,
		validator:                  v,
		txpool:                     txp,
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
