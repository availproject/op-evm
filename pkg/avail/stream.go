package avail

import (
	"fmt"
	"sync/atomic"

	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/hashicorp/go-hclog"
)

// BlockStream represents a stream of Avail blocks.
type BlockStream interface {
	// Chan returns the channel on which the signed blocks are received.
	Chan() <-chan *types.SignedBlock

	// Close closes the block stream.
	Close()
}

// blockStream implements the BlockStream interface.
type blockStream struct {
	closed  *atomic.Bool
	closeCh chan struct{}
	dataCh  chan *types.SignedBlock
	api     *gsrpc.SubstrateAPI
	logger  hclog.Logger
	offset  uint64
}

// newBlockStream creates a new block stream.
// It takes a client of type Client, a logger of type hclog.Logger, and an offset of type uint64.
// It returns a BlockStream instance.
func newBlockStream(client Client, logger hclog.Logger, offset uint64) BlockStream {
	api, err := instance(client)
	if err != nil {
		panic("unsupported client in newBlockStream()")
	}

	bs := &blockStream{
		closed:  new(atomic.Bool),
		closeCh: make(chan struct{}),
		dataCh:  make(chan *types.SignedBlock),
		api:     api,
		logger:  logger.Named("blockstream"),
		offset:  offset,
	}

	go bs.watch()

	return bs
}

// Close closes the block stream.
func (bs *blockStream) Close() {
	if bs.closed.CompareAndSwap(false, true) {
		close(bs.closeCh)
	}
}

// Chan returns the channel on which the signed blocks are received.
func (bs *blockStream) Chan() <-chan *types.SignedBlock {
	return bs.dataCh
}

// watch continuously watches for new blocks and sends them to the data channel.
func (bs *blockStream) watch() {
	hdr, err := bs.api.RPC.Chain.GetHeaderLatest()
	if err != nil {
		bs.logger.Error("couldn't fetch latest block hash", "error", err)
		return
	}

	// Do we need to catch up with HEAD first?
	if bs.offset > 0 {
		err = bs.catchUp(bs.offset, uint64(hdr.Number))
		if err != nil {
			bs.logger.Error("unable to catch up!", "error", err)
			return
		}
	} else {
		bs.logger.Debug("bs.offset == 0; no need to catchUp")
	}

	latestBlockNumber := hdr.Number + 1
	for {
		subscription, err := bs.api.RPC.Chain.SubscribeNewHeads()
		if err != nil {
			bs.logger.Error("failed to subscribe to new heads", "error", err)
			return
		}

	receiveBlocksLoop:
		for {
			var hdr types.Header
			select {
			case <-bs.closeCh:
				close(bs.dataCh)
				return

			case hdr = <-subscription.Chan():
				switch {
				case hdr.Number < latestBlockNumber:
					// Omit blocks that were already streamed
					bs.logger.Debug("block already registered, skipping", "block_number", hdr.Number, "latestBlockNumber", latestBlockNumber)
					continue
				case hdr.Number > latestBlockNumber:
					// Do we need to catch up the last processed block
					// This can happen in two cases:
					// 1) The connection was interrupted for a while
					// 2) There was a delay when catching up with the offset
					err = bs.catchUp(uint64(latestBlockNumber), uint64(hdr.Number))
					if err != nil {
						bs.logger.Error("unable to catch up!", "error", err)
						return
					}
					latestBlockNumber = hdr.Number + 1
					continue
				}

				blockHash, err := bs.api.RPC.Chain.GetBlockHash(uint64(hdr.Number))
				if err != nil {
					bs.logger.Error("couldn't fetch block hash for block", "block_number", hdr.Number, "error", err)
					continue
				}

				bs.logger.Info("Received new avail block", "nbr", hdr.Number, "hash", blockHash.Hex())

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
					latestBlockNumber = hdr.Number + 1
				}

			case err = <-subscription.Err():
				bs.logger.Error("error in Avail's new heads subscription; restarting", "error", err)
				break receiveBlocksLoop
			}
		}
	}
}

// catchUp catches up the blocks from the given offset to the given target offset.
// It returns an error if the catch-up fails.
func (bs *blockStream) catchUp(fromOffset, toOffset uint64) (err error) {
	// Have we reached the HEAD?
	for i := fromOffset; i <= toOffset; i++ {
		blockHash, err := bs.api.RPC.Chain.GetBlockHash(i)
		if err != nil {
			bs.logger.Error("couldn't fetch block hash for block", "block_number", i, "error", err)
			continue
		}

		blk, err := bs.api.RPC.Chain.GetBlock(blockHash)
		if err != nil {
			bs.logger.Error("couldn't fetch block", "block_number", i, "block_hash", blockHash, "error", err)
			continue
		}

		select {
		case <-bs.closeCh:
			close(bs.dataCh)
			return fmt.Errorf("stream closed")
		case bs.dataCh <- blk:
		}
	}
	return nil
}
