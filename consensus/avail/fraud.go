package avail

import (
	"bytes"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/0xPolygon/polygon-edge/crypto"
	"github.com/0xPolygon/polygon-edge/state"
	"github.com/0xPolygon/polygon-edge/txpool"
	"github.com/0xPolygon/polygon-edge/types"
	stypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/hashicorp/go-hclog"
	"github.com/maticnetwork/avail-settlement/consensus/avail/watchtower"
	"github.com/maticnetwork/avail-settlement/pkg/avail"
	"github.com/maticnetwork/avail-settlement/pkg/block"
	"github.com/maticnetwork/avail-settlement/pkg/blockchain"
	"github.com/maticnetwork/avail-settlement/pkg/staking"
)

var (
	ErrTxPoolHashNotFound          = errors.New("hash not found in the txpool")
	ChainProcessingDisabled uint32 = 0
	ChainProcessingEnabled  uint32 = 1
)

// Fraud is a structure that represents the state of a node in a blockchain system that is capable of detecting and handling fraudulent activities.
// It contains various state data and services required to perform its function.
type Fraud struct {
	logger                 hclog.Logger           // logger provides a logging interface for the fraud detection system.
	blockchain             *blockchain.Blockchain // blockchain is a reference to the blockchain being monitored.
	executor               *state.Executor        // executor is a reference to the state executor.
	txpool                 *txpool.TxPool         // txpool refers to the transaction pool where incoming transactions are stored.
	watchtower             watchtower.WatchTower  // watchtower is a reference to the watchtower consensus algorithm.
	blockProductionEnabled *atomic.Bool           // blockProductionEnabled is an atomic boolean representing whether the block production is enabled.

	nodeAddr    types.Address     // nodeAddr represents the address of the node.
	nodeSignKey *ecdsa.PrivateKey // nodeSignKey is the node's private key for signing transactions.
	availSender avail.Sender      // availSender represents a sender in the Avail network.
	nodeType    MechanismType     // nodeType specifies the type of the node.

	fraudBlock          *types.Block       // fraudBlock is the block suspected of fraud.
	lastFraudDisputedTx *types.Transaction // lastFraudDisputedTx is the last transaction that was disputed for fraud.
	chainProcessStatus  uint32             // chainProcessStatus represents the status of the chain processing.
}

// SetBlock sets the block suspected of fraud.
// It is called when a block is detected as fraudulent and needs to be processed for dispute resolution.
func (f *Fraud) SetBlock(b *types.Block) {
	f.fraudBlock = b
}

// GetBlock returns the block that is currently suspected of fraud.
// This block is the one that is currently under dispute resolution.
func (f *Fraud) GetBlock() *types.Block {
	return f.fraudBlock
}

// SetChainStatus is used to update the status of the chain processing.
// This status can be updated based on whether the chain is operating normally or is under a fraud dispute.
func (f *Fraud) SetChainStatus(status uint32) {
	atomic.StoreUint32(&f.chainProcessStatus, status)

	if status == ChainProcessingEnabled {
		f.blockProductionEnabled.Store(true)
	} else {
		f.blockProductionEnabled.Store(false)
	}
}

// IsChainDisabled checks if the chain processing has been disabled.
// This could be due to an ongoing fraud dispute or other reasons.
func (f *Fraud) IsChainDisabled() bool {
	return f.chainProcessStatus == ChainProcessingDisabled
}

// IsReadyToSlash checks if the node is ready to slash a fraudulent block.
// Slashing is the process of penalizing a node that has been proven to perform fraudulent actions.
func (f *Fraud) IsReadyToSlash() bool {
	if f.chainProcessStatus == ChainProcessingDisabled && f.fraudBlock != nil {
		return true
	}

	return false
}

// CheckAndSetFraudBlock checks a list of blocks and sets a block suspected of fraud if it finds one.
// This is done by analyzing the extra data attached to a block.
func (f *Fraud) CheckAndSetFraudBlock(blocks []*types.Block) bool {
	for _, blk := range blocks {
		if fraudProofBlockHash, exists := block.GetExtraDataFraudProofTarget(blk.Header); exists {
			f.logger.Info(
				"Fraud proof parent hash block discovered. Continuing with fraud dispute resolution...",
				"probation_block_hash", fraudProofBlockHash,
				"watchtower_fraud_block_hash", blk.Hash(),
			)
			f.SetBlock(blk)
			return true
		}
	}
	return false
}

