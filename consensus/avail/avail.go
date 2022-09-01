package avail

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"time"

	"github.com/0xPolygon/polygon-edge/blockchain"
	"github.com/0xPolygon/polygon-edge/consensus"
	"github.com/0xPolygon/polygon-edge/helper/progress"
	"github.com/0xPolygon/polygon-edge/network"
	"github.com/0xPolygon/polygon-edge/protocol"
	"github.com/0xPolygon/polygon-edge/secrets"
	"github.com/0xPolygon/polygon-edge/state"
	"github.com/0xPolygon/polygon-edge/txpool"
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/hashicorp/go-hclog"
	"github.com/maticnetwork/avail-settlement/pkg/avail"
)

const (
	// For now hand coded address of the sequencer
	SequencerAddress = "0xF817d12e6933BbA48C14D4c992719B46aD9f5f61"

	// For now hand coded address of the watch tower
	WatchTowerAddress = "0xF817d12e6933BbA48C14D4c992719B46aD9f5f61"
)

type syncerInterface interface {
	Start()
	BestPeer() *protocol.SyncPeer
	BulkSyncWithPeer(p *protocol.SyncPeer, newBlockHandler func(block *types.Block)) error
	WatchSyncWithPeer(p *protocol.SyncPeer, newBlockHandler func(b *types.Block) bool, blockTimeout time.Duration)
	GetSyncProgression() *progress.Progression
	Broadcast(b *types.Block)
}

// Dev consensus protocol seals any new transaction immediately
type Avail struct {
	logger      hclog.Logger
	availClient avail.Client
	mechanisms  []MechanismType
	nodeType    MechanismType

	state *currentState // Reference to the current state

	notifyCh chan struct{}
	closeCh  chan struct{}

	validatorKey     *ecdsa.PrivateKey // Private key for the validator
	validatorKeyAddr types.Address

	syncer syncerInterface // Reference to the sync protocol

	interval uint64
	txpool   *txpool.TxPool

	blockchain *blockchain.Blockchain
	executor   *state.Executor

	updateCh chan struct{} // Update channel

	network        *network.Server // Reference to the networking layer
	secretsManager secrets.SecretsManager
	blockTime      time.Duration // Minimum block generation time in seconds
}

// Factory implements the base factory method
func Factory(
	params *consensus.ConsensusParams,
) (consensus.Consensus, error) {
	logger := params.Logger.Named("avail")

	d := &Avail{
		logger:         logger,
		notifyCh:       make(chan struct{}),
		closeCh:        make(chan struct{}),
		blockchain:     params.Blockchain,
		executor:       params.Executor,
		txpool:         params.Txpool,
		secretsManager: params.SecretsManager,
		network:        params.Network,
		blockTime:      time.Duration(params.BlockTime) * time.Second,
		state:          newState(),
		nodeType:       MechanismType(params.NodeType),
	}

	var err error
	if d.mechanisms, err = ParseMechanismConfigTypes(params.Config.Config["mechanisms"]); err != nil {
		return nil, fmt.Errorf("invalid avail mechanism type/s provided")
	}

	d.availClient, err = avail.NewClient("ws://127.0.0.1:9944/v1/json-rpc")
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

	d.syncer = protocol.NewSyncer(params.Logger, params.Network, params.Blockchain)

	return d, nil
}

// Initialize initializes the consensus
func (d *Avail) Initialize() error {
	return nil
}

// Start starts the consensus mechanism
// TODO: GRPC interface and listener, validator sequence and initialization as well P2P networking
func (d *Avail) Start() error {

	// Start the syncer
	d.syncer.Start()

	if d.nodeType == Sequencer {
		minerKeystore, minerAccount, minerPk, err := getAccountData(SequencerAddress)
		if err != nil {
			return err
		}

		go d.runSequencer(minerKeystore, minerAccount, minerPk)
	}

	if d.nodeType == Validator {
		go d.runValidator()
	}

	if d.nodeType == WatchTower {
		_, wtAccount, wtPK, err := getAccountData(WatchTowerAddress)
		if err != nil {
			return err
		}

		go d.runWatchTower(wtAccount, wtPK)
	}

	return nil
}

