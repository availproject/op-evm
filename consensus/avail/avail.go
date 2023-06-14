package avail

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/0xPolygon/polygon-edge/blockchain"
	"github.com/0xPolygon/polygon-edge/chain"
	"github.com/0xPolygon/polygon-edge/consensus"
	"github.com/0xPolygon/polygon-edge/crypto"
	"github.com/0xPolygon/polygon-edge/helper/progress"
	"github.com/0xPolygon/polygon-edge/network"
	"github.com/0xPolygon/polygon-edge/secrets"
	"github.com/0xPolygon/polygon-edge/state"
	"github.com/0xPolygon/polygon-edge/txpool"
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	avail_types "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/hashicorp/go-hclog"
	"github.com/maticnetwork/avail-settlement/consensus/avail/validator"
	"github.com/maticnetwork/avail-settlement/pkg/avail"
	common_defs "github.com/maticnetwork/avail-settlement/pkg/common"
	"github.com/maticnetwork/avail-settlement/pkg/faucet"
	"github.com/maticnetwork/avail-settlement/pkg/snapshot"
	"github.com/maticnetwork/avail-settlement/pkg/staking"
)

const (
	// 1 AVL == 10^18 Avail fractions.
	AVL = 1_000_000_000_000_000_000

	// DefaultBlockProductionIntervalS - In seconds, default block loop production attempt interval
	DefaultBlockProductionIntervalS = 1

	// For now hand coded address of the sequencer
	SequencerAddress = "0xF817d12e6933BbA48C14D4c992719B46aD9f5f61"

	// For now hand coded address of the watch tower
	WatchTowerAddress = "0xF817d12e6933BbA48C14D4c992719B46aD9f5f61"

	// StakingPollPeersIntervalMs interval to wait for when waiting for peer sto come up before staking
	StakingPollPeersIntervalMs = 200
)

// minBalance is the minimum number of tokens that miner address must have, in
// order to being able to run this node.
var minBalance = big.NewInt(0).Mul(big.NewInt(15), common_defs.ETH)

// Used to sync initial balance (if needed) only once to remove attempts to insert
// same tx multiple times.
var balanceOnce sync.Once

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
	Logger                hclog.Logger
	Network               *network.Server
	NodeType              string
	SecretsManager        secrets.SecretsManager
	Snapshotter           snapshot.Snapshotter
	TxPool                *txpool.TxPool
	AvailAppID            avail_types.UCompact
	NumBlockConfirmations uint64
}

// Dev consensus protocol seals any new transaction immediately
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
}

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

// Initialize initializes the consensus
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

// Start starts the consensus mechanism
// TODO: GRPC interface and listener, validator sequence and initialization as well P2P networking
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

func (d *Avail) startBootstrapSequencer() {
	activeParticipantsQuerier := staking.NewActiveParticipantsQuerier(d.blockchain, d.executor, d.logger)

	sequencerWorker, _ := NewSequencer(
		d.logger.Named(d.nodeType.LogString()), d.blockchain, d.executor, d.txpool,
		d.snapshotter, d.snapshotDistributor,
		d.availClient, d.availAccount, d.availAppID, d.signKey,
		d.minerAddr, d.nodeType, activeParticipantsQuerier, d.stakingNode, d.availSender, d.closeCh,
		d.blockTime, d.blockProductionIntervalSec, d.currentNodeSyncIndex,
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

func (d *Avail) startSequencer() {
	activeParticipantsQuerier := staking.NewActiveParticipantsQuerier(d.blockchain, d.executor, d.logger)

	sequencerWorker, _ := NewSequencer(
		d.logger.Named(d.nodeType.LogString()), d.blockchain, d.executor, d.txpool,
		d.snapshotter, d.snapshotDistributor,
		d.availClient, d.availAccount, d.availAppID, d.signKey,
		d.minerAddr, d.nodeType, activeParticipantsQuerier, d.stakingNode, d.availSender, d.closeCh,
		d.blockTime, d.blockProductionIntervalSec, d.currentNodeSyncIndex,
	)

	d.logger.Info("About to process node staking...", "node_type", d.nodeType)
	if err := d.ensureStaked(nil, activeParticipantsQuerier); err != nil {
		panic(err)
	}

	if err := sequencerWorker.Run(accounts.Account{Address: common.Address(d.minerAddr)}, &keystore.Key{PrivateKey: d.signKey}); err != nil {
		panic(err)
	}
}

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

// Node sync condition. The node's miner account must have at least
// `minBalance` tokens deposited and the syncer must have reached the Avail
// HEAD.
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

func (d *Avail) VerifyHeader(header *types.Header) error {
	return d.verifier.VerifyHeader(header)
}

func (d *Avail) ProcessHeaders(headers []*types.Header) error {
	return d.verifier.ProcessHeaders(headers)
}

func (d *Avail) GetBlockCreator(header *types.Header) (types.Address, error) {
	return d.verifier.GetBlockCreator(header)
}

// PreCommitState a hook to be called before finalizing state transition on inserting block
func (d *Avail) PreCommitState(header *types.Header, tx *state.Transition) error {
	return d.verifier.PreCommitState(header, tx)
}

func (d *Avail) GetSyncProgression() *progress.Progression {
	return nil // d.syncer.GetSyncProgression()
}

// GetBridgeProvider returns an instance of BridgeDataProvider
func (d *Avail) GetBridgeProvider() consensus.BridgeDataProvider {
	return nil
}

func (d *Avail) Prepare(header *types.Header) error {
	// TODO: Remove
	return nil
}

func (d *Avail) Seal(block *types.Block, ctx context.Context) (*types.Block, error) {
	// TODO: Remove
	return nil, nil
}

func (d *Avail) FilterExtra(extra []byte) ([]byte, error) {
	return extra, nil
}

func (d *Avail) Close() error {
	close(d.closeCh)
	return nil
}
