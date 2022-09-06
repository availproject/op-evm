package main

import (
	"context"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

func transferEth(client *ethclient.Client, ks *keystore.KeyStore, fromAccount accounts.Account, toAccount accounts.Account, val int64) (*types.Transaction, error) {
	nonce, err := client.PendingNonceAt(context.Background(), fromAccount.Address)
	if err != nil {
		log.Fatal(err)
	}

	value := big.NewInt(val)  // in wei
	gasLimit := uint64(21000) // in units
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}

	var data []byte
	tx := types.NewTransaction(nonce, toAccount.Address, value, gasLimit, gasPrice, data)

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return nil, err
	}

	err = ks.Unlock(fromAccount, DefaultWalletPassphrase)
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

	// Lock account back again...
	err = ks.Lock(fromAccount.Address)
	if err != nil {
		return nil, err
	}

	return tx, nil
}
