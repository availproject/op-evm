package staking

import (
	"crypto/ecdsa"
	"errors"
	"math/big"
	"time"

	edge_crypto "github.com/0xPolygon/polygon-edge/crypto"
	"github.com/0xPolygon/polygon-edge/state"
	"github.com/0xPolygon/polygon-edge/types"
	staking_contract "github.com/availproject/avail-settlement-contracts/staking/pkg/staking"
	"github.com/availproject/op-evm/pkg/block"
	"github.com/availproject/op-evm/pkg/blockchain"
	"github.com/hashicorp/go-hclog"
	"github.com/umbracle/ethgo/abi"
)

// Threshold represents an interface for managing the staking threshold.
type Threshold interface {
	// Set sets the new staking threshold value.
	// It takes the new amount and a signing key as parameters.
	// An error is returned if the operation fails.
	Set(newAmount *big.Int, signKey *ecdsa.PrivateKey) error

	// Current returns the current staking threshold value.
	// It retrieves the value from the staking system and returns it as a *big.Int.
	// An error is returned if the operation fails.
	Current() (*big.Int, error)
}

type threshold struct {
	blockchain *blockchain.Blockchain
	executor   *state.Executor
	logger     hclog.Logger
}

// NewStakingThresholdQuerier returns a new instance of the staking threshold querier.
func NewStakingThresholdQuerier(blockchain *blockchain.Blockchain, executor *state.Executor, logger hclog.Logger) Threshold {
	return &threshold{
		blockchain: blockchain,
		executor:   executor,
		logger:     logger.ResetNamed("staking_threshold_querier"),
	}
}

// Set sets the new staking threshold value.
func (st *threshold) Set(newAmount *big.Int, signKey *ecdsa.PrivateKey) error {
	builder := block.NewBlockBuilderFactory(st.blockchain, st.executor, st.logger)
	blk, err := builder.FromBlockchainHead()
	if err != nil {
		return err
	}

	pk := signKey.Public().(*ecdsa.PublicKey)
	address := edge_crypto.PubKeyToAddress(pk)

	blk.SetCoinbaseAddress(address)
	blk.SignWith(signKey)

	setThresholdTx, setThresholdTxErr := SetThresholdTx(address, newAmount, st.blockchain.Header().GasLimit)
	if setThresholdTxErr != nil {
		st.logger.Error("failed to query current staking threshold", "error", setThresholdTxErr)
		return err
	}

	blk.AddTransactions(setThresholdTx)

	// Write the block to the blockchain
	if err := blk.Write("staking_threshold_modifier"); err != nil {
		return err
	}

	return nil
}

// Current returns the current staking threshold value.
func (st *threshold) Current() (*big.Int, error) {
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

	transition, err := st.executor.BeginTxn(parent.StateRoot, header, minerAddress)
	if err != nil {
		return nil, err
	}

	threshold, err := GetThresholdTx(transition, header.GasLimit, minerAddress)
	if err != nil {
		st.logger.Error("failed to query current staking threshold", "error", err)
		return nil, err
	}

	return threshold, nil
}

// SetThresholdTx returns a transaction to set the staking threshold.
func SetThresholdTx(from types.Address, amount *big.Int, gasLimit uint64) (*types.Transaction, error) {
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

// GetThresholdTx returns the current staking threshold from the transition state.
func GetThresholdTx(t *state.Transition, gasLimit uint64, from types.Address) (*big.Int, error) {
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
