package avail

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/centrifuge/go-substrate-rpc-client/v4/scale"
)

const (
	// BlobMagic is required to be present in a `Batch` read from Avail.
	BlobMagic = byte(0b10101010)

	// MaxBlobSize defines the maximum length for stored data in a blob.
	MaxBlobSize = 1 << 24 // 2^24 = 16MB
)

var (
	// ErrDataTooLong is the error returned when the data length exceeds the maximum limit.
	ErrDataTooLong = errors.New("data length exceeds maximum limit")

	// ErrInvalidBlobMagic is the error returned when the blob magic byte is invalid.
	ErrInvalidBlobMagic = errors.New("invalid blob magic")
)

// Blob is a wrapper type for data that is stored in Avail.
type Blob struct {
	Magic byte
	Data  []byte
}

// Encode encodes the blob data into the provided scale.Encoder.
func (b *Blob) Encode(e scale.Encoder) error {
	var err error

	if b.Magic != BlobMagic {
		return fmt.Errorf("%w got %d, expected %d", ErrInvalidBlobMagic, b.Magic, BlobMagic)
	}

	if len(b.Data) > MaxBlobSize {
		return ErrDataTooLong
	}

	err = e.PushByte(b.Magic)
	if err != nil {
		return err
	}

	err = e.EncodeUintCompact(*big.NewInt(int64(len(b.Data))))
	if err != nil {
		return err
	}

	err = e.Write(b.Data)
	if err != nil {
		return err
	}

	return nil
}

// Decode decodes the blob data from the provided scale.Decoder.
func (b *Blob) Decode(d scale.Decoder) error {
	var err error

	b.Magic, err = d.ReadOneByte()
	if err != nil {
		return err
	}

	if b.Magic != BlobMagic {
		return fmt.Errorf("%w got %d, expected %d", ErrInvalidBlobMagic, b.Magic, BlobMagic)
	}

	data_len, err := d.DecodeUintCompact()
	if err != nil || data_len == nil {
		return fmt.Errorf("invalid length (%s)", err.Error())
	}

	if !data_len.IsInt64() {
		return fmt.Errorf("corrupted length (is not int64)")
	}

	if data_len.Int64() > MaxBlobSize {
		return ErrDataTooLong
	}

	b.Data = make([]byte, data_len.Int64())
	err = d.Read(b.Data)
	if err != nil {
		return err
	}

	return nil
}
