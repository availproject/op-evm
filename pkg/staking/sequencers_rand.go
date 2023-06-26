package staking

import (
	"bytes"
	"fmt"
	"math/rand"
	"sort"

	"github.com/0xPolygon/polygon-edge/types"
)

// RandomSeedFn is a function interface to provide a seed to the deterministic RNG
// used to shuffle sequencer addresses.
type RandomSeedFn func() int64

// ActiveSequencers is an interface for managing the list of currently active sequencers.
type ActiveSequencers interface {
	// Get returns the list of currently active sequencers.
	// It deterministically randomizes the list using the provided seed and returns it.
	// An error is returned if the operation fails.
	Get() ([]types.Address, error)

	// Contains checks if the given address is in the list of currently active sequencers.
	// It returns true if the address is found, false otherwise.
	// An error is returned if the operation fails.
	Contains(addr types.Address) (bool, error)
}

type randomizedActiveSequencersQuerier struct {
	rngSeedFn RandomSeedFn
	querier   ActiveParticipants
}

// NewRandomizedActiveSequencersQuerier creates a new instance of randomizedActiveSequencersQuerier.
// It returns an implementation of the ActiveSequencers interface that deterministically randomizes
// the list of currently active sequencers. The return value of the Get method will be the same for
// the same seed and list of addresses from ActiveParticipants.
func NewRandomizedActiveSequencersQuerier(rngSeedFn RandomSeedFn, activeParticipants ActiveParticipants) ActiveSequencers {
	return &randomizedActiveSequencersQuerier{
		rngSeedFn: rngSeedFn,
		querier:   activeParticipants,
	}
}

type addresses []types.Address

func (as addresses) Len() int           { return len(as) }
func (as addresses) Less(i, j int) bool { return bytes.Compare(as[i].Bytes(), as[j].Bytes()) < 0 }
func (as addresses) Swap(i, j int)      { tmp := as[i]; as[i] = as[j]; as[j] = tmp }

// Get returns the list of currently active sequencers.
// It sorts the addresses in ascending order and then deterministically shuffles the list using the provided seed.
// An error is returned if the operation fails.
func (rasq *randomizedActiveSequencersQuerier) Get() ([]types.Address, error) {
	as, err := rasq.querier.Get(Sequencer)
	if err != nil {
		return nil, err
	}

	// First sort the Addresses.
	addrs := addresses(as)
	sort.Stable(addrs)

	// Now shuffle the Addresses, using the blockchain head block number as the RNG seed.
	rng := rand.New(rand.NewSource(rasq.rngSeedFn()))
	rng.Shuffle(addrs.Len(), addrs.Swap)

	return addrs, nil
}

// Contains checks if the given address is in the list of currently active sequencers.
// It delegates the check to the underlying querier and returns the result.
// An error is returned if the operation fails.
func (rasq *randomizedActiveSequencersQuerier) Contains(addr types.Address) (bool, error) {
	fmt.Printf("AM I HERE AND WHAT TYPE AM I? type: %v \n", Sequencer)
	return rasq.querier.Contains(addr, Sequencer)
}
