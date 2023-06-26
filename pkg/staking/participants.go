// Package staking contains the definitions and operations related to active participant staking mechanism in the system.
// The package provides two main types: ActiveParticipants and activeParticipantsQuerier.
// ActiveParticipants is an interface that represents the active participants of the system and their operations.
// activeParticipantsQuerier is an implementation of the ActiveParticipants interface that queries the active participants.
package staking

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/0xPolygon/polygon-edge/state"
	"github.com/0xPolygon/polygon-edge/types"
	staking_contract "github.com/availproject/avail-settlement-contracts/staking/pkg/staking"
	"github.com/availproject/op-evm/pkg/blockchain"
	"github.com/hashicorp/go-hclog"
	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/abi"
)

// DumbActiveParticipants is a struct with no behavior, intended for testing purposes.
// All its methods return nil values.
type DumbActiveParticipants struct{}

// Get method of DumbActiveParticipants struct always returns nil values.
// It satisfies the ActiveParticipants interface.
func (dasq *DumbActiveParticipants) Get(nodeType NodeType) ([]types.Address, error) { return nil, nil }

// Contains method of DumbActiveParticipants struct always returns true.
// It satisfies the ActiveParticipants interface.
func (dasq *DumbActiveParticipants) Contains(_ types.Address, nodeType NodeType) (bool, error) {
	return true, nil
}

// GetBalance method of DumbActiveParticipants struct always returns nil values.
// It satisfies the ActiveParticipants interface.
func (dasq *DumbActiveParticipants) GetBalance(_ types.Address) (*big.Int, error) {
	return nil, nil
}

// GetTotalStakedAmount method of DumbActiveParticipants struct always returns nil values.
// It satisfies the ActiveParticipants interface.
func (dasq *DumbActiveParticipants) GetTotalStakedAmount() (*big.Int, error) {
	return nil, nil
}

// InProbation method of DumbActiveParticipants struct always returns true.
// It satisfies the ActiveParticipants interface.
func (dasq *DumbActiveParticipants) InProbation(_ types.Address) (bool, error) {
	return true, nil
}

// ActiveParticipants is an interface for obtaining details about active participants in the network.
// It includes methods for getting participant addresses, checking participant existence,
// checking probation status, and getting balances.
type ActiveParticipants interface {
	Get(nodeType NodeType) ([]types.Address, error)
	Contains(addr types.Address, nodeType NodeType) (bool, error)
	InProbation(address types.Address) (bool, error)
	GetBalance(addr types.Address) (*big.Int, error)
	GetTotalStakedAmount() (*big.Int, error)
}

// activeParticipantsQuerier is a concrete implementation of the ActiveParticipants interface.
// It uses the blockchain, executor, and logger to query participant details from the blockchain.
type activeParticipantsQuerier struct {
	blockchain *blockchain.Blockchain
	executor   *state.Executor
	logger     hclog.Logger
}

// NewActiveParticipantsQuerier creates a new instance of activeParticipantsQuerier.
// It takes a blockchain, executor, and logger as parameters.
// It returns the ActiveParticipants interface.
func NewActiveParticipantsQuerier(blockchain *blockchain.Blockchain, executor *state.Executor, logger hclog.Logger) ActiveParticipants {
	return &activeParticipantsQuerier{
		blockchain: blockchain,
		executor:   executor,
		logger:     logger.Named("active_staking_participants_querier"),
	}
}

// Get method returns the addresses of active participants based on the given node type.
// It takes the nodeType parameter, which represents the type of node (Sequencer or WatchTower).
// It returns a slice of addresses and an error if the operation fails.
func (asq *activeParticipantsQuerier) Get(nodeType NodeType) ([]types.Address, error) {
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

	switch nodeType {
	case Sequencer:
		addrs, err := QueryActiveSequencers(asq.blockchain, asq.executor, transition, gasLimit, minerAddress)
		if err != nil {
			asq.logger.Error("failed to query sequencers", "error", err)
			return nil, err
		}
		return addrs, nil
	case WatchTower:
		addrs, err := QueryWatchtower(transition, gasLimit, minerAddress)
		if err != nil {
			asq.logger.Error("failed to query watchtowers", "error", err)
			return nil, err
		}
		return addrs, nil
	default:
		return nil, fmt.Errorf("failure to query participants due to node type missmatch. '%s' is not node type", nodeType)
	}
}

