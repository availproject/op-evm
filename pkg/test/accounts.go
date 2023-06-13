package test

import (
	"crypto/ecdsa"
	"crypto/rand"
	"math/big"
	"testing"
	"time"

	"github.com/0xPolygon/polygon-edge/chain"
	"github.com/0xPolygon/polygon-edge/consensus"
	edge_crypto "github.com/0xPolygon/polygon-edge/crypto"
	"github.com/0xPolygon/polygon-edge/state"
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/maticnetwork/avail-settlement/pkg/blockchain"
)

// NewAccount is a test helper function that creates a new account with a private key and returns its address and private key.
// It takes a pointer to a testing.T object as argument.
// This function should be used within tests to generate a new account.
// Example usage (within a test):
// address, privateKey := NewAccount(t)
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

// DepositBalance is a test helper function that deposits a specified balance to a given account on a blockchain.
// It takes a pointer to a testing.T object, receiver address, deposit amount, blockchain and executor as arguments.
// This function should be used within tests to make a deposit of tokens to an account.
// Example usage (within a test, assuming receiver, amount, blockchain and executor are already defined):
// DepositBalance(t, receiver, amount, blockchain, executor)
func DepositBalance(t *testing.T, receiver types.Address, amount *big.Int, blockchain *blockchain.Blockchain, executor *state.Executor) {
	t.Helper()

	parent := blockchain.Header()
	if parent == nil {
		t.Fatal("couldn't load header for HEAD block")
	}

	header := &types.Header{
		ParentHash: parent.Hash,
		Number:     parent.Number + 1,
		Miner:      receiver.Bytes(),
		Nonce:      types.Nonce{},
		GasLimit:   parent.GasLimit,
		Timestamp:  uint64(time.Now().Unix()),
	}

	transition, err := executor.BeginTxn(parent.StateRoot, header, receiver)
	if err != nil {
		t.Fatalf("failed to begin transition: %s", err)
	}

	err = transition.SetAccountDirectly(receiver, &chain.GenesisAccount{Balance: amount})
	if err != nil {
		t.Fatalf("failed to set account balance directly: %s", err)
	}

	// Commit the changes
	_, root := transition.Commit()

	// Update the header
	header.StateRoot = root
	header.GasUsed = transition.TotalGas()

	// Build the actual block
	// The header hash is computed inside `BuildBlock()`
	blk := consensus.BuildBlock(consensus.BuildBlockParams{
		Header:   header,
		Txns:     []*types.Transaction{},
		Receipts: transition.Receipts(),
	})

	// Compute the hash, this is only a provisional hash since the final one
	// is sealed after all the committed seals
	blk.Header.ComputeHash()

	err = blockchain.WriteBlock(blk, "test")
	if err != nil {
		t.Fatalf("failed to write balance transfer block: %s", err)
	}
}
