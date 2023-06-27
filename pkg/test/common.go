package test

import (
	"testing"

	"github.com/0xPolygon/polygon-edge/types"
	"github.com/availproject/op-evm/pkg/blockchain"
)

// GetHeadBlock retrieves the head block from the provided blockchain instance.
// If the head block is not available, it retrieves the genesis block.
// It may fail the test if it's unable to fetch the head block.
func GetHeadBlock(t *testing.T, blockchain *blockchain.Blockchain) *types.Block {
	var head *types.Block

	var headBlockHash types.Hash
	hdr := blockchain.Header()
	if hdr != nil {
		headBlockHash = hdr.Hash
	} else {
		headBlockHash = blockchain.Genesis()
	}

	var ok bool
	head, ok = blockchain.GetBlockByHash(headBlockHash, true)
	if !ok {
		t.Fatal("couldn't fetch head block from blockchain")
	}

	return head
}
