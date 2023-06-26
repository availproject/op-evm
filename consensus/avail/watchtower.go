package avail

import (
	"strings"

	"github.com/0xPolygon/polygon-edge/types"
	"github.com/availproject/op-evm/consensus/avail/watchtower"
	"github.com/availproject/op-evm/pkg/avail"
	"github.com/availproject/op-evm/pkg/block"
	"github.com/availproject/op-evm/pkg/staking"
	avail_types "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
)

// runWatchTower is a method of the Avail structure that continuously monitors
// and verifies the blockchain for the Avail system. It utilizes the watchtower concept
// for blockchain monitoring and fraud detection. It operates until the node is closed.
//
// activeParticipantsQuerier is used to determine the active participants in the network.
//
// currentNodeSyncIndex is the blockchain index from where to start watching the blocks.
//
// myAccount is the ethereum account of the node.
//
// signKey is the private key used for signing the transactions.
//
// This function panics if it fails to find the avail call index.
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
		select {
		case <-d.closeCh:
			if err := d.stakingNode.UnStake(signKey.PrivateKey); err != nil {
				d.logger.Error("failed to unstake the node", "error", err)
			}
			availBlockStream.Close()
			return
		case availBlk := <-availBlockStream.Chan():
			blks, err := avail.BlockFromAvail(availBlk, d.availAppID, callIdx, d.logger)
			if err != nil {
				logger.Error("cannot extract Edge blocks from Avail block", "block_number", availBlk.Block.Header.Number, "error", err)
				continue
			}

		blksLoop:
			for _, blk := range blks {
				// Regardless of if block is malicious or not, apply it to the chain
				if err := watchTower.Apply(blk); err != nil {
					logger.Error("cannot apply block to blockchain", "block_number", blk.Header.Number, "block_hash", blk.Header.Hash, "error", err)
					continue blksLoop
				}

				// Periodically verify that we are staked, before proceeding with watchtower
				// logic. In the unexpected case of being slashed and dropping below the
				// required watchtower staking threshold, we must stop processing, because
				// otherwise we just get slashed more.
				watchtowerStaked, sequencerError := activeParticipantsQuerier.Contains(d.minerAddr, staking.WatchTower)
				if sequencerError != nil {
					d.logger.Error("failed to check if my account is among active staked watchtowers; cannot continue", "error", sequencerError)
					continue blksLoop
				}

				if !watchtowerStaked {
					d.logger.Error("my account is not among active staked watchtower; cannot continue", "address", d.minerAddr.String())
					continue blksLoop
				}

				err = watchTower.Check(blk)
				if err != nil {
					// TODO: We should implement something like SafeCheck() to not return errors that should not
					// result in creating fraud proofs for blocks/transactions that should not be checked.
					if err != nil {
						if strings.Contains(err.Error(), "does not belong to active sequencers") {
							continue blksLoop
						}
					}

					// Skip processing of fraudproof block. It's not written to blockchain on sequencers either.
					_, exists := block.GetExtraDataFraudProofTarget(blk.Header)
					if exists {
						continue blksLoop
					}

					logger.Info("Block verification failed. constructing fraudproof", "block_number", blk.Header.Number, "block_hash", blk.Header.Hash, "error", err)

					fp, err := watchTower.ConstructFraudproof(blk)
					if err != nil {
						logger.Error("failed to construct fraudproof for block", "block_number", blk.Header.Number, "block_hash", blk.Header.Hash, "error", err)
						continue blksLoop
					}

					logger.Info("Submitting fraudproof", "block_hash", fp.Header.Hash)

					err = d.availSender.SendAndWaitForStatus(fp, avail_types.ExtrinsicStatus{IsInBlock: true})
					if err != nil {
						logger.Error("Submitting fraud proof to avail failed", "error", err)
						continue blksLoop
					}

					logger.Info("Submitted fraudproof", "block_number", fp.Header.Number, "block_hash", fp.Header.Hash, "txns", len(fp.Transactions))
					continue blksLoop
				}
			}
		}
	}
}
