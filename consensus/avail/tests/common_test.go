package tests

import (
	"crypto/ecdsa"
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/0xPolygon/polygon-edge/blockchain"
	"github.com/0xPolygon/polygon-edge/chain"
	edge_crypto "github.com/0xPolygon/polygon-edge/crypto"
	"github.com/0xPolygon/polygon-edge/state"
	itrie "github.com/0xPolygon/polygon-edge/state/immutable-trie"
	"github.com/0xPolygon/polygon-edge/state/runtime/evm"
	"github.com/0xPolygon/polygon-edge/state/runtime/precompiled"
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/hashicorp/go-hclog"
	"github.com/maticnetwork/avail-settlement/pkg/block"
)

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

	bchain.SetConsensus(block.NewVerifier(hclog.Default()))

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
