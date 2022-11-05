package staking

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/0xPolygon/polygon-edge/types"
)

type staticActiveSequencers struct {
	sequencers []types.Address
}

func (sas *staticActiveSequencers) Get() ([]types.Address, error) {
	lst := make([]types.Address, len(sas.sequencers))
	copy(lst, sas.sequencers)
	return lst, nil
}

func (sas *staticActiveSequencers) Contains(addr types.Address) (bool, error) {
	for _, a := range sas.sequencers {
		if bytes.Equal(a.Bytes(), addr.Bytes()) {
			return true, nil
		}
	}

	return false, nil
}

func Test_RandomizedSequencers(t *testing.T) {

	testCases := []struct {
		name               string
		inputSequencers    []types.Address
		expectedSequencers []types.Address
	}{
		{
			name:               "no sequencers",
			inputSequencers:    []types.Address{},
			expectedSequencers: []types.Address{},
		},
		{
			name: "one sequencer",
			inputSequencers: []types.Address{
				types.StringToAddress("0xAFF12c2B1df7D56144B3CbeDfb64B48d4F018D89"),
			},
			expectedSequencers: []types.Address{
				types.StringToAddress("0xAFF12c2B1df7D56144B3CbeDfb64B48d4F018D89"),
			},
		},
		{
			name: "many sequencers",
			inputSequencers: []types.Address{
				types.StringToAddress("0xAFF12c2B1df7D56144B3CbeDfb64B48d4F018D89"),
				types.StringToAddress("0xC006b2443A1A61d7a1780B81Dbf7A591ceA0b2A0"),
				types.StringToAddress("0x8C037E6dA0A0ACfC2E38A5e046d3dB9EBD2b4Fcc"),
				types.StringToAddress("0x40d170ea21c9477B8360D86CC3C2Baa0D9a9A438"),
				types.StringToAddress("0xA02dE14E4cD7715F17F1bfdCe9370Ac0e82bd4C9"),
				types.StringToAddress("0xB168E70F55e48072b4e01e2ceF9E3B775397c36b"),
				types.StringToAddress("0x412fB31F5C978F6b9268F8431F2a4781B901A9DE"),
				types.StringToAddress("0xec44B3e4C9C8bC64e81120Dc22E4BAB13458D54F"),
				types.StringToAddress("0xc2e6C4379C015a4587826B576ADa0AE73500a9c5"),
				types.StringToAddress("0x11444390fE555f166E44CD8Da9A60f295B4aEB42"),
				types.StringToAddress("0xC26Bd93B02c692F2f42cd44E0647634684C3eeEd"),
				types.StringToAddress("0xd18b41207268C68B6B87C8f01e7e377b4bc3d87c"),
				types.StringToAddress("0x8Fd73b20bb5F59ff0f4ec3fA344d07507B879496"),
				types.StringToAddress("0x2E215bC6b87cEB384cf519C489Eac547da9A2d8E"),
				types.StringToAddress("0xf89204d5d88B435A37B18a068CC33292DBeE791c"),
				types.StringToAddress("0x8f11dA8704c3F44047AF8090488f585b5D03A982"),
				types.StringToAddress("0x20e6E89dAf166D929ccD459fefB51893a9dE4E00"),
				types.StringToAddress("0xa9a7069dDd90aA65723386bE1C81eC5155e82a13"),
				types.StringToAddress("0x64733DF2421D59E0BA156c9b86Ee6493fDd66F3F"),
				types.StringToAddress("0xd7c04ab2fE521E868a619190FD8755aAff402fd7"),
			},
			expectedSequencers: []types.Address{
				types.StringToAddress("0x8C037E6dA0A0ACfC2E38A5e046d3dB9EBD2b4Fcc"),
				types.StringToAddress("0x8Fd73b20bb5F59ff0f4ec3fA344d07507B879496"),
				types.StringToAddress("0xC26Bd93B02c692F2f42cd44E0647634684C3eeEd"),
				types.StringToAddress("0xC006b2443A1A61d7a1780B81Dbf7A591ceA0b2A0"),
				types.StringToAddress("0xA02dE14E4cD7715F17F1bfdCe9370Ac0e82bd4C9"),
				types.StringToAddress("0xd18b41207268C68B6B87C8f01e7e377b4bc3d87c"),
				types.StringToAddress("0xc2e6C4379C015a4587826B576ADa0AE73500a9c5"),
				types.StringToAddress("0x2E215bC6b87cEB384cf519C489Eac547da9A2d8E"),
				types.StringToAddress("0xec44B3e4C9C8bC64e81120Dc22E4BAB13458D54F"),
				types.StringToAddress("0xd7c04ab2fE521E868a619190FD8755aAff402fd7"),
				types.StringToAddress("0xf89204d5d88B435A37B18a068CC33292DBeE791c"),
				types.StringToAddress("0xB168E70F55e48072b4e01e2ceF9E3B775397c36b"),
				types.StringToAddress("0x412fB31F5C978F6b9268F8431F2a4781B901A9DE"),
				types.StringToAddress("0xAFF12c2B1df7D56144B3CbeDfb64B48d4F018D89"),
				types.StringToAddress("0x64733DF2421D59E0BA156c9b86Ee6493fDd66F3F"),
				types.StringToAddress("0x11444390fE555f166E44CD8Da9A60f295B4aEB42"),
				types.StringToAddress("0x40d170ea21c9477B8360D86CC3C2Baa0D9a9A438"),
				types.StringToAddress("0xa9a7069dDd90aA65723386bE1C81eC5155e82a13"),
				types.StringToAddress("0x20e6E89dAf166D929ccD459fefB51893a9dE4E00"),
				types.StringToAddress("0x8f11dA8704c3F44047AF8090488f585b5D03A982"),
			},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("case %d: %s", i, tc.name), func(t *testing.T) {
			sqs := NewRandomizedActiveSequencersQuerier(func() int64 { return 42 }, &staticActiveSequencers{tc.inputSequencers})

			sequencers, err := sqs.Get()
			if err != nil {
				t.Fatal(err)
			}

			if len(sequencers) != len(tc.inputSequencers) {
				t.Fatalf("expected %d sequencers, got %d", len(tc.inputSequencers), len(sequencers))
			}

			for i, s := range sequencers {
				if !bytes.Equal(s.Bytes(), tc.expectedSequencers[i].Bytes()) {
					t.Fatalf("got address %q at index %d, expected %q", s.String(), i, tc.expectedSequencers[i].String())
				}
			}
		})
	}
}
