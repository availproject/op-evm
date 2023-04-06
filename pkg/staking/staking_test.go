package staking

import (
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/0xPolygon/polygon-edge/types"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/hashicorp/go-hclog"
	commontoken "github.com/maticnetwork/avail-settlement/pkg/common"
	"github.com/maticnetwork/avail-settlement/pkg/test"
	"github.com/test-go/testify/assert"
)

func getGenesisBasePath() string {
	path, _ := os.Getwd()
	return filepath.Join(path, "..", "..")
}

func AddrToAccount(addr types.Address) accounts.Account {
	return accounts.Account{Address: common.BytesToAddress([]byte{})}
}

func TestIsContractDeployed(t *testing.T) {
	tAssert := assert.New(t)

	// TODO: Check if verifier is even necessary to be applied. For now skipping it.
	executor, blockchain, err := test.NewBlockchain(nil, getGenesisBasePath())
	tAssert.Nil(err)
	tAssert.NotNil(executor)
	tAssert.NotNil(blockchain)

	stakerAddr := types.StringToAddress("0x064A4a5053F3de5eacF5E72A2E97D5F9CF55f031")

	// Following test only queries contract to see if it's working.
	// Does not necessairly look into the responses.
	sequencerQuerier := NewActiveParticipantsQuerier(blockchain, executor, hclog.Default())
	staked, err := sequencerQuerier.Contains(stakerAddr, Sequencer)
	tAssert.NoError(err)
	tAssert.False(staked)
}

func TestGetSetStakingThreshold(t *testing.T) {
	tAssert := assert.New(t)

	// TODO: Check if verifier is even necessary to be applied. For now skipping it.
	executor, blockchain, err := test.NewBlockchain(NewVerifier(new(DumbActiveParticipants), hclog.Default()), getGenesisBasePath())
	tAssert.Nil(err)
	tAssert.NotNil(executor)
	tAssert.NotNil(blockchain)

	balance := big.NewInt(0).Mul(big.NewInt(1000), commontoken.ETH)
	coinbaseAddr, coinbaseSignKey := test.NewAccount(t)
	test.DepositBalance(t, coinbaseAddr, balance, blockchain, executor)

	defaultStakingThresholdAmount := big.NewInt(0).Mul(big.NewInt(1), commontoken.ETH)
	targetStakingThresholdAmount := big.NewInt(0).Mul(big.NewInt(10), commontoken.ETH)

	stakingThresholdQuerier := NewStakingThresholdQuerier(blockchain, executor, hclog.Default())

	currentThreshold, err := stakingThresholdQuerier.Current()
	tAssert.NoError(err)
	tAssert.Equal(currentThreshold, defaultStakingThresholdAmount)
	fmt.Printf("Default staking threshold (wei): %d \n", currentThreshold)

	setErr := stakingThresholdQuerier.Set(targetStakingThresholdAmount, coinbaseSignKey)
	tAssert.NoError(setErr)

	targetThreshold, err := stakingThresholdQuerier.Current()
	tAssert.NoError(err)
	tAssert.Equal(targetThreshold, targetStakingThresholdAmount)
	fmt.Printf("Target staking threshold (wei): %d \n", targetThreshold)
}

