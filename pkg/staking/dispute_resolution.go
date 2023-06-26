package staking

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	edge_crypto "github.com/0xPolygon/polygon-edge/crypto"
	"github.com/0xPolygon/polygon-edge/state"
	"github.com/0xPolygon/polygon-edge/types"
	staking_contract "github.com/availproject/op-evm-contracts/staking/pkg/staking"
	"github.com/availproject/op-evm/pkg/block"
	"github.com/availproject/op-evm/pkg/blockchain"
	eth_abi "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/hashicorp/go-hclog"
	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/abi"
)

// DisputeResolution defines the methods required for interacting
// with the dispute resolution smart contract. It provides functionality
// for querying and manipulating the contract's state.
type DisputeResolution interface {
	// Get retrieves the addresses of nodes of the given node type
	// that are currently under dispute.
	Get(nodeType NodeType) ([]types.Address, error)

	// Contains checks if a given address is in dispute for a specific node type.
	Contains(addr types.Address, nodeType NodeType) (bool, error)

	// GetSequencerAddr gets the address of a sequencer that's under dispute
	// for a given watchtower address.
	GetSequencerAddr(watchtowerAddr types.Address) (types.Address, error)

	// GetWatchtowerAddr gets the address of a watchtower that's under dispute
	// for a given sequencer address.
	GetWatchtowerAddr(sequencerAddr types.Address) (types.Address, error)

	// Begin initiates a dispute resolution process for a node with a specific address.
	// The process is signed with the provided private key.
	Begin(probationAddr types.Address, signKey *ecdsa.PrivateKey) error

	// End finalizes a dispute resolution process for a node with a specific address.
	// The process is signed with the provided private key.
	End(probationAddr types.Address, signKey *ecdsa.PrivateKey) error
}

// disputeResolution implements the DisputeResolution interface.
// It contains the components required to interact with the smart contract.
type disputeResolution struct {
	blockchain *blockchain.Blockchain
	executor   *state.Executor
	logger     hclog.Logger
	sender     Sender
}

// NewDisputeResolution creates a new instance of disputeResolution with the
// provided blockchain, executor, sender, and logger.
//
// Parameters:
//
//	blockchain - The blockchain instance.
//	executor - The executor instance.
//	sender - The sender instance.
//	logger - The logger instance.
//
// Returns:
//
//	A new instance of disputeResolution.
//
// Example:
//
//	dr := NewDisputeResolution(blockchain, executor, sender, logger)
func NewDisputeResolution(blockchain *blockchain.Blockchain, executor *state.Executor, sender Sender, logger hclog.Logger) DisputeResolution {
	return &disputeResolution{
		blockchain: blockchain,
		executor:   executor,
		logger:     logger.ResetNamed("staking_dispute_resolution"),
		sender:     sender,
	}
}

// Get is a method on the disputeResolution structure that retrieves the
// addresses of nodes of the given node type that are currently under dispute.
// The nodes types are sequencer and watchtower.
//
// Parameters:
//
//	nodeType - The type of the node (sequencer or watchtower).
//
// Returns:
//
//	An array of addresses that are under dispute.
//	An error if there was an issue retrieving the addresses.
//
// Example:
//
//	addresses, err := dr.Get(Sequencer)
//	if err != nil {
//	  log.Fatalf("failed to retrieve addresses: %s", err)
//	}
func (dr *disputeResolution) Get(nodeType NodeType) ([]types.Address, error) {
	parent := dr.blockchain.Header()
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
	gasLimit, err := dr.blockchain.CalculateGasLimit(header.Number)
	if err != nil {
		return nil, err
	}

	transition, err := dr.executor.BeginTxn(parent.StateRoot, header, minerAddress)
	if err != nil {
		return nil, err
	}

	switch nodeType {
	case Sequencer:
		probationAddrs, err := QuerySequencersInProbation(transition, gasLimit, minerAddress)
		if err != nil {
			return nil, err
		}
		return probationAddrs, nil
	case WatchTower:
		probationAddrs, err := QueryDisputedWatchtowers(transition, gasLimit, minerAddress)
		if err != nil {
			return nil, err
		}
		return probationAddrs, nil
	default:
		return nil, fmt.Errorf("unsuported node type provided ':%s'", nodeType)
	}
}

