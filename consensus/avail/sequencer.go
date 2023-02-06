package avail

import (
	"bytes"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/0xPolygon/polygon-edge/consensus"
	"github.com/0xPolygon/polygon-edge/state"
	"github.com/0xPolygon/polygon-edge/types"
	stypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/maticnetwork/avail-settlement/consensus/avail/validator"
	"github.com/maticnetwork/avail-settlement/pkg/avail"
	"github.com/maticnetwork/avail-settlement/pkg/block"
	"github.com/maticnetwork/avail-settlement/pkg/staking"
)

type transitionInterface interface {
	Write(txn *types.Transaction) error
}

func (d *Avail) runSequencer(stakingNode staking.Node, myAccount accounts.Account, signKey *keystore.Key) {
	t := new(atomic.Int64)
	activeParticipantsQuerier := staking.NewActiveParticipantsQuerier(d.blockchain, d.executor, d.logger)
	activeSequencersQuerier := staking.NewRandomizedActiveSequencersQuerier(t.Load, activeParticipantsQuerier)
	validator := validator.New(d.blockchain, d.executor, types.StringToAddress(d.availAccount.Address))

	accBalance, err := avail.GetBalance(d.availClient, d.availAccount)
	if err != nil {
		panic(fmt.Sprintf("Balance failure: %s", err))
	}

	d.logger.Info("Current avail account", "balance", accBalance.Int64())

	callIdx, err := avail.FindCallIndex(d.availClient)
	if err != nil {
		panic(err)
	}

	d.logger.Debug("ensuring sequencer staked")
	err = d.ensureStaked(activeParticipantsQuerier)
	if err != nil {
		d.logger.Error("error while ensuring sequencer staked", "error", err)
		return
	}

	d.logger.Debug("sequencer staked")

	// BlockStream watcher must be started after the staking is done. Otherwise
	// the stream is out-of-sync.
	availBlockStream := avail.NewBlockStream(d.availClient, d.logger, 0)
	defer availBlockStream.Close()

	d.logger.Debug("sequencer started")

	var watchtowerFraudBlock *types.Block

	for blk := range availBlockStream.Chan() {
		d.logger.Info("-----------------------------------------------------------------------------------------------------------------------------------------------")

		// Check if we need to stop.
		select {
		case <-d.closeCh:
			if err := stakingNode.UnStake(signKey.PrivateKey); err != nil {
				d.logger.Error("failed to unstake the node", "error", err)
			}
			return
		default:
		}

		if watchtowerFraudBlock != nil {
			d.logger.Info(
				"Watchtower fraud block discovered, processing with the check...",
				"watchtower_block_hash", watchtowerFraudBlock.Header.Hash.String(),
			)

			if sequencerFraudBlockHash, exists := block.GetExtraDataFraudProofTarget(watchtowerFraudBlock.Header); exists {
				if maliciousBlock, ok := d.blockchain.GetBlockByHash(sequencerFraudBlockHash, true); ok {
					d.logger.Info(
						"Potentially malicious block discovered, processing with the check...",
						"watchtower_block_hash", watchtowerFraudBlock.Header.Hash.String(),
						"malicious_block_hash", sequencerFraudBlockHash.String(),
					)

					sequencerAddr := types.BytesToAddress(maliciousBlock.Header.Miner)
					watchtowerAddr := types.BytesToAddress(watchtowerFraudBlock.Header.Miner)

					d.logger.Info(
						"Verifying if sequencer has permissions to slash the sequencer",
						"current_sequencer_addr", d.minerAddr.String(),
						"malicious_sequencer_addr", sequencerAddr.String(),
					)

					// Case we have probation sequencer and sequencer is not the same a the current miner addr
					//if sequencerAddr.String() != d.minerAddr.String() {
					if err := validator.Check(maliciousBlock); err != nil {
						d.logger.Warn(
							"Fraud proof block check confirmed malicious block, slashing sequencer",
							"watchtower_block_hash", watchtowerFraudBlock.Header.Hash.String(),
							"malicious_block_hash", sequencerFraudBlockHash.String(),
							"sequencer", sequencerAddr,
							"watchtower_addr", watchtowerAddr,
						)
						// Slash sequencer
						// TODO: Figure out what to do when slashing node continuously fails.
						if err := d.slashNode(sequencerAddr, watchtowerFraudBlock.Header); err != nil {
							continue
						}

						d.logger.Warn(
							"Malicious sequencer successfully slashed!",
							"watchtower_block_hash", watchtowerFraudBlock.Header.Hash.String(),
							"malicious_block_hash", sequencerFraudBlockHash.String(),
							"sequencer", sequencerAddr,
							"watchtower_addr", watchtowerAddr,
						)
					} else {
						// Slash watchtower
						d.logger.Warn(
							"Fraud proof block check confirmed correct. Slashing watchtower",
							"watchtower_block_hash", watchtowerFraudBlock.Header.Hash.String(),
							"malicious_block_hash", sequencerFraudBlockHash.String(),
							"sequencer", sequencerAddr,
							"watchtower_addr", watchtowerAddr,
						)
					}

					watchtowerFraudBlock = nil

					//} else {
					//	d.logger.Error("rejecting to process slashing on the malicious node", "address", sequencerAddr)
					//	watchtowerFraudBlock = nil
					//}
				}
			}

		}

		// Periodically verify that we are staked, before proceeding with sequencer
		// logic. In the unexpected case of being slashed and dropping below the
		// required sequencer staking threshold, we must stop processing, because
		// otherwise we just get slashed more.
		sequencerStaked, sequencerError := activeSequencersQuerier.Contains(types.Address(myAccount.Address))
		if sequencerError != nil {
			d.logger.Error("failed to check if my account is among active staked sequencers; cannot continue", "err", sequencerError)
			return
		}

		if !sequencerStaked {
			d.logger.Error("my account is not among active staked sequencers; cannot continue", "myAccount.Address", myAccount.Address)
			return
		}

		// Time `t` is [mostly] monotonic clock, backed by Avail. It's used for all
		// time sensitive logic in sequencer, such as block generation timeouts.
		t.Store(int64(blk.Block.Header.Number))

		sequencers, err := activeSequencersQuerier.Get()
		if err != nil {
			d.logger.Error("querying staked sequencers failed; quitting", "error", err)
			return
		}

		if len(sequencers) == 0 {
			// This is something that should **never** happen.
			panic("no staked sequencers")
		}

		edgeBlks, err := block.FromAvail(blk, d.availAppID, callIdx)
		if len(edgeBlks) == 0 && err != nil {
			d.logger.Error("cannot extract Edge block from Avail block", "block_number", blk.Block.Header.Number, "error", err)
			continue
		}

		// Is it my turn to generate next block?
		if bytes.Equal(sequencers[0].Bytes(), myAccount.Address.Bytes()) {
			header := d.blockchain.Header()
			d.logger.Info("it's my turn; producing a block", "t", blk.Block.Header.Number)

			for _, edgeBlk := range edgeBlks {
				d.logger.Info("Received sequencer block", "number", edgeBlk.Header.Number, "hash", edgeBlk.Hash(), "txn", len(edgeBlk.Transactions))

				if fraudProofBlockHash, exists := block.GetExtraDataFraudProofTarget(edgeBlk.Header); exists {
					d.logger.Info("Fraud proof discovered for block", "hash", fraudProofBlockHash)
					watchtowerFraudBlock = edgeBlk
				}
			}

			if err := d.writeNewBlock(myAccount, signKey, header); err != nil {
				d.logger.Error("failed to mine block", "err", err)
			}

			continue
		} else {
			d.logger.Debug("it's not my turn; skippin' a round", "t", blk.Block.Header.Number)
		}
	}
}