// IsDisputeResolutionEnded checks if the dispute resolution for a suspected fraudulent block has ended.
// This is done by comparing the current block hash with the hash attached as extra data in the block.
func (f *Fraud) IsDisputeResolutionEnded(blk *types.Header) bool {
	if f.fraudBlock == nil {
		return false
	}

	blkDisputeEndHash, _ := block.GetExtraDataEndDisputeResolutionTarget(blk)
	return bytes.Equal(f.fraudBlock.Hash().Bytes(), blkDisputeEndHash.Bytes())
}

// EndDisputeResolution ends the dispute resolution process for a suspected fraudulent block.
// This is done by resetting the fraud block and updating the chain status to enabled.
func (f *Fraud) EndDisputeResolution() {
	f.SetBlock(nil)
	f.SetChainStatus(ChainProcessingEnabled)
}

// ShouldStopProducingBlocks contains the main logic of the fraud detection system.
// It monitors the transaction pool and checks for any transactions indicating fraudulent activities.
// If it detects a fraud, it will update the chain status to disabled and stop producing new blocks.
func (f *Fraud) ShouldStopProducingBlocks(activeParticipantsQuerier staking.ActiveParticipants) {
	for {
		// We've already received begin dispute resolution transaction. Now it's time to wait for
		// processing prior we check tx pool again...
		if f.IsChainDisabled() {
			time.Sleep(200 * time.Millisecond) // Just a bit of the delay...
			continue
		}

		f.txpool.Prepare(f.txpool.GetBaseFee())

	innerLoop:
		for {

			tx := f.txpool.Peek()
			if tx == nil {
				break innerLoop
			}

			isWatchtower, err := activeParticipantsQuerier.Contains(tx.From, staking.WatchTower)
			if err != nil {
				f.logger.Debug("failure while checking if tx from is active watchtower", "error", err)
				continue
			}

			isBeginDisputeResolutionTx, err := staking.IsBeginDisputeResolutionTx(tx)
			if err != nil {
				f.logger.Debug("failure while checking if tx is type of begin dispute resolution", "error", err)
				continue
			}

			f.logger.Debug(
				"New tx pool transaction discovered",
				"hash", tx.Hash,
				"value", tx.Value,
				"originating_addr", tx.From.String(),
				"recipient_addr", tx.To.String(),
				"submitted_via_watchtower", isWatchtower,
				"staking_contract_addr", staking.AddrStakingContract.String(),
				"submitted_towards_contract", bytes.Equal(tx.To.Bytes(), staking.AddrStakingContract.Bytes()),
				"tx_type_of_begin_dispute_resolution", isBeginDisputeResolutionTx,
			)

			if isWatchtower && bytes.Equal(tx.To.Bytes(), staking.AddrStakingContract.Bytes()) && isBeginDisputeResolutionTx {
				// It happens that in time to time, due to multiple push (one tx pool one next block) of the begin dispute resolution txs
				// it can get node into the disputed mode even if dispute mode is already resolved for that specific transaction.
				// This check makes sure we bypass that situation.
				if f.lastFraudDisputedTx != nil && bytes.Equal(tx.Hash.Bytes(), f.lastFraudDisputedTx.Hash.Bytes()) {
					continue
				}

				f.logger.Warn(
					"Discovered valid begin dispute resolution transaction. Chain is entering fraud dispute mode...",
					"originating_watchtower_addr", tx.From,
					"dispute_resolution_tx_hash", tx.Hash,
				)

				// We have proper transaction and therefore we are going to stop processing blocks in the chain
				f.SetChainStatus(ChainProcessingDisabled)
				f.lastFraudDisputedTx = tx
				break innerLoop
			}
		}

		// Just a bit of the time to not break the CPU...
		time.Sleep(100 * time.Millisecond)
	}
}

