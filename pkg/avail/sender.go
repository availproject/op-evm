package avail

import (
	"fmt"
	"log"

	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/maticnetwork/avail-settlement/pkg/future"
)

const (
	// BridgeAppID is the Avail application ID for the bridge.
	BridgeAppID = uint32(0)

	// CallSubmitData is the RPC API call for submitting extrinsic data to Avail.
	CallSubmitData = "DataAvailability.submit_data"
)

// Sender provides interface for sending blocks to Avail. It returns a Future
// to query result of block finalisation.
type Sender interface {
	SubmitData(bs []byte) future.Future[Result]
	SubmitDataWithoutWatch(bs []byte) (*types.Hash, error)
	SubmitDataAndWaitForStatus(bs []byte, status types.ExtrinsicStatus) future.Future[Result]
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

// SubmitData submits data to Avail and returns a Future with Result or an
// error.
func (s *sender) SubmitData(bs []byte) future.Future[Result] {
	api := s.client.instance()
	f := future.New[Result]()

	meta, err := api.RPC.State.GetMetadataLatest()
	if err != nil {
		f.SetError(err)
		return f
	}

	blob := Blob{
		Magic: BlobMagic,
		Data:  bs,
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
			f.SetError(err)
			return f
		}

		call, err = types.NewCall(meta, CallSubmitData, encodedBytes)
		if err != nil {
			f.SetError(err)
			return f
		}
	}

	ext := types.NewExtrinsic(call)

	rv, err := api.RPC.State.GetRuntimeVersionLatest()
	if err != nil {
		f.SetError(err)
		return f
	}

	key, err := types.CreateStorageKey(meta, "System", "Account", s.signingKeyPair.PublicKey)
	if err != nil {
		f.SetError(err)
		return f
	}

	var accountInfo types.AccountInfo
	ok, err := api.RPC.State.GetStorageLatest(key, &accountInfo)
	if err != nil || !ok {
		f.SetError(fmt.Errorf("couldn't fetch latest account storage info"))
		return f
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
		f.SetError(err)
		return f
	}

	sub, err := api.RPC.Author.SubmitAndWatchExtrinsic(ext)
	if err != nil {
		f.SetError(err)
		return f
	}

	go func() {
		defer sub.Unsubscribe()

		for {
			select {
			case status := <-sub.Chan():
				if status.IsFinalized {
					log.Printf("submitted block is finalized.")
					f.SetValue(Result{})
					return
				}
			case err := <-sub.Err():
				// TODO: Consider re-connecting subscription channel on error?
				log.Printf("submitted block subscription returned an error: %s", err)
				f.SetError(err)
				return
			}
		}
	}()

	return f
}

// SubmitDataWithoutWatch submits data to Avail and does not wait for the future blocks
func (s *sender) SubmitDataWithoutWatch(bs []byte) (*types.Hash, error) {
	api := s.client.instance()

	meta, err := api.RPC.State.GetMetadataLatest()
	if err != nil {
		return nil, err
	}

	blob := Blob{
		Magic: BlobMagic,
		Data:  bs,
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
			fmt.Printf("encode to bytes err: %s", err)
			return nil, err
		}

		call, err = types.NewCall(meta, CallSubmitData, encodedBytes)
		if err != nil {
			fmt.Printf("new call err: %s", err)
			return nil, err
		}
	}

	ext := types.NewExtrinsic(call)

	rv, err := api.RPC.State.GetRuntimeVersionLatest()
	if err != nil {
		fmt.Printf("get runtime version latest err: %s", err)
		return nil, err
	}

	key, err := types.CreateStorageKey(meta, "System", "Account", s.signingKeyPair.PublicKey)
	if err != nil {
		fmt.Printf("create storage key err: %s", err)
		return nil, err
	}

	var accountInfo types.AccountInfo
	ok, err := api.RPC.State.GetStorageLatest(key, &accountInfo)
	if err != nil || !ok {
		fmt.Printf("get storage latest err: %s", err)
		return nil, err
	}

	nonce := uint32(accountInfo.Nonce)

	o := types.SignatureOptions{
		// This transaction is Immortal (https://wiki.polkadot.network/docs/build-protocol-info#transaction-mortality)
		// Hence BlockHash: Genesis Hash.
		BlockHash:   s.client.GenesisHash(),
		Era:         types.ExtrinsicEra{IsMortalEra: false},
		GenesisHash: s.client.GenesisHash(),
		Nonce:       types.NewUCompactFromUInt(uint64(nonce + 1)),
		//Nonce:              types.NewUCompactFromUInt(nonce),
		SpecVersion:        rv.SpecVersion,
		Tip:                types.NewUCompactFromUInt(100),
		AppID:              types.NewU32(BridgeAppID),
		TransactionVersion: rv.TransactionVersion,
	}

	err = ext.Sign(s.signingKeyPair, o)
	if err != nil {
		fmt.Printf("sign err: %s", err)
		return nil, err
	}

	hash, err := api.RPC.Author.SubmitExtrinsic(ext)
	if err != nil {
		fmt.Printf("submit extrinsic err: %s", err)
		return nil, err
	}

	return &hash, nil
}

// SubmitDataWithoutWatch submits data to Avail and does not wait for the future blocks
func (s *sender) SubmitDataAndWaitForStatus(bs []byte, dstatus types.ExtrinsicStatus) future.Future[Result] {
	api := s.client.instance()
	f := future.New[Result]()

	meta, err := api.RPC.State.GetMetadataLatest()
	if err != nil {
		f.SetError(err)
		return f
	}

	blob := Blob{
		Magic: BlobMagic,
		Data:  bs,
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
			f.SetError(err)
			return f
		}

		call, err = types.NewCall(meta, CallSubmitData, encodedBytes)
		if err != nil {
			f.SetError(err)
			return f
		}
	}

	ext := types.NewExtrinsic(call)

	rv, err := api.RPC.State.GetRuntimeVersionLatest()
	if err != nil {
		f.SetError(err)
		return f
	}

	key, err := types.CreateStorageKey(meta, "System", "Account", s.signingKeyPair.PublicKey)
	if err != nil {
		f.SetError(err)
		return f
	}

	var accountInfo types.AccountInfo
	ok, err := api.RPC.State.GetStorageLatest(key, &accountInfo)
	if err != nil || !ok {
		f.SetError(fmt.Errorf("couldn't fetch latest account storage info"))
		return f
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
		f.SetError(err)
		return f
	}

	log.Printf("XXXXXXX: Sending Avail block\n")

	sub, err := api.RPC.Author.SubmitAndWatchExtrinsic(ext)
	if err != nil {
		f.SetError(err)
		return f
	}
	log.Printf("XXXXXXX: Sent Avail block\n")

	go func() {
		defer sub.Unsubscribe()

		for {
			log.Printf("waiting for Avail send block subscriptions status\n")
			select {
			case status := <-sub.Chan():
				sts, err := dstatus.MarshalJSON()
				if err != nil {
					panic(err)
				}
				log.Printf("FOOBAR - status from Avail: %q", string(sts))
				if dstatus.IsInBlock {
					if status.IsInBlock {
						log.Printf("submitted block is in block.")
						f.SetValue(Result{})
						return
					}
				} else if dstatus.IsReady {
					if status.IsReady {
						log.Printf("submitted block is ready.")
						f.SetValue(Result{})
						return
					}
				}
			case err := <-sub.Err():
				// TODO: Consider re-connecting subscription channel on error?
				log.Printf("submitted block subscription returned an error: %s", err)
				f.SetError(err)
				return
			}
		}
	}()

	return f
}
