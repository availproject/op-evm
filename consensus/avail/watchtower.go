package avail

import (
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	avail_types "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/maticnetwork/avail-settlement/consensus/avail/watchtower"
	"github.com/maticnetwork/avail-settlement/pkg/avail"
	"github.com/maticnetwork/avail-settlement/pkg/block"
)

func (d *Avail) runWatchTower(watchTowerAccount accounts.Account, watchTowerPK *keystore.Key) {
	availBlockStream := avail.NewBlockStream(d.availClient, d.logger, avail.BridgeAppID, 0)
	availSender := avail.NewSender(d.availClient, signature.TestKeyringPairAlice)
	watchTower := watchtower.New(d.blockchain, d.executor, types.Address(watchTowerAccount.Address), watchTowerPK.PrivateKey)

	callIdx, err := avail.FindCallIndex(d.availClient)
	if err != nil {
		panic(err)
	}

	d.logger.Info("watchtower started")

	// TODO: Figure out where do we need state cycle and how to implement it.
	// Current version only starts the cycles for the future, doing nothing with it.
	for {
		var availBlk *avail_types.SignedBlock

		select {
		case <-d.closeCh:
			availBlockStream.Close()
			return
		case availBlk = <-availBlockStream.Chan():
		}

		blk, err := block.FromAvail(availBlk, avail.BridgeAppID, callIdx)
		if err != nil {
			d.logger.Error("cannot extract Edge block from Avail block %d: %s", availBlk.Block.Header.Number, err)
			continue
		}

		err = watchTower.Check(blk)
		if err != nil {
			fp, err := watchTower.ConstructFraudproof(blk)
			if err != nil {
				d.logger.Error("failed to construct fraudproof for block %d/%q: %s", blk.Header.Number, blk.Header.Hash, err)
				continue
			}

			f := availSender.SubmitDataAndWaitForStatus(fp.MarshalRLP(), avail_types.ExtrinsicStatus{IsInBlock: true})
			go func() {
				if _, err := f.Result(); err != nil {
					d.logger.Error("submitting fraud proof to avail failed", err)
				}
			}()

			continue
		}

		err = watchTower.Apply(blk)
		if err != nil {
			d.logger.Error("cannot apply block %d/%q to blockchain: %s", blk.Header.Number, blk.Header.Hash, err)
		}
	}
}
