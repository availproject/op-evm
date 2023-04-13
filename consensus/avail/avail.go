package avail

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"time"

	"github.com/0xPolygon/polygon-edge/blockchain"
	"github.com/0xPolygon/polygon-edge/consensus"
	"github.com/0xPolygon/polygon-edge/crypto"
	"github.com/0xPolygon/polygon-edge/helper/progress"
	"github.com/0xPolygon/polygon-edge/network"
	"github.com/0xPolygon/polygon-edge/secrets"
	"github.com/0xPolygon/polygon-edge/state"
	"github.com/0xPolygon/polygon-edge/syncer"
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

type Config struct {
	AccountFilePath string
	AvailAccount    signature.KeyringPair
	AvailClient     avail.Client
	AvailSender     avail.Sender
	Blockchain      *blockchain.Blockchain
	BlockTime       uint64
	Bootnode        bool
	Context         context.Context
	Config          *consensus.Config
	Executor        *state.Executor
	Logger          hclog.Logger
	Network         *network.Server
	NodeType        string
	SecretsManager  secrets.SecretsManager
	TxPool          *txpool.TxPool
}

// Dev consensus protocol seals any new transaction immediately
type Avail struct {
	logger     hclog.Logger
	mechanisms []MechanismType
	nodeType   MechanismType

	syncer syncer.Syncer // Reference to the sync protocol

	notifyCh chan struct{}
	closeCh  chan struct{}

	availAppID avail_types.UCompact
	signKey    *ecdsa.PrivateKey
	minerAddr  types.Address

	interval uint64
	txpool   *txpool.TxPool

	blockchain *blockchain.Blockchain
	executor   *state.Executor
	verifier   blockchain.Verifier

	network        *network.Server // Reference to the networking layer
	secretsManager secrets.SecretsManager
	blockTime      time.Duration // Minimum block generation time in seconds

	availAccount signature.KeyringPair
	availClient  avail.Client
	availSender  avail.Sender
	stakingNode  staking.Node

	blockProductionIntervalSec uint64
	validator                  validator.Validator
}

func New(config Config) (consensus.Consensus, error) {
	logger := config.Logger.Named("avail")

	bs, err := config.SecretsManager.GetSecret(secrets.ValidatorKey)
	if err != nil {
		panic("can't find validator key! - " + err.Error())
	}

	validatorKey, err := crypto.BytesToECDSAPrivateKey(bs)
	if err != nil {
		panic("validator key decoding failed: " + err.Error())
	}

	validatorAddr := crypto.PubKeyToAddress(&validatorKey.PublicKey)

	asq := staking.NewActiveParticipantsQuerier(config.Blockchain, config.Executor, logger)

	d := &Avail{
		logger:                     logger,
		notifyCh:                   make(chan struct{}),
		closeCh:                    make(chan struct{}),
		blockchain:                 config.Blockchain,
		executor:                   config.Executor,
		verifier:                   staking.NewVerifier(asq, logger.Named("verifier")),
		txpool:                     config.TxPool,
		secretsManager:             config.SecretsManager,
		network:                    config.Network,
		blockTime:                  time.Duration(config.BlockTime) * time.Second,
		nodeType:                   MechanismType(config.NodeType),
		signKey:                    validatorKey,
		minerAddr:                  validatorAddr,
		validator:                  validator.New(config.Blockchain, config.Executor, validatorAddr, logger),
		blockProductionIntervalSec: DefaultBlockProductionIntervalS,

		availAccount: config.AvailAccount,
		availClient:  config.AvailClient,
		availSender:  config.AvailSender,
	}

	if config.Network != nil {
		d.syncer = syncer.NewSyncer(
			config.Logger,
			config.Network,
			config.Blockchain,
			time.Duration(config.BlockTime)*3*time.Second,
		)
	}

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
	return nil
}

// Start starts the consensus mechanism
// TODO: GRPC interface and listener, validator sequence and initialization as well P2P networking
func (d *Avail) Start() error {
	var (
		activeParticipantsQuerier = staking.NewActiveParticipantsQuerier(d.blockchain, d.executor, d.logger)
		account                   = accounts.Account{Address: common.Address(d.minerAddr)}
		key                       = &keystore.Key{PrivateKey: d.signKey}
	)

	// Enable P2P gossiping.
	d.txpool.SetSealing(true)

	// Start P2P syncing.
	go d.startSyncing()

	d.logger.Info("About to process node staking...", "node_type", d.nodeType)
	if err := d.ensureStaked(nil, activeParticipantsQuerier); err != nil {
		return err
	}

	switch d.nodeType {
	case Sequencer, BootstrapSequencer:
		sequencerWorker, _ := NewSequencer(
			d.logger.Named(d.nodeType.LogString()), d.blockchain, d.executor, d.txpool,
			d.availClient, d.availAccount, d.availAppID, d.signKey,
			d.minerAddr, d.nodeType, activeParticipantsQuerier, d.stakingNode, d.availSender, d.closeCh,
			d.blockTime, d.blockProductionIntervalSec,
		)
		go func() {
			if err := sequencerWorker.Run(accounts.Account{Address: common.Address(d.minerAddr)}, &keystore.Key{PrivateKey: d.signKey}); err != nil {
				panic(err)
			}
		}()
	case WatchTower:
		go d.runWatchTower(activeParticipantsQuerier, account, key)
	default:
		return fmt.Errorf("invalid node type: %q", d.nodeType)
	}

	return nil
}

func (d *Avail) startSyncing() {
	if d.syncer == nil {
		d.logger.Warn("syncer not configured")
		return
	}

	// Start the syncer
	err := d.syncer.Start()
	if err != nil {
		panic(fmt.Sprintf("starting blockchain sync failed: %s", err))
	}

	syncFunc := func(blk *types.FullBlock) bool {
		d.txpool.ResetWithHeaders(blk.Block.Header)
		return false
	}

	err = d.syncer.Sync(syncFunc)
	if err != nil {
		panic(fmt.Sprintf("syncing blockchain failed: %s", err))
	}
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

func (d *Avail) Close() error {
	close(d.closeCh)

	if d.syncer != nil {
		d.syncer.Close()
	}

	return nil
}
