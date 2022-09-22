package block

import (
	"crypto/rand"
	"testing"

	"github.com/0xPolygon/polygon-edge/crypto"
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/umbracle/fastrlp"
)

func Test_WriteSeal(t *testing.T) {
	hdr := &types.Header{}

	hdr.ExtraData = make([]byte, SequencerExtraVanity)

	extraData := append([]byte{}, FraudproofPrefix...)
	extraData = append(extraData, hdr.Hash.Bytes()...)
	ar := &fastrlp.Arena{}
	rlpExtraData, err := ar.NewBytes(extraData).Bytes()
	if err != nil {
		panic(err)
	}

	copy(hdr.ExtraData, rlpExtraData)

	ve := &ValidatorExtra{}
	bs := ve.MarshalRLPTo(nil)
	hdr.ExtraData = append(hdr.ExtraData, bs...)

	key := keystore.NewKeyForDirectICAP(rand.Reader)
	miner := crypto.PubKeyToAddress(&key.PrivateKey.PublicKey)

	hdr.Miner = miner.Bytes()

	hdr, err = WriteSeal(key.PrivateKey, hdr)
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
