package avail

import (
	"fmt"
	"sync/atomic"

	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/hashicorp/go-hclog"
)

type BlockStream struct {
	closed  *atomic.Bool
	closeCh chan struct{}

	dataCh chan *types.SignedBlock

	api    *gsrpc.SubstrateAPI
	logger hclog.Logger

	offset uint64
}

func NewBlockStream(client Client, logger hclog.Logger, offset uint64) *BlockStream {
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
	// Do we need to catch up with HEAD first?
	if bs.offset > 0 {
		err := bs.catchUp()
		if err != nil {
			return
		}
	} else {
		bs.logger.Debug("bs.offset == 0; no need to catchUp")
	}

	for {
		subscription, err := bs.api.RPC.Chain.SubscribeNewHeads()
		if err != nil {
			bs.logger.Error("failed to subscribe to new heads", "error", err)
			return
		}

		for {
			var hdr types.Header
			select {
			case <-bs.closeCh:
				close(bs.dataCh)
				return

			case hdr = <-subscription.Chan():
				blockHash, err := bs.api.RPC.Chain.GetBlockHash(uint64(hdr.Number))
				if err != nil {
					bs.logger.Error("couldn't fetch block hash for block", "block_number", hdr.Number, "error", err)
					continue
				}

				blk, err := bs.api.RPC.Chain.GetBlock(blockHash)
				if err != nil {
					bs.logger.Error("couldn't fetch block", "block_number", hdr.Number, "block_hash", blockHash, "error", err)
					continue
				}

				select {
				case <-bs.closeCh:
					close(bs.dataCh)
					return
				case bs.dataCh <- blk:
				}

			case err = <-subscription.Err():
				bs.logger.Error("error in Avail's new heads subscription; restarting", "error", err)
				break
			}
		}
	}
}

func (bs *BlockStream) catchUp() error {
	var blockHash, previous types.Hash

	for {
		latest, err := bs.api.RPC.Chain.GetBlockHashLatest()
		if err != nil {
			bs.logger.Error("couldn't fetch latest block hash", "error", err)
			continue
		}

		// Have we reached the HEAD?
		if previous == latest {
			return nil
		}

		blockHash, err = bs.api.RPC.Chain.GetBlockHash(bs.offset)
		if err != nil {
			bs.logger.Error("couldn't fetch block hash for block", "block_number", bs.offset, "error", err)
			continue
		}

		blk, err := bs.api.RPC.Chain.GetBlock(blockHash)
		if err != nil {
			bs.logger.Error("couldn't fetch block", "block_number", bs.offset, "block_hash", blockHash, "error", err)
			continue
		}

		select {
		case <-bs.closeCh:
			close(bs.dataCh)
			return fmt.Errorf("stream closed")
		case bs.dataCh <- blk:
		}

		bs.offset = uint64(blk.Block.Header.Number) + 1
		previous = blockHash
	}
}
