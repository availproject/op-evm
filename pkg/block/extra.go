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
	KeyFraudProofOf = "FRAUDPROOF_OF"

	KeyBeginDisputeResolutionOf = "BEGINDISPUTERESOLUTION_OF"
)

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

// AssignExtraValidators is a helper method that adds validators to the extra field in the header
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

// PutIbftExtra sets the extra data field in the header to the passed in istanbul extra data
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

// getValidatorExtra returns the istanbul extra data field from the passed in header
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

func GetExtraDataBeginDisputeResolutionTarget(h *types.Header) (types.Address, bool) {
	kv, err := DecodeExtraDataFields(h.ExtraData)
	if err != nil {
		return types.ZeroAddress, false
	}

	data, exists := kv[KeyBeginDisputeResolutionOf]
	if !exists {
		return types.ZeroAddress, false
	}

	toReturn := types.BytesToAddress(data)

	if toReturn == types.ZeroAddress {
		return types.ZeroAddress, false
	}

	return toReturn, true
}

// IstanbulExtra defines the structure of the extra field for Istanbul
type ValidatorExtra struct {
	Validators    []types.Address
	Seal          []byte
	CommittedSeal [][]byte
}

// MarshalRLPTo defines the marshal function wrapper for ValidatorExtra
func (i *ValidatorExtra) MarshalRLPTo(dst []byte) []byte {
	return types.MarshalRLPTo(i.MarshalRLPWith, dst)
}

// MarshalRLPWith defines the marshal function implementation for ValidatorExtra
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

// UnmarshalRLP defines the unmarshal function wrapper for ValidatorExtra
func (i *ValidatorExtra) UnmarshalRLP(input []byte) error {
	return types.UnmarshalRlp(i.UnmarshalRLPFrom, input)
}

// UnmarshalRLPFrom defines the unmarshal implementation for ValidatorExtra
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
