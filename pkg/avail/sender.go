package avail

import (
	"fmt"

	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/maticnetwork/avail-settlement/pkg/future"
)

// AvailBridgeAppID is the Avail application ID for the bridge.
const AvailBridgeAppID = uint32(0)

// Sender provides interface for sending blocks to Avail. It returns a Future
// to query result of block finalisation.
type Sender interface {
	SubmitData(b *Batch) future.Future[Result]
}

// Result contains the final result of block data submission.
type Result struct{}

type sender struct {
	client Client
}

// NewSender constructs an Avail block data sender.
func NewSender(client Client) Sender {
	return &sender{client: client}
}

// SubmitData submits a Batch to Avail and returns a Future with Result or an
// error.
func (s *sender) SubmitData(b *Batch) future.Future[Result] {
	api := s.client.instance()
	f := future.New[Result]()

	meta, err := api.RPC.State.GetMetadataLatest()
	if err != nil {
		f.SetError(err)
		return f
	}

	// TODO: Refactor Batch to support Encode and serialize Batch here.
	bs := types.NewBytes([]byte("hello world"))

	call, err := types.NewCall(meta, "DataAvailability.submit_data", bs)
	if err != nil {
		f.SetError(err)
		return f
	}

	ext := types.NewExtrinsic(call)

	hash, err := api.RPC.Chain.GetBlockHashLatest()
	if err != nil {
		f.SetError(err)
		return f
	}

	rv, err := api.RPC.State.GetRuntimeVersionLatest()
	if err != nil {
		f.SetError(err)
		return f
	}

	key, err := types.CreateStorageKey(meta, "System", "Account", signature.TestKeyringPairAlice.PublicKey)
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
		BlockHash:          hash,
		Era:                types.ExtrinsicEra{IsMortalEra: false},
		GenesisHash:        s.client.GenesisHash(),
		Nonce:              types.NewUCompactFromUInt(uint64(nonce)),
		SpecVersion:        rv.SpecVersion,
		Tip:                types.NewUCompactFromUInt(100),
		AppID:              types.NewU32(AvailBridgeAppID),
		TransactionVersion: rv.TransactionVersion,
	}

	err = ext.Sign(signature.TestKeyringPairAlice, o)
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
					f.SetValue(Result{})
					return
				}
			case err := <-sub.Err():
				// TODO: Consider re-connecting subscription channel on error?
				f.SetError(err)
				return
			}
		}
	}()

	return f
}
