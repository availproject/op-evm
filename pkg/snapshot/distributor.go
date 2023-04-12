package snapshot

import "github.com/0xPolygon/polygon-edge/network"

const topicNameV1 = "stateSnapshots/0.1"

type Distributor interface {
	Receive() <-chan *Snapshot
	Send(s *Snapshot) error

	Close() error
}

type distributor struct {
	snapshotCh chan *Snapshot
	topic      *network.Topic
}

func NewDistributor(network *network.Server) (Distributor, error) {
	d := &distributor{
		snapshotCh: make(chan *Snapshot),
		topic:      nil, // TODO: Network wiring in separate PR
	}

	return d, nil
}

func (d *distributor) Receive() <-chan *Snapshot {
	return d.snapshotCh
}

func (d *distributor) Send(s *Snapshot) error {
	return nil
}

func (d *distributor) Close() error {
	return nil
}
