// Server package for the blockchain client.
package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	consensusPolyBFT "github.com/0xPolygon/polygon-edge/consensus/polybft"
	"github.com/0xPolygon/polygon-edge/server"
	avail_consensus "github.com/availproject/op-evm/consensus/avail"
	"github.com/availproject/op-evm/pkg/snapshot"

	"github.com/0xPolygon/polygon-edge/archive"
	"github.com/0xPolygon/polygon-edge/chain"
	"github.com/0xPolygon/polygon-edge/consensus"
	"github.com/0xPolygon/polygon-edge/consensus/polybft/statesyncrelayer"
	"github.com/0xPolygon/polygon-edge/consensus/polybft/wallet"
	"github.com/0xPolygon/polygon-edge/contracts"
	"github.com/0xPolygon/polygon-edge/crypto"
	"github.com/0xPolygon/polygon-edge/helper/common"
	"github.com/0xPolygon/polygon-edge/helper/progress"
	"github.com/0xPolygon/polygon-edge/jsonrpc"
	"github.com/0xPolygon/polygon-edge/network"
	"github.com/0xPolygon/polygon-edge/secrets"
	"github.com/0xPolygon/polygon-edge/server/proto"
	"github.com/0xPolygon/polygon-edge/state"
	itrie "github.com/0xPolygon/polygon-edge/state/immutable-trie"
	"github.com/0xPolygon/polygon-edge/state/runtime"
	"github.com/0xPolygon/polygon-edge/state/runtime/tracer"
	"github.com/0xPolygon/polygon-edge/txpool"
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/0xPolygon/polygon-edge/validate"
	"github.com/availproject/op-evm/pkg/blockchain"
	"github.com/hashicorp/go-hclog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/umbracle/ethgo"
	"google.golang.org/grpc"
)

// Error definitions.
var (
	errBlockTimeMissing = errors.New("block time configuration is missing")
	errBlockTimeInvalid = errors.New("block time configuration is invalid")
)

// Server struct defines the central manager of the blockchain client.
type Server struct {
	logger       hclog.Logger
	config       *server.Config
	state        state.State
	stateStorage itrie.Storage

	consensus consensus.Consensus

	// blockchain stack
	blockchain *blockchain.Blockchain
	chain      *chain.Chain

	// state executor
	executor *state.Executor

	// EVM & blockchain state snapshotter
	snapshotter snapshot.Snapshotter

	// jsonrpc stack
	jsonrpcServer *jsonrpc.JSONRPC

	// system grpc server
	grpcServer *grpc.Server

	// libp2p network
	network *network.Server

	// transaction pool
	txpool *txpool.TxPool

	prometheusServer *http.Server

	// secrets manager
	secretsManager secrets.SecretsManager

	// restore
	restoreProgression *progress.ProgressionWrapper

	// stateSyncRelayer is handling state syncs execution (Polybft exclusive)
	stateSyncRelayer *statesyncrelayer.StateSyncRelayer
}

// newFileLogger creates a logger instance that writes all logs to a specified file.
// If log file can't be created, it returns an error.
func newFileLogger(config *server.Config) (hclog.Logger, error) {
	logFileWriter, err := os.Create(config.LogFilePath)
	if err != nil {
		return nil, fmt.Errorf("could not create log file, %w", err)
	}

	return hclog.New(&hclog.LoggerOptions{
		Name:       "polygon",
		Level:      config.LogLevel,
		Output:     logFileWriter,
		JSONFormat: config.JSONLogFormat,
	}), nil
}

// newCLILogger returns a minimal logger instance that sends all logs to standard output.
func newCLILogger(config *server.Config) hclog.Logger {
	return hclog.New(&hclog.LoggerOptions{
		Name:       "polygon",
		Level:      config.LogLevel,
		JSONFormat: config.JSONLogFormat,
	})
}

