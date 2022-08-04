package avail

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/0xPolygon/polygon-edge/consensus/ibft/proto"
	"github.com/0xPolygon/polygon-edge/crypto"
	"github.com/0xPolygon/polygon-edge/helper/keccak"
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/umbracle/fastrlp"
)

func writeSeal(prv *ecdsa.PrivateKey, h *types.Header) (*types.Header, error) {
	h = h.Copy()
	seal, err := signSealImpl(prv, h, false)

	if err != nil {
		return nil, err
	}

	extra, err := getValidatorExtra(h)
	if err != nil {
		return nil, err
	}

	extra.Seal = seal
	if err := PutValidatorExtra(h, extra); err != nil {
		return nil, err
	}

	return h, nil
}

func signSealImpl(prv *ecdsa.PrivateKey, h *types.Header, committed bool) ([]byte, error) {
	hash, err := calculateHeaderHash(h)
	if err != nil {
		return nil, err
	}

	// if we are singing the committed seals we need to do something more
	msg := hash
	if committed {
		msg = commitMsg(hash)
	}

	seal, err := crypto.Sign(prv, crypto.Keccak256(msg))

	if err != nil {
		return nil, err
	}

	return seal, nil
}

func calculateHeaderHash(h *types.Header) ([]byte, error) {
	h = h.Copy() // make a copy since we update the extra field

	arena := fastrlp.DefaultArenaPool.Get()
	defer fastrlp.DefaultArenaPool.Put(arena)

	// when hashing the block for signing we have to remove from
	// the extra field the seal and committed seal items
	extra, err := getValidatorExtra(h)
	if err != nil {
		return nil, err
	}

	// This will effectively remove the Seal and Committed Seal fields,
	// while keeping proposer vanity and validator set
	// because extra.Validators is what we got from `h` in the first place.
	assignExtraValidators(h, extra.Validators)

	fmt.Printf("Got the header extras: %+v \n", extra)

	vv := arena.NewArray()
	vv.Set(arena.NewBytes(h.ParentHash.Bytes()))
	vv.Set(arena.NewBytes(h.Sha3Uncles.Bytes()))
	vv.Set(arena.NewBytes(h.Miner.Bytes()))
	vv.Set(arena.NewBytes(h.StateRoot.Bytes()))
	vv.Set(arena.NewBytes(h.TxRoot.Bytes()))
	vv.Set(arena.NewBytes(h.ReceiptsRoot.Bytes()))
	vv.Set(arena.NewBytes(h.LogsBloom[:]))
	vv.Set(arena.NewUint(h.Difficulty))
	vv.Set(arena.NewUint(h.Number))
	vv.Set(arena.NewUint(h.GasLimit))
	vv.Set(arena.NewUint(h.GasUsed))
	vv.Set(arena.NewUint(h.Timestamp))
	vv.Set(arena.NewCopyBytes(h.ExtraData))

	buf := keccak.Keccak256Rlp(nil, vv)

	return buf, nil
}

func commitMsg(b []byte) []byte {
	// message that the nodes need to sign to commit to a block
	// hash with COMMIT_MSG_CODE which is the same value used in quorum
	return crypto.Keccak256(b, []byte{byte(proto.MessageReq_Commit)})
}

func addressRecoverFromHeader(h *types.Header) (types.Address, error) {
	// get the extra part that contains the seal
	extra, err := getValidatorExtra(h)
	if err != nil {
		return types.Address{}, err
	}

	// get the sig
	msg, err := calculateHeaderHash(h)
	if err != nil {
		return types.Address{}, err
	}

	return addressRecoverImpl(extra.Seal, msg)
}

func addressRecoverImpl(sig, msg []byte) (types.Address, error) {
	pub, err := crypto.RecoverPubkey(sig, crypto.Keccak256(msg))
	if err != nil {
		return types.Address{}, err
	}

	return crypto.PubKeyToAddress(pub), nil
}
