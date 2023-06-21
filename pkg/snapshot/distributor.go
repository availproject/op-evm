package snapshot

import (
	"github.com/0xPolygon/polygon-edge/network"
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/hashicorp/go-hclog"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/maticnetwork/avail-settlement/pkg/snapshot/proto"
)

const (
	// topicNameV1 is the P2P gossip pubsub topic name.
	topicNameV1 = "stateSnapshots/0.1"

	// pendingSnapshotQueueLen is the maximum number of incoming snapshots pending
	// processing in the distributor's channel.
	pendingSnapshotQueueLen = 64
)

// Distributor is responsible for distributing snapshots among network peers.
type Distributor interface {
	Receive() <-chan *Snapshot
	Send(s *Snapshot) error
	Close() error
}

// distributor is an implementation of the Distributor interface.
type distributor struct {
	logger     hclog.Logger
	snapshotCh chan *Snapshot
	topic      *network.Topic
}

// NewDistributor creates a new distributor instance that uses the given logger and network server.
// It subscribes to the gossip topic for snapshot distribution.
func NewDistributor(logger hclog.Logger, network *network.Server) (Distributor, error) {
	topic, err := network.NewTopic(topicNameV1, &proto.StateSnapshot{})
	if err != nil {
		return nil, err
	}

	d := &distributor{
		logger:     logger.Named("snapshot.distributor"),
		snapshotCh: make(chan *Snapshot, pendingSnapshotQueueLen),
		topic:      topic,
	}

	err = topic.Subscribe(d.handleIncomingSnapshot)
	if err != nil {
		topic.Close()
		return nil, err
	}

	return d, nil
}

// handleIncomingSnapshot handles incoming snapshots from network peers.
func (d *distributor) handleIncomingSnapshot(obj interface{}, from peer.ID) {
	d.logger.Debug("received state snapshot", "peer", from.Pretty())

	raw, ok := obj.(*proto.StateSnapshot)
	if !ok {
		d.logger.Error("failed to cast gossiped message to snapshot")
		return
	}

	if raw == nil || raw.BlockchainSnapshot == nil || raw.StateSnapshot == nil {
		d.logger.Error("malformed gossip transaction message received")
		return
	}

	nBlockchainChanges := len(raw.BlockchainSnapshot.Keys)
	nStateChanges := len(raw.StateSnapshot.Keys)

	snapshot := &Snapshot{
		BlockNumber: raw.BlockNumber,
		BlockHash:   types.BytesToHash(raw.BlockHash),
		StateRoot:   types.BytesToHash(raw.StateRoot),

		BlockchainSnapshot: &BlockchainSnapshot{
			Keys:   make([][]byte, nBlockchainChanges),
			Values: make([][]byte, nBlockchainChanges),
		},
		StateSnapshot: &StateStorageSnapshot{
			Keys:   make([][]byte, nStateChanges),
			Values: make([][]byte, nStateChanges),
		},
	}

	for i, k := range raw.BlockchainSnapshot.Keys {
		snapshot.BlockchainSnapshot.Keys[i] = k
		snapshot.BlockchainSnapshot.Values[i] = raw.BlockchainSnapshot.Values[i]
	}

	for i, k := range raw.StateSnapshot.Keys {
		snapshot.StateSnapshot.Keys[i] = k
		snapshot.StateSnapshot.Values[i] = raw.StateSnapshot.Values[i]
	}

	d.snapshotCh <- snapshot
}

// Receive returns the channel for receiving incoming snapshots.
func (d *distributor) Receive() <-chan *Snapshot {
	return d.snapshotCh
}

// Send sends the given snapshot to network peers.
func (d *distributor) Send(s *Snapshot) error {
	snpshot := &proto.StateSnapshot{
		BlockNumber: s.BlockNumber,
		BlockHash:   s.BlockHash.Bytes(),
		StateRoot:   s.StateRoot.Bytes(),

		BlockchainSnapshot: &proto.KeyValuePairs{
			Keys:   s.BlockchainSnapshot.Keys,
			Values: s.BlockchainSnapshot.Values,
		},

		StateSnapshot: &proto.KeyValuePairs{
			Keys:   s.StateSnapshot.Keys,
			Values: s.StateSnapshot.Values,
		},
	}

	return d.topic.Publish(snpshot)
}

// Close closes the distributor and unsubscribes from the gossip topic.
func (d *distributor) Close() error {
	d.topic.Close()
	return nil
}
