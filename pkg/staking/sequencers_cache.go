package staking

import (
	"bytes"
	"sync"

	"github.com/0xPolygon/polygon-edge/blockchain"
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/hashicorp/go-hclog"
)

type cachingActiveSequencersQuerier struct {
	blockchain *blockchain.Blockchain
	logger     hclog.Logger
	querier    ActiveSequencers

	activeSequencers      []types.Address
	activeSequencersMutex *sync.RWMutex
}

func NewCachingActiveSequencersQuerier(blockchain *blockchain.Blockchain, activeSequencers ActiveSequencers, logger hclog.Logger) ActiveSequencers {
	caching := &cachingActiveSequencersQuerier{
		blockchain: blockchain,
		logger:     logger.ResetNamed("caching_active_sequencers_querier"),
		querier:    activeSequencers,

		activeSequencersMutex: &sync.RWMutex{},
	}

	go caching.followBlockchainEvents()

	return caching
}

func (c *cachingActiveSequencersQuerier) Get() ([]types.Address, error) {
	c.activeSequencersMutex.RLock()

	// Refresh active sequencers if there are none.
	if len(c.activeSequencers) == 0 {
		// Unlock the read lock so that state refresh can lock for write.
		c.activeSequencersMutex.RUnlock()
		err := c.refreshActiveSequencers()
		if err != nil {
			return nil, err
		}
		// Re-lock for reading.
		c.activeSequencersMutex.RLock()
	}

	// Deep copy active sequencers.
	as := make([]types.Address, len(c.activeSequencers))
	copy(as, c.activeSequencers)

	defer c.activeSequencersMutex.RUnlock()
	return as, nil
}

func (c *cachingActiveSequencersQuerier) Contains(addr types.Address) (bool, error) {
	c.activeSequencersMutex.RLock()
	defer c.activeSequencersMutex.RUnlock()

	for i := range c.activeSequencers {
		if bytes.Equal(c.activeSequencers[i].Bytes(), addr.Bytes()) {
			return true, nil
		}
	}

	return false, nil
}

func (c *cachingActiveSequencersQuerier) followBlockchainEvents() {
	subscription := c.blockchain.SubscribeEvents()

	for range subscription.GetEventCh() {
		err := c.refreshActiveSequencers()
		if err != nil {
			c.logger.Error("error in refreshing active sequencers", "error", err)
		}
	}
}

func (c *cachingActiveSequencersQuerier) refreshActiveSequencers() error {
	// Lock for writing.
	c.activeSequencersMutex.Lock()
	defer c.activeSequencersMutex.Unlock()

	c.logger.Debug("refreshing active sequencers")

	as, err := c.querier.Get()
	if err != nil {
		return err
	}

	c.activeSequencers = as

	c.logger.Debug("refreshed active sequencers")

	return nil
}
