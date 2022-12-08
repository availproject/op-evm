package avail

import (
	"fmt"

	edgetypes "github.com/0xPolygon/polygon-edge/types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

const (
	// BridgeAppID is the Avail application ID for the bridge.
	BridgeAppID = uint32(0)

	// CallSubmitData is the RPC API call for submitting extrinsic data to Avail.
	CallSubmitData = "DataAvailability.submit_data"
)

// Sender provides interface for sending blocks to Avail.
type Sender interface {
	Send(blk *edgetypes.Block) error
	SendAndWaitForStatus(blk *edgetypes.Block, status types.ExtrinsicStatus) error
}

// Result contains the final result of block data submission.
type Result struct{}

type sender struct {
	client         Client
	signingKeyPair signature.KeyringPair
}

// NewSender constructs an Avail block data sender.
func NewSender(client Client, signingKeyPair signature.KeyringPair) Sender {
	return &sender{
		client:         client,
		signingKeyPair: signingKeyPair,
	}
}

// Send submits data to Avail.
func (s *sender) Send(blk *edgetypes.Block) error {
	return s.SendAndWaitForStatus(blk, types.ExtrinsicStatus{IsFinalized: true})
}

// SendAndWaitForStatus submits data to Avail and does not wait for the future blocks
func (s *sender) SendAndWaitForStatus(blk *edgetypes.Block, dstatus types.ExtrinsicStatus) error {
	api := s.client.instance()

	meta, err := api.RPC.State.GetMetadataLatest()
	if err != nil {
		return err
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
		encodedBytes, err := types.EncodeToBytes(blob)
		if err != nil {
			return err
		}

		call, err = types.NewCall(meta, CallSubmitData, encodedBytes)
		if err != nil {
			return err
		}
	}

	ext := types.NewExtrinsic(call)

	rv, err := api.RPC.State.GetRuntimeVersionLatest()
	if err != nil {
		return err
	}

	key, err := types.CreateStorageKey(meta, "System", "Account", s.signingKeyPair.PublicKey)
	if err != nil {
		return err
	}

	var accountInfo types.AccountInfo
	ok, err := api.RPC.State.GetStorageLatest(key, &accountInfo)
	if err != nil || !ok {
		return fmt.Errorf("couldn't fetch latest account storage info")
	}

	nonce := uint32(accountInfo.Nonce)
	o := types.SignatureOptions{
		// This transaction is Immortal (https://wiki.polkadot.network/docs/build-protocol-info#transaction-mortality)
		// Hence BlockHash: Genesis Hash.
		BlockHash:          s.client.GenesisHash(),
		Era:                types.ExtrinsicEra{IsMortalEra: false},
		GenesisHash:        s.client.GenesisHash(),
		Nonce:              types.NewUCompactFromUInt(uint64(nonce)),
		SpecVersion:        rv.SpecVersion,
		Tip:                types.NewUCompactFromUInt(100),
		AppID:              types.NewU32(BridgeAppID),
		TransactionVersion: rv.TransactionVersion,
	}

	err = ext.Sign(s.signingKeyPair, o)
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
			switch {
			case dstatus.IsFinalized && status.IsFinalized:
				return nil
			case dstatus.IsInBlock && status.IsInBlock:
				return nil
			case dstatus.IsReady && status.IsReady:
				return nil
			default:
				// TODO: Handle other statuses properly.
			}
		case err := <-sub.Err():
			// TODO: Consider re-connecting subscription channel on error?
			return err
		}
	}
}
