package avail

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/0xPolygon/polygon-edge/blockchain"
	"github.com/0xPolygon/polygon-edge/consensus"
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	stypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/maticnetwork/avail-settlement/pkg/avail"
	"github.com/umbracle/fastrlp"
)

var (
	// ErrParentBlockNotFound is returned when the local blockchain doesn't
	// contain block for the referenced parent hash.
	ErrParentBlockNotFound = errors.New("parent block not found")

	// FraudproofPrefix is byte sequence that prefixes the fraudproof objected
	// malicious block hash in `ExtraData` of the fraudproof block header.
	FraudproofPrefix = []byte("FRAUDPROOF_OF:")
)

type watchTower struct {
	blockchain   *blockchain.Blockchain
	fraudproofFn func(block types.Block) types.Block
}

func (wt *watchTower) HandleData(bs []byte) error {
	log.Printf("block handler: received batch w/ %d bytes\n", len(bs))

	block := types.Block{}
	if err := block.UnmarshalRLP(bs); err != nil {
		return err
	}

	if err := wt.blockchain.VerifyFinalizedBlock(&block); err != nil {
		log.Printf("block %d (%q) cannot be verified: %s", block.Number(), block.Hash(), err)
		_ = wt.fraudproofFn(block)
		// TODO: Deal with fraud proof
		log.Printf("fraud proof constructed")
		return nil
	}

	if err := wt.blockchain.WriteBlock(&block, "not-sure-yet-what-source-is"); err != nil {
		return fmt.Errorf("failed to write block while bulk syncing: %w", err)
	}

	log.Printf("Received block header: %+v \n", block.Header)
	log.Printf("Received block transactions: %+v \n", block.Transactions)

	return nil
}

func (wt *watchTower) HandleError(err error) {
	log.Printf("block handler: error %#v\n", err)
}

func (d *Avail) runWatchTower(watchTowerAccount accounts.Account, watchTowerPK *keystore.Key) {
	availSender := avail.NewSender(d.availClient, signature.TestKeyringPairAlice)

	handler := &watchTower{
		blockchain: d.blockchain,
		fraudproofFn: func(block types.Block) types.Block {
			b, err := d.constructFraudproof(watchTowerAccount, watchTowerPK, block)
			if err != nil {
				d.logger.Error("error while constructing fraud proof", err)
				return types.Block{}
			}

			f := availSender.SubmitDataAndWaitForStatus(b.MarshalRLP(), stypes.ExtrinsicStatus{IsInBlock: true})
			go func() {
				if _, err := f.Result(); err != nil {
					d.logger.Error("Error while submitting fraud proof to avail", err)
				}
			}()
			return b
		},
	}

	watcher, err := avail.NewBlockDataWatcher(d.availClient, avail.BridgeAppID, handler)
	if err != nil {
		return
	}

	defer watcher.Stop()

	// Consensus always starts in SyncState mode in case it needs
	// to sync with Avail and/or other nodes.
	d.setState(SyncState)

	d.logger.Info("watch tower started")

	var once sync.Once
	for {
		select {
		case <-d.closeCh:
			return
		default: // Default is here because we would block until we receive something in the closeCh
		}

		if d.isState(WatchTowerState) {
			once.Do(func() {
				err := watcher.Start()
				if err != nil {
					panic(err)
				}
			})
		}

		// Start the state machine loop
		d.runWatchTowerCycle()
	}
}

func (d *Avail) runWatchTowerCycle() {
	// Based on the current state, execute the corresponding section
	switch d.getState() {
	case AcceptState:
		d.runAcceptState()

	case ValidateState:
		d.runValidateState()

	case SyncState:
		d.runSyncState()
		d.setState(WatchTowerState)

	case WatchTowerState:
		return
	}
}

func (d *Avail) constructFraudproof(watchTowerAccount accounts.Account, watchTowerPK *keystore.Key, maliciousBlock types.Block) (types.Block, error) {
	header := &types.Header{
		ParentHash: maliciousBlock.ParentHash(),
		Number:     maliciousBlock.Number(),
		Miner:      watchTowerAccount.Address.Bytes(),
		Nonce:      types.Nonce{},
		GasLimit:   maliciousBlock.Header.GasLimit, // TODO(tuommaki): This needs adjusting.
		Timestamp:  uint64(time.Now().Unix()),
	}

	parentHdr, found := d.blockchain.GetParent(maliciousBlock.Header)
	if !found {
		return types.Block{}, ErrParentBlockNotFound
	}

	transition, err := d.executor.BeginTxn(parentHdr.StateRoot, header, types.StringToAddress(watchTowerAccount.Address.Hex()))
	if err != nil {
		return types.Block{}, err
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
	block := consensus.BuildBlock(consensus.BuildBlockParams{
		Header:   header,
		Txns:     txns,
		Receipts: transition.Receipts(),
	})

	// Write the seal of the block after all the fields are completed
	{
		block.Header.ExtraData = make([]byte, SequencerExtraVanity)
		extraData := append([]byte{}, FraudproofPrefix...)
		extraData = append(extraData, []byte(maliciousBlock.Hash().String())...)
		ar := &fastrlp.Arena{}
		rlpExtraData, err := ar.NewBytes(extraData).Bytes()
		if err != nil {
			panic(err)
		}

		copy(header.ExtraData, rlpExtraData)

		ve := &ValidatorExtra{}
		bs := ve.MarshalRLPTo(nil)
		header.ExtraData = append(header.ExtraData, bs...)
	}

	header, err = writeSeal(watchTowerPK.PrivateKey, block.Header)
	if err != nil {
		return types.Block{}, err
	}

	block.Header = header

	// Compute the hash, this is only a provisional hash since the final one
	// is sealed after all the committed seals
	block.Header.ComputeHash()

	return *block, nil
}

// constructFraudproofTxs returns set of transactions that challenge the
// malicious block and submit watchtower's stake.
func constructFraudproofTxs(maliciousBlock types.Block) []*types.Transaction {
	return []*types.Transaction{}
}
