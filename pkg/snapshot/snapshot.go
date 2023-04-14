package snapshot

import (
	"github.com/0xPolygon/polygon-edge/blockchain/storage"
	itrie "github.com/0xPolygon/polygon-edge/state/immutable-trie"
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/hashicorp/go-hclog"
	"github.com/vedhavyas/go-subkey/scale"
)

type Snapshot struct {
	BlockNumber uint64
	BlockHash   types.Hash
	StateRoot   types.Hash

	BlockchainSnapshot *BlockchainSnapshot
	StateSnapshot      *StateStorageSnapshot
}

func (s *Snapshot) Encode(e scale.Encoder) error {
	return e.Encode(s)
}

func (s *Snapshot) Decode(d scale.Decoder) error {
	return d.Decode(s)
}

type Snapshotter interface {
	Begin()
	End() *Snapshot

	Apply(*Snapshot) error
}

type snapshotter struct {
	blockchainSnapshotter BlockchainKVSnapshotter
	stateSnapshotter      StateStorageSnapshotter
}

func NewSnapshotter(logger hclog.Logger, stateStorage itrie.Storage, blockchainStoragePath string) (Snapshotter, storage.Storage, itrie.Storage, error) {
	var blockchainSnapshotter BlockchainKVSnapshotter
	var err error

	if blockchainStoragePath == "" {
		blockchainSnapshotter, err = NewMemoryDBSnapshotStorage()
	} else {
		blockchainSnapshotter, err = NewLevelDBSnapshotStorage(blockchainStoragePath)
	}
	if err != nil {
		return nil, nil, nil, err
	}

	stateSnapshotter := StateWrapper(stateStorage)

	s := &snapshotter{
		blockchainSnapshotter: blockchainSnapshotter,
		stateSnapshotter:      stateSnapshotter,
	}

	bchainStorage := storage.NewKeyValueStorage(logger.Named("leveldb"), blockchainSnapshotter)

	return s, bchainStorage, stateSnapshotter, nil
}

func (s *snapshotter) Begin() {
	s.stateSnapshotter.Begin()
	s.blockchainSnapshotter.Begin()
}

func (s *snapshotter) End() *Snapshot {
	snapshot := &Snapshot{
		BlockchainSnapshot: s.blockchainSnapshotter.End(),
		StateSnapshot:      s.stateSnapshotter.End(),
	}

	return snapshot
}

func (s *snapshotter) Apply(snapshot *Snapshot) error {
	err := s.blockchainSnapshotter.Apply(snapshot.BlockchainSnapshot)
	if err != nil {
		return err
	}

	return s.stateSnapshotter.Apply(snapshot.StateSnapshot)
}