// Contains is a method on the disputeResolution structure that checks if a
// given address is in dispute for a specific node type.
//
// Parameters:
//
//	addr - The address to check.
//	nodeType - The type of the node (sequencer or watchtower).
//
// Returns:
//
//	A boolean indicating whether the address is in dispute.
//	An error if there was an issue checking the address.
//
// Example:
//
//	isInDispute, err := dr.Contains(addr, Sequencer)
//	if err != nil {
//	  log.Fatalf("failed to check address: %s", err)
//	}
func (dr *disputeResolution) Contains(addr types.Address, nodeType NodeType) (bool, error) {
	addrs, err := dr.Get(nodeType)
	if err != nil {
		return false, err
	}

	for _, a := range addrs {
		if a == addr {
			return true, nil
		}
	}

	return false, nil
}

// GetSequencerAddr is a method on the disputeResolution structure that gets
// the address of a sequencer that's under dispute for a given watchtower address.
//
// Parameters:
//
//	watchtowerAddr - The address of the watchtower.
//
// Returns:
//
//	The address of the disputed sequencer.
//	An error if there was an issue retrieving the address.
//
// Example:
//
//	sequencerAddr, err := dr.GetSequencerAddr(watchtowerAddr)
//	if err != nil {
//	  log.Fatalf("failed to retrieve sequencer address: %s", err)
//	}
func (dr *disputeResolution) GetSequencerAddr(watchtowerAddr types.Address) (types.Address, error) {
	parent := dr.blockchain.Header()
	minerAddress := types.BytesToAddress(parent.Miner)

	dr.logger.Info("Got addresses", "miner", minerAddress.String(), "watchtower", watchtowerAddr.String())

	header := &types.Header{
		ParentHash: parent.Hash,
		Number:     parent.Number + 1,
		Miner:      minerAddress.Bytes(),
		Nonce:      types.Nonce{},
		GasLimit:   parent.GasLimit, // Inherit from parent for now, will need to adjust dynamically later.
		Timestamp:  uint64(time.Now().Unix()),
	}

	// calculate gas limit based on parent header
	gasLimit, err := dr.blockchain.CalculateGasLimit(header.Number)
	if err != nil {
		return types.Address{}, err
	}

	transition, err := dr.executor.BeginTxn(parent.StateRoot, header, minerAddress)
	if err != nil {
		return types.Address{}, err
	}

	sequencerAddr, err := QueryDisputedSequencerAddr(transition, gasLimit, minerAddress, watchtowerAddr)
	if err != nil {
		return types.Address{}, err
	}

	return sequencerAddr, nil
}

// GetWatchtowerAddr is a method on the disputeResolution structure that gets
// the address of a watchtower that's under dispute for a given sequencer address.
//
// Parameters:
//
//	sequencerAddr - The address of the sequencer.
//
// Returns:
//
//	The address of the disputed watchtower.
//	An error if there was an issue retrieving the address.
//
// Example:
//
//	watchtowerAddr, err := dr.GetWatchtowerAddr(sequencerAddr)
//	if err != nil {
//	  log.Fatalf("failed to retrieve watchtower address: %s", err)
//	}
func (dr *disputeResolution) GetWatchtowerAddr(sequencerAddr types.Address) (types.Address, error) {
	parent := dr.blockchain.Header()
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
	gasLimit, err := dr.blockchain.CalculateGasLimit(header.Number)
	if err != nil {
		return types.Address{}, err
	}

	transition, err := dr.executor.BeginTxn(parent.StateRoot, header, minerAddress)
	if err != nil {
		return types.Address{}, err
	}

	watchtowerAddr, err := QueryDisputedWatchtowerAddr(transition, gasLimit, minerAddress, sequencerAddr)
	if err != nil {
		return types.Address{}, err
	}

	return watchtowerAddr, nil
}

