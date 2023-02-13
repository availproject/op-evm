package avail

import (
	"crypto/ecdsa"
	"errors"
	"fmt"

	"github.com/0xPolygon/polygon-edge/blockchain"
	"github.com/0xPolygon/polygon-edge/crypto"
	"github.com/0xPolygon/polygon-edge/state"
	"github.com/0xPolygon/polygon-edge/types"
	stypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/hashicorp/go-hclog"
	"github.com/maticnetwork/avail-settlement/consensus/avail/validator"
	"github.com/maticnetwork/avail-settlement/pkg/avail"
	"github.com/maticnetwork/avail-settlement/pkg/block"
	"github.com/maticnetwork/avail-settlement/pkg/staking"
)

type Fraud struct {
	logger      hclog.Logger
	blockchain  *blockchain.Blockchain
	executor    *state.Executor
	validator   validator.Validator
	nodeAddr    types.Address
	nodeSignKey *ecdsa.PrivateKey
	availSender avail.Sender

	fraudBlock *types.Block
}

func (f *Fraud) SetBlock(b *types.Block) {
	f.fraudBlock = b
}

func (f *Fraud) GetBlock() *types.Block {
	return f.fraudBlock
}

func (f *Fraud) InProcess() bool {
	return f.fraudBlock != nil
}

func (f *Fraud) CheckAndSetFraudBlock(blocks []*types.Block) bool {
	for _, blk := range blocks {
		if fraudProofBlockHash, exists := block.GetExtraDataFraudProofTarget(blk.Header); exists {
			f.logger.Info(
				"Fraud proof parent hash discovered. Chain is entering into the dispute mode...",
				"probation_block_hash", fraudProofBlockHash,
			)
			f.SetBlock(blk)
			return true
		}
	}
	return false
}

func (f *Fraud) CheckAndSlash() (bool, error) {

	// There is no block attached from previous sequencer runs and therefore we assume
	// no fraud should be checked in this moment...
	if !f.InProcess() {
		f.logger.Info("no fraud to process")
		return false, nil
	}

	fraudBlockTargetHash, exists := block.GetExtraDataFraudProofTarget(f.fraudBlock.Header)
	if !exists {
		// It seems that fraud block is set but the proof target cannot be calculated
		// therefore we are going to log this problem and unset the fraud block as it seems
		// that there is corruption.
		// TODO: Figure out what to do with dispute resolution as if it's set, sequencer is
		// blocked. This should never happen and if it does, via malicious watchtower we should
		// be resolve this problem.
		f.logger.Error(
			"failed to extract proof target from the fraud block",
			"block_hash", f.fraudBlock.Hash(),
		)

		// Disregard entirely this specific fraud block
		f.SetBlock(nil)

		return false, fmt.Errorf(
			"failed to extract fraud proof targed from the fraud block hash %s",
			f.fraudBlock.Hash(),
		)
	}

	f.logger.Info(
		"Discovered fraud proof block hash targed",
		"targeted_block_hash", fraudBlockTargetHash,
		"watchtower_block_hash", f.fraudBlock.Hash(),
	)

	// Extract the potentially malicious block from the database but fetch only headers
	// as rest is not really needed.
	maliciousBlock, mbExists := f.blockchain.GetBlockByHash(fraudBlockTargetHash, false)
	if !mbExists {
		f.logger.Info(
			"Potentially malicious block not discovered, rejecting future verification",
			"watchtower_block_hash", f.fraudBlock.Hash(),
			"potentially_malicious_block_hash", fraudBlockTargetHash,
		)

		return false, fmt.Errorf(
			"failed to discover potentially malicious block hash: %s, watchtower_block_hash: %s",
			f.fraudBlock.Header.Hash, fraudBlockTargetHash,
		)
	}

	f.logger.Info(
		"Potentially malicious block discovered, processing with the check...",
		"watchtower_block_hash", f.fraudBlock.Hash(),
		"potentially_malicious_block_hash", maliciousBlock.Hash(),
	)

	sequencerAddr := types.BytesToAddress(maliciousBlock.Header.Miner)
	watchtowerAddr := types.BytesToAddress(f.fraudBlock.Header.Miner)

	// Slashing should not occur from the node that produced actual malicious block
	if sequencerAddr.String() == f.nodeAddr.String() {
		f.logger.Warn("Potentially malicious node cannot process (slash) block it produced",
			"malicious_addr", sequencerAddr,
			"node_addr", f.nodeAddr,
			"watchtower_block_hash", f.fraudBlock.Hash(),
			"potentially_malicious_block_hash", maliciousBlock.Hash(),
		)

		return false, errors.New(
			"potentially malicious node cannot process (slash) block it produced",
		)
	}

	// Discover who needs to be slashed.
	// If watchtower produced block that proves sequencer to be corrupted, sequencer needs to be slashed.
	// If watchtower produced block that proves sequencer to be correct, watchtower needs to be slashed.
	if err := f.validator.Check(maliciousBlock); err != nil {
		f.logger.Warn(
			"Fraud proof block check confirmed malicious block. Slashing sequencer...",
			"watchtower_block_hash", f.fraudBlock.Hash(),
			"potentially_malicious_block_hash", maliciousBlock.Hash(),
			"sequencer", sequencerAddr,
			"watchtower_addr", watchtowerAddr,
			"error", err,
		)

		if err := f.slashNode(sequencerAddr, maliciousBlock.Header); err != nil {
			f.logger.Error(
				"failed to slash node (sequencer)",
				"watchtower_block_hash", f.fraudBlock.Hash(),
				"potentially_malicious_block_hash", maliciousBlock.Hash(),
				"sequencer", sequencerAddr,
				"watchtower_addr", watchtowerAddr,
				"error", err,
			)
			return false, err
		}

		f.SetBlock(nil)
		return true, nil

	} else {
		f.logger.Warn(
			"Fraud proof block check confirmed block is not malicious. Slashing watchtower...",
			"watchtower_block_hash", f.fraudBlock.Hash(),
			"potentially_malicious_block_hash", maliciousBlock.Hash(),
			"sequencer", sequencerAddr,
			"watchtower_addr", watchtowerAddr,
			"error", err,
		)

		if err := f.slashNode(watchtowerAddr, maliciousBlock.Header); err != nil {
			f.logger.Error(
				"failed to slash node (watchtower)",
				"watchtower_block_hash", f.fraudBlock.Hash(),
				"potentially_malicious_block_hash", maliciousBlock.Hash(),
				"sequencer", sequencerAddr,
				"watchtower_addr", watchtowerAddr,
				"error", err,
			)
			return false, err
		}

		f.SetBlock(nil)
		return true, nil
	}
}

