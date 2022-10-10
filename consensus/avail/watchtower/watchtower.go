package watchtower

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"time"

	"github.com/0xPolygon/polygon-edge/blockchain"
	"github.com/0xPolygon/polygon-edge/consensus"
	"github.com/0xPolygon/polygon-edge/state"
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/hashicorp/go-hclog"
	"github.com/maticnetwork/avail-settlement/pkg/block"
	block_seal "github.com/maticnetwork/avail-settlement/pkg/block"
	"github.com/umbracle/fastrlp"
)

var (
	// ErrInvalidBlock is general error used when block structure is invalid
	// or its field values are inconsistent.
	ErrInvalidBlock = errors.New("invalid block")

	// ErrParentBlockNotFound is returned when the local blockchain doesn't
	// contain block for the referenced parent hash.
	ErrParentBlockNotFound = errors.New("parent block not found")

	// FraudproofPrefix is byte sequence that prefixes the fraudproof objected
	// malicious block hash in `ExtraData` of the fraudproof block header.
	FraudproofPrefix = []byte("FRAUDPROOF_OF:")
)

type WatchTower interface {
	Apply(blk *types.Block) error
	Check(blk *types.Block) error
	ConstructFraudproof(blk *types.Block) (*types.Block, error)
}

type watchTower struct {
	blockchain *blockchain.Blockchain
	executor   *state.Executor
	logger     hclog.Logger

	account types.Address
	signKey *ecdsa.PrivateKey
}

func New(blockchain *blockchain.Blockchain, executor *state.Executor, account types.Address, signKey *ecdsa.PrivateKey) WatchTower {
	return &watchTower{
		blockchain: blockchain,
		executor:   executor,
		logger:     hclog.Default(),

		account: account,
		signKey: signKey,
	}
}

func (wt *watchTower) Check(blk *types.Block) error {
	if blk == nil {
		return fmt.Errorf("%w: block == nil", ErrInvalidBlock)
	}

	if blk.Header == nil {
		return fmt.Errorf("%w: block.Header == nil", ErrInvalidBlock)
	}

	if err := wt.blockchain.VerifyFinalizedBlock(blk); err != nil {
		wt.logger.Info("block %d (%q) cannot be verified: %s", blk.Number(), blk.Hash(), err)
		return err
	}

	return nil
}

func (wt *watchTower) Apply(blk *types.Block) error {
	if err := wt.blockchain.WriteBlock(blk, block.SourceWatchTower); err != nil {
		return fmt.Errorf("failed to write block while bulk syncing: %w", err)
	}

	wt.logger.Debug("Received block header: %+v \n", blk.Header)
	wt.logger.Debug("Received block transactions: %+v \n", blk.Transactions)

	return nil
}

func (wt *watchTower) ConstructFraudproof(maliciousBlock *types.Block) (*types.Block, error) {
	header := &types.Header{
		ParentHash: maliciousBlock.ParentHash(),
		Number:     maliciousBlock.Number(),
		Miner:      wt.account.Bytes(),
		Nonce:      types.Nonce{},
		GasLimit:   maliciousBlock.Header.GasLimit, // TODO(tuommaki): This needs adjusting.
		Timestamp:  uint64(time.Now().Unix()),
	}

	parentHdr, found := wt.blockchain.GetParent(maliciousBlock.Header)
	if !found {
		return nil, ErrParentBlockNotFound
	}

	transition, err := wt.executor.BeginTxn(parentHdr.StateRoot, header, wt.account)
	if err != nil {
		return nil, err
	}

	txns := constructFraudproofTxs(maliciousBlock)
	for _, tx := range txns {
		err := transition.Write(tx)
		if err != nil {
			// TODO(tuommaki): This needs re-assesment. Theoretically there
			// should NEVER be situation where fraud proof transaction writing
			// could fail and hence panic here is appropriate. There is some
			// debugging aspects though, which might need revisiting,
			// especially if the malicious block can cause situation where this
			// section fails - it would then defeat the purpose of watch tower.
			panic(err)
		}
	}

	// Commit the changes
	_, root := transition.Commit()

	// Update the header
	header.StateRoot = root
	header.GasUsed = transition.TotalGas()

	// Build the actual block
	// The header hash is computed inside buildBlock
	blk := consensus.BuildBlock(consensus.BuildBlockParams{
		Header:   header,
		Txns:     txns,
		Receipts: transition.Receipts(),
	})

	// Write the seal of the block after all the fields are completed
	{
		blk.Header.ExtraData = make([]byte, block_seal.SequencerExtraVanity)
		extraData := append([]byte{}, FraudproofPrefix...)
		extraData = append(extraData, []byte(maliciousBlock.Hash().String())...)
		ar := &fastrlp.Arena{}
		rlpExtraData, err := ar.NewBytes(extraData).Bytes()
		if err != nil {
			panic(err)
		}

		copy(header.ExtraData, rlpExtraData)

		ve := &block_seal.ValidatorExtra{}
		bs := ve.MarshalRLPTo(nil)
		header.ExtraData = append(header.ExtraData, bs...)
	}

	header, err = block_seal.WriteSeal(wt.signKey, blk.Header)
	if err != nil {
		return nil, err
	}

	blk.Header = header

	// Compute the hash, this is only a provisional hash since the final one
	// is sealed after all the committed seals
	blk.Header.ComputeHash()

	return blk, nil
}

// constructFraudproofTxs returns set of transactions that challenge the
// malicious block and submit watchtower's stake.
func constructFraudproofTxs(maliciousBlock *types.Block) []*types.Transaction {
	return []*types.Transaction{}
}
