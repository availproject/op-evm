package block

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"time"

	"github.com/0xPolygon/polygon-edge/consensus"
	"github.com/0xPolygon/polygon-edge/state"
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/hashicorp/go-hclog"
)

// ErrInvalidHash represents an error indicating an invalid hash.
var ErrInvalidHash = errors.New("invalid hash")

// ErrSignKeyMissing represents an error indicating a missing signing key.
var ErrSignKeyMissing = errors.New("signing key missing")

// Builder provides a builder interface for constructing blocks.
type Builder interface {
	// SetBlockNumber sets the block number and returns the builder.
	SetBlockNumber(number uint64) Builder

	// SetCoinbaseAddress sets the coinbase address and returns the builder.
	SetCoinbaseAddress(coinbaseAddr types.Address) Builder

	// SetDifficulty sets the block difficulty and returns the builder.
	SetDifficulty(d uint64) Builder

	// SetExtraDataField sets an extra data field with a key-value pair and returns the builder.
	SetExtraDataField(key string, value []byte) Builder

	// SetGasLimit sets the gas limit and returns the builder.
	SetGasLimit(limit uint64) Builder

	// SetParentStateRoot sets the parent state root and returns the builder.
	SetParentStateRoot(parentRoot types.Hash) Builder

	// AddTransactions adds transactions to the block and returns the builder.
	AddTransactions(txs ...*types.Transaction) Builder

	// SignWith signs the block with the provided private key and returns the builder.
	SignWith(signKey *ecdsa.PrivateKey) Builder

	// Build constructs and returns the built block.
	Build() (*types.Block, error)

	// Write writes the built block to the specified source.
	Write(src string) error
}

type blockchain interface {
	CalculateGasLimit(number uint64) (uint64, error)
	GetHeaderByHash(types.Hash) (*types.Header, bool)
	Header() *types.Header
	WriteBlock(block *types.Block, source string) error
}

// blockBuilder is a builder for constructing blocks.
type blockBuilder struct {
	blockchain blockchain
	executor   *state.Executor
	logger     hclog.Logger

	coinbase   *types.Address
	difficulty *uint64
	parentRoot *types.Hash
	parentHash *types.Hash
	gasLimit   *uint64

	header *types.Header
	parent *types.Header

	transition   *state.Transition
	extraData    map[string][]byte
	transactions []*types.Transaction
	signKey      *ecdsa.PrivateKey
}

// BlockBuilderFactory is a factory interface for creating block builders.
type BlockBuilderFactory interface {
	// FromParentHash creates a block builder using the parent block's hash.
	FromParentHash(parent types.Hash) (Builder, error)

	// FromBlockchainHead creates a block builder using the blockchain's head.
	FromBlockchainHead() (Builder, error)
}

// blockBuilderFactory is a factory for creating block builders.
type blockBuilderFactory struct {
	blockchain blockchain
	executor   *state.Executor
	logger     hclog.Logger
}

// NewBlockBuilderFactory creates a new block builder factory.
func NewBlockBuilderFactory(blockchain blockchain, executor *state.Executor, logger hclog.Logger) BlockBuilderFactory {
	return &blockBuilderFactory{
		blockchain: blockchain,
		executor:   executor,
		logger:     logger.ResetNamed("block_builder_factory"),
	}
}

// FromBlockchainHead creates a block builder using the blockchain's head.
// It returns an error if the blockchain head is not available.
func (bbf *blockBuilderFactory) FromBlockchainHead() (Builder, error) {
	hdr := bbf.blockchain.Header()
	return bbf.FromParentHeader(hdr)
}

// FromParentHash creates a block builder using the parent block's hash.
// It returns an error if the parent block is not found in the blockchain.
func (bbf *blockBuilderFactory) FromParentHash(parent types.Hash) (Builder, error) {
	hdr, found := bbf.blockchain.GetHeaderByHash(parent)
	if !found {
		return nil, fmt.Errorf("%w: not found", ErrInvalidHash)
	}

	return bbf.FromParentHeader(hdr)
}

// FromParentHeader creates a block builder using the parent block's header.
func (bbf *blockBuilderFactory) FromParentHeader(parent *types.Header) (Builder, error) {
	bb := &blockBuilder{
		blockchain: bbf.blockchain,
		executor:   bbf.executor,
		logger:     bbf.logger.ResetNamed("block_builder"),

		header: &types.Header{
			ParentHash: parent.Hash,
			Number:     parent.Number + 1,
			GasLimit:   parent.GasLimit,
		},
		parent: parent,

		extraData: make(map[string][]byte),
	}

	return bb, nil
}

