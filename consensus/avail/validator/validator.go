package validator

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/0xPolygon/polygon-edge/blockchain"
	"github.com/0xPolygon/polygon-edge/state"
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/0xPolygon/polygon-edge/types/buildroot"
	"github.com/hashicorp/go-hclog"
	"github.com/maticnetwork/avail-settlement/pkg/block"
)

/****************************************************************************/

// For now hand coded address of the sequencer
const SequencerAddress = "0xF817d12e6933BbA48C14D4c992719B46aD9f5f61"

var (
	// ErrInvalidBlock is general error used when block structure is invalid
	// or its field values are inconsistent.
	ErrInvalidBlock         = errors.New("invalid block")
	ErrInvalidBlockSequence = errors.New("invalid block sequence")
	ErrInvalidParentHash    = errors.New("parent block hash is invalid")
	ErrInvalidSha3Uncles    = errors.New("invalid block sha3 uncles root")
	ErrInvalidTxRoot        = errors.New("invalid block transactions root")
	ErrNoBlock              = errors.New("no block data passed in")
	ErrParentHashMismatch   = errors.New("invalid parent block hash")
	ErrParentNotFound       = errors.New("parent block not found")
)

type Validator interface {
	Apply(block *types.Block) error
	Check(block *types.Block) error
	ProcessFraudproof(block *types.Block) error
}

type ValidatorSet []types.Address

type validator struct {
	blockchain *blockchain.Blockchain
	executor   *state.Executor

	logger           hclog.Logger
	sequencerAddress types.Address
}

func New(blockchain *blockchain.Blockchain, executor *state.Executor, sequencer types.Address, logger hclog.Logger) Validator {
	return &validator{
		blockchain: blockchain,
		executor:   executor,

		logger:           logger.Named("validator"),
		sequencerAddress: sequencer,
	}
}

func (v *validator) Apply(blk *types.Block) error {
	if err := v.blockchain.WriteBlock(blk, block.SourceAvail); err != nil {
		return fmt.Errorf("failed to write block while bulk syncing: %w", err)
	}

	v.logger.Debug("Received block header", "header", blk.Header)
	v.logger.Debug("Received block transactions", "transactions", blk.Transactions)

	return nil
}

func (v *validator) Check(blk *types.Block) error {
	if blk.Header == nil {
		return fmt.Errorf("%w: block.Header == nil", ErrInvalidBlock)
	}

	if err := v.verifyFinalizedBlock(blk); err != nil {
		return fmt.Errorf("unable to verify block, %w", err)
	}
	return nil
}

func (v *validator) ProcessFraudproof(blk *types.Block) error {
	extraDataKV, err := block.DecodeExtraDataFields(blk.Header.ExtraData)
	if err != nil {
		return err
	}

	hashBS, exists := extraDataKV[block.KeyFraudProofOf]
	if exists {
		v.logger.Warn("**************** FRAUD PROOF FOUND ************************")

		blkHash := types.BytesToHash(hashBS)
		v.logger.Info("Fraudproof for block", "hash", blkHash)

		// TODO(tuommaki): Process fraud proof.
	}

	return nil
}

// verifyFinalizedBlock is modified version of `blockchain.Blockchain.VerifyFinalizedBlock()`
func (v *validator) verifyFinalizedBlock(blk *types.Block) error {
	// Make sure the consensus layer verifies this block header
	if err := v.verifyHeader(blk.Header); err != nil {
		return fmt.Errorf("failed to verify the header: %w", err)
	}

	// Do the initial block verification
	if err := v.verifyBlock(blk); err != nil {
		return err
	}

	return nil

}

// verifyBlock does the base (common) block verification steps by
// verifying the block body as well as the parent information
func (v *validator) verifyBlock(blk *types.Block) error {
	// Make sure the block is present
	if blk == nil {
		return ErrNoBlock
	}

	// Make sure the block is in line with the parent block
	if err := v.verifyBlockParent(blk); err != nil {
		return err
	}

	// Make sure the block body data is valid
	if err := v.verifyBlockBody(blk); err != nil {
		return err
	}

	return nil
}

