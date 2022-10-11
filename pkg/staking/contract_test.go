package staking

import (
	"log"
	"math/big"
	"testing"

	"github.com/0xPolygon/polygon-edge/blockchain"
	"github.com/0xPolygon/polygon-edge/chain"
	"github.com/0xPolygon/polygon-edge/state"
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/hashicorp/go-hclog"
	"github.com/maticnetwork/avail-settlement/consensus/avail/verifier"
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

// First we need to write block that contains the balance. Otherwise it wont allow
// writing the block as account is not yet written into the database, resulting in rejection
// due to not being able to pay for the gas cost.
func getCoinbaseTransition(t *testing.T, tAssert *assert.Assertions, bch *blockchain.Blockchain, exec *state.Executor) (test.BlockFactory, types.Address) {
	coinbaseAddr, signKey := test.NewAccount(t)
	balance := big.NewInt(0).Mul(big.NewInt(1000), ETH)

	bf := test.NewBasicBlockFactory(t, exec, coinbaseAddr, signKey)

	parentBlock, parentBlockFound := bch.GetBlockByHash(bch.Header().Hash, false)
	tAssert.True(parentBlockFound)

	transition, err := bf.GetTransition(parentBlock)
	tAssert.NoError(err)

	transition.SetAccountDirectly(coinbaseAddr, &chain.GenesisAccount{
		Balance: balance,
	})

	blockAccount := bf.BuildBlockWithTransition(parentBlock, transition, []*types.Transaction{})
	tAssert.NoError(bch.WriteBlock(blockAccount, "test"))

	return bf, coinbaseAddr
}

func TestIsContractStaked(t *testing.T) {
	tAssert := assert.New(t)

	executor, blockchain := test.NewBlockchain(t, verifier.NewVerifier(hclog.Default()))
	tAssert.NotNil(executor)
	tAssert.NotNil(blockchain)

	bf, coinbaseAddr := getCoinbaseTransition(t, tAssert, blockchain, executor)

	coinbaseBlock, coinbaseBlockFound := blockchain.GetBlockByHash(blockchain.Header().Hash, false)
	tAssert.True(coinbaseBlockFound)

	coinbaseTransition, err := bf.GetTransition(coinbaseBlock)
	tAssert.NoError(err)

	// Now lets go build the stake tx and push it to the blockchain.
	stakeTx, err := ApplyStakeTx(coinbaseTransition, coinbaseAddr, 1000000)
	tAssert.NoError(err)

	block := bf.BuildBlock(coinbaseBlock, []*types.Transaction{stakeTx})

	// Write the block to the blockchain
	tAssert.NoError(blockchain.WriteBlock(block, "test"))

	// Following test only queries contract to see if coinbase address is in staked set.
	staked, err := IsStaked(coinbaseAddr, blockchain, executor)
	tAssert.NoError(err)
	tAssert.True(staked)
}

// TestIsContractUnStaked - Is a bit more complex unit test that requires to write multiple blocks
// in order to satisfy the states. It will produce 5 blocks, written into the database and as a outcome,
// staker address will be staked and removed from the sequencer list resulting in a passing test.
// Note that there has to be 2 stakers at least as minimum staker amount in the contract is 1.
func TestIsContractUnStaked(t *testing.T) {
	tAssert := assert.New(t)

	executor, blockchain := test.NewBlockchain(t, verifier.NewVerifier(hclog.Default()))
	tAssert.NotNil(executor)
	tAssert.NotNil(blockchain)

	/** GET THE REQUIRED ADDRESSES **/

	bfCoinbase, coinbaseAddr := getCoinbaseTransition(t, tAssert, blockchain, executor)
	bfStaker, stakerAddr := getCoinbaseTransition(t, tAssert, blockchain, executor)

	/** APPLY THE COINBASE STAKE **/

	coinbaseBlock, coinbaseBlockFound := blockchain.GetBlockByHash(blockchain.Header().Hash, false)
	tAssert.True(coinbaseBlockFound)

	coinbaseTransition, err := bfCoinbase.GetTransition(coinbaseBlock)
	tAssert.NoError(err)

	// Now lets go build the stake tx and push it to the blockchain.
	stakeTx, err := ApplyStakeTx(coinbaseTransition, coinbaseAddr, 1000000)
	tAssert.NoError(err)

	block := bfCoinbase.BuildBlock(coinbaseBlock, []*types.Transaction{stakeTx})

	// Write the block to the blockchain
	tAssert.NoError(blockchain.WriteBlock(block, "test"))

	/** APPLY THE STAKER STAKE - THIS ONE WILL BE REMOVED **/

	coinbaseStakerBlock, coinbaseStakerBlockFound := blockchain.GetBlockByHash(blockchain.Header().Hash, false)
	tAssert.True(coinbaseStakerBlockFound)

	coinbaseStakerTransition, err := bfCoinbase.GetTransition(coinbaseStakerBlock)
	tAssert.NoError(err)

	// Now lets go build the stake tx and push it to the blockchain.
	stakerTx, err := ApplyStakeTx(coinbaseStakerTransition, stakerAddr, 1000000)
	tAssert.NoError(err)

	block = bfCoinbase.BuildBlock(coinbaseStakerBlock, []*types.Transaction{stakerTx})

	// Write the block to the blockchain
	tAssert.NoError(blockchain.WriteBlock(block, "test"))

	// Following test only queries contract to see if it's working.
	// Does not necessairly look into the responses.
	staked, err := IsStaked(coinbaseAddr, blockchain, executor)
	tAssert.NoError(err)
	tAssert.True(staked)

	/** DO THE UNSTAKE **/

	stakeBlock, stakeBlockFound := blockchain.GetBlockByHash(blockchain.Header().Hash, false)
	tAssert.True(stakeBlockFound)

	stakeTransition, err := bfCoinbase.GetTransition(stakeBlock)
	tAssert.NoError(err)

	// Make sure that staker and the contract contain enough of the funds
	coinbaseBalance := stakeTransition.GetBalance(coinbaseAddr)
	tAssert.NotZero(coinbaseBalance.Int64())

	stakingContractBalance := stakeTransition.GetBalance(AddrStakingContract)
	tAssert.NotZero(stakingContractBalance.Int64())

	// Now lets go build the stake tx and push it to the blockchain.
	unstakeTx, err := ApplyUnStakeTx(stakeTransition, stakerAddr, 1000000)
	tAssert.NoError(err)
	tAssert.NotNil(unstakeTx)

	unstakeBlock := bfStaker.BuildBlock(stakeBlock, []*types.Transaction{unstakeTx})

	// Write the block to the blockchain
	tAssert.NoError(blockchain.WriteBlock(unstakeBlock, "test"))

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

	executor, blockchain := test.NewBlockchain(t, verifier.NewVerifier(hclog.Default()))
	tAssert.NotNil(executor)
	tAssert.NotNil(blockchain)

	/** GET THE REQUIRED ADDRESSES **/

	bfCoinbase, coinbaseAddr := getCoinbaseTransition(t, tAssert, blockchain, executor)
	//_, stakerAddr := getCoinbaseTransition(t, tAssert, blockchain, executor)

	/** APPLY THE COINBASE STAKE **/

	coinbaseBlock, coinbaseBlockFound := blockchain.GetBlockByHash(blockchain.Header().Hash, false)
	tAssert.True(coinbaseBlockFound)

	coinbaseTransition, err := bfCoinbase.GetTransition(coinbaseBlock)
	tAssert.NoError(err)

	// Now lets go build the stake tx and push it to the blockchain.
	stakeTx, err := ApplyStakeTx(coinbaseTransition, coinbaseAddr, 1000000)
	tAssert.NoError(err)

	block := bfCoinbase.BuildBlock(coinbaseBlock, []*types.Transaction{stakeTx})

	// Write the block to the blockchain
	tAssert.NoError(blockchain.WriteBlock(block, "test"))

	/** DO THE SLASHING **/

	slashBlock, slashBlockFound := blockchain.GetBlockByHash(blockchain.Header().Hash, false)
	tAssert.True(slashBlockFound)

	stakeTransition, err := bfCoinbase.GetTransition(slashBlock)
	tAssert.NoError(err)

	// Make sure that staker and the contract contain enough of the funds
	coinbaseBalance := stakeTransition.GetBalance(coinbaseAddr)
	log.Printf("Coinbase balance (after-staking) but prior slash: %s wei", coinbaseBalance.String())
	tAssert.NotZero(coinbaseBalance)

	stakingContractBalance := stakeTransition.GetBalance(AddrStakingContract)
	tAssert.NotZero(stakingContractBalance)

	// Now lets go build the slash tx and push it to the blockchain.
	slashTx, err := ApplySlashStakerTx(stakeTransition, coinbaseAddr, big.NewInt(10), 100000)
	tAssert.NoError(err)
	tAssert.NotNil(slashTx)

	// Make sure that staker and the contract contain enough of the funds
	coinbaseBalanceAfterSlash := stakeTransition.GetBalance(coinbaseAddr)
	log.Printf("Coinbase balance (after-staking) and after slash: %s wei", coinbaseBalanceAfterSlash.String())

	newSlashBlock := bfCoinbase.BuildBlock(slashBlock, []*types.Transaction{slashTx})

	// Write the block to the blockchain
	tAssert.NoError(blockchain.WriteBlock(newSlashBlock, "test"))

	// Following test only queries contract to see if it's working.
	// Does not necessairly look into the responses.
	unstaked, err := IsStaked(coinbaseAddr, blockchain, executor)
	tAssert.NoError(err)
	tAssert.True(unstaked)
}
