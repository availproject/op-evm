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
	staking_contract "github.com/maticnetwork/avail-settlement-contracts/staking/pkg/staking"
	"github.com/maticnetwork/avail-settlement/pkg/block"
	"github.com/umbracle/ethgo/abi"
)

type Threshold interface {
	Current() (*big.Int, error)
	Set(newAmount *big.Int, signKey *ecdsa.PrivateKey) error
}

type threshold struct {
	blockchain *blockchain.Blockchain
	executor   *state.Executor
	logger     hclog.Logger
}

func NewStakingThresholdQuerier(blockchain *blockchain.Blockchain, executor *state.Executor, logger hclog.Logger) Threshold {
	return &threshold{
		blockchain: blockchain,
		executor:   executor,
		logger:     logger.ResetNamed("staking_threshold_querier"),
	}
}

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
		st.logger.Error("failed to query current staking threshold", "err", setThresholdTxErr)
		return err
	}

	blk.AddTransactions(setThresholdTx)

	// Write the block to the blockchain
	if err := blk.Write("staking_threshold_modifier"); err != nil {
		return err
	}

	return nil
}

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
		st.logger.Error("failed to query current staking threshold", "err", err)
		return nil, err
	}

	return threshold, nil
}

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