// DiscoverDisputeResolutionTx searches for a transaction matching the provided hash within the transaction pool.
// The function iterates through the transaction pool, and upon finding a match, pops the transaction from the pool,
// logs its discovery, and returns it with a nil error.
// If no matching transaction is found, the function returns a nil transaction and an error stating the transaction hash was not found.
func (f *Fraud) DiscoverDisputeResolutionTx(hash types.Hash) (*types.Transaction, error) {
	f.txpool.Prepare(f.txpool.GetBaseFee())

	for {
		tx := f.txpool.Peek()
		if tx == nil {
			break
		}

		if bytes.Equal(tx.Hash.Bytes(), hash.Bytes()) {
			f.logger.Info(
				"Discovered txpool dispute resolution transaction",
				"hash", tx.Hash,
				"nonce", tx.Nonce,
				"account_from", tx.From,
			)

			// no errors, pop the tx from the pool
			f.txpool.Pop(tx)
			return tx, nil
		}
	}

	return nil, ErrTxPoolHashNotFound
}

// GetBeginDisputeResolutionTxHash retrieves the hash of the transaction that initiated the dispute resolution process.
// This is done by extracting the dispute resolution target from the extra data in the fraud block's header.
func (f *Fraud) GetBeginDisputeResolutionTxHash() types.Hash {
	hash, _ := block.GetExtraDataBeginDisputeResolutionTarget(f.fraudBlock.Header)
	return hash
}

// IsFraudProofBlock checks if the given block has evidence of fraudulent activity.
// The function checks the block's extra data for a fraud proof target. If found, the function returns true. If not, it returns false.
func (f *Fraud) IsFraudProofBlock(blk *types.Block) bool {
	_, exists := block.GetExtraDataFraudProofTarget(blk.Header)
	return exists
}

// CheckAndSlash conducts a fraud investigation. It checks if the system is ready to slash a fraudulent block, and if so, it retrieves the hash
// of the suspected fraudulent block and checks the existence of a block with this hash in the blockchain.
// If the suspected block does not exist or if the block was produced by the same node running this function, it logs the issue and returns.
// If the suspected block exists and was produced by a different node, it checks the block for fraud using the implemented watchtower mechanism.
// If the check detects fraud, the function initiates the process of slashing the node that created the block.
// If the check does not detect fraud, the function slashes the watchtower node instead, as it incorrectly flagged the block as fraudulent.
// The function returns true if a node was slashed and false if not, along with an error if any occurred.
func (f *Fraud) CheckAndSlash() (bool, error) {
	// There is no block attached from previous sequencer runs and therefore we assume
	// no fraud should be checked in this moment...
	if !f.IsReadyToSlash() {
		f.logger.Debug("not yet ready to process the block and slash the participant...")
		return false, nil
	}

	fraudBlockTargetHash, exists := block.GetExtraDataFraudProofTarget(f.fraudBlock.Header)
	if !exists {
		// Disregard entirely this specific fraud block
		f.EndDisputeResolution()

		// It seems that fraud block is set but the proof target cannot be calculated
		// therefore we are going to log this problem and panic as this should *NEVER EVER HAPPEN*
		// Block should not be set if it's not fraud block via `CheckAndSetFraudBlock` in the first place.
		panic(fmt.Sprintf(
			"failed to extract fraud proof targed from the fraud block hash `%s`",
			f.fraudBlock.Hash(),
		))
	}

	f.logger.Info(
		"Discovered fraud proof block hash targed",
		"targeted_block_hash", fraudBlockTargetHash,
		"watchtower_block_hash", f.fraudBlock.Hash(),
	)

	maliciousBlock, mbExists := f.blockchain.GetBlockByHash(fraudBlockTargetHash, false)
	if !mbExists {
		f.logger.Info(
			"Potentially malicious block not discovered, rejecting future verification",
			"watchtower_block_hash", f.fraudBlock.Hash(),
			"potentially_malicious_block_hash", fraudBlockTargetHash,
		)

		return false, fmt.Errorf(
			"failed to discover potentially malicious block hash: %s, watchtower_block_hash: %s",
			f.fraudBlock.Header.Hash, fraudBlockTargetHash,
		)
	}

	f.logger.Info(
		"Potentially malicious block discovered, processing with the check...",
		"watchtower_block_hash", f.fraudBlock.Hash(),
		"potentially_malicious_block_hash", maliciousBlock.Hash(),
	)

	sequencerAddr := types.BytesToAddress(maliciousBlock.Header.Miner)
	watchtowerAddr := types.BytesToAddress(f.fraudBlock.Header.Miner)

	// Slashing should not occur from the node that produced actual malicious block
	if sequencerAddr.String() == f.nodeAddr.String() {
		f.logger.Warn(
			"Potentially malicious node cannot process (slash) block it produced",
			"malicious_addr", sequencerAddr,
			"node_addr", f.nodeAddr,
			"watchtower_block_hash", f.fraudBlock.Hash(),
			"potentially_malicious_block_hash", maliciousBlock.Hash(),
		)

		return false, errors.New(
			"potentially malicious node cannot process with slashing itself",
		)
	}

	// Discover who needs to be slashed.
	// If watchtower produced block that proves sequencer to be corrupted, sequencer needs to be slashed.
	// If watchtower produced block that proves sequencer to be correct, watchtower needs to be slashed.
	if err := f.watchtower.Check(maliciousBlock); err != nil {
		f.logger.Warn(
			"Fraud proof block check confirmed malicious block. Slashing sequencer...",
			"watchtower_block_hash", f.fraudBlock.Hash(),
			"potentially_malicious_block_hash", maliciousBlock.Hash(),
			"potentially_malicious_block_parent_hash", maliciousBlock.ParentHash(),
			"potentially_malicious_block_number", maliciousBlock.Number(),
			"sequencer", sequencerAddr,
			"watchtower_addr", watchtowerAddr,
			"error", err,
		)

		if err := f.slashNode(sequencerAddr, maliciousBlock.Header, Sequencer); err != nil {
			f.logger.Error(
				"failed to slash node (sequencer)",
				"watchtower_block_hash", f.fraudBlock.Hash(),
				"potentially_malicious_block_hash", maliciousBlock.Hash(),
				"sequencer", sequencerAddr,
				"watchtower_addr", watchtowerAddr,
				"error", err,
			)
			return false, err
		}
		return true, nil

	} else {
		f.logger.Warn(
			"Fraud proof block check confirmed block is not malicious. Slashing watchtower...",
			"watchtower_block_hash", f.fraudBlock.Hash(),
			"potentially_malicious_block_hash", maliciousBlock.Hash(),
			"sequencer", sequencerAddr,
			"watchtower_addr", watchtowerAddr,
			"error", err,
		)

		if err := f.slashNode(watchtowerAddr, maliciousBlock.Header, WatchTower); err != nil {
			f.logger.Error(
				"failed to slash node (watchtower)",
				"watchtower_block_hash", f.fraudBlock.Hash(),
				"potentially_malicious_block_hash", maliciousBlock.Hash(),
				"sequencer", sequencerAddr,
				"watchtower_addr", watchtowerAddr,
				"error", err,
			)
			return false, err
		}
		return true, nil
	}
}

