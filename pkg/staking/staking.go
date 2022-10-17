package staking

import (
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/0xPolygon/polygon-edge/blockchain"
	"github.com/0xPolygon/polygon-edge/helper/common"
	"github.com/0xPolygon/polygon-edge/state"
	"github.com/maticnetwork/avail-settlement/contracts/staking"
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

func StakeTx(from types.Address, gasLimit uint64) (*types.Transaction, error) {
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

func SlashStakerTx(from types.Address, ethValue *big.Int, gasLimit uint64) (*types.Transaction, error) {
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
		GasPrice: big.NewInt(10000),
		Gas:      gasLimit,
	}

	return tx, nil
}

func IsStaked(addr types.Address, bc *blockchain.Blockchain, exec *state.Executor) (bool, error) {
	parent := bc.Header()

	header := &types.Header{
		ParentHash: parent.Hash,
		Number:     parent.Number + 1,
		Miner:      addr.Bytes(),
		Nonce:      types.Nonce{},
		GasLimit:   parent.GasLimit, // Inherit from parent for now, will need to adjust dynamically later.
		Timestamp:  uint64(time.Now().Unix()),
	}

	// calculate gas limit based on parent header
	gasLimit, err := bc.CalculateGasLimit(header.Number)
	if err != nil {
		return false, err
	}

	transition, err := exec.BeginTxn(parent.StateRoot, header, addr)
	if err != nil {
		return false, err
	}

	addrs, err := QuerySequencers(transition, gasLimit, addr)
	if err != nil {
		return false, err
	}

	fmt.Printf("Requested staked addr: %v - available addrs: %v \n", addr.String(), addrs)

	for _, sequencerAddr := range addrs {
		if sequencerAddr.String() == addr.String() {
			return true, nil
		}
	}

	return false, nil
}
