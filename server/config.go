package server

import (
	"net"

	"github.com/hashicorp/go-hclog"

	"github.com/0xPolygon/polygon-edge/chain"
	"github.com/0xPolygon/polygon-edge/network"
	"github.com/0xPolygon/polygon-edge/secrets"
)

const (
	// DefaultGRPCPort is the default port for the GRPC server
	DefaultGRPCPort int = 9632

	// DefaultJSONRPCPort is the default port for the JSON-RPC server
	DefaultJSONRPCPort int = 8545
)

// Config is the main configuration struct for the server. It holds
// various sub-configurations related to different components of the server.
type Config struct {
	// Chain configuration
	Chain *chain.Chain

	// JSON-RPC server configuration
	JSONRPC *JSONRPC

	// TCP address for the GRPC server
	GRPCAddr *net.TCPAddr

	// TCP address for the LibP2P server
	LibP2PAddr *net.TCPAddr

	// Various blockchain-related parameters
	PriceLimit         uint64
	MaxAccountEnqueued uint64
	MaxSlots           uint64
	BlockTime          uint64

	// Telemetry configuration for metric services
	Telemetry *Telemetry

	// Network configuration
	Network *network.Config

	// Data directory for the server
	DataDir string

	// Restore file path, if any
	RestoreFile *string

	// Flag to enable or disable sealing
	Seal bool

	// Secrets manager configuration
	SecretsManager *secrets.SecretsManagerConfig

	// Log level for the server
	LogLevel hclog.Level

	// Flag to enable or disable JSON log format
	JSONLogFormat bool

	// Path for the log file
	LogFilePath string

	// Flag to enable or disable relayer mode
	Relayer bool

	// Number of block confirmations required
	NumBlockConfirmations uint64
}

// Telemetry holds the configuration details for metric services.
type Telemetry struct {
	// TCP address for the Prometheus server
	PrometheusAddr *net.TCPAddr
}

// JSONRPC holds the configuration details for the JSON-RPC server.
type JSONRPC struct {
	// TCP address for the JSON-RPC server
	JSONRPCAddr *net.TCPAddr

	// Allowed origins for CORS
	AccessControlAllowOrigin []string

	// Limit for the number of items in a batch request
	BatchLengthLimit uint64

	// Limit for the number of blocks in a range request
	BlockRangeLimit uint64
}
