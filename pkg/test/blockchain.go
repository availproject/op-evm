package test

import (
	"encoding/json"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/0xPolygon/polygon-edge/blockchain"
	"github.com/0xPolygon/polygon-edge/chain"
	"github.com/0xPolygon/polygon-edge/crypto"
	"github.com/0xPolygon/polygon-edge/helper/hex"
	"github.com/0xPolygon/polygon-edge/state"
	itrie "github.com/0xPolygon/polygon-edge/state/immutable-trie"
	"github.com/0xPolygon/polygon-edge/txpool"
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/hashicorp/go-hclog"
	"github.com/maticnetwork/avail-settlement/pkg/common"
)

func NewBlockchain(t *testing.T, verifier blockchain.Verifier, basepath string) (*state.Executor, *blockchain.Blockchain) {
	chain := NewChain(t, basepath)
	executor := NewInMemExecutor(t, chain)

	gr, err := executor.WriteGenesis(chain.Genesis.Alloc, types.ZeroHash)
	if err != nil {
		t.Fatal(err)
	}

	chain.Genesis.StateRoot = gr

	// use the eip155 signer
	signer := crypto.NewEIP155Signer(chain.Params.Forks.At(0), uint64(chain.Params.ChainID))

	bchain, err := blockchain.NewBlockchain(hclog.Default(), "", chain, nil, executor, signer)
	if err != nil {
		t.Fatal(err)
	}

	bchain.SetConsensus(verifier)
	executor.GetHash = bchain.GetHashHelper

	if err := bchain.ComputeGenesis(); err != nil {
		t.Fatal(err)
	}

	return executor, bchain
}

func NewBlockchainWithTxPool(t *testing.T, chainSpec *chain.Chain, verifier blockchain.Verifier) (*state.Executor, *blockchain.Blockchain, *txpool.TxPool) {
	executor := NewInMemExecutor(t, chainSpec)

	gr, err := executor.WriteGenesis(chainSpec.Genesis.Alloc, types.ZeroHash)
	if err != nil {
		t.Fatal(err)
	}

	chainSpec.Genesis.StateRoot = gr

	// use the eip155 signer
	signer := crypto.NewEIP155Signer(chainSpec.Params.Forks.At(0), uint64(chainSpec.Params.ChainID))

	bchain, err := blockchain.NewBlockchain(hclog.Default(), "", chainSpec, nil, executor, signer)
	if err != nil {
		t.Fatal(err)
	}

	bchain.SetConsensus(verifier)
	executor.GetHash = bchain.GetHashHelper

	if err := bchain.ComputeGenesis(); err != nil {
		t.Fatal(err)
	}

	txPool, err := txpool.NewTxPool(
		hclog.Default(),
		chainSpec.Params.Forks.At(0),
		NewTxpoolHub(executor.State(), bchain),
		nil,
		nil,
		&txpool.Config{MaxSlots: 10, MaxAccountEnqueued: 100},
	)
	if err != nil {
		t.Fatal(err)
	}

	txPool.SetSigner(signer)
	txPool.Start()

	return executor, bchain, txPool
}

func NewInMemExecutor(t *testing.T, c *chain.Chain) *state.Executor {
	t.Helper()

	storage := itrie.NewMemoryStorage()
	st := itrie.NewState(storage)

	e := state.NewExecutor(c.Params, st, hclog.Default())

	return e
}

func getStakingContractBytecode(t *testing.T, basepath string) []byte {
	jsonFile, err := os.Open(filepath.Join(basepath, "configs/genesis.json"))
	if err != nil {
		t.Fatal(err)
	}

	byteValue, _ := io.ReadAll(jsonFile)

	var data map[string]interface{}

	if err := json.Unmarshal(byteValue, &data); err != nil {
		t.Fatal(err)
	}

	genesisData := data["genesis"].(map[string]interface{})
	allocData := genesisData["alloc"].(map[string]interface{})
	for addr, addrData := range allocData {
		if addr == "0x0110000000000000000000000000000000000001" {
			addrDataMap := addrData.(map[string]interface{})
			bytecode, err := hex.DecodeHex(addrDataMap["code"].(string))
			if err != nil {
				t.Fatal(err)
			}
			return bytecode
		}
	}
	return nil
}

func NewChain(t *testing.T, basepath string) *chain.Chain {
	balance := big.NewInt(0).Mul(big.NewInt(1000), common.ETH)
	scBytecode := getStakingContractBytecode(t, basepath)

	return &chain.Chain{
		Genesis: &chain.Genesis{
			Alloc: map[types.Address]*chain.GenesisAccount{
				types.StringToAddress("0x064A4a5053F3de5eacF5E72A2E97D5F9CF55f031"): {
					Balance: balance,
				},
				types.StringToAddress("0x0110000000000000000000000000000000000001"): {
					Code:    scBytecode,
					Balance: balance,
					Storage: map[types.Hash]types.Hash{
						types.StringToHash("0x0000000000000000000000000000000000000000000000000000000000000005"): types.StringToHash("0x0000000000000000000000000000000000000000000000000000000000000001"),
						types.StringToHash("0x0000000000000000000000000000000000000000000000000000000000000006"): types.StringToHash("0x000000000000000000000000000000000000000000000000000000000000000a"),
					},
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
