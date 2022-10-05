package block

import (
	"fmt"

	"github.com/0xPolygon/polygon-edge/blockchain"
	"github.com/0xPolygon/polygon-edge/state"
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/hashicorp/go-hclog"
)

var (
	// XXX: For now hand coded address of the sequencer. Will be removed soon.
	SequencerAddress = "0xF817d12e6933BbA48C14D4c992719B46aD9f5f61"
)

type verifier struct {
	logger hclog.Logger
}

func NewVerifier(logger hclog.Logger) blockchain.Verifier {
	return &verifier{
		logger: logger,
	}
}

func (v *verifier) VerifyHeader(header *types.Header) error {
	signer, err := AddressRecoverFromHeader(header)
	if err != nil {
		return err
	}

	v.logger.Info("Verify header", "signer", signer.String())

	if signer != types.StringToAddress(SequencerAddress) {
		v.logger.Info("Passing, how is it possible? 222")
		return fmt.Errorf("signer address '%s' does not match sequencer address '%s'", signer, SequencerAddress)
	}

	v.logger.Info("Seal signer address successfully verified!", "signer", signer, "sequencer", SequencerAddress)

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
