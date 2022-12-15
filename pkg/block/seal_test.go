package block

import (
	"crypto/rand"
	"testing"

	"github.com/0xPolygon/polygon-edge/crypto"
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/ethereum/go-ethereum/accounts/keystore"
)

func Test_WriteSeal(t *testing.T) {
	hdr := &types.Header{}

	kv := make(map[string][]byte)
	kv[KeyFraudProofOf] = hdr.Hash.Bytes()

	ve := &ValidatorExtra{}
	kv[KeyExtraValidators] = ve.MarshalRLPTo(nil)

	hdr.ExtraData = EncodeExtraDataFields(kv)

	key := keystore.NewKeyForDirectICAP(rand.Reader)
	miner := crypto.PubKeyToAddress(&key.PrivateKey.PublicKey)

	hdr.Miner = miner.Bytes()

	hdr, err := WriteSeal(key.PrivateKey, hdr)
	if err != nil {
		t.Fatal(err)
	}

	signer, err := AddressRecoverFromHeader(hdr)
	if err != nil {
		t.Fatal(err)
	}

	if signer != miner {
		t.Fatalf("signer != miner, signer: %q, miner: %q", signer, miner)
	}
}
