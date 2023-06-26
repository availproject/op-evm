package staking

import (
	"crypto/ecdsa"
	"errors"
	"math/big"
	"time"

	edge_crypto "github.com/0xPolygon/polygon-edge/crypto"
	"github.com/0xPolygon/polygon-edge/state"
	"github.com/0xPolygon/polygon-edge/types"
	staking_contract "github.com/availproject/op-evm-contracts/staking/pkg/staking"
	"github.com/availproject/op-evm/pkg/block"
	"github.com/availproject/op-evm/pkg/blockchain"
	"github.com/hashicorp/go-hclog"
	"github.com/umbracle/ethgo/abi"
)

// SequencerRate is an interface for managing the minimum and maximum number of sequencers.
type SequencerRate interface {
	// CurrentMinimum returns the current minimum number of sequencers.
	// It retrieves the value from the staking system and returns it as a *big.Int.
	// An error is returned if the operation fails.
	CurrentMinimum() (*big.Int, error)

	// CurrentMaximum returns the current maximum number of sequencers.
	// It retrieves the value from the staking system and returns it as a *big.Int.
	// An error is returned if the operation fails.
	CurrentMaximum() (*big.Int, error)

	// SetMinimum sets the minimum number of sequencers.
	// It takes the new minimum value and a signing key as parameters.
	// An error is returned if the operation fails.
	SetMinimum(newMin *big.Int, signKey *ecdsa.PrivateKey) error

	// SetMaximum sets the maximum number of sequencers.
	// It takes the new maximum value and a signing key as parameters.
	// An error is returned if the operation fails.
	SetMaximum(newMin *big.Int, signKey *ecdsa.PrivateKey) error
}

// sequencerRate is a concrete implementation of the SequencerRate interface.
// It uses a blockchain, executor, and logger to manage the sequencer rates.
type sequencerRate struct {
	blockchain *blockchain.Blockchain
	executor   *state.Executor
	logger     hclog.Logger
}

// NewSequencerRater creates a new instance of sequencerRate.
// It takes a blockchain, executor, and logger as parameters and returns a SequencerRate interface.
func NewSequencerRater(blockchain *blockchain.Blockchain, executor *state.Executor, logger hclog.Logger) SequencerRate {
	return &sequencerRate{
		blockchain: blockchain,
		executor:   executor,
		logger:     logger.ResetNamed("staking_rate_sequencers"),
	}
}

// SetMinimum sets the minimum number of sequencers.
// It takes the new minimum value and a signing key as parameters.
// It returns an error if the operation fails.
func (st *sequencerRate) SetMinimum(newMin *big.Int, signKey *ecdsa.PrivateKey) error {
	builder := block.NewBlockBuilderFactory(st.blockchain, st.executor, st.logger)
	blk, err := builder.FromBlockchainHead()
	if err != nil {
		return err
	}

	pk := signKey.Public().(*ecdsa.PublicKey)
	address := edge_crypto.PubKeyToAddress(pk)

	blk.SetCoinbaseAddress(address)
	blk.SignWith(signKey)

	setThresholdTx, setThresholdTxErr := SetMinimumSequencersTx(address, newMin, st.blockchain.Header().GasLimit)
	if setThresholdTxErr != nil {
		st.logger.Error("failed to set new minimum sequencers count", "error", setThresholdTxErr)
		return err
	}

	blk.AddTransactions(setThresholdTx)

	// Write the block to the blockchain
	if err := blk.Write("staking_rate_sequencers_modifier"); err != nil {
		return err
	}

	return nil
}

// SetMaximum sets the maximum number of sequencers.
// It takes the new maximum value and a signing key as parameters.
// It returns an error if the operation fails.
func (st *sequencerRate) SetMaximum(newMin *big.Int, signKey *ecdsa.PrivateKey) error {
	builder := block.NewBlockBuilderFactory(st.blockchain, st.executor, st.logger)
	blk, err := builder.FromBlockchainHead()
	if err != nil {
		return err
	}

	pk := signKey.Public().(*ecdsa.PublicKey)
	address := edge_crypto.PubKeyToAddress(pk)

	blk.SetCoinbaseAddress(address)
	blk.SignWith(signKey)

	setThresholdTx, setThresholdTxErr := SetMaximumSequencersTx(address, newMin, st.blockchain.Header().GasLimit)
	if setThresholdTxErr != nil {
		st.logger.Error("failed to set new maximum sequencers count", "error", setThresholdTxErr)
		return err
	}

	blk.AddTransactions(setThresholdTx)

	// Write the block to the blockchain
	if err := blk.Write("staking_rate_sequencers_modifier"); err != nil {
		return err
	}

	return nil
}

