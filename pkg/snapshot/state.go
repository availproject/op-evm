package snapshot

import (
	"sync"

	itrie "github.com/0xPolygon/polygon-edge/state/immutable-trie"
	"github.com/0xPolygon/polygon-edge/types"
)

// codePrefix is the code prefix for leveldb
var codePrefix = []byte("code")

type StateStorageSnapshot interface {
	Diff() map[string][]byte
}

type stateStorageSnapshot map[string][]byte

func (sss stateStorageSnapshot) Diff() map[string][]byte {
	return sss
}

type StateStorageSnapshotter interface {
	itrie.Storage

	Begin()
	End() StateStorageSnapshot
	Apply(snapshot StateStorageSnapshot) error
}

type StateStorage struct {
	underlying itrie.Storage

	mutex   *sync.Mutex
	changes map[string][]byte
}

func StateWrapper(s itrie.Storage) StateStorageSnapshotter {
	return &StateStorage{
		underlying: s,

		mutex: &sync.Mutex{},
	}
}

func (ss *StateStorage) Begin() {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()

	ss.changes = make(map[string][]byte)
}

func (ss *StateStorage) End() StateStorageSnapshot {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()

	ret := stateStorageSnapshot(ss.changes)
	ss.changes = nil

	return ret
}

func (ss *StateStorage) Apply(snapshot StateStorageSnapshot) error {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()

	for k, v := range snapshot.Diff() {
		ss.underlying.Put([]byte(k), v)
	}

	return nil
}

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

func (ss *StateStorage) Get(k []byte) ([]byte, bool) {
	return ss.underlying.Get(k)
}

func (ss *StateStorage) Batch() itrie.Batch {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()

	return &KVBatch{parent: ss}
}

func (ss *StateStorage) SetCode(hash types.Hash, code []byte) {
	ss.Put(append(codePrefix, hash.Bytes()...), code)
}

func (ss *StateStorage) GetCode(hash types.Hash) ([]byte, bool) {
	return ss.underlying.GetCode(hash)
}

func (ss *StateStorage) Close() error {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()

	return ss.underlying.Close()
}

// KVBatch is a batch write for leveldb
type KVBatch struct {
	parent *StateStorage
}

func (b *KVBatch) Put(k, v []byte) {
	b.parent.Put(k, v)
}

func (b *KVBatch) Write() {
	// NOP
}
