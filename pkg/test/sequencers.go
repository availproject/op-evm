package test

import (
	"github.com/0xPolygon/polygon-edge/types"
)

type DumbActiveSequencers struct{}

func (dasq *DumbActiveSequencers) Get() ([]types.Address, error)          { return nil, nil }
func (dasq *DumbActiveSequencers) Contains(_ types.Address) (bool, error) { return true, nil }