// newLoggerFromConfig creates a logger that logs to a specified file.
// If log file is not set it outputs to standard output (console).
// If log file is specified, and it can't be created the server command will error out.
func newLoggerFromConfig(config *server.Config) (hclog.Logger, error) {
	if config.LogFilePath != "" {
		fileLoggerInstance, err := newFileLogger(config)
		if err != nil {
			return nil, err
		}

		return fileLoggerInstance, nil
	}

	return newCLILogger(config), nil
}

// NewServer creates a new minimal server, using the passed in configuration.
func NewServer(config *server.Config, consensusCfg avail_consensus.Config) (*Server, error) {
	logger, err := newLoggerFromConfig(config)
	if err != nil {
		return nil, fmt.Errorf("could not setup new logger instance, %w", err)
	}

	m := &Server{
		logger:             logger.Named("server"),
		config:             config,
		chain:              config.Chain,
		grpcServer:         grpc.NewServer(grpc.UnaryInterceptor(unaryInterceptor)),
		restoreProgression: progress.NewProgressionWrapper(progress.ChainSyncRestore),
	}

	m.logger.Info("Data dir", "path", config.DataDir)

	dirPaths := []string{
		"blockchain",
		"trie",
	}

	// Generate all the paths in the dataDir
	if err := common.SetupDataDir(config.DataDir, dirPaths, 0o770); err != nil {
		return nil, fmt.Errorf("failed to create data directories: %w", err)
	}

	if config.Telemetry.PrometheusAddr != nil {
		// Only setup telemetry if `PrometheusAddr` has been configured.
		if err := m.setupTelemetry(); err != nil {
			return nil, err
		}

		m.prometheusServer = m.startPrometheusServer(config.Telemetry.PrometheusAddr)
	}

	// Set up datadog profiler
	if ddErr := m.enableDataDogProfiler(); err != nil {
		m.logger.Error("DataDog profiler setup failed", "error", ddErr.Error())
	}

	// Set up the secrets manager
	if err := m.setupSecretsManager(); err != nil {
		return nil, fmt.Errorf("failed to set up the secrets manager: %w", err)
	}

	// start libp2p
	{
		netConfig := config.Network
		netConfig.Chain = m.config.Chain
		netConfig.DataDir = filepath.Join(m.config.DataDir, "libp2p")
		netConfig.SecretsManager = m.secretsManager

		network, err := network.NewServer(logger, netConfig)
		if err != nil {
			return nil, err
		}
		m.network = network
	}

	// start blockchain object
	stateStorage, err := itrie.NewLevelDBStorage(filepath.Join(m.config.DataDir, "trie"), logger)
	if err != nil {
		return nil, err
	}

	m.stateStorage = stateStorage

	blockchainDBPath := m.config.DataDir
	if blockchainDBPath != "" {
		blockchainDBPath = filepath.Join(m.config.DataDir, "blockchain")
	}

	snapshotter, blockchainStorage, wrappedStateStorage, err := snapshot.NewSnapshotter(logger, stateStorage, blockchainDBPath)
	if err != nil {
		return nil, err
	}

	m.snapshotter = snapshotter

	st := itrie.NewState(wrappedStateStorage)
	m.state = st

	m.executor = state.NewExecutor(config.Chain.Params, st, logger)

	initialStateRoot := types.ZeroHash

	genesisRoot, err := m.executor.WriteGenesis(config.Chain.Genesis.Alloc, initialStateRoot)
	if err != nil {
		return nil, err
	}

	// compute the genesis root state
	config.Chain.Genesis.StateRoot = genesisRoot

	// Use the london signer with eip-155 as a fallback one
	var signer crypto.TxSigner = crypto.NewLondonSigner(
		uint64(m.config.Chain.Params.ChainID),
		config.Chain.Params.Forks.IsActive(chain.Homestead, 0),
		crypto.NewEIP155Signer(
			uint64(m.config.Chain.Params.ChainID),
			config.Chain.Params.Forks.IsActive(chain.Homestead, 0),
		),
	)

	// blockchain object
	m.blockchain, err = blockchain.NewBlockchain(
		logger,
		blockchainStorage,
		config.Chain,
		nil,
		m.executor,
		signer,
	)
	if err != nil {
		return nil, err
	}

	m.executor.GetHash = m.blockchain.GetHashHelper

	{
		hub := &txpoolHub{
			state:      m.state,
			Blockchain: m.blockchain,
		}

		// start transaction pool
		m.txpool, err = txpool.NewTxPool(
			logger,
			m.chain.Params.Forks.At(0),
			hub,
			m.grpcServer,
			m.network,
			&txpool.Config{
				MaxSlots:           m.config.MaxSlots,
				PriceLimit:         m.config.PriceLimit,
				MaxAccountEnqueued: m.config.MaxAccountEnqueued,
			},
		)
		if err != nil {
			return nil, err
		}

		m.txpool.SetSigner(signer)
	}

	{
		// Setup consensus
		if err := m.setupConsensus(consensusCfg); err != nil {
			return nil, err
		}
		m.blockchain.SetConsensus(m.consensus)
	}

	// after consensus is done, we can mine the genesis block in blockchain
	// This is done because consensus might use a custom Hash function so we need
	// to wait for consensus because we do any block hashing like genesis
	if err := m.blockchain.ComputeGenesis(); err != nil {
		return nil, err
	}

	// initialize data in consensus layer
	if err := m.consensus.Initialize(); err != nil {
		return nil, err
	}

	// setup and start grpc server
	if err := m.setupGRPC(); err != nil {
		return nil, err
	}

	if err := m.network.Start(); err != nil {
		return nil, err
	}

	// TxPool Start() is only listener method that notifies specific function targets
	// if channel information is met. It should be started prior JSONRPC as new tx
	// can be applied in between (edge case) and result in tx hitting blackhole.
	// When network starts, txpool should already be started so if any tx is pushed from any node
	// we know that at that time, txpool is already ready and will receive the tx properly.
	m.txpool.Start()

	// setup and start jsonrpc server
	if err := m.setupJSONRPC(); err != nil {
		return nil, err
	}

	// restore archive data before starting
	if err := m.restoreChain(); err != nil {
		return nil, err
	}

	// start relayer
	if config.Relayer {
		if err := m.setupRelayer(); err != nil {
			return nil, err
		}
	}

	// start consensus
	if err := m.consensus.Start(); err != nil {
		return nil, err
	}

	return m, nil
}

