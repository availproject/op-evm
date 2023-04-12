package snapshot

import (
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
)

// NewLevelDBSnapshotStorage creates the new storage reference with leveldb and snapshot support.
func NewLevelDBSnapshotStorage(path string) (BlockchainKVSnapshotter, error) {
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, err
	}

	kv := &blockchainKVStorage{
		db:    &levelDBKV{db},
		mutex: &sync.Mutex{},
	}

	return kv, nil
}

// levelDBKV is the leveldb implementation of the kv storage
type levelDBKV struct {
	db *leveldb.DB
}

// Set sets the key-value pair in leveldb storage
func (l *levelDBKV) Set(p []byte, v []byte) error {
	return l.db.Put(p, v, nil)
}

// Get retrieves the key-value pair in leveldb storage
func (l *levelDBKV) Get(p []byte) ([]byte, bool, error) {
	data, err := l.db.Get(p, nil)
	if err != nil {
		if err.Error() == "leveldb: not found" {
			return nil, false, nil
		}

		return nil, false, err
	}

	return data, true, nil
}

// Close closes the leveldb storage instance
func (l *levelDBKV) Close() error {
	return l.db.Close()
}
