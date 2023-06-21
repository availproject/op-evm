package block

import (
	"crypto/ecdsa"

	"github.com/0xPolygon/polygon-edge/crypto"
	"github.com/0xPolygon/polygon-edge/helper/keccak"
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/umbracle/fastrlp"
)

// WriteSeal signs the block and writes serialized `ValidatorExtra` into the block's `ExtraData`.
// It takes the private key and the header, and returns the updated header with the seal and an error if there is an issue.
func WriteSeal(prv *ecdsa.PrivateKey, h *types.Header) (*types.Header, error) {
	h = h.Copy()
	seal, err := signSealImpl(prv, h)

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

// signSealImpl signs the header using the private key and returns the seal.
// It takes the private key and the header, and returns the seal as a byte slice and an error if there is an issue.
func signSealImpl(prv *ecdsa.PrivateKey, h *types.Header) ([]byte, error) {
	hash, err := calculateHeaderHash(h)
	if err != nil {
		return nil, err
	}

	msg := hash
	seal, err := crypto.Sign(prv, crypto.Keccak256(msg))

	if err != nil {
		return nil, err
	}

	return seal, nil
}

// calculateHeaderHash calculates the hash of the header.
// It takes the header and returns the hash as a byte slice and an error if there is an issue.
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
	err = AssignExtraValidators(h, extra.Validators)
	if err != nil {
		return nil, err
	}

	vv := arena.NewArray()
	vv.Set(arena.NewBytes(h.ParentHash.Bytes()))
	vv.Set(arena.NewBytes(h.Sha3Uncles.Bytes()))
	vv.Set(arena.NewBytes(h.Miner))
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

// AddressRecoverFromHeader recovers the address from the header seal.
// It takes the header and returns the recovered address and an error if there is an issue.
func AddressRecoverFromHeader(h *types.Header) (types.Address, error) {
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

// addressRecoverImpl recovers the address from the signature and message.
// It takes the signature as a byte slice, the message as a byte slice, and returns the recovered address and an error if there is an issue.
func addressRecoverImpl(sig, msg []byte) (types.Address, error) {
	pub, err := crypto.RecoverPubkey(sig, crypto.Keccak256(msg))
	if err != nil {
		return types.Address{}, err
	}

	return crypto.PubKeyToAddress(pub), nil
}
