package avail

import (
	"fmt"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

// SearchFunc is an interface to function that determines seek offset based on current Avail block.
type SearchFunc func(*types.SignedBlock, uint64, types.CallIndex) (int, bool, error)

// SearchBlock
// What we really need is to figure out in which
func (c *client) SearchBlock(offset int, targetEdgeBlock uint64, searchFunc SearchFunc) (*types.SignedBlock, error) {
	meta, err := c.instance().RPC.State.GetMetadataLatest()
	if err != nil {
		return nil, err
	}

	callIdx, err := meta.FindCallIndex(CallSubmitData)
	if err != nil {
		return nil, err
	}

	// In case offset is zero, it means that we have new chain node and we need to sync it
	// from latest head in avail towards first block.
	if offset == 0 {
		header, err := c.api.RPC.Chain.GetHeaderLatest()
		if err != nil {
			return nil, err
		}
		offset = int(header.Number)
	}

	blkHash, err := c.api.RPC.Chain.GetBlockHash(uint64(offset))
	if err != nil {
		return nil, err
	}

	blk, err := c.api.RPC.Chain.GetBlock(blkHash)
	if err != nil {
		return nil, err
	}

	offset, _, err = searchFunc(blk, targetEdgeBlock, callIdx)
	if err != nil {
		return nil, err
	}

	var found bool

	for {
		if offset == 0 {
			return blk, nil
		}

		if offset < 0 && blk.Block.Header.Number <= 1 {
			break
		}

		fmt.Printf("NEXT BLOCK HASH: %d - OFFSET: %d \n", uint64(blk.Block.Header.Number)+uint64(offset), offset)

		blkHash, err := c.api.RPC.Chain.GetBlockHash(uint64(blk.Block.Header.Number) + uint64(offset))
		if err != nil {
			return nil, err
		}

		blk, err = c.api.RPC.Chain.GetBlock(blkHash)
		if err != nil {
			return nil, err
		}

		offset, found, err = searchFunc(blk, targetEdgeBlock, callIdx)
		if err != nil {
			return nil, err
		}

		if found {
			return blk, nil
		}
	}

	return nil, fmt.Errorf("can't find block")
}
