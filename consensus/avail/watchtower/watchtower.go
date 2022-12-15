package watchtower

import (
	"crypto/ecdsa"
	"errors"
	"fmt"

	"github.com/0xPolygon/polygon-edge/blockchain"
	"github.com/0xPolygon/polygon-edge/state"
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/hashicorp/go-hclog"
	"github.com/maticnetwork/avail-settlement/pkg/block"
	"github.com/maticnetwork/avail-settlement/pkg/staking"
)

var (
	// ErrInvalidBlock is general error used when block structure is invalid
	// or its field values are inconsistent.
	ErrInvalidBlock = errors.New("invalid block")

	// ErrParentBlockNotFound is returned when the local blockchain doesn't
	// contain block for the referenced parent hash.
	ErrParentBlockNotFound = errors.New("parent block not found")

	// FraudproofPrefix is byte sequence that prefixes the fraudproof objected
	// malicious block hash in `ExtraData` of the fraudproof block header.
	FraudproofPrefix = []byte("FRAUDPROOF_OF:")
)

type WatchTower interface {
	Apply(blk *types.Block) error
	Check(blk *types.Block) error
	ConstructFraudproof(blk *types.Block) (*types.Block, error)
}

type watchTower struct {
	blockchain          *blockchain.Blockchain
	executor            *state.Executor
	blockBuilderFactory block.BlockBuilderFactory
	logger              hclog.Logger

	account types.Address
	signKey *ecdsa.PrivateKey
}

func New(blockchain *blockchain.Blockchain, executor *state.Executor, logger hclog.Logger, account types.Address, signKey *ecdsa.PrivateKey) WatchTower {
	return &watchTower{
		blockchain:          blockchain,
		executor:            executor,
		logger:              logger,
		blockBuilderFactory: block.NewBlockBuilderFactory(blockchain, executor, hclog.Default()),

		account: account,
		signKey: signKey,
	}
}

func (wt *watchTower) Check(blk *types.Block) error {
	if blk == nil {
		return fmt.Errorf("%w: block == nil", ErrInvalidBlock)
	}

	if blk.Header == nil {
		return fmt.Errorf("%w: block.Header == nil", ErrInvalidBlock)
	}

	if err := wt.blockchain.VerifyFinalizedBlock(blk); err != nil {
		wt.logger.Info("block cannot be verified", "block_number", blk.Number(), "block_hash", blk.Hash(), "error", err)
		return err
	}

	return nil
}

func (wt *watchTower) Apply(blk *types.Block) error {
	if err := wt.blockchain.WriteBlock(blk, block.SourceWatchTower); err != nil {
		return fmt.Errorf("failed to write block: %w", err)
	}

	wt.logger.Info("Block committed to blockchain", "block_number", blk.Header.Number, "hash", blk.Header.Hash.String())
	wt.logger.Debug("Received block header", "block_header", blk.Header)
	wt.logger.Debug("Received block transactions", "block_transactions", blk.Transactions)

	return nil
}

func (wt *watchTower) ConstructFraudproof(maliciousBlock *types.Block) (*types.Block, error) {
	builder, err := wt.blockBuilderFactory.FromParentHash(maliciousBlock.ParentHash())
	if err != nil {
		return nil, err
	}

	fraudProofTxs, err := constructFraudproofTxs(wt.account, maliciousBlock)
	if err != nil {
		return nil, err
	}

	blk, err := builder.
		SetCoinbaseAddress(wt.account).
		SetGasLimit(maliciousBlock.Header.GasLimit).
		SetExtraDataField(block.KEY_FRAUDPROOF_OF, maliciousBlock.Hash().Bytes()).
		AddTransactions(fraudProofTxs...).
		SignWith(wt.signKey).
		Build()

	if err != nil {
		return nil, err
	}

	return blk, nil
}

// constructFraudproofTxs returns set of transactions that challenge the
// malicious block and submit watchtower's stake.
func constructFraudproofTxs(watchtowerAddress types.Address, maliciousBlock *types.Block) ([]*types.Transaction, error) {
	bdrTx, err := constructBeginDisputeResolutionTx(watchtowerAddress, maliciousBlock)
	if err != nil {
		return []*types.Transaction{}, err
	}

	return []*types.Transaction{bdrTx}, nil
}

func constructBeginDisputeResolutionTx(watchtowerAddress types.Address, maliciousBlock *types.Block) (*types.Transaction, error) {
	tx, err := staking.BeginDisputeResolutionTx(watchtowerAddress, types.BytesToAddress(maliciousBlock.Header.Miner), maliciousBlock.Header.GasLimit)
	if err != nil {
		return nil, err
	}

	return tx, nil
}
