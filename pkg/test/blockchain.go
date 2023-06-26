package test

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/json"
	"io"
	"math/big"
	"os"
	"path/filepath"

	"github.com/0xPolygon/polygon-edge/blockchain/storage/memory"
	"github.com/0xPolygon/polygon-edge/chain"
	edgechain "github.com/0xPolygon/polygon-edge/chain"
	"github.com/0xPolygon/polygon-edge/crypto"
	"github.com/0xPolygon/polygon-edge/helper/hex"
	"github.com/0xPolygon/polygon-edge/state"
	itrie "github.com/0xPolygon/polygon-edge/state/immutable-trie"
	"github.com/0xPolygon/polygon-edge/txpool"
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/availproject/op-evm/pkg/blockchain"
	"github.com/availproject/op-evm/pkg/common"
	geth_crypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/hashicorp/go-hclog"
)

// FaucetAccount and FaucetSignKey are used as a pair for an account with a plentiful balance for test purposes.
var (
	FaucetSignKey *ecdsa.PrivateKey = GenFaucetSignKey()
	FaucetAccount types.Address     = GetAccountFromPrivateKey(FaucetSignKey)
)

// GenFaucetSignKey generates a new ECDSA private key which is used for the FaucetAccount.
// It panics if the key generation fails.
func GenFaucetSignKey() *ecdsa.PrivateKey {
	key, err := ecdsa.GenerateKey(geth_crypto.S256(), rand.Reader)
	if err != nil {
		panic(err)
	}

	return key
}

// GetAccountFromPrivateKey takes an ECDSA private key and returns the associated account address.
func GetAccountFromPrivateKey(key *ecdsa.PrivateKey) types.Address {
	pk := key.Public().(*ecdsa.PublicKey)
	return crypto.PubKeyToAddress(pk)
}

// NewBlockchain creates a new in-memory blockchain with a specified verifier and basepath.
// It returns an executor, a blockchain, and an error if any occurred during the initialization.
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

	// Use the london signer with eip-155 as a fallback one
	var signer crypto.TxSigner = crypto.NewLondonSigner(
		uint64(chain.Params.ChainID),
		chain.Params.Forks.IsActive(edgechain.Homestead, 0),
		crypto.NewEIP155Signer(
			uint64(chain.Params.ChainID),
			chain.Params.Forks.IsActive(edgechain.Homestead, 0),
		),
	)

	db, err := memory.NewMemoryStorage(nil)
	if err != nil {
		return nil, nil, err
	}

	bchain, err := blockchain.NewBlockchain(hclog.Default(), db, chain, nil, executor, signer)
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

// NewBlockchainWithTxPool creates a new in-memory blockchain with a specified chain specification and verifier.
// It also initializes a transaction pool with default parameters.
// It returns an executor, a blockchain, a transaction pool, and an error if any occurred during the initialization.
func NewBlockchainWithTxPool(chainSpec *chain.Chain, verifier blockchain.Verifier) (*state.Executor, *blockchain.Blockchain, *txpool.TxPool, error) {
	executor := NewInMemExecutor(chainSpec)

	gr, err := executor.WriteGenesis(chainSpec.Genesis.Alloc, types.ZeroHash)
	if err != nil {
		return nil, nil, nil, err
	}

	chainSpec.Genesis.StateRoot = gr

	// Use the london signer with eip-155 as a fallback one
	var signer crypto.TxSigner = crypto.NewLondonSigner(
		uint64(chainSpec.Params.ChainID),
		chainSpec.Params.Forks.IsActive(edgechain.Homestead, 0),
		crypto.NewEIP155Signer(
			uint64(chainSpec.Params.ChainID),
			chainSpec.Params.Forks.IsActive(edgechain.Homestead, 0),
		),
	)

	db, err := memory.NewMemoryStorage(nil)
	if err != nil {
		return nil, nil, nil, err
	}

	bchain, err := blockchain.NewBlockchain(hclog.Default(), db, chainSpec, nil, executor, signer)
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

// NewInMemExecutor creates a new executor with an in-memory state and returns it.
func NewInMemExecutor(c *chain.Chain) *state.Executor {
	storage := itrie.NewMemoryStorage()
	st := itrie.NewState(storage)
	return state.NewExecutor(c.Params, st, hclog.Default())
}

// getStakingContractBytecode retrieves the bytecode of the staking contract from a genesis.json file located at the given basepath.
// It returns the bytecode as a byte slice and an error if any occurred during the retrieval.
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

// NewChain creates a new Chain instance with a predefined genesis account (FaucetAccount) and staking contract.
// The function takes a basepath as an argument, which is used to locate the staking contract's bytecode.
// It returns a Chain instance and an error if any occurred during the initialization.
func NewChain(basepath string) (*chain.Chain, error) {
	balance := big.NewInt(0).Mul(big.NewInt(10000), common.ETH)
	scBytecode, err := getStakingContractBytecode(basepath)
	if err != nil {
		return nil, err
	}

	return &chain.Chain{
		Genesis: &chain.Genesis{
			Alloc: map[types.Address]*chain.GenesisAccount{
				FaucetAccount: {
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
			Forks:   chain.AllForksEnabled,
			ChainID: 100,
			Engine: map[string]interface{}{
				"avail": map[string]interface{}{
					"mechanisms": []string{"sequencer", "validator"},
				},
			},
			BurnContract: map[uint64]string{
				0: "0x0000000000000000000000000000000000000000",
			},
		},
	}, nil
}
