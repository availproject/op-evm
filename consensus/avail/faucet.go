package avail

import (
	"strings"

	avail_types "github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

/*
	XXX:
	There's some kind of a race between the above `ensureAccountBalance()` and
	the one below in the `syncConditionFn` - if one of them are removed,
	the sequencer doesn't get balance deposit and therefore won't boot. If
	they are both enabled, there are errors about "already known tx" from
	the TxPool, which is completely understandable.

	The question is: Where is that race? What causes it and how can there be
	a correct synchronization between the bootstrap sequencer and a new ordinary
	sequencer node, booting online?
*/

type FaucetHelper struct {
	*Avail
}

// Node sync condition. The node's miner account must have at least
// `minBalance` tokens deposited and the syncer must have reached the Avail
// HEAD.
func (f *FaucetHelper) SyncConditionFn(blk *avail_types.SignedBlock) bool {
	hdr, err := f.availClient.GetLatestHeader()
	if err != nil {
		f.logger.Error("couldn't fetch latest block hash from Avail", "error", err)
		return false
	}

	if hdr.Number == blk.Block.Header.Number {
		accountBalance, err := f.GetAccountBalance(f.minerAddr)
		if err != nil && strings.HasPrefix(err.Error(), "state not found") {
			// No need to log this.
			return false
		} else if err != nil {
			f.logger.Error("failed to query miner account balance", "error", err)
			return false
		}

		// Sync until our deposit tx is through.
		if accountBalance.Cmp(minBalance) < 0 {
			err = f.ensureAccountBalance()
			if err != nil {
				f.logger.Error("failed to ensure account balance", "error", err)
			}
			return false
		}

		// Our miner account has enough funds to operate and we have reached Avail
		// HEAD. Sync complete.
		return true
	}

	return false
}
