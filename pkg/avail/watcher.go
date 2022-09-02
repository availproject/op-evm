package avail

import (
	"bytes"
	"log"

	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
	"github.com/centrifuge/go-substrate-rpc-client/v4/rpc/chain"
	"github.com/centrifuge/go-substrate-rpc-client/v4/scale"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

// BlockDataHandler is a function type for a callback invoked on new block.
type BlockDataHandler interface {
	HandleData(bs []byte) error
	HandleError(err error)
}

// BlockDataWatcher provides an implementation that is watching for new blocks from
// Avail and filters extrinsics with embedded `Blob` data, invoking handler with
// the decoded `Blob`.
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

	//err := watcher.start()
	//if err != nil {
	//	return nil, err
	//}

	return &watcher, nil
}

func (bw *BlockDataWatcher) Start() error {
	api := bw.client.instance()

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
				if extrinsic.Signature.AppID != types.U32(BridgeAppID) {
					log.Printf("block %d extrinsic %d: AppID doesn't match (%d vs. %d)", head.Number, i, extrinsic.Signature.AppID, BridgeAppID)
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
					err = types.DecodeFromBytes(extrinsic.Method.Args, &bs)
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
