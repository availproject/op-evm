// Package avail provides the implementation of the Avail consensus protocol for the blockchain network.
// Avail is a modular and flexible optimistic consensus mechanism designed to ensure secure and efficient transaction processing and block creation.
// It offers different node types, including Sequencer and WatchTower, each with specific roles and responsibilities in the consensus process.
//
// The package includes functionalities for initializing and starting the Avail consensus mechanism, handling node synchronization,
// verifying block headers, processing headers, retrieving block creators, and managing account balances.
// It also provides hooks for pre-commit state operations and sealing blocks.
//
// With Avail, nodes can participate in the consensus protocol, validate transactions, and contribute to the creation of new blocks.
// The consensus mechanism employs active participant querying and employs staking to ensure the integrity and security of the network.
package avail

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/0xPolygon/polygon-edge/chain"
	"github.com/0xPolygon/polygon-edge/consensus"
	"github.com/0xPolygon/polygon-edge/crypto"
	"github.com/0xPolygon/polygon-edge/helper/progress"
	"github.com/0xPolygon/polygon-edge/network"
	"github.com/0xPolygon/polygon-edge/secrets"
	"github.com/0xPolygon/polygon-edge/state"
	"github.com/0xPolygon/polygon-edge/txpool"
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/availproject/op-evm/consensus/avail/validator"
	"github.com/availproject/op-evm/pkg/avail"
	"github.com/availproject/op-evm/pkg/blockchain"
	common_defs "github.com/availproject/op-evm/pkg/common"
	"github.com/availproject/op-evm/pkg/faucet"
	"github.com/availproject/op-evm/pkg/snapshot"
	"github.com/availproject/op-evm/pkg/staking"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	avail_types "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/hashicorp/go-hclog"
)

// Constants and Variables
const (
	// AVL represents 1 Avail token which is equivalent to 10^18 fractions of an Avail token.
	AVL = 1_000_000_000_000_000_000

	// DefaultBlockProductionIntervalS represents the default interval in seconds for attempting block production.
	DefaultBlockProductionIntervalS = 1

	// StakingPollPeersIntervalMs is the interval in milliseconds to wait for when waiting for peers to come up before staking.
	StakingPollPeersIntervalMs = 200
)

// minBalance is the minimum number of tokens that miner address must have, in
// order to being able to run this node.
var minBalance = big.NewInt(0).Mul(big.NewInt(15), common_defs.ETH)

// Used to sync initial balance (if needed) only once to remove attempts to insert
// same tx multiple times.
var balanceOnce sync.Once

// Config is a structure that holds various configuration options required by the Avail consensus protocol.
type Config struct {
	AccountFilePath       string
	AvailAccount          signature.KeyringPair
	AvailClient           avail.Client
	AvailSender           avail.Sender
	Blockchain            *blockchain.Blockchain
	BlockTime             uint64
	Bootnode              bool
	Chain                 *chain.Chain
	Context               context.Context
	Config                *consensus.Config
	Executor              *state.Executor
	FraudListenerAddr     string
	Logger                hclog.Logger
	Network               *network.Server
	NodeType              string
	SecretsManager        secrets.SecretsManager
	Snapshotter           snapshot.Snapshotter
	TxPool                *txpool.TxPool
	AvailAppID            avail_types.UCompact
	NumBlockConfirmations uint64
}

// Avail represents the consensus protocol for the Avail network.
// It implements the Consensus interface and contains various configurations and mechanisms for consensus.
type Avail struct {
	logger     hclog.Logger
	mechanisms []MechanismType
	nodeType   MechanismType

	notifyCh chan struct{}
	closeCh  chan struct{}

	availAppID avail_types.UCompact
	signKey    *ecdsa.PrivateKey
	minerAddr  types.Address

	interval uint64
	txpool   *txpool.TxPool

	chain               *chain.Chain
	blockchain          *blockchain.Blockchain
	executor            *state.Executor
	snapshotter         snapshot.Snapshotter
	snapshotDistributor snapshot.Distributor
	verifier            blockchain.Verifier

	network        *network.Server // Reference to the networking layer
	secretsManager secrets.SecretsManager
	blockTime      time.Duration // Minimum block generation time in seconds

	availAccount signature.KeyringPair
	availClient  avail.Client
	availSender  avail.Sender
	stakingNode  staking.Node

	blockProductionIntervalSec uint64
	validator                  validator.Validator
	currentNodeSyncIndex       uint64
	fraudListenerAddr          string
}

