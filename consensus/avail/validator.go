package avail

import (
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/maticnetwork/avail-settlement/consensus/avail/validator"
	"github.com/maticnetwork/avail-settlement/pkg/avail"
	"github.com/maticnetwork/avail-settlement/pkg/block"

	avail_types "github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

type ValidatorSet []types.Address

func (d *Avail) runValidator() {
	availBlockStream := avail.NewBlockStream(d.availClient, d.logger, 1)
	validator := validator.New(d.blockchain, d.executor, types.StringToAddress(SequencerAddress))

	callIdx, err := avail.FindCallIndex(d.availClient)
	if err != nil {
		panic(err)
	}

	d.logger.Info("validator started")

	// TODO: Figure out where do we need state cycle and how to implement it.
	// Current version only starts the cycles for the future, doing nothing with it.
	for {
		var avail_blk *avail_types.SignedBlock

		select {
		case <-d.closeCh:
			availBlockStream.Close()
			return
		case avail_blk = <-availBlockStream.Chan():
		}

		blks, err := block.FromAvail(avail_blk, d.availAppID, callIdx, d.logger)
		if err != nil {
			d.logger.Error("cannot extract Edge block from Avail block", "avail_block_number", avail_blk.Block.Header.Number, "error", err)
			continue
		}

		for _, blk := range blks {
			err = validator.Check(blk)
			if err != nil {
				d.logger.Error("invalid block", "block_number", blk.Header.Number, "block_hash", blk.Header.Hash, "error", err)
				continue
			}

			err = validator.Apply(blk)
			if err != nil {
				d.logger.Error("cannot apply block to blockchain", "block_nubmer", blk.Header.Number, "block_hash", blk.Header.Hash, "error", err)
			}
		}
	}
}
