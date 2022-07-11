package avail

import (
	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
	"github.com/centrifuge/go-substrate-rpc-client/v4/rpc/chain"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

// BlockDataHandler is a function type for a callback invoked on new block.
type BlockDataHandler interface {
	HandleData(bs []byte) error
	HandleError(err error)
}

// BlockDataWatcher provides an implementation that is watching for new blocks from
// Avail and filters extrinsics with embedded Batch data, invoking handler with
// the decoded Batch.
type BlockDataWatcher struct {
	appID   types.U32
	client  Client
	handler BlockDataHandler
	stop    chan struct{}
}

// NewBlockDataWatcher constructs and starts the watcher following Avail blocks.
func NewBlockDataWatcher(client Client, appID uint32, handler BlockDataHandler) (*BlockDataWatcher, error) {
	watcher := BlockDataWatcher{
		appID:   types.U32(appID),
		client:  client,
		handler: handler,
		stop:    make(chan struct{}),
	}

	err := watcher.start()
	if err != nil {
		return nil, err
	}

	return &watcher, nil
}

func (bw *BlockDataWatcher) start() error {
	api := bw.client.instance()

	sub, err := api.RPC.Chain.SubscribeNewHeads()
	if err != nil {
		return err
	}

	go bw.processBlocks(api, sub)

	return nil
}

func (bw *BlockDataWatcher) processBlocks(api *gsrpc.SubstrateAPI, sub *chain.NewHeadsSubscription) {
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

			for _, extrinsic := range availBatch.Block.Extrinsics {
				if extrinsic.Signature.AppID != bw.appID {
					continue
				}

				blob := &Blob{}
				err = types.DecodeFromBytes(extrinsic.Method.Args, &blob)
				if err != nil {
					// Don't invoke HandleError() on this because there is no
					// way of filtering uninteresting extrinsics / method.Args
					// and failing decoding is the only way to distinct those.
					continue
				}

				if blob.Magic != BlobMagic {
					// Don't invoke HandleError() on this because there is no
					// way of filtering uninteresting extrinsics / method.Args
					// and failing decoding is the only way to distinct those.
					continue
				}

				bw.handler.HandleData(blob.Data)
			}
		case err := <-sub.Err():
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
