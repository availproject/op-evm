package avail

import (
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/0xPolygon/polygon-edge/crypto"
	"github.com/0xPolygon/polygon-edge/types"
	stypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"

	"github.com/maticnetwork/avail-settlement/pkg/block"
	"github.com/maticnetwork/avail-settlement/pkg/common"
	"github.com/maticnetwork/avail-settlement/pkg/staking"
)

func (sw *SequencerWorker) waitForStakedSequencer(activeParticipantsQuerier staking.ActiveSequencers, nodeAddr types.Address) bool {
	for {
		sequencerStaked, sequencerError := activeParticipantsQuerier.Contains(nodeAddr)
		if sequencerError != nil {
			sw.logger.Error("failed to check if my account is among active staked sequencers. Retrying in few seconds...", "err", sequencerError)
			time.Sleep(3 * time.Second)
			continue
		}

		if !sequencerStaked {
			sw.logger.Warn("my account is not among active staked sequencers. Retrying in few seconds...", "address", nodeAddr.String())
			time.Sleep(3 * time.Second)
			continue
		}
		break
	}
	return true
}

func (d *Avail) ensureStaked(wg *sync.WaitGroup, activeParticipantsQuerier staking.ActiveParticipants) error {
	var nodeType staking.NodeType

	switch d.nodeType {
	case BootstrapSequencer, Sequencer:
		nodeType = staking.Sequencer
	case WatchTower:
		nodeType = staking.WatchTower
	default:
		return fmt.Errorf("unknown node type: %q", d.nodeType)
	}

	var returnErr error

	go func() {
		for {
			inProbation, err := activeParticipantsQuerier.InProbation(d.minerAddr)
			if err != nil {
				d.logger.Error("Failed to check if participant is currently in probation... Rechecking in a second...", "err", err)
				time.Sleep(3 * time.Second)
				continue
			}

			if inProbation {
				d.logger.Warn("Participant (node/miner) is currently in probation.... Rechecking in few seconds...", "err", err)
				time.Sleep(5 * time.Second)
				continue
			}

			staked, err := activeParticipantsQuerier.Contains(d.minerAddr, nodeType)
			if err != nil {
				d.logger.Error("Failed to check if participant exists... Rechecking in a second...", "err", err)
				time.Sleep(3 * time.Second)
				continue
			}

			if staked {
				d.logger.Debug("Node is successfully staked... Checking in 2 seconds for state change...")
				time.Sleep(3 * time.Second)
				continue
			}

			switch MechanismType(d.nodeType) {
			case BootstrapSequencer:
				// We cannot pass machine type as it won't be staked.
				// Bootstrap sequencer does not exist as category in the smart contract.
				returnErr = d.stakeParticipant(false, Sequencer.String())
			case Sequencer:
				returnErr = d.stakeParticipantThroughTxPool(activeParticipantsQuerier)
			case WatchTower:
				returnErr = d.stakeParticipantThroughTxPool(activeParticipantsQuerier)
			}
		}
	}()

	return returnErr
}

func (d *Avail) stakeParticipant(shouldWait bool, nodeType string) error {
	// Bootnode does not need to wait for any additional peers to be discovered prior pushing the
	// block towards rest of the community, however, sequencers and watchtowers must!
	if shouldWait {
		for {
			if d.network.GetBootnodeConnCount() > 0 {
				break
			}

			time.Sleep(1 * time.Second)
			continue
		}
	}

	// First, build the staking block.
	blockBuilderFactory := block.NewBlockBuilderFactory(d.blockchain, d.executor, d.logger)
	bb, err := blockBuilderFactory.FromBlockchainHead()
	if err != nil {
		return err
	}

	bb.SetCoinbaseAddress(d.minerAddr)
	bb.SignWith(d.signKey)

	stakeAmount := big.NewInt(0).Mul(big.NewInt(10), common.ETH)
	tx, err := staking.StakeTx(d.minerAddr, stakeAmount, nodeType, 1_000_000)
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
		d.logger.Error("failed to build staking block", "node_type", nodeType, "err", err)
		return err
	}

	d.logger.Debug("sending block with staking tx to Avail")
	err = d.availSender.SendAndWaitForStatus(blk, stypes.ExtrinsicStatus{IsInBlock: true})
	if err != nil {
		d.logger.Error("error while submitting data to avail", "error", err)
		return err
	}

	d.logger.Info(
		"Successfully wrote staking block to the blockchain",
		"hash", blk.Hash().String(),
	)

	err = d.blockchain.WriteBlock(blk, d.nodeType.String())
	if err != nil {
		return err
	}

	return nil
}

func (d *Avail) stakeParticipantThroughTxPool(activeParticipantsQuerier staking.ActiveParticipants) error {
	// We need to have at least one node available to be able successfully push tx to the neighborhood peers
	for {
		if d.network.GetBootnodeConnCount() > 0 {
			break
		}

		time.Sleep(1 * time.Second)
		continue
	}

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
