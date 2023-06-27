package avail

import (
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/0xPolygon/polygon-edge/crypto"
	stypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"

	"github.com/availproject/op-evm/pkg/block"
	"github.com/availproject/op-evm/pkg/common"
	"github.com/availproject/op-evm/pkg/staking"
)

// ensureStaked verifies whether a node is staked in the network.
// It takes as arguments a WaitGroup and an ActiveParticipants object.
// It determines the node type and checks if the node is under probation.
// If the node is not under probation and not already staked, the function tries to stake it
// and returns an error if staking fails.
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

	inProbation, err := activeParticipantsQuerier.InProbation(d.minerAddr)
	if err != nil {
		d.logger.Error("failed to check if participant is currently in probation", "error", err)
		return err
	}

	if inProbation {
		d.logger.Warn("Participant (node/miner) is currently in probation.", "error", err)
		return errors.New("participant is under probation")
	}

	staked, err := activeParticipantsQuerier.Contains(d.minerAddr, nodeType)
	if err != nil {
		d.logger.Error("failed to check if participant exists...", "error", err)
		return err
	}

	if staked {
		d.logger.Info("Node is successfully staked...")
		return nil
	}

	switch MechanismType(d.nodeType) {
	case BootstrapSequencer:
		// Staking smart contract does not support `BootstrapSequencer` MachineType.
		if returnErr := d.stakeParticipant(false, Sequencer.String()); returnErr != nil {
			return returnErr
		}
	case Sequencer, WatchTower:
		staked, returnErr := d.stakeParticipantThroughTxPool(activeParticipantsQuerier)
		if returnErr != nil {
			return returnErr
		}

		if staked {
			return nil
		}
	}

	return nil
}

// stakeParticipant stakes a participant in the network.
// It takes as arguments a boolean value indicating whether to wait for discovery of additional peers
// before pushing the block towards the rest of the community, and a string representing the node type.
// It first builds a staking block, signs it, and then submits it to the Avail network.
// After a successful submission, it writes the block to the local blockchain.
// Function is used only if staked participant is bootstrap sequencer.
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
		d.logger.Error("failed to build staking block", "node_type", nodeType, "error", err)
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

// stakeParticipantThroughTxPool stakes a participant through the transaction pool.
// It takes as argument an ActiveParticipants object.
// Before proceeding, it checks for network connection.
// It creates and signs a staking transaction, and attempts to add it to the transaction pool,
// retrying up to 10 times if unsuccessful. If successful, it waits for the main sequencer loop
// to do the synchronization.
// Function is used only if staked participant is sequencer or watchtower.
func (d *Avail) stakeParticipantThroughTxPool(activeParticipantsQuerier staking.ActiveParticipants) (bool, error) {
	// We need to have at least one node available to be able successfully push tx
	// to the neighborhood peers.
	for d.network == nil || d.network.GetBootnodeConnCount() < 1 {
		time.Sleep(1 * time.Second)
	}

	// XXX: This is a workaround for now.
	// TODO: Fix this with peer check to get rid of static sleep.
	// Apparently, we still need to wait a bit more time than boot node count to be able
	// process staking. If there's only bootstrap sequencer and one sequencer without this sleep
	// txpool tx will be added but bootstrap sequencer won't receive it.
	time.Sleep(5 * time.Second)

	stakeAmount := big.NewInt(0).Mul(big.NewInt(10), common.ETH)
	tx, err := staking.StakeTx(d.minerAddr, stakeAmount, d.nodeType.String(), 1_000_000)
	if err != nil {
		return false, err
	}

	txSigner := &crypto.FrontierSigner{}
	tx, err = txSigner.SignTx(tx, d.signKey)
	if err != nil {
		return false, err
	}

	for retries := 0; retries < 10; retries++ {
		d.logger.Info("Submitting stake to the tx pool", "retry", retries)
		// Submit staking transaction for execution by active sequencer.
		if err := d.txpool.AddTx(tx); err != nil {
			d.logger.Error("failure to add staking tx to the txpool err: %s", err)
			time.Sleep(1 * time.Second)
			continue
		}
		d.logger.Info("Stake submitted to the tx pool", "retry", retries)
		break
	}

	if err != nil {
		return false, err
	}

	// Assume staked if it's sent as we're going to wait for main sequencer loop to
	// do the synchronization...
	return true, nil
}
