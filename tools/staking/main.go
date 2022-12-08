package main

import (
	"log"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/common"
	golog "github.com/ipfs/go-log/v2"
	"github.com/maticnetwork/avail-settlement-contracts/staking/pkg/staking"
	commontoken "github.com/maticnetwork/avail-settlement/pkg/common"
)

// curl  http://127.0.0.1:30002 -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"txpool_content","params":[],"id":1}'
// curl  http://127.0.0.1:30002 -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"txpool_inspect","params":[],"id":1}'

const (
	SequencerAddr           = "http://127.0.0.1:10002"
	ValidatorAddr           = "http://127.0.0.1:20002"
	WalletsDir              = "./data/wallets"
	DefaultWalletPassphrase = "secret"
	GenesisAccountHex       = "0x064A4a5053F3de5eacF5E72A2E97D5F9CF55f031"
	TestAccountHex          = "0x65F0bDe66C970F391bd648B7ea22e1c193221c65"
	ContractAccountHex      = "0x137De958553BB76FdFD8D64a55E8fA466768FE6a"
)

var (
	StakingAddress = common.HexToAddress("0x0110000000000000000000000000000000000001")
	MinerAddress   = common.HexToAddress("0xF817d12e6933BbA48C14D4c992719B46aD9f5f61")
)

func toETH(wei *big.Int) *big.Int {
	return big.NewInt(0).Div(wei, commontoken.ETH)
}

// Test case description:
// - Transfer Eth from one account to another and check if balance matches between nodes
// - Deploy contract and check if responses matches between the nodes
func main() {
	golog.SetAllLoggers(golog.LevelDebug)

	sequencerClient, err := getSequencerClient()
	if err != nil {
		log.Fatalf("sequencer client error: %s \n", err)
	}

	// Keystore and necessary account extraction

	ks, err := getKeystore()
	if err != nil {
		log.Fatalf("failure to retrieve keystore: %s \n", err)
	}

	genesisAccount, _, _ := getAccounts(ks)

	genesisSequencerCurrentBalance, err := getAccountBalance(sequencerClient, genesisAccount.Address, nil)
	if err != nil {
		log.Fatalf("failure to get genesis account balance - sequencer: %s \n", err)
	}

	sequencerCurrentBalance, err := getAccountBalance(sequencerClient, MinerAddress, nil)
	if err != nil {
		log.Fatalf("failure to get miner account balance - sequencer: %s \n", err)
	}

	log.Printf(
		"Genesis Account Hex: %s | Sequencer(Miner) Account Hex: %s",
		genesisAccount.Address.Hex(),
		MinerAddress.Hex(),
	)

	log.Printf(
		"Balances -> Genesis: %d | Sequencer (Miner): %d \n",
		toETH(genesisSequencerCurrentBalance),
		toETH(sequencerCurrentBalance),
	)

	contract, err := staking.NewStaking(StakingAddress, sequencerClient)
	if err != nil {
		log.Fatalf("Contract -> Failure to build new contract due: %s \n", err)
	}

	isSequencer, err := contract.IsSequencer(nil, MinerAddress)
	if err != nil {
		log.Fatalf("Contract -> Failure to check if contract is sequencer: %s \n", err)
	}

	log.Printf("Staking Smart Contract -> Is miner address staked: %v", isSequencer)
	os.Exit(0)
}