// slashNode initiates the process of slashing a node that produced a fraudulent block.
// It first generates a "begin dispute resolution" block, followed by a "slash" block.
// The process involves penalizing the malicious node by reducing its stake and rights in the blockchain network.
// After the slashing process is completed, the fraud detection system ends the dispute resolution process, as the fraudulent action has been addressed.
// The function returns an error if any occurred during the process.
func (f *Fraud) slashNode(maliciousAddr types.Address, maliciousHeader *types.Header, nodeType MechanismType) error {
	blockBuilderFactory := block.NewBlockBuilderFactory(f.blockchain, f.executor, f.logger)

	disputeBlk, err := f.produceBeginDisputeResolutionBlock(blockBuilderFactory, maliciousAddr, maliciousHeader, nodeType)
	if err != nil {
		return err
	}

	// TODO: Fix this function and remove this warning.
	{
		_, file, ln, ok := runtime.Caller(0)
		if ok {
			f.logger.Warn(fmt.Sprintf("%s:%d: TODO: Make {begin, end} dispute resolution process transactions atomic.", file, ln))
		}
	}

	_, err = f.produceSlashBlock(blockBuilderFactory, disputeBlk, maliciousAddr, maliciousHeader, nodeType)
	if err != nil {
		return err
	}

	// No longer is it required for the chain to be in the disputed mode
	f.EndDisputeResolution()
	return nil
}

