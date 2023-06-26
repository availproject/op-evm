package config

import (
	"errors"
	"net"

	"github.com/0xPolygon/polygon-edge/chain"
	"github.com/0xPolygon/polygon-edge/command/helper"
	"github.com/0xPolygon/polygon-edge/network/common"
	"github.com/0xPolygon/polygon-edge/secrets"
	"github.com/maticnetwork/avail-settlement/consensus/avail"
	"github.com/multiformats/go-multiaddr"
)

// ParseGenesisConfig parses the genesis configuration from the provided Config and returns a *chain.Chain instance.
// It imports the chain from the specified genesis path and handles any parsing errors.
func ParseGenesisConfig(cfg *Config) (*chain.Chain, error) {
	if chain, parseErr := chain.Import(cfg.GenesisPath); parseErr != nil {
		return nil, parseErr
	} else {
		return chain, nil
	}
}

// ParsePrometheusAddress parses the Prometheus address from the configuration file.
// If the Prometheus address is not defined or empty, it returns nil.
// Otherwise, it resolves the address using the helper.ResolveAddr function.
func ParsePrometheusAddress(cfg *Config) (*net.TCPAddr, error) {
	if cfg.Telemetry == nil || cfg.Telemetry.PrometheusAddr == "" {
		return nil, nil
	}

	return helper.ResolveAddr(cfg.Telemetry.PrometheusAddr, helper.AllInterfacesBinding)
}

// ParseGrpcAddress parses the gRPC address from the configuration file.
// It resolves the address using the helper.ResolveAddr function.
func ParseGrpcAddress(cfg *Config) (*net.TCPAddr, error) {
	return helper.ResolveAddr(cfg.GRPCAddr, helper.LocalHostBinding)
}

// ParseLibp2pAddress parses the libp2p address from the configuration file.
// It resolves the address using the helper.ResolveAddr function.
func ParseLibp2pAddress(cfg *Config) (*net.TCPAddr, error) {
	return helper.ResolveAddr(cfg.Network.Libp2pAddr, helper.LocalHostBinding)
}

// ParseJsonRpcAddress parses the JSON-RPC address from the configuration file.
// It resolves the address using the helper.ResolveAddr function.
func ParseJsonRpcAddress(cfg *Config) (*net.TCPAddr, error) {
	return helper.ResolveAddr(cfg.JSONRPCAddr, helper.AllInterfacesBinding)
}

// ParseNatAddress parses the NAT address from the configuration file.
// If the NAT address is not defined or empty, it returns nil.
// Otherwise, it parses the address as a net.IP instance.
func ParseNatAddress(cfg *Config) (net.IP, error) {
	if cfg.Network.NatAddr == "" {
		return nil, nil
	}

	addr := net.ParseIP(cfg.Network.NatAddr)
	if addr == nil {
		return nil, errors.New("invalid network NAT address provided")
	}
	return addr, nil
}

// ParseDNSAddress parses the DNS address from the configuration file.
// If the DNS address is not defined or empty, it returns nil.
// Otherwise, it constructs a multiaddr.Multiaddr instance using the common.MultiAddrFromDNS function.
func ParseDNSAddress(cfg *Config, p2pPort int) (multiaddr.Multiaddr, error) {
	if cfg.Network.DNSAddr == "" {
		return nil, nil
	}

	return common.MultiAddrFromDNS(cfg.Network.DNSAddr, p2pPort)
}

// ParseSecretsConfig parses the secrets configuration from the provided Config and returns a *secrets.SecretsManagerConfig instance.
// If the secrets configuration path is not defined or empty, it returns nil.
// Otherwise, it reads the configuration using the secrets.ReadConfig function.
func ParseSecretsConfig(cfg *Config) (*secrets.SecretsManagerConfig, error) {
	if cfg.SecretsConfigPath == "" {
		return nil, nil
	}

	return secrets.ReadConfig(cfg.SecretsConfigPath)
}

// ParseNodeType parses the node type from the configuration file and returns an avail.MechanismType value.
// If the node type is not defined or empty, it returns avail.Validator.
// Otherwise, it parses the node type using the avail.ParseType function.
func ParseNodeType(cfg *Config) (avail.MechanismType, error) {
	if cfg.NodeType == "" {
		return avail.Sequencer, nil
	}

	return avail.ParseType(cfg.NodeType)
}