// unaryInterceptor is a GRPC interceptor that validates requests.
func unaryInterceptor(
	ctx context.Context,
	req interface{},
	_ *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	// Validate request
	if err := validate.ValidateRequest(req); err != nil {
		return nil, err
	}

	return handler(ctx, req)
}

// restoreChain restores the blockchain from a backup file.
func (s *Server) restoreChain() error {
	if s.config.RestoreFile == nil {
		return nil
	}

	if err := archive.RestoreChain(s.blockchain, *s.config.RestoreFile, s.restoreProgression); err != nil {
		return err
	}

	return nil
}

// txpoolHub provides an interface between the transaction pool and the blockchain.
type txpoolHub struct {
	state state.State
	*blockchain.Blockchain
}

// getAccountImpl retrieves the account state for a specific address at a given root hash.
// This is a utility function used in GetNonce and GetBalance methods.
func getAccountImpl(state state.State, root types.Hash, addr types.Address) (*state.Account, error) {
	snap, err := state.NewSnapshotAt(root)
	if err != nil {
		return nil, fmt.Errorf("unable to get snapshot for root '%s': %w", root, err)
	}

	account, err := snap.GetAccount(addr)
	if err != nil {
		return nil, err
	}

	if account == nil {
		return nil, jsonrpc.ErrStateNotFound
	}

	return account, nil
}

