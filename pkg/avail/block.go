package avail

import (
	"errors"
	"fmt"
)

const (
	// BlockMagic is required to be present in a block read from Avail.
	BlockMagic = byte(0b10101010)
)

var (
	// ErrInvalidBlock is used when block validation fails.
	ErrInvalidBlock = errors.New("invalid block")
)

// Block is a wrapper type for data that is stored in Avail.
type Block struct {
	Magic byte
	Data  []byte
}

// Validate Block.
func (b *Block) Validate() error {
	if b.Magic != BlockMagic {
		return fmt.Errorf("%w: invalid magic", ErrInvalidBlock)
	}

	if b.Data == nil {
		return fmt.Errorf("%w: nil data", ErrInvalidBlock)
	}

	return nil
}
