package staking

import (
	"crypto/ecdsa"
	"errors"
	"math/big"

	"github.com/0xPolygon/polygon-edge/helper/common"
	"github.com/0xPolygon/polygon-edge/state"
	"github.com/availproject/op-evm-contracts/staking/pkg/staking"
	"github.com/availproject/op-evm/pkg/block"
	"github.com/availproject/op-evm/pkg/blockchain"
	commontoken "github.com/availproject/op-evm/pkg/common"
	"github.com/hashicorp/go-hclog"
	"github.com/umbracle/ethgo/abi"

	"github.com/0xPolygon/polygon-edge/types"
)

// AddrStakingContract represents the staking contract address.
var AddrStakingContract = types.StringToAddress("0x0110000000000000000000000000000000000001")

// MinSequencerCount is the minimum number of sequencers required.
var MinSequencerCount = uint64(1)

// MaxSequencerCount is the maximum number of sequencers allowed.
var MaxSequencerCount = common.MaxSafeJSInt

// Stake stakes the specified amount for the given staker address and node type.
// It builds a block, signs it with the staker's key, adds the stake transaction,
// sends the block to the sender, and writes the block to the blockchain.
func Stake(bh *blockchain.Blockchain, exec *state.Executor, sender Sender, logger hclog.Logger, nodeType string, stakerAddr types.Address, stakerKey *ecdsa.PrivateKey, amount *big.Int, gasLimit uint64, src string) error {
	builder := block.NewBlockBuilderFactory(bh, exec, logger)
	blk, err := builder.FromBlockchainHead()
	if err != nil {
		return err
	}

	blk.SetCoinbaseAddress(stakerAddr)
	blk.SignWith(stakerKey)

	tx, err := StakeTx(stakerAddr, amount, nodeType, gasLimit)
	if err != nil {
		return err
	}

	blk.AddTransactions(tx)

	fBlock, err := blk.Build()
	if err != nil {
		return err
	}

	if err := sender.Send(fBlock); err != nil {
		return err
	}

	if err := bh.WriteBlock(fBlock, src); err != nil {
		return err
	}

	return nil
}

// UnStake unstakes the given staker address.
// It builds a block, signs it with the staker's key, adds the unstake transaction,
// sends the block to the sender, and writes the block to the blockchain.
func UnStake(bh *blockchain.Blockchain, exec *state.Executor, sender Sender, logger hclog.Logger, stakerAddr types.Address, stakerKey *ecdsa.PrivateKey, gasLimit uint64, src string) error {
	builder := block.NewBlockBuilderFactory(bh, exec, logger)
	blk, err := builder.FromBlockchainHead()
	if err != nil {
		return err
	}

	blk.SetCoinbaseAddress(stakerAddr)
	blk.SignWith(stakerKey)

	tx, err := UnStakeTx(stakerAddr, gasLimit)
	if err != nil {
		return err
	}

	blk.AddTransactions(tx)

	fBlock, err := blk.Build()
	if err != nil {
		return err
	}

	if err := sender.Send(fBlock); err != nil {
		return err
	}

	if err := bh.WriteBlock(fBlock, src); err != nil {
		return err
	}

	return nil
}

// Slash slashes the malicious staker address by the active sequencer.
// It builds a block, signs it with the active sequencer's key, adds the slash transaction,
// and writes the block to the blockchain.
func Slash(bh *blockchain.Blockchain, exec *state.Executor, logger hclog.Logger, activeSequencerAddr types.Address, activeSequencerSignKey *ecdsa.PrivateKey, maliciousStakerAddr types.Address, gasLimit uint64, nodeType string) error {
	builder := block.NewBlockBuilderFactory(bh, exec, logger)
	blk, err := builder.FromBlockchainHead()
	if err != nil {
		return err
	}

	blk.SetCoinbaseAddress(activeSequencerAddr)
	blk.SignWith(activeSequencerSignKey)

	tx, err := SlashStakerTx(activeSequencerAddr, maliciousStakerAddr, gasLimit)
	if err != nil {
		return err
	}

	blk.AddTransactions(tx)

	// Write the block to the blockchain
	if err := blk.Write(nodeType); err != nil {
		return err
	}

	return nil
}

// StakeTx returns a stake transaction with the specified parameters.
func StakeTx(from types.Address, amount *big.Int, nodeType string, gasLimit uint64) (*types.Transaction, error) {
	method, ok := abi.MustNewABI(staking.StakingABI).Methods["stake"]
	if !ok {
		return nil, errors.New("stake method doesn't exist in Staking contract ABI")
	}

	selector := method.ID()

	encodedInput, encodeErr := method.Inputs.Encode(
		map[string]interface{}{
			"nodeType": nodeType,
		},
	)
	if encodeErr != nil {
		return nil, encodeErr
	}

	tx := &types.Transaction{
		From:     from,
		To:       &AddrStakingContract,
		Value:    big.NewInt(0).Mul(big.NewInt(10), commontoken.ETH), // 10 ETH
		Input:    append(selector, encodedInput...),
		GasPrice: big.NewInt(5000),
		Gas:      gasLimit,
	}

	return tx, nil
}

// UnStakeTx returns an unstake transaction for the specified address.
func UnStakeTx(from types.Address, gasLimit uint64) (*types.Transaction, error) {
	method, ok := abi.MustNewABI(staking.StakingABI).Methods["unstake"]
	if !ok {
		return nil, errors.New("unstake method doesn't exist in Staking contract ABI")
	}

	selector := method.ID()

	tx := &types.Transaction{
		From:     from,
		To:       &AddrStakingContract,
		Value:    big.NewInt(0),
		Input:    selector,
		GasPrice: big.NewInt(50000),
		Gas:      gasLimit,
	}

	return tx, nil
}

// SlashStakerTx returns a slash transaction to slash the malicious staker address.
func SlashStakerTx(activeSequencerAddr types.Address, maliciousStakerAddr types.Address, gasLimit uint64) (*types.Transaction, error) {
	method, ok := abi.MustNewABI(staking.StakingABI).Methods["slash"]
	if !ok {
		return nil, errors.New("Slash method doesn't exist in Staking contract ABI")
	}

	selector := method.ID()

	encodedInput, encodeErr := method.Inputs.Encode(
		map[string]interface{}{
			"slashAddr": maliciousStakerAddr.Bytes(),
		},
	)
	if encodeErr != nil {
		return nil, encodeErr
	}

	tx := &types.Transaction{
		From:     activeSequencerAddr,
		To:       &AddrStakingContract,
		Value:    big.NewInt(0),
		Input:    append(selector, encodedInput...),
		GasPrice: big.NewInt(10000),
		Gas:      gasLimit,
	}

	return tx, nil
}
