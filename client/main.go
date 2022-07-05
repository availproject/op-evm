package main

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	golog "github.com/ipfs/go-log/v2"
)

// curl  http://127.0.0.1:30002 -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"txpool_content","params":[],"id":1}'
// curl  http://127.0.0.1:30002 -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"txpool_inspect","params":[],"id":1}'

func getKeystoreAccounts() (*keystore.KeyStore, error) {
	ks := keystore.NewKeyStore("./data/wallets", keystore.StandardScryptN, keystore.StandardScryptP)
	return ks, nil
}

func createKs() {
	ks := keystore.NewKeyStore("./data/wallets", keystore.StandardScryptN, keystore.StandardScryptP)
	password := "secret"
	account, err := ks.NewAccount(password)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(account.Address.Hex())
}

func main() {
	golog.SetAllLoggers(golog.LevelDebug)

	client, err := ethclient.Dial("http://127.0.0.1:10002")
	if err != nil {
		log.Fatal(err)
	}

	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Got the header number: %s", header.Number.String())

	block, err := client.BlockByNumber(context.Background(), header.Number)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Block number: %v", block.Number().Uint64())
	log.Printf("Block time: %v", block.Time())
	log.Printf("Block difficulty: %v", block.Difficulty().Uint64())
	log.Printf("Block hash (hex): %v", block.Hash().Hex())
	log.Printf("Block transactions length: %v", len(block.Transactions()))

	accKeystore, err := getKeystoreAccounts()
	if err != nil {
		log.Fatal(err)
	}

	accounts := accKeystore.Accounts()
	genesisAddress := accounts[0]
	ownerAddress := accounts[1]

	log.Printf("Preminted account hex: %s", genesisAddress.Address.Hex())

	balanceBaseAcc, err := client.BalanceAt(context.Background(), genesisAddress.Address, nil)
	if err != nil {
		log.Fatal(err)
	}

	eth := big.NewInt(1000000000000000000)
	log.Printf("Preminted account balance: %v ETH", big.NewInt(0).Div(balanceBaseAcc, eth))

	log.Printf("Test account hex: %s", ownerAddress.Address.Hex())

	//for i := 0; i < 100; i++ {
	tx, err := transferEth(
		client,
		accKeystore,
		genesisAddress,
		ownerAddress,
		eth.Int64(), // Send 1 ETH
	)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Test transfer of 1 ETH. Tx hash: %v", tx.Hash().String())
	//}

	balanceTestAcc, err := client.BalanceAt(context.Background(), ownerAddress.Address, nil)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Test account balance: %v ETH", big.NewInt(0).Div(balanceTestAcc, eth))

}

func transferEth(client *ethclient.Client, ks *keystore.KeyStore, fromAccount accounts.Account, toAccount accounts.Account, val int64) (*types.Transaction, error) {
	nonce, err := client.PendingNonceAt(context.Background(), fromAccount.Address)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Nonce: %v", nonce)

	value := big.NewInt(val)  // in wei
	gasLimit := uint64(21000) // in units
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}

	var data []byte
	tx := types.NewTransaction(nonce, toAccount.Address, value, gasLimit, gasPrice, data)

	log.Printf("TX HASH: %v", tx.Hash().String())
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return nil, err
	}

	ks.Unlock(fromAccount, "secret")

	signedTx, err := ks.SignTx(fromAccount, tx, chainID)
	if err != nil {
		return nil, err
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return nil, err
	}

	return tx, nil
}