// New creates and initializes a new instance of the Avail consensus protocol with the provided configuration.
// It also sets up necessary dependencies including the staking node, private signing key, miner address, snapshot distributor etc. It validates the configuration and returns the Avail consensus protocol instance.
// The function can panic if it fails to find or decode the signing key. Returns error if the configuration is invalid or it fails to setup any of the dependencies.
func New(config Config) (consensus.Consensus, error) {
	logger := config.Logger.Named("avail")

	bs, err := config.SecretsManager.GetSecret(secrets.ValidatorKey)
	if err != nil {
		panic("can't find sign key! - " + err.Error())
	}

	signKey, err := crypto.BytesToECDSAPrivateKey(bs)
	if err != nil {
		panic("sign key decoding failed: " + err.Error())
	}

	minerAddr := crypto.PubKeyToAddress(&signKey.PublicKey)

	asq := staking.NewActiveParticipantsQuerier(config.Blockchain, config.Executor, logger)

	d := &Avail{
		logger:                     logger,
		notifyCh:                   make(chan struct{}),
		chain:                      config.Chain,
		closeCh:                    make(chan struct{}),
		blockchain:                 config.Blockchain,
		executor:                   config.Executor,
		snapshotter:                config.Snapshotter,
		verifier:                   staking.NewVerifier(asq, logger.Named("verifier")),
		txpool:                     config.TxPool,
		secretsManager:             config.SecretsManager,
		network:                    config.Network,
		blockTime:                  time.Duration(config.BlockTime) * time.Second,
		nodeType:                   MechanismType(config.NodeType),
		signKey:                    signKey,
		minerAddr:                  minerAddr,
		validator:                  validator.New(config.Blockchain, minerAddr, logger),
		blockProductionIntervalSec: DefaultBlockProductionIntervalS,
		availAccount:               config.AvailAccount,
		availClient:                config.AvailClient,
		availSender:                config.AvailSender,
		availAppID:                 config.AvailAppID,
		fraudListenerAddr:          config.FraudListenerAddr,
	}

	if config.Network != nil {
		d.snapshotDistributor, err = snapshot.NewDistributor(d.logger, d.network)
		if err != nil {
			return nil, err
		}
	} /* TODO: Implement /dev/null snapshot distributor for no-network situations.
		else {
	} */

	if d.mechanisms, err = ParseMechanismConfigTypes(config.Config.Config["mechanisms"]); err != nil {
		return nil, fmt.Errorf("invalid avail mechanism type/s provided")
	}

	if d.nodeType == BootstrapSequencer && !config.Bootnode {
		return nil, fmt.Errorf("invalid avail node type provided: cannot specify bootstrap-sequencer type without -bootnode flag")
	}

	if d.nodeType == Sequencer && config.Bootnode {
		d.nodeType = BootstrapSequencer
	}

	rawInterval, ok := config.Config.Config["interval"]
	if ok {
		interval, ok := rawInterval.(uint64)
		if !ok {
			return nil, fmt.Errorf("interval expected int")
		}

		d.interval = interval
	}

	blockProductionIntervalSecRaw, ok := config.Config.Config["blockProductionIntervalSec"]
	if ok {
		blockProductionIntervalSec, ok := blockProductionIntervalSecRaw.(uint64)
		if !ok {
			return nil, fmt.Errorf("blockProductionIntervalSec expected int")
		}

		d.blockProductionIntervalSec = blockProductionIntervalSec
	}

	d.stakingNode = staking.NewNode(d.blockchain, d.executor, d.availSender, d.logger, staking.NodeType(d.nodeType))

	return d, nil
}

