package avail

import (
	avail_types "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/maticnetwork/avail-settlement/pkg/block"
)

func (d *Avail) syncNode() uint64 {
	head := d.blockchain.Header()

	//panic(fmt.Sprintf("Block number: %d", head.Number))

	/// We have new blockchain. Allow syncing from last to 1st block
	if head.Number == 0 {
		return 1
	}

	blk, err := d.availClient.SearchBlock(0, head.Number, d.syncFunc)
	if err != nil {
		d.logger.Error("failure to sync node", "error", err)
		return 0
	}
	//panic(fmt.Sprintf("Block number: %d", uint64(blk.Block.Header.Number)))
	return uint64(blk.Block.Header.Number)
}

// Searches for the edge block in the Avail and returns back avail block for future catch up by the node
func (d *Avail) syncFunc(availBlk *avail_types.SignedBlock, targetEdgeBlock uint64, callIdx avail_types.CallIndex) (int, bool, error) {
	blks, err := block.FromAvail(availBlk, d.availAppID, callIdx, d.logger)
	if err != nil && err != block.ErrNoExtrinsicFound {
		return -1, false, err
	}

	if blks == nil || len(blks) < 1 {
		return -1, false, nil
	}

	for _, blk := range blks {
		if blk.Header.Number == targetEdgeBlock {
			return int(availBlk.Block.Header.Number), true, nil
		}
	}

	return -1, false, nil
}
