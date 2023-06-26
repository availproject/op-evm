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

// WatchtowerRate is an interface for managing the minimum and maximum number of watchtowers.
type WatchtowerRate interface {
	// CurrentMinimum returns the current minimum number of watchtowers.
	// It retrieves the value from the staking system and returns it as a *big.Int.
	// An error is returned if the operation fails.
	CurrentMinimum() (*big.Int, error)

	// CurrentMaximum returns the current maximum number of watchtowers.
	// It retrieves the value from the staking system and returns it as a *big.Int.
	// An error is returned if the operation fails.
	CurrentMaximum() (*big.Int, error)

	// SetMinimum sets the minimum number of watchtowers.
	// It takes the new minimum value and a signing key as parameters.
	// An error is returned if the operation fails.
	SetMinimum(newMin *big.Int, signKey *ecdsa.PrivateKey) error

	// SetMaximum sets the maximum number of watchtowers.
	// It takes the new maximum value and a signing key as parameters.
	// An error is returned if the operation fails.
	SetMaximum(newMax *big.Int, signKey *ecdsa.PrivateKey) error
}

type watchtowerRate struct {
	blockchain *blockchain.Blockchain
	executor   *state.Executor
	logger     hclog.Logger
}

// NewWatchtowerRater creates a new instance of the WatchtowerRate interface.
func NewWatchtowerRater(blockchain *blockchain.Blockchain, executor *state.Executor, logger hclog.Logger) WatchtowerRate {
	return &watchtowerRate{
		blockchain: blockchain,
		executor:   executor,
		logger:     logger.ResetNamed("staking_rate_watchtowers"),
	}
}

// SetMinimum sets the minimum number of watchtowers.
// It takes the new minimum value and a signing key as parameters.
// An error is returned if the operation fails.
func (st *watchtowerRate) SetMinimum(newMin *big.Int, signKey *ecdsa.PrivateKey) error {
	builder := block.NewBlockBuilderFactory(st.blockchain, st.executor, st.logger)
	blk, err := builder.FromBlockchainHead()
	if err != nil {
		return err
	}

	pk := signKey.Public().(*ecdsa.PublicKey)
	address := edge_crypto.PubKeyToAddress(pk)

	blk.SetCoinbaseAddress(address)
	blk.SignWith(signKey)

	setThresholdTx, setThresholdTxErr := SetMinimumWatchtowersTx(address, newMin, st.blockchain.Header().GasLimit)
	if setThresholdTxErr != nil {
		st.logger.Error("failed to set new minimum watchtowers count", "error", setThresholdTxErr)
		return err
	}

	blk.AddTransactions(setThresholdTx)

	// Write the block to the blockchain
	if err := blk.Write("staking_rate_watchtowers_modifier"); err != nil {
		return err
	}

	return nil
}

// SetMaximum sets the maximum number of watchtowers.
// It takes the new maximum value and a signing key as parameters.
// An error is returned if the operation fails.
func (st *watchtowerRate) SetMaximum(newMax *big.Int, signKey *ecdsa.PrivateKey) error {
	builder := block.NewBlockBuilderFactory(st.blockchain, st.executor, st.logger)
	blk, err := builder.FromBlockchainHead()
	if err != nil {
		return err
	}

	pk := signKey.Public().(*ecdsa.PublicKey)
	address := edge_crypto.PubKeyToAddress(pk)

	blk.SetCoinbaseAddress(address)
	blk.SignWith(signKey)

	setThresholdTx, setThresholdTxErr := SetMaximumWatchtowersTx(address, newMax, st.blockchain.Header().GasLimit)
	if setThresholdTxErr != nil {
		st.logger.Error("failed to set new maximum watchtowers count", "error", setThresholdTxErr)
		return err
	}

	blk.AddTransactions(setThresholdTx)

	// Write the block to the blockchain
	if err := blk.Write("staking_rate_watchtowers_modifier"); err != nil {
		return err
	}

	return nil
}

// CurrentMinimum returns the current minimum number of watchtowers.
// It retrieves the value from the staking system and returns it as a *big.Int.
// An error is returned if the operation fails.
func (st *watchtowerRate) CurrentMinimum() (*big.Int, error) {
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

	threshold, err := GetMinimumWatchtowersTx(transition, header.GasLimit, minerAddress)
	if err != nil {
		st.logger.Error("failed to query current minimum watchtowers allowed", "error", err)
		return nil, err
	}

	return threshold, nil
}

// CurrentMaximum returns the current maximum number of watchtowers.
// It retrieves the value from the staking system and returns it as a *big.Int.
// An error is returned if the operation fails.
func (st *watchtowerRate) CurrentMaximum() (*big.Int, error) {
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

	threshold, err := GetMaximumWatchtowersTx(transition, header.GasLimit, minerAddress)
	if err != nil {
		st.logger.Error("failed to query current maximum watchtowers allowed", "error", err)
		return nil, err
	}

	return threshold, nil
}

// SetMinimumWatchtowersTx creates a transaction to set the minimum number of watchtowers.
// It takes the sender address, the new minimum value, and the gas limit as parameters.
// The transaction is returned or an error if it fails.
func SetMinimumWatchtowersTx(from types.Address, amount *big.Int, gasLimit uint64) (*types.Transaction, error) {
	method, ok := abi.MustNewABI(staking_contract.StakingABI).Methods["SetMinNumWatchtowers"]
	if !ok {
		return nil, errors.New("SetMinNumWatchtowers method doesn't exist in Staking contract ABI")
	}

	selector := method.ID()

	encodedInput, encodeErr := method.Inputs.Encode(
		map[string]interface{}{
			"minimumNumWatchtowers": amount,
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

// GetMinimumWatchtowersTx retrieves the minimum number of watchtowers from the staking system.
// It takes a state transition, gas limit, and sender address as parameters.
// The minimum number of watchtowers is returned as a *big.Int or an error if the operation fails.
func GetMinimumWatchtowersTx(t *state.Transition, gasLimit uint64, from types.Address) (*big.Int, error) {
	method, ok := abi.MustNewABI(staking_contract.StakingABI).Methods["GetMinNumWatchtowers"]
	if !ok {
		return nil, errors.New("GetMinNumWatchtowers method doesn't exist in Staking contract ABI")
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

// SetMaximumWatchtowersTx creates a transaction to set the maximum number of watchtowers.
// It takes the sender address, the new maximum value, and the gas limit as parameters.
// The transaction is returned or an error if it fails.
func SetMaximumWatchtowersTx(from types.Address, amount *big.Int, gasLimit uint64) (*types.Transaction, error) {
	method, ok := abi.MustNewABI(staking_contract.StakingABI).Methods["SetMaxNumWatchtowers"]
	if !ok {
		return nil, errors.New("SetMaxNumWatchtowers method doesn't exist in Staking contract ABI")
	}

	selector := method.ID()

	encodedInput, encodeErr := method.Inputs.Encode(
		map[string]interface{}{
			"maximumNumWatchtowers": amount,
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

// GetMaximumWatchtowersTx retrieves the maximum number of watchtowers from the staking system.
// It takes a state transition, gas limit, and sender address as parameters.
// The maximum number of watchtowers is returned as a *big.Int or an error if the operation fails.
func GetMaximumWatchtowersTx(t *state.Transition, gasLimit uint64, from types.Address) (*big.Int, error) {
	method, ok := abi.MustNewABI(staking_contract.StakingABI).Methods["GetMaxNumWatchtowers"]
	if !ok {
		return nil, errors.New("GetMaxNumWatchtowers method doesn't exist in Staking contract ABI")
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