// produceBeginDisputeResolutionBlock initiates the creation of a dispute resolution block to flag a potential fraudulent activity by a node.
// Depending on the node type, it will either create a new block by forking the chain (in case of a sequencer node) or just create a block from the current head of the blockchain (in case of a watchtower node).
// The function then sets the block number, coinbase address, and signs the block.
// It fetches the transaction hash for beginning the dispute resolution from the fraudulent block and discovers the associated transaction, which is then added to the dispute resolution block.
// The block is built and sent to the Avail network. On successful submission, the block is written to the blockchain.
// The function also resets the transaction pool with the current block header to remove stale transactions.
// It logs the successful creation and addition of the dispute resolution block to the blockchain, then returns the block and a nil error.
// If at any point an error occurs, the function logs the error and returns a nil block along with the error.
func (f *Fraud) produceBeginDisputeResolutionBlock(blockBuilderFactory block.BlockBuilderFactory, maliciousAddr types.Address, maliciousHeader *types.Header, nodeType MechanismType) (*types.Block, error) {
	var bb block.Builder
	var err error

	// We are going to fork the chain but only if the malicious participant is sequencer.
	// Otherwise we are making sure we slash the watchtower and continue normal operation...
	switch nodeType {
	case Sequencer:
		bb, err = blockBuilderFactory.FromParentHash(maliciousHeader.ParentHash)
	case WatchTower:
		bb, err = blockBuilderFactory.FromBlockchainHead()
	default:
		panic("unsupported node type: " + nodeType)
	}

	if err != nil {
		return nil, err
	}

	// Force sequential block number, to ensure it's correct in case of fork as well.
	bb.SetBlockNumber(f.blockchain.Header().Number + 1)

	bb.SetCoinbaseAddress(f.nodeAddr)
	bb.SignWith(f.nodeSignKey)

	// Append begin disputed resolution txn
	disputeTxHash := f.GetBeginDisputeResolutionTxHash()
	f.logger.Info("Dispute resolution tx hash from fraud block", "hash", disputeTxHash.String())
	disputeBeginTx, err := f.DiscoverDisputeResolutionTx(disputeTxHash)
	if err != nil {
		f.logger.Error(
			"failed to discover begin dispute resoultion transaction for the block",
			"correct_block_hash", maliciousHeader.ParentHash,
			"error", err,
		)
		return nil, err
	}
	bb.AddTransactions(disputeBeginTx)

	blk, err := bb.Build()
	if err != nil {
		f.logger.Error("failed to build begin dispute resolution block", "error", err)
		return nil, err
	}

	f.logger.Info(
		"Sending begin dispute resolution block to the Avail",
		"hash", blk.Hash(),
		"malicious_block_hash", maliciousHeader.Hash,
		"parent_block_hash", maliciousHeader.ParentHash,
	)

	err = f.availSender.SendAndWaitForStatus(blk, stypes.ExtrinsicStatus{IsInBlock: true})
	if err != nil {
		f.logger.Error("error while submitting begin dispute resolution block to avail", "error", err)
		return nil, err
	}

	err = f.blockchain.WriteBlock(blk, f.nodeType.String())
	if err != nil {
		f.logger.Error("failed to write begin dispute resolution block to the blockchain", "error", err)
		return nil, err
	}

	// After the block has been written we reset the txpool to remove stale transactions.
	f.txpool.ResetWithHeaders(blk.Header)

	f.logger.Info(
		"Successfully sent and wrote begin dispute resolution block to the blockchain...",
		"txn_count", len(blk.Transactions),
		"hash", blk.Hash(),
		"block_number", blk.Number(),
		"malicious_block_hash", maliciousHeader.Hash,
		"parent_block_hash", maliciousHeader.ParentHash,
	)

	return blk, nil
}

