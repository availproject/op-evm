package config

import (
	"github.com/0xPolygon/polygon-edge/command/server/config"
	"github.com/0xPolygon/polygon-edge/network"
	"github.com/0xPolygon/polygon-edge/server"
	"github.com/hashicorp/go-hclog"

	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/hcl"
	"gopkg.in/yaml.v3"
)

type CustomServerConfig struct {
	Config   *server.Config
	NodeType string
}

// Config defines the server configuration params
type Config struct {
	GenesisPath              string            `json:"chain_config" yaml:"chain_config"`
	SecretsConfigPath        string            `json:"secrets_config" yaml:"secrets_config"`
	DataDir                  string            `json:"data_dir" yaml:"data_dir"`
	BlockGasTarget           string            `json:"block_gas_target" yaml:"block_gas_target"`
	GRPCAddr                 string            `json:"grpc_addr" yaml:"grpc_addr"`
	JSONRPCAddr              string            `json:"jsonrpc_addr" yaml:"jsonrpc_addr"`
	Telemetry                *config.Telemetry `json:"telemetry" yaml:"telemetry"`
	Network                  *config.Network   `json:"network" yaml:"network"`
	ShouldSeal               bool              `json:"seal" yaml:"seal"`
	TxPool                   *config.TxPool    `json:"tx_pool" yaml:"tx_pool"`
	LogLevel                 string            `json:"log_level" yaml:"log_level"`
	RestoreFile              string            `json:"restore_file" yaml:"restore_file"`
	BlockTime                uint64            `json:"block_time_s" yaml:"block_time_s"`
	Headers                  *config.Headers   `json:"headers" yaml:"headers"`
	LogFilePath              string            `json:"log_to" yaml:"log_to"`
	JSONRPCBatchRequestLimit uint64            `json:"json_rpc_batch_request_limit" yaml:"json_rpc_batch_request_limit"`
	JSONRPCBlockRangeLimit   uint64            `json:"json_rpc_block_range_limit" yaml:"json_rpc_block_range_limit"`
	JSONLogFormat            bool              `json:"json_log_format" yaml:"json_log_format"`

	Relayer               bool   `json:"relayer" yaml:"relayer"`
	NumBlockConfirmations uint64 `json:"num_block_confirmations" yaml:"num_block_confirmations"`
	NodeType              string `json:"node_type" yaml:"node_type"`
}

// DefaultConfig returns the default server configuration
func DefaultConfig() *Config {
	defaultNetworkConfig := network.DefaultConfig()

	return &Config{
		GenesisPath:    "./genesis.json",
		DataDir:        "",
		BlockGasTarget: "0x0", // Special value signaling the parent gas limit should be applied
		Network: &config.Network{
			NoDiscover:       defaultNetworkConfig.NoDiscover,
			MaxPeers:         defaultNetworkConfig.MaxPeers,
			MaxOutboundPeers: defaultNetworkConfig.MaxOutboundPeers,
			MaxInboundPeers:  defaultNetworkConfig.MaxInboundPeers,
			Libp2pAddr: fmt.Sprintf("%s:%d",
				defaultNetworkConfig.Addr.IP,
				defaultNetworkConfig.Addr.Port,
			),
		},
		Telemetry:  &config.Telemetry{},
		ShouldSeal: true,
		TxPool: &config.TxPool{
			PriceLimit:         0,
			MaxSlots:           4096,
			MaxAccountEnqueued: 128,
		},
		LogLevel:    "INFO",
		RestoreFile: "",
		BlockTime:   1686644797, // We are not using it as we produce blocks at our own peace.
		Headers: &config.Headers{
			AccessControlAllowOrigins: []string{"*"},
		},
		LogFilePath:              "",
		JSONRPCBatchRequestLimit: config.DefaultJSONRPCBatchRequestLimit,
		JSONRPCBlockRangeLimit:   config.DefaultJSONRPCBlockRangeLimit,
		Relayer:                  false,
		NumBlockConfirmations:    config.DefaultNumBlockConfirmations,
	}
}

