package avail

import (
	"github.com/availproject/op-evm/consensus/avail/validator"
	"github.com/availproject/op-evm/pkg/avail"
	avail_types "github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

// getNextAvailBlockNumber determines the next Avail block number to be processed.
// It starts from the first block if the current blockchain is new, otherwise,
// it searches for the block in the Avail chain that corresponds to the last block in
// the local chain. In case of any failure, it returns 0.
func (d *Avail) getNextAvailBlockNumber() uint64 {
	head := d.blockchain.Header()

	/// We have new blockchain. Allow syncing from last to 1st block
	if head.Number == 0 {
		return 1
	}

	callIdx, err := avail.FindCallIndex(d.availClient)
	if err != nil {
		return 0
	}

	blk, err := d.availClient.SearchBlock(0, d.syncFunc(int64(head.Number), callIdx))
	if err != nil {
		d.logger.Error("failure to sync node", "error", err)
		return 0
	}

	return uint64(blk.Block.Header.Number)
}

// syncNode synchronizes the local node with the Avail chain until it reaches
// the latest block. It fetches the latest block from the Avail chain and syncs
// the local node until it catches up to this block. In case of any error, it
// logs the error message and returns the error.
func (d *Avail) syncNode() (uint64, error) {
	hdr, err := d.availClient.GetLatestHeader()
	if err != nil {
		d.logger.Error("couldn't fetch latest block hash from Avail", "error", err)
		return 0, err
	}

	fn := func(blk *avail_types.SignedBlock) bool {
		// Stop the syncing when we are up to date with latest header.
		return hdr.Number == blk.Block.Header.Number
	}

	return d.syncNodeUntil(fn)
}

// syncNodeUntil synchronizes the local node with the Avail chain until a
// specified condition is met. It fetches the Avail blocks and validates
// and writes them to the local blockchain. It continues this process until
// the provided stopConditionFn function returns true. In case of any error,
// it returns the number of the next Avail block to be fetched along with the error.
func (d *Avail) syncNodeUntil(stopConditionFn func(blk *avail_types.SignedBlock) bool) (uint64, error) {
	availNextBlockNumber := d.getNextAvailBlockNumber()

	callIdx, err := avail.FindCallIndex(d.availClient)
	if err != nil {
		return availNextBlockNumber, err
	}

	fraudResolver := NewFraudResolver(d.logger, d.blockchain, d.executor, d.txpool, nil, nil, d.minerAddr, d.signKey, d.availSender, d.nodeType)
	validator := validator.New(d.blockchain, d.minerAddr, d.logger)

	// BlockStream watcher must be started after the staking is done. Otherwise
	// the stream is out-of-sync.
	availBlockStream := d.availClient.BlockStream(availNextBlockNumber)
	defer availBlockStream.Close()

	for {
		var blk *avail_types.SignedBlock

		select {
		case blk = <-availBlockStream.Chan():

		case <-d.closeCh:
			if err := d.stakingNode.UnStake(d.signKey); err != nil {
				d.logger.Error("failed to unstake the node", "error", err)
				return availNextBlockNumber, nil
			}
			return 0, nil
		}

		edgeBlks, err := avail.BlockFromAvail(blk, d.availAppID, callIdx, d.logger)
		if len(edgeBlks) == 0 && err != nil {
			if err != avail.ErrNoExtrinsicFound {
				d.logger.Warn("unexpected error while extracting OpEVM blocks from Avail block", "error", err)
				continue
			}
		}

		// Write down blocks received from avail to make sure we're synced before processing with the
		// fraud check or writing down new blocks...
		for _, edgeBlk := range edgeBlks {
			if !fraudResolver.IsFraudProofBlock(edgeBlk) {
				if err := validator.Check(edgeBlk); err == nil {
					if err := d.blockchain.WriteBlock(edgeBlk, d.nodeType.String()); err != nil {
						d.logger.Warn(
							"failed to write edge block received from avail",
							"edge_block_hash", edgeBlk.Hash(),
							"error", err,
						)
					}
				} else {
					d.logger.Warn(
						"failed to validate edge block received from avail",
						"edge_block_hash", edgeBlk.Hash(),
						"error", err,
					)
				}
			}
		}

		availNextBlockNumber = uint64(blk.Block.Header.Number)

		// Stop syncing when stopCondition is met.
		if stopConditionFn(blk) {
			break
		}
	}

	return availNextBlockNumber, nil
}

// syncFunc generates a function that, given an Avail block, calculates the
// offset to the target Edge block. The generated function finds the Edge blocks
// in the Avail block and determines the offset to the target Edge block for
// each Edge block found. It then returns the smallest offset and a boolean
// indicating whether the target Edge block was found in the Avail block. In
// case of any error during the process, it returns the error.
func (d *Avail) syncFunc(targetEdgeBlock int64, callIdx avail_types.CallIndex) avail.SearchFunc {
	return func(availBlk *avail_types.SignedBlock) (int64, bool, error) {
		blks, err := avail.BlockFromAvail(availBlk, d.availAppID, callIdx, d.logger)
		if err != nil && err != avail.ErrNoExtrinsicFound {
			return -1, false, err
		}

		if blks == nil || len(blks) < 1 {
			return -1, false, nil
		}
		availBlockNum := int64(availBlk.Block.Header.Number)

		// Compute Avail block offsets for all the Edge blocks we can find from the
		// current Avail block extrinsincs.
		offsets := []int64{}
		for _, blk := range blks {
			edgeBlockNum := int64(blk.Header.Number)

			switch {
			case edgeBlockNum > targetEdgeBlock:
				offsets = append(offsets, availBlockNum-(edgeBlockNum-targetEdgeBlock))
			case edgeBlockNum == targetEdgeBlock:
				return int64(availBlk.Block.Header.Number), true, nil
			case edgeBlockNum < targetEdgeBlock:
				offsets = append(offsets, availBlockNum+(targetEdgeBlock-edgeBlockNum))
			}
		}

		// Search the smallest offset from the offsets.
		var smallest int64
		for i := 0; i < len(offsets); i++ {
			// Did we find the targetEdgeBlock from this Avail block?
			if offsets[i] == 0 {
				return 0, true, nil
			}

			if abs(smallest) > abs(offsets[i]) || smallest == 0 {
				smallest = offsets[i]
			}
		}

		return smallest, false, nil
	}
}

// abs returns the absolute value of the given int64 value.
func abs(x int64) int64 {
	if x < 0 {
		return -x
	} else {
		return x
	}
}
