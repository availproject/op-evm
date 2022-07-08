package avail

import (
	"errors"
	"fmt"
)

const (
	// BatchMagic is required to be present in a `Batch` read from Avail.
	BatchMagic = byte(0b10101010)
)

var (
	// ErrInvalidBatch is used when batch validation fails.
	ErrInvalidBatch = errors.New("invalid batch")
)

// Batch is a wrapper type for data that is stored in Avail.
type Batch struct {
	Magic byte
	Data  []byte
}

// Validate Batch.
func (b *Batch) Validate() error {
	if b.Magic != BatchMagic {
		return fmt.Errorf("%w: invalid magic", ErrInvalidBatch)
	}

	if b.Data == nil {
		return fmt.Errorf("%w: nil data", ErrInvalidBatch)
	}

	return nil
}
