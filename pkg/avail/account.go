package avail

import (
	"fmt"
	"math/big"
	"os"

	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/tyler-smith/go-bip39"
)

const (
	// 1 AVL == 10^18 Avail fractions.
	AVL = 1_000_000_000_000_000_000
)

// NewAccount generates a new Avail account by creating a mnemonic phrase and deriving the key pair.
// It returns the generated key pair and an error if there is an issue.
func NewAccount() (signature.KeyringPair, error) {
	entropy, err := bip39.NewEntropy(128)
	if err != nil {
		return signature.KeyringPair{}, err
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return signature.KeyringPair{}, err
	}

	keyPair, err := signature.KeyringPairFromSecret(mnemonic, 42)
	if err != nil {
		return signature.KeyringPair{}, err
	}

	return keyPair, nil
}

// NewAccountFromMnemonic generates an Avail account using the provided mnemonic phrase.
// It returns the generated key pair and an error if there is an issue.
func NewAccountFromMnemonic(mnemonic string) (signature.KeyringPair, error) {
	keyPair, err := signature.KeyringPairFromSecret(mnemonic, 42)
	if err != nil {
		return signature.KeyringPair{}, err
	}

	return keyPair, nil
}

// AccountFromFile reads an Avail account from a file containing the mnemonic phrase.
// It returns the generated key pair and an error if there is an issue.
func AccountFromFile(filePath string) (signature.KeyringPair, error) {
	accountBytes, err := os.ReadFile(filePath)
	if err != nil {
		return signature.KeyringPair{}, fmt.Errorf("failure to read account file '%s'", err)
	}

	availAccount, err := NewAccountFromMnemonic(string(accountBytes))
	if err != nil {
		return signature.KeyringPair{}, err
	}

	return availAccount, nil
}

// AccountExistsFromMnemonic checks if an Avail account exists on the blockchain using the provided mnemonic phrase.
// It takes a client and the file path of the mnemonic phrase, and returns a boolean indicating if the account exists and an error if there is an issue.
func AccountExistsFromMnemonic(client Client, filePath string) (bool, error) {
	account, err := AccountFromFile(filePath)
	if err != nil {
		return false, err
	}

	api, err := instance(client)
	if err != nil {
		return false, err
	}

	meta, err := api.RPC.State.GetMetadataLatest()
	if err != nil {
		return false, err
	}

	key, err := types.CreateStorageKey(meta, "System", "Account", account.PublicKey, nil)
	if err != nil {
		return false, err
	}

	var accountInfo types.AccountInfo
	return api.RPC.State.GetStorageLatest(key, &accountInfo)
}

// DepositBalance deposits a specified amount of Avail tokens from the specified account to the specified recipient.
// It takes a client, the account key pair, the amount to deposit, and the nonce increment.
// It returns an error if there is an issue.
func DepositBalance(client Client, account signature.KeyringPair, amount, nonceIncrement uint64) error {
	api, err := instance(client)
	if err != nil {
		return err
	}

	meta, err := api.RPC.State.GetMetadataLatest()
	if err != nil {
		return err
	}

	addr, err := types.NewMultiAddressFromAccountID(account.PublicKey)
	if err != nil {
		return err
	}

	c, err := types.NewCall(meta, "Balances.transfer", addr, types.NewUCompactFromUInt(amount))
	if err != nil {
		return err
	}

	// Create the extrinsic
	ext := types.NewExtrinsic(c)

	genesisHash, err := api.RPC.Chain.GetBlockHash(0)
	if err != nil {
		return err
	}

	rv, err := api.RPC.State.GetRuntimeVersionLatest()
	if err != nil {
		return err
	}

	key, err := types.CreateStorageKey(meta, "System", "Account", signature.TestKeyringPairAlice.PublicKey, nil)
	if err != nil {
		return err
	}

	var accountInfo types.AccountInfo
	ok, err := api.RPC.State.GetStorageLatest(key, &accountInfo)
	if err != nil || !ok {
		return fmt.Errorf("couldn't fetch latest alice account storage info: %w", err)
	}

	nonce := uint64(accountInfo.Nonce)

	if nonceIncrement > 0 {
		nonce = nonce + nonceIncrement
	}

	o := types.SignatureOptions{
		BlockHash:          genesisHash,
		Era:                types.ExtrinsicEra{IsMortalEra: false},
		GenesisHash:        genesisHash,
		Nonce:              types.NewUCompactFromUInt(nonce),
		SpecVersion:        rv.SpecVersion,
		Tip:                types.NewUCompactFromUInt(0),
		AppID:              types.NewUCompactFromUInt(0),
		TransactionVersion: rv.TransactionVersion,
	}

	// Sign the transaction using Alice's default account
	err = ext.Sign(signature.TestKeyringPairAlice, o)
	if err != nil {
		return err
	}

	// Send the extrinsic
	sub, err := api.RPC.Author.SubmitAndWatchExtrinsic(ext)
	if err != nil {
		return err
	}

	defer sub.Unsubscribe()

	for {
		select {
		case status := <-sub.Chan():
			// NOTE: See first line of this function for supported extrinsic status expectations.
			switch {
			case status.IsFinalized:
				return nil
			case status.IsInBlock:
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

// GetBalance retrieves the Avail token balance of the specified account.
// It takes a client and the account key pair, and returns the account balance as a *big.Int and an error if there is an issue.
func GetBalance(client Client, account signature.KeyringPair) (*big.Int, error) {
	api, err := instance(client)
	if err != nil {
		return nil, err
	}

	meta, err := api.RPC.State.GetMetadataLatest()
	if err != nil {
		return nil, err
	}

	key, err := types.CreateStorageKey(meta, "System", "Account", account.PublicKey, nil)
	if err != nil {
		return nil, err
	}

	var accountInfo types.AccountInfo
	ok, err := api.RPC.State.GetStorageLatest(key, &accountInfo)
	if err != nil || !ok {
		return nil, err
	}

	return new(big.Int).Div(new(big.Int).SetUint64(accountInfo.Data.Free.Uint64()), big.NewInt(AVL)), nil
}
