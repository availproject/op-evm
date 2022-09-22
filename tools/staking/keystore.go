package main

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

/* func createKeystore() {
	ks := keystore.NewKeyStore(WalletsDir, keystore.StandardScryptN, keystore.StandardScryptP)
	password := DefaultWalletPassphrase
	account, err := ks.NewAccount(password)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(account.Address.Hex())
}
*/

func getKeystore() (*keystore.KeyStore, error) {
	ks := keystore.NewKeyStore(WalletsDir, keystore.StandardScryptN, keystore.StandardScryptP)
	return ks, nil
}

func getAccounts(ks *keystore.KeyStore) (accounts.Account, accounts.Account, accounts.Account) {
	accounts := ks.Accounts()
	return accounts[0], accounts[1], accounts[2]
}

func getAccountBalance(client *ethclient.Client, address common.Address, blockNum *big.Int) (*big.Int, error) {
	return client.BalanceAt(context.Background(), address, blockNum)
}
