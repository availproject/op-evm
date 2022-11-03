package staking

import (
	"crypto/ecdsa"
	"errors"
	"math/big"
	"time"

	"github.com/0xPolygon/polygon-edge/blockchain"
	edge_crypto "github.com/0xPolygon/polygon-edge/crypto"
	"github.com/0xPolygon/polygon-edge/state"
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/hashicorp/go-hclog"
	staking_contract "github.com/maticnetwork/avail-settlement/contracts/staking"
	"github.com/maticnetwork/avail-settlement/pkg/block"
	"github.com/umbracle/ethgo/abi"
)

type SequencerRate interface {
	CurrentMinimum() (*big.Int, error)
	CurrentMaximum() (*big.Int, error)
	SetMinimum(newMin *big.Int, signKey *ecdsa.PrivateKey) error
	SetMaximum(newMin *big.Int, signKey *ecdsa.PrivateKey) error
}

type sequencerRate struct {
	blockchain *blockchain.Blockchain
	executor   *state.Executor
	logger     hclog.Logger
}

func NewSequencerRater(blockchain *blockchain.Blockchain, executor *state.Executor, logger hclog.Logger) SequencerRate {
	return &sequencerRate{
		blockchain: blockchain,
		executor:   executor,
		logger:     logger.ResetNamed("staking_rate_sequencers"),
	}
}

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

	gasLimit, err := st.blockchain.CalculateGasLimit(st.blockchain.Header().Number)
	if err != nil {
		return err
	}

	setThresholdTx, setThresholdTxErr := SetMinimumSequencersTx(address, newMin, gasLimit)
	if setThresholdTxErr != nil {
		st.logger.Error("failed to set new minimum sequencers count", "err", setThresholdTxErr)
		return err
	}

	blk.AddTransactions(setThresholdTx)

	// Write the block to the blockchain
	if err := blk.Write("staking_rate_sequencers_modifier"); err != nil {
		return err
	}

	return nil
}

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

	gasLimit, err := st.blockchain.CalculateGasLimit(st.blockchain.Header().Number)
	if err != nil {
		return err
	}

	setThresholdTx, setThresholdTxErr := SetMaximumSequencersTx(address, newMin, gasLimit)
	if setThresholdTxErr != nil {
		st.logger.Error("failed to set new maximum sequencers count", "err", setThresholdTxErr)
		return err
	}

	blk.AddTransactions(setThresholdTx)

	// Write the block to the blockchain
	if err := blk.Write("staking_rate_sequencers_modifier"); err != nil {
		return err
	}

	return nil
}

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

	// calculate gas limit based on parent header
	gasLimit, err := st.blockchain.CalculateGasLimit(header.Number)
	if err != nil {
		return nil, err
	}

	transition, err := st.executor.BeginTxn(parent.StateRoot, header, minerAddress)
	if err != nil {
		return nil, err
	}

	threshold, err := GetMinimumSequencersTx(transition, gasLimit, minerAddress)
	if err != nil {
		st.logger.Error("failed to query current minimum sequencers allowed", "err", err)
		return nil, err
	}

	return threshold, nil
}

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

	// calculate gas limit based on parent header
	gasLimit, err := st.blockchain.CalculateGasLimit(header.Number)
	if err != nil {
		return nil, err
	}

	transition, err := st.executor.BeginTxn(parent.StateRoot, header, minerAddress)
	if err != nil {
		return nil, err
	}

	threshold, err := GetMaximumSequencersTx(transition, gasLimit, minerAddress)
	if err != nil {
		st.logger.Error("failed to query current maximum sequencers allowed", "err", err)
		return nil, err
	}

	return threshold, nil
}

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
