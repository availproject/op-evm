package avail

import (
	"github.com/0xPolygon/polygon-edge/types"
	avail_types "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/maticnetwork/avail-settlement/consensus/avail/watchtower"
	"github.com/maticnetwork/avail-settlement/pkg/avail"
	"github.com/maticnetwork/avail-settlement/pkg/block"
	"github.com/maticnetwork/avail-settlement/pkg/staking"
)

func (d *Avail) runWatchTower(stakingNode staking.Node, myAccount accounts.Account, signKey *keystore.Key) {
	activeParticipantsQuerier := staking.NewActiveParticipantsQuerier(d.blockchain, d.executor, d.logger)
	logger := d.logger.Named("watchtower")
	watchTower := watchtower.New(d.blockchain, d.executor, d.txpool, logger, types.Address(myAccount.Address), signKey.PrivateKey)

	d.logger.Debug("ensuring watchtower staked")
	err := d.ensureStaked(activeParticipantsQuerier)
	if err != nil {
		d.logger.Error("error while ensuring sequencer staked", "error", err)
		return
	}

	d.logger.Debug("ensured watchtower staked")

	// Start watching HEAD from Avail.
	availBlockStream := avail.NewBlockStream(d.availClient, d.logger, 0)

	// Stop P2P blockchain syncing and follow the blockstream only via Avail.
	//d.syncer.Close()

	callIdx, err := avail.FindCallIndex(d.availClient)
	if err != nil {
		panic(err)
	}

	logger.Info("watchtower started")

	// TODO: Figure out where do we need state cycle and how to implement it.
	// Current version only starts the cycles for the future, doing nothing with it.
	for {
		var availBlk *avail_types.SignedBlock

		select {
		case <-d.closeCh:
			if err := stakingNode.UnStake(signKey.PrivateKey); err != nil {
				d.logger.Error("failed to unstake the node", "error", err)
			}
			availBlockStream.Close()
			return
		case availBlk = <-availBlockStream.Chan():
		}

		/* 		accBalance, err := avail.GetBalance(d.availClient, d.availAccount)
		   		if err != nil {
		   			panic(fmt.Sprintf("Balance failure: %s", err))
		   		}

		   		d.logger.Info("Current avail account", "balance", accBalance.Int64()) */

		blks, err := block.FromAvail(availBlk, d.availAppID, callIdx)
		if err != nil {
			logger.Error("cannot extract Edge block from Avail block", "block_number", availBlk.Block.Header.Number, "error", err)
			continue
		}

		fraudProof := false

	blksLoop:
		for _, blk := range blks {
			err = watchTower.Check(blk)
			if err != nil {
				logger.Info("block verification failed. constructing fraudproof", "block_number", blk.Header.Number, "block_hash", blk.Header.Hash, "error", err)

				fp, err := watchTower.ConstructFraudproof(blk)
				if err != nil {
					logger.Error("failed to construct fraudproof for block", "block_number", blk.Header.Number, "block_hash", blk.Header.Hash, "error", err)
					continue
				}

				logger.Info("submitting fraudproof", "block_hash", fp, "err", err)
				logger.Info("submitting fraudproof", "block_hash", fp.Header.Hash)

				err = d.availSender.SendAndWaitForStatus(fp, avail_types.ExtrinsicStatus{IsInBlock: true})
				if err != nil {
					logger.Error("submitting fraud proof to avail failed", "error", err)
					continue
				}

				logger.Info("submitted fraudproof", "block_number", fp.Header.Number, "block_hash", fp.Header.Hash, "txns", len(fp.Transactions))

				err = watchTower.Apply(fp)
				if err != nil {
					logger.Error("cannot apply fraud block to blockchain", "block_number", blk.Header.Number, "block_hash", blk.Header.Hash, "error", err)
				}
				fraudProof = true
				continue blksLoop
			}

			err = watchTower.Apply(blk)
			if err != nil {
				logger.Error("cannot apply block to blockchain", "block_number", blk.Header.Number, "block_hash", blk.Header.Hash, "error", err)
			}

			if fraudProof {
				// Basically sequencer will get into the fraud proof state and won't be discovered as block that follows
				// will be corrupted so once fraud proof is pushed, p
				return

			}
		}
	}
}