// SetBlockNumber sets the block number for the block builder.
func (bb *blockBuilder) SetBlockNumber(n uint64) Builder {
	bb.header.Number = n
	return bb
}

// SetCoinbaseAddress sets the coinbase address for the block builder.
func (bb *blockBuilder) SetCoinbaseAddress(coinbaseAddr types.Address) Builder {
	bb.coinbase = &coinbaseAddr
	return bb
}

// SetDifficulty sets the difficulty for the block builder.
func (bb *blockBuilder) SetDifficulty(d uint64) Builder {
	bb.difficulty = &d
	return bb
}

// SetExtraDataField sets the value of an extra data field for the block builder.
func (bb *blockBuilder) SetExtraDataField(key string, value []byte) Builder {
	bb.extraData[key] = value
	return bb
}

// SetGasLimit sets the gas limit for the block builder.
func (bb *blockBuilder) SetGasLimit(limit uint64) Builder {
	bb.gasLimit = &limit
	return bb
}

// SetParentStateRoot sets the parent state root for the block builder.
func (bb *blockBuilder) SetParentStateRoot(parentRoot types.Hash) Builder {
	bb.parentRoot = &parentRoot
	return bb
}

// AddTransactions adds one or more transactions to the block builder.
func (bb *blockBuilder) AddTransactions(tx ...*types.Transaction) Builder {
	bb.transactions = append(bb.transactions, tx...)
	return bb
}

// SignWith signs the block with the given private key.
func (bb *blockBuilder) SignWith(signKey *ecdsa.PrivateKey) Builder {
	bb.signKey = signKey
	return bb
}

// Write builds and writes the block to the blockchain.
func (bb *blockBuilder) Write(src string) error {
	blk, err := bb.Build()
	if err != nil {
		return err
	}

	err = bb.blockchain.WriteBlock(blk, src)
	if err != nil {
		return err
	}

	return nil
}

func (bb *blockBuilder) setDefaults() {
	if bb.coinbase == nil {
		bb.coinbase = new(types.Address)
		*bb.coinbase = types.BytesToAddress(bb.parent.Miner)
	}

	if bb.difficulty == nil {
		bb.difficulty = new(uint64)
		*bb.difficulty = 0
	}

	if bb.parentRoot == nil {
		bb.parentRoot = new(types.Hash)
		*bb.parentRoot = bb.parent.StateRoot
	}

	if bb.parentHash == nil {
		bb.parentHash = new(types.Hash)
		*bb.parentHash = bb.parent.ParentHash
	}

	if bb.gasLimit == nil {
		bb.gasLimit = new(uint64)
		*bb.gasLimit = 0
	}
}

// Build creates a new block using the provided parameters.
func (bb *blockBuilder) Build() (*types.Block, error) {
	var err error

	// ASSERTIONS
	if bb.signKey == nil {
		return nil, ErrSignKeyMissing
	}

	// Set defaults for missing unset parameters.
	bb.setDefaults()

	// Finalize header details before transaction processing.
	bb.header.Difficulty = *bb.difficulty
	bb.header.ExtraData = EncodeExtraDataFields(bb.extraData)
	bb.header.GasLimit = *bb.gasLimit
	bb.header.Miner = bb.coinbase.Bytes()
	bb.header.Timestamp = uint64(time.Now().Unix())

	// Copy gaslimit from genesis block if first post-genesis.
	if bb.header.GasLimit == 0 && bb.parent.Number == 0 {
		bb.header.GasLimit = bb.parent.GasLimit
	}

	// Check if the gas limit needs to be calculated based on parent block.
	if bb.header.GasLimit == 0 && bb.parent.Number != 0 {
		// Calculate gas limit based on parent header.
		bb.header.GasLimit, err = bb.blockchain.CalculateGasLimit(bb.parent.Number)
		if err != nil {
			return nil, err
		}
	}

	// Create a block transition.
	bb.transition, err = bb.executor.BeginTxn(*bb.parentRoot, bb.header, *bb.coinbase)
	if err != nil {
		return nil, err
	}

	// Write all transactions in-order.
	for _, tx := range bb.transactions {
		if tx.Nonce == 0 {
			tx.Nonce = bb.transition.GetNonce(tx.From)
		}

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
