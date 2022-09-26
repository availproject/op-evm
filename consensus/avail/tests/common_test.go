package tests

import (
	"crypto/ecdsa"
	"crypto/rand"
	"math/big"
	"testing"
	"time"

	"github.com/0xPolygon/polygon-edge/blockchain"
	"github.com/0xPolygon/polygon-edge/chain"
	"github.com/0xPolygon/polygon-edge/consensus"
	edge_crypto "github.com/0xPolygon/polygon-edge/crypto"
	"github.com/0xPolygon/polygon-edge/state"
	itrie "github.com/0xPolygon/polygon-edge/state/immutable-trie"
	"github.com/0xPolygon/polygon-edge/state/runtime/evm"
	"github.com/0xPolygon/polygon-edge/state/runtime/precompiled"
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/hashicorp/go-hclog"
	"github.com/maticnetwork/avail-settlement/consensus/avail"
	"github.com/maticnetwork/avail-settlement/pkg/block"
)

type BlockFactory interface {
	BuildBlock(parent *types.Block, txs []*types.Transaction) *types.Block
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

		T: t,
	}
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

func newAccount(t *testing.T) (types.Address, *ecdsa.PrivateKey) {
	t.Helper()

	privateKey, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	pk := privateKey.Public().(*ecdsa.PublicKey)
	address := edge_crypto.PubKeyToAddress(pk)

	return address, privateKey
}

func newBlockchain(t *testing.T) (*state.Executor, *blockchain.Blockchain) {
	chain := newChain(t)
	executor := newInMemExecutor(t, chain)

	gr := executor.WriteGenesis(chain.Genesis.Alloc)
	chain.Genesis.StateRoot = gr

	bchain, err := blockchain.NewBlockchain(hclog.Default(), "", chain, nil, executor)
	if err != nil {
		t.Fatal(err)
	}

	bchain.SetConsensus(avail.NewVerifier(hclog.Default()))

	executor.GetHash = bchain.GetHashHelper

	err = bchain.ComputeGenesis()
	if err != nil {
		t.Fatal(err)
	}

	return executor, bchain
}

func newInMemExecutor(t *testing.T, c *chain.Chain) *state.Executor {
	t.Helper()

	storage := itrie.NewMemoryStorage()
	st := itrie.NewState(storage)

	e := state.NewExecutor(c.Params, st, hclog.Default())

	e.SetRuntime(precompiled.NewPrecompiled())
	e.SetRuntime(evm.NewEVM())

	return e
}

func newChain(t *testing.T) *chain.Chain {
	balance := new(big.Int)
	balance.SetString("0x3635c9adc5dea00000", 16)

	return &chain.Chain{
		Genesis: &chain.Genesis{
			Alloc: map[types.Address]*chain.GenesisAccount{
				types.StringToAddress("0x064A4a5053F3de5eacF5E72A2E97D5F9CF55f031"): &chain.GenesisAccount{
					Balance: balance,
				},
			},
		},
		Params: &chain.Params{
			Forks: &chain.Forks{
				Homestead:      new(chain.Fork),
				Byzantium:      new(chain.Fork),
				Constantinople: new(chain.Fork),
				Petersburg:     new(chain.Fork),
				Istanbul:       new(chain.Fork),
				EIP150:         new(chain.Fork),
				EIP158:         new(chain.Fork),
				EIP155:         new(chain.Fork),
			},
			ChainID: 100,
			Engine: map[string]interface{}{
				"avail": map[string]interface{}{
					"mechanisms": []string{"sequencer", "validator"},
				},
			},
		},
	}
}

func getHeadBlock(t *testing.T, blockchain *blockchain.Blockchain) *types.Block {
	var head *types.Block

	var headBlockHash types.Hash
	hdr := blockchain.Header()
	if hdr != nil {
		headBlockHash = hdr.Hash
	} else {
		headBlockHash = blockchain.Genesis()
	}

	var ok bool
	head, ok = blockchain.GetBlockByHash(headBlockHash, true)
	if !ok {
		t.Fatal("couldn't fetch head block from blockchain")
	}

	return head
}
