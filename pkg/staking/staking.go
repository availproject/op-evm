package staking

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/0xPolygon/polygon-edge/helper/common"
	"github.com/0xPolygon/polygon-edge/state"
	"github.com/maticnetwork/avail-settlement/contracts/staking"
	"github.com/umbracle/ethgo/abi"

	"github.com/0xPolygon/polygon-edge/types"
)

var (
	// staking contract address
	AddrStakingContract = types.StringToAddress("0x0110000000000000000000000000000000000001")

	MinSequencerCount = uint64(1)
	MaxSequencerCount = common.MaxSafeJSInt
)

func ApplyStakeTx(txn *state.Transition, from types.Address, gasLimit uint64) (*types.Transaction, error) {
	method, ok := abi.MustNewABI(staking.StakingABI).Methods["stake"]
	if !ok {
		return nil, errors.New("stake method doesn't exist in Staking contract ABI")
	}

	selector := method.ID()

	tx := &types.Transaction{
		From:     from,
		To:       &AddrStakingContract,
		Value:    big.NewInt(0).Mul(big.NewInt(10), ETH), // 10 ETH
		Input:    selector,
		GasPrice: big.NewInt(5000),
		Gas:      gasLimit,
		Nonce:    txn.GetNonce(from),
	}

	res, err := txn.Apply(tx)

	if err != nil {
		return nil, err
	}

	if res.Failed() {
		return nil, res.Err
	}

	return tx, nil
}

func ApplyUnStakeTx(txn *state.Transition, from types.Address, gasLimit uint64) (*types.Transaction, error) {
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
		Nonce:    txn.GetNonce(from),
	}

	res, err := txn.Apply(tx)

	if err != nil {
		return nil, err
	}

	if res.Failed() {
		fmt.Printf("failure while unstaking: %s - err: %s \n", res.ReturnValue, res.Err)
		return nil, res.Err
	}

	return tx, nil
}

func ApplySlashStakerTx(txn *state.Transition, from types.Address, ethValue *big.Int, gasLimit uint64) (*types.Transaction, error) {
	method, ok := abi.MustNewABI(staking.StakingABI).Methods["slash"]
	if !ok {
		return nil, errors.New("unstake method doesn't exist in Staking contract ABI")
	}

	selector := method.ID()

	tx := &types.Transaction{
		From:     from,
		To:       &AddrStakingContract,
		Value:    big.NewInt(0), // 10 ETH
		Input:    selector,
		GasPrice: big.NewInt(100000),
		Gas:      gasLimit,
		Nonce:    txn.GetNonce(from),
	}

	res, err := txn.Apply(tx)
	if err != nil {
		fmt.Printf("failure while attempt to slash - err: %s \n", err)
		return nil, err
	}

	if res.Failed() {
		fmt.Printf("failure while slashing: %+v - %s - err: %s \n", res, res.ReturnValue, res.Err)
		return nil, res.Err
	}

	return tx, nil
}
