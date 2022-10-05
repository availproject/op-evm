package test

import (
	"crypto/ecdsa"
	"crypto/rand"
	"math/big"
	"testing"

	edge_crypto "github.com/0xPolygon/polygon-edge/crypto"
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/ethereum/go-ethereum/crypto"
)

var (
	ETH = big.NewInt(1000000000000000000)
)

func NewAccount(t *testing.T) (types.Address, *ecdsa.PrivateKey) {
	t.Helper()

	privateKey, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	pk := privateKey.Public().(*ecdsa.PublicKey)
	address := edge_crypto.PubKeyToAddress(pk)

	return address, privateKey
}
