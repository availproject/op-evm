package block

import (
	"bytes"
	"errors"

	"github.com/0xPolygon/polygon-edge/types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/scale"
	avail_types "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/hashicorp/go-hclog"
	"github.com/maticnetwork/avail-settlement/pkg/avail"
)

const (
	// SourceAvail is the constant for Avail as a block source.
	SourceAvail = "Avail"

	// SourceWatchTower is the constant for watch tower as a block source.
	SourceWatchTower = "WatchTower"
)

var (
	// ErrNoExtrinsicFound is returned when FromAvail(avail_blk) cannot decode
	// working Edge block from Avail block's extrinsic data.
	ErrNoExtrinsicFound = errors.New("no compatible extrinsic found")
)

func FromAvail(avail_blk *avail_types.SignedBlock, appID avail_types.U32, callIdx avail_types.CallIndex, logger hclog.Logger) ([]*types.Block, error) {
	toReturn := []*types.Block{}

	for i, extrinsic := range avail_blk.Block.Extrinsics {
		if extrinsic.Signature.AppID != appID {
			logger.Debug("block extrinsic's  AppID doesn't match", "avail_block_number", avail_blk.Block.Header.Number, "extrinsic_index", i, "extrinsic_app_id", extrinsic.Signature.AppID, "filter_app_id", appID)
			continue
		}

		if extrinsic.Method.CallIndex != callIdx {
			logger.Debug("block extrinsic's Method.CallIndex doesn't match", "avail_block_number", avail_blk.Block.Header.Number, "extrinsic_index", i, "extrinsic_call_index", extrinsic.Method.CallIndex, "filter_call_index", callIdx)
			continue
		}

		var blob avail.Blob
		{
			// XXX: This decoding process is an inefficient hack to
			// workaround problem in the encoding pipeline from client
			// code to Avail server. See more information about this in
			// sender.SubmitData().
			var bs avail_types.Bytes
			err := avail_types.DecodeFromBytes(extrinsic.Method.Args, &bs)
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

		blk := types.Block{}
		if err := blk.UnmarshalRLP(blob.Data); err != nil {
			return nil, err
		}

		logger.Info("Received new edge block from avail.", "hash", blk.Header.Hash, "parent_hash", blk.Header.ParentHash)

		toReturn = append(toReturn, &blk)
	}

	if len(toReturn) == 0 {
		return nil, ErrNoExtrinsicFound
	}

	return toReturn, nil
}
