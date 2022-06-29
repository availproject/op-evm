package config

import (
	"github.com/0xPolygon/polygon-edge/command/server/config"
	"github.com/0xPolygon/polygon-edge/network"
	"github.com/0xPolygon/polygon-edge/server"
	"github.com/hashicorp/go-hclog"
)

func NewServerConfig(path string) (*server.Config, error) {
	rawConfig, err := config.ReadConfigFile(path)
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

	return &server.Config{
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
		DataDir:        rawConfig.DataDir,
		Seal:           rawConfig.ShouldSeal,
		PriceLimit:     rawConfig.TxPool.PriceLimit,
		MaxSlots:       rawConfig.TxPool.MaxSlots,
		SecretsManager: secretsConfig,
		RestoreFile: func(cfg *config.Config) *string {
			if cfg.RestoreFile != "" {
				return &cfg.RestoreFile
			}

			return nil
		}(rawConfig),
		BlockTime:       rawConfig.BlockTime,
		IBFTBaseTimeout: rawConfig.IBFTBaseTimeout,
		LogLevel:        hclog.LevelFromString(rawConfig.LogLevel),
		LogFilePath:     rawConfig.LogFilePath,
	}, nil
}
