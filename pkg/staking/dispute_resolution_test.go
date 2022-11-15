package staking

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/maticnetwork/avail-settlement/pkg/test"
	"github.com/test-go/testify/assert"
)

func TestBeginDisputeResolution(t *testing.T) {
	tAssert := assert.New(t)

	// TODO: Check if verifier is even necessary to be applied. For now skipping it.
	executor, blockchain := test.NewBlockchain(t, NewVerifier(new(DumbActiveParticipants), hclog.Default()), getGenesisBasePath())
	tAssert.NotNil(executor)
	tAssert.NotNil(blockchain)

	balance := big.NewInt(0).Mul(big.NewInt(1000), ETH)
	coinbaseAddr, coinbaseSignKey := test.NewAccount(t)
	test.DepositBalance(t, coinbaseAddr, balance, blockchain, executor)

	byzantineSequencerAddr, byzantineSequencerSignKey := test.NewAccount(t)
	test.DepositBalance(t, byzantineSequencerAddr, balance, blockchain, executor)
	_ = byzantineSequencerSignKey

	dr := NewDisputeResolution(blockchain, executor, hclog.Default())

	err := dr.Begin(byzantineSequencerAddr, coinbaseSignKey)
	tAssert.NoError(err)

	probationSequencers, err := dr.Get()
	tAssert.NoError(err)

	fmt.Printf("Probation Sequencers: %v \n", probationSequencers)

	isProbationSequencer, isProbationSequencerErr := dr.Contains(byzantineSequencerAddr)
	tAssert.NoError(isProbationSequencerErr)
	tAssert.True(isProbationSequencer)
}

func TestEndDisputeResolution(t *testing.T) {
	tAssert := assert.New(t)

	// TODO: Check if verifier is even necessary to be applied. For now skipping it.
	executor, blockchain := test.NewBlockchain(t, NewVerifier(new(DumbActiveParticipants), hclog.Default()), getGenesisBasePath())
	tAssert.NotNil(executor)
	tAssert.NotNil(blockchain)

	balance := big.NewInt(0).Mul(big.NewInt(1000), ETH)
	coinbaseAddr, coinbaseSignKey := test.NewAccount(t)
	test.DepositBalance(t, coinbaseAddr, balance, blockchain, executor)

	byzantineSequencerAddr, byzantineSequencerSignKey := test.NewAccount(t)
	test.DepositBalance(t, byzantineSequencerAddr, balance, blockchain, executor)
	_ = byzantineSequencerSignKey

	dr := NewDisputeResolution(blockchain, executor, hclog.Default())

	// BEGIN THE DISPUTE RESOLUTION

	err := dr.Begin(byzantineSequencerAddr, coinbaseSignKey)
	tAssert.NoError(err)

	probationSequencers, err := dr.Get()
	tAssert.NoError(err)

	fmt.Printf("Probation Sequencers: %v \n", probationSequencers)

	isProbationSequencer, isProbationSequencerErr := dr.Contains(byzantineSequencerAddr)
	tAssert.NoError(isProbationSequencerErr)
	tAssert.True(isProbationSequencer)

	// END THE DISPUTE RESOLUTION

	err = dr.End(byzantineSequencerAddr, coinbaseSignKey)
	tAssert.NoError(err)

	probationSequencers, err = dr.Get()
	tAssert.NoError(err)

	fmt.Printf("Probation Sequencers: %v \n", probationSequencers)

	isProbationSequencer, isProbationSequencerErr = dr.Contains(byzantineSequencerAddr)
	tAssert.NoError(isProbationSequencerErr)
	tAssert.False(isProbationSequencer)
}
