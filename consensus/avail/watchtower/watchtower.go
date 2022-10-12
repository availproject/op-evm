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

func New(blockchain *blockchain.Blockchain, executor *state.Executor, account types.Address, signKey *ecdsa.PrivateKey) WatchTower {
	return &watchTower{
		blockchain:          blockchain,
		executor:            executor,
		logger:              hclog.Default(),
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
		wt.logger.Info("block %d (%q) cannot be verified: %s", blk.Number(), blk.Hash(), err)
		return err
	}

	return nil
}

func (wt *watchTower) Apply(blk *types.Block) error {
	if err := wt.blockchain.WriteBlock(blk, block.SourceWatchTower); err != nil {
		return fmt.Errorf("failed to write block while bulk syncing: %w", err)
	}

	wt.logger.Debug("Received block header: %+v \n", blk.Header)
	wt.logger.Debug("Received block transactions: %+v \n", blk.Transactions)

	return nil
}

func (wt *watchTower) ConstructFraudproof(maliciousBlock *types.Block) (*types.Block, error) {
	builder, err := wt.blockBuilderFactory.FromParentHash(maliciousBlock.ParentHash())
	if err != nil {
		return nil, err
	}

	blk, err := builder.
		SetCoinbaseAddress(wt.account).
		SetGasLimit(maliciousBlock.Header.GasLimit).
		SetExtraDataField(block.KeyFraudproof, maliciousBlock.Hash().Bytes()).
		AddTransactions(constructFraudproofTxs(maliciousBlock)...).
		SignWith(wt.signKey).
		Build()

	if err != nil {
		return nil, err
	}

	return blk, nil
}

// constructFraudproofTxs returns set of transactions that challenge the
// malicious block and submit watchtower's stake.
func constructFraudproofTxs(maliciousBlock *types.Block) []*types.Transaction {
	return []*types.Transaction{}
}
