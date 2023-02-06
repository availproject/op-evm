package avail

import (
	"fmt"
	"math/big"
	"time"

	"github.com/0xPolygon/polygon-edge/crypto"
	"github.com/0xPolygon/polygon-edge/types"
	stypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"

	"github.com/maticnetwork/avail-settlement/pkg/block"
	"github.com/maticnetwork/avail-settlement/pkg/common"
	"github.com/maticnetwork/avail-settlement/pkg/staking"
)

func (d *Avail) ensureStaked(activeParticipantsQuerier staking.ActiveParticipants) error {
	var nodeType staking.NodeType
	switch d.nodeType {
	case BootstrapSequencer, Sequencer:
		nodeType = staking.Sequencer
	case WatchTower:
		nodeType = staking.WatchTower
	case Validator:
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

	stakeAmount := big.NewInt(0).Mul(big.NewInt(10), common.ETH)
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

	d.logger.Debug("sending block with staking tx to Avail")
	err = d.availSender.SendAndWaitForStatus(blk, stypes.ExtrinsicStatus{IsInBlock: true})
	if err != nil {
		d.logger.Error("error while submitting data to avail", "error", err)
		panic(err)
	}

	d.logger.Debug("writing block with staking tx to local blockchain")

	err = d.blockchain.WriteBlock(blk, "sequencer")
	if err != nil {
		panic("bootstrap sequencer couldn't stake: " + err.Error())
	}

	return nil
}

func (d *Avail) stakeParticipant(activeParticipantsQuerier staking.ActiveParticipants) error {
	// Sleep time is added because we need to stake the participant after the peer discovery is
	// fully setup. If removed, transaction will never go through from watchtower to sequencer.
	time.Sleep(5 * time.Second)

	stakeAmount := big.NewInt(0).Mul(big.NewInt(10), common.ETH)
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

func (d *Avail) slashNode(maliciousAddr types.Address, maliciousHeader *types.Header) error {
	// First, build the staking block.
	blockBuilderFactory := block.NewBlockBuilderFactory(d.blockchain, d.executor, d.logger)
	bb, err := blockBuilderFactory.FromBlockchainHead()
	if err != nil {
		return err
	}

	lastKnownCorrectHeader, ok := d.blockchain.GetBlockByHash(maliciousHeader.ParentHash, false)
	if !ok {
		return fmt.Errorf("failed to discover block by parent hash '%s'", maliciousHeader.ParentHash)
	}

	bb.SetParentStateRoot(lastKnownCorrectHeader.Header.StateRoot)
	bb.SetCoinbaseAddress(d.minerAddr)
	bb.SignWith(d.signKey)

	tx, err := staking.SlashStakerTx(d.minerAddr, maliciousAddr, 1_000_000)
	if err != nil {
		d.logger.Error("failed to construct slash transaction", "err", err)
		return err
	}

	txSigner := &crypto.FrontierSigner{}
	tx, err = txSigner.SignTx(tx, d.signKey)
	if err != nil {
		d.logger.Error("failed to sign slashing transaction", "err", err)
		return err
	}

	bb.AddTransactions(tx)
	blk, err := bb.Build()
	if err != nil {
		d.logger.Error("failed to build slashing block: ", "err", err)
		return err
	}

	d.logger.Info("sending block with slashing tx to Avail")
	err = d.availSender.SendAndWaitForStatus(blk, stypes.ExtrinsicStatus{IsInBlock: true})
	if err != nil {
		d.logger.Error("error while submitting slashing block to avail", "err", err)
		return err
	}

	d.logger.Info("writing block with slashing tx to local blockchain")

	err = d.blockchain.WriteBlock(blk, "sequencer")
	if err != nil {
		d.logger.Error("bootstrap sequencer couldn't slash staker: ", "err", err)
		return err
	}

	return nil
}