// ReadConfigFile reads the config file from the specified path, builds a Config object
// and returns it.
//
// Supported file types: .json, .hcl, .yaml, .yml
func ReadConfigFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var unmarshalFunc func([]byte, interface{}) error

	switch {
	case strings.HasSuffix(path, ".hcl"):
		unmarshalFunc = hcl.Unmarshal
	case strings.HasSuffix(path, ".json"):
		unmarshalFunc = json.Unmarshal
	case strings.HasSuffix(path, ".yaml"), strings.HasSuffix(path, ".yml"):
		unmarshalFunc = yaml.Unmarshal
	default:
		return nil, fmt.Errorf("suffix of %s is neither hcl, json, yaml nor yml", path)
	}

	config := DefaultConfig()
	config.Network.MaxPeers = -1
	config.Network.MaxInboundPeers = -1
	config.Network.MaxOutboundPeers = -1

	if err := unmarshalFunc(data, config); err != nil {
		return nil, err
	}

	return config, nil
}

// --------------

func NewServerConfig(path string) (*CustomServerConfig, error) {
	rawConfig, err := ReadConfigFile(path)
	if err != nil {
		return nil, err
	}

	chain, err := ParseGenesisConfig(rawConfig)
	if err != nil {
		return nil, err
	}

	jsonRpcAddr, err := ParseJsonRpcAddress(rawConfig)
	if err != nil {
		return nil, err
	}

	grpcAddr, err := ParseGrpcAddress(rawConfig)
	if err != nil {
		return nil, err
	}

	libp2pAddr, err := ParseLibp2pAddress(rawConfig)
	if err != nil {
		return nil, err
	}

	prometheusAddr, err := ParsePrometheusAddress(rawConfig)
	if err != nil {
		return nil, err
	}

	natAddr, err := ParseNatAddress(rawConfig)
	if err != nil {
		return nil, err
	}

	dnsAddr, err := ParseDNSAddress(rawConfig, libp2pAddr.Port)
	if err != nil {
		return nil, err
	}

	secretsConfig, err := ParseSecretsConfig(rawConfig)
	if err != nil {
		return nil, err
	}

	nodeType, err := ParseNodeType(rawConfig) //nolint:typecheck
	if err != nil {
		return nil, err
	}

	serverCfg := &server.Config{
		Chain: chain,
		JSONRPC: &server.JSONRPC{
			JSONRPCAddr:              jsonRpcAddr,
			AccessControlAllowOrigin: rawConfig.Headers.AccessControlAllowOrigins,
		},
		GRPCAddr:   grpcAddr,
		LibP2PAddr: libp2pAddr,
		Telemetry: &server.Telemetry{
			PrometheusAddr: prometheusAddr,
		},
		Network: &network.Config{
			NoDiscover:       rawConfig.Network.NoDiscover,
			Addr:             libp2pAddr,
			NatAddr:          natAddr,
			DNS:              dnsAddr,
			DataDir:          rawConfig.DataDir,
			MaxPeers:         rawConfig.Network.MaxPeers,
			MaxInboundPeers:  rawConfig.Network.MaxInboundPeers,
			MaxOutboundPeers: rawConfig.Network.MaxOutboundPeers,
			Chain:            chain,
		},
		DataDir:            rawConfig.DataDir,
		Seal:               rawConfig.ShouldSeal,
		PriceLimit:         rawConfig.TxPool.PriceLimit,
		MaxSlots:           rawConfig.TxPool.MaxSlots,
		MaxAccountEnqueued: rawConfig.TxPool.MaxAccountEnqueued,
		SecretsManager:     secretsConfig,
		LogLevel:           hclog.LevelFromString(rawConfig.LogLevel),
		LogFilePath:        rawConfig.LogFilePath,
	}

	return &CustomServerConfig{
		Config:   serverCfg,
		NodeType: nodeType.String(),
	}, nil
}
