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
	"github.com/0xPolygon/polygon-edge/protocol"
	"github.com/0xPolygon/polygon-edge/secrets"
	"github.com/0xPolygon/polygon-edge/state"
	"github.com/0xPolygon/polygon-edge/txpool"
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	"github.com/hashicorp/go-hclog"
	"github.com/maticnetwork/avail-settlement/pkg/avail"
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
		state:          &currentState{},
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
func (d *Avail) Start() error {

	// Start the syncer
	d.syncer.Start()

	if d.nodeType == Sequencer {
		go d.runSequencer()
	}

	if d.nodeType == Validator {
		go d.runValidator()
	}

	return nil
}

func (d *Avail) nextNotify() chan struct{} {
	if d.interval == 0 {
		d.interval = 10
	}

	go func() {
		<-time.After(time.Duration(d.interval) * time.Second)
		d.notifyCh <- struct{}{}
	}()

	return d.notifyCh
}

func (d *Avail) runSequencer() {
	d.logger.Info("sequencer started")

	for {
		// wait until there is a new txn
		select {
		case <-d.nextNotify():
		case <-d.closeCh:
			return
		}

		// There are new transactions in the pool, try to seal them
		header := d.blockchain.Header()
		if err := d.writeNewBlock(header); err != nil {
			d.logger.Error("failed to mine block", "err", err)
		}
	}
}

func (d *Avail) runValidator() {
	d.logger.Info("validator started")
}

// TODO:
func (d *Avail) runWatchtower() {
	d.logger.Info("watch tower started")
}

// createKey sets the validator's private key from the secrets manager
func (d *Avail) createKey() error {
	//d.msgQueue = newMsgQueue()
	d.closeCh = make(chan struct{})
	d.updateCh = make(chan struct{})

	if d.validatorKey == nil {
		// Check if the validator key is initialized
		var key *ecdsa.PrivateKey

		if d.secretsManager.HasSecret(secrets.ValidatorKey) {
			// The validator key is present in the secrets manager, load it
			validatorKey, readErr := crypto.ReadConsensusKey(d.secretsManager)
			if readErr != nil {
				return fmt.Errorf("unable to read validator key from Secrets Manager, %w", readErr)
			}

			key = validatorKey
		} else {
			// The validator key is not present in the secrets manager, generate it
			validatorKey, validatorKeyEncoded, genErr := crypto.GenerateAndEncodePrivateKey()
			if genErr != nil {
				return fmt.Errorf("unable to generate validator key for Secrets Manager, %w", genErr)
			}

			// Save the key to the secrets manager
			saveErr := d.secretsManager.SetSecret(secrets.ValidatorKey, validatorKeyEncoded)
			if saveErr != nil {
				return fmt.Errorf("unable to save validator key to Secrets Manager, %w", saveErr)
			}

			key = validatorKey
		}

		d.validatorKey = key
		d.validatorKeyAddr = crypto.PubKeyToAddress(&key.PublicKey)
	}

	return nil
}

type transitionInterface interface {
	Write(txn *types.Transaction) error
}

func (d *Avail) writeTransactions(gasLimit uint64, transition transitionInterface) []*types.Transaction {
	var successful []*types.Transaction

	d.txpool.Prepare()

	for {
		tx := d.txpool.Peek()
		if tx == nil {
			break
		}

		if tx.ExceedsBlockGasLimit(gasLimit) {
			d.txpool.Drop(tx)

			continue
		}

		if err := transition.Write(tx); err != nil {
			if _, ok := err.(*state.GasLimitReachedTransitionApplicationError); ok { // nolint:errorlint
				break
			} else if appErr, ok := err.(*state.TransitionApplicationError); ok && appErr.IsRecoverable { // nolint:errorlint
				d.txpool.Demote(tx)
			} else {
				d.txpool.Drop(tx)
			}

			continue
		}

		// no errors, pop the tx from the pool
		d.txpool.Pop(tx)

		successful = append(successful, tx)
	}

	d.logger.Info("picked out txns from pool", "num", len(successful), "remaining", d.txpool.Length())

	return successful
}

