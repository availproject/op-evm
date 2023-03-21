package staking

import (
	"crypto/ecdsa"
	"math/big"

	"github.com/0xPolygon/polygon-edge/blockchain"
	edge_crypto "github.com/0xPolygon/polygon-edge/crypto"
	"github.com/0xPolygon/polygon-edge/state"
	"github.com/hashicorp/go-hclog"
)

type NodeType string

const (
	Sequencer  NodeType = "sequencer"
	WatchTower NodeType = "watchtower"
)

type Node interface {
	ShouldStake(pkey *ecdsa.PrivateKey) bool
	Stake(amount *big.Int, pkey *ecdsa.PrivateKey) error
	UnStake(pkey *ecdsa.PrivateKey) error
}

type node struct {
	blockchain *blockchain.Blockchain
	executor   *state.Executor
	logger     hclog.Logger
	nodeType   NodeType
	sender     Sender
}

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

func (n *node) Stake(amount *big.Int, pkey *ecdsa.PrivateKey) error {
	pk := pkey.Public().(*ecdsa.PublicKey)
	address := edge_crypto.PubKeyToAddress(pk)
	gasLimit := uint64(1_000_000)
	return Stake(
		n.blockchain, n.executor, n.sender, n.logger, string(n.nodeType),
		address, pkey, amount, gasLimit, string(n.nodeType),
	)
}

func (n *node) UnStake(pkey *ecdsa.PrivateKey) error {
	pk := pkey.Public().(*ecdsa.PublicKey)
	address := edge_crypto.PubKeyToAddress(pk)
	gasLimit := uint64(1_000_000)
	return UnStake(
		n.blockchain, n.executor, n.sender, n.logger, address, pkey,
		gasLimit, string(n.nodeType),
	)
}

func NewNode(blockchain *blockchain.Blockchain, executor *state.Executor, sender Sender, logger hclog.Logger, nodeType NodeType) Node {
	return &node{
		blockchain: blockchain,
		executor:   executor,
		logger:     logger.ResetNamed("staking_node"),
		nodeType:   nodeType,
		sender:     sender,
	}
}
