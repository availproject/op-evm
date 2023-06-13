package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/0xPolygon/polygon-edge/network/common"
	"github.com/0xPolygon/polygon-edge/server/proto"
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/maticnetwork/avail-settlement/pkg/blockchain"
	empty "google.golang.org/protobuf/types/known/emptypb"
)

// systemService is an implementation of the SystemServer gRPC service which is part of the server package.
type systemService struct {
	proto.UnimplementedSystemServer
	server *Server
}

// GetStatus method in the systemService struct retrieves the status of the current system.
// It returns details about the network chain ID, current block header and the libp2p network address.
func (s *systemService) GetStatus(ctx context.Context, req *empty.Empty) (*proto.ServerStatus, error) {
	header := s.server.blockchain.Header()

	addr, err := common.AddrInfoToString(s.server.network.AddrInfo())
	if err != nil {
		return nil, err
	}

	status := &proto.ServerStatus{
		Network: s.server.chain.Params.ChainID,
		Current: &proto.ServerStatus_Block{
			Number: int64(header.Number),
			Hash:   header.Hash.String(),
		},
		P2PAddr: addr,
	}

	return status, nil
}

// Subscribe method provides a subscription service to blockchain events.
// It pushes blockchain events to the provided stream until the stream is closed.
func (s *systemService) Subscribe(req *empty.Empty, stream proto.System_SubscribeServer) error {
	sub := s.server.blockchain.SubscribeEvents()

	for {
		evnt := sub.GetEvent()
		if evnt == nil {
			break
		}

		pEvent := &proto.BlockchainEvent{
			Added:   []*proto.BlockchainEvent_Header{},
			Removed: []*proto.BlockchainEvent_Header{},
		}

		for _, h := range evnt.NewChain {
			pEvent.Added = append(
				pEvent.Added,
				&proto.BlockchainEvent_Header{Hash: h.Hash.String(), Number: int64(h.Number)},
			)
		}

		for _, h := range evnt.OldChain {
			pEvent.Removed = append(
				pEvent.Removed,
				&proto.BlockchainEvent_Header{Hash: h.Hash.String(), Number: int64(h.Number)},
			)
		}

		err := stream.Send(pEvent)

		if err != nil {
			break
		}
	}

	sub.Close()

	return nil
}

// PeersAdd method takes a peer ID and attempts to join it to the server network.
// It returns a response message along with any error that may have occurred during the operation.
func (s *systemService) PeersAdd(_ context.Context, req *proto.PeersAddRequest) (*proto.PeersAddResponse, error) {
	if joinErr := s.server.JoinPeer(req.Id); joinErr != nil {
		return &proto.PeersAddResponse{
			Message: "Unable to successfully add peer",
		}, joinErr
	}

	return &proto.PeersAddResponse{
		Message: "Peer address marked ready for dialing",
	}, nil
}

// PeersStatus method takes a peer ID and returns information about the peer, if it exists.
func (s *systemService) PeersStatus(ctx context.Context, req *proto.PeersStatusRequest) (*proto.Peer, error) {
	peerID, err := peer.Decode(req.Id)
	if err != nil {
		return nil, err
	}

	peer, err := s.getPeer(peerID)
	if err != nil {
		return nil, err
	}

	return peer, nil
}

// getPeer method returns the information of a specific peer by using the peer ID.
func (s *systemService) getPeer(id peer.ID) (*proto.Peer, error) {
	protocols, err := s.server.network.GetProtocols(id)
	if err != nil {
		return nil, err
	}

	info := s.server.network.GetPeerInfo(id)

	addrs := []string{}
	for _, addr := range info.Addrs {
		addrs = append(addrs, addr.String())
	}

	peer := &proto.Peer{
		Id:        id.String(),
		Protocols: protocols,
		Addrs:     addrs,
	}

	return peer, nil
}

// PeersList method retrieves a list of all peers in the network.
func (s *systemService) PeersList(
	ctx context.Context,
	req *empty.Empty,
) (*proto.PeersListResponse, error) {
	resp := &proto.PeersListResponse{
		Peers: []*proto.Peer{},
	}

	peers := s.server.network.Peers()
	for _, p := range peers {
		peer, err := s.getPeer(p.Info.ID)
		if err != nil {
			return nil, err
		}

		resp.Peers = append(resp.Peers, peer)
	}

	return resp, nil
}

