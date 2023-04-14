package snapshot

import (
	"sync"

	"github.com/0xPolygon/polygon-edge/blockchain/storage"
	"github.com/vedhavyas/go-subkey/scale"
)

type BlockchainSnapshot struct {
	Keys   [][]byte
	Values [][]byte
}

func (bs *BlockchainSnapshot) Encode(e scale.Encoder) error {
	return e.Encode(bs)
}

func (bs *BlockchainSnapshot) Decode(d scale.Decoder) error {
	return d.Decode(bs)
}

type BlockchainKVSnapshotter interface {
	storage.KV

	Begin()
	End() *BlockchainSnapshot
	Apply(snapshot *BlockchainSnapshot) error
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
