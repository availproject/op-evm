package devnet

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"net"
	"net/netip"
	"os"
	"path"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/0xPolygon/polygon-edge/chain"
	"github.com/0xPolygon/polygon-edge/command/server/config"
	"github.com/0xPolygon/polygon-edge/network"
	"github.com/0xPolygon/polygon-edge/secrets/helper"
	edge_server "github.com/0xPolygon/polygon-edge/server"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/hashicorp/go-hclog"
	"github.com/libp2p/go-libp2p/core/peer"
	consensus "github.com/maticnetwork/avail-settlement/consensus/avail"
	"github.com/maticnetwork/avail-settlement/pkg/avail"
	"github.com/maticnetwork/avail-settlement/pkg/common"
	pkg_config "github.com/maticnetwork/avail-settlement/pkg/config"
	"github.com/maticnetwork/avail-settlement/server"
)

type Context struct {
	servers []instance
}

type instance struct {
	nodeType    consensus.MechanismType
	accountPath string
	config      *edge_server.Config
	server      *server.Server
	fraudAddr   string
}

//go:embed genesis.json
var genesisBytes []byte

// StartNodes starts configured nodes
func StartNodes(logger hclog.Logger, bindAddr netip.Addr, availAddr, accountsPath string, nodeTypes ...consensus.MechanismType) (*Context, error) {
	ctx := &Context{}
	if err := createAvailAccounts(logger, availAddr, accountsPath, nodeTypes); err != nil {
		return nil, err
	}

	// Set up a [TCP] port allocator.
	pa := NewPortAllocator(bindAddr)

	nnh := newNodeNameHelper(accountsPath)
	for _, nt := range nodeTypes {
		cfg, err := configureNode(pa, nt)
		if err != nil {
			_ = pa.Release()
			return nil, err
		}

		fraudAddr, err := pa.Allocate()
		if err != nil {
			return nil, err
		}

		ctx.servers = append(ctx.servers, instance{
			nodeType:    nt,
			config:      cfg.Config,
			accountPath: nnh.nextAccountPath(nt),
			fraudAddr:   fraudAddr.String(),
		})
	}

	// Release allocated [TCP] ports to be used in Edge nodes.
	err := pa.Release()
	if err != nil {
		return nil, err
	}

	for i, si := range ctx.servers {
		bootnodes := make(map[consensus.MechanismType]string)

		// Adjust the blockchain Genesis spec.
		for j := range ctx.servers {
			if len(ctx.servers[j].config.Chain.Bootnodes) > 0 {
				// Collect one per node type. The logic here is that the
				// `bootstrap-sequencer` is the preferred one, but one `sequencer` is a
				// good second choice. If there are no sequencers -> return an error.
				//
				// In the chain spec, there is expected to be only one. See
				// `configureNode()` below.
				bootnodes[ctx.servers[j].nodeType] = ctx.servers[j].config.Chain.Bootnodes[0]
			}

			if i == j {
				// Skip `self` for the rest.
				continue
			}

			// Sync all premined accounts.
			for k, v := range ctx.servers[j].config.Chain.Genesis.Alloc {
				if _, exists := si.config.Chain.Genesis.Alloc[k]; !exists {
					si.config.Chain.Genesis.Alloc[k] = v
				}
			}
		}

		bootnodeAddr, exists := bootnodes[consensus.BootstrapSequencer]
		if !exists {
			bootnodeAddr, exists = bootnodes[consensus.Sequencer]
		}

		if !exists {
			return nil, fmt.Errorf("at least one sequencer must be configured")
		}

		// Reset the bootnode list.
		si.config.Chain.Bootnodes = []string{bootnodeAddr}
		si.config.Network.Chain.Bootnodes = []string{bootnodeAddr}

		srv, err := startNode(si.config, availAddr, si.accountPath, si.fraudAddr, si.nodeType)
		if err != nil {
			return nil, err
		}

		ctx.servers[i].server = srv

		logger.Info("started node", "i", i, "nodeType", si.nodeType)
	}

	logger.Info("all nodes started", "servers_count", len(ctx.servers))

	return ctx, nil
}

