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
}
