package avail

import (
	"sync/atomic"
	"time"

	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/hashicorp/go-hclog"
)

// pollInterval is the time to wait after latest head block has been
// processed and before checking for new block head from Avail.
var pollInterval = 15 * time.Second

type BlockStream struct {
	closed  *atomic.Bool
	closeCh chan struct{}

	dataCh chan *types.SignedBlock

	api    *gsrpc.SubstrateAPI
	logger hclog.Logger

	offset uint64
}

func NewBlockStream(client Client, logger hclog.Logger, appID uint32, offset uint64) *BlockStream {
	bs := &BlockStream{
		closed:  new(atomic.Bool),
		closeCh: make(chan struct{}),
		dataCh:  make(chan *types.SignedBlock),
		api:     client.instance(),
		logger:  logger.Named("blockstream"),
		offset:  offset,
	}

	go bs.watch()

	return bs
}

func (bs *BlockStream) Close() {
	if bs.closed.CompareAndSwap(false, true) {
		close(bs.closeCh)
	}
}

func (bs *BlockStream) Chan() <-chan *types.SignedBlock {
	return bs.dataCh
}

func (bs *BlockStream) watch() {
	var blockHash, previous types.Hash
	var headReached bool

	for {
		latest, err := bs.api.RPC.Chain.GetBlockHashLatest()
		if err != nil {
			bs.logger.Error("couldn't fetch latest block hash", "error", err)
			continue
		}

		if previous == latest {
			headReached = true

			// We have processed the latest head block. Time to sleep.
			time.Sleep(pollInterval)
			continue
		}

		if headReached || bs.offset == 0 {
			blockHash = latest
		} else {
			blockHash, err = bs.api.RPC.Chain.GetBlockHash(bs.offset)
			if err != nil {
				bs.logger.Error("couldn't fetch block hash for block", "block_number", bs.offset, "error", err)
				continue
			}
		}

		blk, err := bs.api.RPC.Chain.GetBlock(blockHash)
		if err != nil {
			bs.logger.Error("couldn't fetch block", "block_number", bs.offset, "block_hash", blockHash, "error", err)
			continue
		}

		select {
		case <-bs.closeCh:
			close(bs.dataCh)
			return

		case bs.dataCh <- blk:
		}

		bs.offset = uint64(blk.Block.Header.Number) + 1
		previous = blockHash
	}
}
