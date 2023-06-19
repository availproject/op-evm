package avail

import (
	"sync/atomic"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
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