// Initialize verifies the initial balance of the miner's account.
// If the account does not exist or does not have a balance yet (returns a 'state not found' error), it returns nil.
// If the account's balance is less than the minimum required balance, the function attempts to find the account in the faucet.
// If the account is not found in the faucet or any other error occurs, an error is returned.
func (d *Avail) Initialize() error {
	balance, err := d.GetAccountBalance(d.minerAddr)
	if err != nil && strings.HasPrefix(err.Error(), "state not found") {
		// On accounts that don't have balance / don't exist
		// -> a `state not found at hash ...` error is returned.
		return nil
	} else if err != nil {
		return err
	}

	if balance.Cmp(minBalance) < 0 {
		_, err := faucet.FindAccount(d.chain)
		if err == faucet.ErrAccountNotFound {
			return fmt.Errorf("not enough balance on account - cannot continue")
		}
		if err != nil {
			return err
		}
	}

	return nil
}

// Start initiates the consensus mechanism.
// It enables P2P gossiping and verifies if the node type is not a BootstrapSequencer.
// For node types other than BootstrapSequencer, it ensures that at least one bootnode is available before syncing.
// If there are no nodes to push transactions towards, this function waits for 2 seconds before attempting to sync again.
// After a successful sync, the function checks the node type and starts the respective process.
// If the node type is invalid, an error is returned.
// Note: A panic occurs if the node fails to sync.
func (d *Avail) Start() error {
	// Enable P2P gossiping.
	d.txpool.SetSealing(true)

	if d.nodeType != BootstrapSequencer {
		// When node starts, txpool is started but because peer count is not yet updated and
		// there is no nodes to push transactions towards, we should first wait for at least
		// 1 bootnode to be available prior we continue syncing.
		// Syncing will at the last step attempt to top up the faucet balance if needed which may fail and
		// usually fails due to txpool not having any peers to send tx towards.
		// This results in local tx being applied and node goes into corrupted mode.
		// Following for functionality is here as well to ensure we do not unecessary sleeps
		// in server.go when txpool and network is starting.
		for d.network == nil || d.network.GetBootnodeConnCount() < 1 {
			time.Sleep(2 * time.Second)
		}

		// Sync the node from Avail.
		var err error
		d.currentNodeSyncIndex, err = d.syncNodeUntil(d.syncConditionFn)
		if err != nil {
			panic(fmt.Sprintf("failure to sync node: %s", err))
		}
	}

	switch d.nodeType {
	case BootstrapSequencer:
		go d.startBootstrapSequencer()

	case Sequencer:
		go d.startSequencer()

	case WatchTower:
		go d.startWatchTower()

	default:
		return fmt.Errorf("invalid node type: %q", d.nodeType)
	}

	return nil
}

// startBootstrapSequencer starts the process for a BootstrapSequencer node type.
// It initializes a new Sequencer, syncs the node, and ensures the node is staked.
// If the node successfully syncs and stakes, it starts running the Sequencer worker.
// Note: The function panics if it fails to sync the node, ensure the node is staked, or run the Sequencer worker.
func (d *Avail) startBootstrapSequencer() {
	activeParticipantsQuerier := staking.NewActiveParticipantsQuerier(d.blockchain, d.executor, d.logger)

	sequencerWorker, _ := NewSequencer(
		d.logger.Named(d.nodeType.LogString()), d.blockchain, d.executor, d.txpool,
		d.snapshotter, d.snapshotDistributor,
		d.availClient, d.availAccount, d.availAppID, d.signKey,
		d.minerAddr, d.nodeType, activeParticipantsQuerier, d.stakingNode, d.availSender, d.closeCh,
		d.blockTime, d.blockProductionIntervalSec, d.currentNodeSyncIndex,
		d.fraudListenerAddr,
	)

	// Sync the node from Avail.
	var err error
	d.currentNodeSyncIndex, err = d.syncNode()
	if err != nil {
		panic(err)
	}

	d.logger.Info("About to process node staking...", "node_type", d.nodeType)
	if err := d.ensureStaked(nil, activeParticipantsQuerier); err != nil {
		panic(err)
	}

	if err := sequencerWorker.Run(accounts.Account{Address: common.Address(d.minerAddr)}, &keystore.Key{PrivateKey: d.signKey}); err != nil {
		panic(err)
	}
}

