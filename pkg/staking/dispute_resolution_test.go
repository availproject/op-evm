package staking

import (
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
	coinbaseAddr, _ := test.NewAccount(t)
	test.DepositBalance(t, coinbaseAddr, balance, blockchain, executor)

	byzantineSequencerAddr, byzantineSequencerSignKey := test.NewAccount(t)
	test.DepositBalance(t, byzantineSequencerAddr, balance, blockchain, executor)

	sender := NewTestDisputeResolutionSender()
	dr := NewDisputeResolution(blockchain, executor, sender, hclog.Default())

	err := dr.Begin(byzantineSequencerAddr, byzantineSequencerSignKey)
	tAssert.NoError(err)

	probationSequencers, err := dr.Get()
	tAssert.NoError(err)

	t.Logf("Probation Sequencers: %v \n", probationSequencers)

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
	coinbaseAddr, _ := test.NewAccount(t)
	test.DepositBalance(t, coinbaseAddr, balance, blockchain, executor)

	byzantineSequencerAddr, byzantineSequencerSignKey := test.NewAccount(t)
	test.DepositBalance(t, byzantineSequencerAddr, balance, blockchain, executor)

	sender := NewTestDisputeResolutionSender()
	dr := NewDisputeResolution(blockchain, executor, sender, hclog.Default())

	// BEGIN THE DISPUTE RESOLUTION

	err := dr.Begin(byzantineSequencerAddr, byzantineSequencerSignKey)
	tAssert.NoError(err)

	probationSequencers, err := dr.Get()
	tAssert.NoError(err)

	t.Logf("Probation Sequencers: %v \n", probationSequencers)

	isProbationSequencer, isProbationSequencerErr := dr.Contains(byzantineSequencerAddr)
	tAssert.NoError(isProbationSequencerErr)
	tAssert.True(isProbationSequencer)

	// END THE DISPUTE RESOLUTION

	err = dr.End(byzantineSequencerAddr, byzantineSequencerSignKey)
	tAssert.NoError(err)

	probationSequencers, err = dr.Get()
	tAssert.NoError(err)

	t.Logf("Probation Sequencers: %v \n", probationSequencers)

	isProbationSequencer, isProbationSequencerErr = dr.Contains(byzantineSequencerAddr)
	tAssert.NoError(isProbationSequencerErr)
	tAssert.False(isProbationSequencer)
}

func TestFailedEndDisputeResolution(t *testing.T) {
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

	sender := NewTestDisputeResolutionSender()
	dr := NewDisputeResolution(blockchain, executor, sender, hclog.Default())

	// BEGIN THE DISPUTE RESOLUTION

	err := dr.Begin(byzantineSequencerAddr, byzantineSequencerSignKey)
	tAssert.NoError(err)

	probationSequencers, err := dr.Get()
	tAssert.NoError(err)

	t.Logf("Probation Sequencers: %v \n", probationSequencers)

	isProbationSequencer, isProbationSequencerErr := dr.Contains(byzantineSequencerAddr)
	tAssert.NoError(isProbationSequencerErr)
	tAssert.True(isProbationSequencer)

	// END THE DISPUTE RESOLUTION

	err = dr.End(coinbaseAddr, coinbaseSignKey)
	// Error will be under the receipt, not here as a failure to apply the transaction.
	tAssert.NoError(err)

	probationSequencers, err = dr.Get()
	tAssert.NoError(err)

	t.Logf("Probation Sequencers: %v \n", probationSequencers)

	isProbationSequencer, isProbationSequencerErr = dr.Contains(coinbaseAddr)
	tAssert.NoError(isProbationSequencerErr)
	tAssert.False(isProbationSequencer)

	isProbationSequencer, isProbationSequencerErr = dr.Contains(byzantineSequencerAddr)
	tAssert.NoError(isProbationSequencerErr)
	tAssert.True(isProbationSequencer)
}
