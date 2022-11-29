package staking

import (
	"github.com/0xPolygon/polygon-edge/types"
	stypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/maticnetwork/avail-settlement/pkg/avail"
)

type AvailSender interface {
	Send(blk *types.Block) error
}

type availSender struct {
	sender avail.Sender
}

func (s *availSender) Send(blk *types.Block) error {
	f := s.sender.SubmitDataAndWaitForStatus(blk.MarshalRLP(), stypes.ExtrinsicStatus{IsInBlock: true})
	if _, err := f.Result(); err != nil {
		return err
	}

	return nil
}

func NewAvailSender(sender avail.Sender) AvailSender {
	return &availSender{sender: sender}
}

type testAvailSender struct{}

func (s *testAvailSender) Send(blk *types.Block) error {
	return nil
}

func NewTestAvailSender() AvailSender {
	return &testAvailSender{}
}
