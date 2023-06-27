package staking

import (
	"github.com/0xPolygon/polygon-edge/types"
)

// Sender is an interface for sending blocks.
type Sender interface {
	// Send sends the given block.
	// It takes a *types.Block as a parameter and returns an error if the operation fails.
	Send(blk *types.Block) error
}

// testAvailSender is an implementation of the Sender interface for testing purposes.
type testAvailSender struct{}

// Send sends the given block.
// It takes a *types.Block as a parameter and always returns nil.
func (s *testAvailSender) Send(blk *types.Block) error {
	return nil
}

// NewTestAvailSender creates a new instance of the testAvailSender.
// It returns a Sender interface that can be used for testing purposes.
func NewTestAvailSender() Sender {
	return &testAvailSender{}
}
