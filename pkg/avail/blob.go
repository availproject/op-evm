package avail

import (
	"errors"
	"fmt"
)

const (
	// BlobMagic is required to be present in a `Batch` read from Avail.
	BlobMagic = byte(0b10101010)
)

var (
	// ErrInvalidBlob is used when batch validation fails.
	ErrInvalidBlob = errors.New("invalid blob")
)

// Blob is a wrapper type for data that is stored in Avail.
type Blob struct {
	Magic byte
	Data  []byte
}

// Validate Batch.
func (b *Blob) Validate() error {
	if b.Magic != BlobMagic {
		return fmt.Errorf("%w: invalid magic", ErrInvalidBlob)
	}

	if b.Data == nil {
		return fmt.Errorf("%w: nil data", ErrInvalidBlob)
	}

	return nil
}
