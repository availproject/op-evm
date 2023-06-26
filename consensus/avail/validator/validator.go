package validator

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/0xPolygon/polygon-edge/types"
	"github.com/0xPolygon/polygon-edge/types/buildroot"
	"github.com/availproject/op-evm/pkg/block"
	"github.com/availproject/op-evm/pkg/blockchain"
	"github.com/availproject/op-evm/pkg/staking"
	"github.com/hashicorp/go-hclog"
)

/****************************************************************************/

var (
	// ErrInvalidBlock is a general error used when the block structure is invalid or its field values are inconsistent.
	ErrInvalidBlock = errors.New("invalid block")

	// ErrInvalidBlockSequence is returned when the block sequence is invalid.
	ErrInvalidBlockSequence = errors.New("invalid block sequence")

	// ErrInvalidParentHash is returned when the parent block hash is invalid.
	ErrInvalidParentHash = errors.New("parent block hash is invalid")

	// ErrInvalidSha3Uncles is returned when the block's sha3 uncles root is invalid.
	ErrInvalidSha3Uncles = errors.New("invalid block sha3 uncles root")

	// ErrInvalidTxRoot is returned when the block's transactions root is invalid.
	ErrInvalidTxRoot = errors.New("invalid block transactions root")

	// ErrNoBlock is returned when no block data is passed in.
	ErrNoBlock = errors.New("no block data passed in")

	// ErrParentHashMismatch is returned when the parent block hash doesn't match.
	ErrParentHashMismatch = errors.New("invalid parent block hash")

	// ErrParentNotFound is returned when the parent block is not found.
	ErrParentNotFound = errors.New("parent block not found")
)

// Validator is an interface that defines methods for applying, checking, and processing fraudproof blocks.
type Validator interface {
	Apply(block *types.Block) error
	Check(block *types.Block) error
	ProcessFraudproof(block *types.Block) error
}

// ValidatorSet represents a set of validators.
type ValidatorSet []types.Address

// validator implements the Validator interface and provides the actual implementation for the methods.
type validator struct {
	blockchain *blockchain.Blockchain

	logger           hclog.Logger
	sequencerAddress types.Address
}

// New creates a new instance of Validator with the provided parameters.
func New(blockchain *blockchain.Blockchain, sequencer types.Address, logger hclog.Logger) Validator {
	return &validator{
		blockchain: blockchain,

		logger:           logger.Named("validator"),
		sequencerAddress: sequencer,
	}
}

// Apply applies a block to the blockchain by writing it to the blockchain.
func (v *validator) Apply(blk *types.Block) error {
	if err := v.blockchain.WriteBlock(blk, block.SourceAvail); err != nil {
		return fmt.Errorf("failed to write block while bulk syncing: %w", err)
	}

	v.logger.Debug("Received block header", "header", blk.Header)
	v.logger.Debug("Received block transactions", "transactions", blk.Transactions)

	return nil
}

// Check checks the validity of a block by verifying its header and performing block verification.
// It returns an error if the block is invalid.
func (v *validator) Check(blk *types.Block) error {
	if blk.Header == nil {
		return fmt.Errorf("%w: block.Header == nil", ErrInvalidBlock)
	}

	if err := v.verifyFinalizedBlock(blk); err != nil {
		return fmt.Errorf("unable to verify block, %w", err)
	}
	return nil
}

// ProcessFraudproof processes a fraudproof block by extracting the fraudproof information from its header.
// It performs the necessary actions based on the fraudproof information.
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

// verifyFinalizedBlock verifies a finalized block by performing header verification and block verification.
// It returns an error if the block is invalid.
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

// verifyBlock performs the base block verification steps by verifying the block's parent information and block body.
// It returns an error if the block is invalid.
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

// verifyHeader verifies the header of a block by checking the sequencer address and the miner address.
// It returns an error if the header is invalid.
func (v *validator) verifyHeader(header *types.Header) error {
	signer, err := block.AddressRecoverFromHeader(header)
	if err != nil {
		return err
	}

	v.logger.Info("About to process block header verification",
		"block_hash", header.Hash,
		"signer", signer.String(),
	)

	minerAddr := types.BytesToAddress(header.Miner)

	if !bytes.Equal(signer.Bytes(), header.Miner) {
		return fmt.Errorf(
			"signer address '%s' does not match sequencer address '%s' for block hash '%s'",
			signer, minerAddr, header.Hash,
		)
	}

	v.logger.Info(
		"Seal signer address successfully verified!",
		"block_hash", header.Hash,
		"signer", signer,
		"sequencer", minerAddr,
	)

	return nil
}

// verifyBlockParent verifies that the child block is in line with the locally saved parent block.
// It checks the existence of the parent block, the matching of hashes, the matching of block numbers,
// and the matching of gas limit/gas used.
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
		// Check if one of the transactions is `BeginDisputeResolutionTx`, which can
		// perform a fork in case the corresponding sequencer made fraud.
		for _, tx := range childBlk.Transactions {
			isDisputeResolutionFork, err := staking.IsBeginDisputeResolutionTx(tx)
			if err != nil {
				return err
			}

			if !isDisputeResolutionFork {
				v.logger.Error(
					"block number sequence not correct",
					"child_block_number", childBlk.Number(),
					"parent_block_number", parent.Number,
				)
				return ErrInvalidBlockSequence
			}
		}
	}

	// Make sure the gas limit is within correct bounds
	if gasLimitErr := v.verifyGasLimit(childBlk.Header, parent); gasLimitErr != nil {
		return fmt.Errorf("invalid gas limit, %w", gasLimitErr)
	}

	return nil
}

// verifyGasLimit verifies the gas limit of a block header.
// It returns an error if the gas limit is invalid.
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

	// blockchain.BlockGasTargetDivisor is now private (v0.8.1 -> v1.0.0) so instead using it's value 1024
	limit := parentHeader.GasLimit / 1024
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

// verifyBlockBody verifies that the block body is valid by checking the uncle root and transactions root.
// It returns an error if the block body is invalid.
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
