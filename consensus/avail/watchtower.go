package avail

import (
	"fmt"
	"strings"

	"github.com/0xPolygon/polygon-edge/types"
	avail_types "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/maticnetwork/avail-settlement/consensus/avail/watchtower"
	"github.com/maticnetwork/avail-settlement/pkg/avail"
	"github.com/maticnetwork/avail-settlement/pkg/block"
	"github.com/maticnetwork/avail-settlement/pkg/staking"
)

func (d *Avail) runWatchTower(activeParticipantsQuerier staking.ActiveParticipants, currentNodeSyncIndex uint64, myAccount accounts.Account, signKey *keystore.Key) {
	logger := d.logger.Named("watchtower")
	watchTower := watchtower.New(d.blockchain, d.executor, d.txpool, logger, types.Address(myAccount.Address), signKey.PrivateKey)

	// Start watching HEAD from Avail.
	availBlockStream := d.availClient.BlockStream(currentNodeSyncIndex)

	callIdx, err := avail.FindCallIndex(d.availClient)
	if err != nil {
		panic(err)
	}

	logger.Info("watchtower started")

	for {
		var blks []*types.Block
		var err error

		select {
		case availBlk := <-availBlockStream.Chan():
			blks, err = block.FromAvail(availBlk, d.availAppID, callIdx, d.logger)
			if err != nil {
				logger.Error("cannot extract Edge blocks from Avail block", "block_number", availBlk.Block.Header.Number, "error", err)
				continue
			}
			// `blks` processing continues below in the main for-loop body.
		case <-d.closeCh:
			err = d.stakingNode.UnStake(signKey.PrivateKey)
			if err != nil {
				d.logger.Error("failed to unstake the node", "error", err)
			}
			availBlockStream.Close()
			return
		}

		for _, blk := range blks {
			// Check if the block is a fraud proof block (that should be skipped).
			_, isFraudProofBlock := block.GetExtraDataFraudProofTarget(blk.Header)
			if isFraudProofBlock {
				logger.Info("skipping fraud proof block", "block_number", blk.Header.Number, "block_hash", blk.Header.Hash)
				continue
			}

			// Write sequencer generated block into the blockchain prior we continue
			// checking for staking status.
			err = watchTower.Apply(blk)
			if err != nil {
				logger.Error("cannot apply block to blockchain", "block_number", blk.Header.Number, "block_hash", blk.Header.Hash, "error", err)
				continue
			}

			// Periodically verify that we are staked, before proceeding with watchtower
			// logic. In the unexpected case of being slashed and dropping below the
			// required watchtower staking threshold, we must stop processing, because
			// otherwise we just get slashed more.
			watchtowerStaked, err := activeParticipantsQuerier.Contains(d.minerAddr, staking.WatchTower)
			if err != nil {
				d.logger.Error("failed to check if my account is among active staked watchtowers; cannot continue", "error", err)
				continue
			}

			if !watchtowerStaked {
				d.logger.Error("my account is not among active staked watchtower; cannot continue", "address", d.minerAddr.String())
				continue
			}

			refreshedBlk, found := d.blockchain.GetBlock(blk.Hash(), 0, true)
			if !found {
				d.logger.Error("failed to discover block prior fraud proof check", "hash", blk.Hash())
				continue
			}

			err = watchTower.Check(refreshedBlk)
			if err != nil { //  || blk.Number() == 4  - test the fraud
				// TODO: We should implement something like SafeCheck() to not return errors that should not
				// result in creating fraud proofs for blocks/transactions that should not be checked.
				if err != nil {
					if strings.Contains(err.Error(), "does not belong to active sequencers") {
						continue
					}
				}

				logger.Info("Block verification failed. constructing fraudproof", "block_number", refreshedBlk.Header.Number, "block_hash", refreshedBlk.Header.Hash, "error", err)

				// Ensure that fraud proof blocks aren't present in canonical blockchain.
				_, exists := block.GetExtraDataFraudProofTarget(refreshedBlk.Header)
				if exists {
					panic(fmt.Sprintf("fraud proof block (number: %d, hash: %q) written in blockchain; execution must not get here", refreshedBlk.Header.Number, refreshedBlk.Header.Hash))
				}

				fp, err := watchTower.ConstructFraudproof(refreshedBlk)
				if err != nil {
					logger.Error("failed to construct fraudproof for block", "block_number", refreshedBlk.Header.Number, "block_hash", refreshedBlk.Header.Hash, "error", err)
					continue
				}

				logger.Info("Submitting fraudproof", "block_hash", fp.Header.Hash)

				err = d.availSender.SendAndWaitForStatus(fp, avail_types.ExtrinsicStatus{IsInBlock: true})
				if err != nil {
					logger.Error("Submitting fraud proof to avail failed", "error", err)
					continue
				}

				logger.Info("Submitted fraudproof", "block_number", fp.Header.Number, "block_hash", fp.Header.Hash, "txns", len(fp.Transactions))
				continue
			}
		}
	}
}
