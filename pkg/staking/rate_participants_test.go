package staking

import (
	"math/big"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/maticnetwork/avail-settlement/pkg/test"
	"github.com/test-go/testify/assert"
)

func TestMinParticipantRater(t *testing.T) {
	tAssert := assert.New(t)

	// TODO: Check if verifier is even necessary to be applied. For now skipping it.
	executor, blockchain := test.NewBlockchain(t, NewVerifier(new(DumbActiveParticipants), hclog.Default()), getGenesisBasePath())
	tAssert.NotNil(executor)
	tAssert.NotNil(blockchain)

	balance := big.NewInt(0).Mul(big.NewInt(1000), ETH)
	coinbaseAddr, coinbaseSignKey := test.NewAccount(t)
	test.DepositBalance(t, coinbaseAddr, balance, blockchain, executor)

	participantRater := NewParticipantRater(blockchain, executor, hclog.Default())
	minimum, err := participantRater.CurrentMinimum()
	tAssert.NoError(err)
	tAssert.Equal(minimum.Int64(), big.NewInt(0).Int64())

	err = participantRater.SetMinimum(big.NewInt(10), coinbaseSignKey)
	tAssert.NoError(err)

	nextMinimum, err := participantRater.CurrentMinimum()
	tAssert.NoError(err)
	tAssert.Equal(nextMinimum.Int64(), big.NewInt(10).Int64())
}

func TestMaxParticipantRater(t *testing.T) {
	tAssert := assert.New(t)

	// TODO: Check if verifier is even necessary to be applied. For now skipping it.
	executor, blockchain := test.NewBlockchain(t, NewVerifier(new(DumbActiveParticipants), hclog.Default()), getGenesisBasePath())
	tAssert.NotNil(executor)
	tAssert.NotNil(blockchain)

	balance := big.NewInt(0).Mul(big.NewInt(1000), ETH)
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
