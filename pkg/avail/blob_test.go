package avail

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/centrifuge/go-substrate-rpc-client/v4/scale"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/maticnetwork/avail-settlement/pkg/test"
)

func Test_BlobEncoding(t *testing.T) {
	testCases := []struct {
		name             string
		input            []byte
		encodeErrMatcher func(error) bool
		decodeErrMatcher func(error) bool
	}{
		{
			name:             "test nil data",
			input:            nil,
			encodeErrMatcher: nil,
			decodeErrMatcher: nil,
		},
		{
			name:             "test empty data",
			input:            []byte{},
			encodeErrMatcher: nil,
			decodeErrMatcher: nil,
		},
		{
			name:             "test 1 byte data",
			input:            []byte{0x1A},
			encodeErrMatcher: nil,
			decodeErrMatcher: nil,
		},
		{
			name:             "test 512 bytes of random data",
			input:            test.RandomBytes(t, 512),
			encodeErrMatcher: nil,
			decodeErrMatcher: nil,
		},
		{
			name:             "test 16 MB of random data",
			input:            test.RandomBytes(t, 1<<24),
			encodeErrMatcher: nil,
			decodeErrMatcher: nil,
		},
		{
			name:             "test 16 MB + 1 of random data",
			input:            test.RandomBytes(t, 1<<24+1),
			encodeErrMatcher: func(err error) bool { return errors.Is(err, ErrDataTooLong) },
			decodeErrMatcher: nil,
		},
	}

	for i, tc := range testCases {
		ok := t.Run(fmt.Sprintf("case %d: %s", i, tc.name), func(t *testing.T) {
			input := Blob{
				Data: tc.input,
			}

			buf := bytes.NewBuffer(nil)

			// Encode Blob
			{
				encoder := scale.NewEncoder(buf)
				err := input.Encode(*encoder)
				switch {
				case err == nil && tc.encodeErrMatcher == nil:
					// correct; carry on
				case err != nil && tc.encodeErrMatcher == nil:
					t.Fatalf("error == %#v, want nil", err)
				case err == nil && tc.encodeErrMatcher != nil:
					t.Fatalf("error == nil, want non-nil")
				case !tc.encodeErrMatcher(err):
					t.Fatalf("error == %#v, want matching", err)
				case tc.encodeErrMatcher(err):
					// When encoding is expected to return an error, decoding
					// couldn't work after that, so test case need to "short
					// circuit" here.
					return
				}
			}

			// Decode blob
			{
				var output Blob
				decoder := scale.NewDecoder(buf)
				err := output.Decode(*decoder)
				switch {
				case err == nil && tc.decodeErrMatcher == nil:
					// correct; carry on
				case err != nil && tc.decodeErrMatcher == nil:
					t.Fatalf("error == %#v, want nil", err)
				case err == nil && tc.decodeErrMatcher != nil:
					t.Fatalf("error == nil, want non-nil")
				case !tc.decodeErrMatcher(err):
					t.Fatalf("error == %#v, want matching", err)
				}

				if !cmp.Equal(input.Data, output.Data, cmpopts.EquateEmpty()) {
					t.Fatalf("input & output bytes don't match: expected %v, got %v", input.Data, output.Data)
				}
			}
		})

		if !ok {
			t.Logf("random seed: %d", *test.Seed)
		}
	}
}