func (d *Avail) startSyncing() {
	// Start the syncer
	err := d.syncer.Start()
	if err != nil {
		panic(fmt.Sprintf("starting blockchain sync failed: %s", err))
	}

	syncFunc := func(blk *types.Block) bool {
		d.txpool.ResetWithHeaders(blk.Header)
		return false
	}

	err = d.syncer.Sync(syncFunc)
	if err != nil {
		panic(fmt.Sprintf("syncing blockchain failed: %s", err))
	}
}

func (d *Avail) writeTransactions(gasLimit uint64, transition transitionInterface) []*types.Transaction {
	var successful []*types.Transaction

	d.txpool.Prepare()

	for {
		tx := d.txpool.Peek()
		if tx == nil {
			break
		}

		d.logger.Debug("found transaction from txpool", "hash", tx.Hash.String())

		if tx.ExceedsBlockGasLimit(gasLimit) {
			d.txpool.Drop(tx)
			continue
		}

		if err := transition.Write(tx); err != nil {
			if _, ok := err.(*state.GasLimitReachedTransitionApplicationError); ok { // nolint:errorlint
				d.logger.Warn("transaction reached gas limit during excution", "hash", tx.Hash.String())
				break
			} else if appErr, ok := err.(*state.TransitionApplicationError); ok && appErr.IsRecoverable { // nolint:errorlint
				d.logger.Warn("transaction caused application error", "hash", tx.Hash.String())
				d.txpool.Demote(tx)
			} else {
				d.logger.Error("transaction caused unknown error", "error", err)
				d.txpool.Drop(tx)
			}

			continue
		}

		// no errors, pop the tx from the pool
		d.txpool.Pop(tx)

		successful = append(successful, tx)
	}

	return successful
}

