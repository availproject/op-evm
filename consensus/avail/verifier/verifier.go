package verifier

import (
	"fmt"

	"github.com/0xPolygon/polygon-edge/blockchain"
	"github.com/0xPolygon/polygon-edge/state"
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/hashicorp/go-hclog"
	"github.com/maticnetwork/avail-settlement/consensus/avail/common"
	"github.com/maticnetwork/avail-settlement/pkg/block"
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
	signer, err := block.AddressRecoverFromHeader(header)
	if err != nil {
		return err
	}

	v.logger.Info("Verify header", "signer", signer.String())

	if signer != types.StringToAddress(common.SequencerAddress) {
		v.logger.Info("Passing, how is it possible? 222")
		return fmt.Errorf("signer address '%s' does not match sequencer address '%s'", signer, common.SequencerAddress)
	}

	v.logger.Info("Seal signer address successfully verified!", "signer", signer, "sequencer", common.SequencerAddress)

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
