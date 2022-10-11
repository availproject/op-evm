package staking

import (
	"errors"
	"math/big"
	"time"

	"github.com/0xPolygon/polygon-edge/blockchain"
	"github.com/0xPolygon/polygon-edge/state"
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/hashicorp/go-hclog"
	staking_contract "github.com/maticnetwork/avail-settlement/contracts/staking"
	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/abi"
)

type ActiveSequencers interface {
	Get() ([]types.Address, error)
	Contains(addr types.Address) (bool, error)
}

type activeSequencersQuerier struct {
	blockchain *blockchain.Blockchain
	executor   *state.Executor
	logger     hclog.Logger
}

func NewActiveSequencersQuerier(blockchain *blockchain.Blockchain, executor *state.Executor, logger hclog.Logger) ActiveSequencers {
	return &activeSequencersQuerier{
		blockchain: blockchain,
		executor:   executor,
		logger:     logger.ResetNamed("active_sequencer_querier"),
	}
}

func (asq *activeSequencersQuerier) Get() ([]types.Address, error) {
	parent := asq.blockchain.Header()
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
	gasLimit, err := asq.blockchain.CalculateGasLimit(header.Number)
	if err != nil {
		return nil, err
	}

	transition, err := asq.executor.BeginTxn(parent.StateRoot, header, minerAddress)
	if err != nil {
		return nil, err
	}

	addrs, err := QuerySequencers(transition, gasLimit, minerAddress)
	if err != nil {
		asq.logger.Error("failed to query sequencers", "err", err)
		return nil, err
	}

	return addrs, nil
}

func (asq *activeSequencersQuerier) Contains(addr types.Address) (bool, error) {
	addrs, err := asq.Get()
	if err != nil {
		return false, err
	}

	for _, a := range addrs {
		if a == addr {
			asq.logger.Info("Sequencer stake discovered no need to stake the sequencer.")
			return true, nil
		}
	}

	asq.logger.Info("Staking contract address discovery information", "sequencers", addrs)
	asq.logger.Warn("Sequencer stake not discovered. Need to stake the sequencer.")

	return false, nil

}

func QuerySequencers(t *state.Transition, gasLimit uint64, from types.Address) ([]types.Address, error) {
	method, ok := abi.MustNewABI(staking_contract.StakingABI).Methods["CurrentSequencers"]
	if !ok {
		return nil, errors.New("sequencers method doesn't exist in Staking contract ABI")
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

	return DecodeSequencers(method, res.ReturnValue)
}

func DecodeSequencers(method *abi.Method, returnValue []byte) ([]types.Address, error) {
	decodedResults, err := method.Outputs.Decode(returnValue)
	if err != nil {
		return nil, err
	}

	results, ok := decodedResults.(map[string]interface{})
	if !ok {
		return nil, errors.New("failed type assertion from decodedResults to map")
	}

	web3Addresses, ok := results["0"].([]ethgo.Address)

	if !ok {
		return nil, errors.New("failed type assertion from results[0] to []ethgo.Address")
	}

	addresses := make([]types.Address, len(web3Addresses))
	for idx, waddr := range web3Addresses {
		addresses[idx] = types.Address(waddr)
	}

	return addresses, nil
}
