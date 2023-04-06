package avail

import (
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/maticnetwork/avail-settlement/pkg/avail"
	"github.com/maticnetwork/avail-settlement/pkg/common"
	"github.com/maticnetwork/avail-settlement/pkg/staking"
	"github.com/maticnetwork/avail-settlement/pkg/test"
)

func getGenesisBasePath() string {
	path, _ := os.Getwd()
	return filepath.Join(path, "..", "..")
}

func NewTestAvail(t *testing.T, nodeType MechanismType) (*Avail, staking.ActiveParticipants) {
	chain, err := test.NewChain(getGenesisBasePath())
	if err != nil {
		t.Fatal(err)
	}

	executor, blockchain, txpool, err := test.NewBlockchainWithTxPool(chain, staking.NewVerifier(new(staking.DumbActiveParticipants), hclog.Default()))
	if err != nil {
		t.Fatal(err)
	}

	asq := staking.NewActiveParticipantsQuerier(blockchain, executor, hclog.Default())

	balance := big.NewInt(0).Mul(big.NewInt(1000), common.ETH)
	sequencerAddr, sequencerSignKey := test.NewAccount(t)
	test.DepositBalance(t, sequencerAddr, balance, blockchain, executor)

	// Set the verifier as we cannot pass it directly to the NewBlockchain
	verifier := staking.NewVerifier(asq, hclog.Default())
	blockchain.SetConsensus(verifier)

	sender := avail.NewBlackholeSender()
	stakingNode := staking.NewNode(blockchain, executor, sender, hclog.Default(), staking.NodeType(nodeType))

	return &Avail{
		logger:      hclog.Default(),
		notifyCh:    make(chan struct{}),
		closeCh:     make(chan struct{}),
		blockchain:  blockchain,
		executor:    executor,
		verifier:    verifier,
		txpool:      txpool,
		blockTime:   time.Duration(1) * time.Second,
		nodeType:    nodeType,
		signKey:     sequencerSignKey,
		minerAddr:   sequencerAddr,
		availSender: sender,
		stakingNode: stakingNode,
	}, asq
}
