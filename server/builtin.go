package server

import (
	"github.com/0xPolygon/polygon-edge/chain"
	"github.com/0xPolygon/polygon-edge/secrets"
	"github.com/0xPolygon/polygon-edge/secrets/awsssm"
	"github.com/0xPolygon/polygon-edge/secrets/gcpssm"
	"github.com/0xPolygon/polygon-edge/secrets/hashicorpvault"
	"github.com/0xPolygon/polygon-edge/secrets/local"
	"github.com/0xPolygon/polygon-edge/state"
)

// GenesisFactoryHook is a type definition for a function that takes a chain configuration
// and an engine name, and returns a function that accepts a state transition object and returns an error.
// It is typically used to modify the behavior of the blockchain's genesis block production based on specific engine.
type GenesisFactoryHook func(config *chain.Chain, engineName string) func(*state.Transition) error

// secretsManagerBackends is a map that associates a SecretsManagerType with a SecretsManagerFactory function.
// This allows for the creation of different types of secrets manager depending on the desired backend,
// including local storage, Hashicorp Vault, AWS SSM, and GCP SSM.
var secretsManagerBackends = map[secrets.SecretsManagerType]secrets.SecretsManagerFactory{
	secrets.Local:          local.SecretsManagerFactory,
	secrets.HashicorpVault: hashicorpvault.SecretsManagerFactory,
	secrets.AWSSSM:         awsssm.SecretsManagerFactory,
	secrets.GCPSSM:         gcpssm.SecretsManagerFactory,
}
