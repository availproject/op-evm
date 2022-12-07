package avail

import (
	"fmt"
	"math/big"
	"time"

	"github.com/0xPolygon/polygon-edge/crypto"
	"github.com/maticnetwork/avail-settlement/pkg/block"
	"github.com/maticnetwork/avail-settlement/pkg/staking"
)

func (d *Avail) ensureStaked(activeParticipantsQuerier staking.ActiveParticipants) error {
	var nodeType staking.NodeType
	switch d.nodeType {
	case "bootstrap-sequencer", "sequencer":
		nodeType = staking.Sequencer
	case "watchtower":
		nodeType = staking.WatchTower
	case "validator":
		nodeType = staking.Validator
	default:
		return fmt.Errorf("unknown node type: %q", d.nodeType)
	}

	staked, err := activeParticipantsQuerier.Contains(d.minerAddr, nodeType)
	if err != nil {
		return err
	}

	if staked {
		d.logger.Debug("already staked")
		return nil
	}

	switch MechanismType(d.nodeType) {
	case BootstrapSequencer:
		return d.stakeBootstrapSequencer()
	case Sequencer:
		return d.stakeParticipant(activeParticipantsQuerier)
	case WatchTower:
		return d.stakeParticipant(activeParticipantsQuerier)
	default:
		panic("invalid node type: " + d.nodeType)
	}
}

// stakeBootstrapSequencer takes care of sequencer staking for the very first
// bootstrap sequencer. This needs special handling, because there won't be any
// other sequencer forging a block with staking transaction, therefore it must
// be build by the bootstrap node itself.
func (d *Avail) stakeBootstrapSequencer() error {
	// First, build the staking block.
	blockBuilderFactory := block.NewBlockBuilderFactory(d.blockchain, d.executor, d.logger)
	bb, err := blockBuilderFactory.FromBlockchainHead()
	if err != nil {
		return err
	}

	bb.SetCoinbaseAddress(d.minerAddr)
	bb.SignWith(d.signKey)

	stakeAmount := big.NewInt(0).Mul(big.NewInt(10), staking.ETH)
	tx, err := staking.StakeTx(d.minerAddr, stakeAmount, "sequencer", 1_000_000)
	if err != nil {
		return err
	}

	txSigner := &crypto.FrontierSigner{}
	tx, err = txSigner.SignTx(tx, d.signKey)
	if err != nil {
		return err
	}

	bb.AddTransactions(tx)
	blk, err := bb.Build()
	if err != nil {
		return err
	}

	for {
		d.logger.Debug("sending block with staking tx to Avail")
		err, malicious := d.sendBlockToAvail(blk)
		if err != nil {
			panic(err)
		}

		if !malicious {
			break
		}
	}

	d.logger.Debug("writing block with staking tx to local blockchain")

	err = d.blockchain.WriteBlock(blk, "sequencer")
	if err != nil {
		panic("bootstrap sequencer couldn't stake: " + err.Error())
	}

	return nil
}

func (d *Avail) stakeParticipant(activeParticipantsQuerier staking.ActiveParticipants) error {
	time.Sleep(5 * time.Second)

	stakeAmount := big.NewInt(0).Mul(big.NewInt(10), staking.ETH)
	tx, err := staking.StakeTx(d.minerAddr, stakeAmount, d.nodeType.String(), 1_000_000)
	if err != nil {
		return err
	}

	txSigner := &crypto.FrontierSigner{}
	tx, err = txSigner.SignTx(tx, d.signKey)
	if err != nil {
		return err
	}

	for retries := 0; retries < 10; retries++ {
		// Submit staking transaction for execution by active sequencer.
		err = d.txpool.AddTx(tx)
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}

		break
	}

	if err != nil {
		return err
	}

	// Syncer will be syncing the blockchain in the background, so once an active
	// sequencer picks up the staking transaction from the txpool, it becomes
	// effective and visible to us as well, via blockchain.
	var staked bool
	for !staked {
		staked, err = activeParticipantsQuerier.Contains(d.minerAddr, staking.NodeType(d.nodeType))
		if err != nil {
			return err
		}

		// Wait a bit before checking again.
		time.Sleep(3 * time.Second)
	}

	return nil
}
