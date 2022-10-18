package staking

import (
	"bytes"
	"math/rand"
	"sort"

	"github.com/0xPolygon/polygon-edge/types"
)

// RandomSeedFn is a function interface to provide seed to deterministic RNG
// used to shuffle sequencer addresses.
type RandomSeedFn func() int64

type randomizedActiveSequencersQuerier struct {
	rngSeedFn RandomSeedFn
	querier   ActiveSequencers
}

// NewRandomizedActiveSequencersQuerier returns an implementation of
// `ActiveSequencers` that deterministically randomizes list of currently
// active sequencers. Given same number from `RandomSeedFn` and list of
// addresses from `ActiveSequencers`, the return value of `Get()` is the same.
func NewRandomizedActiveSequencersQuerier(rngSeedFn RandomSeedFn, activeSequencers ActiveSequencers) ActiveSequencers {
	return &randomizedActiveSequencersQuerier{
		rngSeedFn: rngSeedFn,
		querier:   activeSequencers,
	}
}

type addresses []types.Address

func (as addresses) Len() int           { return len(as) }
func (as addresses) Less(i, j int) bool { return bytes.Compare(as[i].Bytes(), as[j].Bytes()) < 0 }
func (as addresses) Swap(i, j int)      { tmp := as[i]; as[i] = as[j]; as[j] = tmp }

func (rasq *randomizedActiveSequencersQuerier) Get() ([]types.Address, error) {
	as, err := rasq.querier.Get()
	if err != nil {
		return nil, err
	}

	// First sort the Addresses.
	addrs := addresses(as)
	sort.Stable(addrs)

	// Now shuffle the Addresses, using blockchain head block number as rng seed.
	rng := rand.New(rand.NewSource(rasq.rngSeedFn()))
	rng.Shuffle(addrs.Len(), addrs.Swap)

	return addrs, nil
}

func (rasq *randomizedActiveSequencersQuerier) Contains(addr types.Address) (bool, error) {
	as, err := rasq.querier.Get()
	if err != nil {
		return false, err
	}

	for _, a := range as {
		if bytes.Equal(addr.Bytes(), a.Bytes()) {
			return true, nil
		}
	}

	return false, nil
}