// startSequencer starts the process for a Sequencer node type.
// It initializes a new Sequencer, ensures the node is staked, and runs the Sequencer worker.
// Note: The function panics if it fails to ensure the node is staked or run the Sequencer worker.
func (d *Avail) startSequencer() {
	activeParticipantsQuerier := staking.NewActiveParticipantsQuerier(d.blockchain, d.executor, d.logger)

	sequencerWorker, _ := NewSequencer(
		d.logger.Named(d.nodeType.LogString()), d.blockchain, d.executor, d.txpool,
		d.snapshotter, d.snapshotDistributor,
		d.availClient, d.availAccount, d.availAppID, d.signKey,
		d.minerAddr, d.nodeType, activeParticipantsQuerier, d.stakingNode, d.availSender, d.closeCh,
		d.blockTime, d.blockProductionIntervalSec, d.currentNodeSyncIndex,
		d.fraudListenerAddr,
	)

	d.logger.Info("About to process node staking...", "node_type", d.nodeType)
	if err := d.ensureStaked(nil, activeParticipantsQuerier); err != nil {
		panic(err)
	}

	if err := sequencerWorker.Run(accounts.Account{Address: common.Address(d.minerAddr)}, &keystore.Key{PrivateKey: d.signKey}); err != nil {
		panic(err)
	}
}

// startWatchTower starts the process for a WatchTower node type.
// It ensures the node is staked and runs the WatchTower process.
// Note: The function panics if it fails to ensure the node is staked or run the WatchTower process.
func (d *Avail) startWatchTower() {
	activeParticipantsQuerier := staking.NewActiveParticipantsQuerier(d.blockchain, d.executor, d.logger)
	key := &keystore.Key{PrivateKey: d.signKey}

	d.logger.Info("About to process node staking...", "node_type", d.nodeType)
	if err := d.ensureStaked(nil, activeParticipantsQuerier); err != nil {
		panic(err)
	}

	acc := accounts.Account{Address: common.Address(d.minerAddr)}
	d.runWatchTower(activeParticipantsQuerier, d.currentNodeSyncIndex, acc, key)
}

// ensureAccountBalance verifies the account balance of the miner.
// If the current balance is less than the minimum required balance,
// the function tops up the account balance by depositing additional tokens from the faucet account.
// Note: The function returns an error if any operation fails.
func (d *Avail) ensureAccountBalance() error {
	faucetSignKey, err := faucet.FindAccount(d.chain)
	if err != nil {
		return err
	}

	// Query the current balance of the miner account.
	currentBalance, err := d.GetAccountBalance(d.minerAddr)

	// 'state not found' means that the account doesn't exist yet.
	if err != nil && strings.HasPrefix(err.Error(), "state not found") {
		currentBalance = big.NewInt(0)
	} else if err != nil {
		return err
	}

	if currentBalance.Cmp(minBalance) >= 0 {
		// No need to top up the account balance.
		return nil
	}

	// Necessary amount of tokens to be deposited to miner account.
	amount := big.NewInt(0).Sub(minBalance, currentBalance)

	var txn *state.Transition
	{
		hdr := d.blockchain.Header()
		if hdr == nil {
			return fmt.Errorf("blockchain returned nil header")
		}

		txn, err = d.executor.BeginTxn(hdr.StateRoot, hdr, d.minerAddr)
		if err != nil {
			return err
		}
	}

	faucetAddr := crypto.PubKeyToAddress(&faucetSignKey.PublicKey)

	tx := &types.Transaction{
		From:     faucetAddr,
		To:       &d.minerAddr,
		Value:    amount,
		GasPrice: big.NewInt(5000),
		Gas:      1_000_000,
		Nonce:    txn.GetNonce(faucetAddr),
	}

	txSigner := &crypto.FrontierSigner{}
	tx, err = txSigner.SignTx(tx, faucetSignKey)
	if err != nil {
		return err
	}

	err = d.txpool.AddTx(tx)
	if err != nil {
		return err
	}

	return nil
}

// GetAccountBalance retrieves the balance of an account.
// It fetches the latest header from Avail and returns the balance associated with the specified address.
// If the balance is not found or any error occurs, an error is returned.
func (d *Avail) GetAccountBalance(addr types.Address) (*big.Int, error) {
	hdr := d.blockchain.Header()
	if hdr == nil {
		return nil, fmt.Errorf("blockchain returned nil header")
	}

	txn, err := d.executor.BeginTxn(hdr.StateRoot, hdr, addr)
	if err != nil {
		return nil, err
	}

	return txn.GetBalance(addr), nil
}