// Contains method checks if the given address is contained in the active participants list.
// It takes the addr parameter, which represents the address to check, and the nodeType parameter, which represents the type of node (Sequencer or WatchTower).
// It returns a boolean value indicating whether the address is found and an error if the operation fails.
func (asq *activeParticipantsQuerier) Contains(addr types.Address, nodeType NodeType) (bool, error) {
	addrs, err := asq.Get(nodeType)
	if err != nil {
		return false, err
	}

	for _, a := range addrs {
		if a == addr {
			asq.logger.Debug(fmt.Sprintf("Stake discovered no need to stake the %s.", strings.ToLower(string(nodeType))))
			return true, nil
		}
	}

	asq.logger.Debug("Staking contract address discovery information", strings.ToLower(string(nodeType)), addrs)
	asq.logger.Debug(fmt.Sprintf("Stake not discovered for '%s'. Need to stake the %s.", addr, strings.ToLower(string(nodeType))))

	return false, nil

}

// InProbation method checks if the given address is in probation.
// It takes the address parameter, which represents the address to check.
// It returns a boolean value indicating whether the address is in probation and an error if the operation fails.
func (asq *activeParticipantsQuerier) InProbation(address types.Address) (bool, error) {
	parent := asq.blockchain.Header()
	minerAddress := types.BytesToAddress(parent.Miner)

	header := &types.Header{
		ParentHash: parent.Hash,
		Number:     parent.Number + 1,
		Miner:      minerAddress.Bytes(),
		Nonce:      types.Nonce{},
		GasLimit:   parent.GasLimit,
		Timestamp:  uint64(time.Now().Unix()),
	}

	// calculate gas limit based on parent header
	gasLimit, err := asq.blockchain.CalculateGasLimit(header.Number)
	if err != nil {
		return false, err
	}

	transition, err := asq.executor.BeginTxn(parent.StateRoot, header, minerAddress)
	if err != nil {
		return false, err
	}

	probationAddrs, err := QuerySequencersInProbation(transition, gasLimit, minerAddress)
	if err != nil {
		return false, err
	}

	for _, probationAddr := range probationAddrs {
		if bytes.Equal(probationAddr.Bytes(), address.Bytes()) {
			return true, nil
		}
	}

	return false, nil
}

// GetBalance method retrieves the balance of the given address.
// It takes the address parameter, which represents the address to query.
// It returns the balance as a big.Int value and an error if the operation fails.
func (asq *activeParticipantsQuerier) GetBalance(address types.Address) (*big.Int, error) {
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

	balance, err := QueryParticipantBalance(transition, gasLimit, minerAddress, address)
	if err != nil {
		return nil, err
	}

	return balance, nil
}

// GetTotalStakedAmount method retrieves the total staked amount in the system.
// It returns the total staked amount as a big.Int value and an error if the operation fails.
func (asq *activeParticipantsQuerier) GetTotalStakedAmount() (*big.Int, error) {
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

	balance, err := QueryParticipantTotalStakedAmount(transition, gasLimit, minerAddress)
	if err != nil {
		return nil, err
	}

	return balance, nil
}

