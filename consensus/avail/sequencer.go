package avail

import (
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
)

type transitionInterface interface {
	Write(txn *types.Transaction) error
}

func (d *Avail) runSequencer(minerKeystore *keystore.KeyStore, miner accounts.Account, minerPK *keystore.Key) {
	d.logger.Info("sequencer started")

	for {
		// wait until there is a new txn
		select {
		case <-d.nextNotify():
		case <-d.closeCh:
			return
		}

		// For now it's here as is, not the best path moving forward for sure but the idea
		// is to check if sequencer is staked prior we allow any futhure block manipulations.
		sequencerStaked, sequencerError := d.isSequencerStaked(miner)
		if sequencerError != nil {
			d.logger.Error("failed to check if sequencer is staked", "err", sequencerError)
			continue
		}

		// TODO: Figure out how to do this check properly.
		// For now disabled until I figure out how to stake properly.
		if !sequencerStaked {
			d.logger.Error("Forbidding block generation until sequencer is staked properly...")
			continue
		}

		// There are new transactions in the pool, try to seal them
		header := d.blockchain.Header()
		if err := d.writeNewBlock(minerKeystore, miner, minerPK, header); err != nil {
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
func (d *Avail) writeNewBlock(minerKeystore *keystore.KeyStore, minerAccount accounts.Account, minerPK *keystore.Key, parent *types.Header) error {
	header := &types.Header{
		ParentHash: parent.Hash,
		Number:     parent.Number + 1,
		Miner:      minerAccount.Address.Bytes(),
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
	assignExtraValidators(header, ValidatorSet{types.StringToAddress(minerAccount.Address.Hex())})

	transition, err := d.executor.BeginTxn(parent.StateRoot, header, types.StringToAddress(minerAccount.Address.Hex()))
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

	// write the seal of the block after all the fields are completed
	header, err = writeSeal(minerPK.PrivateKey, block.Header)
	if err != nil {
		d.logger.Info("FAILING HERE? 5")
		return err
	}

	block.Header = header

	// compute the hash, this is only a provisional hash since the final one
	// is sealed after all the committed seals
	block.Header.ComputeHash()

	err, malicious := d.sendBlockToAvail(block)
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
	if err := d.blockchain.WriteBlock(block, "not-sure-what-source-yet-is"); err != nil {
		d.logger.Info("FAILING HERE? 7")
		return err
	}

	// after the block has been written we reset the txpool so that
	// the old transactions are removed
	d.txpool.ResetWithHeaders(block.Header)

	fmt.Printf("Written block information: %+v", block.Header)

	return nil
}

func (d *Avail) sendBlockToAvail(block *types.Block) (error, bool) {
	malicious := false
	sender := avail.NewSender(d.availClient, signature.TestKeyringPairAlice)

	// XXX: Test watch tower and validator. This breaks a block every now and then.
	if rand.Intn(3) == 2 {
		d.logger.Warn("XXX - I'm gonna break a block submitted to Avail")
		block.Header.StateRoot[0] = 42
		malicious = true
	}

	d.logger.Info("Submitting block to avail...")
	f := sender.SubmitDataAndWaitForStatus(block.MarshalRLP(), stypes.ExtrinsicStatus{IsInBlock: true})
	if _, err := f.Result(); err != nil {
		d.logger.Error("Error while submitting data to avail", err)
		return err, malicious
	}

	d.logger.Info("Submitted block to avail", "block", block.Header.Number)
	return nil, malicious
}