// GetNonce retrieves the nonce for an account identified by the given address and at the given root hash.
// Returns 0 if the account state cannot be fetched.
func (t *txpoolHub) GetNonce(root types.Hash, addr types.Address) uint64 {
	account, err := getAccountImpl(t.state, root, addr)
	if err != nil {
		return 0
	}

	return account.Nonce
}

// GetBalance retrieves the balance for an account identified by the given address and at the given root hash.
// Returns big.NewInt(0) and potentially an error if the account state cannot be fetched.
func (t *txpoolHub) GetBalance(root types.Hash, addr types.Address) (*big.Int, error) {
	account, err := getAccountImpl(t.state, root, addr)
	if err != nil {
		if errors.Is(err, jsonrpc.ErrStateNotFound) {
			return big.NewInt(0), nil
		}

		return big.NewInt(0), err
	}

	return account.Balance, nil
}

// GetBlockByHash retrieves a block identified by the given hash from the blockchain.
// If 'full' is set to true, then the transactions are populated with all their details, else only their hashes are returned.
func (t *txpoolHub) GetBlockByHash(h types.Hash, full bool) (*types.Block, bool) {
	return t.Blockchain.GetBlockByHash(h, full)
}

// Header retrieves the latest header from the blockchain.
func (t *txpoolHub) Header() *types.Header {
	return t.Blockchain.Header()
}

// setupSecretsManager sets up the secrets manager based on the server's configuration.
// It supports a local secrets manager which stores secrets in a local directory.
// If the type of secrets manager specified in the configuration is not found, an error is returned.
func (s *Server) setupSecretsManager() error {
	secretsManagerConfig := s.config.SecretsManager
	if secretsManagerConfig == nil {
		// No config provided, use default
		secretsManagerConfig = &secrets.SecretsManagerConfig{
			Type: secrets.Local,
		}
	}

	secretsManagerType := secretsManagerConfig.Type
	secretsManagerParams := &secrets.SecretsManagerParams{
		Logger: s.logger,
	}

	if secretsManagerType == secrets.Local {
		// Only the base directory is required for
		// the local secrets manager
		secretsManagerParams.Extra = map[string]interface{}{
			secrets.Path: s.config.DataDir,
		}
	}

	// Grab the factory method
	secretsManagerFactory, ok := secretsManagerBackends[secretsManagerType]
	if !ok {
		return fmt.Errorf("secrets manager type '%s' not found", secretsManagerType)
	}

	// Instantiate the secrets manager
	secretsManager, factoryErr := secretsManagerFactory(
		secretsManagerConfig,
		secretsManagerParams,
	)

	if factoryErr != nil {
		return fmt.Errorf("unable to instantiate secrets manager, %w", factoryErr)
	}

	s.secretsManager = secretsManager

	return nil
}

// setupConsensus sets up the consensus mechanism for the server.
// It reads the consensus engine name from the chain parameters in the server's configuration.
// If the consensus engine is neither DummyConsensus nor DevConsensus, it also tries to extract the block time from the engine's configuration.
// After that, it creates the consensus configuration and fills it with all necessary dependencies.
// Finally, it creates a new consensus mechanism using the created configuration.
// If the creation of the consensus mechanism fails, it returns an error.
func (s *Server) setupConsensus(consensusCfg avail_consensus.Config) error {
	engineName := s.config.Chain.Params.GetEngine()
	engineConfig, ok := s.config.Chain.Params.Engine[engineName].(map[string]interface{})
	if !ok {
		engineConfig = map[string]interface{}{}
	}

	var (
		blockTime = common.Duration{Duration: 0}
		err       error
	)

	if engineName != string(server.DummyConsensus) && engineName != string(server.DevConsensus) {
		blockTime, err = ExtractBlockTime(engineConfig)
		if err != nil {
			return err
		}
	}

	config := &consensus.Config{
		Params: s.config.Chain.Params,
		Config: engineConfig,
		Path:   filepath.Join(s.config.DataDir, "consensus"),
	}

	// Fill-in server dependencies.
	consensusCfg.Blockchain = s.blockchain
	consensusCfg.BlockTime = uint64(blockTime.Seconds())
	consensusCfg.Chain = s.config.Chain
	consensusCfg.Config = config
	consensusCfg.Context = context.Background()
	consensusCfg.Executor = s.executor
	consensusCfg.Logger = s.logger
	consensusCfg.Network = s.network
	consensusCfg.TxPool = s.txpool
	consensusCfg.SecretsManager = s.secretsManager
	consensusCfg.Snapshotter = s.snapshotter
	consensusCfg.NumBlockConfirmations = s.config.NumBlockConfirmations

	consensus, err := avail_consensus.New(consensusCfg)
	if err != nil {
		return err
	}

	s.consensus = consensus

	return nil
}

