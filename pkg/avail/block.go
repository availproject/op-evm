package avail

import (
	"sync/atomic"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

type DummyBlockSource struct {
	blockNumber atomic.Int32
}

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
