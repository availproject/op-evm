package avail

import (
	"errors"

	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/hashicorp/go-hclog"
)

var ErrUnsupportedClient = errors.New("unsupported client")

// Client is an abstraction on Avail JSON-RPC client.
type Client interface {
	// BlockStream creates a new Avail block stream, starting from block height
	// `offset`.
	BlockStream(offset uint64) BlockStream
	GenesisHash() types.Hash
	GetLatestHeader() (*types.Header, error)
	SearchBlock(offset int, searchFunc SearchFunc) (*types.SignedBlock, error)
}

type client struct {
	api         *gsrpc.SubstrateAPI
	genesisHash types.Hash
}

// NewClient constructs a new Avail Client for `url`.
func NewClient(url string) (Client, error) {
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
	}, nil
}

func instance(c Client) (*gsrpc.SubstrateAPI, error) {
	c2, ok := c.(*client)
	if ok {
		return c2.instance(), nil
	}

	return nil, ErrUnsupportedClient
}

func (c *client) instance() *gsrpc.SubstrateAPI {
	return c.api
}

func (c *client) BlockStream(offset uint64) BlockStream {
	return newBlockStream(c, hclog.Default(), offset)
}

func (c *client) GenesisHash() types.Hash {
	return c.genesisHash
}

func (c *client) GetLatestHeader() (*types.Header, error) {
	return c.api.RPC.Chain.GetHeaderLatest()
}

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
