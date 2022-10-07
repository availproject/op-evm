package avail

import (
	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

// Client is an abstraction on Avail JSON-RPC client.
type Client interface {
	GenesisHash() types.Hash
	instance() *gsrpc.SubstrateAPI
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

func (c *client) instance() *gsrpc.SubstrateAPI {
	return c.api
}

func (c *client) GenesisHash() types.Hash {
	return c.genesisHash
}

func FindCallIndex(client Client) (types.CallIndex, error) {
	meta, err := client.instance().RPC.State.GetMetadataLatest()
	if err != nil {
		return types.CallIndex{}, err
	}

	callIdx, err := meta.FindCallIndex(CallSubmitData)
	if err != nil {
		return types.CallIndex{}, err
	}

	return callIdx, nil
}
