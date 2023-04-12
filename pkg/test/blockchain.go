package test

import (
	"encoding/json"
	"io"
	"math/big"
	"os"
	"path/filepath"

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

func NewBlockchain(verifier blockchain.Verifier, basepath string) (*state.Executor, *blockchain.Blockchain, error) {
	chain, err := NewChain(basepath)
	if err != nil {
		return nil, nil, err
	}

	executor := NewInMemExecutor(chain)

	gr, err := executor.WriteGenesis(chain.Genesis.Alloc, types.ZeroHash)
	if err != nil {
		return nil, nil, err
	}

	chain.Genesis.StateRoot = gr

	// use the eip155 signer
	signer := crypto.NewEIP155Signer(chain.Params.Forks.At(0), uint64(chain.Params.ChainID))

	bchain, err := blockchain.NewBlockchain(hclog.Default(), "", chain, nil, executor, signer)
	if err != nil {
		return nil, nil, err
	}

	bchain.SetConsensus(verifier)
	executor.GetHash = bchain.GetHashHelper

	if err := bchain.ComputeGenesis(); err != nil {
		return nil, nil, err
	}

	return executor, bchain, nil
}

func NewBlockchainWithTxPool(chainSpec *chain.Chain, verifier blockchain.Verifier) (*state.Executor, *blockchain.Blockchain, *txpool.TxPool, error) {
	executor := NewInMemExecutor(chainSpec)

	gr, err := executor.WriteGenesis(chainSpec.Genesis.Alloc, types.ZeroHash)
	if err != nil {
		return nil, nil, nil, err
	}

	chainSpec.Genesis.StateRoot = gr

	// use the eip155 signer
	signer := crypto.NewEIP155Signer(chainSpec.Params.Forks.At(0), uint64(chainSpec.Params.ChainID))

	bchain, err := blockchain.NewBlockchain(hclog.Default(), "", chainSpec, nil, executor, signer)
	if err != nil {
		return nil, nil, nil, err
	}

	bchain.SetConsensus(verifier)
	executor.GetHash = bchain.GetHashHelper

	if err := bchain.ComputeGenesis(); err != nil {
		return nil, nil, nil, err
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
		return nil, nil, nil, err
	}

	txPool.SetSigner(signer)
	txPool.Start()

	return executor, bchain, txPool, nil
}

func NewInMemExecutor(c *chain.Chain) *state.Executor {
	storage := itrie.NewMemoryStorage()
	st := itrie.NewState(storage)
	return state.NewExecutor(c.Params, st, hclog.Default())
}

func getStakingContractBytecode(basepath string) ([]byte, error) {
	jsonFile, err := os.Open(filepath.Join(basepath, "configs/genesis.json"))
	if err != nil {
		return nil, err
	}

	byteValue, _ := io.ReadAll(jsonFile)

	var data map[string]interface{}

	if err := json.Unmarshal(byteValue, &data); err != nil {
		return nil, err
	}

	genesisData := data["genesis"].(map[string]interface{})
	allocData := genesisData["alloc"].(map[string]interface{})
	for addr, addrData := range allocData {
		if addr == "0x0110000000000000000000000000000000000001" {
			addrDataMap := addrData.(map[string]interface{})
			bytecode, err := hex.DecodeHex(addrDataMap["code"].(string))
			if err != nil {
				return nil, err
			}

			return bytecode, nil
		}
	}
	return nil, nil
}

func NewChain(basepath string) (*chain.Chain, error) {
	balance := big.NewInt(0).Mul(big.NewInt(1000), common.ETH)
	scBytecode, err := getStakingContractBytecode(basepath)
	if err != nil {
		return nil, err
	}

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
	}, nil
}
