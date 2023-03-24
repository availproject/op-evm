package snapshot

import (
	"sync"

	"github.com/0xPolygon/polygon-edge/blockchain/storage"
	"github.com/syndtr/goleveldb/leveldb"
)

type BlockchainSnapshot interface {
	Diff() map[string][]byte
}

type blockchainSnapshot map[string][]byte

func (bs blockchainSnapshot) Diff() map[string][]byte {
	return bs
}

type BlockchainKVSnapshotter interface {
	storage.KV

	Begin()
	End() BlockchainSnapshot
	Apply(snapshot BlockchainSnapshot) error
}

// NewLevelDBSnapshotStorage creates the new storage reference with leveldb and snapshot support.
func NewLevelDBSnapshotStorage(path string) (BlockchainKVSnapshotter, error) {
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, err
	}

	kv := &blockchainKVStorage{
		db:    db,
		mutex: &sync.Mutex{},
	}

	return kv, nil
}

// levelDBKV is the leveldb implementation of the kv storage
type blockchainKVStorage struct {
	db *leveldb.DB

	mutex   *sync.Mutex
	changes map[string][]byte
}

func (bs *blockchainKVStorage) Begin() {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()

	bs.changes = make(map[string][]byte)
}

func (bs *blockchainKVStorage) End() BlockchainSnapshot {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()

	ret := blockchainSnapshot(bs.changes)
	bs.changes = nil

	return ret
}

func (bs *blockchainKVStorage) Apply(snapshot BlockchainSnapshot) error {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()

	for k, v := range snapshot.Diff() {
		err := bs.db.Put([]byte(k), v, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

func (bs *blockchainKVStorage) Close() error {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()

	return bs.db.Close()
}

func (bs *blockchainKVStorage) Set(p, v []byte) error {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()

	if bs.changes != nil {
		bs.changes[string(p)] = v
	}

	return bs.db.Put(p, v, nil)
}

func (bs *blockchainKVStorage) Get(p []byte) ([]byte, bool, error) {
	data, err := bs.db.Get(p, nil)
	if err != nil {
		if err.Error() == "leveldb: not found" {
			return nil, false, nil
		}

		return nil, false, err
	}

	return data, true, nil
}

func (bs *blockchainKVStorage) Delete(p []byte) error {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()

	// TODO: Handle Delete in changes

	return bs.db.Delete(p, nil)
}
