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
	executor, blockchain := test.NewBlockchain(t, nil, getGenesisBasePath())

	tAssert.NotNil(executor)
	tAssert.NotNil(blockchain)

	stakerAddr := types.StringToAddress("0x064A4a5053F3de5eacF5E72A2E97D5F9CF55f031")

	// Following test only queries contract to see if it's working.
	// Does not necessairly look into the responses.
	sequencerQuerier := NewActiveSequencersQuerier(blockchain, executor, hclog.Default())
	staked, err := sequencerQuerier.Contains(stakerAddr)
	tAssert.NoError(err)
	tAssert.False(staked)
}

func TestGetSetStakingThreshold(t *testing.T) {
	tAssert := assert.New(t)

	// TODO: Check if verifier is even necessary to be applied. For now skipping it.
	executor, blockchain := test.NewBlockchain(t, NewVerifier(new(test.DumbActiveSequencers), hclog.Default()), getGenesisBasePath())
	tAssert.NotNil(executor)
	tAssert.NotNil(blockchain)

	balance := big.NewInt(0).Mul(big.NewInt(1000), ETH)
	coinbaseAddr, coinbaseSignKey := test.NewAccount(t)
	test.DepositBalance(t, coinbaseAddr, balance, blockchain, executor)

	defaultStakingThresholdAmount := big.NewInt(0).Mul(big.NewInt(1), ETH)
	targetStakingThresholdAmount := big.NewInt(0).Mul(big.NewInt(10), ETH)

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

	executor, blockchain := test.NewBlockchain(t, NewVerifier(new(test.DumbActiveSequencers), hclog.Default()), getGenesisBasePath())
	tAssert.NotNil(executor)
	tAssert.NotNil(blockchain)

	stakeAmount := big.NewInt(0).Mul(big.NewInt(10), ETH)
	balance := big.NewInt(0).Mul(big.NewInt(1000), ETH)

	// GET THE REQUIRED ADDRESSES

	coinbaseAddr, coinbaseSignKey := test.NewAccount(t)
	test.DepositBalance(t, coinbaseAddr, balance, blockchain, executor)

	stakingThresholdQuerier := NewStakingThresholdQuerier(blockchain, executor, hclog.Default())
	setErr := stakingThresholdQuerier.Set(big.NewInt(10), coinbaseSignKey)
	tAssert.NoError(setErr)

	stakerAddr, stakerSignKey := test.NewAccount(t)
	test.DepositBalance(t, stakerAddr, balance, blockchain, executor)

	sequencerQuerier := NewActiveSequencersQuerier(blockchain, executor, hclog.Default())

	// Base staker, necessary for unstaking to be available (needs at least one active staker as a leftover)
	coinbaseStakeErr := Stake(blockchain, executor, hclog.Default(), "sequencer", coinbaseAddr, coinbaseSignKey, stakeAmount, 1_000_000, "test")
	tAssert.NoError(coinbaseStakeErr)

	// Following test only queries contract to see if it's working.
	// Does not necessairly look into the responses.
	staked, err := sequencerQuerier.Contains(coinbaseAddr)
	tAssert.NoError(err)
	tAssert.True(staked)

	// Staker that we are going to attempt to stake and unstake.
	stakeErr := Stake(blockchain, executor, hclog.Default(), "sequencer", stakerAddr, stakerSignKey, stakeAmount, 1_000_000, "test")
	tAssert.NoError(stakeErr)

	// Following test only queries contract to see if it's working.
	// Does not necessairly look into the responses.
	staked, err = sequencerQuerier.Contains(stakerAddr)
	tAssert.NoError(err)
	tAssert.True(staked)

	// DO THE UNSTAKE

	// Staker that we are going to attempt to stake and unstake.
	unStakeErr := UnStake(blockchain, executor, hclog.Default(), coinbaseAddr, coinbaseSignKey, 1_000_000, "test")
	tAssert.NoError(unStakeErr)

	// Following test only queries contract to see if it's working.
	// Does not necessairly look into the responses.
	unstaked, err := sequencerQuerier.Contains(coinbaseAddr)
	tAssert.NoError(err)
	tAssert.False(unstaked)
}

func TestSlashStaker(t *testing.T) {
	tAssert := assert.New(t)

	executor, blockchain := test.NewBlockchain(t, NewVerifier(new(test.DumbActiveSequencers), hclog.Default()), getGenesisBasePath())
	tAssert.NotNil(executor)
	tAssert.NotNil(blockchain)

	// GET THE REQUIRED ADDRESSES

	stakeAmount := big.NewInt(0).Mul(big.NewInt(10), ETH)
	balance := big.NewInt(0).Mul(big.NewInt(1000), ETH)

	coinbaseAddr, coinbaseSignKey := test.NewAccount(t)
	test.DepositBalance(t, coinbaseAddr, balance, blockchain, executor)

	// Base staker, necessary for unstaking to be available (needs at least one active staker as a leftover)
	coinbaseStakeErr := Stake(blockchain, executor, hclog.Default(), "sequencer", coinbaseAddr, coinbaseSignKey, stakeAmount, 1_000_000, "test")
	tAssert.NoError(coinbaseStakeErr)

	// Base staker, necessary for unstaking to be available (needs at least one active staker as a leftover)
	coinbaseSlashErr := Slash(blockchain, executor, hclog.Default(), coinbaseAddr, coinbaseSignKey, 1_000_000, "test")
	tAssert.NoError(coinbaseSlashErr)
}
