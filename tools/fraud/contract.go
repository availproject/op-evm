package main

import (
	"context"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	fraud "github.com/maticnetwork/avail-settlement/tools/fraud/contract"
)

func deployContract(client *ethclient.Client, ks *keystore.KeyStore, fromAccount accounts.Account) (*common.Address, *types.Transaction, error) {
	nonce, err := client.PendingNonceAt(context.Background(), fromAccount.Address)
	if err != nil {
		return nil, nil, err
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, nil, err
	}

	passpharse := "secret"
	err = ks.Unlock(fromAccount, passpharse)
	if err != nil {
		return nil, nil, err
	}

	keyjson, err := ks.Export(fromAccount, passpharse, passpharse)
	if err != nil {
		return nil, nil, err
	}

	privatekey, err := keystore.DecryptKey(keyjson, passpharse)
	if err != nil {
		return nil, nil, err
	}

	auth := bind.NewKeyedTransactor(privatekey.PrivateKey)
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)     // in wei
	auth.GasLimit = uint64(700000) // in units
	auth.GasPrice = gasPrice

	address, tx, _, err := fraud.DeployFraud(auth, client)
	if err != nil {
		return nil, nil, err
	}

	return &address, tx, nil
}

func writeToContract(client *ethclient.Client, chainID *big.Int, ks *keystore.KeyStore, fromAccount accounts.Account, instance *fraud.Fraud, val *big.Int) (*types.Transaction, error) {
	nonce, err := client.PendingNonceAt(context.Background(), fromAccount.Address)
	if err != nil {
		log.Fatal(err)
	}

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
