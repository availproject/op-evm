package avail

import (
	"errors"

	"github.com/0xPolygon/polygon-edge/types"
)

var (
	ErrNoSyncPeer = errors.New("no sync peer")
)

func (d *Avail) runSyncState() {
	if !d.isState(SyncState) {
		return
	}

	callInsertBlockHook := func(block *types.Block) bool {
		d.logger.Debug("syncing block %d", block.Number())
		d.txpool.ResetWithHeaders(block.Header)
		return false
	}

	switch d.nodeType {
	case Sequencer:
		d.setState(AcceptState)
	case Validator:
		d.setState(ValidateState)
	case WatchTower:
		d.setState(WatchTowerState)
	}

	if err := d.syncer.Sync(
		callInsertBlockHook,
	); err != nil {
		d.logger.Error("watch sync failed", "err", err)
	}

	/* 	var bestPeer *protocol.SyncPeer

	   	deadline := time.Now().Add(30 * time.Second)
	   	for bestPeer == nil && time.Now().Before(deadline) {
	   		bestPeer = d.syncer.BestPeer()
	   		if bestPeer == nil {
	   			// Wait for best peer to come online.
	   			time.Sleep(1 * time.Second)
	   		}
	   	}

	   	if bestPeer == nil {
	   		return
	   	}

	   	curHdr := d.blockchain.Header()
	   	if curHdr == nil {
	   		panic("blockchain is uninitialized")
	   	}

	   	if bestPeer.Number() <= curHdr.Number {
	   		d.logger.Debug("no need to sync")

	   		switch d.nodeType {
	   		case Sequencer:
	   			d.setState(AcceptState)
	   		case Validator:
	   			d.setState(ValidateState)
	   		case WatchTower:
	   			d.setState(WatchTowerState)
	   		}

	   		return
	   	}

	   	d.runFullSync(bestPeer) */
}

/* func (d *Avail) runFullSync(peer *protocol.SyncPeer) {
	newBlockHandler := func(b *types.Block) {
		d.logger.Debug("syncing block %d", b.Number())
	}

	err := d.syncer.BulkSyncWithPeer(peer, newBlockHandler)
	if err != nil {
		d.logger.Error("bulk sync failed: %s", err)
	}
}
*/
