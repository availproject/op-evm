package block

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"

	"github.com/maticnetwork/avail-settlement/pkg/test"
)

func Test_ExtraData_Encoding(t *testing.T) {
	key := func() string {
		min, max := 2, 32
		return string(test.RandomBytes(t, rand.Intn(max-min)+min))
	}

	value := func() []byte {
		min, max := 2, 256
		return test.RandomBytes(t, rand.Intn(max-min)+min)
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