// QueryParticipants queries the current participants from the staking contract.
// It takes a transaction transition, gas limit, and the address of the sender as parameters.
// It returns a slice of addresses representing the current participants and an error if the operation fails.
func QueryParticipants(t *state.Transition, gasLimit uint64, from types.Address) ([]types.Address, error) {
	method, ok := abi.MustNewABI(staking_contract.StakingABI).Methods["GetCurrentParticipants"]
	if !ok {
		return nil, errors.New("GetCurrentParticipants method doesn't exist in Staking contract ABI")
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

	return DecodeParticipants(method, res.ReturnValue)
}

// QueryActiveSequencers queries the current active sequencers from the staking contract.
// It takes a blockchain, an executor, a transaction transition, gas limit, and the address of the sender as parameters.
// It returns a slice of addresses representing the current active sequencers and an error if the operation fails.
func QueryActiveSequencers(blockchain *blockchain.Blockchain, executor *state.Executor, t *state.Transition, gasLimit uint64, from types.Address) ([]types.Address, error) {
	toReturn := []types.Address{}

	addrs, err := QuerySequencers(t, gasLimit, from)
	if err != nil {
		return nil, err
	}

	parent := blockchain.Header()

	header := &types.Header{
		ParentHash: parent.Hash,
		Number:     parent.Number + 1,
		Miner:      from.Bytes(),
		Nonce:      types.Nonce{},
		GasLimit:   parent.GasLimit, // Inherit from parent for now, will need to adjust dynamically later.
		Timestamp:  uint64(time.Now().Unix()),
	}

	// calculate gas limit based on parent header
	probationGasLimit, err := blockchain.CalculateGasLimit(header.Number)
	if err != nil {
		return nil, err
	}

	transition, err := executor.BeginTxn(parent.StateRoot, header, from)
	if err != nil {
		return nil, err
	}

	probationAddrs, err := QuerySequencersInProbation(transition, probationGasLimit, from)
	if err != nil {
		return nil, err
	}

mainLoop:
	for _, addr := range addrs {
		for _, probationAddr := range probationAddrs {
			if addr.String() == probationAddr.String() {
				continue mainLoop
			}
		}

		toReturn = append(toReturn, addr)
	}

	return toReturn, nil
}

// QuerySequencers queries the current sequencers from the staking contract.
// It takes a transaction transition, gas limit, and the address of the sender as parameters.
// It returns a slice of addresses representing the current sequencers and an error if the operation fails.
func QuerySequencers(t *state.Transition, gasLimit uint64, from types.Address) ([]types.Address, error) {
	method, ok := abi.MustNewABI(staking_contract.StakingABI).Methods["GetCurrentSequencers"]
	if !ok {
		return nil, errors.New("GetCurrentSequencers method doesn't exist in Staking contract ABI")
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

	return DecodeParticipants(method, res.ReturnValue)
}

// QuerySequencersInProbation queries the current sequencers in probation from the staking contract.
// It takes a transaction transition, gas limit, and the address of the sender as parameters.
// It returns a slice of addresses representing the sequencers in probation and an error if the operation fails.
func QuerySequencersInProbation(t *state.Transition, gasLimit uint64, from types.Address) ([]types.Address, error) {
	method, ok := abi.MustNewABI(staking_contract.StakingABI).Methods["GetCurrentSequencersInProbation"]
	if !ok {
		return nil, errors.New("GetCurrentSequencersInProbation method doesn't exist in Staking contract ABI")
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

	return DecodeParticipants(method, res.ReturnValue)
}

// QueryWatchtower queries the current watchtowers from the staking contract.
// It takes a transaction transition, gas limit, and the address of the sender as parameters.
// It returns a slice of addresses representing the current watchtowers and an error if the operation fails.
func QueryWatchtower(t *state.Transition, gasLimit uint64, from types.Address) ([]types.Address, error) {
	method, ok := abi.MustNewABI(staking_contract.StakingABI).Methods["GetCurrentWatchtowers"]
	if !ok {
		return nil, errors.New("GetCurrentWatchtowers method doesn't exist in Staking contract ABI")
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

	return DecodeParticipants(method, res.ReturnValue)
}

// DecodeParticipants decodes the returned results from the staking contract into addresses.
// It takes a method object and the returned value as parameters.
// It returns a slice of addresses decoded from the returned value and an error if the operation fails.
func DecodeParticipants(method *abi.Method, returnValue []byte) ([]types.Address, error) {
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

// QueryParticipantBalance queries the staked amount of a participant from the staking contract.
// It takes a transaction transition, gas limit, the address of the sender, and the address of the participant as parameters.
// It returns the staked amount as a big.Int value and an error if the operation fails.
func QueryParticipantBalance(t *state.Transition, gasLimit uint64, from types.Address, addr types.Address) (*big.Int, error) {
	method, ok := abi.MustNewABI(staking_contract.StakingABI).Methods["GetCurrentAccountStakedAmount"]
	if !ok {
		return nil, errors.New("GetCurrentAccountStakedAmount method doesn't exist in Staking contract ABI")
	}

	selector := method.ID()

	encodedInput, encodeErr := method.Inputs.Encode(
		map[string]interface{}{
			"addr": addr.Bytes(),
		},
	)
	if encodeErr != nil {
		return nil, encodeErr
	}

	res, err := t.Apply(&types.Transaction{
		From:     from,
		To:       &AddrStakingContract,
		Value:    big.NewInt(0),
		Input:    append(selector, encodedInput...),
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

	return new(big.Int).SetBytes(res.ReturnValue), nil
}

// QueryParticipantTotalStakedAmount queries the total staked amount from the staking contract.
// It takes a transaction transition, gas limit, and the address of the sender as parameters.
// It returns the total staked amount as a big.Int value and an error if the operation fails.
func QueryParticipantTotalStakedAmount(t *state.Transition, gasLimit uint64, from types.Address) (*big.Int, error) {
	method, ok := abi.MustNewABI(staking_contract.StakingABI).Methods["GetCurrentStakedAmount"]
	if !ok {
		return nil, errors.New("GetCurrentStakedAmount method doesn't exist in Staking contract ABI")
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

	return new(big.Int).SetBytes(res.ReturnValue), nil
}
