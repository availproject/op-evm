package avail

import (
	"fmt"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

// SearchFunc is an interface to function that determines seek offset based on current Avail block.
type SearchFunc func(*types.SignedBlock, types.CallIndex) (int, error)

// SearchBlock
// -
func (c *client) SearchBlock(offset int, searchFunc SearchFunc) (*types.SignedBlock, error) {
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
		fmt.Print("HERE ERROR GBH")
		return nil, err
	}

	blk, err := c.api.RPC.Chain.GetBlock(blkHash)
	if err != nil {
		return nil, err
	}

	fmt.Printf("GOT BLOCK %+v \n", blk)

	offset, err = searchFunc(blk, callIdx)
	if err != nil {
		return nil, err
	}

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
			fmt.Print("HERE ERROR GBH L \n")
			return nil, err
		}

		blk, err = c.api.RPC.Chain.GetBlock(blkHash)
		if err != nil {
			return nil, err
		}

		fmt.Printf("Offset: %d \n", blk.Block.Header.Number)

		offset, err = searchFunc(blk, callIdx)
		if err != nil {
			return nil, err
		}
	}

	return nil, fmt.Errorf("can't find block")
}
