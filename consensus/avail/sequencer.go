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
	availBlockStream := avail.NewBlockStream(d.availClient, d.logger, avail.BridgeAppID, 0)
	defer availBlockStream.Close()

	d.logger.Debug("ensuring sequencer staked")
	err := d.ensureStaked(activeParticipantsQuerier)
	if err != nil {
		d.logger.Error("error while ensuring sequencer staked", "error", err)
		return
	}

	d.logger.Debug("ensured sequencer staked")
	d.logger.Debug("sequencer started")

	for blk := range availBlockStream.Chan() {
		// Check if we need to stop.
		select {
		case <-d.closeCh:
			if err := stakingNode.UnStake(signKey.PrivateKey); err != nil {
				d.logger.Error("failed to unstake the node", "error", err)
			}
			return
		default:
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

		// Is it my turn to generate next block?
		if bytes.Equal(sequencers[0].Bytes(), myAccount.Address.Bytes()) {
			header := d.blockchain.Header()
			d.logger.Debug("it's my turn; producing a block", "t", blk.Block.Header.Number)
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

		if tx.ExceedsBlockGasLimit(gasLimit) {
			d.txpool.Drop(tx)
			continue
		}

		if err := transition.Write(tx); err != nil {
			if _, ok := err.(*state.GasLimitReachedTransitionApplicationError); ok { // nolint:errorlint
				break
			} else if appErr, ok := err.(*state.TransitionApplicationError); ok && appErr.IsRecoverable { // nolint:errorlint
				d.txpool.Demote(tx)
			} else {
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

	blk.Header = header

	// compute the hash, this is only a provisional hash since the final one
	// is sealed after all the committed seals
	blk.Header.ComputeHash()

	d.logger.Debug("sending block to avail")

	err = d.sender.SendAndWaitForStatus(blk, stypes.ExtrinsicStatus{IsInBlock: true})
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