func configureNode(pa *PortAllocator, nodeType consensus.MechanismType) (_ *pkg_config.CustomServerConfig, err error) {
	rawConfig := config.DefaultConfig()
	rawConfig.DataDir, err = os.MkdirTemp("", "*")
	if err != nil {
		return nil, err
	}

	chainSpec := &chain.Chain{}
	if err := json.Unmarshal(genesisBytes, chainSpec); err != nil {
		return nil, err
	}

	// Reset bootnodes, in case there are any in the JSON file.
	chainSpec.Bootnodes = nil

	jsonRpcAddr, err := pa.Allocate()
	if err != nil {
		return nil, err
	}

	grpcAddr, err := pa.Allocate()
	if err != nil {
		return nil, err
	}

	libp2pAddr, err := pa.Allocate()
	if err != nil {
		return nil, err
	}

	secretsManager, err := helper.SetupLocalSecretsManager(rawConfig.DataDir)
	if err != nil {
		return nil, err
	}

	_, err = helper.InitBLSValidatorKey(secretsManager)
	if err != nil {
		return nil, err
	}

	minerAddr, err := helper.InitECDSAValidatorKey(secretsManager)
	if err != nil {
		return nil, err
	}

	libp2pKey, err := helper.InitNetworkingPrivateKey(secretsManager)
	if err != nil {
		return nil, err
	}

	p2pID, err := peer.IDFromPrivateKey(libp2pKey)
	if err != nil {
		return nil, err
	}

	var bootnodeAddr string
	switch {
	case libp2pAddr.Addr().Is4():
		bootnodeAddr = fmt.Sprintf("/ip4/%s/tcp/%d/p2p/%s", libp2pAddr.Addr().String(), libp2pAddr.Port(), p2pID)
	case libp2pAddr.Addr().Is6():
		bootnodeAddr = fmt.Sprintf("/ip6/%s/tcp/%d/p2p/%s", libp2pAddr.Addr().String(), libp2pAddr.Port(), p2pID)
	default:
		return nil, fmt.Errorf("invalid p2p network address: %q", libp2pAddr.String())
	}

	if nodeType == consensus.BootstrapSequencer || len(chainSpec.Bootnodes) == 0 {
		chainSpec.Bootnodes = append(chainSpec.Bootnodes, bootnodeAddr)
	}

	chainSpec.Genesis.Alloc[minerAddr] = &chain.GenesisAccount{
		Balance: big.NewInt(0).Mul(big.NewInt(1000), common.ETH),
	}

	cfg := &pkg_config.CustomServerConfig{
		Config: &edge_server.Config{
			Chain: chainSpec,
			JSONRPC: &edge_server.JSONRPC{
				JSONRPCAddr:              net.TCPAddrFromAddrPort(jsonRpcAddr),
				AccessControlAllowOrigin: rawConfig.Headers.AccessControlAllowOrigins,
			},
			GRPCAddr:   net.TCPAddrFromAddrPort(grpcAddr),
			LibP2PAddr: net.TCPAddrFromAddrPort(libp2pAddr),
			Telemetry:  new(edge_server.Telemetry),
			Network: &network.Config{
				NoDiscover:       rawConfig.Network.NoDiscover,
				Addr:             net.TCPAddrFromAddrPort(libp2pAddr),
				DataDir:          rawConfig.DataDir,
				MaxPeers:         rawConfig.Network.MaxPeers,
				MaxInboundPeers:  rawConfig.Network.MaxInboundPeers,
				MaxOutboundPeers: rawConfig.Network.MaxOutboundPeers,
				Chain:            chainSpec,
			},
			DataDir:            rawConfig.DataDir,
			Seal:               true, // Seal enables TxPool P2P gossiping
			PriceLimit:         rawConfig.TxPool.PriceLimit,
			MaxAccountEnqueued: 128,
			MaxSlots:           rawConfig.TxPool.MaxSlots,
			SecretsManager:     nil,
			RestoreFile:        nil,
			LogLevel:           hclog.Info,
			LogFilePath:        rawConfig.LogFilePath,
		},
		NodeType: nodeType.String(),
	}

	return cfg, nil
}