// TestIsContractStakedAndUnStaked - Is a bit more complex unit test that requires to write multiple blocks
// in order to satisfy the states. It will produce 5 blocks, written into the database and as a outcome,
// staker address will be staked and removed from the sequencer list resulting in a passing test.
// Note that there has to be 2 stakers at least as minimum staker amount in the contract is 1.
func TestIsContractStakedAndUnStaked(t *testing.T) {
	tAssert := assert.New(t)

	executor, blockchain, err := test.NewBlockchain(NewVerifier(new(DumbActiveParticipants), hclog.Default()), getGenesisBasePath())
	tAssert.Nil(err)
	tAssert.NotNil(executor)
	tAssert.NotNil(blockchain)

	stakeAmount := big.NewInt(0).Mul(big.NewInt(10), commontoken.ETH)
	balance := big.NewInt(0).Mul(big.NewInt(1000), commontoken.ETH)

	// GET THE REQUIRED ADDRESSES

	coinbaseAddr, coinbaseSignKey := test.NewAccount(t)
	test.DepositBalance(t, coinbaseAddr, balance, blockchain, executor)

	stakingThresholdQuerier := NewStakingThresholdQuerier(blockchain, executor, hclog.Default())
	setErr := stakingThresholdQuerier.Set(big.NewInt(10), coinbaseSignKey)
	tAssert.NoError(setErr)

	stakerAddr, stakerSignKey := test.NewAccount(t)
	test.DepositBalance(t, stakerAddr, balance, blockchain, executor)

	sequencerQuerier := NewActiveParticipantsQuerier(blockchain, executor, hclog.Default())
	sender := NewTestAvailSender()

	// Base staker, necessary for unstaking to be available (needs at least one active staker as a leftover)
	coinbaseStakeErr := Stake(blockchain, executor, sender, hclog.Default(), string(WatchTower), coinbaseAddr, coinbaseSignKey, stakeAmount, 1_000_000, "test")
	tAssert.NoError(coinbaseStakeErr)

	// Following test only queries contract to see if it's working.
	// Does not necessairly look into the responses.
	staked, err := sequencerQuerier.Contains(coinbaseAddr, WatchTower)
	tAssert.NoError(err)
	tAssert.True(staked)

	// Staker that we are going to attempt to stake and unstake.
	stakeErr := Stake(blockchain, executor, sender, hclog.Default(), string(WatchTower), stakerAddr, stakerSignKey, stakeAmount, 1_000_000, "test")
	tAssert.NoError(stakeErr)

	// Following test only queries contract to see if it's working.
	// Does not necessairly look into the responses.
	staked, err = sequencerQuerier.Contains(stakerAddr, WatchTower)
	tAssert.NoError(err)
	tAssert.True(staked)

	// DO THE UNSTAKE

	// Staker that we are going to attempt to stake and unstake.
	unStakeErr := UnStake(blockchain, executor, sender, hclog.Default(), coinbaseAddr, coinbaseSignKey, 1_000_000, "test")
	tAssert.NoError(unStakeErr)

	// Following test only queries contract to see if it's working.
	// Does not necessairly look into the responses.
	unstaked, err := sequencerQuerier.Contains(coinbaseAddr, WatchTower)
	tAssert.NoError(err)
	tAssert.False(unstaked)
}

