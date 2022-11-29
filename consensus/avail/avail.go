package avail

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"time"

	"github.com/0xPolygon/polygon-edge/blockchain"
	"github.com/0xPolygon/polygon-edge/consensus"
	"github.com/0xPolygon/polygon-edge/crypto"
	"github.com/0xPolygon/polygon-edge/helper/progress"
	"github.com/0xPolygon/polygon-edge/network"
	"github.com/0xPolygon/polygon-edge/syncer"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"

	"github.com/0xPolygon/polygon-edge/secrets"
	"github.com/0xPolygon/polygon-edge/state"
	"github.com/0xPolygon/polygon-edge/txpool"
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/hashicorp/go-hclog"
	"github.com/maticnetwork/avail-settlement/pkg/avail"
	"github.com/maticnetwork/avail-settlement/pkg/staking"
)

const (
	// For now hand coded address of the sequencer
	SequencerAddress = "0xF817d12e6933BbA48C14D4c992719B46aD9f5f61"

	// For now hand coded address of the watch tower
	WatchTowerAddress = "0xF817d12e6933BbA48C14D4c992719B46aD9f5f61"
)

var (
	ETH = big.NewInt(1000000000000000000)
)

type Config struct {
	AvailAddr string
	Bootnode  bool
}

// Dev consensus protocol seals any new transaction immediately
type Avail struct {
	logger      hclog.Logger
	availClient avail.Client
	mechanisms  []MechanismType
	nodeType    MechanismType

	syncer syncer.Syncer // Reference to the sync protocol

	notifyCh chan struct{}
	closeCh  chan struct{}

	signKey   *ecdsa.PrivateKey
	minerAddr types.Address

	interval uint64
	txpool   *txpool.TxPool

	blockchain *blockchain.Blockchain
	executor   *state.Executor
	verifier   blockchain.Verifier

	updateCh chan struct{} // nolint:unused // Update channel

	network        *network.Server // Reference to the networking layer
	secretsManager secrets.SecretsManager
	blockTime      time.Duration // Minimum block generation time in seconds
}

// Factory returns the consensus factory method
func Factory(config Config) func(params *consensus.Params) (consensus.Consensus, error) {
	return func(params *consensus.Params) (consensus.Consensus, error) {
		logger := params.Logger.Named("avail")

		bs, err := params.SecretsManager.GetSecret(secrets.ValidatorKey)
		if err != nil {
			panic("can't find validator key! - " + err.Error())
		}

		validatorKey, err := crypto.BytesToECDSAPrivateKey(bs)
		if err != nil {
			panic("validator key decoding failed: " + err.Error())
		}

		validatorAddr := crypto.PubKeyToAddress(&validatorKey.PublicKey)

		asq := staking.NewActiveParticipantsQuerier(params.Blockchain, params.Executor, logger)

		d := &Avail{
			logger:         logger,
			notifyCh:       make(chan struct{}),
			closeCh:        make(chan struct{}),
			blockchain:     params.Blockchain,
			executor:       params.Executor,
			verifier:       staking.NewVerifier(asq, logger.Named("verifier")),
			txpool:         params.TxPool,
			secretsManager: params.SecretsManager,
			network:        params.Network,
			blockTime:      time.Duration(params.BlockTime) * time.Second,
			nodeType:       MechanismType(params.NodeType),
			syncer: syncer.NewSyncer(
				params.Logger,
				params.Network,
				params.Blockchain,
				time.Duration(params.BlockTime)*3*time.Second,
			),
			signKey:   validatorKey,
			minerAddr: validatorAddr,
		}

		if d.mechanisms, err = ParseMechanismConfigTypes(params.Config.Config["mechanisms"]); err != nil {
			return nil, fmt.Errorf("invalid avail mechanism type/s provided")
		}

		if d.nodeType == BootstrapSequencer && !config.Bootnode {
			return nil, fmt.Errorf("invalid avail node type provided: cannot specify bootstrap-sequencer type without -bootnode flag")
		}

		if d.nodeType == Sequencer && config.Bootnode {
			d.nodeType = BootstrapSequencer
		}

		d.availClient, err = avail.NewClient(config.AvailAddr)
		if err != nil {
			return nil, err
		}

		rawInterval, ok := params.Config.Config["interval"]
		if ok {
			interval, ok := rawInterval.(uint64)
			if !ok {
				return nil, fmt.Errorf("interval expected int")
			}

			d.interval = interval
		}

		return d, nil
	}
}

// Initialize initializes the consensus
func (d *Avail) Initialize() error {
	return nil
}

// Start starts the consensus mechanism
// TODO: GRPC interface and listener, validator sequence and initialization as well P2P networking
func (d *Avail) Start() error {
	stakingSender := staking.NewAvailSender(avail.NewSender(d.availClient, signature.TestKeyringPairAlice))
	stakingNode := staking.NewNode(d.blockchain, d.executor, stakingSender, d.logger, staking.NodeType(d.nodeType))

	if d.nodeType == Sequencer || d.nodeType == BootstrapSequencer {
		go d.runSequencer(stakingNode, accounts.Account{Address: common.Address(d.minerAddr)}, &keystore.Key{PrivateKey: d.signKey})
	}

	if d.nodeType == Validator {
		go d.runValidator()
	}

	if d.nodeType == WatchTower {
		wtAccount := accounts.Account{Address: common.Address(d.minerAddr)}
		wtPK := &keystore.Key{PrivateKey: d.signKey}

		activeParticipantsQuerier := staking.NewActiveParticipantsQuerier(d.blockchain, d.executor, d.logger)
		staked, err := activeParticipantsQuerier.Contains(d.minerAddr, staking.NodeType(d.nodeType))
		if err != nil {
			return err
		}

		if !staked {
			if err := d.stakeParticipant(activeParticipantsQuerier); err != nil {
				d.logger.Error("failure to build staking block", "error", err)
				return err
			}
		}

		go d.runWatchTower(stakingNode, wtAccount, wtPK)
	}

	return nil
}

// REQUIRED BASE INTERFACE METHODS //
// BeginDisputeResolution -

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
	return nil //d.syncer.GetSyncProgression()
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

	return nil
}
