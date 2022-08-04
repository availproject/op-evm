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

	genesisAccount, testAccount := getAccounts(ks)

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

	// Transfer 1 ETH from genesis to test account and verify that validator received the amount

	log.Println("Sequencer -> initiating transfer of 1 ETH from genesis to test account...")

	transferTime := time.Now()

	tx, err := transferEth(
		sequencerClient,
		ks,
		genesisAccount,
		testAccount,
		ETH.Int64(), // Send 1 ETH
	)

	if err != nil {
		log.Fatalf("Sequencer -> failure to transfer 1 ETH from genesis account to test account: %s", err)
	}

	log.Printf("Sequencer -> genesis to test account 1 ETH transfer success! Tx hash: %v \n", tx.Hash().String())

	// First we are going to wait for sequencer client to report back that test account has 1 ETH
	// and genesis account balance-value (1000 ETH - 1 ETH) = 999 ETH
	// We're going to set the deadline for test failure

	// I've went this route to send and check. More proper one would be to await for the receipts
	// of the transactions. That is something we should consider in the future as it would be
	// more round-robbin test.

	transferTicker := time.NewTicker(5 * time.Second)
	transferDoneTicker := time.NewTicker(60 * time.Second)

	// We need to make sure we know what's the new balance we're targeting before the loop
	var genesisSequencerTransferBalance *big.Int
	var testSequencerTransferBalance *big.Int

	genesisSequencerNextBalance := toETH(new(big.Int).Sub(genesisSequencerCurrentBalance, ETH))
	testSequencerNextBalance := toETH(new(big.Int).Add(testSequencerCurrentBalance, ETH))

	log.Printf(
		"Sequencer -> Awaiting for balance confirmation. Target -> Genesis: %d | Test: %d",
		genesisSequencerNextBalance, testSequencerNextBalance,
	)

transferGoto:
	for {
		select {
		case <-transferDoneTicker.C:
			transferTicker.Stop()
			transferDoneTicker.Stop()
			log.Fatalf("Sequencer -> failure to receive transfer balance changes in 60s")
			os.Exit(1)
		case <-transferTicker.C:
			genesisSequencerTransferBalance, err = getAccountBalance(sequencerClient, genesisAccount.Address, nil)
			if err != nil {
				log.Fatalf("transfer: failure to get genesis account balance - sequencer: %s \n", err)
			}

			testSequencerTransferBalance, err = getAccountBalance(sequencerClient, testAccount.Address, nil)
			if err != nil {
				log.Fatalf("transfer: failure to get test account balance - sequencer: %s \n", err)
			}

			log.Printf(
				"Sequencer -> Ticker balance check -> Genesis Account: %d | Test Account: %d \n",
				toETH(genesisSequencerTransferBalance),
				toETH(testSequencerTransferBalance),
			)

			if toETH(genesisSequencerTransferBalance).Int64() == genesisSequencerNextBalance.Int64() &&
				toETH(testSequencerTransferBalance).Int64() == testSequencerNextBalance.Int64() {
				log.Printf(
					"Sequencer -> Balance transfer confirmation successful! Time took: %v \n",
					time.Since(transferTime),
				)
				transferTicker.Stop()
				transferDoneTicker.Stop()
				break transferGoto
			} else {
				transferTicker.Reset(5 * time.Second)
				log.Print("Sequencer -> Balances not matching yet... Waiting 5 seconds and rechecking...")
			}

		}
	}

	// Next is to check if the validator node as well has the proper balance set...

	log.Println("Validator -> Starting transfer confirmation check...")

	transferValidatorTicker := time.NewTicker(5 * time.Second)
	transferValidatorDoneTicker := time.NewTicker(60 * time.Second)

transferValidatorGoto:
	for {
		select {
		case <-transferValidatorDoneTicker.C:
			transferValidatorTicker.Stop()
			transferValidatorDoneTicker.Stop()
			log.Fatalf("Validator -> failure to receive transfer balance changes in 60s")
			os.Exit(1)
		case <-transferValidatorTicker.C:
			genesisValidatorTransferBalance, err := getAccountBalance(validatorClient, genesisAccount.Address, nil)
			if err != nil {
				log.Fatalf("validator transfer: failure to get genesis account balance - sequencer: %s \n", err)
			}

			testValidatorTransferBalance, err := getAccountBalance(validatorClient, testAccount.Address, nil)
			if err != nil {
				log.Fatalf("validator transfer: failure to get test account balance - sequencer: %s \n", err)
			}

			log.Printf(
				"Validator -> Ticker balance check -> Genesis Account: %d | Test Account: %d \n",
				toETH(genesisValidatorTransferBalance),
				toETH(testValidatorTransferBalance),
			)

			if toETH(genesisValidatorTransferBalance).Int64() == toETH(genesisSequencerTransferBalance).Int64() &&
				toETH(testValidatorTransferBalance).Int64() == toETH(testSequencerTransferBalance).Int64() {
				log.Printf(
					"Validator -> Balance transfer confirmation successful! Total time took: %v \n",
					time.Since(transferTime),
				)
				transferValidatorTicker.Stop()
				transferValidatorDoneTicker.Stop()
				log.Println("E2E BALANCE TEST SUCCESSFUL!")
				break transferValidatorGoto
			} else {
				transferValidatorTicker.Reset(5 * time.Second)
				log.Print("Validator -> Balances not matching yet... Waiting 5 seconds and rechecking...")
			}

		}
	}

}
