package avail

import (
	"fmt"
	"sync/atomic"
)

type AvailState uint32

// Define the states in IBFT
const (
	AcceptState AvailState = iota
	ValidateState
	CommitState
	SyncState
)

// String returns the string representation of the passed in state
func (i AvailState) String() string {
	switch i {
	case AcceptState:
		return "AcceptState"

	case ValidateState:
		return "ValidateState"

	case CommitState:
		return "CommitState"

	case SyncState:
		return "SyncState"
	}

	panic(fmt.Sprintf("BUG: state not found %d", i))
}

// currentState defines the current state object in IBFT
type currentState struct {
	// state is the current state
	state uint64
}

// newState creates a new state
func newState() *currentState {
	return &currentState{}
}

// getState returns the current state
func (c *currentState) getState() AvailState {
	stateAddr := &c.state

	return AvailState(atomic.LoadUint64(stateAddr))
}

// setState sets the current state
func (c *currentState) setState(s AvailState) {
	stateAddr := &c.state

	atomic.StoreUint64(stateAddr, uint64(s))
}
