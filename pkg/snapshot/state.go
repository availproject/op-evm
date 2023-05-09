package snapshot

import (
	"sync"

	itrie "github.com/0xPolygon/polygon-edge/state/immutable-trie"
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/vedhavyas/go-subkey/scale"
)

// codePrefix is the code prefix for leveldb
var codePrefix = []byte("code")

type StateStorageSnapshot struct {
	Keys   [][]byte
	Values [][]byte
}

func (ss *StateStorageSnapshot) Encode(e scale.Encoder) error {
	return e.Encode(ss)
}

func (ss *StateStorageSnapshot) Decode(d scale.Decoder) error {
	return d.Decode(ss)
}

type StateStorageSnapshotter interface {
	itrie.Storage

	Begin()
	End() *StateStorageSnapshot
	Apply(snapshot *StateStorageSnapshot) error
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

func (ss *StateStorage) Apply(snapshot *StateStorageSnapshot) error {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()

	for i := 0; i < len(snapshot.Keys); i++ {
		ss.underlying.Put([]byte(snapshot.Keys[i]), snapshot.Values[i])
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
