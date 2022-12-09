package staking

import (
	"github.com/0xPolygon/polygon-edge/types"
)

type Sender interface {
	Send(blk *types.Block) error
}

type testAvailSender struct{}

func (s *testAvailSender) Send(blk *types.Block) error {
	return nil
}

func NewTestAvailSender() Sender {
	return &testAvailSender{}
}
