package staking

import (
	"math/big"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/maticnetwork/avail-settlement/pkg/common"
	"github.com/maticnetwork/avail-settlement/pkg/test"
	"github.com/test-go/testify/assert"
)

func TestBeginDisputeResolution(t *testing.T) {
	tAssert := assert.New(t)

	// TODO: Check if verifier is even necessary to be applied. For now skipping it.
	executor, blockchain, err := test.NewBlockchain(NewVerifier(new(DumbActiveParticipants), hclog.Default()), getGenesisBasePath())
	tAssert.Nil(err)
	tAssert.NotNil(executor)
	tAssert.NotNil(blockchain)

	balance := big.NewInt(0).Mul(big.NewInt(1000), common.ETH)
	watchtowerAddr, watchtowerSignKey := test.NewAccount(t)
	test.DepositBalance(t, watchtowerAddr, balance, blockchain, executor)

	byzantineSequencerAddr, _ := test.NewAccount(t)
	test.DepositBalance(t, byzantineSequencerAddr, balance, blockchain, executor)

	// In order to begin the dispute resolution, onlyWatchtower modifier needs to be met.
	// In other words, we first need to stake watchtower as .Begin() can be called only by the staked watchtower.
	stakeAmount := big.NewInt(0).Mul(big.NewInt(10), common.ETH)
	sender := NewTestAvailSender()
	coinbaseStakeErr := Stake(blockchain, executor, sender, hclog.Default(), string(WatchTower), watchtowerAddr, watchtowerSignKey, stakeAmount, 1_000_000, "test")
	tAssert.NoError(coinbaseStakeErr)

	dr := NewDisputeResolution(blockchain, executor, sender, hclog.Default())

	err = dr.Begin(byzantineSequencerAddr, watchtowerSignKey)
	tAssert.NoError(err)

	probationSequencers, err := dr.Get(Sequencer)
	tAssert.NoError(err)

	t.Logf("Probation Sequencers: %v \n", probationSequencers)

	isProbationSequencer, isProbationSequencerErr := dr.Contains(byzantineSequencerAddr, Sequencer)
	tAssert.NoError(isProbationSequencerErr)
	tAssert.True(isProbationSequencer)

	contractSequencerAddr, contractSequencerAddrErr := dr.GetSequencerAddr(watchtowerAddr)
	tAssert.NoError(contractSequencerAddrErr)

	t.Logf("Disputed sequencer addr: %s \n", contractSequencerAddr)

	contractWatchtowerAddr, contractWatchtowerAddrErr := dr.GetWatchtowerAddr(byzantineSequencerAddr)
	tAssert.NoError(contractWatchtowerAddrErr)

	t.Logf("Disputed watchtower addr: %s \n", contractWatchtowerAddr)

	tAssert.Equal(watchtowerAddr, contractWatchtowerAddr)
	tAssert.Equal(byzantineSequencerAddr, contractSequencerAddr)
}

func TestIsBeginDisputeResolutionTx(t *testing.T) {
	from, _ := test.NewAccount(t)
	probationAddr, _ := test.NewAccount(t)
	gasLimit := uint64(1_000_000)

	tx, err := BeginDisputeResolutionTx(from, probationAddr, gasLimit)
	if err != nil {
		t.Fatal(err)
	}

	res, err := IsBeginDisputeResolutionTx(tx)
	if err != nil {
		t.Fatal(err)
	}

	if !res {
		t.Fatalf("IsBeginDisputeResolutionTx(): got %t, expected %t", res, true)
	}
}

