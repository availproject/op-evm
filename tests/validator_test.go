package tests

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/0xPolygon/polygon-edge/types"
	"github.com/availproject/op-evm/consensus/avail/validator"
	"github.com/availproject/op-evm/pkg/block"
	"github.com/availproject/op-evm/pkg/staking"
	"github.com/availproject/op-evm/pkg/test"
	"github.com/hashicorp/go-hclog"
)

func getGenesisBasePath() string {
	path, _ := os.Getwd()
	return filepath.Join(path, "..")
}

func TestValidatorBlockCheck(t *testing.T) {
	testCases := []struct {
		name         string
		block        func(blockBuilder block.Builder) *types.Block
		errorMatcher func(err error) bool
	}{
		{
			name:         "zero block",
			block:        func(blockBuilder block.Builder) *types.Block { return &types.Block{} },
			errorMatcher: func(err error) bool { return errors.Is(err, validator.ErrInvalidBlock) },
		},
		{
			name:  "coinbase block",
			block: func(blockBuilder block.Builder) *types.Block { b, _ := blockBuilder.Build(); return b },
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("case %d: %s", i, tc.name), func(t *testing.T) {
			verifier := staking.NewVerifier(new(staking.DumbActiveParticipants), hclog.Default())
			executor, blockchain, err := test.NewBlockchain(verifier, getGenesisBasePath())
			if err != nil {
				t.Fatal(err)
			}

			coinbaseAddr, signKey := test.NewAccount(t)
			head := test.GetHeadBlock(t, blockchain)

			blockBuilder, err := block.NewBlockBuilderFactory(blockchain, executor, hclog.Default()).FromParentHash(head.Hash())
			if err != nil {
				t.Fatal(err)
			}

			blockBuilder.SetCoinbaseAddress(coinbaseAddr).SignWith(signKey)

			v := validator.New(blockchain, coinbaseAddr, hclog.Default())
			err = v.Check(tc.block(blockBuilder))
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
		block        func(blockBuilder block.Builder) *types.Block
		errorMatcher func(err error) bool
	}{
		{
			name:  "coinbase block",
			block: func(blockBuilder block.Builder) *types.Block { b, _ := blockBuilder.Build(); return b },
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("case %d: %s", i, tc.name), func(t *testing.T) {
			verifier := staking.NewVerifier(new(staking.DumbActiveParticipants), hclog.Default())

			executor, blockchain, err := test.NewBlockchain(verifier, getGenesisBasePath())
			if err != nil {
				t.Fatal(err)
			}

			coinbaseAddr, signKey := test.NewAccount(t)
			head := test.GetHeadBlock(t, blockchain)

			blockBuilder, err := block.NewBlockBuilderFactory(blockchain, executor, hclog.Default()).FromParentHash(head.Hash())
			if err != nil {
				t.Fatal(err)
			}

			blockBuilder.SetCoinbaseAddress(coinbaseAddr).SignWith(signKey)

			v := validator.New(blockchain, coinbaseAddr, hclog.Default())

			err = v.Apply(tc.block(blockBuilder))
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
		block        func(blockBuilder block.Builder) *types.Block
		errorMatcher func(err error) bool
	}{
		{
			name:  "coinbase block",
			block: func(blockBuilder block.Builder) *types.Block { b, _ := blockBuilder.Build(); return b },
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("case %d: %s", i, tc.name), func(t *testing.T) {
			verifier := staking.NewVerifier(new(staking.DumbActiveParticipants), hclog.Default())
			executor, blockchain, err := test.NewBlockchain(verifier, getGenesisBasePath())
			if err != nil {
				t.Fatal(err)
			}

			coinbaseAddr, signKey := test.NewAccount(t)
			head := test.GetHeadBlock(t, blockchain)

			blockBuilder, err := block.NewBlockBuilderFactory(blockchain, executor, hclog.Default()).FromParentHash(head.Hash())
			if err != nil {
				t.Fatal(err)
			}

			blockBuilder.SetCoinbaseAddress(coinbaseAddr).SignWith(signKey)

			v := validator.New(blockchain, coinbaseAddr, hclog.Default())

			err = v.ProcessFraudproof(tc.block(blockBuilder))
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
