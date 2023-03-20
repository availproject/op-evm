package server

import (
	"fmt"

	"github.com/0xPolygon/polygon-edge/consensus"
	consensusDummy "github.com/0xPolygon/polygon-edge/consensus/dummy"
	"github.com/0xPolygon/polygon-edge/secrets"
	"github.com/0xPolygon/polygon-edge/secrets/awsssm"
	"github.com/0xPolygon/polygon-edge/secrets/gcpssm"
	"github.com/0xPolygon/polygon-edge/secrets/hashicorpvault"
	"github.com/0xPolygon/polygon-edge/secrets/local"
)

type ConsensusType string

const (
	DummyConsensus ConsensusType = "dummy"
)

var consensusBackends = map[ConsensusType]consensus.Factory{
	DummyConsensus: consensusDummy.Factory,
}

// secretsManagerBackends defines the SecretManager factories for different
// secret management solutions
var secretsManagerBackends = map[secrets.SecretsManagerType]secrets.SecretsManagerFactory{
	secrets.Local:          local.SecretsManagerFactory,
	secrets.HashicorpVault: hashicorpvault.SecretsManagerFactory,
	secrets.AWSSSM:         awsssm.SecretsManagerFactory,
	secrets.GCPSSM:         gcpssm.SecretsManagerFactory,
}

func ConsensusSupported(value string) bool {
	_, ok := consensusBackends[ConsensusType(value)]

	return ok
}

func RegisterConsensus(ct ConsensusType, f consensus.Factory) error {
	if ConsensusSupported(string(ct)) {
		return fmt.Errorf("provided consensus '%s' is already registered", ct)
	}
	consensusBackends[ct] = f
	return nil
}

func UnRegisterConsensus(ct ConsensusType) error {
	if !ConsensusSupported(string(ct)) {
		return fmt.Errorf("provided consensus '%s' is not registered", ct)
	}

	delete(consensusBackends, ct)
	return nil
}
