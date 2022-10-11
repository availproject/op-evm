package staking

import (
	"math/big"
	"testing"

	"github.com/0xPolygon/polygon-edge/types"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/hashicorp/go-hclog"
	"github.com/maticnetwork/avail-settlement/pkg/block"
	"github.com/maticnetwork/avail-settlement/pkg/test"
	"github.com/test-go/testify/assert"
)

func AddrToAccount(addr types.Address) accounts.Account {
	return accounts.Account{Address: common.BytesToAddress([]byte{})}
}

func TestIsContractDeployed(t *testing.T) {
	tAssert := assert.New(t)

	// TODO: Check if verifier is even necessary to be applied. For now skipping it.
	executor, blockchain := test.NewBlockchain(t, nil)

	tAssert.NotNil(executor)
	tAssert.NotNil(blockchain)

	stakerAddr := types.StringToAddress("0x064A4a5053F3de5eacF5E72A2E97D5F9CF55f031")

	// Following test only queries contract to see if it's working.
	// Does not necessairly look into the responses.
	staked, err := IsStaked(stakerAddr, blockchain, executor)
	tAssert.NoError(err)
	tAssert.False(staked)
}

// TestIsContractStakedAndUnStaked - Is a bit more complex unit test that requires to write multiple blocks
// in order to satisfy the states. It will produce 5 blocks, written into the database and as a outcome,
// staker address will be staked and removed from the sequencer list resulting in a passing test.
// Note that there has to be 2 stakers at least as minimum staker amount in the contract is 1.
func TestIsContractStakedAndUnStaked(t *testing.T) {
	tAssert := assert.New(t)

	executor, blockchain := test.NewBlockchain(t, NewVerifier(new(test.DumbActiveSequencers), hclog.Default()))
	tAssert.NotNil(executor)
	tAssert.NotNil(blockchain)

	// GET THE REQUIRED ADDRESSES

	balance := big.NewInt(0).Mul(big.NewInt(1000), ETH)

	coinbaseAddr, coinbaseSignKey := test.NewAccount(t)
	test.DepositBalance(t, coinbaseAddr, balance, blockchain, executor)

	stakerAddr, stakerSignKey := test.NewAccount(t)
	test.DepositBalance(t, stakerAddr, balance, blockchain, executor)

	bfCoinbase := block.NewBlockBuilderFactory(blockchain, executor, hclog.Default())
	blck, err := bfCoinbase.FromParentHash(blockchain.Header().Hash)
	tAssert.NoError(err)

	blck.SetCoinbaseAddress(coinbaseAddr)
	blck.SignWith(coinbaseSignKey)

	// APPLY THE COINBASE STAKE

	// Now lets go build the stake tx and push it to the blockchain.
	stakeTx, err := StakeTx(coinbaseAddr, 1000000)
	tAssert.NoError(err)

	blck.AddTransactions(stakeTx)

	// Write the block to the blockchain
	tAssert.NoError(blck.Write("test"))

	// APPLY THE STAKER STAKE - THIS ONE WILL BE REMOVED

	bfStake := block.NewBlockBuilderFactory(blockchain, executor, hclog.Default())
	stakeBlck, err := bfStake.FromParentHash(blockchain.Header().Hash)
	tAssert.NoError(err)

	stakeBlck.SetCoinbaseAddress(stakerAddr)
	stakeBlck.SignWith(stakerSignKey)

	// Now lets go build the stake tx and push it to the blockchain.
	stakerTx, err := StakeTx(stakerAddr, 1000000)
	tAssert.NoError(err)

	stakeBlck.AddTransactions(stakerTx)

	// Write the block to the blockchain
	tAssert.NoError(stakeBlck.Write("test"))

	// Following test only queries contract to see if it's working.
	// Does not necessairly look into the responses.
	staked, err := IsStaked(stakerAddr, blockchain, executor)
	tAssert.NoError(err)
	tAssert.True(staked)

	// DO THE UNSTAKE

	bfUnStake := block.NewBlockBuilderFactory(blockchain, executor, hclog.Default())
	unStakeBlck, err := bfUnStake.FromParentHash(blockchain.Header().Hash)
	tAssert.NoError(err)

	unStakeBlck.SetCoinbaseAddress(stakerAddr)
	unStakeBlck.SignWith(stakerSignKey)

	// Now lets go build the stake tx and push it to the blockchain.
	unstakeTx, err := UnStakeTx(stakerAddr, 100000)
	tAssert.NoError(err)
	tAssert.NotNil(unstakeTx)

	unStakeBlck.AddTransactions(unstakeTx)

	// Write the block to the blockchain
	tAssert.NoError(unStakeBlck.Write("test"))

	// Following test only queries contract to see if it's working.
	// Does not necessairly look into the responses.
	unstaked, err := IsStaked(stakerAddr, blockchain, executor)
	tAssert.NoError(err)
	tAssert.False(unstaked)
}

// TestIsContractUnStaked - Is a bit more complex unit test that requires to write multiple blocks
// in order to satisfy the states. It will produce 5 blocks, written into the database and as a outcome,
// staker address will be staked and removed from the sequencer list resulting in a passing test.
// Note that there has to be 2 stakers at least as minimum staker amount in the contract is 1.
func TestSlashStaker(t *testing.T) {
	tAssert := assert.New(t)

	executor, blockchain := test.NewBlockchain(t, NewVerifier(new(test.DumbActiveSequencers), hclog.Default()))
	tAssert.NotNil(executor)
	tAssert.NotNil(blockchain)

	// GET THE REQUIRED ADDRESSES

	balance := big.NewInt(0).Mul(big.NewInt(1000), ETH)

	coinbaseAddr, coinbaseSignKey := test.NewAccount(t)
	test.DepositBalance(t, coinbaseAddr, balance, blockchain, executor)

	// APPLY THE COINBASE STAKE

	bfCoinbase := block.NewBlockBuilderFactory(blockchain, executor, hclog.Default())
	blck, err := bfCoinbase.FromParentHash(blockchain.Header().Hash)
	tAssert.NoError(err)

	blck.SetCoinbaseAddress(coinbaseAddr)
	blck.SignWith(coinbaseSignKey)

	// Now lets go build the stake tx and push it to the blockchain.
	stakeTx, err := StakeTx(coinbaseAddr, 1000000)
	tAssert.NoError(err)

	blck.AddTransactions(stakeTx)
	tAssert.NoError(blck.Write("test"))

	// DO THE SLASHING

	sbfCoinbase := block.NewBlockBuilderFactory(blockchain, executor, hclog.Default())
	slashBlck, err := sbfCoinbase.FromParentHash(blockchain.Header().Hash)
	tAssert.NoError(err)

	slashBlck.SetCoinbaseAddress(coinbaseAddr)
	slashBlck.SignWith(coinbaseSignKey)

	// Now lets go build the slash tx and push it to the blockchain.
	slashTx, err := SlashStakerTx(coinbaseAddr, big.NewInt(10), 100000)
	tAssert.NoError(err)
	tAssert.NotNil(slashTx)

	slashBlck.AddTransactions(slashTx)

	// Write the block to the blockchain
	tAssert.NoError(slashBlck.Write("test"))

	// Following test only queries contract to see if it's working.
	// Does not necessairly look into the responses.
	unstaked, err := IsStaked(coinbaseAddr, blockchain, executor)
	tAssert.NoError(err)
	tAssert.True(unstaked)
}