func startNode(cfg *edge_server.Config, availAddr, accountPath, fraudListenerAddr string, nodeType consensus.MechanismType) (*server.Server, error) {
	var bootnode bool
	if nodeType == consensus.BootstrapSequencer {
		bootnode = true
	}

	availAccount, err := avail.AccountFromFile(accountPath)
	if err != nil {
		log.Fatalf("failed to read Avail account from %q: %s\n", accountPath, err)
	}

	availClient, err := avail.NewClient(availAddr)
	if err != nil {
		log.Fatalf("failed to create Avail client: %s\n", err)
	}

	appID, err := avail.EnsureApplicationKeyExists(availClient, avail.ApplicationKey, availAccount)
	if err != nil {
		log.Fatalf("failed to get AppID from Avail: %s\n", err)
	}

	availSender := avail.NewSender(availClient, appID, availAccount)

	consensusCfg := consensus.Config{
		Bootnode:          bootnode,
		AvailAccount:      availAccount,
		AvailClient:       availClient,
		AvailSender:       availSender,
		AccountFilePath:   accountPath,
		FraudListenerAddr: "",
		NodeType:          string(nodeType),
		AvailAppID:        appID,
	}

	serverInstance, err := server.NewServer(cfg, consensusCfg)
	if err != nil {
		return nil, fmt.Errorf("failure to start node: %w", err)
	}

	return serverInstance, nil
}

type PortAllocator struct {
	bindAddr  netip.Addr
	listeners []net.Listener
}

func NewPortAllocator(bindAddr netip.Addr) *PortAllocator {
	return &PortAllocator{
		bindAddr: bindAddr,
	}
}

func (pa *PortAllocator) Allocate() (netip.AddrPort, error) {
	addrPort := netip.AddrPortFrom(pa.bindAddr, 0)
	lst, err := net.Listen("tcp", addrPort.String())
	if err != nil {
		return netip.AddrPort{}, err
	}

	pa.listeners = append(pa.listeners, lst)

	return netip.ParseAddrPort(lst.Addr().String())
}

func (pa *PortAllocator) Release() error {
	var lastErr error

	for _, l := range pa.listeners {
		err := l.Close()
		if err != nil {
			lastErr = err
			log.Printf("error: %#v", err)
		}
	}

	return lastErr
}

func (sc *Context) GethClient(nodeType consensus.MechanismType) (*ethclient.Client, error) {
	if len(sc.servers) == 0 {
		return nil, fmt.Errorf("no json-rpc URLs available")
	}
	addr, err := sc.FirstRPCAddrForNodeType(nodeType)
	if err != nil {
		return nil, err
	}

	return ethclient.Dial(fmt.Sprintf("http://%s/", addr))
}

func (sc *Context) Output(w io.Writer) {
	tw := tabwriter.NewWriter(w, 0, 0, 0, ' ', tabwriter.Debug)
	fmt.Fprintf(tw, "\t NODE TYPE \t JSONRPC URL \t FRAUD SERVER URL \t GRPC ADDR \t\n")
	for _, s := range sc.servers {
		fmt.Fprintf(tw, "\t %s \t http://%s/ \t http://%s/ \t %s \t\n",
			s.nodeType, s.config.JSONRPC.JSONRPCAddr, s.fraudAddr, s.config.GRPCAddr)
	}
	tw.Flush()
}

func (sc *Context) StopAll() {
	for _, srvInstance := range sc.servers {
		srvInstance.server.Close()
	}
}

