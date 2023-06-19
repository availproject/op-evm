package staking

import "github.com/0xPolygon/polygon-edge/types"

// cachingRandomizedActiveSequencersQuerier is an implementation of the ActiveSequencers interface
// that provides deterministic randomization of the list of currently active sequencers.
type cachingRandomizedActiveSequencersQuerier struct {
	rngSeedFn      RandomSeedFn
	querier        ActiveSequencers
	lastSeed       int64
	lastSequencers []types.Address
}

// NewCachingRandomizedActiveSequencersQuerier creates a new cachingRandomizedActiveSequencersQuerier instance.
// It returns an implementation of the ActiveSequencers interface that deterministically randomizes the list of
// currently active sequencers. The return value of the Get method will be the same for the same seed and list
// of addresses from ActiveSequencers.
func NewCachingRandomizedActiveSequencersQuerier(rngSeedFn RandomSeedFn, activeParticipants ActiveParticipants) ActiveSequencers {
	return &cachingRandomizedActiveSequencersQuerier{
		rngSeedFn: rngSeedFn,
		querier:   NewRandomizedActiveSequencersQuerier(rngSeedFn, activeParticipants),
	}
}

// Get returns the list of currently active sequencers.
// If the seed is the same as the last seed and there is a cached list of sequencers, the cached value is returned.
// Otherwise, it retrieves the sequencers from the underlying querier and caches the result.
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

// Contains checks if the provided address is in the list of currently active sequencers.
// If the seed is the same as the last seed and there is a cached list of sequencers, the cached list is used for the check.
// Otherwise, it retrieves the sequencers from the underlying querier, caches the result, and performs the check.
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
