package avail

import (
	"testing"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
)

func TestStreamBlocksCorrectSequence(t *testing.T) {
	t.Skip("multi-sequencer benchmarks disabled in CI/CD due to lack of Avail")

	offset := 1
	availClient, err := NewClient("ws://127.0.0.1:9944/v1/json-rpc", hclog.Default())
	if err != nil {
		t.Fatal(err)
	}
	logger := hclog.New(&hclog.LoggerOptions{Name: "polygon", Level: hclog.Off})

	bc := newBlockStream(availClient, logger, uint64(offset))

	timeout := time.After(20 * time.Second)
	var blockSeq []uint64
loop:
	for {
		select {
		case b := <-bc.Chan():
			blockSeq = append(blockSeq, uint64(b.Block.Header.Number))
			t.Log("Block number", b.Block.Header.Number)
		case <-timeout:
			break loop
		}
	}
	var expectSequence []uint64
	for i := uint64(offset); i <= blockSeq[len(blockSeq)-1]; i++ {
		expectSequence = append(expectSequence, i)
	}

	assert.Equal(t, expectSequence, blockSeq)
}
