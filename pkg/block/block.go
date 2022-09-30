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

func FromAvail(avail_blk *avail_types.SignedBlock, appID uint32, callIdx avail_types.CallIndex) (*types.Block, error) {
	logger := hclog.Default().Named("block")

	for i, extrinsic := range avail_blk.Block.Extrinsics {
		if extrinsic.Signature.AppID != avail_types.U32(appID) {
			logger.Debug("block %d extrinsic %d: AppID doesn't match (%d vs. %d)", avail_blk.Block.Header.Number, i, extrinsic.Signature.AppID, appID)
			continue
		}

		if extrinsic.Method.CallIndex != callIdx {
			logger.Debug("block %d extrinsic %d: Method.CallIndex doesn't match (got %v, expected %v)", avail_blk.Block.Header.Number, i, extrinsic.Method.CallIndex, callIdx)
			continue
		}

		logger.Debug("block %d extrinsic %d: len(extrinsic.Method.Args): %d, extrinsic.Method.Args: '%v'", avail_blk.Block.Header.Number, i, len(extrinsic.Method.Args), extrinsic.Method.Args)

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
				logger.Debug("block %d extrinsic %d: decoding raw bytes from args failed: %s", avail_blk.Block.Header.Number, i, err)
				continue
			}

			decoder := scale.NewDecoder(bytes.NewBuffer(bs))
			err = blob.Decode(*decoder)
			if err != nil {
				// Don't return just yet because there is no way of filtering
				// uninteresting extrinsics / method.Args and failing decoding
				// is the only way to distinct those.
				logger.Debug("block %d extrinsic %d: decoding blob from bytes failed: %s", avail_blk.Block.Header.Number, i, err)
				continue
			}
		}

		blk := types.Block{}
		if err := blk.UnmarshalRLP(blob.Data); err != nil {
			return nil, err
		}

		return &blk, nil
	}

	return nil, ErrNoExtrinsicFound
}