// syncConditionFn defines the condition for node synchronization.
// It checks if the miner's account balance is equal to or greater than the minimum required balance
// and if the syncer has reached the Avail HEAD.
// The function returns true if the conditions are met; otherwise, it returns false.
func (d *Avail) syncConditionFn(blk *avail_types.SignedBlock) bool {
	hdr, err := d.availClient.GetLatestHeader()
	if err != nil {
		d.logger.Error("couldn't fetch latest block hash from Avail", "error", err)
		return false
	}

	if hdr.Number == blk.Block.Header.Number {
		accountBalance, err := d.GetAccountBalance(d.minerAddr)
		if err != nil && strings.HasPrefix(err.Error(), "state not found") {
			// No need to log this.
			return false
		} else if err != nil {
			d.logger.Error("failed to query miner account balance", "error", err)
			return false
		}

		// Sync until our deposit tx is through.
		if accountBalance.Cmp(minBalance) < 0 {
			balanceOnce.Do(func() {
				if err := d.ensureAccountBalance(); err != nil {
					d.logger.Error("failed to ensure account balance", "error", err)
					panic(fmt.Sprintf("failure to apply faucet balance to the txpool: %s", err))
				}
			})

			return false
		}

		// Our miner account has enough funds to operate and we have reached Avail
		// HEAD. Sync complete.
		return true
	}

	return false
}

// REQUIRED BASE INTERFACE METHODS //

// VerifyHeader verifies the validity of a block header.
// It delegates the verification process to the underlying verifier.
// The function returns an error if the header is invalid.
func (d *Avail) VerifyHeader(header *types.Header) error {
	return d.verifier.VerifyHeader(header)
}

// ProcessHeaders processes a batch of block headers.
// It delegates the processing to the underlying verifier.
// The function returns an error if any header fails to process.
func (d *Avail) ProcessHeaders(headers []*types.Header) error {
	return d.verifier.ProcessHeaders(headers)
}

// GetBlockCreator returns the address of the block creator for a given header.
// It delegates the retrieval process to the underlying verifier.
// The function returns the block creator address or an error if the retrieval fails.
func (d *Avail) GetBlockCreator(header *types.Header) (types.Address, error) {
	return d.verifier.GetBlockCreator(header)
}

// PreCommitState is a hook called before finalizing the state transition on inserting a block.
// It delegates the pre-commit state process to the underlying verifier.
// The function returns an error if the pre-commit state operation fails.
func (d *Avail) PreCommitState(header *types.Header, tx *state.Transition) error {
	return d.verifier.PreCommitState(header, tx)
}

// GetSyncProgression returns the progression of the node's sync process.
// As of now, the function returns nil as the sync progression is not implemented.
func (d *Avail) GetSyncProgression() *progress.Progression {
	return nil
}

// GetBridgeProvider returns an instance of BridgeDataProvider.
// As of now, the function returns nil as the BridgeDataProvider is not implemented.
func (d *Avail) GetBridgeProvider() consensus.BridgeDataProvider {
	return nil
}

// Prepare is a method that is part of the consensus.BaseConsensus interface.
// As of now, the function is empty and does not perform any specific operations.
func (d *Avail) Prepare(header *types.Header) error {
	return nil
}

// Seal is a method that is part of the consensus.BaseConsensus interface.
// As of now, the function is empty and does not perform any specific operations.
func (d *Avail) Seal(block *types.Block, ctx context.Context) (*types.Block, error) {
	return nil, nil
}

// FilterExtra is a method that is part of the consensus.BaseConsensus interface.
// As of now, the function returns the input 'extra' without any modifications.
// It does not perform any filtering operations.
func (d *Avail) FilterExtra(extra []byte) ([]byte, error) {
	return extra, nil
}

// Close closes the Avail consensus.
// It closes the internal close channel and returns nil.
func (d *Avail) Close() error {
	close(d.closeCh)
	return nil
}
