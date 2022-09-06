package avail

import (
	"testing"

	"github.com/0xPolygon/polygon-edge/blockchain"
	"github.com/0xPolygon/polygon-edge/types"
)

func TestWatchTower(t *testing.T) {
	b := blockchain.NewTestBlockchain(t, nil)

	fraudProofGenerated := false
	fpFn := func(block types.Block) FraudProof { fraudProofGenerated = true; return FraudProof{} }

	wt := watchTower{
		blockchain:   b,
		fraudproofFn: fpFn,
	}

	emptyBlock := types.Block{
		Header:       new(types.Header),
		Transactions: []*types.Transaction{},
		Uncles:       []*types.Header{},
	}

	bs := emptyBlock.MarshalRLP()

	err := wt.HandleData(bs)
	if err != nil {
		t.Fatal(err)
	}

	if !fraudProofGenerated {
		t.Fatal("fraud proof was not generated for invalid block")
	}
}
