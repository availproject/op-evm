package avail

import (
	"crypto/rand"
	"testing"

	"github.com/0xPolygon/polygon-edge/types"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/umbracle/fastrlp"
)

// TODO(tuommaki): This will be mostly useless test. Should be removed after initial development.
func Test_writeSeal(t *testing.T) {
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

	_, err = writeSeal(key.PrivateKey, hdr)
	if err != nil {
		t.Fatal(err)
	}
}