// TODO - Check if miner address was the same (active sequencer at that point in the time)
// through the avail block
func (v *validator) verifyHeader(header *types.Header) error {
	signer, err := block.AddressRecoverFromHeader(header)
	if err != nil {
		return err
	}

	v.logger.Info("Verify header", "signer", signer.String())

	minerAddr := types.BytesToAddress(header.Miner)

	if !bytes.Equal(signer.Bytes(), header.Miner) {
		return fmt.Errorf("signer address '%s' does not match sequencer address '%s'", signer, minerAddr)
	}

	v.logger.Info("Seal signer address successfully verified!", "signer", signer, "sequencer", minerAddr)

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

// verifyBlockParent makes sure that the child block is in line
// with the locally saved parent block. This means checking:
// - The parent exists
// - The hashes match up
// - The block numbers match up
// - The block gas limit / used matches up
func (v *validator) verifyBlockParent(childBlk *types.Block) error {
	// Grab the parent block
	parentHash := childBlk.ParentHash()
	parent, ok := v.blockchain.GetHeaderByHash(parentHash)

	if !ok {
		v.logger.Error(
			"parent not found",
			"child_block_hash", childBlk.Hash().String(),
			"child_block_number", childBlk.Number(),
			"parent_block_hash", parentHash,
		)

		return ErrParentNotFound
	}

	// Make sure the hash is valid
	if parent.Hash == types.ZeroHash {
		return ErrInvalidParentHash
	}

	// Make sure the hashes match up
	if parentHash != parent.Hash {
		return ErrParentHashMismatch
	}

	// Make sure the block numbers are correct
	if childBlk.Number()-1 != parent.Number {
		v.logger.Error(
			"block number sequence not correct",
			"child_block_number", childBlk.Number(),
			"parent_block_number", parent.Number,
		)

		return ErrInvalidBlockSequence
	}

	// Make sure the gas limit is within correct bounds
	if gasLimitErr := v.verifyGasLimit(childBlk.Header, parent); gasLimitErr != nil {
		return fmt.Errorf("invalid gas limit, %w", gasLimitErr)
	}

	return nil
}

// verifyGasLimit is a helper function for validating a gas limit in a header
func (v *validator) verifyGasLimit(header *types.Header, parentHeader *types.Header) error {
	if header.GasUsed > header.GasLimit {
		return fmt.Errorf(
			"block gas used exceeds gas limit, limit = %d, used=%d",
			header.GasLimit,
			header.GasUsed,
		)
	}

	// Skip block limit difference check for genesis
	if header.Number == 0 {
		return nil
	}

	// Find the absolute delta between the limits
	diff := int64(parentHeader.GasLimit) - int64(header.GasLimit)
	if diff < 0 {
		diff *= -1
	}

	limit := parentHeader.GasLimit / blockchain.BlockGasTargetDivisor
	if uint64(diff) > limit {
		return fmt.Errorf(
			"invalid gas limit, limit = %d, want %d +- %d",
			header.GasLimit,
			parentHeader.GasLimit,
			limit-1,
		)
	}

	return nil
}

// verifyBlockBody verifies that the block body is valid. This means checking:
// - The trie roots match up (state, transactions, receipts, uncles)
// - The receipts match up
// - The execution result matches up
func (v *validator) verifyBlockBody(blk *types.Block) error {
	// Make sure the Uncles root matches up
	if hash := buildroot.CalculateUncleRoot(blk.Uncles); hash != blk.Header.Sha3Uncles {
		v.logger.Error(fmt.Sprintf(
			"uncle root hash mismatch: have %s, want %s",
			hash,
			blk.Header.Sha3Uncles,
		))

		return ErrInvalidSha3Uncles
	}

	// Make sure the transactions root matches up
	if hash := buildroot.CalculateTransactionsRoot(blk.Transactions); hash != blk.Header.TxRoot {
		v.logger.Error(fmt.Sprintf(
			"transaction root hash mismatch: have %s, want %s",
			hash,
			blk.Header.TxRoot,
		))

		return ErrInvalidTxRoot
	}

	// Transaction execution skipped, contrary to
	// `blockchain.Blockchain.verifyBlockBody()` implementation.

	return nil
}