func TestEndDisputeResolution(t *testing.T) {
	tAssert := assert.New(t)

	// TODO: Check if verifier is even necessary to be applied. For now skipping it.
	executor, blockchain, err := test.NewBlockchain(NewVerifier(new(DumbActiveParticipants), hclog.Default()), getGenesisBasePath())
	tAssert.Nil(err)
	tAssert.NotNil(executor)
	tAssert.NotNil(blockchain)

	balance := big.NewInt(0).Mul(big.NewInt(1000), common.ETH)
	watchtowerAddr, watchtowerSignKey := test.NewAccount(t)
	test.DepositBalance(t, watchtowerAddr, balance, blockchain, executor)

	byzantineSequencerAddr, _ := test.NewAccount(t)
	test.DepositBalance(t, byzantineSequencerAddr, balance, blockchain, executor)

	// In order to begin the dispute resolution, onlyWatchtower modifier needs to be met.
	// In other words, we first need to stake watchtower as .Begin() can be called only by the staked watchtower.
	stakeAmount := big.NewInt(0).Mul(big.NewInt(10), common.ETH)
	sender := NewTestAvailSender()
	coinbaseStakeErr := Stake(blockchain, executor, sender, hclog.Default(), string(WatchTower), watchtowerAddr, watchtowerSignKey, stakeAmount, 1_000_000, "test")
	tAssert.NoError(coinbaseStakeErr)

	dr := NewDisputeResolution(blockchain, executor, sender, hclog.Default())

	// BEGIN THE DISPUTE RESOLUTION

	err = dr.Begin(byzantineSequencerAddr, watchtowerSignKey)
	tAssert.NoError(err)

	probationSequencers, err := dr.Get(Sequencer)
	tAssert.NoError(err)

	t.Logf("Probation Sequencers: %v \n", probationSequencers)

	isProbationSequencer, isProbationSequencerErr := dr.Contains(byzantineSequencerAddr, Sequencer)
	tAssert.NoError(isProbationSequencerErr)
	tAssert.True(isProbationSequencer)

	// END THE DISPUTE RESOLUTION

	err = dr.End(byzantineSequencerAddr, watchtowerSignKey)
	tAssert.NoError(err)

	probationSequencers, err = dr.Get(Sequencer)
	tAssert.NoError(err)

	t.Logf("Probation Sequencers: %v \n", probationSequencers)

	isProbationSequencer, isProbationSequencerErr = dr.Contains(byzantineSequencerAddr, Sequencer)
	tAssert.NoError(isProbationSequencerErr)
	tAssert.False(isProbationSequencer)
}

func TestFailedEndDisputeResolution(t *testing.T) {
	tAssert := assert.New(t)

	// TODO: Check if verifier is even necessary to be applied. For now skipping it.
	executor, blockchain, err := test.NewBlockchain(NewVerifier(new(DumbActiveParticipants), hclog.Default()), getGenesisBasePath())
	tAssert.Nil(err)
	tAssert.NotNil(executor)
	tAssert.NotNil(blockchain)

	balance := big.NewInt(0).Mul(big.NewInt(1000), common.ETH)
	watchtowerAddr, watchtowerSignKey := test.NewAccount(t)
	test.DepositBalance(t, watchtowerAddr, balance, blockchain, executor)

	byzantineSequencerAddr, byzantineSequencerSignKey := test.NewAccount(t)
	test.DepositBalance(t, byzantineSequencerAddr, balance, blockchain, executor)

	// In order to begin the dispute resolution, onlyWatchtower modifier needs to be met.
	// In other words, we first need to stake watchtower as .Begin() can be called only by the staked watchtower.
	stakeAmount := big.NewInt(0).Mul(big.NewInt(10), common.ETH)
	sender := NewTestAvailSender()
	coinbaseStakeErr := Stake(blockchain, executor, sender, hclog.Default(), string(WatchTower), watchtowerAddr, watchtowerSignKey, stakeAmount, 1_000_000, "test")
	tAssert.NoError(coinbaseStakeErr)

	dr := NewDisputeResolution(blockchain, executor, sender, hclog.Default())

	// BEGIN THE DISPUTE RESOLUTION

	err = dr.Begin(byzantineSequencerAddr, byzantineSequencerSignKey)
	tAssert.NoError(err)

	probationSequencers, err := dr.Get(Sequencer)
	tAssert.NoError(err)

	t.Logf("Probation Sequencers: %v \n", probationSequencers)

	isProbationSequencer, isProbationSequencerErr := dr.Contains(byzantineSequencerAddr, Sequencer)
	tAssert.NoError(isProbationSequencerErr)
	tAssert.False(isProbationSequencer)

	// END THE DISPUTE RESOLUTION

	err = dr.End(watchtowerAddr, watchtowerSignKey)
	// Error will be under the receipt, not here as a failure to apply the transaction.
	tAssert.NoError(err)

	probationSequencers, err = dr.Get(Sequencer)
	tAssert.NoError(err)

	t.Logf("Probation Sequencers: %v \n", probationSequencers)

	isProbationWatchtower, isProbationWatchtowerErr := dr.Contains(watchtowerAddr, WatchTower)
	tAssert.NoError(isProbationWatchtowerErr)
	tAssert.False(isProbationWatchtower)

	isProbationSequencer, isProbationSequencerErr = dr.Contains(byzantineSequencerAddr, Sequencer)
	tAssert.NoError(isProbationSequencerErr)
	tAssert.False(isProbationSequencer)
}
