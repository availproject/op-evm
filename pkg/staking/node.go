package staking

import (
	"crypto/ecdsa"
	"math/big"

	edge_crypto "github.com/0xPolygon/polygon-edge/crypto"
	"github.com/0xPolygon/polygon-edge/state"
	"github.com/hashicorp/go-hclog"
	"github.com/maticnetwork/avail-settlement/pkg/blockchain"
)

// NodeType represents a type of node, either Sequencer or Watchtower.
type NodeType string

// These constants represent the different types of nodes.
const (
	Sequencer  NodeType = "sequencer"
	WatchTower NodeType = "watchtower"
)

// Node interface represents the staking-related operations a node can perform.
type Node interface {
	ShouldStake(pkey *ecdsa.PrivateKey) bool
	Stake(amount *big.Int, pkey *ecdsa.PrivateKey) error
	UnStake(pkey *ecdsa.PrivateKey) error
}

// node structure represents a specific node on the network, containing
// blockchain, executor, logger, nodeType and sender instances.
type node struct {
	blockchain *blockchain.Blockchain
	executor   *state.Executor
	logger     hclog.Logger
	nodeType   NodeType
	sender     Sender
}

// ShouldStake is a method on the node structure that determines if the node should stake.
// The decision is based on whether the node is already staked.
//
// Parameters:
//
//	pkey - The private key of the node.
//
// Returns:
//
//	A boolean indicating whether the node should stake.
func (n *node) ShouldStake(pkey *ecdsa.PrivateKey) bool {
	participantsQuerier := NewActiveParticipantsQuerier(n.blockchain, n.executor, n.logger)

	pk := pkey.Public().(*ecdsa.PublicKey)
	address := edge_crypto.PubKeyToAddress(pk)

	staked, err := participantsQuerier.Contains(address, n.nodeType)
	if err != nil {
		n.logger.Error("failed to check if sequencer is staked", "error", err)
		return false
	}

	return !staked
}

// Stake is a method on the node structure that stakes a specific amount for the node.
// The staking transaction is signed with the provided private key.
//
// Parameters:
//
//	amount - The amount to stake.
//	pkey - The private key used to sign the staking transaction.
//
// Returns:
//
//	An error if there was an issue staking the amount.
func (n *node) Stake(amount *big.Int, pkey *ecdsa.PrivateKey) error {
	pk := pkey.Public().(*ecdsa.PublicKey)
	address := edge_crypto.PubKeyToAddress(pk)
	gasLimit := uint64(1_000_000)
	return Stake(
		n.blockchain, n.executor, n.sender, n.logger, string(n.nodeType),
		address, pkey, amount, gasLimit, string(n.nodeType),
	)
}

// UnStake is a method on the node structure that unstakes the node.
// The unstaking transaction is signed with the provided private key.
//
// Parameters:
//
//	pkey - The private key used to sign the unstaking transaction.
//
// Returns:
//
//	An error if there was an issue unstaking.
func (n *node) UnStake(pkey *ecdsa.PrivateKey) error {
	pk := pkey.Public().(*ecdsa.PublicKey)
	address := edge_crypto.PubKeyToAddress(pk)
	gasLimit := uint64(1_000_000)
	return UnStake(
		n.blockchain, n.executor, n.sender, n.logger, address, pkey,
		gasLimit, string(n.nodeType),
	)
}

// NewNode creates a new instance of node with the provided blockchain, executor,
// sender, logger, and node type.
//
// Parameters:
//
//	blockchain - The blockchain instance.
//	executor - The executor instance.
//	sender - The sender instance.
//	logger - The logger instance.
//	nodeType - The type of the node (sequencer or watchtower).
//
// Returns:
//
//	A new instance of node.
//
// Example:
//
//	n := NewNode(blockchain, executor, sender, logger, Sequencer)
func NewNode(blockchain *blockchain.Blockchain, executor *state.Executor, sender Sender, logger hclog.Logger, nodeType NodeType) Node {
	return &node{
		blockchain: blockchain,
		executor:   executor,
		logger:     logger.ResetNamed("staking_node"),
		nodeType:   nodeType,
		sender:     sender,
	}
}
