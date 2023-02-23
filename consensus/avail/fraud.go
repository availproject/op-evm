package avail

import (
	"bytes"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"time"

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

type DisputeProcessChainStatus uint16

var (
	ErrTxPoolHashNotFound = errors.New("hash not found in the txpool")

	ChainProcessingDisabled DisputeProcessChainStatus = 0
	ChainProcessingEnabled  DisputeProcessChainStatus = 1
)

type Fraud struct {
	logger                  hclog.Logger
	blockchain              *blockchain.Blockchain
	executor                *state.Executor
	txpool                  *txpool.TxPool
	validator               validator.Validator
	enableBlockProductionCh chan bool

	nodeAddr    types.Address
	nodeSignKey *ecdsa.PrivateKey
	availSender avail.Sender
	nodeType    MechanismType

	fraudBlock         *types.Block
	chainProcessStatus DisputeProcessChainStatus
}

func (f *Fraud) SetBlock(b *types.Block) {
	f.fraudBlock = b
}

func (f *Fraud) GetBlock() *types.Block {
	return f.fraudBlock
}

func (f *Fraud) SetChainStatus(status DisputeProcessChainStatus) {
	f.chainProcessStatus = status

	if status == ChainProcessingEnabled {
		f.enableBlockProductionCh <- true
	} else {
		f.enableBlockProductionCh <- false
	}
}

func (f *Fraud) IsChainDisabled() bool {
	return f.chainProcessStatus == ChainProcessingDisabled
}

func (f *Fraud) IsReadyToSlash() bool {
	if f.chainProcessStatus == ChainProcessingDisabled && f.fraudBlock != nil {
		return true
	}

	return false
}

func (f *Fraud) CheckAndSetFraudBlock(blocks []*types.Block) bool {
	for _, blk := range blocks {
		if fraudProofBlockHash, exists := block.GetExtraDataFraudProofTarget(blk.Header); exists {
			f.logger.Info(
				"Fraud proof parent hash block discovered. Continuing with fraud dispute resolution...",
				"probation_block_hash", fraudProofBlockHash,
				"watchtower_fraud_block_hash", blk.Hash(),
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

func (f *Fraud) EndDisputeResolution() {
	f.SetBlock(nil)
	f.SetChainStatus(ChainProcessingEnabled)
}

func (f *Fraud) ShouldStopProducingBlocks(activeParticipantsQuerier staking.ActiveParticipants) {
	for {
		// We've already received begin dispute resolution transaction. Now it's time to wait for
		// processing prior we check tx pool again...
		if f.IsChainDisabled() {
			time.Sleep(1 * time.Second)
			continue
		}

		f.txpool.Prepare()

	innerLoop:
		for {

			tx := f.txpool.Peek()
			if tx == nil {
				break innerLoop
			}

			isWatchtower, err := activeParticipantsQuerier.Contains(tx.From, staking.WatchTower)
			if err != nil {
				f.logger.Debug("failure while checking if tx from is active watchtower", "error", err)
				time.Sleep(1 * time.Second)
				continue
			}

			isBeginDisputeResolutionTx, err := staking.IsBeginDisputeResolutionTx(tx)
			if err != nil {
				f.logger.Debug("failure while checking if tx is type of begin dispute resolution", "error", err)
				time.Sleep(1 * time.Second)
				continue
			}

			f.logger.Debug(
				"New tx pool transaction discovered",
				"hash", tx.Hash,
				"value", tx.Value,
				"originating_addr", tx.From.String(),
				"recipient_addr", tx.To.String(),
				"submitted_via_watchtower", isWatchtower,
				"staking_contract_addr", staking.AddrStakingContract.String(),
				"submitted_towards_contract", bytes.Equal(tx.To.Bytes(), staking.AddrStakingContract.Bytes()),
				"tx_type_of_begin_dispute_resolution", isBeginDisputeResolutionTx,
			)

			if isWatchtower && bytes.Equal(tx.To.Bytes(), staking.AddrStakingContract.Bytes()) && isBeginDisputeResolutionTx {
				f.logger.Warn(
					"Discovered valid begin dispute resolution transaction. Chain is entering fraud dispute mode...",
					"originating_watchtower_addr", tx.From,
					"dispute_resolution_tx_hash", tx.Hash,
				)

				// We have proper transaction and therefore we are going to stop processing blocks in the chain
				f.SetChainStatus(ChainProcessingDisabled)
				break innerLoop
			}
		}

		// Just a bit of the time to not break the CPU...
		time.Sleep(100 * time.Millisecond)
	}
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

	return nil, ErrTxPoolHashNotFound
}

func (f *Fraud) GetBeginDisputeResolutionTxHash() types.Hash {
	hash, _ := block.GetExtraDataBeginDisputeResolutionTarget(f.fraudBlock.Header)
	return hash
}

func (f *Fraud) IsFraudProofBlock(blk *types.Block) bool {
	_, exists := block.GetExtraDataFraudProofTarget(blk.Header)
	return exists
}

func (f *Fraud) CheckAndSlash() (bool, error) {

	// There is no block attached from previous sequencer runs and therefore we assume
	// no fraud should be checked in this moment...
	if !f.IsReadyToSlash() {
		f.logger.Debug("not yet ready to process the block and slash the participant...")
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
	blockBuilderFactory := block.NewBlockBuilderFactory(f.blockchain, f.executor, f.logger)

	disputeBlk, err := f.produceBeginDisputeResolutionBlock(blockBuilderFactory, maliciousAddr, maliciousHeader, nodeType)
	if err != nil {
		return err
	}

	_, err = f.produceSlashBlock(blockBuilderFactory, disputeBlk, maliciousAddr, maliciousHeader, nodeType)
	if err != nil {
		return err
	}

	// No longer is it required for the chain to be in the disputed mode
	f.EndDisputeResolution()
	return nil
}

func (f *Fraud) produceBeginDisputeResolutionBlock(blockBuilderFactory block.BlockBuilderFactory, maliciousAddr types.Address, maliciousHeader *types.Header, nodeType MechanismType) (*types.Block, error) {
	bb, err := blockBuilderFactory.FromBlockchainHead()
	if err != nil {
		return nil, err
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
		return nil, err
	}
	bb.AddTransactions(disputeBeginTx)

	blk, err := bb.Build()
	if err != nil {
		f.logger.Error("failed to build begin dispute resolution block", "err", err)
		return nil, err
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
		return nil, err
	}

	err = f.blockchain.WriteBlock(blk, f.nodeType.String())
	if err != nil {
		f.logger.Error("failed to write begin dispute resolution block to the blockchain", "err", err)
		return nil, err
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

	return blk, nil
}

func (f *Fraud) produceSlashBlock(blockBuilderFactory block.BlockBuilderFactory, disputeBlk *types.Block, maliciousAddr types.Address, maliciousHeader *types.Header, nodeType MechanismType) (*types.Block, error) {
	slashBlk, err := blockBuilderFactory.FromBlockchainHead()
	if err != nil {
		return nil, err
	}

	slashBlk.SetCoinbaseAddress(f.nodeAddr)
	slashBlk.SignWith(f.nodeSignKey)

	disputeResolutionTx, err := staking.SlashStakerTx(f.nodeAddr, maliciousAddr, 1_000_000)
	if err != nil {
		f.logger.Error("failed to end new fraud dispute resolution", "err", err)
		return nil, err
	}

	hdr, _ := f.blockchain.GetHeaderByHash(disputeBlk.Hash())
	transition, err := f.executor.BeginTxn(hdr.StateRoot, hdr, f.nodeAddr)
	if err != nil {
		f.logger.Error("failed to begin the transition for the end dispute resolution", "err", err)
		return nil, err
	}
	disputeResolutionTx.Nonce = transition.GetNonce(disputeResolutionTx.From)

	txSigner := &crypto.FrontierSigner{}
	dtx, err := txSigner.SignTx(disputeResolutionTx, f.nodeSignKey)
	if err != nil {
		f.logger.Error("failed to sign slashing transaction", "err", err)
		return nil, err
	}

	slashBlk.AddTransactions(dtx)

	// Used to ensure we can end fraud dispute for a specific fraud block on all of the nodes!
	// Work, please work...
	slashBlk.SetExtraDataField(block.KeyEndDisputeResolutionOf, f.fraudBlock.Hash().Bytes())

	blk, err := slashBlk.Build()
	if err != nil {
		f.logger.Error("failed to build slashing block", "err", err)
		return nil, err
	}

	f.logger.Info(
		"Sending slashing block to the Avail",
		"hash", blk.Hash(),
		"malicious_block_hash", maliciousHeader.Hash,
		"parent_block_hash", maliciousHeader.ParentHash,
	)

	err = f.availSender.SendAndWaitForStatus(blk, stypes.ExtrinsicStatus{IsInBlock: true})
	if err != nil {
		f.logger.Error("error while submitting slashing block to avail", "err", err)
		return nil, err
	}

	err = f.blockchain.WriteBlock(blk, f.nodeType.String())
	if err != nil {
		f.logger.Error("failed to write slashing block to the blockchain", "err", err)
		return nil, err
	}

	// after the block has been written we reset the txpool so that
	// the old transactions are removed
	f.txpool.ResetWithHeaders(blk.Header)

	f.logger.Info(
		"Successfully sent and wrote slashing block to the blockchain... Resuming chain activity...",
		"txn_count", len(blk.Transactions),
		"hash", blk.Hash(),
		"block_number", blk.Number(),
		"malicious_block_hash", maliciousHeader.Hash,
		"parent_block_hash", maliciousHeader.ParentHash,
	)

	return blk, nil
}

func NewFraudResolver(logger hclog.Logger, b *blockchain.Blockchain, e *state.Executor, txp *txpool.TxPool, v validator.Validator, enableBlockProductionCh chan bool, nodeAddr types.Address, nodeSignKey *ecdsa.PrivateKey, availSender avail.Sender, nodeType MechanismType) *Fraud {
	return &Fraud{
		logger:                  logger,
		blockchain:              b,
		executor:                e,
		txpool:                  txp,
		validator:               v,
		nodeAddr:                nodeAddr,
		nodeType:                nodeType,
		nodeSignKey:             nodeSignKey,
		availSender:             availSender,
		chainProcessStatus:      ChainProcessingEnabled,
		enableBlockProductionCh: enableBlockProductionCh,
	}
}
