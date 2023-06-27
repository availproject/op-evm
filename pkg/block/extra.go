package block

import (
	"errors"
	"fmt"
	"sort"

	"github.com/0xPolygon/polygon-edge/types"
	"github.com/umbracle/fastrlp"
)

const (
	// KeyExtraValidators is key that identifies the `ValidatorsExtra` object
	// serialized in `ExtraData`.
	KeyExtraValidators = "EXTRA_VALIDATORS"

	// KeyFraudProofOf is key that identifies the fraudproof objected malicious
	// block hash in `ExtraData` of the fraudproof block header.
	KeyFraudProofOf = "FRAUD_PROOF_OF"

	// KeyBeginDisputeResolutionOf used to understand which tx from the txpool we need to pick
	// when writing fraud slash block
	KeyBeginDisputeResolutionOf = "BEGIN_DISPUTE_RESOLUTION_OF"

	// KeyEndDisputeResolutionOf used to understand which block hash was used to slash the node
	// in order to end dispute resolution on all of the nodes
	KeyEndDisputeResolutionOf = "END_DISPUTE_RESOLUTION_OF"
)

// EncodeExtraDataFields encodes the given map of extra data fields into a byte slice.
// It takes a map of string keys to byte slice values and returns a byte slice representation of the encoded data.
func EncodeExtraDataFields(data map[string][]byte) []byte {
	var keys []string
	for k := range data {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	a := &fastrlp.Arena{}

	vv := a.NewArray()

	for _, k := range keys {
		vv.Set(a.NewString(k))
		vv.Set(a.NewBytes(data[k]))
	}

	return vv.MarshalTo(nil)
}

// DecodeExtraDataFields decodes the byte slice into a map of extra data fields.
// It takes a byte slice representing the encoded data and returns a map of string keys to byte slice values.
func DecodeExtraDataFields(data []byte) (map[string][]byte, error) {
	kv := make(map[string][]byte)

	if len(data) == 0 {
		return kv, nil
	}

	p := &fastrlp.Parser{}
	v, err := p.Parse(data)
	if err != nil {
		return nil, err
	}

	vs, err := v.GetElems()
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(vs); i += 2 {
		k, err := vs[i].GetString()
		if err != nil {
			return nil, err
		}

		v, err := vs[i+1].Bytes()
		if err != nil {
			return nil, err
		}

		kv[k] = v
	}

	return kv, nil
}

// AssignExtraValidators adds the validators to the extra data field in the header.
// It takes the header and a slice of validator addresses and modifies the header's extra data field accordingly.
// Returns an error if there is an issue decoding or encoding the extra data field.
func AssignExtraValidators(h *types.Header, validators []types.Address) error {
	kv, err := DecodeExtraDataFields(h.ExtraData)
	if err != nil {
		return err
	}

	ibftExtra := &ValidatorExtra{
		Validators:    validators,
		Seal:          []byte{},
		CommittedSeal: [][]byte{},
	}

	bs := ibftExtra.MarshalRLPTo(nil)
	kv[KeyExtraValidators] = bs

	h.ExtraData = EncodeExtraDataFields(kv)

	return nil
}

// PutValidatorExtra sets the extra data field in the header to the given ValidatorExtra.
// It takes the header and a ValidatorExtra struct and modifies the header's extra data field accordingly.
// Returns an error if there is an issue decoding or encoding the extra data field.
func PutValidatorExtra(h *types.Header, istanbulExtra *ValidatorExtra) error {
	kv, err := DecodeExtraDataFields(h.ExtraData)
	if err != nil {
		return err
	}

	data := istanbulExtra.MarshalRLPTo(nil)

	kv[KeyExtraValidators] = data

	h.ExtraData = EncodeExtraDataFields(kv)

	return nil
}

// getValidatorExtra returns the ValidatorExtra from the extra data field in the header.
// It takes the header and returns the ValidatorExtra struct decoded from the header's extra data field.
// Returns an error if there is an issue decoding the extra data field or if the extra data field is not found.
func getValidatorExtra(h *types.Header) (*ValidatorExtra, error) {
	kv, err := DecodeExtraDataFields(h.ExtraData)
	if err != nil {
		return nil, err
	}

	data, exists := kv[KeyExtraValidators]
	if !exists {
		return nil, errors.New("no validators extra object found")
	}

	extra := &ValidatorExtra{}

	if err := extra.UnmarshalRLP(data); err != nil {
		return nil, err
	}

	return extra, nil
}

// GetExtraDataFraudProofTarget returns the fraudproof target from the extra data field in the header.
// It takes the header and returns the fraudproof target as a Hash value.
// Returns the fraudproof target and a boolean indicating if it was found in the extra data field.
func GetExtraDataFraudProofTarget(h *types.Header) (types.Hash, bool) {
	kv, err := DecodeExtraDataFields(h.ExtraData)
	if err != nil {
		return types.ZeroHash, false
	}

	data, exists := kv[KeyFraudProofOf]
	if !exists {
		return types.ZeroHash, false
	}

	toReturn := types.BytesToHash(data)

	if toReturn == types.ZeroHash {
		return types.ZeroHash, false
	}

	return toReturn, true
}

// GetExtraDataBeginDisputeResolutionTarget returns the begin dispute resolution target from the extra data field in the header.
// It takes the header and returns the begin dispute resolution target as a Hash value.
// Returns the begin dispute resolution target and a boolean indicating if it was found in the extra data field.
func GetExtraDataBeginDisputeResolutionTarget(h *types.Header) (types.Hash, bool) {
	kv, err := DecodeExtraDataFields(h.ExtraData)
	if err != nil {
		return types.ZeroHash, false
	}

	data, exists := kv[KeyBeginDisputeResolutionOf]
	if !exists {
		return types.ZeroHash, false
	}

	toReturn := types.BytesToHash(data)

	if toReturn == types.ZeroHash {
		return types.ZeroHash, false
	}

	return toReturn, true
}

// GetExtraDataEndDisputeResolutionTarget returns the end dispute resolution target from the extra data field in the header.
// It takes the header and returns the end dispute resolution target as a Hash value.
// Returns the end dispute resolution target and a boolean indicating if it was found in the extra data field.
func GetExtraDataEndDisputeResolutionTarget(h *types.Header) (types.Hash, bool) {
	kv, err := DecodeExtraDataFields(h.ExtraData)
	if err != nil {
		return types.ZeroHash, false
	}

	data, exists := kv[KeyEndDisputeResolutionOf]
	if !exists {
		return types.ZeroHash, false
	}

	toReturn := types.BytesToHash(data)

	if toReturn == types.ZeroHash {
		return types.ZeroHash, false
	}

	return toReturn, true
}

// ValidatorExtra defines the structure of the extra data field for validators.
type ValidatorExtra struct {
	Validators    []types.Address
	Seal          []byte
	CommittedSeal [][]byte
}

// MarshalRLPTo marshals the ValidatorExtra struct to an RLP-encoded byte slice.
// It takes the ValidatorExtra struct and a destination byte slice, and returns the marshaled byte slice.
func (i *ValidatorExtra) MarshalRLPTo(dst []byte) []byte {
	return types.MarshalRLPTo(i.MarshalRLPWith, dst)
}

// MarshalRLPWith marshals the ValidatorExtra struct to an RLP value using the given RLP arena.
// It takes the ValidatorExtra struct and an RLP arena, and returns the RLP value.
func (i *ValidatorExtra) MarshalRLPWith(ar *fastrlp.Arena) *fastrlp.Value {
	vv := ar.NewArray()

	// Validators
	vals := ar.NewArray()
	for _, a := range i.Validators {
		vals.Set(ar.NewBytes(a.Bytes()))
	}

	vv.Set(vals)

	// Seal
	if len(i.Seal) == 0 {
		vv.Set(ar.NewNull())
	} else {
		vv.Set(ar.NewBytes(i.Seal))
	}

	// CommittedSeal
	if len(i.CommittedSeal) == 0 {
		vv.Set(ar.NewNullArray())
	} else {
		committed := ar.NewArray()
		for _, a := range i.CommittedSeal {
			if len(a) == 0 {
				vv.Set(ar.NewNull())
			} else {
				committed.Set(ar.NewBytes(a))
			}
		}
		vv.Set(committed)
	}

	return vv
}

// UnmarshalRLP unmarshals the ValidatorExtra struct from an RLP-encoded byte slice.
// It takes the input byte slice and returns an error if there is an issue unmarshaling the data.
func (i *ValidatorExtra) UnmarshalRLP(input []byte) error {
	return types.UnmarshalRlp(i.UnmarshalRLPFrom, input)
}

// UnmarshalRLPFrom unmarshals the ValidatorExtra struct from an RLP value using the given RLP parser and value.
// It takes the RLP parser, RLP value, and returns an error if there is an issue unmarshaling the data.
func (i *ValidatorExtra) UnmarshalRLPFrom(p *fastrlp.Parser, v *fastrlp.Value) error {
	elems, err := v.GetElems()
	if err != nil {
		return err
	}

	if len(elems) < 3 {
		return fmt.Errorf("incorrect number of elements to decode istambul extra, expected 3 but found %d", len(elems))
	}

	// Validators
	{
		vals, err := elems[0].GetElems()
		if err != nil {
			return fmt.Errorf("list expected for validators")
		}
		i.Validators = make([]types.Address, len(vals))
		for indx, val := range vals {
			if err = val.GetAddr(i.Validators[indx][:]); err != nil {
				return err
			}
		}
	}

	// Seal
	{
		if i.Seal, err = elems[1].GetBytes(i.Seal); err != nil {
			return err
		}
	}

	// Committed
	{
		vals, err := elems[2].GetElems()
		if err != nil {
			return fmt.Errorf("list expected for committed")
		}
		i.CommittedSeal = make([][]byte, len(vals))
		for indx, val := range vals {
			if i.CommittedSeal[indx], err = val.GetBytes(i.CommittedSeal[indx]); err != nil {
				return err
			}
		}
	}

	return nil
}
