package avail

import (
	"errors"

	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/hashicorp/go-hclog"
)

// ErrUnsupportedClient indicates that the client is not supported.
var ErrUnsupportedClient = errors.New("unsupported client")

// Client is an abstraction on Avail JSON-RPC client.
type Client interface {
	// BlockStream creates a new Avail block stream, starting from the specified block height offset.
	BlockStream(offset uint64) BlockStream

	// GenesisHash returns the genesis hash of the Avail network.
	GenesisHash() types.Hash

	// GetLatestHeader retrieves the latest header from the Avail network.
	GetLatestHeader() (*types.Header, error)

	// SearchBlock searches for a block at the specified offset using the provided search function.
	SearchBlock(offset int64, searchFunc SearchFunc) (*types.SignedBlock, error)
}

// client is an implementation of the Client interface.
type client struct {
	api         *gsrpc.SubstrateAPI
	genesisHash types.Hash
	logger      hclog.Logger
}

// NewClient constructs a new Avail Client for the specified URL.
//
// Parameters:
//   - url: The URL of the Avail JSON-RPC server.
//   - logger: The logger instance.
//
// Return:
//   - Client: The Avail client instance.
//   - error: An error if the client initialization fails.
func NewClient(url string, logger hclog.Logger) (Client, error) {

	api, err := gsrpc.NewSubstrateAPI(url)
	if err != nil {
		return nil, err
	}

	// Cache genesis hash as it will never change.
	genesisHash, err := api.RPC.Chain.GetBlockHash(0)
	if err != nil {
		return nil, err
	}

	return &client{
		api:         api,
		genesisHash: genesisHash,
		logger:      logger,
	}, nil
}

// instance returns the underlying SubstrateAPI instance.
//
// Return:
//   - *gsrpc.SubstrateAPI: The SubstrateAPI instance.
//   - error: An error if the client is not supported or found.
func instance(c Client) (*gsrpc.SubstrateAPI, error) {
	c2, ok := c.(*client)
	if ok {
		return c2.instance(), nil
	}

	return nil, ErrUnsupportedClient
}

// instance returns the underlying SubstrateAPI instance.
//
// Return:
//   - *gsrpc.SubstrateAPI: The SubstrateAPI instance.
func (c *client) instance() *gsrpc.SubstrateAPI {
	return c.api
}

// BlockStream creates a new Avail block stream starting from the specified offset.
//
// Parameters:
//   - offset: The block height offset to start the stream from.
//
// Return:
//   - BlockStream: The block stream.
func (c *client) BlockStream(offset uint64) BlockStream {
	return newBlockStream(c, c.logger, offset)
}

// GenesisHash returns the genesis hash of the Avail network.
//
// Return:
//   - types.Hash: The genesis hash.
func (c *client) GenesisHash() types.Hash {
	return c.genesisHash
}

// GetLatestHeader retrieves the latest header from the Avail network.
//
// Return:
//   - *types.Header: The latest header.
//   - error: An error if the retrieval fails.
func (c *client) GetLatestHeader() (*types.Header, error) {
	return c.api.RPC.Chain.GetHeaderLatest()
}

// FindCallIndex finds the call index for CallSubmitData in the Avail network.
//
// Parameters:
//   - client: The Avail client.
//
// Return:
//   - types.CallIndex: The call index for CallSubmitData.
//   - error: An error if the call index retrieval fails.
func FindCallIndex(client Client) (types.CallIndex, error) {
	api, err := instance(client)
	if err == ErrUnsupportedClient {
		return types.CallIndex{}, nil
	} else if err != nil {
		return types.CallIndex{}, err
	}

	meta, err := api.RPC.State.GetMetadataLatest()
	if err != nil {
		return types.CallIndex{}, err
	}

	callIdx, err := meta.FindCallIndex(CallSubmitData)
	if err != nil {
		return types.CallIndex{}, err
	}

	return callIdx, nil
}
