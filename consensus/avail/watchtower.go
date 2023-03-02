package avail

import (
	"strings"
	"time"

	"github.com/0xPolygon/polygon-edge/types"
	avail_types "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/maticnetwork/avail-settlement/consensus/avail/watchtower"
	"github.com/maticnetwork/avail-settlement/pkg/avail"
	"github.com/maticnetwork/avail-settlement/pkg/block"
	"github.com/maticnetwork/avail-settlement/pkg/staking"
)

func (d *Avail) runWatchTower(activeParticipantsQuerier staking.ActiveParticipants, myAccount accounts.Account, signKey *keystore.Key) {
	logger := d.logger.Named("watchtower")
	watchTower := watchtower.New(d.blockchain, d.executor, d.txpool, nil, logger, types.Address(myAccount.Address), signKey.PrivateKey)

	// Start watching HEAD from Avail.
	availBlockStream := avail.NewBlockStream(d.availClient, d.logger, 0)

	callIdx, err := avail.FindCallIndex(d.availClient)
	if err != nil {
		panic(err)
	}

	for {
		watchtowerStaked, sequencerError := activeParticipantsQuerier.Contains(d.minerAddr, staking.WatchTower)
		if sequencerError != nil {
			d.logger.Error("failed to check if my account is among active staked watchtowers. Retrying in few seconds...", "error", sequencerError)
			time.Sleep(3 * time.Second)
			continue
		}

		if !watchtowerStaked {
			d.logger.Warn("my account is not among active staked watchtower. Retrying in few seconds...", "address", d.minerAddr.String())
			time.Sleep(3 * time.Second)
			continue
		}

		// Stop P2P blockchain syncing and follow the blockstream only via Avail.
		d.syncer.Close()
		break
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
			// Periodically verify that we are staked, before proceeding with watchtower
			// logic. In the unexpected case of being slashed and dropping below the
			// required watchtower staking threshold, we must stop processing, because
			// otherwise we just get slashed more.
			watchtowerStaked, sequencerError := activeParticipantsQuerier.Contains(d.minerAddr, staking.WatchTower)
			if sequencerError != nil {
				d.logger.Error("failed to check if my account is among active staked watchtowers; cannot continue", "error", sequencerError)
				continue
			}

			if !watchtowerStaked {
				d.logger.Error("my account is not among active staked watchtower; cannot continue", "address", d.minerAddr.String())
				continue
			}

			blks, err := block.FromAvail(availBlk, d.availAppID, callIdx, d.logger)
			if err != nil {
				logger.Error("cannot extract Edge blocks from Avail block", "block_number", availBlk.Block.Header.Number, "error", err)
				continue
			}

		blksLoop:
			for _, blk := range blks {
				err = watchTower.CheckBlockFully(blk)
				if err != nil { //  || blk.Number() == 4  - test the fraud
					// TODO: We should implement something like SafeCheck() to not return errors that should not
					// result in creating fraud proofs for blocks/transactions that should not be checked.
					if err != nil {
						if strings.Contains(err.Error(), "does not belong to active sequencers") {
							continue blksLoop
						}
					}

					logger.Info("Block verification failed. constructing fraudproof", "block_number", blk.Header.Number, "block_hash", blk.Header.Hash, "error", err)

					// Skip processing of fraudproof block. It's not written to blockchain on sequencers
					// either.
					_, exists := block.GetExtraDataFraudProofTarget(blk.Header)
					if exists {
						continue
					}

					// Apply block into local blockchain, even if the block was "invalid/malicious", because
					// otherwise the local blockchain wouldn't be consistent with the one on sequencers.
					if err := watchTower.Apply(blk); err != nil {
						logger.Error("cannot apply block to blockchain prior constructing fraud proof", "block_number", blk.Header.Number, "block_hash", blk.Header.Hash, "error", err)
						continue blksLoop
					}

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

				_, exists := block.GetExtraDataFraudProofTarget(blk.Header)
				if !exists {
					err = watchTower.Apply(blk)
					if err != nil {
						logger.Error("cannot apply block to blockchain", "block_number", blk.Header.Number, "block_hash", blk.Header.Hash, "error", err)
					}
				}
			}
		}
	}
}