/* func (d *Avail) sendBlockToAvail(block *types.Block) error {
	sender := avail.NewSender(d.availClient, signature.TestKeyringPairAlice)
	d.logger.Info("Submitting block to avail...")
	hash, err := sender.SubmitDataWithoutWatch(block.MarshalRLP())
	if err != nil {
		d.logger.Error("Error while submitting data to avail", err)
		return err
	}
	d.logger.Info("Submitted block to avail", "block", block.Header.Hash, "avail_block", hash.Hex())
	return nil
} */

// STATE MACHINE METHODS //

// getState returns the current IBFT state
func (d *Avail) getState() AvailState {
	return d.state.getState()
}

// isState checks if the node is in the passed in state
func (d *Avail) isState(s AvailState) bool {
	return d.state.getState() == s
}

// setState sets the IBFT state
func (d *Avail) setState(s AvailState) {
	d.logger.Info("state change", "new", s)
	d.state.setState(s)
}

// REQUIRED BASE INTERFACE METHODS //

func (d *Avail) VerifyHeader(header *types.Header) error {

	signer, err := addressRecoverFromHeader(header)
	if err != nil {
		return err
	}

	d.logger.Info("Verify header", "signer", signer.String())

	if signer != types.StringToAddress(SequencerAddress) {
		d.logger.Info("Passing, how is it possible? 222")
		return fmt.Errorf("signer address '%s' does not match sequencer address '%s'", signer, SequencerAddress)
	}

	d.logger.Info("Seal signer address successfully verified!", "signer", signer, "sequencer", SequencerAddress)

	/*
		parent, ok := i.blockchain.GetHeaderByNumber(header.Number - 1)
		if !ok {
			return fmt.Errorf(
				"unable to get parent header for block number %d",
				header.Number,
			)
		}

		snap, err := i.getSnapshot(parent.Number)
		if err != nil {
			return err
		}

		// verify all the header fields + seal
		if err := i.verifyHeaderImpl(snap, parent, header); err != nil {
			return err
		}

		// verify the committed seals
		if err := verifyCommittedFields(snap, header, i.quorumSize(header.Number)); err != nil {
			return err
		}

		return nil
	*/
	return nil
}

func (d *Avail) ProcessHeaders(headers []*types.Header) error {
	return nil
}

func (d *Avail) GetBlockCreator(header *types.Header) (types.Address, error) {
	//return addressRecoverFromHeader(header)
	return header.Miner, nil
}

// PreStateCommit a hook to be called before finalizing state transition on inserting block
func (d *Avail) PreStateCommit(_header *types.Header, _txn *state.Transition) error {
	return nil
}

func (d *Avail) GetSyncProgression() *progress.Progression {
	return d.syncer.GetSyncProgression()
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

// TODO: This is just a demo implementation, to get miner & watch tower
// addresses working. Implementing bare minimum out of which, when working
// correctly we can extract into more proper functions in the future.
func getAccountData(address string) (*keystore.KeyStore, accounts.Account, *keystore.Key, error) {
	ks := keystore.NewKeyStore("./data/wallets", keystore.StandardScryptN, keystore.StandardScryptP)
	acc, err := ks.Find(accounts.Account{Address: common.HexToAddress(address)})
	if err != nil {
		return nil, accounts.Account{}, nil, fmt.Errorf("failure to load sequencer miner account: %s", err)
	}

	passpharse := "secret"
	keyjson, err := ks.Export(acc, passpharse, passpharse)
	if err != nil {
		return nil, accounts.Account{}, nil, err
	}

	privatekey, err := keystore.DecryptKey(keyjson, passpharse)
	if err != nil {
		return nil, accounts.Account{}, nil, err
	}

	return ks, acc, privatekey, err
}