// writeNewBLock generates a new block based on transactions from the pool,
// and writes them to the blockchain
func (d *Avail) writeNewBlock(myAccount accounts.Account, signKey *keystore.Key, parent *types.Header) error {
	header := &types.Header{
		ParentHash: parent.Hash,
		Number:     parent.Number + 1,
		Miner:      myAccount.Address.Bytes(),
		Nonce:      types.Nonce{},
		GasLimit:   parent.GasLimit, // Inherit from parent for now, will need to adjust dynamically later.
		Timestamp:  uint64(time.Now().Unix()),
	}

	// calculate gas limit based on parent header
	gasLimit, err := d.blockchain.CalculateGasLimit(header.Number)
	if err != nil {
		return err
	}

	header.GasLimit = gasLimit

	// set the timestamp
	parentTime := time.Unix(int64(parent.Timestamp), 0)
	headerTime := parentTime.Add(d.blockTime)

	if headerTime.Before(time.Now()) {
		headerTime = time.Now()
	}

	header.Timestamp = uint64(headerTime.Unix())

	// we need to include in the extra field the current set of validators
	err = block.AssignExtraValidators(header, ValidatorSet{types.StringToAddress(myAccount.Address.Hex())})
	if err != nil {
		return err
	}

	transition, err := d.executor.BeginTxn(parent.StateRoot, header, types.StringToAddress(myAccount.Address.Hex()))
	if err != nil {
		return err
	}

	txns := d.writeTransactions(gasLimit, transition)

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

	if header.Number == 5 {
		header.ExtraData = []byte{1, 2, 3}
	}

	// Corrupt miner -> fraud check.
	//header.Miner = types.ZeroAddress.Bytes()

	blk.Header = header

	// compute the hash, this is only a provisional hash since the final one
	// is sealed after all the committed seals
	blk.Header.ComputeHash()

	d.logger.Info("sending block to avail")

	err = d.availSender.SendAndWaitForStatus(blk, stypes.ExtrinsicStatus{IsInBlock: true})
	if err != nil {
		d.logger.Error("Error while submitting data to avail", "error", err)
		return err
	}

	d.logger.Info("sent block to avail")
	d.logger.Info("writing block to blockchain")

	// Write the block to the blockchain
	if err := d.blockchain.WriteBlock(blk, "sequencer"); err != nil {
		return err
	}

	// after the block has been written we reset the txpool so that
	// the old transactions are removed
	d.txpool.ResetWithHeaders(blk.Header)

	return nil
}

// writeNewBLock generates a new block based on transactions from the pool,
// and writes them to the blockchain
func (d *Avail) writeNewAvailBlock(myAccount accounts.Account, signKey *keystore.Key, parent *types.Header, txs []*types.Transaction) error {
	header := &types.Header{
		ParentHash: parent.Hash,
		Number:     parent.Number + 1,
		Miner:      myAccount.Address.Bytes(),
		Nonce:      types.Nonce{},
		GasLimit:   parent.GasLimit, // Inherit from parent for now, will need to adjust dynamically later.
		Timestamp:  uint64(time.Now().Unix()),
	}

	// calculate gas limit based on parent header
	gasLimit, err := d.blockchain.CalculateGasLimit(header.Number)
	if err != nil {
		return err
	}

	header.GasLimit = gasLimit

	// set the timestamp
	parentTime := time.Unix(int64(parent.Timestamp), 0)
	headerTime := parentTime.Add(d.blockTime)

	if headerTime.Before(time.Now()) {
		headerTime = time.Now()
	}

	header.Timestamp = uint64(headerTime.Unix())

	// we need to include in the extra field the current set of validators
	err = block.AssignExtraValidators(header, ValidatorSet{types.StringToAddress(myAccount.Address.Hex())})
	if err != nil {
		return err
	}

	transition, err := d.executor.BeginTxn(parent.StateRoot, header, types.StringToAddress(myAccount.Address.Hex()))
	if err != nil {
		return err
	}

	txns := []*types.Transaction{}

	for _, txn := range txs {
		if err := transition.Write(txn); err != nil {
			return err
		}
		txns = append(txns, txn)
	}

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

	if header.Number == 4 {
		header.ExtraData = []byte{1, 2, 3}
	}

	// Corrupt miner -> fraud check.
	//header.Miner = types.ZeroAddress.Bytes()

	blk.Header = header

	// compute the hash, this is only a provisional hash since the final one
	// is sealed after all the committed seals
	blk.Header.ComputeHash()

	d.logger.Debug("sending block to avail")

	err = d.availSender.SendAndWaitForStatus(blk, stypes.ExtrinsicStatus{IsInBlock: true})
	if err != nil {
		d.logger.Error("Error while submitting data to avail", "error", err)
		return err
	}

	d.logger.Debug("sent block to avail")
	d.logger.Debug("writing block to blockchain")

	// Write the block to the blockchain
	if err := d.blockchain.WriteBlock(blk, "sequencer"); err != nil {
		return err
	}

	// after the block has been written we reset the txpool so that
	// the old transactions are removed
	d.txpool.ResetWithHeaders(blk.Header)

	return nil
}
