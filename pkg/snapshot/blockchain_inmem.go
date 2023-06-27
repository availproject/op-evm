package snapshot

import (
	"sync"

	"github.com/0xPolygon/polygon-edge/helper/hex"
)

// memoryKV is an in-memory implementation of the kv storage.
type memoryKV struct {
	db map[string][]byte
}

// Set sets the value of the given key in the in-memory key-value storage.
func (m *memoryKV) Set(p []byte, v []byte) error {
	m.db[hex.EncodeToHex(p)] = v
	return nil
}

// Get retrieves the value associated with the given key from the in-memory key-value storage.
func (m *memoryKV) Get(p []byte) ([]byte, bool, error) {
	v, ok := m.db[hex.EncodeToHex(p)]
	if !ok {
		return nil, false, nil
	}
	return v, true, nil
}

// Close closes the in-memory key-value storage.
func (m *memoryKV) Close() error {
	return nil
}

// NewMemoryDBSnapshotStorage creates a new storage reference with an in-memory key-value store and snapshot support.
func NewMemoryDBSnapshotStorage() (BlockchainKVSnapshotter, error) {
	db := &memoryKV{db: map[string][]byte{}}

	kv := &blockchainKVStorage{
		db:    db,
		mutex: &sync.Mutex{},
	}

	return kv, nil
}