// Begin is a method on the disputeResolution structure that initiates
// a dispute resolution process for a node with a specific address.
// The process is signed with the provided private key.
//
// Parameters:
//
//	probationAddr - The address of the node under dispute.
//	signKey - The private key used to sign the process.
//
// Returns:
//
//	An error if there was an issue initiating the process.
//
// Example:
//
//	err := dr.Begin(probationAddr, signKey)
//	if err != nil {
//	  log.Fatalf("failed to initiate dispute resolution process: %s", err)
//	}
func (dr *disputeResolution) Begin(probationAddr types.Address, signKey *ecdsa.PrivateKey) error {
	builder := block.NewBlockBuilderFactory(dr.blockchain, dr.executor, dr.logger)
	blk, err := builder.FromBlockchainHead()
	if err != nil {
		return err
	}

	pk := signKey.Public().(*ecdsa.PublicKey)
	address := edge_crypto.PubKeyToAddress(pk)

	blk.SetCoinbaseAddress(address)
	blk.SignWith(signKey)
	blk.SetExtraDataField(block.KeyBeginDisputeResolutionOf, probationAddr.Bytes())

	disputeResolutionTx, err := BeginDisputeResolutionTx(address, probationAddr, dr.blockchain.Header().GasLimit)
	if err != nil {
		dr.logger.Error("failed to begin new fraud dispute resolution", "error", err)
		return err
	}

	blk.AddTransactions(disputeResolutionTx)

	fBlock, err := blk.Build()
	if err != nil {
		return err
	}

	dr.logger.Info("Submitting begin dispute resolution block", "hash", fBlock.Header.Hash)
	if err := dr.sender.Send(fBlock); err != nil {
		return err
	}

	if err := dr.blockchain.WriteBlock(fBlock, "staking_fraud_dispute_resolution_modifier"); err != nil {
		return err
	}

	dr.logger.Info("Submitted begin dispute resolution block", "hash", fBlock.Header.Hash)
	return nil
}

// End is a method on the disputeResolution structure that finalizes
// a dispute resolution process for a node with a specific address.
// The process is signed with the provided private key.
//
// Parameters:
//
//	probationAddr - The address of the node under dispute.
//	signKey - The private key used to sign the process.
//
// Returns:
//
//	An error if there was an issue finalizing the process.
//
// Example:
//
//	err := dr.End(probationAddr, signKey)
//	if err != nil {
//	  log.Fatalf("failed to finalize dispute resolution process: %s", err)
//	}
func (dr *disputeResolution) End(probationAddr types.Address, signKey *ecdsa.PrivateKey) error {
	builder := block.NewBlockBuilderFactory(dr.blockchain, dr.executor, dr.logger)
	blk, err := builder.FromBlockchainHead()
	if err != nil {
		return err
	}

	pk := signKey.Public().(*ecdsa.PublicKey)
	address := edge_crypto.PubKeyToAddress(pk)

	blk.SetCoinbaseAddress(address)
	blk.SignWith(signKey)

	disputeResolutionTx, err := EndDisputeResolutionTx(address, probationAddr, dr.blockchain.Header().GasLimit)
	if err != nil {
		dr.logger.Error("failed to end new fraud dispute resolution", "error", err)
		return err
	}

	blk.AddTransactions(disputeResolutionTx)

	fBlock, err := blk.Build()
	if err != nil {
		return err
	}

	if err := dr.sender.Send(fBlock); err != nil {
		return err
	}

	if err := dr.blockchain.WriteBlock(fBlock, "staking_fraud_dispute_resolution_modifier"); err != nil {
		return err
	}

	return nil
}

