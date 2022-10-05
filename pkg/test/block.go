package test

import (
	"crypto/ecdsa"
	"testing"
	"time"

	"github.com/0xPolygon/polygon-edge/consensus"
	"github.com/0xPolygon/polygon-edge/state"
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/maticnetwork/avail-settlement/pkg/block"
)

type BlockFactory interface {
	BuildBlock(parent *types.Block, txs []*types.Transaction) *types.Block
	BuildBlockWithTransition(parent *types.Block, transition *state.Transition, txs []*types.Transaction) *types.Block
	GetTransition(parent *types.Block) (*state.Transition, error)
}

type BasicBlockFactory struct {
	Executor *state.Executor
	Coinbase types.Address
	SignKey  *ecdsa.PrivateKey

	T *testing.T
}

func NewBasicBlockFactory(t *testing.T, executor *state.Executor, coinbaseAddr types.Address, signKey *ecdsa.PrivateKey) BlockFactory {
	return &BasicBlockFactory{
		Executor: executor,
		Coinbase: coinbaseAddr,
		SignKey:  signKey,
		T:        t,
	}
}

func (bbf *BasicBlockFactory) GetTransition(parent *types.Block) (*state.Transition, error) {
	header := &types.Header{
		ParentHash: parent.Hash(),
		Number:     parent.Number() + 1,
		Miner:      bbf.Coinbase.Bytes(),
		Nonce:      types.Nonce{},
		GasLimit:   parent.Header.GasLimit, // TODO(tuommaki): This needs adjusting.
		Timestamp:  uint64(time.Now().Unix()),
	}

	return bbf.Executor.BeginTxn(parent.Header.StateRoot, header, bbf.Coinbase)
}

func (bbf *BasicBlockFactory) BuildBlock(parent *types.Block, txs []*types.Transaction) *types.Block {
	header := &types.Header{
		ParentHash: parent.Hash(),
		Number:     parent.Number() + 1,
		Miner:      bbf.Coinbase.Bytes(),
		Nonce:      types.Nonce{},
		GasLimit:   parent.Header.GasLimit, // TODO(tuommaki): This needs adjusting.
		Timestamp:  uint64(time.Now().Unix()),
	}

	transition, err := bbf.Executor.BeginTxn(parent.Header.StateRoot, header, bbf.Coinbase)
	if err != nil {
		bbf.T.Fatal(err)
	}

	for _, tx := range txs {
		err := transition.Write(tx)
		if err != nil {
			// TODO(tuommaki): This needs re-assesment. Theoretically there
			// should NEVER be situation where fraud proof transaction writing
			// could fail and hence panic here is appropriate. There is some
			// debugging aspects though, which might need revisiting,
			// especially if the malicious block can cause situation where this
			// section fails - it would then defeat the purpose of watch tower.
			bbf.T.Fatal(err)
		}
	}

	// Commit the changes
	_, root := transition.Commit()

	// Update the header
	header.StateRoot = root
	header.GasUsed = transition.TotalGas()

	// Build the actual blk
	// The header hash is computed inside buildBlock
	blk := consensus.BuildBlock(consensus.BuildBlockParams{
		Header:   header,
		Txns:     txs,
		Receipts: transition.Receipts(),
	})

	err = block.PutValidatorExtra(blk.Header, &block.ValidatorExtra{Validators: []types.Address{bbf.Coinbase}})
	if err != nil {
		bbf.T.Fatalf("put validator extra failed: %s", err)
	}

	blk.Header, err = block.WriteSeal(bbf.SignKey, blk.Header)
	if err != nil {
		bbf.T.Fatalf("sealing block failed: %s", err)
	}

	// Compute the hash, this is only a provisional hash since the final one
	// is sealed after all the committed seals
	blk.Header.ComputeHash()

	return blk
}

func (bbf *BasicBlockFactory) BuildBlockWithTransition(parent *types.Block, transition *state.Transition, txs []*types.Transaction) *types.Block {
	header := &types.Header{
		ParentHash: parent.Hash(),
		Number:     parent.Number() + 1,
		Miner:      bbf.Coinbase.Bytes(),
		Nonce:      types.Nonce{},
		GasLimit:   parent.Header.GasLimit, // TODO(tuommaki): This needs adjusting.
		Timestamp:  uint64(time.Now().Unix()),
	}

	for _, tx := range txs {
		err := transition.Write(tx)
		if err != nil {
			// TODO(tuommaki): This needs re-assesment. Theoretically there
			// should NEVER be situation where fraud proof transaction writing
			// could fail and hence panic here is appropriate. There is some
			// debugging aspects though, which might need revisiting,
			// especially if the malicious block can cause situation where this
			// section fails - it would then defeat the purpose of watch tower.
			bbf.T.Fatal(err)
		}
	}

	// Commit the changes
	_, root := transition.Commit()

	// Update the header
	header.StateRoot = root
	header.GasUsed = transition.TotalGas()

	// Build the actual blk
	// The header hash is computed inside buildBlock
	blk := consensus.BuildBlock(consensus.BuildBlockParams{
		Header:   header,
		Txns:     txs,
		Receipts: transition.Receipts(),
	})

	err := block.PutValidatorExtra(blk.Header, &block.ValidatorExtra{Validators: []types.Address{bbf.Coinbase}})
	if err != nil {
		bbf.T.Fatalf("put validator extra failed: %s", err)
	}

	blk.Header, err = block.WriteSeal(bbf.SignKey, blk.Header)
	if err != nil {
		bbf.T.Fatalf("sealing block failed: %s", err)
	}

	// Compute the hash, this is only a provisional hash since the final one
	// is sealed after all the committed seals
	blk.Header.ComputeHash()

	return blk
}
