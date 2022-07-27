package avail

import (
	"fmt"
	"time"

	"github.com/0xPolygon/polygon-edge/consensus"
	"github.com/0xPolygon/polygon-edge/state"
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	stypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/maticnetwork/avail-settlement/pkg/avail"
)

type transitionInterface interface {
	Write(txn *types.Transaction) error
}

func (d *Avail) runSequencer() {
	d.logger.Info("sequencer started")

	for {
		// wait until there is a new txn
		select {
		case <-d.nextNotify():
		case <-d.closeCh:
			return
		}

		// There are new transactions in the pool, try to seal them
		header := d.blockchain.Header()
		if err := d.writeNewBlock(header); err != nil {
			d.logger.Error("failed to mine block", "err", err)
		}
	}
}

func (d *Avail) nextNotify() chan struct{} {
	if d.interval == 0 {
		d.interval = 2
	}

	go func() {
		<-time.After(time.Duration(d.interval) * time.Second)
		d.notifyCh <- struct{}{}
	}()

	return d.notifyCh
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
func (d *Avail) writeNewBlock(parent *types.Header) error {
	header := &types.Header{
		ParentHash: parent.Hash,
		Number:     parent.Number + 1,
		Miner:      types.Address{},
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

	miner, err := d.GetBlockCreator(header)
	if err != nil {
		d.logger.Info("FAILING HERE? 2")
		return err
	}

	transition, err := d.executor.BeginTxn(parent.StateRoot, header, miner)
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
	block := consensus.BuildBlock(consensus.BuildBlockParams{
		Header:   header,
		Txns:     txns,
		Receipts: transition.Receipts(),
	})

	if err := d.blockchain.VerifyFinalizedBlock(block); err != nil {
		d.logger.Info("FAILING HERE? 4")
		return err
	}

	// Write block to the avail
	if err := d.sendBlockToAvail(block); err != nil {
		d.logger.Info("FAILING HERE? 5")
		return err
	}

	// Write the block to the blockchain
	if err := d.blockchain.WriteBlock(block); err != nil {
		d.logger.Info("FAILING HERE? 6")
		return err
	}

	// after the block has been written we reset the txpool so that
	// the old transactions are removed
	d.txpool.ResetWithHeaders(block.Header)

	fmt.Printf("Written block information: %+v", block.Header)

	return nil
}

func (d *Avail) sendBlockToAvail(block *types.Block) error {
	sender := avail.NewSender(d.availClient, signature.TestKeyringPairAlice)
	d.logger.Info("Submitting block to avail...")
	f := sender.SubmitDataAndWaitForStatus(block.MarshalRLP(), stypes.ExtrinsicStatus{IsInBlock: true})
	if _, err := f.Result(); err != nil {
		d.logger.Error("Error while submitting data to avail", err)
		return err
	}
	d.logger.Info("Submitted block to avail", "block", block.Header.Number)
	return nil
}
