package avail

import (
	"bytes"
	"fmt"
	"math/rand"
	"time"

	"github.com/0xPolygon/polygon-edge/consensus"
	"github.com/0xPolygon/polygon-edge/state"
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
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

func (d *Avail) runSequencer(myAccount accounts.Account, signKey *keystore.Key) {
	activeSequencersQuerier := staking.NewActiveParticipantsQuerier(d.blockchain, d.executor, d.logger)
	availBlockStream := avail.NewBlockStream(d.availClient, d.logger, avail.BridgeAppID, 0)
	defer availBlockStream.Close()

	d.logger.Info("sequencer started")

	for blk := range availBlockStream.Chan() {
		// Check if we need to stop.
		select {
		case <-d.closeCh:
			return
		default:
		}

		// Periodically verify that we are staked, before proceeding with sequencer
		// logic. In the unexpected case of being slashed and dropping below the
		// required sequencer staking threshold, we must stop processing, because
		// otherwise we just get slashed more.
		sequencerStaked, sequencerError := activeSequencersQuerier.Contains(types.Address(myAccount.Address), staking.Sequencer)
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
		t := blk.Block.Header.Number

		d.logger.Debug("sequencer time", "t", t)

		sequencers, err := activeSequencersQuerier.Get(staking.Sequencer)
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
			if err := d.writeNewBlock(myAccount, signKey, header); err != nil {
				d.logger.Error("failed to mine block", "err", err)
			}

			continue
		} else {
			// XXX: This is just for debugging.
			var activeSequencers []string
			for _, s := range sequencers {
				activeSequencers = append(activeSequencers, s.String())
			}
			d.logger.Debug("it's not my turn to produce a block", "myAccount.Address", myAccount.Address.String(), "activeSequencers", activeSequencers)
		}

		// TODO: What if node fails to publish a block on its turn?
		// TODO: What if a node fails to get information about published block?
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

	d.logger.Info("picked out txns from pool", "num", len(successful), "remaining", d.txpool.Length())

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
		d.logger.Info("FAILING HERE? 1")
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
		d.logger.Info("FAILING HERE? 3")
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
		d.logger.Info("FAILING HERE? 5")
		return err
	}

	blk.Header = header

	// compute the hash, this is only a provisional hash since the final one
	// is sealed after all the committed seals
	blk.Header.ComputeHash()

	err, malicious := d.sendBlockToAvail(blk)
	if err != nil {
		d.logger.Info("FAILING HERE? 6")
		return err
	}

	if malicious {
		// Don't write malicious block into blockchain. It messes up the parent
		// state of blocks when validator nor watch tower has it.
		return nil
	}

	// Write the block to the blockchain
	if err := d.blockchain.WriteBlock(blk, "not-sure-what-source-yet-is"); err != nil {
		d.logger.Info("FAILING HERE? 7")
		return err
	}

	// after the block has been written we reset the txpool so that
	// the old transactions are removed
	d.txpool.ResetWithHeaders(blk.Header)

	fmt.Printf("Written block information: %+v", blk.Header)

	return nil
}

func (d *Avail) sendBlockToAvail(blk *types.Block) (error, bool) {
	malicious := false
	sender := avail.NewSender(d.availClient, signature.TestKeyringPairAlice)

	// XXX: Test watch tower and validator. This breaks a block every now and then.
	if rand.Intn(3) == 2 {
		d.logger.Warn("XXX - I'm gonna break a block submitted to Avail")
		blk.Header.StateRoot[0] = 42
		malicious = true
	}

	d.logger.Info("Submitting block to avail...")
	f := sender.SubmitDataAndWaitForStatus(blk.MarshalRLP(), stypes.ExtrinsicStatus{IsInBlock: true})
	if _, err := f.Result(); err != nil {
		d.logger.Error("Error while submitting data to avail", err)
		return err, malicious
	}

	d.logger.Info("Submitted block to avail", "block", blk.Header.Number)
	return nil, malicious
}
