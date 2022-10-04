package test

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
)

func NewAccount(t *testing.T) (types.Address, *ecdsa.PrivateKey) {
	t.Helper()

	privateKey, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	pk := privateKey.Public().(*ecdsa.PublicKey)
	address := edge_crypto.PubKeyToAddress(pk)

	return address, privateKey
}

func NewBlockchain(t *testing.T) (*state.Executor, *blockchain.Blockchain) {
	chain := newChain(t)
	executor := newInMemExecutor(t, chain)

	gr := executor.WriteGenesis(chain.Genesis.Alloc)
	chain.Genesis.StateRoot = gr

	bchain, err := blockchain.NewBlockchain(hclog.Default(), "", chain, nil, executor)
	if err != nil {
		t.Fatal(err)
	}

	bchain.SetConsensus(nil)

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
