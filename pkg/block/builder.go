package block

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"time"

	"github.com/0xPolygon/polygon-edge/blockchain"
	"github.com/0xPolygon/polygon-edge/consensus"
	"github.com/0xPolygon/polygon-edge/state"
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/hashicorp/go-hclog"
)

var (
	ErrInvalidHash    = errors.New("invalid hash")
	ErrSignKeyMissing = errors.New("signing key missing")
)

// Builder provides a builder interface for constructing blocks.
type Builder interface {
	SetBlockNumber(number uint64) Builder
	SetCoinbaseAddress(coinbaseAddr types.Address) Builder
	SetParentStateRoot(parentRoot types.Hash) Builder

	AddTransactions(txs ...*types.Transaction) Builder

	SignWith(signKey *ecdsa.PrivateKey) Builder

	Build() (*types.Block, error)
}

type blockBuilder struct {
	executor *state.Executor

	coinbase   *types.Address
	parentRoot *types.Hash

	header *types.Header
	parent *types.Header

	transition   *state.Transition
	transactions []*types.Transaction
	signKey      *ecdsa.PrivateKey
}

type BlockBuilderFactory interface {
	FromParentHash(hash types.Hash) (Builder, error)
}

type blockBuilderFactory struct {
	blockchain *blockchain.Blockchain
	executor   *state.Executor
	logger     hclog.Logger
}

func NewBlockBuilderFactory(blockchain *blockchain.Blockchain, executor *state.Executor, logger hclog.Logger) BlockBuilderFactory {
	return &blockBuilderFactory{
		blockchain: blockchain,
		executor:   executor,
		logger:     logger.Named("block_build_factory"),
	}
}

func (bbf *blockBuilderFactory) FromParentHash(parent types.Hash) (Builder, error) {
	hdr, found := bbf.blockchain.GetHeaderByHash(parent)
	if !found {
		return nil, fmt.Errorf("%w: not found", ErrInvalidHash)
	}

	return bbf.FromParentHeader(hdr)
}

func (bbf *blockBuilderFactory) FromParentHeader(parent *types.Header) (Builder, error) {
	bb := &blockBuilder{
		executor: bbf.executor,

		header: &types.Header{
			ParentHash: parent.Hash,
			Number:     parent.Number + 1,
			GasLimit:   parent.GasLimit,
		},
		parent: parent,
	}

	return bb, nil
}

func (bb *blockBuilder) SetBlockNumber(n uint64) Builder {
	bb.header.Number = n
	return bb
}

func (bb *blockBuilder) SetCoinbaseAddress(coinbaseAddr types.Address) Builder {
	bb.coinbase = &coinbaseAddr
	return bb
}

func (bb *blockBuilder) SetParentStateRoot(parentRoot types.Hash) Builder {
	bb.parentRoot = &parentRoot
	return bb
}

func (bb *blockBuilder) AddTransactions(tx ...*types.Transaction) Builder {
	bb.transactions = append(bb.transactions, tx...)
	return bb
}

func (bb *blockBuilder) SignWith(signKey *ecdsa.PrivateKey) Builder {
	bb.signKey = signKey
	return bb
}

func (bb *blockBuilder) setDefaults() {
	if bb.coinbase == nil {
		bb.coinbase = new(types.Address)
		*bb.coinbase = types.BytesToAddress(bb.parent.Miner)
	}

	if bb.parentRoot == nil {
		bb.parentRoot = new(types.Hash)
		*bb.parentRoot = bb.parent.StateRoot
	}
}

func (bb *blockBuilder) Build() (*types.Block, error) {
	var err error

	// ASSERTIONS
	if bb.signKey == nil {
		return nil, ErrSignKeyMissing
	}

	// Set defaults for missing unset parameters.
	bb.setDefaults()

	// Finalize header details before transaction processing.
	bb.header.Miner = bb.coinbase.Bytes()
	bb.header.Timestamp = uint64(time.Now().Unix())

	// Create a block transition.
	bb.transition, err = bb.executor.BeginTxn(*bb.parentRoot, bb.header, *bb.coinbase)
	if err != nil {
		return nil, err
	}

	// Write all transactions in-order.
	for _, tx := range bb.transactions {
		err := bb.transition.Write(tx)
		if err != nil {
			return nil, err
		}
	}
	// Commit the changes.
	_, root := bb.transition.Commit()

	// Update the headers.
	bb.header.StateRoot = root
	bb.header.GasUsed = bb.transition.TotalGas()

	// Build the actual block.
	// The header hash is computed inside BuildBlock().
	blk := consensus.BuildBlock(consensus.BuildBlockParams{
		Header:   bb.header,
		Txns:     bb.transactions,
		Receipts: bb.transition.Receipts(),
	})

	// Initialize the block header's `ExtraData`.
	err = PutValidatorExtra(blk.Header, &ValidatorExtra{Validators: []types.Address{types.BytesToAddress(bb.header.Miner)}})
	if err != nil {
		return nil, err
	}

	// ...and sign the block.
	blk.Header, err = WriteSeal(bb.signKey, blk.Header)
	if err != nil {
		return nil, err
	}

	// Compute the hash, this is only a provisional hash since the final one
	// is sealed after all the committed seals.
	blk.Header.ComputeHash()

	return blk, nil
}