// produceSlashBlock initiates the creation of a slashing block to penalize a malicious node.
// It begins by creating a new block based on the head of the blockchain. The function then sets the coinbase address for the block and signs the block.
// Next, a transaction is prepared to slash the staker of the malicious node, removing an amount from their stake.
// The function also prepares the state transition context and increments the nonce of the slashing transaction to prevent transaction replay.
// The slashing transaction is then signed and added to the slashing block.
// After successfully building and writing the slashing block to the blockchain, the transaction pool is reset with the current block header to remove stale transactions.
// The function logs the successful creation and addition of the slashing block to the blockchain, then returns the block and a nil error.
// If at any point an error occurs, the function logs the error and returns a nil block along with the error.
func (f *Fraud) produceSlashBlock(blockBuilderFactory block.BlockBuilderFactory, disputeBlk *types.Block, maliciousAddr types.Address, maliciousHeader *types.Header, nodeType MechanismType) (*types.Block, error) {
	slashBlk, err := blockBuilderFactory.FromBlockchainHead()
	if err != nil {
		return nil, err
	}

	slashBlk.SetCoinbaseAddress(f.nodeAddr)
	slashBlk.SignWith(f.nodeSignKey)

	disputeResolutionTx, err := staking.SlashStakerTx(f.nodeAddr, maliciousAddr, 1_000_000)
	if err != nil {
		f.logger.Error("failed to end new fraud dispute resolution", "error", err)
		return nil, err
	}

	hdr, exists := f.blockchain.GetHeaderByHash(disputeBlk.Hash())
	if !exists {
		return nil, fmt.Errorf("cannot find block with disputed header %q", disputeBlk.Hash().String())
	}

	transition, err := f.executor.BeginTxn(hdr.StateRoot, hdr, f.nodeAddr)
	if err != nil {
		f.logger.Error("failed to begin the transition for the end dispute resolution", "error", err)
		return nil, err
	}
	disputeResolutionTx.Nonce = transition.GetNonce(disputeResolutionTx.From)

	txSigner := &crypto.FrontierSigner{}
	dtx, err := txSigner.SignTx(disputeResolutionTx, f.nodeSignKey)
	if err != nil {
		f.logger.Error("failed to sign slashing transaction", "error", err)
		return nil, err
	}

	slashBlk.AddTransactions(dtx)

	// Used to ensure we can end fraud dispute for a specific fraud block on all of the nodes!
	slashBlk.SetExtraDataField(block.KeyEndDisputeResolutionOf, f.fraudBlock.Hash().Bytes())

	blk, err := slashBlk.Build()
	if err != nil {
		f.logger.Error("failed to build slashing block", "error", err)
		return nil, err
	}

	f.logger.Info(
		"Sending slashing block to the Avail",
		"hash", blk.Hash(),
		"malicious_block_hash", maliciousHeader.Hash,
		"parent_block_hash", maliciousHeader.ParentHash,
	)

	err = f.availSender.SendAndWaitForStatus(blk, stypes.ExtrinsicStatus{IsInBlock: true})
	if err != nil {
		f.logger.Error("error while submitting slashing block to avail", "error", err)
		return nil, err
	}

	err = f.blockchain.WriteBlock(blk, f.nodeType.String())
	if err != nil {
		f.logger.Error("failed to write slashing block to the blockchain", "error", err)
		return nil, err
	}

	// After the block has been written we reset the txpool to remove stale transactions.
	f.txpool.ResetWithHeaders(blk.Header)

	f.logger.Info(
		"Successfully sent and wrote slashing block to the blockchain... Resuming chain activity...",
		"txn_count", len(blk.Transactions),
		"hash", blk.Hash(),
		"block_number", blk.Number(),
		"malicious_block_hash", maliciousHeader.Hash,
		"parent_block_hash", maliciousHeader.ParentHash,
	)

	return blk, nil
}

// NewFraudResolver creates a new FraudResolver instance which is used to detect and handle fraudulent activity within the blockchain network.
// The FraudResolver uses several components such as a logger, a blockchain, an executor, a transaction pool, and a watchtower to perform its functions.
// It also requires several settings such as the node address, node signing key, a sender for Avail network communication, and the node type (sequencer or watchtower).
// The created FraudResolver also includes information on the status of chain processing and block production.
func NewFraudResolver(logger hclog.Logger, b *blockchain.Blockchain, e *state.Executor, txp *txpool.TxPool, w watchtower.WatchTower, blockProductionEnabled *atomic.Bool, nodeAddr types.Address, nodeSignKey *ecdsa.PrivateKey, availSender avail.Sender, nodeType MechanismType) *Fraud {
	return &Fraud{
		logger:                 logger,
		blockchain:             b,
		executor:               e,
		txpool:                 txp,
		watchtower:             w,
		nodeAddr:               nodeAddr,
		nodeType:               nodeType,
		nodeSignKey:            nodeSignKey,
		availSender:            availSender,
		chainProcessStatus:     ChainProcessingEnabled,
		blockProductionEnabled: blockProductionEnabled,
	}
}