func (f *Fraud) slashNode(maliciousAddr types.Address, maliciousHeader *types.Header) error {
	// First, build the staking block.
	blockBuilderFactory := block.NewBlockBuilderFactory(f.blockchain, f.executor, f.logger)
	bb, err := blockBuilderFactory.FromBlockchainHead()
	if err != nil {
		return err
	}

	lastKnownCorrectHeader, ok := f.blockchain.GetBlockByHash(maliciousHeader.ParentHash, false)
	if !ok {
		return fmt.Errorf("failed to discover block by parent hash '%s'", maliciousHeader.ParentHash)
	}

	bb.SetParentStateRoot(lastKnownCorrectHeader.Header.StateRoot)
	bb.SetCoinbaseAddress(f.nodeAddr)
	bb.SignWith(f.nodeSignKey)

	tx, err := staking.SlashStakerTx(f.nodeAddr, maliciousAddr, 1_000_000)
	if err != nil {
		f.logger.Error("failed to construct slash transaction", "err", err)
		return err
	}

	txSigner := &crypto.FrontierSigner{}
	tx, err = txSigner.SignTx(tx, f.nodeSignKey)
	if err != nil {
		f.logger.Error("failed to sign slashing transaction", "err", err)
		return err
	}

	bb.AddTransactions(tx)
	blk, err := bb.Build()
	if err != nil {
		f.logger.Error("failed to build slashing block: ", "err", err)
		return err
	}

	f.logger.Info("sending block with slashing tx to Avail")
	err = f.availSender.SendAndWaitForStatus(blk, stypes.ExtrinsicStatus{IsInBlock: true})
	if err != nil {
		f.logger.Error("error while submitting slashing block to avail", "err", err)
		return err
	}

	f.logger.Info("writing block with slashing tx to local blockchain")

	err = f.blockchain.WriteBlock(blk, "bootstrap-sequencer")
	if err != nil {
		f.logger.Error("bootstrap sequencer couldn't slash staker: ", "err", err)
		return err
	}

	return nil
}

func NewFraudResolver(logger hclog.Logger, b *blockchain.Blockchain, e *state.Executor, v validator.Validator, nodeAddr types.Address, nodeSignKey *ecdsa.PrivateKey, availSender avail.Sender) *Fraud {
	return &Fraud{
		logger:      logger,
		blockchain:  b,
		executor:    e,
		validator:   v,
		nodeAddr:    nodeAddr,
		nodeSignKey: nodeSignKey,
		availSender: availSender,
	}
}
