package avail

import (
	"fmt"

	edgetypes "github.com/0xPolygon/polygon-edge/types"
	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types/codec"
)

const (
	// CallSubmitData is the RPC API call for submitting extrinsic data to Avail.
	CallSubmitData = "DataAvailability.submit_data"
)

// Sender is an interface for sending blocks to Avail.
type Sender interface {
	// Send sends a block to Avail without waiting for any status response.
	Send(blk *edgetypes.Block) error
	// SendAndWaitForStatus sends a block to Avail and waits for the specified extrinsic status.
	SendAndWaitForStatus(blk *edgetypes.Block, status types.ExtrinsicStatus) error
}

// Result represents the final result of block data submission.
type Result struct{}

// blackholeSender is an implementation of Sender that ignores sent blocks.
type blackholeSender struct{}

// Send ignores the sent block.
func (t *blackholeSender) Send(blk *edgetypes.Block) error {
	return nil
}

// SendAndWaitForStatus ignores the sent block and the specified status.
func (t *blackholeSender) SendAndWaitForStatus(blk *edgetypes.Block, status types.ExtrinsicStatus) error {
	return nil
}

// NewBlackholeSender constructs an Avail block data sender that ignores sent
// blocks - i.e. blackholes them.
func NewBlackholeSender() Sender {
	return &blackholeSender{}
}

// sender is an implementation of Sender for Avail block data submission.
type sender struct {
	appID          types.UCompact
	client         Client
	signingKeyPair signature.KeyringPair
	nextNonce      uint64
}

// NewSender constructs a block data sender for Avail.
// It takes a Client instance, appID of type types.UCompact, and a signingKeyPair of type signature.KeyringPair.
// It returns a Sender instance.
func NewSender(client Client, appID types.UCompact, signingKeyPair signature.KeyringPair) Sender {
	return &sender{
		appID:          appID,
		client:         client,
		signingKeyPair: signingKeyPair,
	}
}

// Send submits data to Avail without waiting for any status response.
// It takes a blk parameter of type *edgetypes.Block.
// It returns an error if there was a problem sending the data.
func (s *sender) Send(blk *edgetypes.Block) error {
	api, err := instance(s.client)
	if err != nil {
		return err
	}

	ext, err := s.prepareExtrinsicForSend(api, blk)
	if err != nil {
		return err
	}

	_, err = api.RPC.Author.SubmitExtrinsic(ext)
	if err != nil {
		return err
	}

	return nil
}

// SendAndWaitForStatus submits data to Avail and does not wait for the future blocks.
// It takes blk parameter of type *edgetypes.Block and dstatus parameter of type types.ExtrinsicStatus.
// It returns an error if there was a problem sending the data or if the specified status expectation is not supported.
func (s *sender) SendAndWaitForStatus(blk *edgetypes.Block, dstatus types.ExtrinsicStatus) error {
	// Only these three are supported for now.
	// NOTE: If adding new types here, handle them correspondingly in the end of
	//       the function as well!
	if !dstatus.IsFinalized && !dstatus.IsReady && !dstatus.IsInBlock {
		return fmt.Errorf("unsupported extrinsic status expectation: %#v", dstatus)
	}

	api, err := instance(s.client)
	if err != nil {
		return err
	}

	ext, err := s.prepareExtrinsicForSend(api, blk)
	if err != nil {
		return err
	}

	sub, err := api.RPC.Author.SubmitAndWatchExtrinsic(ext)
	if err != nil {
		return err
	}

	defer sub.Unsubscribe()

	for {
		select {
		case status := <-sub.Chan():
			_, err := dstatus.MarshalJSON()
			if err != nil {
				panic(err)
			}
			// NOTE: See first line of this function for supported extrinsic status expectations.
			switch {
			case dstatus.IsFinalized && status.IsFinalized:
				return nil
			case dstatus.IsInBlock && status.IsInBlock:
				return nil
			case dstatus.IsReady && status.IsReady:
				return nil
			default:
				if status.IsDropped || status.IsInvalid {
					return fmt.Errorf("unexpected extrinsic status from Avail: %#v", status)
				}
			}
		case err := <-sub.Err():
			// TODO: Consider re-connecting subscription channel on error?
			return err
		}
	}
}

// prepareExtrinsicForSend prepares the extrinsic for sending the block data.
// It takes api parameter of type *gsrpc.SubstrateAPI and blk parameter of type *edgetypes.Block.
// It returns a types.Extrinsic and an error if there was a problem preparing the extrinsic.
func (s *sender) prepareExtrinsicForSend(api *gsrpc.SubstrateAPI, blk *edgetypes.Block) (types.Extrinsic, error) {
	meta, err := api.RPC.State.GetMetadataLatest()
	if err != nil {
		return types.Extrinsic{}, err
	}

	blob := Blob{
		Magic: BlobMagic,
		Data:  blk.MarshalRLP(),
	}

	var call types.Call
	{
		// XXX: This encoding process is an inefficient hack to workaround
		// problem in the encoding pipeline from client code to Avail server.
		// `Blob` implements `scale.Encodeable` interface, but it it's passed
		// directly to `types.NewCall()`, the server will return an error. This
		// requires further investigation to fix.
		encodedBytes, err := codec.Encode(blob)
		if err != nil {
			return types.Extrinsic{}, err
		}

		call, err = types.NewCall(meta, CallSubmitData, encodedBytes)
		if err != nil {
			return types.Extrinsic{}, err
		}
	}

	ext := types.NewExtrinsic(call)

	rv, err := api.RPC.State.GetRuntimeVersionLatest()
	if err != nil {
		return types.Extrinsic{}, err
	}

	key, err := types.CreateStorageKey(meta, "System", "Account", s.signingKeyPair.PublicKey)
	if err != nil {
		return types.Extrinsic{}, err
	}

	var accountInfo types.AccountInfo
	ok, err := api.RPC.State.GetStorageLatest(key, &accountInfo)
	if err != nil || !ok {
		return types.Extrinsic{}, fmt.Errorf("couldn't fetch latest account storage info")
	}

	nonce := uint64(accountInfo.Nonce)
	if s.nextNonce > nonce {
		nonce = s.nextNonce
	}
	s.nextNonce = nonce + 1
	o := types.SignatureOptions{
		// This transaction is Immortal (https://wiki.polkadot.network/docs/build-protocol-info#transaction-mortality)
		// Hence BlockHash: Genesis Hash.
		BlockHash:          s.client.GenesisHash(),
		Era:                types.ExtrinsicEra{IsMortalEra: false},
		GenesisHash:        s.client.GenesisHash(),
		Nonce:              types.NewUCompactFromUInt(nonce),
		SpecVersion:        rv.SpecVersion,
		Tip:                types.NewUCompactFromUInt(100),
		AppID:              s.appID,
		TransactionVersion: rv.TransactionVersion,
	}

	err = ext.Sign(s.signingKeyPair, o)
	if err != nil {
		return types.Extrinsic{}, err
	}

	return ext, nil
}
