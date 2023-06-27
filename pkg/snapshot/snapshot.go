// Package snapshot provides functionality for managing snapshots of blockchain state and storage in the context of decentralized applications (DApps).
// It offers utilities for creating, encoding, decoding, and applying snapshots of the blockchain's key-value storage and state storage.
//
// The package includes interfaces, implementations, and helper functions for creating and manipulating snapshots.
// It supports various storage backends, including in-memory storage, LevelDB storage, and other key-value storage implementations.
// Snapshots can be used to capture and restore the blockchain state at specific points in time, enabling efficient state management, data integrity, and data recovery.
//
// # Key Features
//
// - Snapshotting of blockchain key-value storage and state storage.
//
// - Encoding and decoding of snapshots using the Substrate scale format.
//
// - Support for in-memory storage and LevelDB storage.
//
// - Snapshot creation, application, and retrieval.
//
// # Usage
//
// To use the snapshot package, import it as follows:
//
//	import "github.com/example/snapshot"
//
// This package provides several types and functions for working with snapshots, including the Snapshotter interface, which defines methods for managing snapshots.
// The package also includes implementations of the Snapshotter interface, such as NewMemoryDBSnapshotStorage and NewLevelDBSnapshotStorage, which create instances of snapshot storage backends.
//
// Example:
//
//	logger := hclog.Default() // Create a logger instance
//	stateStorage := ...       // Create an instance of state storage
//	blockchainStoragePath := ... // Path to LevelDB storage
//
//	snapshotter, blockchainStorage, _, err := snapshot.NewSnapshotter(logger, stateStorage, blockchainStoragePath)
//	if err != nil {
//	    // Handle error
//	}
//
//	// Use the snapshotter and storage backend
//
// Refer to the package's documentation and individual function/method documentation for more details on each component and its usage.
package snapshot

import (
	"github.com/0xPolygon/polygon-edge/blockchain/storage"
	itrie "github.com/0xPolygon/polygon-edge/state/immutable-trie"
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/hashicorp/go-hclog"
	"github.com/vedhavyas/go-subkey/scale"
)

// Snapshot represents a snapshot of the blockchain state.
type Snapshot struct {
	BlockNumber        uint64                // BlockNumber is the number of the block at which the snapshot was taken.
	BlockHash          types.Hash            // BlockHash is the hash of the block at which the snapshot was taken.
	StateRoot          types.Hash            // StateRoot is the root hash of the state at which the snapshot was taken.
	BlockchainSnapshot *BlockchainSnapshot   // BlockchainSnapshot represents a snapshot of the blockchain key-value storage.
	StateSnapshot      *StateStorageSnapshot // StateSnapshot represents a snapshot of the state storage.
}

// Encode encodes the Snapshot using the provided scale.Encoder.
func (s *Snapshot) Encode(e scale.Encoder) error {
	return e.Encode(s)
}

// Decode decodes the Snapshot using the provided scale.Decoder.
func (s *Snapshot) Decode(d scale.Decoder) error {
	return d.Decode(s)
}

// Snapshotter is responsible for creating and applying snapshots.
type Snapshotter interface {
	Begin()                // Begin starts a new transaction to create a snapshot.
	End() *Snapshot        // End finalizes the transaction and returns the created snapshot.
	Apply(*Snapshot) error // Apply applies the changes from the given snapshot to the underlying storage.
}

// snapshotter is an implementation of the Snapshotter interface.
type snapshotter struct {
	blockchainSnapshotter BlockchainKVSnapshotter // blockchainSnapshotter is responsible for managing the blockchain key-value storage snapshots.
	stateSnapshotter      StateStorageSnapshotter // stateSnapshotter is responsible for managing the state storage snapshots.
}

// NewSnapshotter creates a new Snapshotter instance that uses the provided logger, state storage,
// and blockchain storage path. It returns the Snapshotter, blockchain storage, state storage,
// and any error that occurred.
func NewSnapshotter(logger hclog.Logger, stateStorage itrie.Storage, blockchainStoragePath string) (Snapshotter, storage.Storage, itrie.Storage, error) {
	var blockchainSnapshotter BlockchainKVSnapshotter
	var err error

	if blockchainStoragePath == "" {
		blockchainSnapshotter, err = NewMemoryDBSnapshotStorage()
	} else {
		blockchainSnapshotter, err = NewLevelDBSnapshotStorage(blockchainStoragePath)
	}
	if err != nil {
		return nil, nil, nil, err
	}

	stateSnapshotter := StateWrapper(stateStorage)

	s := &snapshotter{
		blockchainSnapshotter: blockchainSnapshotter,
		stateSnapshotter:      stateSnapshotter,
	}

	blockchainStorage := storage.NewKeyValueStorage(logger.Named("leveldb"), blockchainSnapshotter)

	return s, blockchainStorage, stateSnapshotter, nil
}

// Begin starts a new transaction to create a snapshot.
func (s *snapshotter) Begin() {
	s.stateSnapshotter.Begin()
	s.blockchainSnapshotter.Begin()
}

// End finalizes the transaction and returns the created snapshot.
func (s *snapshotter) End() *Snapshot {
	snapshot := &Snapshot{
		BlockchainSnapshot: s.blockchainSnapshotter.End(),
		StateSnapshot:      s.stateSnapshotter.End(),
	}

	return snapshot
}

// Apply applies the changes from the given snapshot to the underlying storage.
func (s *snapshotter) Apply(snapshot *Snapshot) error {
	err := s.blockchainSnapshotter.Apply(snapshot.BlockchainSnapshot)
	if err != nil {
		return err
	}

	return s.stateSnapshotter.Apply(snapshot.StateSnapshot)
}
