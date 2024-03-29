package watchtower

import (
	"crypto/ecdsa"
	"errors"
	"fmt"

	"github.com/0xPolygon/polygon-edge/crypto"
	"github.com/0xPolygon/polygon-edge/state"
	"github.com/0xPolygon/polygon-edge/txpool"
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/availproject/op-evm/pkg/block"
	"github.com/availproject/op-evm/pkg/blockchain"
	"github.com/availproject/op-evm/pkg/staking"
	"github.com/hashicorp/go-hclog"
)

var (
	// ErrInvalidBlock is a general error used when the block structure is invalid or its field values are inconsistent.
	ErrInvalidBlock = errors.New("invalid block")

	// ErrParentBlockNotFound is returned when the local blockchain doesn't contain a block for the referenced parent hash.
	ErrParentBlockNotFound = errors.New("parent block not found")

	// FraudproofPrefix is a byte sequence that prefixes the fraudproof objected malicious block hash in the `ExtraData` of the fraudproof block header.
	FraudproofPrefix = []byte("FRAUDPROOF_OF:")
)

// WatchTower is an interface that defines methods for applying, checking, and constructing fraudproof blocks.
type WatchTower interface {
	Apply(blk *types.Block) error
	Check(blk *types.Block) error
	ConstructFraudproof(blk *types.Block) (*types.Block, error)
}

// watchTower implements the WatchTower interface and provides the actual implementation for the methods.
type watchTower struct {
	blockchain          *blockchain.Blockchain
	executor            *state.Executor
	txpool              *txpool.TxPool
	blockBuilderFactory block.BlockBuilderFactory
	logger              hclog.Logger

	account types.Address
	signKey *ecdsa.PrivateKey
}

// New creates a new instance of WatchTower with the provided parameters.
func New(blockchain *blockchain.Blockchain, executor *state.Executor, txp *txpool.TxPool, logger hclog.Logger, account types.Address, signKey *ecdsa.PrivateKey) WatchTower {
	return &watchTower{
		blockchain:          blockchain,
		executor:            executor,
		txpool:              txp,
		logger:              logger,
		blockBuilderFactory: block.NewBlockBuilderFactory(blockchain, executor, hclog.Default()),

		account: account,
		signKey: signKey,
	}
}

// Check checks the validity of a block by verifying it using the local blockchain.
// It returns an error if the block is invalid.
func (wt *watchTower) Check(blk *types.Block) error {
	if blk == nil {
		return fmt.Errorf("%w: block == nil", ErrInvalidBlock)
	}

	if blk.Header == nil {
		return fmt.Errorf("%w: block.Header == nil", ErrInvalidBlock)
	}

	if _, err := wt.blockchain.VerifyFinalizedBlock(blk); err != nil {
		wt.logger.Info("block cannot be verified", "block_number", blk.Number(), "block_hash", blk.Hash(), "parent_block_hash", blk.ParentHash(), "error", err)
		return err
	}

	return nil
}

// Apply applies a block to the blockchain by writing it to the blockchain and resetting the transaction pool.
func (wt *watchTower) Apply(blk *types.Block) error {
	if err := wt.blockchain.WriteBlock(blk, block.SourceWatchTower); err != nil {
		return fmt.Errorf("failed to write block: %w", err)
	}

	// after the block has been written we reset the txpool so that
	// the old transactions are removed
	wt.txpool.ResetWithHeaders(blk.Header)

	wt.logger.Info("Block committed to blockchain", "block_number", blk.Header.Number, "hash", blk.Header.Hash.String(), "txns", len(blk.Transactions))
	wt.logger.Debug("Received block header", "block_header", blk.Header)
	wt.logger.Debug("Received block transactions", "block_transactions", blk.Transactions)

	return nil
}

// ConstructFraudproof constructs a fraudproof block by challenging a malicious block and submitting the watchtower's stake.
// It returns the constructed fraudproof block if successful.
func (wt *watchTower) ConstructFraudproof(maliciousBlock *types.Block) (*types.Block, error) {
	builder, err := wt.blockBuilderFactory.FromParentHash(maliciousBlock.ParentHash())
	if err != nil {
		return nil, err
	}

	fraudProofTxs, err := constructFraudproofTxs(wt.account, maliciousBlock)
	if err != nil {
		return nil, err
	}

	hdr, _ := wt.blockchain.GetHeaderByHash(maliciousBlock.ParentHash())
	transition, err := wt.executor.BeginTxn(hdr.StateRoot, hdr, wt.account)
	if err != nil {
		return nil, err
	}

	txSigner := &crypto.FrontierSigner{}
	fpTx := fraudProofTxs[0]
	fpTx.Nonce = transition.GetNonce(fpTx.From)
	tx, err := txSigner.SignTx(fpTx, wt.signKey)
	if err != nil {
		return nil, err
	}

	if wt.txpool != nil { // Tests sometimes do not have txpool so we need to do this check.
		if err := wt.txpool.AddTx(tx); err != nil {
			wt.logger.Error("failed to add fraud proof txn to the pool", "error", err)
			return nil, err
		}
	}

	wt.logger.Info(
		"Applied dispute resolution transaction to the txpool",
		"hash", tx.Hash,
		"nonce", tx.Nonce,
		"account_from", tx.From,
	)

	// Build the block that is going to be sent out to the Avail.
	blk, err := builder.
		SetCoinbaseAddress(wt.account).
		SetGasLimit(maliciousBlock.Header.GasLimit).
		SetExtraDataField(block.KeyFraudProofOf, maliciousBlock.Hash().Bytes()).
		SetExtraDataField(block.KeyBeginDisputeResolutionOf, tx.Hash.Bytes()).
		AddTransactions(fraudProofTxs...).
		SignWith(wt.signKey).
		Build()

	if err != nil {
		return nil, err
	}

	return blk, nil
}

// constructFraudproofTxs returns a set of transactions that challenge the malicious block and submit the watchtower's stake.
func constructFraudproofTxs(watchtowerAddress types.Address, maliciousBlock *types.Block) ([]*types.Transaction, error) {
	bdrTx, err := constructBeginDisputeResolutionTx(watchtowerAddress, maliciousBlock)
	if err != nil {
		return []*types.Transaction{}, err
	}

	return []*types.Transaction{bdrTx}, nil
}

// constructBeginDisputeResolutionTx constructs a transaction for beginning the dispute resolution process.
func constructBeginDisputeResolutionTx(watchtowerAddress types.Address, maliciousBlock *types.Block) (*types.Transaction, error) {
	tx, err := staking.BeginDisputeResolutionTx(watchtowerAddress, types.BytesToAddress(maliciousBlock.Header.Miner), maliciousBlock.Header.GasLimit)
	if err != nil {
		return nil, err
	}

	return tx, nil
}
