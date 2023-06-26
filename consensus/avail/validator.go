package avail

import (
	"github.com/0xPolygon/polygon-edge/types"
)

// ValidatorSet represents a collection of addresses acting as validators in the settlement layer system.
// It is implemented as a slice of Address values, each of which represents a unique validator in the network.
type ValidatorSet []types.Address
