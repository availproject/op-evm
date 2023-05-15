package faucet

import (
	"crypto/ecdsa"
	"errors"

	"github.com/0xPolygon/polygon-edge/chain"
	"github.com/0xPolygon/polygon-edge/crypto"
	"github.com/0xPolygon/polygon-edge/types"
)

var ErrAccountNotFound = errors.New("faucet account not found")

func FindAccount(c *chain.Chain) (*ecdsa.PrivateKey, error) {
	var found bool
	var wealthiestAccount types.Address

	for addr, acc := range c.Genesis.Alloc {
		if len(acc.PrivateKey) > 0 {
			// Ensure that the account has correct private key present.
			_, err := crypto.BytesToECDSAPrivateKey(acc.PrivateKey)
			if err != nil {
				continue
			}

			// Find the wealthiest account to steal tokens from :D
			if !found || c.Genesis.Alloc[wealthiestAccount].Balance.Cmp(acc.Balance) < 0 {
				found = true
				wealthiestAccount = addr
			}
		}
	}

	if !found {
		return nil, ErrAccountNotFound
	}

	key, err := crypto.BytesToECDSAPrivateKey(c.Genesis.Alloc[wealthiestAccount].PrivateKey)
	if err != nil {
		return nil, err
	}

	return key, nil
}