// ExtractBlockTime attempts to extract a blockTime parameter from a given consensus engine configuration map.
// It returns a Duration type. If blockTime is missing or invalid, it returns an appropriate error.
func ExtractBlockTime(engineConfig map[string]interface{}) (common.Duration, error) {
	blockTimeGeneric, ok := engineConfig["blockTime"]
	if !ok {
		return common.Duration{}, errBlockTimeMissing
	}

	blockTimeRaw, err := json.Marshal(blockTimeGeneric)
	if err != nil {
		return common.Duration{}, errBlockTimeInvalid
	}

	var blockTime common.Duration

	if err := json.Unmarshal(blockTimeRaw, &blockTime); err != nil {
		return common.Duration{}, errBlockTimeInvalid
	}

	if blockTime.Seconds() < 1 {
		return common.Duration{}, errBlockTimeInvalid
	}

	return blockTime, nil
}

// setupRelayer initializes the server's relayer component.
// It first creates a new account from the secret stored in the server's secrets manager.
// Then it retrieves the PolyBFT configuration from the server's chain configuration.
// Afterwards, it determines the starting blocks for tracking events for the State Receiver contract.
// Finally, it creates a new relayer, starts it and in case of any errors, returns them.
func (s *Server) setupRelayer() error {
	account, err := wallet.NewAccountFromSecret(s.secretsManager)
	if err != nil {
		return fmt.Errorf("failed to create account from secret: %w", err)
	}

	polyBFTConfig, err := consensusPolyBFT.GetPolyBFTConfig(s.config.Chain)
	if err != nil {
		return fmt.Errorf("failed to extract polybft config: %w", err)
	}

	trackerStartBlockConfig := map[types.Address]uint64{}
	if polyBFTConfig.Bridge != nil {
		trackerStartBlockConfig = polyBFTConfig.Bridge.EventTrackerStartBlocks
	}

	relayer := statesyncrelayer.NewRelayer(
		s.config.DataDir,
		s.config.JSONRPC.JSONRPCAddr.String(),
		ethgo.Address(contracts.StateReceiverContract),
		trackerStartBlockConfig[contracts.StateReceiverContract],
		s.logger.Named("relayer"),
		wallet.NewEcdsaSigner(wallet.NewKey(account)),
	)

	// start relayer
	if err := relayer.Start(); err != nil {
		return fmt.Errorf("failed to start relayer: %w", err)
	}

	return nil
}

// jsonRPCHub represents a hub for handling JSON-RPC requests. It includes the
// blockchain, transaction pool, executor, network server, consensus, and bridge
// data provider as embedded types.
type jsonRPCHub struct {
	state              state.State
	restoreProgression *progress.ProgressionWrapper

	*blockchain.Blockchain
	*txpool.TxPool
	*state.Executor
	*network.Server
	consensus.Consensus
	consensus.BridgeDataProvider
}

// GetPeers returns the number of peers connected to the server.
func (j *jsonRPCHub) GetPeers() int {
	return len(j.Server.Peers())
}

