package avail

import (
	"errors"
	"fmt"

	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

const (
	// DefaultAppID is the Avail application ID.
	DefaultAppID = types.U32(0)

	// CallCreateApplicationKey is the RPC API call for creating new AppID on Avail.
	CallCreateApplicationKey = "DataAvailability.create_application_key"
)

var (
	ErrAppIDNotFound = errors.New("AppID not found")
)

func EnsureApplicationKeyExists(client Client, applicationKey string, signingKeyPair signature.KeyringPair) (types.U32, error) {
	appID, err := QueryAppID(client, applicationKey)
	if errors.Is(err, ErrAppIDNotFound) {
		appID, err = CreateApplicationKey(client, applicationKey, signingKeyPair)
		if err != nil {
			return 0, err
		}
	} else if err != nil {
		return 0, err
	}

	return appID, nil
}

func QueryAppID(client Client, applicationKey string) (types.U32, error) {
	api := client.instance()

	meta, err := api.RPC.State.GetMetadataLatest()
	if err != nil {
		return 0, err
	}

	encodedBytes, err := types.EncodeToBytes("avail-settlement")
	if err != nil {
		return 0, err
	}

	key, err := types.CreateStorageKey(meta, "DataAvailability", "AppKeys", encodedBytes)
	if err != nil {
		return 0, err
	}

	type AppKeyInfo struct {
		AccountID types.AccountID
		AppID     types.U32
	}

	var aki AppKeyInfo
	ok, err := api.RPC.State.GetStorageLatest(key, &aki)
	if err != nil {
		return 0, err
	}

	if ok {
		return aki.AppID, nil
	} else {
		return 0, ErrAppIDNotFound
	}
}

func CreateApplicationKey(client Client, applicationKey string, signingKeyPair signature.KeyringPair) (types.U32, error) {
	api := client.instance()

	meta, err := api.RPC.State.GetMetadataLatest()
	if err != nil {
		return 0, err
	}

	call, err := types.NewCall(meta, CallCreateApplicationKey, []byte(applicationKey))
	if err != nil {
		return 0, err
	}

	ext := types.NewExtrinsic(call)

	rv, err := api.RPC.State.GetRuntimeVersionLatest()
	if err != nil {
		return 0, err
	}

	key, err := types.CreateStorageKey(meta, "System", "Account", signingKeyPair.PublicKey)
	if err != nil {
		return 0, err
	}

	var accountInfo types.AccountInfo
	ok, err := api.RPC.State.GetStorageLatest(key, &accountInfo)
	if err != nil || !ok {
		return 0, fmt.Errorf("couldn't fetch latest account storage info: %w", err)
	}

	genesisHash, err := api.RPC.Chain.GetBlockHash(0)
	if err != nil {
		return 0, err
	}

	nonce := uint64(accountInfo.Nonce)
	o := types.SignatureOptions{
		// This transaction is Immortal (https://wiki.polkadot.network/docs/build-protocol-info#transaction-mortality)
		// Hence BlockHash: Genesis Hash.
		BlockHash:          genesisHash,
		Era:                types.ExtrinsicEra{IsMortalEra: false},
		GenesisHash:        genesisHash,
		Nonce:              types.NewUCompactFromUInt(nonce),
		SpecVersion:        rv.SpecVersion,
		Tip:                types.NewUCompactFromUInt(100),
		AppID:              DefaultAppID,
		TransactionVersion: rv.TransactionVersion,
	}

	err = ext.Sign(signingKeyPair, o)
	if err != nil {
		return 0, err
	}

	sub, err := api.RPC.Author.SubmitAndWatchExtrinsic(ext)
	if err != nil {
		return 0, err
	}

	defer sub.Unsubscribe()

	for {
		select {
		case status := <-sub.Chan():
			if status.IsInBlock {
				return QueryAppID(client, applicationKey)
			}

			if status.IsDropped || status.IsInvalid {
				return 0, fmt.Errorf("unexpected extrinsic status from Avail: %#v", status)
			}

		case err = <-sub.Err():
			return 0, fmt.Errorf("error while waiting for application key creation status: %w", err)
		}
	}
}