// FirstRPCAddrForNodeType looks up and returns the url of the node for the node type
func (sc *Context) FirstRPCAddrForNodeType(nodeType consensus.MechanismType) (*net.TCPAddr, error) {
	for _, srv := range sc.servers {
		if srv.nodeType == nodeType {
			return sc.servers[0].config.JSONRPC.JSONRPCAddr, nil
		}
	}

	return nil, fmt.Errorf("no %s node present in the servers", nodeType)
}

func createAvailAccounts(logger hclog.Logger, availAddr, accountPath string, nodeTypes []consensus.MechanismType) error {
	nnh := newNodeNameHelper(accountPath)

	var accountWg sync.WaitGroup

	availClient, err := avail.NewClient(availAddr)
	if err != nil {
		return err
	}

	var nonceIncrement uint64
	errCh := make(chan error)
	for _, nt := range nodeTypes {
		accountWg.Add(1)

		go func(accountPath string, nonceIncrement uint64) {
			defer accountWg.Done()
			// Initiate creation of the avail account if not present
			err := createAvailAccount(logger, availClient, accountPath, nonceIncrement)
			if err != nil {
				errCh <- fmt.Errorf("failed to create new avail account: %w", err)
				return
			}
		}(nnh.nextAccountPath(nt), nonceIncrement)

		nonceIncrement++
		time.Sleep(250 * time.Millisecond)
	}

	logger.Info("Waiting for Avail accounts to be created...")
	wait := make(chan struct{})
	go func() {
		accountWg.Wait()
		wait <- struct{}{}
	}()
	select {
	case err := <-errCh:
		return err
	case <-wait:
	}

	logger.Info("Avail accounts created")
	return nil
}

func createAvailAccount(logger hclog.Logger, availClient avail.Client, accountPath string, nonceIncrement uint64) error {
	// If file exists, make sure that we return the file and not go through account creation process.
	// In rare cases, funds may be depleted but in that case we can erase files and run it again.
	// TODO: Potentially add lookup for account balance check and if it's too low, process with creation
	if _, err := os.Stat(accountPath); !errors.Is(err, os.ErrNotExist) {
		// In case that account path exists but is not visible in Avail (restart)
		// make sure to go through the process of the account creation.
		if ok, err := avail.AccountExistsFromMnemonic(availClient, accountPath); err == nil && ok {
			return nil
		}
	}

	availAccount, err := avail.NewAccount()
	if err != nil {
		return err
	}

	err = avail.DepositBalance(availClient, availAccount, 15*avail.AVL, nonceIncrement)
	if err != nil {
		return err
	}

	if _, err := avail.QueryAppID(availClient, avail.ApplicationKey); err != nil {
		if err == avail.ErrAppIDNotFound {
			_, err = avail.EnsureApplicationKeyExists(availClient, avail.ApplicationKey, availAccount)
			if err != nil {
				return err
			}
		}

		return err
	}

	logger.Info("Successfully deposited", "avl", 15, "to", availAccount.Address)

	if err := os.WriteFile(accountPath, []byte(availAccount.URI), 0o644); err != nil {
		return err
	}

	logger.Info("Successfully written mnemonic", "into", accountPath)

	return nil
}

type nodeNameHelper struct {
	accountsPath string
	nodeCounter  map[consensus.MechanismType]int
}

func newNodeNameHelper(accountsPath string) nodeNameHelper {
	return nodeNameHelper{
		accountsPath: accountsPath,
		nodeCounter:  make(map[consensus.MechanismType]int),
	}
}

func (h *nodeNameHelper) next(nodeType consensus.MechanismType) string {
	h.nodeCounter[nodeType]++
	return fmt.Sprintf("%s-%d", nodeType, h.nodeCounter[nodeType])
}

func (h *nodeNameHelper) nextAccountPath(nodeType consensus.MechanismType) string {
	return path.Join(h.accountsPath, h.next(nodeType))
}
