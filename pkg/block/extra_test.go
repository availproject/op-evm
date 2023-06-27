package block

import (
	"flag"
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"time"
)

func Test_ExtraData_Encoding(t *testing.T) {
	flag.Parse()

	key := func() string {
		min, max := 2, 32
		return string(randomBytes(t, rand.Intn(max-min)+min))
	}

	value := func() []byte {
		min, max := 2, 256
		return randomBytes(t, rand.Intn(max-min)+min)
	}

	TestRounds := 100

	for i := 0; i < TestRounds; i++ {
		kv := make(map[string][]byte)
		n := rand.Intn(64)
		for j := 0; j < n; j++ {
			kv[key()] = value()
		}

		bs := EncodeExtraDataFields(kv)
		decodedKV, err := DecodeExtraDataFields(bs)
		if err != nil {
			t.Fatal(err)
		}

		for k, v := range kv {
			v2, exists := decodedKV[k]
			if !exists {
				t.Fatalf("couldn't find key %q from decoded key values", k)
			}

			if !reflect.DeepEqual(v, v2) {
				t.Fatalf("v != v2; got len(v): %d, len(v2): %d", len(v), len(v2))
			}

			delete(decodedKV, k)
		}

		if len(decodedKV) > 0 {
			t.Fatalf("len(decodedKV): %d, expected 0", len(decodedKV))
		}
	}
}

func Test_ExtraData_Decoding(t *testing.T) {
	testCases := []struct {
		name         string
		input        []byte
		expected     map[string][]byte
		errorMatcher func(error) bool
	}{
		{
			name:         "nil input",
			input:        nil,
			expected:     make(map[string][]byte),
			errorMatcher: nil,
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("case %d: %s", i, tc.name), func(t *testing.T) {
			kv, err := DecodeExtraDataFields(tc.input)

			switch {
			case err == nil && tc.errorMatcher == nil:
				// correct; carry on
			case err != nil && tc.errorMatcher == nil:
				t.Fatalf("error == %#v, want nil", err)
			case err == nil && tc.errorMatcher != nil:
				t.Fatalf("error == nil, want non-nil")
			case !tc.errorMatcher(err):
				t.Fatalf("error == %#v, want matching", err)
			}

			if !reflect.DeepEqual(kv, tc.expected) {
				t.Fatalf("expected %#v, got %#v", tc.expected, kv)
			}
		})
	}
}

// Seed is a global variable used in functions that generate random data.
// It's value can be specified via a command-line flag `-seed`.
// By default, it uses the current Unix time.
var Seed = flag.Int64("seed", time.Now().Unix(), "random seed used in tests")

// RandomBytes generates a slice of random bytes of the specified size.
// This function uses the global Seed variable for random number generation.
// It's used to produce deterministic results when Seed is specified.
//
// t is a pointer to testing.T, which is the parallel testing interface.
// size is the number of random bytes to generate.
//
// Returns a slice of random bytes.
//
// Example usage:
//
//	func TestRandomBytes(t *testing.T) {
//		bytes := RandomBytes(t, 10)
//		// bytes now holds a slice of 10 random bytes
//	}
func randomBytes(t *testing.T, size int) []byte {
	t.Helper()

	// This allows deterministic tests when seed is specified.
	rnd := rand.New(rand.NewSource(*Seed))
	buf := make([]byte, size)

	bytesRead := 0
	for bytesRead < size {
		n, err := rnd.Read(buf[bytesRead:])
		if err != nil {
			t.Fatalf("error while generating random bytes: %s", err)
		}

		bytesRead += n
	}

	return buf
}
