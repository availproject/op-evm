// Package snapshot provides functionality for managing snapshots of a blockchain key-value storage.
// It includes an interface for creating and applying snapshots, as well as an implementation using leveldb.
package snapshot

import (
	"sync"

	"github.com/0xPolygon/polygon-edge/blockchain/storage"
	"github.com/vedhavyas/go-subkey/scale"
)

// BlockchainSnapshot represents a snapshot of a blockchain key-value storage.
type BlockchainSnapshot struct {
	Keys   [][]byte
	Values [][]byte
}

// Encode encodes the BlockchainSnapshot using the provided scale.Encoder.
func (bs *BlockchainSnapshot) Encode(e scale.Encoder) error {
	return e.Encode(bs)
}

// Decode decodes the BlockchainSnapshot using the provided scale.Decoder.
func (bs *BlockchainSnapshot) Decode(d scale.Decoder) error {
	return d.Decode(bs)
}

// BlockchainKVSnapshotter is an interface that extends storage.KV and provides additional snapshot-related functionality.
type BlockchainKVSnapshotter interface {
	storage.KV

	// Begin starts a new transaction to create a snapshot.
	Begin()

	// End finalizes the transaction and returns the created snapshot.
	End() *BlockchainSnapshot

	// Apply applies the changes from the given snapshot to the underlying key-value storage.
	Apply(snapshot *BlockchainSnapshot) error
}

// blockchainKVStorage is the leveldb implementation of the kv storage.
type blockchainKVStorage struct {
	db storage.KV

	mutex   *sync.Mutex
	changes map[string][]byte
}

// Begin starts a new transaction to create a snapshot.
func (bs *blockchainKVStorage) Begin() {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()

	bs.changes = make(map[string][]byte)
}

// End finalizes the transaction and returns the created snapshot.
func (bs *blockchainKVStorage) End() *BlockchainSnapshot {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()

	ret := &BlockchainSnapshot{
		Keys:   make([][]byte, len(bs.changes)),
		Values: make([][]byte, len(bs.changes)),
	}

	i := 0
	for k, v := range bs.changes {
		ret.Keys[i] = []byte(k)
		ret.Values[i] = v
		i++
	}

	bs.changes = nil

	return ret
}

// Apply applies the changes from the given snapshot to the underlying key-value storage.
func (bs *blockchainKVStorage) Apply(snapshot *BlockchainSnapshot) error {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()

	for i := 0; i < len(snapshot.Keys); i++ {
		err := bs.db.Set([]byte(snapshot.Keys[i]), snapshot.Values[i])
		if err != nil {
			return err
		}
	}

	return nil
}

// Close closes the underlying key-value storage.
func (bs *blockchainKVStorage) Close() error {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()

	return bs.db.Close()
}

// Set sets the value of the given key in the underlying key-value storage.
func (bs *blockchainKVStorage) Set(p, v []byte) error {
	bs.mutex.Lock()
	defer bs.mutex.Unlock()

	if bs.changes != nil {
		bs.changes[string(p)] = v
	}

	return bs.db.Set(p, v)
}

// Get retrieves the value associated with the given key from the underlying key-value storage.
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