func TestSlashStaker(t *testing.T) {
	tAssert := assert.New(t)

	executor, blockchain, err := test.NewBlockchain(NewVerifier(new(DumbActiveParticipants), hclog.Default()), getGenesisBasePath())
	tAssert.Nil(err)
	tAssert.NotNil(executor)
	tAssert.NotNil(blockchain)

	// GET THE REQUIRED ADDRESSES

	stakeAmount := big.NewInt(0).Mul(big.NewInt(10), commontoken.ETH)
	balance := big.NewInt(0).Mul(big.NewInt(1000), commontoken.ETH)

	coinbaseAddr, coinbaseSignKey := test.NewAccount(t)
	test.DepositBalance(t, coinbaseAddr, balance, blockchain, executor)

	sequencerAddr, sequencerSignKey := test.NewAccount(t)
	test.DepositBalance(t, sequencerAddr, balance, blockchain, executor)

	maliciousSequencerAddr, maliciousSignKey := test.NewAccount(t)
	test.DepositBalance(t, maliciousSequencerAddr, balance, blockchain, executor)

	sender := NewTestAvailSender()

	// Base staker, necessary for unstaking to be available (needs at least one active staker as a leftover)
	coinbaseStakeErr := Stake(blockchain, executor, sender, hclog.Default(), string(WatchTower), coinbaseAddr, coinbaseSignKey, stakeAmount, 1_000_000, "test")
	tAssert.NoError(coinbaseStakeErr)

	// Sequencer that is going to slash the malicious sequencer
	sequencerStakeErr := Stake(blockchain, executor, sender, hclog.Default(), string(Sequencer), sequencerAddr, sequencerSignKey, stakeAmount, 1_000_000, "test")
	tAssert.NoError(sequencerStakeErr)

	// Sequencer that pretends it's malicious
	maliciousSequencerStakeErr := Stake(blockchain, executor, sender, hclog.Default(), string(Sequencer), maliciousSequencerAddr, maliciousSignKey, stakeAmount, 1_000_000, "test")
	tAssert.NoError(maliciousSequencerStakeErr)

	// Checking for the correct balance before and after slashing.
	participantQuerier := NewActiveParticipantsQuerier(blockchain, executor, hclog.Default())

	totalContractBalance, tcbErr := participantQuerier.GetTotalStakedAmount()
	tAssert.NoError(tcbErr)

	t.Logf("Contract balance before slashing: %v", totalContractBalance)
	totalContractBalanceBefore, _ := new(big.Int).SetString("30000000000000000000", 10)
	tAssert.Equal(totalContractBalance, totalContractBalanceBefore)

	coinbaseBalance, cbErr := participantQuerier.GetBalance(coinbaseAddr)
	tAssert.NoError(cbErr)

	t.Logf("Coinbase contract balance before slashing: %v", coinbaseBalance)
	coinbaseBalanceBefore, _ := new(big.Int).SetString("10000000000000000000", 10)
	tAssert.Equal(coinbaseBalance, coinbaseBalanceBefore)

	sequencerBalance, sbErr := participantQuerier.GetBalance(sequencerAddr)
	tAssert.NoError(sbErr)
	t.Logf("Sequencer contract balance before slashing: %v", sequencerBalance)

	maliciousSequencerBalance, msbErr := participantQuerier.GetBalance(maliciousSequencerAddr)
	tAssert.NoError(msbErr)
	t.Logf("Malicious sequencer contract balance before slashing: %v", maliciousSequencerBalance)

	parentHeader := blockchain.Header()
	transition, tErr := executor.BeginTxn(parentHeader.StateRoot, parentHeader, coinbaseAddr)
	tAssert.NoError(tErr)

	balanceBefore := transition.GetBalance(coinbaseAddr)
	t.Logf("Watchtower (recipient of fee) wallet balance before slashing: %v", balanceBefore)
	watchtowerWalletBalanceBefore, _ := new(big.Int).SetString("990000000000000000000", 10)
	tAssert.Equal(balanceBefore, watchtowerWalletBalanceBefore)

	dr := NewDisputeResolution(blockchain, executor, sender, hclog.Default())

	err = dr.Begin(maliciousSequencerAddr, coinbaseSignKey)
	tAssert.NoError(err)

	isProbationSequencer, isProbationSequencerErr := dr.Contains(maliciousSequencerAddr, Sequencer)
	tAssert.NoError(isProbationSequencerErr)
	tAssert.True(isProbationSequencer)

	// Must be executed with a correct sequencer (valid one), should fail if it's not.
	// Slashing implements onSequencer modifier (decorator).
	coinbaseSlashErr := Slash(blockchain, executor, hclog.Default(), sequencerAddr, sequencerSignKey, maliciousSequencerAddr, 1_000_000, "test")
	tAssert.NoError(coinbaseSlashErr)

	isProbationSequencer, isProbationSequencerErr = dr.Contains(maliciousSequencerAddr, Sequencer)
	tAssert.NoError(isProbationSequencerErr)
	tAssert.False(isProbationSequencer)

	totalContractBalance, tcbErr = participantQuerier.GetTotalStakedAmount()
	tAssert.NoError(tcbErr)

	t.Logf("Contract balance after slashing: %v", totalContractBalance)
	totalContractBalanceBefore, _ = new(big.Int).SetString("29900000000000000000", 10)
	tAssert.Equal(totalContractBalance, totalContractBalanceBefore)

	coinbaseBalance, cbErr = participantQuerier.GetBalance(coinbaseAddr)
	tAssert.NoError(cbErr)

	t.Logf("Coinbase contract balance after slashing: %v", coinbaseBalance)
	coinbaseBalanceEnd, _ := new(big.Int).SetString("10000000000000000000", 10)
	tAssert.Equal(coinbaseBalance, coinbaseBalanceEnd)

	sequencerBalance, sbErr = participantQuerier.GetBalance(sequencerAddr)
	tAssert.NoError(sbErr)
	sequencerBalanceAfter, _ := new(big.Int).SetString("10000000000000000000", 10)
	tAssert.Equal(sequencerBalance, sequencerBalanceAfter)

	t.Logf("Sequencer contract balance after slashing: %v", sequencerBalance)

	maliciousSequencerBalance, msbErr = participantQuerier.GetBalance(maliciousSequencerAddr)
	tAssert.NoError(msbErr)
	maliciousSequencerBalanceAfter, _ := new(big.Int).SetString("9900000000000000000", 10)
	tAssert.Equal(maliciousSequencerBalance, maliciousSequencerBalanceAfter)

	t.Logf("Malicious sequencer contract balance after slashing: %v", maliciousSequencerBalance)

	parentHeader = blockchain.Header()
	transition, tErr = executor.BeginTxn(parentHeader.StateRoot, parentHeader, coinbaseAddr)
	tAssert.NoError(tErr)

	balanceAfter := transition.GetBalance(coinbaseAddr)
	t.Logf("Watchtower (recipient of fee) wallet balance after slashing: %v", balanceAfter)
	watchtowerWalletBalanceAfter, _ := new(big.Int).SetString("990100000000000000000", 10)
	tAssert.Equal(balanceAfter, watchtowerWalletBalanceAfter)

}
