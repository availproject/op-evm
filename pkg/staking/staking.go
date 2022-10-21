package staking

import (
	"crypto/ecdsa"
	"errors"
	"math/big"

	"github.com/0xPolygon/polygon-edge/blockchain"
	"github.com/0xPolygon/polygon-edge/helper/common"
	"github.com/0xPolygon/polygon-edge/state"
	"github.com/hashicorp/go-hclog"
	"github.com/maticnetwork/avail-settlement/contracts/staking"
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

func Stake(bh *blockchain.Blockchain, exec *state.Executor, logger hclog.Logger, nodeType string, stakerAddr types.Address, stakerKey *ecdsa.PrivateKey, amount *big.Int, gasLimit uint64, src string) error {
	builder := block.NewBlockBuilderFactory(bh, exec, logger)
	blck, err := builder.FromParentHash(bh.Header().Hash)
	if err != nil {
		return err
	}

	blck.SetCoinbaseAddress(stakerAddr)
	blck.SignWith(stakerKey)

	stakeTx, err := StakeTx(stakerAddr, amount, nodeType, gasLimit)
	if err != nil {
		return err
	}

	blck.AddTransactions(stakeTx)

	// Write the block to the blockchain
	if err := blck.Write(src); err != nil {
		return err
	}

	return nil

}

func UnStake(bh *blockchain.Blockchain, exec *state.Executor, logger hclog.Logger, stakerAddr types.Address, stakerKey *ecdsa.PrivateKey, gasLimit uint64, src string) error {
	builder := block.NewBlockBuilderFactory(bh, exec, logger)
	blck, err := builder.FromParentHash(bh.Header().Hash)
	if err != nil {
		return err
	}

	blck.SetCoinbaseAddress(stakerAddr)
	blck.SignWith(stakerKey)

	stakeTx, err := UnStakeTx(stakerAddr, gasLimit)
	if err != nil {
		return err
	}

	blck.AddTransactions(stakeTx)

	// Write the block to the blockchain
	if err := blck.Write(src); err != nil {
		return err
	}

	return nil

}

func Slash(bh *blockchain.Blockchain, exec *state.Executor, logger hclog.Logger, stakerAddr types.Address, stakerKey *ecdsa.PrivateKey, gasLimit uint64, src string) error {
	builder := block.NewBlockBuilderFactory(bh, exec, logger)
	blck, err := builder.FromParentHash(bh.Header().Hash)
	if err != nil {
		return err
	}

	blck.SetCoinbaseAddress(stakerAddr)
	blck.SignWith(stakerKey)

	stakeTx, err := SlashStakerTx(stakerAddr, gasLimit)
	if err != nil {
		return err
	}

	blck.AddTransactions(stakeTx)

	// Write the block to the blockchain
	if err := blck.Write(src); err != nil {
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
		Value:    amount, // big.NewInt(0).Mul(big.NewInt(10), ETH), // 10 ETH
		Input:    append(selector, encodedInput...),
		GasPrice: big.NewInt(50000),
		//V:        big.NewInt(1), // it is necessary to encode in rlp,
		Gas: gasLimit,
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

func SlashStakerTx(from types.Address, gasLimit uint64) (*types.Transaction, error) {
	method, ok := abi.MustNewABI(staking.StakingABI).Methods["slash"]
	if !ok {
		return nil, errors.New("unstake method doesn't exist in Staking contract ABI")
	}

	selector := method.ID()

	tx := &types.Transaction{
		From:     from,
		To:       &AddrStakingContract,
		Value:    big.NewInt(0),
		Input:    selector,
		GasPrice: big.NewInt(10000),
		Gas:      gasLimit,
	}

	return tx, nil
}