// GetAccount retrieves the account with the given address from the given state root.
// It returns the account in JSON-RPC format, or an error if one occurs.
func (j *jsonRPCHub) GetAccount(root types.Hash, addr types.Address) (*jsonrpc.Account, error) {
	acct, err := getAccountImpl(j.state, root, addr)
	if err != nil {
		return nil, err
	}

	account := &jsonrpc.Account{
		Nonce:   acct.Nonce,
		Balance: new(big.Int).Set(acct.Balance),
	}

	return account, nil
}

// GetForksInTime retrieves the active forks at the specified block number.
func (j *jsonRPCHub) GetForksInTime(blockNumber uint64) chain.ForksInTime {
	return j.Executor.GetForksInTime(blockNumber)
}

// GetStorage retrieves the storage at the given slot for the specified account address from the given state root.
// It returns the storage value as a byte slice, or an error if one occurs.
func (j *jsonRPCHub) GetStorage(stateRoot types.Hash, addr types.Address, slot types.Hash) ([]byte, error) {
	account, err := getAccountImpl(j.state, stateRoot, addr)
	if err != nil {
		return nil, err
	}

	snap, err := j.state.NewSnapshotAt(stateRoot)
	if err != nil {
		return nil, err
	}

	res := snap.GetStorage(addr, account.Root, slot)

	return res.Bytes(), nil
}

// GetCode retrieves the contract code for the account with the given address from the given state root.
// It returns the code as a byte slice, or an error if one occurs.
func (j *jsonRPCHub) GetCode(root types.Hash, addr types.Address) ([]byte, error) {
	account, err := getAccountImpl(j.state, root, addr)
	if err != nil {
		return nil, err
	}

	code, ok := j.state.GetCode(types.BytesToHash(account.CodeHash))
	if !ok {
		return nil, fmt.Errorf("unable to fetch code")
	}

	return code, nil
}

// ApplyTxn applies a given transaction on the state defined by the provided header.
// If an override is provided, it is used to update the state before the transaction is applied.
// The method returns the result of executing the transaction, or an error if one occurs.
func (j *jsonRPCHub) ApplyTxn(
	header *types.Header,
	txn *types.Transaction,
	override types.StateOverride,
) (result *runtime.ExecutionResult, err error) {
	blockCreator, err := j.Blockchain.GetConsensus().GetBlockCreator(header)
	if err != nil {
		return nil, err
	}

	transition, err := j.BeginTxn(header.StateRoot, header, blockCreator)
	if err != nil {
		return
	}

	if override != nil {
		if err = transition.WithStateOverride(override); err != nil {
			return
		}
	}

	result, err = transition.Apply(txn)

	return
}

// TraceBlock traces all transactions in the given block using the specified tracer and returns the results.
// An error is returned if the block is a genesis block, the parent header is not found, or an error occurs while applying transactions.
func (j *jsonRPCHub) TraceBlock(
	block *types.Block,
	tracer tracer.Tracer,
) ([]interface{}, error) {
	if block.Number() == 0 {
		return nil, errors.New("genesis block can't have transaction")
	}

	parentHeader, ok := j.Blockchain.GetHeaderByHash(block.ParentHash())
	if !ok {
		return nil, errors.New("parent header not found")
	}

	blockCreator, err := j.Blockchain.GetConsensus().GetBlockCreator(block.Header)
	if err != nil {
		return nil, err
	}

	transition, err := j.BeginTxn(parentHeader.StateRoot, block.Header, blockCreator)
	if err != nil {
		return nil, err
	}

	transition.SetTracer(tracer)

	results := make([]interface{}, len(block.Transactions))

	for idx, tx := range block.Transactions {
		tracer.Clear()

		if _, err := transition.Apply(tx); err != nil {
			return nil, err
		}

		if results[idx], err = tracer.GetResult(); err != nil {
			return nil, err
		}
	}

	return results, nil
}

