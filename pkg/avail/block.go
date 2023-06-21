package avail

import (
	"bytes"
	"errors"
	"sync/atomic"

	edge_types "github.com/0xPolygon/polygon-edge/types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/scale"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types/codec"
	"github.com/hashicorp/go-hclog"
)

// DummyBlockSource is a dummy block source that generates dummy blocks with incremented block numbers.
type DummyBlockSource struct {
	blockNumber atomic.Int32
}

// DummyBlock generates a dummy block with the specified application ID, call index, and extrinsics.
//
// Parameters:
//   - appID: The application ID for the block.
//   - callIdx: The call index for the block.
//   - extrinsics: Optional extrinsics to include in the block.
//
// Return:
//   - *types.SignedBlock: The generated dummy block.
func (dbs *DummyBlockSource) DummyBlock(appID types.UCompact, callIdx types.CallIndex, extrinsics ...types.Extrinsic) *types.SignedBlock {
	blockNum := dbs.blockNumber.Add(1)

	blk := &types.SignedBlock{
		Block: types.Block{
			Header: types.Header{
				Number: types.BlockNumber(blockNum),
			},
			Extrinsics: []types.Extrinsic{},
		},
	}

	for _, e := range extrinsics {
		e.Method.CallIndex = callIdx
		e.Signature.AppID = appID

		blk.Block.Extrinsics = append(blk.Block.Extrinsics, e)
	}

	return blk
}

// Error returned when no compatible extrinsic is found in Avail block's extrinsic data
var ErrNoExtrinsicFound = errors.New("no compatible extrinsic found")

// BlockFromAvail converts Avail blocks into Edge blocks.
// It takes an Avail block, appID, callIdx, and logger as parameters.
// It returns a slice of Edge blocks or an error if conversion fails.
func BlockFromAvail(avail_blk *types.SignedBlock, appID types.UCompact, callIdx types.CallIndex, logger hclog.Logger) ([]*edge_types.Block, error) {
	toReturn := []*edge_types.Block{}

	for i, extrinsic := range avail_blk.Block.Extrinsics {
		if extrinsic.Signature.AppID.Int64() != appID.Int64() {
			logger.Debug("block extrinsic's AppID doesn't match", "avail_block_number", avail_blk.Block.Header.Number, "extrinsic_index", i, "extrinsic_app_id", extrinsic.Signature.AppID, "filter_app_id", appID)
			continue
		}

		if extrinsic.Method.CallIndex != callIdx {
			logger.Debug("block extrinsic's Method.CallIndex doesn't match", "avail_block_number", avail_blk.Block.Header.Number, "extrinsic_index", i, "extrinsic_call_index", extrinsic.Method.CallIndex, "filter_call_index", callIdx)
			continue
		}

		var blob Blob
		{
			// XXX: This decoding process is an inefficient hack to
			// workaround problem in the encoding pipeline from client
			// code to Avail server. See more information about this in
			// sender.SubmitData().
			var bs types.Bytes
			err := codec.Decode(extrinsic.Method.Args, &bs)
			if err != nil {
				// Don't return just yet because there is no way of filtering
				// uninteresting extrinsics / method.Args and failing decoding
				// is the only way to distinct those.
				logger.Info("decoding block extrinsic's raw bytes from args failed", "avail_block_number", avail_blk.Block.Header.Number, "extrinsic_index", i, "error", err)
				continue
			}

			decoder := scale.NewDecoder(bytes.NewBuffer(bs))
			err = blob.Decode(*decoder)
			if err != nil {
				// Don't return just yet because there is no way of filtering
				// uninteresting extrinsics / method.Args and failing decoding
				// is the only way to distinct those.
				logger.Info("decoding blob from extrinsic data failed", "avail_block_number", avail_blk.Block.Header.Number, "extrinsic_index", i, "error", err)
				continue
			}
		}

		blk := edge_types.Block{}
		if err := blk.UnmarshalRLP(blob.Data); err != nil {
			return nil, err
		}

		logger.Info("Received new edge block from avail.", "hash", blk.Header.Hash, "parent_hash", blk.Header.ParentHash, "avail_block_number", blk.Header.Number)

		toReturn = append(toReturn, &blk)
	}

	if len(toReturn) == 0 {
		return nil, ErrNoExtrinsicFound
	}

	return toReturn, nil
}
