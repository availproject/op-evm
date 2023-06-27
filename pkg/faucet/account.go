// Package faucet provides functionality to find and interact with a faucet account in a blockchain.
// It offers utilities for locating the faucet account with the highest balance and retrieving its private key.
// The faucet account is typically used for distributing tokens to users during testing or development phases.
package faucet

import (
	"crypto/ecdsa"
	"errors"

	"github.com/0xPolygon/polygon-edge/chain"
	"github.com/0xPolygon/polygon-edge/crypto"
	"github.com/0xPolygon/polygon-edge/types"
)

// ErrAccountNotFound is returned when the faucet account is not found.
var ErrAccountNotFound = errors.New("faucet account not found")

// FindAccount finds the faucet account in the given chain and returns its private key.
// It selects the wealthiest account with a valid private key present in the chain's genesis allocation.
// If no valid faucet account is found, it returns ErrAccountNotFound.
func FindAccount(c *chain.Chain) (*ecdsa.PrivateKey, error) {
	var found bool
	var wealthiestAccount types.Address

	// Iterate through the accounts in the chain's genesis allocation
	for addr, acc := range c.Genesis.Alloc {
		if len(acc.PrivateKey) > 0 {
			// Ensure that the account has a valid private key present
			_, err := crypto.BytesToECDSAPrivateKey(acc.PrivateKey)
			if err != nil {
				continue
			}

			// Find the wealthiest account to use as the faucet account
			if !found || c.Genesis.Alloc[wealthiestAccount].Balance.Cmp(acc.Balance) < 0 {
				found = true
				wealthiestAccount = addr
			}
		}
	}

	if !found {
		return nil, ErrAccountNotFound
	}

	// Convert the private key bytes to an ECDSA private key
	key, err := crypto.BytesToECDSAPrivateKey(c.Genesis.Alloc[wealthiestAccount].PrivateKey)
	if err != nil {
		return nil, err
	}

	return key, nil
}
