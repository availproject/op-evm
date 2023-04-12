package snapshot

import (
	"sync"

	"github.com/0xPolygon/polygon-edge/blockchain/storage"
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

// blockchainKVStorage is the leveldb implementation of the kv storage
type blockchainKVStorage struct {
	db storage.KV

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
		err := bs.db.Set([]byte(k), v)
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

	return bs.db.Set(p, v)
}

func (bs *blockchainKVStorage) Get(p []byte) ([]byte, bool, error) {
	data, found, err := bs.db.Get(p)
	if err != nil {
		if err.Error() == "leveldb: not found" {
			return nil, false, nil
		}

		return nil, false, err
	}

	return data, found, nil
}
