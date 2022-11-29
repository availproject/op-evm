package staking

import (
	"fmt"

	"github.com/0xPolygon/polygon-edge/blockchain"
	"github.com/0xPolygon/polygon-edge/state"
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/hashicorp/go-hclog"
	"github.com/maticnetwork/avail-settlement/pkg/block"
)

type verifier struct {
	activeSequencers ActiveParticipants
	logger           hclog.Logger
}

func NewVerifier(as ActiveParticipants, logger hclog.Logger) blockchain.Verifier {
	return &verifier{
		activeSequencers: as,
		logger:           logger,
	}
}

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

	/*
		parent, ok := i.blockchain.GetHeaderByNumber(header.Number - 1)
		if !ok {
			return fmt.Errorf(
				"unable to get parent header for block number %d",
				header.Number,
			)
		}

		snap, err := i.getSnapshot(parent.Number)
		if err != nil {
			return err
		}

		// verify all the header fields + seal
		if err := i.verifyHeaderImpl(snap, parent, header); err != nil {
			return err
		}

		// verify the committed seals
		if err := verifyCommittedFields(snap, header, i.quorumSize(header.Number)); err != nil {
			return err
		}

		return nil
	*/
	return nil
}

func (v *verifier) ProcessHeaders(headers []*types.Header) error {
	return nil
}

func (v *verifier) GetBlockCreator(header *types.Header) (types.Address, error) {
	return types.BytesToAddress(header.Miner), nil
}

// PreCommitState a hook to be called before finalizing state transition on inserting block
func (v *verifier) PreCommitState(_header *types.Header, _txn *state.Transition) error {
	return nil
}
