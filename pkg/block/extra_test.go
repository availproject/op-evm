package block

import (
	crand "crypto/rand"
	"fmt"
	"math/rand"
	"reflect"
	"testing"

	"github.com/0xPolygon/polygon-edge/crypto"
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/maticnetwork/avail-settlement/pkg/test"
	"github.com/test-go/testify/assert"
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

func TestExtraDataFraudProofKeyExists(t *testing.T) {
	tAssert := assert.New(t)

	testCases := []struct {
		name                string
		input               types.Hash
		expectedHash        types.Hash
		expectedExistsState bool
		expectedError       error
	}{
		{
			name:                "zero address input",
			input:               types.ZeroHash,
			expectedHash:        types.ZeroHash,
			expectedExistsState: false,
			expectedError:       nil,
		},
		{
			name:                "correct hash input",
			input:               types.StringToHash("1234567890"),
			expectedHash:        types.StringToHash("1234567890"),
			expectedExistsState: true,
			expectedError:       nil,
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("case %d: %s", i, tc.name), func(t *testing.T) {
			hdr := &types.Header{}

			kv := make(map[string][]byte)
			kv[KEY_FRAUDPROOF_OF] = tc.input.Bytes()

			ve := &ValidatorExtra{}
			kv[KEY_EXTRA_VALIDATORS] = ve.MarshalRLPTo(nil)

			hdr.ExtraData = EncodeExtraDataFields(kv)

			key := keystore.NewKeyForDirectICAP(crand.Reader)
			miner := crypto.PubKeyToAddress(&key.PrivateKey.PublicKey)

			hdr.Miner = miner.Bytes()

			hdr, err := WriteSeal(key.PrivateKey, hdr)
			tAssert.NoError(err)

			hashValue, exists, err := GetExtraDataFraudProofKey(hdr)

			if tc.expectedError == nil {
				tAssert.NoError(err)
			} else {
				tAssert.EqualError(tc.expectedError, err.Error())
			}

			tAssert.Equal(tc.expectedExistsState, exists)
			tAssert.Equal(tc.expectedHash.Bytes(), hashValue.Bytes())
		})
	}
}