// BeginDisputeResolutionTx constructs a transaction to initiate the dispute resolution process on the Staking contract.
//
// It takes in the initiator's address, the probation address and a gas limit.
//
// The function generates a selector for the BeginDisputeResolution method from the Staking contract ABI.
// Then, it encodes the probation address as an input for this method and includes this data in the newly created transaction.
//
// This transaction can be submitted to the network to initiate the dispute resolution process.
//
// Parameters:
//
//	from - The address of the transaction initiator.
//	probationAddr - The address of the probation sequencer.
//	gasLimit - The gas limit for the transaction.
//
// Returns:
//
//	A pointer to the newly created transaction.
//	An error if there was an issue creating the transaction.
//
// Example:
//
//	tx, err := BeginDisputeResolutionTx(fromAddress, probationAddress, 50000)
//	if err != nil {
//	  log.Fatalf("failed to create dispute resolution transaction: %s", err)
//	}
func BeginDisputeResolutionTx(from types.Address, probationAddr types.Address, gasLimit uint64) (*types.Transaction, error) {
	method, ok := abi.MustNewABI(staking_contract.StakingABI).Methods["BeginDisputeResolution"]
	if !ok {
		panic("BeginDisputeResolution method doesn't exist in Staking contract ABI. Contract is broken.")
	}

	selector := method.ID()

	encodedInput, encodeErr := method.Inputs.Encode(
		map[string]interface{}{
			"sequencerAddr": probationAddr.Bytes(),
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

// IsBeginDisputeResolutionTx checks if the given transaction is a dispute resolution initiation transaction.
//
// This is done by decoding the input data in the transaction and comparing it with the expected method selector and parameters.
//
// Parameters:
//
//	tx - The transaction to check.
//
// Returns:
//
//	true if the transaction is a dispute resolution initiation transaction, false otherwise.
//	An error if there was an issue checking the transaction.
//
// Example:
//
//	isDisputeTx, err := IsBeginDisputeResolutionTx(tx)
//	if err != nil {
//	  log.Fatalf("failed to check dispute resolution transaction: %s", err)
//	}
func IsBeginDisputeResolutionTx(tx *types.Transaction) (bool, error) {
	stakingAbi, err := eth_abi.JSON(strings.NewReader(staking_contract.StakingMetaData.ABI))
	if err != nil {
		panic(fmt.Sprintf("Failed to resolve staking contract abi: %s", err))
	}

	// Make sure not to process tx as it's not really ready, however DO NOT return error as it's spamming the hell
	// out of the stdout
	if tx == nil || len(tx.Input) < 4 {
		return false, nil
	}

	method, err := stakingAbi.MethodById(tx.Input[:4])
	if err == nil && method != nil && method.RawName == "BeginDisputeResolution" {
		return true, nil
	}

	return false, nil
}

// EndDisputeResolutionTx constructs a transaction to conclude the dispute resolution process on the Staking contract.
//
// Similarly to BeginDisputeResolutionTx, it creates a transaction which includes the EndDisputeResolution method selector and the encoded input parameters.
//
// Parameters:
//
//	from - The address of the transaction initiator.
//	probationAddr - The address of the probation sequencer.
//	gasLimit - The gas limit for the transaction.
//
// Returns:
//
//	A pointer to the newly created transaction.
//	An error if there was an issue creating the transaction.
//
// Example:
//
//	tx, err := EndDisputeResolutionTx(fromAddress, probationAddress, 50000)
//	if err != nil {
//	  log.Fatalf("failed to create dispute resolution conclusion transaction: %s", err)
//	}
func EndDisputeResolutionTx(from types.Address, probationAddr types.Address, gasLimit uint64) (*types.Transaction, error) {
	method, ok := abi.MustNewABI(staking_contract.StakingABI).Methods["EndDisputeResolution"]
	if !ok {
		panic("EndDisputeResolution method doesn't exist in Staking contract ABI. Contract is broken.")
	}

	selector := method.ID()

	encodedInput, encodeErr := method.Inputs.Encode(
		map[string]interface{}{
			"sequencerAddr": probationAddr.Bytes(),
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

// QueryDisputedSequencerAddr calls the GetDisputedSequencerAddrs method on the Staking contract.
//
// It sends a transaction to query the contract and returns the result.
//
// Parameters:
//
//	t - The state transition object.
//	gasLimit - The gas limit for the transaction.
//	from - The address of the query initiator.
//	watchtowerAddr - The address of the watchtower.
//
// Returns:
//
//	The disputed sequencer address.
//	An error if there was an issue querying the address.
//
// Example:
//
//	disputedSequencer, err := QueryDisputedSequencerAddr(transition, 50000, fromAddress, watchtowerAddress)
//	if err != nil {
//	  log.Fatalf("failed to query disputed sequencer address: %s", err)
//	}
func QueryDisputedSequencerAddr(t *state.Transition, gasLimit uint64, from types.Address, watchtowerAddr types.Address) (types.Address, error) {
	method, ok := abi.MustNewABI(staking_contract.StakingABI).Methods["GetDisputedSequencerAddrs"]
	if !ok {
		return types.Address{}, errors.New("GetDisputedSequencerAddrs method doesn't exist in Staking contract ABI")
	}

	encodedInput, encodeErr := method.Inputs.Encode(
		map[string]interface{}{
			"watchtowerAddr": watchtowerAddr.Bytes(),
		},
	)
	if encodeErr != nil {
		return types.Address{}, encodeErr
	}

	selector := method.ID()
	res, err := t.Apply(&types.Transaction{
		From:     from,
		To:       &AddrStakingContract,
		Value:    big.NewInt(0),
		Input:    append(selector, encodedInput...),
		GasPrice: big.NewInt(5000),
		Gas:      gasLimit,
		Nonce:    t.GetNonce(from),
	})

	if err != nil {
		return types.Address{}, err
	}

	if res.Failed() {
		return types.Address{}, res.Err
	}

	decodedResults, err := method.Outputs.Decode(res.ReturnValue)
	if err != nil {
		return types.Address{}, err
	}

	results, ok := decodedResults.(map[string]interface{})
	if !ok {
		return types.Address{}, errors.New("failed type assertion from decodedResults to map")
	}

	address, _ := results["0"].(ethgo.Address)
	return types.Address(address), nil
}

// QueryDisputedWatchtowerAddr calls the GetDisputedWatchtowerAddr method on the Staking contract.
//
// It sends a transaction to query the contract and returns the result.
//
// Parameters:
//
//	t - The state transition object.
//	gasLimit - The gas limit for the transaction.
//	from - The address of the query initiator.
//	sequencerAddr - The address of the sequencer.
//
// Returns:
//
//	The disputed watchtower address.
//	An error if there was an issue querying the address.
//
// Example:
//
//	disputedWatchtower, err := QueryDisputedWatchtowerAddr(transition, 50000, fromAddress, sequencerAddress)
//	if err != nil {
//	  log.Fatalf("failed to query disputed watchtower address: %s", err)
//	}
func QueryDisputedWatchtowerAddr(t *state.Transition, gasLimit uint64, from types.Address, sequencerAddr types.Address) (types.Address, error) {
	method, ok := abi.MustNewABI(staking_contract.StakingABI).Methods["GetDisputedWatchtowerAddr"]
	if !ok {
		return types.Address{}, errors.New("GetDisputedWatchtowerAddr method doesn't exist in Staking contract ABI")
	}

	encodedInput, encodeErr := method.Inputs.Encode(
		map[string]interface{}{
			"sequencerAddr": sequencerAddr.Bytes(),
		},
	)
	if encodeErr != nil {
		return types.Address{}, encodeErr
	}

	selector := method.ID()
	res, err := t.Apply(&types.Transaction{
		From:     from,
		To:       &AddrStakingContract,
		Value:    big.NewInt(0),
		Input:    append(selector, encodedInput...),
		GasPrice: big.NewInt(5000),
		Gas:      gasLimit,
		Nonce:    t.GetNonce(from),
	})

	if err != nil {
		return types.Address{}, err
	}

	if res.Failed() {
		return types.Address{}, res.Err
	}

	decodedResults, err := method.Outputs.Decode(res.ReturnValue)
	if err != nil {
		return types.Address{}, err
	}

	results, ok := decodedResults.(map[string]interface{})
	if !ok {
		return types.Address{}, errors.New("failed type assertion from decodedResults to map")
	}

	address, _ := results["0"].(ethgo.Address)
	return types.Address(address), nil
}

// QueryDisputedWatchtowers calls the GetCurrentDisputeWatchtowers method on the Staking contract.
//
// It sends a transaction to query the contract and returns the result.
//
// Parameters:
//
//	t - The state transition object.
//	gasLimit - The gas limit for the transaction.
//	from - The address of the query initiator.
//
// Returns:
//
//	An array of disputed watchtower addresses.
//	An error if there was an issue querying the addresses.
//
// Example:
//
//	disputedWatchtowers, err := QueryDisputedWatchtowers(transition, 50000, fromAddress)
//	if err != nil {
//	  log.Fatalf("failed to query disputed watchtower addresses: %s", err)
//	}
func QueryDisputedWatchtowers(t *state.Transition, gasLimit uint64, from types.Address) ([]types.Address, error) {
	method, ok := abi.MustNewABI(staking_contract.StakingABI).Methods["GetCurrentDisputeWatchtowers"]
	if !ok {
		return nil, errors.New("GetCurrentDisputeWatchtowers method doesn't exist in Staking contract ABI")
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