// TraceTxn traces a specified transaction in the given block using the provided tracer and returns the result.
// An error is returned if the block is a genesis block, the parent header or the target transaction is not found, or an error occurs while applying transactions.
func (j *jsonRPCHub) TraceTxn(
	block *types.Block,
	targetTxHash types.Hash,
	tracer tracer.Tracer,
) (interface{}, error) {
	if block.Number() == 0 {
		return nil, errors.New("genesis block can't have transaction")
	}

	parentHeader, ok := j.GetHeaderByHash(block.ParentHash())
	if !ok {
		return nil, errors.New("parent header not found")
	}

	blockCreator, err := j.GetConsensus().GetBlockCreator(block.Header)
	if err != nil {
		return nil, err
	}

	transition, err := j.BeginTxn(parentHeader.StateRoot, block.Header, blockCreator)
	if err != nil {
		return nil, err
	}

	var targetTx *types.Transaction

	for _, tx := range block.Transactions {
		if tx.Hash == targetTxHash {
			targetTx = tx

			break
		}

		// Execute transactions without tracer until reaching the target transaction
		if _, err := transition.Apply(tx); err != nil {
			return nil, err
		}
	}

	if targetTx == nil {
		return nil, errors.New("target tx not found")
	}

	transition.SetTracer(tracer)

	if _, err := transition.Apply(targetTx); err != nil {
		return nil, err
	}

	return tracer.GetResult()
}

// TraceCall traces a call made by the given transaction on the state defined by the parent header using the provided tracer.
// It returns the result of the call, or an error if one occurs.
func (j *jsonRPCHub) TraceCall(
	tx *types.Transaction,
	parentHeader *types.Header,
	tracer tracer.Tracer,
) (interface{}, error) {
	blockCreator, err := j.GetConsensus().GetBlockCreator(parentHeader)
	if err != nil {
		return nil, err
	}

	transition, err := j.BeginTxn(parentHeader.StateRoot, parentHeader, blockCreator)
	if err != nil {
		return nil, err
	}

	transition.SetTracer(tracer)

	if _, err := transition.Apply(tx); err != nil {
		return nil, err
	}

	return tracer.GetResult()
}

func (j *jsonRPCHub) GetSyncProgression() *progress.Progression {
	// restore progression
	if restoreProg := j.restoreProgression.GetProgression(); restoreProg != nil {
		return restoreProg
	}

	// consensus sync progression
	if consensusSyncProg := j.Consensus.GetSyncProgression(); consensusSyncProg != nil {
		return consensusSyncProg
	}

	return nil
}

// SETUP //

// setupJSONRPC initializes the JSONRPC server based on the server's
// configuration. It uses the server's existing services (like state, blockchain,
// txpool, executor, and others) to create a new jsonRPCHub. It then constructs
// a new JSONRPC server and assigns it to the server's jsonrpcServer property.
//
// If an error occurs while creating the JSONRPC server, it is returned immediately.
// Otherwise, the method returns nil.
func (s *Server) setupJSONRPC() error {
	hub := &jsonRPCHub{
		state:              s.state,
		restoreProgression: s.restoreProgression,
		Blockchain:         s.blockchain,
		TxPool:             s.txpool,
		Executor:           s.executor,
		Consensus:          s.consensus,
		Server:             s.network,
		BridgeDataProvider: s.consensus.GetBridgeProvider(),
	}

	conf := &jsonrpc.Config{
		Store:                    hub,
		Addr:                     s.config.JSONRPC.JSONRPCAddr,
		ChainID:                  uint64(s.config.Chain.Params.ChainID),
		ChainName:                s.chain.Name,
		AccessControlAllowOrigin: s.config.JSONRPC.AccessControlAllowOrigin,
		PriceLimit:               s.config.PriceLimit,
		BatchLengthLimit:         s.config.JSONRPC.BatchLengthLimit,
		BlockRangeLimit:          s.config.JSONRPC.BlockRangeLimit,
	}

	srv, err := jsonrpc.NewJSONRPC(s.logger, conf)
	if err != nil {
		return err
	}

	s.jsonrpcServer = srv

	return nil
}