// writeNewBLock generates a new block based on transactions from the pool,
// and writes them to the blockchain
func (d *Avail) writeNewBlock(parent *types.Header) error {
	header := &types.Header{
		ParentHash: parent.Hash,
		Number:     parent.Number + 1,
		Miner:      types.Address{},
		Nonce:      types.Nonce{},
		GasLimit:   parent.GasLimit, // Inherit from parent for now, will need to adjust dynamically later.
		Timestamp:  uint64(time.Now().Unix()),
	}

	// calculate gas limit based on parent header
	gasLimit, err := d.blockchain.CalculateGasLimit(header.Number)
	if err != nil {
		d.logger.Info("FAILING HERE? 1")
		return err
	}

	header.GasLimit = gasLimit

	// set the timestamp
	parentTime := time.Unix(int64(parent.Timestamp), 0)
	headerTime := parentTime.Add(d.blockTime)

	if headerTime.Before(time.Now()) {
		headerTime = time.Now()
	}

	header.Timestamp = uint64(headerTime.Unix())

	miner, err := d.GetBlockCreator(header)
	if err != nil {
		d.logger.Info("FAILING HERE? 2")
		return err
	}

	transition, err := d.executor.BeginTxn(parent.StateRoot, header, miner)
	if err != nil {
		d.logger.Info("FAILING HERE? 3")
		return err
	}

	txns := d.writeTransactions(gasLimit, transition)

	// Commit the changes
	_, root := transition.Commit()

	// Update the header
	header.StateRoot = root
	header.GasUsed = transition.TotalGas()

	// Build the actual block
	// The header hash is computed inside buildBlock
	block := consensus.BuildBlock(consensus.BuildBlockParams{
		Header:   header,
		Txns:     txns,
		Receipts: transition.Receipts(),
	})

	if err := d.blockchain.VerifyFinalizedBlock(block); err != nil {
		d.logger.Info("FAILING HERE? 4")
		return err
	}

	// Write block to the avail
	if err := d.sendBlockToAvail(block); err != nil {
		d.logger.Info("FAILING HERE? 5")
		return err
	}

	// Write the block to the blockchain
	if err := d.blockchain.WriteBlock(block); err != nil {
		d.logger.Info("FAILING HERE? 6")
		return err
	}

	// after the block has been written we reset the txpool so that
	// the old transactions are removed
	d.txpool.ResetWithHeaders(block.Header)

	fmt.Printf("Written block information: %+v", block.Header)

	return nil
}

func (d *Avail) sendBlockToAvail(block *types.Block) error {
	sender := avail.NewSender(d.availClient, signature.TestKeyringPairAlice)
	d.logger.Info("Submitting block to avail...")
	hash, err := sender.SubmitDataWithoutWatch(block.MarshalRLP())
	if err != nil {
		d.logger.Error("Error while submitting data to avail", err)
		return err
	}
	d.logger.Info("Submitted block to avail", "block", block.Header.Hash, "avail_block", hash.Hex())
	return nil
}

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
	// All blocks are valid
	return nil
}

func (d *Avail) ProcessHeaders(headers []*types.Header) error {
	return nil
}

func (d *Avail) GetBlockCreator(header *types.Header) (types.Address, error) {
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

/**
package main
​
import (
	"flag"
	"log"
	"sync"
​
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	"github.com/maticnetwork/avail-settlement/pkg/avail"
)
​
type dataHandler struct {
	WG sync.WaitGroup
}
​
func (bh *dataHandler) HandleData(bs []byte) error {
	log.Printf("block handler: received batch w/ %d bytes\n", len(bs))
	bh.WG.Done()
	return nil
}
func (bh *dataHandler) HandleError(err error) {
	log.Printf("block handler: error %#v\n", err)
	bh.WG.Done()
}
​
var rpcUrlFlag = flag.String("rpc-url", "ws://127.0.0.1:9944/v1/json-rpc", "Avail JSON-RPC URL")
​
func main() {
	flag.Parse()
​
	client, err := avail.NewClient(*rpcUrlFlag)
	if err != nil {
		log.Fatal(err)
	}
​
	sender := avail.NewSender(client, signature.TestKeyringPairAlice)
	handler := &dataHandler{}
	handler.WG.Add(1)
	watcher, err := avail.NewBlockDataWatcher(client, avail.BridgeAppID, handler)
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Stop()
​
	data := []byte("foobar quux")
	f := sender.SubmitData(data)
	_, err = f.Result()
	if err != nil {
		log.Fatal(err)
	}
​
	log.Println("got Result. All good.")
	handler.WG.Wait()
}
**/
