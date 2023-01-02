package tests

import (
	"fmt"
	"log"
	"math/big"
	"net"
	"net/netip"
	"net/url"
	"testing"

	"github.com/0xPolygon/polygon-edge/chain"
	"github.com/0xPolygon/polygon-edge/command/server/config"
	"github.com/0xPolygon/polygon-edge/network"
	"github.com/0xPolygon/polygon-edge/secrets/helper"
	"github.com/0xPolygon/polygon-edge/server"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/hashicorp/go-hclog"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/maticnetwork/avail-settlement/consensus/avail"
	"github.com/maticnetwork/avail-settlement/pkg/common"
)

type Context struct {
	servers     []instance
	jsonRPCURLs []*url.URL
}

type instance struct {
	nodeType avail.MechanismType
	config   *server.Config
	server   *server.Server
}

// StartServers starts configured nodes
func StartNodes(t testing.TB, bindAddr netip.Addr, genesisCfgPath, availAddr string, nodeTypes ...avail.MechanismType) (*Context, error) {
	t.Helper()

	ctx := &Context{}

	// Set up a [TCP] port allocator.
	pa := NewPortAllocator(bindAddr)

	for _, nt := range nodeTypes {
		cfg, err := configureNode(t, pa, nt, genesisCfgPath)
		if err != nil {
			_ = pa.Release()
			return nil, err
		}

		si := instance{
			nodeType: nt,
			config:   cfg,
		}

		u, err := url.Parse("http://" + cfg.JSONRPC.JSONRPCAddr.String() + "/")
		if err != nil {
			return nil, err
		}
		ctx.jsonRPCURLs = append(ctx.jsonRPCURLs, u)

		ctx.servers = append(ctx.servers, si)
	}

	// Release allocated [TCP] ports to be used in Edge nodes.
	err := pa.Release()
	if err != nil {
		return nil, err
	}

	for i, si := range ctx.servers {
		bootnodes := make(map[avail.MechanismType]string)

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

		bootnodeAddr, exists := bootnodes[avail.BootstrapSequencer]
		if !exists {
			bootnodeAddr, exists = bootnodes[avail.Sequencer]
		}

		if !exists {
			return nil, fmt.Errorf("at least one sequencer must be configured")
		}

		// Reset the bootnode list.
		si.config.Chain.Bootnodes = []string{bootnodeAddr}
		si.config.Network.Chain.Bootnodes = []string{bootnodeAddr}

		srv, err := startNode(si.config, availAddr, si.nodeType)
		if err != nil {
			return nil, err
		}

		ctx.servers[i].server = srv

		t.Logf("%d: started node %q", i, si.nodeType)
	}

	t.Logf("all %d nodes started", len(ctx.servers))

	return ctx, nil
}

func configureNode(t testing.TB, pa *PortAllocator, nodeType avail.MechanismType, genesisCfgPath string) (*server.Config, error) {
	t.Helper()

	rawConfig := config.DefaultConfig()
	rawConfig.DataDir = t.TempDir()
	rawConfig.GenesisPath = genesisCfgPath

	chainSpec, err := chain.Import(genesisCfgPath)
	if err != nil {
		return nil, err
	}

	// Reset bootnodes, in case there're any in the JSON file.
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

	if nodeType == avail.BootstrapSequencer || len(chainSpec.Bootnodes) == 0 {
		chainSpec.Bootnodes = append(chainSpec.Bootnodes, bootnodeAddr)
	}

	chainSpec.Genesis.Alloc[minerAddr] = &chain.GenesisAccount{
		Balance: big.NewInt(0).Mul(big.NewInt(1000), common.ETH),
	}

	cfg := &server.Config{
		Chain: chainSpec,
		JSONRPC: &server.JSONRPC{
			JSONRPCAddr:              net.TCPAddrFromAddrPort(jsonRpcAddr),
			AccessControlAllowOrigin: rawConfig.Headers.AccessControlAllowOrigins,
		},
		GRPCAddr:   net.TCPAddrFromAddrPort(grpcAddr),
		LibP2PAddr: net.TCPAddrFromAddrPort(libp2pAddr),
		Telemetry:  new(server.Telemetry),
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
		BlockTime:          rawConfig.BlockTime,
		NodeType:           nodeType.String(),
		LogLevel:           hclog.Error,
		LogFilePath:        rawConfig.LogFilePath,
	}

	return cfg, nil
}

func startNode(cfg *server.Config, availAddr string, nodeType avail.MechanismType) (*server.Server, error) {

	bootnode := false
	if nodeType == avail.BootstrapSequencer {
		bootnode = true
	}

	// Attach the concensus to the Edge
	err := server.RegisterConsensus(server.ConsensusType("avail"), avail.Factory(avail.Config{Bootnode: bootnode, AvailAddr: availAddr}))
	if err != nil {
		return nil, fmt.Errorf("failure to register consensus: %w", err)
	}

	serverInstance, err := server.NewServer(cfg)
	if err != nil {
		return nil, fmt.Errorf("failure to start node: %w", err)
	}

	// Remove consensus from Edge to clean up our factory configuration.
	err = server.UnRegisterConsensus(server.ConsensusType("avail"))
	if err != nil {
		return nil, err
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

func (sc *Context) GethClient() (*ethclient.Client, error) {
	if len(sc.jsonRPCURLs) == 0 {
		return nil, fmt.Errorf("no json-rpc URLs available")
	}

	return ethclient.Dial(sc.jsonRPCURLs[0].String())
}

func (sc *Context) JSONRPCURLs() []*url.URL {
	return sc.jsonRPCURLs
}

func (sc *Context) StopAll() {
	for _, srvInstance := range sc.servers {
		srvInstance.server.Close()
	}
}

// FirstRPCURLForNodeType looks up and returns the url of the node for the node type
func (sc *Context) FirstRPCURLForNodeType(nodeType avail.MechanismType) (*url.URL, error) {
	if len(sc.servers) != len(sc.jsonRPCURLs) {
		return nil, fmt.Errorf("servers and jsonRPCURLs have different lengths")
	}
	for i, srv := range sc.servers {
		if srv.nodeType == nodeType {
			return sc.jsonRPCURLs[i], nil
		}
	}

	return nil, fmt.Errorf("no %s node present in the servers", nodeType)
}
