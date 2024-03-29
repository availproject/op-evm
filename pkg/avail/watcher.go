package avail

import (
	"bytes"
	"log"

	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
	"github.com/centrifuge/go-substrate-rpc-client/v4/rpc/chain"
	"github.com/centrifuge/go-substrate-rpc-client/v4/scale"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types/codec"
)

// BlockDataHandler is an interface for handling Avail block data.
type BlockDataHandler interface {
	// HandleData is called when block data is received.
	HandleData(bs []byte) error

	// HandleError is called when an error occurs during block processing.
	HandleError(err error)
}

// BlockDataWatcher watches for new Avail blocks and filters extrinsics with embedded `Blob` data.
// It invokes the handler with the decoded `Blob`.
type BlockDataWatcher struct {
	appID   types.UCompact
	client  Client
	handler BlockDataHandler
	stop    chan struct{}
}

// NewBlockDataWatcher creates and starts a new BlockDataWatcher.
// It takes a client of type Client, an appID of type types.UCompact, and a handler of type BlockDataHandler.
// It returns a pointer to the BlockDataWatcher instance and an error if any.
func NewBlockDataWatcher(client Client, appID types.UCompact, handler BlockDataHandler) (*BlockDataWatcher, error) {
	watcher := BlockDataWatcher{
		appID:   appID,
		client:  client,
		handler: handler,
		stop:    make(chan struct{}),
	}
	return &watcher, nil
}

// Start starts the BlockDataWatcher and begins processing blocks.
// It returns an error if the watcher fails to start.
func (bw *BlockDataWatcher) Start() error {
	api, err := instance(bw.client)
	if err != nil {
		return err
	}

	meta, err := api.RPC.State.GetMetadataLatest()
	if err != nil {
		return err
	}

	callIdx, err := meta.FindCallIndex(CallSubmitData)
	if err != nil {
		return err
	}

	sub, err := api.RPC.Chain.SubscribeNewHeads()
	if err != nil {
		return err
	}

	go bw.processBlocks(api, callIdx, sub)

	return nil
}

// processBlocks listens for new block heads and filters extrinsics with embedded `Blob` data.
// It invokes the handler with the decoded `Blob` data.
func (bw *BlockDataWatcher) processBlocks(api *gsrpc.SubstrateAPI, callIdx types.CallIndex, sub *chain.NewHeadsSubscription) {
	defer sub.Unsubscribe()

	for {
		select {
		case head := <-sub.Chan():
			blockHash, err := api.RPC.Chain.GetBlockHash(uint64(head.Number))
			if err != nil {
				bw.handler.HandleError(err)
				return
			}

			availBatch, err := api.RPC.Chain.GetBlock(blockHash)
			if err != nil {
				bw.handler.HandleError(err)
				return
			}

			for i, extrinsic := range availBatch.Block.Extrinsics {
				if extrinsic.Signature.AppID.Int64() != bw.appID.Int64() {
					log.Printf("block %d extrinsic %d: AppID doesn't match (%d vs. %d)", head.Number, i, extrinsic.Signature.AppID.Int64(), bw.appID.Int64())
					continue
				}

				if extrinsic.Method.CallIndex != callIdx {
					log.Printf("block %d extrinsic %d: Method.CallIndex doesn't match (got %v, expected %v)", head.Number, i, extrinsic.Method.CallIndex, callIdx)
					continue
				}

				log.Printf("block %d extrinsic %d: len(extrinsic.Method.Args): %d, extrinsic.Method.Args: '%v'", head.Number, i, len(extrinsic.Method.Args), extrinsic.Method.Args)

				var blob Blob
				{
					// XXX: This decoding process is an inefficient hack to
					// workaround problem in the encoding pipeline from client
					// code to Avail server. See more information about this in
					// sender.SubmitData().
					var bs types.Bytes
					err = codec.Decode(extrinsic.Method.Args, &bs)
					if err != nil {
						// Don't invoke HandleError() on this because there is no
						// way of filtering uninteresting extrinsics / method.Args
						// and failing decoding is the only way to distinct those.
						log.Printf("block %d extrinsic %d: decoding raw bytes from args failed: %s", head.Number, i, err)
						continue
					}

					decoder := scale.NewDecoder(bytes.NewBuffer(bs))
					err = blob.Decode(*decoder)
					if err != nil {
						// Don't invoke HandleError() on this because there is no
						// way of filtering uninteresting extrinsics / method.Args
						// and failing decoding is the only way to distinct those.
						log.Printf("block %d extrinsic %d: decoding blob from bytes failed: %s", head.Number, i, err)
						continue
					}
				}

				err = bw.handler.HandleData(blob.Data)
				if err != nil {
					log.Printf("block %d extrinsic %d: data handler returned an error: %s", head.Number, i, err)
				}
			}
		case err := <-sub.Err():
			log.Printf("block watcher error: %s", err)
			bw.handler.HandleError(err)
		case <-bw.stop:
			return
		}
	}
}

// Stop stops active watcher.
func (bw *BlockDataWatcher) Stop() {
	select {
	case <-bw.stop:
		return
	default:
		close(bw.stop)
	}
}
