package blockchain

import (
	"math/big"
	"sync"

	"github.com/0xPolygon/polygon-edge/blockchain"
	"github.com/0xPolygon/polygon-edge/types"
)

type void struct{}

// Subscription is the blockchain subscription interface.
// It represents a subscription to blockchain events.
type Subscription interface {
	// GetEventCh returns a channel to receive blockchain events.
	GetEventCh() chan *blockchain.Event

	// GetEvent returns the next blockchain event (BLOCKING).
	GetEvent() *blockchain.Event

	// Close closes the subscription.
	Close()
}

// MockSubscription is a mock implementation of the Subscription interface for testing purposes.
type MockSubscription struct {
	*subscription
}

// NewMockSubscription creates a new instance of MockSubscription.
func NewMockSubscription() *MockSubscription {
	return &MockSubscription{
		subscription: &subscription{
			updateCh: make(chan *blockchain.Event),
			closeCh:  make(chan void),
		},
	}
}

// Push pushes a new event to the mock subscription.
func (m *MockSubscription) Push(e *blockchain.Event) {
	m.updateCh <- e
}

// subscription is the Blockchain event subscription object.
// It represents a subscription to blockchain events.
type subscription struct {
	updateCh chan *blockchain.Event // Channel for update information
	closeCh  chan void              // Channel for close signals
}

// GetEventCh returns the channel for receiving blockchain events.
func (s *subscription) GetEventCh() chan *blockchain.Event {
	return s.updateCh
}

// GetEvent returns the next blockchain event (BLOCKING).
func (s *subscription) GetEvent() *blockchain.Event {
	// Wait for an update
	select {
	case ev := <-s.updateCh:
		return ev
	case <-s.closeCh:
		return nil
	}
}

// Close closes the subscription.
func (s *subscription) Close() {
	close(s.closeCh)
}

// EventType represents the type of a blockchain event.
type EventType int

// Constants for different types of blockchain events.
const (
	EventHead  EventType = iota // New head event
	EventReorg                  // Chain reorganization event
	EventFork                   // Chain fork event
)

// Event represents a blockchain event that gets passed to the listeners.
type Event struct {
	// OldChain represents the old chain (removed headers) if there was a reorg.
	OldChain []*types.Header

	// NewChain represents the new part of the chain (or a fork).
	NewChain []*types.Header

	// Difficulty is the new difficulty created with this event.
	Difficulty *big.Int

	// Type is the type of event.
	Type EventType

	// Source is the source that generated the blocks for the event.
	// It can be either the Sealer or the Syncer.
	Source string
}

// Header returns the latest block header for the event.
func (e *Event) Header() *types.Header {
	return e.NewChain[len(e.NewChain)-1]
}

// SetDifficulty sets the event difficulty.
func (e *Event) SetDifficulty(b *big.Int) {
	e.Difficulty = new(big.Int).Set(b)
}

// AddNewHeader appends a header to the event's NewChain array.
func (e *Event) AddNewHeader(newHeader *types.Header) {
	header := newHeader.Copy()

	if e.NewChain == nil {
		// Array doesn't exist yet, create it.
		e.NewChain = []*types.Header{}
	}

	e.NewChain = append(e.NewChain, header)
}

// AddOldHeader appends a header to the event's OldChain array.
func (e *Event) AddOldHeader(oldHeader *types.Header) {
	header := oldHeader.Copy()

	if e.OldChain == nil {
		// Array doesn't exist yet, create it.
		e.OldChain = []*types.Header{}
	}

	e.OldChain = append(e.OldChain, header)
}

// SubscribeEvents returns a blockchain event subscription.
func (b *Blockchain) SubscribeEvents() blockchain.Subscription {
	return b.stream.subscribe()
}

// eventStream is the structure that contains the event list,
// as well as the update channel which it uses to notify of updates.
type eventStream struct {
	sync.Mutex

	// updateCh is the list of channels to notify updates.
	updateCh []chan *blockchain.Event
}

// subscribe creates a new blockchain event subscription.
func (e *eventStream) subscribe() *subscription {
	return &subscription{
		updateCh: e.newUpdateCh(),
		closeCh:  make(chan void),
	}
}

// newUpdateCh returns a new event update channel.
func (e *eventStream) newUpdateCh() chan *blockchain.Event {
	e.Lock()
	defer e.Unlock()

	ch := make(chan *blockchain.Event, 5)
	e.updateCh = append(e.updateCh, ch)

	return ch
}

// push adds a new event and notifies listeners.
func (e *eventStream) push(event *blockchain.Event) {
	e.Lock()
	defer e.Unlock()

	// Notify the listeners.
	for _, update := range e.updateCh {
		update <- event
	}
}