// setupGRPC initializes the gRPC server and begins listening on the TCP address
// specified in the server's configuration. It registers a systemService instance
// with the server and starts a goroutine that serves incoming requests indefinitely.
//
// If an error occurs while setting up the server, it is returned immediately.
// Otherwise, the method returns nil.
func (s *Server) setupGRPC() error {
	proto.RegisterSystemServer(s.grpcServer, &systemService{server: s})

	lis, err := net.Listen("tcp", s.config.GRPCAddr.String())
	if err != nil {
		return err
	}

	// Start server with infinite retries
	go func() {
		if err := s.grpcServer.Serve(lis); err != nil {
			s.logger.Error(err.Error())
		}
	}()

	s.logger.Info("GRPC server running", "addr", s.config.GRPCAddr.String())

	return nil
}

// Chain retrieves the server's Chain instance. Chain represents the blockchain
// associated with this server.
func (s *Server) Chain() *chain.Chain {
	return s.chain
}

// JoinPeer attempts to add a new peer to the server's network. The peer is
// identified by the provided multiaddress. If an error occurs while joining the
// peer, it is returned immediately.
func (s *Server) JoinPeer(rawPeerMultiaddr string) error {
	return s.network.JoinPeer(rawPeerMultiaddr)
}

// Close shuts down all components of the server, including the blockchain,
// networking layer, consensus layer, and state storage. If a Prometheus server
// is running, it is also shut down. Errors during shutdown are logged but not
// returned, as the method always succeeds.
func (s *Server) Close() {
	// Close the blockchain layer
	if err := s.blockchain.Close(); err != nil {
		s.logger.Error("failed to close blockchain", "error", err.Error())
	}

	// Close the networking layer
	if err := s.network.Close(); err != nil {
		s.logger.Error("failed to close networking", "error", err.Error())
	}

	// Close the consensus layer
	if err := s.consensus.Close(); err != nil {
		s.logger.Error("failed to close consensus", "error", err.Error())
	}

	// Close the state storage
	if err := s.stateStorage.Close(); err != nil {
		s.logger.Error("failed to close storage for trie", "error", err.Error())
	}

	if s.prometheusServer != nil {
		if err := s.prometheusServer.Shutdown(context.Background()); err != nil {
			s.logger.Error("Prometheus server shutdown error", err)
		}
	}

	// Stop state sync relayer
	if s.stateSyncRelayer != nil {
		s.stateSyncRelayer.Stop()
	}

	// Close the txpool's main loop
	s.txpool.Close()

	// Close DataDog profiler
	s.closeDataDogProfiler()
}

// Entry represents a configuration entry for a consensus protocol. It includes
// an Enabled field that indicates whether the entry is in use, and a Config
// field that holds the entry's configuration data. The Config field is a map
// of string keys to values of any type.
type Entry struct {
	Enabled bool
	Config  map[string]interface{}
}

// startPrometheusServer creates and starts a new Prometheus server that listens
// on the provided TCP address. The server uses the default Prometheus registerer
// and gatherer, and has a read header timeout of 60 seconds. A log message is
// written when the server starts. If an error occurs while the server is running,
// it is logged and the server is shut down.
//
// The method returns the created *http.Server instance.
func (s *Server) startPrometheusServer(listenAddr *net.TCPAddr) *http.Server {
	srv := &http.Server{
		Addr: listenAddr.String(),
		Handler: promhttp.InstrumentMetricHandler(
			prometheus.DefaultRegisterer, promhttp.HandlerFor(
				prometheus.DefaultGatherer,
				promhttp.HandlerOpts{},
			),
		),
		ReadHeaderTimeout: 60 * time.Second,
	}

	s.logger.Info("Prometheus server started", "addr=", listenAddr.String())

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				s.logger.Error("Prometheus HTTP server ListenAndServe", "error", err)
			}
		}
	}()

	return srv
}
