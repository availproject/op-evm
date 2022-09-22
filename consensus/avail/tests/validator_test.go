package tests

import (
	"errors"
	"fmt"
	"testing"

	"github.com/0xPolygon/polygon-edge/types"
	"github.com/maticnetwork/avail-settlement/consensus/avail/validator"
)

func TestValidatorBlockCheck(t *testing.T) {
	testCases := []struct {
		name         string
		block        func(bf BlockFactory, parent *types.Block) *types.Block
		errorMatcher func(err error) bool
	}{
		{
			name:         "zero block",
			block:        func(bf BlockFactory, parent *types.Block) *types.Block { return &types.Block{} },
			errorMatcher: func(err error) bool { return errors.Is(err, validator.ErrInvalidBlock) },
		},
		{
			name:  "coinbase block",
			block: func(bf BlockFactory, parent *types.Block) *types.Block { return bf.BuildBlock(parent, nil) },
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("case %d: %s", i, tc.name), func(t *testing.T) {
			executor, blockchain := newBlockchain(t)
			coinbaseAddr, signKey := newAccount(t)
			bf := NewBasicBlockFactory(t, executor, coinbaseAddr, signKey)

			v := validator.New(blockchain, executor, coinbaseAddr)

			head := getHeadBlock(t, blockchain)

			err := v.Check(tc.block(bf, head))
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
		})
	}
}

func TestValidatorApplyBlockToBlockchain(t *testing.T) {
	testCases := []struct {
		name         string
		block        func(bf BlockFactory, parent *types.Block) *types.Block
		errorMatcher func(err error) bool
	}{
		{
			name:  "coinbase block",
			block: func(bf BlockFactory, parent *types.Block) *types.Block { return bf.BuildBlock(parent, nil) },
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("case %d: %s", i, tc.name), func(t *testing.T) {
			executor, blockchain := newBlockchain(t)
			coinbaseAddr, signKey := newAccount(t)
			bf := NewBasicBlockFactory(t, executor, coinbaseAddr, signKey)

			v := validator.New(blockchain, executor, coinbaseAddr)

			head := getHeadBlock(t, blockchain)

			err := v.Apply(tc.block(bf, head))
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
		})
	}
}

func TestValidatorProcessesFraudproof(t *testing.T) {
	testCases := []struct {
		name         string
		block        func(bf BlockFactory, parent *types.Block) *types.Block
		errorMatcher func(err error) bool
	}{
		{
			name:  "coinbase block",
			block: func(bf BlockFactory, parent *types.Block) *types.Block { return bf.BuildBlock(parent, nil) },
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("case %d: %s", i, tc.name), func(t *testing.T) {
			executor, blockchain := newBlockchain(t)
			coinbaseAddr, signKey := newAccount(t)
			bf := NewBasicBlockFactory(t, executor, coinbaseAddr, signKey)

			v := validator.New(blockchain, executor, coinbaseAddr)

			head := getHeadBlock(t, blockchain)

			err := v.ProcessFraudproof(tc.block(bf, head))
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
		})
	}
}
