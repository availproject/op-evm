package staking

import (
	"crypto/ecdsa"
	"errors"
	"math/big"

	"github.com/0xPolygon/polygon-edge/blockchain"
	"github.com/0xPolygon/polygon-edge/helper/common"
	"github.com/0xPolygon/polygon-edge/state"
	"github.com/hashicorp/go-hclog"
	"github.com/maticnetwork/avail-settlement-contracts/staking/pkg/staking"
	"github.com/maticnetwork/avail-settlement/pkg/block"
	"github.com/umbracle/ethgo/abi"

	"github.com/0xPolygon/polygon-edge/types"
)

var (
	// staking contract address
	AddrStakingContract = types.StringToAddress("0x0110000000000000000000000000000000000001")
	ETH                 = big.NewInt(1000000000000000000)

	MinSequencerCount = uint64(1)
	MaxSequencerCount = common.MaxSafeJSInt
)

func Stake(bh *blockchain.Blockchain, exec *state.Executor, sender AvailSender, logger hclog.Logger, nodeType string, stakerAddr types.Address, stakerKey *ecdsa.PrivateKey, amount *big.Int, gasLimit uint64, src string) error {
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

func UnStake(bh *blockchain.Blockchain, exec *state.Executor, sender AvailSender, logger hclog.Logger, stakerAddr types.Address, stakerKey *ecdsa.PrivateKey, gasLimit uint64, src string) error {
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

func Slash(bh *blockchain.Blockchain, exec *state.Executor, logger hclog.Logger, stakerAddr types.Address, stakerKey *ecdsa.PrivateKey, slashAddr types.Address, gasLimit uint64, src string) error {
	builder := block.NewBlockBuilderFactory(bh, exec, logger)
	blk, err := builder.FromBlockchainHead()
	if err != nil {
		return err
	}

	blk.SetCoinbaseAddress(stakerAddr)
	blk.SignWith(stakerKey)

	tx, err := SlashStakerTx(stakerAddr, slashAddr, gasLimit)
	if err != nil {
		return err
	}

	blk.AddTransactions(tx)

	// Write the block to the blockchain
	if err := blk.Write(src); err != nil {
		return err
	}

	return nil

}

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
		Value:    big.NewInt(0).Mul(big.NewInt(10), ETH), // 10 ETH
		Input:    append(selector, encodedInput...),
		GasPrice: big.NewInt(5000),
		Gas:      gasLimit,
	}

	return tx, nil
}

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

func SlashStakerTx(from types.Address, slashAddr types.Address, gasLimit uint64) (*types.Transaction, error) {
	method, ok := abi.MustNewABI(staking.StakingABI).Methods["slash"]
	if !ok {
		return nil, errors.New("unstake method doesn't exist in Staking contract ABI")
	}

	selector := method.ID()

	encodedInput, encodeErr := method.Inputs.Encode(
		map[string]interface{}{
			"slashAddr":   slashAddr.Bytes(),
			"slashAmount": big.NewInt(0).Mul(big.NewInt(1), ETH),
		},
	)
	if encodeErr != nil {
		return nil, encodeErr
	}

	tx := &types.Transaction{
		From:     from,
		To:       &AddrStakingContract,
		Value:    big.NewInt(0),
		Input:    append(selector, encodedInput...),
		GasPrice: big.NewInt(10000),
		Gas:      gasLimit,
	}

	return tx, nil
}
