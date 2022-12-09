package test

import (
	"fmt"
	"math/big"

	"github.com/0xPolygon/polygon-edge/blockchain"
	"github.com/0xPolygon/polygon-edge/state"
	"github.com/0xPolygon/polygon-edge/types"
)

type txpoolHub struct {
	state state.State
	*blockchain.Blockchain
}

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

func NewTxpoolHub(s state.State, bc *blockchain.Blockchain) *txpoolHub {
	return &txpoolHub{state: s, Blockchain: bc}
}
