package snapshot

import (
	"sync"

	itrie "github.com/0xPolygon/polygon-edge/state/immutable-trie"
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/vedhavyas/go-subkey/scale"
)

// codePrefix is the code prefix for leveldb
var codePrefix = []byte("code")

// StateStorageSnapshot represents a snapshot of the state storage.
type StateStorageSnapshot struct {
	Keys   [][]byte
	Values [][]byte
}

// Encode encodes the StateStorageSnapshot using the provided scale.Encoder.
func (ss *StateStorageSnapshot) Encode(e scale.Encoder) error {
	return e.Encode(ss)
}

// Decode decodes the StateStorageSnapshot using the provided scale.Decoder.
func (ss *StateStorageSnapshot) Decode(d scale.Decoder) error {
	return d.Decode(ss)
}

// StateStorageSnapshotter is responsible for managing snapshots of the state storage.
type StateStorageSnapshotter interface {
	itrie.Storage

	Begin()
	End() *StateStorageSnapshot
	Apply(snapshot *StateStorageSnapshot) error
}

// StateStorage is a wrapper around itrie.Storage that provides snapshotting functionality.
type StateStorage struct {
	underlying itrie.Storage

	mutex   *sync.Mutex
	changes map[string][]byte
}

// StateWrapper wraps the given itrie.Storage with snapshotting functionality.
func StateWrapper(s itrie.Storage) StateStorageSnapshotter {
	return &StateStorage{
		underlying: s,
		mutex:      &sync.Mutex{},
	}
}

// Begin starts a new transaction to create a state storage snapshot.
func (ss *StateStorage) Begin() {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()

	ss.changes = make(map[string][]byte)
}

// End finalizes the transaction and returns the created state storage snapshot.
func (ss *StateStorage) End() *StateStorageSnapshot {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()

	ret := &StateStorageSnapshot{
		Keys:   make([][]byte, len(ss.changes)),
		Values: make([][]byte, len(ss.changes)),
	}

	i := 0
	for k, v := range ss.changes {
		ret.Keys[i] = []byte(k)
		ret.Values[i] = v
		i++
	}

	ss.changes = nil

	return ret
}

// Apply applies the changes from the given state storage snapshot to the underlying storage.
func (ss *StateStorage) Apply(snapshot *StateStorageSnapshot) error {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()

	for i := 0; i < len(snapshot.Keys); i++ {
		ss.underlying.Put([]byte(snapshot.Keys[i]), snapshot.Values[i])
	}

	return nil
}

// Put adds a key-value pair to the state storage and the current transaction.
func (ss *StateStorage) Put(k, v []byte) {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()

	buf := make([]byte, len(v))
	copy(buf[:], v[:])

	if ss.changes != nil {
		ss.changes[string(k)] = buf
	}

	ss.underlying.Put(k, buf)
}

// Get retrieves the value associated with the given key from the state storage.
func (ss *StateStorage) Get(k []byte) ([]byte, bool) {
	return ss.underlying.Get(k)
}

// Batch returns a new batch write for the state storage.
func (ss *StateStorage) Batch() itrie.Batch {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()

	return &KVBatch{parent: ss}
}

// SetCode sets the code associated with the given hash in the state storage.
func (ss *StateStorage) SetCode(hash types.Hash, code []byte) {
	ss.Put(append(codePrefix, hash.Bytes()...), code)
}

// GetCode retrieves the code associated with the given hash from the state storage.
func (ss *StateStorage) GetCode(hash types.Hash) ([]byte, bool) {
	return ss.underlying.GetCode(hash)
}

// Close closes the state storage.
func (ss *StateStorage) Close() error {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()

	return ss.underlying.Close()
}

// KVBatch is a batch write for leveldb.
type KVBatch struct {
	parent *StateStorage
}

// Put adds a key-value pair to the batch write.
func (b *KVBatch) Put(k, v []byte) {
	b.parent.Put(k, v)
}

// Write applies the batch write to the state storage.
func (b *KVBatch) Write() {
	// NOP
}
