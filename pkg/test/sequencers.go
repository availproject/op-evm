package test

import (
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/maticnetwork/avail-settlement/pkg/staking"
)

type dumbActiveSequencerQuerier struct{}

func (dasq *dumbActiveSequencerQuerier) Get() ([]types.Address, error)          { return nil, nil }
func (dasq *dumbActiveSequencerQuerier) Contains(_ types.Address) (bool, error) { return true, nil }

func DumbActiveSequencers() staking.ActiveSequencers {
	return &dumbActiveSequencerQuerier{}
}
