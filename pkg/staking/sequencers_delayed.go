package staking

import "github.com/0xPolygon/polygon-edge/types"

type cachingRandomizedActiveSequencersQuerier struct {
	rngSeedFn RandomSeedFn
	querier   ActiveSequencers

	lastSeed       int64
	lastSequencers []types.Address
}

// NewRandomizedActiveSequencersQuerier returns an implementation of
// `ActiveSequencers` that deterministically randomizes list of currently
// active sequencers. Given same number from `RandomSeedFn` and list of
// addresses from `ActiveSequencers`, the return value of `Get()` is the same.
func NewCachingRandomizedActiveSequencersQuerier(rngSeedFn RandomSeedFn, activeParticipants ActiveParticipants) ActiveSequencers {
	return &cachingRandomizedActiveSequencersQuerier{
		rngSeedFn: rngSeedFn,
		querier:   NewRandomizedActiveSequencersQuerier(rngSeedFn, activeParticipants),
	}
}

func (c *cachingRandomizedActiveSequencersQuerier) Get() ([]types.Address, error) {
	seed := c.rngSeedFn()

	if c.lastSeed == seed {
		return c.lastSequencers, nil
	}

	c.lastSeed = seed
	sequencers, err := c.querier.Get()
	if err != nil {
		return nil, err
	}

	c.lastSequencers = sequencers
	return sequencers, nil
}

func (c *cachingRandomizedActiveSequencersQuerier) Contains(addr types.Address) (bool, error) {
	seed := c.rngSeedFn()

	if c.lastSequencers == nil || c.lastSeed != seed {
		// Refresh cache.
		_, err := c.Get()
		if err != nil {
			return false, err
		}
	}

	for _, s := range c.lastSequencers {
		if s == addr {
			return true, nil
		}
	}

	return false, nil
}
