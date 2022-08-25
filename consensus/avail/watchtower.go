package avail

import (
	"fmt"
	"log"
	"sync"

	"github.com/0xPolygon/polygon-edge/blockchain"
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/maticnetwork/avail-settlement/pkg/avail"
)

type FraudProof struct {
	types.Block
}

type watchTower struct {
	blockchain   *blockchain.Blockchain
	fraudproofFn func(block types.Block) FraudProof
}

func (wt *watchTower) HandleData(bs []byte) error {
	log.Printf("block handler: received batch w/ %d bytes\n", len(bs))

	block := types.Block{}
	if err := block.UnmarshalRLP(bs); err != nil {
		return err
	}

	if err := wt.blockchain.VerifyFinalizedBlock(&block); err != nil {
		log.Printf("block %d (%q) cannot be verified: %s", block.Number(), block.Hash(), err)
		_ = wt.fraudproofFn(block)
		// TODO: Deal with fraudproof
		return nil
	}

	if err := wt.blockchain.WriteBlock(&block); err != nil {
		return fmt.Errorf("failed to write block while bulk syncing: %w", err)
	}

	log.Printf("Received block header: %+v \n", block.Header)
	log.Printf("Received block transactions: %+v \n", block.Transactions)

	return nil
}

func (wt *watchTower) HandleError(err error) {
	log.Printf("block handler: error %#v\n", err)
}

func (d *Avail) runWatchTower() {
	d.logger.Info("watch tower started")

	// consensus always starts in SyncState mode in case it needs
	// to sync with Avail and/or other nodes.
	d.setState(SyncState)

	handler := &watchTower{
		blockchain:   d.blockchain,
		fraudproofFn: func(blk types.Block) FraudProof { panic(blk) },
	}

	watcher, err := avail.NewBlockDataWatcher(d.availClient, avail.BridgeAppID, handler)
	if err != nil {
		return
	}

	defer watcher.Stop()

	// TODO: Following state machine drive loop was copied from validator. It
	// is unclear right now, if it's _really_ needed.

	var once sync.Once
	for {
		select {
		case <-d.closeCh:
			return
		default: // Default is here because we would block until we receive something in the closeCh
		}

		if d.isState(WatchTowerState) {
			once.Do(func() {
				err := watcher.Start()
				if err != nil {
					panic(err)
				}
			})
		}

		// Start the state machine loop
		d.runWatchTowerCycle()
	}
}

func (d *Avail) runWatchTowerCycle() {
	// Based on the current state, execute the corresponding section
	switch d.getState() {
	case AcceptState:
		d.runAcceptState()

	case ValidateState:
		d.runValidateState()

	case SyncState:
		d.runSyncState()

	case WatchTowerState:
		return
	}

}