// CurrentMinimum returns the current minimum number of sequencers.
// It retrieves the value from the staking contract and returns it as a big.Int.
// It returns an error if the operation fails.
func (st *sequencerRate) CurrentMinimum() (*big.Int, error) {
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

	threshold, err := GetMinimumSequencersTx(transition, header.GasLimit, minerAddress)
	if err != nil {
		st.logger.Error("failed to query current minimum sequencers allowed", "error", err)
		return nil, err
	}

	return threshold, nil
}

// CurrentMaximum returns the current maximum number of sequencers.
// It retrieves the value from the staking contract and returns it as a big.Int.
// It returns an error if the operation fails.
func (st *sequencerRate) CurrentMaximum() (*big.Int, error) {
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

	threshold, err := GetMaximumSequencersTx(transition, header.GasLimit, minerAddress)
	if err != nil {
		st.logger.Error("failed to query current maximum sequencers allowed", "error", err)
		return nil, err
	}

	return threshold, nil
}

// SetMinimumSequencersTx creates a transaction to set the minimum number of sequencers.
// It takes the sender address, the new minimum value, and the gas limit as parameters.
// It returns the transaction and an error if the operation fails.
func SetMinimumSequencersTx(from types.Address, amount *big.Int, gasLimit uint64) (*types.Transaction, error) {
	method, ok := abi.MustNewABI(staking_contract.StakingABI).Methods["SetMinNumSequencers"]
	if !ok {
		return nil, errors.New("SetMinNumSequencers method doesn't exist in Staking contract ABI")
	}

	selector := method.ID()

	encodedInput, encodeErr := method.Inputs.Encode(
		map[string]interface{}{
			"minimumNumSequencers": amount,
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

// GetMinimumSequencersTx retrieves the current minimum number of sequencers from the staking contract.
// It takes a transaction transition, gas limit, and the address of the sender as parameters.
// It returns the current minimum value as a big.Int and an error if the operation fails.
func GetMinimumSequencersTx(t *state.Transition, gasLimit uint64, from types.Address) (*big.Int, error) {
	method, ok := abi.MustNewABI(staking_contract.StakingABI).Methods["GetMinNumSequencers"]
	if !ok {
		return nil, errors.New("GetMinNumSequencers method doesn't exist in Staking contract ABI")
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

// SetMaximumSequencersTx creates a transaction to set the maximum number of sequencers.
// It takes the sender address, the new maximum value, and the gas limit as parameters.
// It returns the transaction and an error if the operation fails.
func SetMaximumSequencersTx(from types.Address, amount *big.Int, gasLimit uint64) (*types.Transaction, error) {
	method, ok := abi.MustNewABI(staking_contract.StakingABI).Methods["SetMaxNumSequencers"]
	if !ok {
		return nil, errors.New("SetMaxNumSequencers method doesn't exist in Staking contract ABI")
	}

	selector := method.ID()

	encodedInput, encodeErr := method.Inputs.Encode(
		map[string]interface{}{
			"maximumNumSequencers": amount,
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

// GetMaximumSequencersTx retrieves the current maximum number of sequencers from the staking contract.
// It takes a transaction transition, gas limit, and the address of the sender as parameters.
// It returns the current maximum value as a big.Int and an error if the operation fails.
func GetMaximumSequencersTx(t *state.Transition, gasLimit uint64, from types.Address) (*big.Int, error) {
	method, ok := abi.MustNewABI(staking_contract.StakingABI).Methods["GetMaxNumSequencers"]
	if !ok {
		return nil, errors.New("GetMaxNumSequencers method doesn't exist in Staking contract ABI")
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
