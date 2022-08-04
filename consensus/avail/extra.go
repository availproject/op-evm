package avail

import (
	"fmt"

	"github.com/0xPolygon/polygon-edge/types"
	"github.com/umbracle/fastrlp"
)

var (
	// ValidatorExtraVanity represents a fixed number of extra-data bytes reserved for proposer vanity
	ValidatorExtraVanity = 32
)

var zeroBytes = make([]byte, 32)

// assignExtraValidators is a helper method that adds validators to the extra field in the header
func assignExtraValidators(h *types.Header, validators []types.Address) {
	// Pad zeros to the right up to istanbul vanity
	extra := h.ExtraData
	if len(extra) < ValidatorExtraVanity {
		extra = append(extra, zeroBytes[:ValidatorExtraVanity-len(extra)]...)
	} else {
		extra = extra[:ValidatorExtraVanity]
	}

	ibftExtra := &ValidatorExtra{
		Validators:    validators,
		Seal:          []byte{},
		CommittedSeal: [][]byte{},
	}

	extra = ibftExtra.MarshalRLPTo(extra)
	h.ExtraData = extra
}

// PutIbftExtra sets the extra data field in the header to the passed in istanbul extra data
func PutValidatorExtra(h *types.Header, istanbulExtra *ValidatorExtra) error {
	// Pad zeros to the right up to istanbul vanity
	extra := h.ExtraData
	if len(extra) < ValidatorExtraVanity {
		extra = append(extra, zeroBytes[:ValidatorExtraVanity-len(extra)]...)
	} else {
		extra = extra[:ValidatorExtraVanity]
	}

	data := istanbulExtra.MarshalRLPTo(nil)
	extra = append(extra, data...)
	h.ExtraData = extra

	return nil
}

// getValidatorExtra returns the istanbul extra data field from the passed in header
func getValidatorExtra(h *types.Header) (*ValidatorExtra, error) {
	if len(h.ExtraData) < ValidatorExtraVanity {
		return nil, fmt.Errorf("wrong extra size: %d", len(h.ExtraData))
	}

	data := h.ExtraData[ValidatorExtraVanity:]
	extra := &ValidatorExtra{}

	if err := extra.UnmarshalRLP(data); err != nil {
		return nil, err
	}

	return extra, nil
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
