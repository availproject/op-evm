package staking

import (
	"math/big"
	"testing"

	"github.com/availproject/op-evm/pkg/common"
	"github.com/availproject/op-evm/pkg/test"
	"github.com/hashicorp/go-hclog"
	"github.com/test-go/testify/assert"
)

func TestMinSequencerRater(t *testing.T) {
	tAssert := assert.New(t)

	// TODO: Check if verifier is even necessary to be applied. For now skipping it.
	executor, blockchain, err := test.NewBlockchain(NewVerifier(new(DumbActiveParticipants), hclog.Default()), getGenesisBasePath())
	tAssert.Nil(err)
	tAssert.NotNil(executor)
	tAssert.NotNil(blockchain)

	balance := big.NewInt(0).Mul(big.NewInt(1000), common.ETH)
	coinbaseAddr, coinbaseSignKey := test.NewAccount(t)
	test.DepositBalance(t, coinbaseAddr, balance, blockchain, executor)

	sequencerRater := NewSequencerRater(blockchain, executor, hclog.Default())
	minimum, err := sequencerRater.CurrentMinimum()
	tAssert.NoError(err)
	tAssert.Equal(minimum.Int64(), big.NewInt(0).Int64())

	err = sequencerRater.SetMinimum(big.NewInt(10), coinbaseSignKey)
	tAssert.NoError(err)

	nextMinimum, err := sequencerRater.CurrentMinimum()
	tAssert.NoError(err)
	tAssert.Equal(nextMinimum.Int64(), big.NewInt(10).Int64())
}

func TestMaxSequencerRater(t *testing.T) {
	tAssert := assert.New(t)

	// TODO: Check if verifier is even necessary to be applied. For now skipping it.
	executor, blockchain, err := test.NewBlockchain(NewVerifier(new(DumbActiveParticipants), hclog.Default()), getGenesisBasePath())
	tAssert.Nil(err)
	tAssert.NotNil(executor)
	tAssert.NotNil(blockchain)

	balance := big.NewInt(0).Mul(big.NewInt(1000), common.ETH)
	coinbaseAddr, coinbaseSignKey := test.NewAccount(t)
	test.DepositBalance(t, coinbaseAddr, balance, blockchain, executor)

	participantRater := NewParticipantRater(blockchain, executor, hclog.Default())
	maximum, err := participantRater.CurrentMaximum()
	tAssert.NoError(err)
	tAssert.Equal(maximum.Int64(), big.NewInt(0).Int64())

	err = participantRater.SetMaximum(big.NewInt(10), coinbaseSignKey)
	tAssert.NoError(err)

	nextMaximum, err := participantRater.CurrentMaximum()
	tAssert.NoError(err)
	tAssert.Equal(nextMaximum.Int64(), big.NewInt(10).Int64())
}
