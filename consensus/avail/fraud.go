package avail

import (
	"bytes"
	"crypto/ecdsa"
	"errors"
	"fmt"

	"github.com/0xPolygon/polygon-edge/blockchain"
	"github.com/0xPolygon/polygon-edge/crypto"
	"github.com/0xPolygon/polygon-edge/state"
	"github.com/0xPolygon/polygon-edge/txpool"
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
	txpool      *txpool.TxPool
	validator   validator.Validator
	nodeAddr    types.Address
	nodeSignKey *ecdsa.PrivateKey
	availSender avail.Sender
	nodeType    MechanismType

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

func (f *Fraud) IsDisputeResolutionEnded(blk *types.Header) bool {
	if f.fraudBlock == nil {
		return false
	}

	blkDisputeEndHash, _ := block.GetExtraDataEndDisputeResolutionTarget(blk)
	return bytes.Equal(f.fraudBlock.Hash().Bytes(), blkDisputeEndHash.Bytes())
}

func (f *Fraud) DiscoverDisputeResolutionTx(hash types.Hash) (*types.Transaction, error) {
	f.txpool.Prepare()

	for {
		tx := f.txpool.Peek()
		if tx == nil {
			break
		}

		if bytes.Equal(tx.Hash.Bytes(), hash.Bytes()) {
			f.logger.Info(
				"Discovered txpool dispute resolution transaction",
				"hash", tx.Hash,
				"nonce", tx.Nonce,
				"account_from", tx.From,
			)

			// no errors, pop the tx from the pool
			f.txpool.Pop(tx)
			return tx, nil
		}
	}

	return nil, fmt.Errorf("failed to discover dispute resolution tx at txpool hash: %s", hash)
}

func (f *Fraud) GetBeginDisputeResolutionTxHash() types.Hash {
	hash, _ := block.GetExtraDataBeginDisputeResolutionTarget(f.fraudBlock.Header)
	return hash
}

func (f *Fraud) IsFraudProofBlock(blk *types.Block) bool {
	_, exists := block.GetExtraDataFraudProofTarget(blk.Header)
	return exists
}

func (f *Fraud) CheckAndSlash(isMaliciousNode bool) (bool, error) {

	// There is no block attached from previous sequencer runs and therefore we assume
	// no fraud should be checked in this moment...
	if !f.InProcess() {
		f.logger.Info("no fraud to process")
		return false, nil
	}

	if isMaliciousNode {
		f.logger.Info("it is not my turn to check the potentially fraudlent block and slash. Skipping...")
		return false, errors.New("i am not allowed to check this particular fraudlent block")
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
		f.logger.Warn(
			"Potentially malicious node cannot process (slash) block it produced",
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
			"potentially_malicious_block_parent_hash", maliciousBlock.ParentHash(),
			"potentially_malicious_block_number", maliciousBlock.Number(),
			"sequencer", sequencerAddr,
			"watchtower_addr", watchtowerAddr,
			"error", err,
		)

		if err := f.slashNode(sequencerAddr, maliciousBlock.Header, Sequencer); err != nil {
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

		if err := f.slashNode(watchtowerAddr, maliciousBlock.Header, WatchTower); err != nil {
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
		return true, nil
	}
}

func (f *Fraud) slashNode(maliciousAddr types.Address, maliciousHeader *types.Header, nodeType MechanismType) error {
	// First, build the staking block.
	blockBuilderFactory := block.NewBlockBuilderFactory(f.blockchain, f.executor, f.logger)

	// Get the latest head
	bb, err := blockBuilderFactory.FromBlockchainHead()
	if err != nil {
		return err
	}

	bb.SetCoinbaseAddress(f.nodeAddr)
	bb.SignWith(f.nodeSignKey)

	// We are going to fork the chain but only if the malicious participant is sequencer.
	// Otherwise we are making sure we slash the watchtower and continue normal operation...
	if nodeType == Sequencer {
		bb.SetParentHash(maliciousHeader.ParentHash)
	}

	// Append begin disputed resolution txn
	disputeTxHash := f.GetBeginDisputeResolutionTxHash()
	f.logger.Info("Dispute resolution tx hash from fraud block", "hash", disputeTxHash.String())
	disputeBeginTx, err := f.DiscoverDisputeResolutionTx(disputeTxHash)
	if err != nil {
		f.logger.Error(
			"failed to discover begin dispute resoultion transaction for the block",
			"correct_block_hash", maliciousHeader.ParentHash,
			"err", err,
		)
		return err
	}
	bb.AddTransactions(disputeBeginTx)

	blk, err := bb.Build()
	if err != nil {
		f.logger.Error("failed to build begin dispute resolution block", "err", err)
		return err
	}

	f.logger.Info(
		"Sending begin dispute resolution block to the Avail",
		"hash", blk.Hash(),
		"malicious_block_hash", maliciousHeader.Hash,
		"parent_block_hash", maliciousHeader.ParentHash,
	)

	err = f.availSender.SendAndWaitForStatus(blk, stypes.ExtrinsicStatus{IsInBlock: true})
	if err != nil {
		f.logger.Error("error while submitting begin dispute resolution block to avail", "err", err)
		return err
	}

	err = f.blockchain.WriteBlock(blk, f.nodeType.String())
	if err != nil {
		f.logger.Error("failed to write begin dispute resolution block to the blockchain", "err", err)
		return err
	}

	// after the block has been written we reset the txpool so that
	// the old transactions are removed
	f.txpool.ResetWithHeaders(blk.Header)

	f.logger.Info(
		"Successfully sent and wrote begin dispute resolution block to the blockchain...",
		"txn_count", len(blk.Transactions),
		"hash", blk.Hash(),
		"block_number", blk.Number(),
		"malicious_block_hash", maliciousHeader.Hash,
		"parent_block_hash", maliciousHeader.ParentHash,
	)

	// Get the latest head
	slashBlk, err := blockBuilderFactory.FromBlockchainHead()
	if err != nil {
		return err
	}

	// We are going to fork the chain but only if the malicious participant is sequencer.
	// Otherwise we are making sure we slash the watchtower and continue normal operation...
	/* 	if nodeType == Sequencer {
		slashBlk.SetParentHash(maliciousHeader.ParentHash)
	} */

	// If sequencer set parent hash
	// bb.SetParentHash(maliciousHeader.ParentHash)

	slashBlk.SetCoinbaseAddress(f.nodeAddr)
	slashBlk.SignWith(f.nodeSignKey)

	disputeResolutionTx, err := staking.SlashStakerTx(f.nodeAddr, maliciousAddr, 1_000_000)
	if err != nil {
		f.logger.Error("failed to end new fraud dispute resolution", "err", err)
		return err
	}

	hdr, _ := f.blockchain.GetHeaderByHash(blk.Hash())
	transition, err := f.executor.BeginTxn(hdr.StateRoot, hdr, f.nodeAddr)
	if err != nil {
		f.logger.Error("failed to begin the transition for the end dispute resolution", "err", err)
		return err
	}
	disputeResolutionTx.Nonce = transition.GetNonce(disputeResolutionTx.From)

	txSigner := &crypto.FrontierSigner{}
	dtx, err := txSigner.SignTx(disputeResolutionTx, f.nodeSignKey)
	if err != nil {
		f.logger.Error("failed to sign slashing transaction", "err", err)
		return err
	}

	slashBlk.AddTransactions(dtx)

	// Used to ensure we can end fraud dispute for a specific fraud block on all of the nodes!
	// Work, please work...
	slashBlk.SetExtraDataField(block.KeyEndDisputeResolutionOf, f.fraudBlock.Hash().Bytes())

	blk2, err := slashBlk.Build()
	if err != nil {
		f.logger.Error("failed to build slashing block", "err", err)
		return err
	}

	f.logger.Info(
		"Sending slashing block to the Avail",
		"hash", blk2.Hash(),
		"malicious_block_hash", maliciousHeader.Hash,
		"parent_block_hash", maliciousHeader.ParentHash,
	)

	err = f.availSender.SendAndWaitForStatus(blk2, stypes.ExtrinsicStatus{IsInBlock: true})
	if err != nil {
		f.logger.Error("error while submitting slashing block to avail", "err", err)
		return err
	}

	err = f.blockchain.WriteBlock(blk2, f.nodeType.String())
	if err != nil {
		f.logger.Error("failed to write slashing block to the blockchain", "err", err)
		return err
	}

	// after the block has been written we reset the txpool so that
	// the old transactions are removed
	f.txpool.ResetWithHeaders(blk2.Header)

	f.logger.Info(
		"Successfully sent and wrote slashing block to the blockchain... Resuming with the chain activity...",
		"txn_count", len(blk2.Transactions),
		"hash", blk2.Hash(),
		"block_number", blk2.Number(),
		"malicious_block_hash", maliciousHeader.Hash,
		"parent_block_hash", maliciousHeader.ParentHash,
	)

	return nil
}

func NewFraudResolver(logger hclog.Logger, b *blockchain.Blockchain, e *state.Executor, txp *txpool.TxPool, v validator.Validator, nodeAddr types.Address, nodeSignKey *ecdsa.PrivateKey, availSender avail.Sender, nodeType MechanismType) *Fraud {
	return &Fraud{
		logger:      logger,
		blockchain:  b,
		executor:    e,
		txpool:      txp,
		validator:   v,
		nodeAddr:    nodeAddr,
		nodeType:    nodeType,
		nodeSignKey: nodeSignKey,
		availSender: availSender,
	}
}
