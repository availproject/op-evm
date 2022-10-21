package staking

import (
	"crypto/ecdsa"
	"errors"
	"math/big"
	"time"

	"github.com/0xPolygon/polygon-edge/blockchain"
	"github.com/0xPolygon/polygon-edge/state"
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/hashicorp/go-hclog"
	staking_contract "github.com/maticnetwork/avail-settlement/contracts/staking"
	"github.com/maticnetwork/avail-settlement/pkg/block"
	"github.com/umbracle/ethgo/abi"
)

type StakingThreshold interface {
	Current() (*big.Int, error)
	Set(newAmount *big.Int) error
	SetAddress(address types.Address)
	SetSignKey(key *ecdsa.PrivateKey)
}

type stakingThreshold struct {
	blockchain *blockchain.Blockchain
	executor   *state.Executor
	logger     hclog.Logger
	address    types.Address
	signKey    *ecdsa.PrivateKey
}

func NewStakingThresholdQuerier(blockchain *blockchain.Blockchain, executor *state.Executor, logger hclog.Logger) StakingThreshold {
	return &stakingThreshold{
		blockchain: blockchain,
		executor:   executor,
		logger:     logger.ResetNamed("staking_threshold_querier"),
	}
}

func (st *stakingThreshold) SetAddress(address types.Address) {
	st.address = address
}

func (st *stakingThreshold) SetSignKey(key *ecdsa.PrivateKey) {
	st.signKey = key
}

func (st *stakingThreshold) Set(newAmount *big.Int) error {
	builder := block.NewBlockBuilderFactory(st.blockchain, st.executor, st.logger)
	blck, err := builder.FromParentHash(st.blockchain.Header().Hash)
	if err != nil {
		return err
	}

	blck.SetCoinbaseAddress(st.address)
	blck.SignWith(st.signKey)

	gasLimit, err := st.blockchain.CalculateGasLimit(st.blockchain.Header().Number)
	if err != nil {
		return err
	}

	setThresholdTx, setThresholdTxErr := SetStakingThresholdTx(st.address, newAmount, gasLimit)
	if setThresholdTxErr != nil {
		st.logger.Error("failed to query current staking threshold", "err", setThresholdTxErr)
		return err
	}

	blck.AddTransactions(setThresholdTx)

	// Write the block to the blockchain
	if err := blck.Write("staking_threshold_modifier"); err != nil {
		return err
	}

	return nil
}

func (st *stakingThreshold) Current() (*big.Int, error) {
	parent := st.blockchain.Header()
	minerAddress := types.BytesToAddress(parent.Miner)

	header := &types.Header{
		ParentHash: parent.Hash,
		Number:     parent.Number + 1,
		Miner:      minerAddress.Bytes(),
		Nonce:      types.Nonce{},
		GasLimit:   parent.GasLimit, // Inherit from parent for now, will need to adjust dynamically later.
		Timestamp:  uint64(time.Now().Unix()),
	}

	// calculate gas limit based on parent header
	gasLimit, err := st.blockchain.CalculateGasLimit(header.Number)
	if err != nil {
		return nil, err
	}

	transition, err := st.executor.BeginTxn(parent.StateRoot, header, minerAddress)
	if err != nil {
		return nil, err
	}

	threshold, err := GetStakingThresholdTx(transition, gasLimit, minerAddress)
	if err != nil {
		st.logger.Error("failed to query current staking threshold", "err", err)
		return nil, err
	}

	return threshold, nil
}

func SetStakingThresholdTx(from types.Address, amount *big.Int, gasLimit uint64) (*types.Transaction, error) {
	method, ok := abi.MustNewABI(staking_contract.StakingABI).Methods["SetStakingMinThreshold"]
	if !ok {
		return nil, errors.New("SetStakingMinThreshold method doesn't exist in Staking contract ABI")
	}

	selector := method.ID()

	encodedInput, encodeErr := method.Inputs.Encode(
		map[string]interface{}{
			"newThreshold": amount,
		},
	)
	if encodeErr != nil {
		return nil, encodeErr
	}

	return &types.Transaction{
		From:     from,
		To:       &AddrStakingContract,
		Value:    big.NewInt(0),
		Input:    append(selector, encodedInput...),
		GasPrice: big.NewInt(5000),
		Gas:      gasLimit,
	}, nil
}

func GetStakingThresholdTx(t *state.Transition, gasLimit uint64, from types.Address) (*big.Int, error) {
	method, ok := abi.MustNewABI(staking_contract.StakingABI).Methods["GetCurrentStakingThreshold"]
	if !ok {
		return nil, errors.New("GetCurrentStakingThreshold method doesn't exist in Staking contract ABI")
	}

	selector := method.ID()
	res, err := t.Apply(&types.Transaction{
		From:     from,
		To:       &AddrStakingContract,
		Value:    big.NewInt(0),
		Input:    selector,
		GasPrice: big.NewInt(0),
		Gas:      gasLimit,
		Nonce:    t.GetNonce(from),
	})

	if err != nil {
		return nil, err
	}

	if res.Failed() {
		return nil, res.Err
	}

	toReturn := new(big.Int)
	toReturn.SetBytes(res.ReturnValue)
	return toReturn, nil
}