// BlockByNumber method takes a block number and returns the corresponding block from the blockchain.
func (s *systemService) BlockByNumber(
	ctx context.Context,
	req *proto.BlockByNumberRequest,
) (*proto.BlockResponse, error) {
	block, ok := s.server.blockchain.GetBlockByNumber(req.Number, true)
	if !ok {
		return nil, fmt.Errorf("block #%d not found", req.Number)
	}

	return &proto.BlockResponse{
		Data: block.MarshalRLP(),
	}, nil
}

// Export method exports blocks from the blockchain to a given stream.
// The blocks are retrieved in chunks, from a starting block number to an ending block number.
func (s *systemService) Export(req *proto.ExportRequest, stream proto.System_ExportServer) error {
	var (
		from uint64 = 0
		to   *uint64
	)

	if req.From != from {
		from = req.From
	}

	if req.To != 0 {
		if from >= req.To {
			return errors.New("to must be greater than from")
		}

		to = &req.To
	}

	canLoop := func(i uint64) bool {
		if to == nil {
			current := s.server.blockchain.Header()

			return current != nil && i <= current.Number
		} else {
			return i <= *to
		}
	}

	writer := newBlockStreamWriter(stream, s.server.blockchain, defaultMaxGRPCPayloadSize)
	i := from

	for canLoop(i) {
		block, ok := s.server.blockchain.GetBlockByNumber(i, true)
		if !ok {
			break
		}

		if err := writer.appendBlock(block); err != nil {
			return err
		}

		i++
	}

	if err := writer.flush(); err != nil {
		return err
	}

	return nil
}

const (
	defaultMaxGRPCPayloadSize uint64 = 512 * 1024 // 4MB

	// Number of header fields * bytes per field (From, To, Latest all them uint64)
	maxHeaderInfoSize int = 3 * 8
)

// blockStreamWriter struct helps in streaming blockchain data over gRPC.
type blockStreamWriter struct {
	buf         bytes.Buffer
	blockchain  *blockchain.Blockchain
	stream      proto.System_ExportServer
	maxPayload  uint64
	pendingFrom *uint64 // first block height in buffer
	pendingTo   *uint64 // last block height in buffer
}

// newBlockStreamWriter returns an instance of blockStreamWriter with a specified maximum payload size.
func newBlockStreamWriter(
	stream proto.System_ExportServer,
	blockchain *blockchain.Blockchain,
	maxPayload uint64,
) *blockStreamWriter {
	return &blockStreamWriter{
		buf:        *bytes.NewBuffer(make([]byte, 0, maxPayload)),
		blockchain: blockchain,
		stream:     stream,
		maxPayload: maxPayload,
	}
}

// appendBlock method adds a block to the blockStreamWriter's internal buffer.
// If the block data causes the buffer to exceed its maximum capacity, the buffer is first flushed.
func (w *blockStreamWriter) appendBlock(b *types.Block) error {
	data := b.MarshalRLP()
	if uint64(maxHeaderInfoSize+w.buf.Len()+len(data)) >= w.maxPayload {
		// send buffered data to client first
		if err := w.flush(); err != nil {
			return err
		}
	}

	w.buf.Write(data)

	n := b.Number()
	if w.pendingFrom == nil {
		w.pendingFrom = &n
	}

	w.pendingTo = &n

	return nil
}

// flush method sends the data in the buffer to the client.
// After sending, it resets the buffer and block height tracking fields.
func (w *blockStreamWriter) flush() error {
	// nothing happens in case of empty buffer
	if w.buf.Len() == 0 {
		return nil
	}

	if w.pendingFrom == nil || w.pendingTo == nil {
		// should not reach
		return errors.New("pendingFrom or pendingTo is nil")
	}

	err := w.stream.Send(&proto.ExportEvent{
		From:   *w.pendingFrom,
		To:     *w.pendingTo,
		Latest: w.blockchain.Header().Number,
		Data:   w.buf.Bytes(),
	})

	if err != nil {
		return err
	}

	w.reset()

	return nil
}

// reset method clears the buffer and resets the pending block height values.
func (w *blockStreamWriter) reset() {
	w.buf.Reset()
	w.pendingFrom = nil
	w.pendingTo = nil
}
