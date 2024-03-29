package staking

import (
	"fmt"

	"github.com/0xPolygon/polygon-edge/state"
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/availproject/op-evm/pkg/block"
	"github.com/availproject/op-evm/pkg/blockchain"
	"github.com/hashicorp/go-hclog"
)

// verifier is a struct that implements the blockchain.Verifier interface.
type verifier struct {
	activeSequencers ActiveParticipants
	logger           hclog.Logger
}

// NewVerifier returns a new instance of the blockchain.Verifier.
// It takes ActiveParticipants and a logger as parameters.
func NewVerifier(as ActiveParticipants, logger hclog.Logger) blockchain.Verifier {
	return &verifier{
		activeSequencers: as,
		logger:           logger,
	}
}

// VerifyHeader verifies the given header by checking if the signer address belongs to the active sequencers.
// It takes the header as a parameter and returns an error if verification fails.
func (v *verifier) VerifyHeader(header *types.Header) error {
	signer, err := block.AddressRecoverFromHeader(header)
	if err != nil {
		return err
	}

	v.logger.Info("Verify header", "signer", signer.String())

	activeSequencers, err := v.activeSequencers.Get(Sequencer)
	if err != nil {
		return err
	}

	// XXX: Is this ok? Verification of the very first signature is chicken-egg
	// problem, because initially there are no sequencers staked and the first
	// block needs to be passed through, in order to "register" the staking.
	//
	// This check can also function as an escape hatch to prevent blockchain
	// halting in case all sequencers unstake for some reason.
	if len(activeSequencers) == 0 {
		v.logger.Warn("no active sequencers staked atm. - skipping signer verification")
		return nil
	}

	minerIsActiveSequencer := false
	for _, s := range activeSequencers {
		if signer == s {
			minerIsActiveSequencer = true
			break
		}
	}

	if !minerIsActiveSequencer {
		v.logger.Error("failed to verify signer address", "address", signer)
		return fmt.Errorf("signer address '%s' does not belong to active sequencers", signer)
	}

	v.logger.Info("Seal signer address successfully verified!", "signer", signer)

	return nil
}

// ProcessHeaders processes the given headers.
// It takes headers as a parameter and returns an error if processing fails.
func (v *verifier) ProcessHeaders(headers []*types.Header) error {
	return nil
}

// GetBlockCreator returns the address of the block creator based on the header.
// It takes the header as a parameter and returns the block creator address and an error if retrieval fails.
func (v *verifier) GetBlockCreator(header *types.Header) (types.Address, error) {
	return types.BytesToAddress(header.Miner), nil
}

// PreCommitState is a hook called before finalizing the state transition on inserting a block.
// It takes the header and transition as parameters and returns an error if pre-committing fails.
func (v *verifier) PreCommitState(_header *types.Header, _txn *state.Transition) error {
	return nil
}
