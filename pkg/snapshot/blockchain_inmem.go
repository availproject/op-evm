package snapshot

import (
	"sync"

	"github.com/0xPolygon/polygon-edge/helper/hex"
)

// memoryKV is an in memory implementation of the kv storage
type memoryKV struct {
	db map[string][]byte
}

func (m *memoryKV) Set(p []byte, v []byte) error {
	m.db[hex.EncodeToHex(p)] = v

	return nil
}

func (m *memoryKV) Get(p []byte) ([]byte, bool, error) {
	v, ok := m.db[hex.EncodeToHex(p)]
	if !ok {
		return nil, false, nil
	}

	return v, true, nil
}

func (m *memoryKV) Close() error {
	return nil
}

// NewMemoryDBSnapshotStorage creates the new storage reference with inmem kv store and snapshot support.
func NewMemoryDBSnapshotStorage() (BlockchainKVSnapshotter, error) {
	db := &memoryKV{map[string][]byte{}}

	kv := &blockchainKVStorage{
		db:    db,
		mutex: &sync.Mutex{},
	}

	return kv, nil
}
