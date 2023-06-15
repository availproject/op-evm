package test

import (
	"testing"

	"github.com/maticnetwork/avail-settlement/pkg/blockchain"
	"github.com/0xPolygon/polygon-edge/types"
)

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
