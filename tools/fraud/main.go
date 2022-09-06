package main

import (
	"context"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	golog "github.com/ipfs/go-log/v2"
	fraud "github.com/maticnetwork/avail-settlement/tools/fraud/contract"
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
	ETH = big.NewInt(1000000000000000000)
)

func getHeaderByNumber(client *ethclient.Client, number *big.Int) (*types.Header, error) {
	return client.HeaderByNumber(context.Background(), number)
}

func headerNumbersMatch(sequencer *types.Header, validator *types.Header) bool {
	return sequencer.Number.Int64() == validator.Number.Int64()
}

func toETH(wei *big.Int) *big.Int {
	return big.NewInt(0).Div(wei, ETH)
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

	validatorClient, err := getValidatorClient()
	if err != nil {
		log.Fatalf("validator client error: %s \n", err)
	}

	// Fetch the current headers
	currentSequencerHeader, err := getHeaderByNumber(sequencerClient, nil)
	if err != nil {
		log.Fatalf("sequencer client error - current header -: %s \n", err)
	}

	currentValidatorHeader, err := getHeaderByNumber(validatorClient, nil)
	if err != nil {
		log.Fatalf("validator client error - current header -: %s \n", err)
	}

	log.Printf(
		"Current Headers -> Sequencer: %d | Validator: %d | Synced: %v \n",
		currentSequencerHeader.Number.Int64(),
		currentValidatorHeader.Number.Int64(),
		headerNumbersMatch(currentSequencerHeader, currentValidatorHeader),
	)

	// Keystore and necessary account extraction

	ks, err := getKeystore()
	if err != nil {
		log.Fatalf("failure to retrieve keystore: %s \n", err)
	}

	genesisAccount, testAccount, ownerAccount := getAccounts(ks)

	log.Printf(
		"Genesis Account Hex: %s | Test Account Hex: %s",
		genesisAccount.Address.Hex(),
		testAccount.Address.Hex(),
	)

	// Fetch current account balances from sequencer and validator and check if they match

	genesisSequencerCurrentBalance, err := getAccountBalance(sequencerClient, genesisAccount.Address, nil)
	if err != nil {
		log.Fatalf("failure to get genesis account balance - sequencer: %s \n", err)
	}

	testSequencerCurrentBalance, err := getAccountBalance(sequencerClient, testAccount.Address, nil)
	if err != nil {
		log.Fatalf("failure to get test account balance - sequencer: %s \n", err)
	}

	log.Printf(
		"Sequencer Balances -> Genesis: %d | Test: %d \n",
		toETH(genesisSequencerCurrentBalance),
		toETH(testSequencerCurrentBalance),
	)

	genesisValidatorCurrentBalance, err := getAccountBalance(validatorClient, genesisAccount.Address, nil)
	if err != nil {
		log.Fatalf("failure to get genesis account balance - validator: %s \n", err)
	}

	testValidatorCurrentBalance, err := getAccountBalance(validatorClient, testAccount.Address, nil)
	if err != nil {
		log.Fatalf("failure to get test account balance - validator: %s \n", err)
	}

	log.Printf(
		"Validator Balances -> Genesis: %d | Test: %d \n",
		toETH(genesisValidatorCurrentBalance),
		toETH(testValidatorCurrentBalance),
	)

	if genesisSequencerCurrentBalance.Int64() == genesisValidatorCurrentBalance.Int64() &&
		testSequencerCurrentBalance.Int64() == testValidatorCurrentBalance.Int64() {
		log.Print("Initial balances are matching between sequencer and validator nodes!")
	} else {
		log.Fatal("Initial balances do not match between the sequencer and validator nodes!")
	}

	transferTime := time.Now()

	// Now as we have enough of the cash on the contract owner account we can process with
	// creation of the contract. For each test run, we are going to re-create the contract for now.
	// NOTE: Solidity contracts are not per default, upgradeable. We won't do that trick here!
	// More info about upgrades can be found here: https://docs.openzeppelin.com/contracts/4.x/upgradeable

	log.Println("Contract -> Deploying the fraud contract...")

	contractOwnerBalance, err := getAccountBalance(validatorClient, ownerAccount.Address, nil)
	if err != nil {
		log.Fatalf("Contract -> failure to get owner account balance - validator: %s \n", err)
	}

	log.Printf("Contract -> Current owner '%s' account balance: %d ETH", ownerAccount.Address, toETH(contractOwnerBalance))

	ownerTx, err := transferEth(
		sequencerClient,
		ks,
		genesisAccount,
		ownerAccount,
		ETH.Int64(), // Send 1 ETH
	)

	if err != nil {
		log.Fatalf("Contract -> failure to transfer 1 ETH from genesis account to owner account: %s", err)
	}

	log.Printf("Contract -> Genesis to owner account 1 ETH transfer success! Tx hash: %v \n", ownerTx.Hash().String())

	log.Println("Contract -> Starting transfer confirmation check...")

	ownerBalanceTransferTicker := time.NewTicker(5 * time.Second)
	ownerBalanceTransferDoneTicker := time.NewTicker(120 * time.Second)

ownerBalanceTransferGoto:
	for {
		select {
		case <-ownerBalanceTransferDoneTicker.C:
			ownerBalanceTransferTicker.Stop()
			ownerBalanceTransferDoneTicker.Stop()
			log.Fatalf("Contract -> failure to receive owner transfer balance changes in 60s")
			os.Exit(1)
		case <-ownerBalanceTransferTicker.C:
			ownerValidatorTransferBalance, err := getAccountBalance(validatorClient, ownerAccount.Address, nil)
			if err != nil {
				log.Fatalf("validator transfer: failure to get genesis account balance - sequencer: %s \n", err)
			}

			log.Printf(
				"Contract -> Check owner transfer balance update -> Current: %d | Wanted: %d \n",
				toETH(ownerValidatorTransferBalance),
				toETH(big.NewInt(0).Add(contractOwnerBalance, ETH)),
			)

			if toETH(ownerValidatorTransferBalance).Int64() == toETH(big.NewInt(0).Add(contractOwnerBalance, ETH)).Int64() {
				log.Printf(
					"Contract -> Balance transfer confirmation successful! Total time took: %v \n",
					time.Since(transferTime),
				)
				ownerBalanceTransferTicker.Stop()
				ownerBalanceTransferDoneTicker.Stop()
				log.Println("Chain -> CONTRACT OWNER BALANCE PROPAGATED SUCCESSFULLY!")
				break ownerBalanceTransferGoto
			} else {
				ownerBalanceTransferTicker.Reset(5 * time.Second)
				log.Print("Contract -> Balances not matching yet... Waiting 5 seconds and rechecking...")
			}

		}
	}

	chainID, _ := sequencerClient.ChainID(context.Background())

	contractAddress, _, err := deployContract(sequencerClient, chainID, ks, ownerAccount)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Contract -> Deployment transaction executed! Contract address (hex): %s", contractAddress)

	log.Println("Contract -> Checking if contract exists on validator node...")

	contractDeployedTicker := time.NewTicker(5 * time.Second)
	contractDeployedDoneTicker := time.NewTicker(120 * time.Second)

	// Going to check if code exists on validator node under deployed address. If it does
	// it means contract is deployed and ready for use...
contractDeployedGoto:
	for {
		select {
		case <-contractDeployedDoneTicker.C:
			contractDeployedTicker.Stop()
			contractDeployedDoneTicker.Stop()
			log.Fatalf("Contract -> failure to check contract validator existence in 120s")
			os.Exit(1)
		case <-contractDeployedTicker.C:
			byteCode, err := validatorClient.CodeAt(context.Background(), *contractAddress, nil)
			if err != nil {
				contractDeployedTicker.Reset(5 * time.Second)
				log.Println("Contract -> Code not yet present on validator... Rechecking in 5s... (CodeAt)")
				continue
			}

			if len(byteCode) < 1 {
				contractDeployedTicker.Reset(5 * time.Second)
				log.Println("Contract -> Code not yet present on validator... Rechecking in 5s... (EmptyBytecode)")
				continue
			}

			log.Printf("Contract -> Bytecode: %v \n", byteCode)

			contractDeployedTicker.Stop()
			contractDeployedDoneTicker.Stop()
			log.Println("Contract -> DEPLOYED SUCCESSFULLY!")
			break contractDeployedGoto

		}
	}

	contract, err := fraud.NewFraud(*contractAddress, sequencerClient)
	if err != nil {
		log.Fatalf("Contract -> Failure to build new contract due: %s \n", err)
	}

	getVal, getErr := contract.Get(nil)
	if getErr != nil {
		log.Fatalf("Contract -> Failure to get latest contract information: %s \n", getErr)
	}

	log.Printf("Contract -> BEFORE SET -> Get() -> Response: %v", getVal)

	setTx, err := writeToContract(
		sequencerClient,
		chainID,
		ks,
		ownerAccount,
		contract,
		big.NewInt(0).Add(getVal, big.NewInt(100)),
	)

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Contract -> Set(100) -> Tx Response: %+v", setTx)

	fraudContractSetTicker := time.NewTicker(5 * time.Second)
	fraudContractSetDoneTicker := time.NewTicker(120 * time.Second)

fraudContractSetGoto:
	for {
		select {
		case <-fraudContractSetDoneTicker.C:
			fraudContractSetTicker.Stop()
			fraudContractSetDoneTicker.Stop()
			log.Fatalf("Contract -> failure to receive owner transfer balance changes in 60s")
			os.Exit(1)
		case <-fraudContractSetTicker.C:
			getVal, getErr = contract.Get(nil)
			if getErr != nil {
				log.Fatalf("Contract -> Failure to fetch the Get() value: %s", err)
			}

			log.Printf("Contract -> Received Get() value: %d \n", getVal)

			if getVal.Int64() == big.NewInt(100).Int64() {
				fraudContractSetTicker.Stop()
				fraudContractSetDoneTicker.Stop()
				log.Println("Chain -> CONTRACT SET-GET TEST COMPLETED!")
				break fraudContractSetGoto
			} else {
				fraudContractSetTicker.Reset(5 * time.Second)
				log.Print("Contract -> Values not matching yet... Waiting 5 seconds and rechecking...")
			}

		}
	}

	os.Exit(0)
}
