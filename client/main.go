package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	golog "github.com/ipfs/go-log/v2"
	setget "github.com/maticnetwork/avail-settlement/contracts/setget"
)

var chainID = big.NewInt(100)

// curl  http://127.0.0.1:30002 -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"txpool_content","params":[],"id":1}'
// curl  http://127.0.0.1:30002 -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"txpool_inspect","params":[],"id":1}'

func getKeystoreAccounts() (*keystore.KeyStore, error) {
	ks := keystore.NewKeyStore("./data/wallets", keystore.StandardScryptN, keystore.StandardScryptP)
	return ks, nil
}

// nolint:unused
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

	log.Printf("Owner account balance: %v ETH", big.NewInt(0).Div(balanceTestAcc, eth))

	/* 	contractTx, err := deployContract(
	   		client,
	   		accKeystore,
	   		ownerAddress,
	   	)
	   	if err != nil {
	   		log.Fatal(err)
	   	}

	   	log.Printf("Owner account setget contract tx: %#v", contractTx) */

	contractAddress := common.HexToAddress("0xBEe483807d80fBC3c933CF35FEe2D7CfE2e9d973")
	// contractTxHash := common.HexToHash("0xdb3e7f0c4ad6cab919944de0929003edc653510b9909f5040d2d4a39f71bf918")

	contract, err := setget.NewSetget(contractAddress, client)
	if err != nil {
		log.Fatal(err)
	}

	getVal, getErr := contract.Get(nil)
	if getErr != nil {
		log.Fatal(err)
	}

	log.Printf("Owner -> Contract -> BEFORE SET -> Get() -> Response: %v", getVal)

	setTx, err := writeToContract(
		client,
		accKeystore,
		ownerAddress,
		contract,
		big.NewInt(0).Add(getVal, big.NewInt(100)),
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Owner -> Contract -> Set() -> Tx Response: %+v", setTx)

	time.Sleep(5 * time.Second)

	getVal, getErr = contract.Get(nil)
	if getErr != nil {
		log.Fatal(err)
	}

	log.Printf("Owner -> Contract -> AFTER SET -> Get() -> Response: %v", getVal)

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

	err = ks.Unlock(fromAccount, "secret")
	if err != nil {
		return nil, err
	}

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

// nolint:unused
func deployContract(client *ethclient.Client, ks *keystore.KeyStore, fromAccount accounts.Account) (*types.Transaction, error) {
	nonce, err := client.PendingNonceAt(context.Background(), fromAccount.Address)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Nonce: %v", nonce)

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}

	passpharse := "secret"
	err = ks.Unlock(fromAccount, passpharse)
	if err != nil {
		return nil, err
	}

	keyjson, err := ks.Export(fromAccount, passpharse, passpharse)
	if err != nil {
		return nil, err
	}

	privatekey, err := keystore.DecryptKey(keyjson, passpharse)
	if err != nil {
		return nil, err
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privatekey.PrivateKey, chainID)
	if err != nil {
		return nil, err
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)     // in wei
	auth.GasLimit = uint64(300000) // in units
	auth.GasPrice = gasPrice

	address, tx, _, err := setget.DeploySetget(auth, client)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(address.Hex())
	fmt.Println(tx.Hash().Hex())

	return tx, nil
}

func writeToContract(client *ethclient.Client, ks *keystore.KeyStore, fromAccount accounts.Account, instance *setget.Setget, val *big.Int) (*types.Transaction, error) {
	nonce, err := client.PendingNonceAt(context.Background(), fromAccount.Address)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Nonce: %v", nonce)

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}

	passpharse := "secret"
	err = ks.Unlock(fromAccount, passpharse)
	if err != nil {
		return nil, err
	}

	keyjson, err := ks.Export(fromAccount, passpharse, passpharse)
	if err != nil {
		return nil, err
	}

	privatekey, err := keystore.DecryptKey(keyjson, passpharse)
	if err != nil {
		return nil, err
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privatekey.PrivateKey, chainID)
	if err != nil {
		return nil, err
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)     // in wei
	auth.GasLimit = uint64(300000) // in units
	auth.GasPrice = gasPrice

	tx, err := instance.Set(auth, val)
	if err != nil {
		log.Fatal(err)
	}

	return tx, nil
}
