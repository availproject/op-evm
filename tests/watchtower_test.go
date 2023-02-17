package tests

import (
	"errors"
	"fmt"
	"math/big"
	"testing"

	"github.com/0xPolygon/polygon-edge/types"
	"github.com/hashicorp/go-hclog"
	"github.com/maticnetwork/avail-settlement/consensus/avail/watchtower"
	"github.com/maticnetwork/avail-settlement/pkg/block"
	"github.com/maticnetwork/avail-settlement/pkg/common"
	"github.com/maticnetwork/avail-settlement/pkg/staking"
	"github.com/maticnetwork/avail-settlement/pkg/test"
	"github.com/test-go/testify/assert"
)

func TestWatchTowerBlockCheck(t *testing.T) {
	coinbaseAddr, signKey := test.NewAccount(t)

	testCases := []struct {
		name         string
		block        func(blockBuilder block.Builder) *types.Block
		errorMatcher func(err error) bool
	}{
		{
			name:         "zero block",
			block:        func(blockBuilder block.Builder) *types.Block { return &types.Block{} },
			errorMatcher: func(err error) bool { return errors.Is(err, watchtower.ErrInvalidBlock) },
		},
		{
			name: "coinbase block",
			block: func(blockBuilder block.Builder) *types.Block {
				b, _ := blockBuilder.SignWith(signKey).Build()
				return b
			},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("case %d: %s", i, tc.name), func(t *testing.T) {
			verifier := staking.NewVerifier(new(staking.DumbActiveParticipants), hclog.Default())
			executor, blockchain := test.NewBlockchain(t, verifier, getGenesisBasePath())
			head := test.GetHeadBlock(t, blockchain)

			blockBuilder, err := block.NewBlockBuilderFactory(blockchain, executor, hclog.Default()).FromParentHash(head.Hash(), 0)
			if err != nil {
				t.Fatal(err)
			}

			wt := watchtower.New(blockchain, executor, nil, hclog.Default(), coinbaseAddr, signKey)

			err = wt.Check(tc.block(blockBuilder))
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

func TestWatchTowerBlockConstructFraudProof(t *testing.T) {
	tAssert := assert.New(t)
	verifier := staking.NewVerifier(new(staking.DumbActiveParticipants), hclog.Default())
	executor, blockchain := test.NewBlockchain(t, verifier, getGenesisBasePath())

	balance := big.NewInt(0).Mul(big.NewInt(1000), common.ETH)

	coinbaseAddr, signKey := test.NewAccount(t)
	test.DepositBalance(t, coinbaseAddr, balance, blockchain, executor)

	asq := staking.NewActiveParticipantsQuerier(blockchain, executor, hclog.Default())
	verifier = staking.NewVerifier(asq, hclog.Default())
	blockchain.SetConsensus(verifier)

	wt := watchtower.New(blockchain, executor, nil, hclog.Default(), coinbaseAddr, signKey)

	stakeAmount := big.NewInt(0).Mul(big.NewInt(20), common.ETH)
	sender := staking.NewTestAvailSender()
	coinbaseStakeErr := staking.Stake(blockchain, executor, sender, hclog.Default(), string(staking.WatchTower), coinbaseAddr, signKey, stakeAmount, 1_000_000, "test")
	tAssert.NoError(coinbaseStakeErr)

	testCases := []struct {
		name         string
		block        func(blockBuilder block.Builder) *types.Block
		errorMatcher func(err error) bool
	}{
		{
			name: "coinbase block",
			block: func(blockBuilder block.Builder) *types.Block {
				b, _ := blockBuilder.SignWith(signKey).Build()
				return b
			},
		},
		{
			name: "malicious block _ no validators extradata",
			block: func(blockBuilder block.Builder) *types.Block {
				b, _ := blockBuilder.SignWith(signKey).Build()
				// Screw up the extra data...
				b.Header.ExtraData = []byte{1, 2, 3}
				return b
			},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("case %d: %s", i, tc.name), func(t *testing.T) {
			head := test.GetHeadBlock(t, blockchain)

			blockBuilder, err := block.NewBlockBuilderFactory(blockchain, executor, hclog.Default()).FromParentHash(head.Hash(), 0)
			if err != nil {
				t.Fatal(err)
			}

			blk := tc.block(blockBuilder)
			if err := wt.Check(blk); err != nil {
				fpBlk, err := wt.ConstructFraudproof(blk)
				tAssert.NoError(err)

				data, err := block.DecodeExtraDataFields(fpBlk.Header.ExtraData)
				tAssert.NoError(err)
				tAssert.Equal(blk.Hash(), types.BytesToHash(data[block.KeyFraudProofOf]))
			}

		})
	}
}
