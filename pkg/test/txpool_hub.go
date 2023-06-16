package test

import (
	"fmt"
	"math/big"

	"github.com/0xPolygon/polygon-edge/state"
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/maticnetwork/avail-settlement/pkg/blockchain"
)

// txpoolHub is a struct that holds a state and a Blockchain instance.
// This struct implements various methods that interact with the blockchain.
type txpoolHub struct {
	state state.State
	*blockchain.Blockchain
}

// GetNonce returns the nonce of an account at a specific state root.
// root is the hash of the state root.
// addr is the address of the account.
// Returns 0 if an error occurs when creating a new snapshot or getting the account.
func (t *txpoolHub) GetNonce(root types.Hash, addr types.Address) uint64 {
	// TODO: Use a function that returns only Account
	snap, err := t.state.NewSnapshotAt(root)
	if err != nil {
		return 0
	}

	account, err := snap.GetAccount(addr)
	if err != nil {
		return 0
	}

	return account.Nonce
}

// GetBalance returns the balance of an account at a specific state root.
// root is the hash of the state root.
// addr is the address of the account.
// Returns a big integer representing the account's balance.
// An error is returned if it fails to get a snapshot for the state root or if the account does not exist.
func (t *txpoolHub) GetBalance(root types.Hash, addr types.Address) (*big.Int, error) {
	snap, err := t.state.NewSnapshotAt(root)
	if err != nil {
		return nil, fmt.Errorf("unable to get snapshot for root, %w", err)
	}

	account, err := snap.GetAccount(addr)
	if err != nil {
		return big.NewInt(0), err
	}

	return account.Balance, nil
}

// GetBlockByHash retrieves a block by its hash from the blockchain.
// h is the hash of the block.
// full is a boolean that determines whether to retrieve a full block or not.
// Returns the block corresponding to the hash and a boolean indicating the success of the operation.
func (t *txpoolHub) GetBlockByHash(h types.Hash, full bool) (*types.Block, bool) {
	return t.Blockchain.GetBlockByHash(h, full)
}

// Header returns the header of the latest block from the blockchain.
// Returns a pointer to the header of the latest block.
func (t *txpoolHub) Header() *types.Header {
	return t.Blockchain.Header()
}

// NewTxpoolHub creates a new txpoolHub instance with the provided state and blockchain.
// s is the current state of the blockchain.
// bc is the current blockchain instance.
// Returns a pointer to the newly created txpoolHub instance.
//
// Example usage:
//
//	func TestNewTxpoolHub(t *testing.T) {
//		state := ... // initialize state
//		bc := ... // initialize blockchain
//		hub := NewTxpoolHub(state, bc)
//		// hub now holds a new txpoolHub instance
//	}
func NewTxpoolHub(s state.State, bc *blockchain.Blockchain) *txpoolHub {
	return &txpoolHub{state: s, Blockchain: bc}
}
