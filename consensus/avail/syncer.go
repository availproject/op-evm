package avail

import (
	avail_types "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/maticnetwork/avail-settlement/pkg/block"
)

func (d *Avail) syncNode() uint64 {
	head := d.blockchain.Header()

	/// We have new blockchain. Allow syncing from last to 1st block
	if head.Number == 0 {
		return 1
	}

	_, err := d.availClient.SearchBlock(0, d.syncFunc)
	if err != nil {
		d.logger.Error("failure to sync node", "error", err)
		return 0
	}

	return 0
}

func (d *Avail) syncFunc(availBlk *avail_types.SignedBlock, callIdx avail_types.CallIndex) (int, error) {
	head := d.blockchain.Header()

	blks, err := block.FromAvail(availBlk, d.availAppID, callIdx, d.logger)
	if err != nil && err != block.ErrNoExtrinsicFound {
		return -1, err
	}

	if blks == nil || len(blks) < 1 {
		return -1, nil
	}

	offset := blks[0].Number() - head.Number
	//fmt.Printf("SEARCH OFFSET: %d\n", offset)
	return -1 * int(offset), nil
}
