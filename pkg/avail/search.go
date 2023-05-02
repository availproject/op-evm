package avail

import (
	"fmt"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

// SearchFunc is an interface to function that determines seek offset based on
// current Avail block. The function returns the next block number to continue
// search from, a boolean whether the searched block was found or an error if
// something went wrong.
type SearchFunc func(*types.SignedBlock) (int64, bool, error)

// SearchBlock
// What we really need is to figure out in which
func (c *client) SearchBlock(offset int64, searchFunc SearchFunc) (*types.SignedBlock, error) {
	// In case offset is zero, it means that we have new chain node and we need to sync it
	// from latest head in avail towards first block.
	if offset == 0 {
		header, err := c.api.RPC.Chain.GetHeaderLatest()
		if err != nil {
			return nil, err
		}
		offset = int64(header.Number)
	}

	blkHash, err := c.api.RPC.Chain.GetBlockHash(uint64(offset))
	if err != nil {
		return nil, err
	}

	blk, err := c.api.RPC.Chain.GetBlock(blkHash)
	if err != nil {
		return nil, err
	}

	offset, _, err = searchFunc(blk)
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

		offset, found, err = searchFunc(blk)
		if err != nil {
			return nil, err
		}

		if found {
			return blk, nil
		}
	}

	return nil, fmt.Errorf("can't find block")
}
