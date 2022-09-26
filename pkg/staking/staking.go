package staking

import (
	"fmt"
	"math/big"

	"github.com/0xPolygon/polygon-edge/helper/common"

	"github.com/0xPolygon/polygon-edge/chain"
	"github.com/0xPolygon/polygon-edge/helper/hex"
	"github.com/0xPolygon/polygon-edge/helper/keccak"
	"github.com/0xPolygon/polygon-edge/types"
	stakingContract "github.com/maticnetwork/avail-settlement/contracts/staking"
)

var (
	// staking contract address
	AddrStakingContract = types.StringToAddress("0x0110000000000000000000000000000000000001")

	MinSequencerCount = uint64(1)
	MaxSequencerCount = common.MaxSafeJSInt
)

// getAddressMapping returns the key for the SC storage mapping (address => something)
//
// More information:
// https://docs.soliditylang.org/en/latest/internals/layout_in_storage.html
func getAddressMapping(address types.Address, slot int64) []byte {
	bigSlot := big.NewInt(slot)

	finalSlice := append(
		common.PadLeftOrTrim(address.Bytes(), 32),
		common.PadLeftOrTrim(bigSlot.Bytes(), 32)...,
	)
	keccakValue := keccak.Keccak256(nil, finalSlice)

	return keccakValue
}

// getIndexWithOffset is a helper method for adding an offset to the already found keccak hash
func getIndexWithOffset(keccakHash []byte, offset int64) []byte {
	bigOffset := big.NewInt(offset)
	bigKeccak := big.NewInt(0).SetBytes(keccakHash)

	bigKeccak.Add(bigKeccak, bigOffset)

	return bigKeccak.Bytes()
}

// getStorageIndexes is a helper function for getting the correct indexes
// of the storage slots which need to be modified during bootstrap.
//
// It is SC dependant, and based on the SC located at:
// https://github.com/0xPolygon/staking-contracts/
func getStorageIndexes(address types.Address, index int64) *StorageIndexes {
	storageIndexes := StorageIndexes{}

	// Get the indexes for the mappings
	// The index for the mapping is retrieved with:
	// keccak(address . slot)
	// . stands for concatenation (basically appending the bytes)
	storageIndexes.AddressToIsSequencerIndex = getAddressMapping(address, addressToIsSequencerSlot)
	storageIndexes.AddressToStakedAmountIndex = getAddressMapping(address, addressToStakedAmountSlot)
	storageIndexes.AddressToSequencerIndexIndex = getAddressMapping(address, addressToSequencerIndexSlot)

	// Get the indexes for _sequencers, _stakedAmount
	// Index for regular types is calculated as just the regular slot
	storageIndexes.StakedAmountIndex = big.NewInt(stakedAmountSlot).Bytes()

	// Index for array types is calculated as keccak(slot) + index
	// The slot for the dynamic arrays that's put in the keccak needs to be in hex form (padded 64 chars)
	storageIndexes.SequencersIndex = getIndexWithOffset(
		keccak.Keccak256(nil, common.PadLeftOrTrim(big.NewInt(sequencersSlot).Bytes(), 32)),
		index,
	)

	// For any dynamic array in Solidity, the size of the actual array should be
	// located on slot x
	storageIndexes.SequencersArraySizeIndex = []byte{byte(sequencersSlot)}

	return &storageIndexes
}

// PredeployParams contains the values used to predeploy the staking contract
type PredeployParams struct {
	MinSequencerCount uint64
	MaxSequencerCount uint64
}

// StorageIndexes is a wrapper for different storage indexes that
// need to be modified
type StorageIndexes struct {
	SequencersIndex              []byte // []address
	SequencersArraySizeIndex     []byte // []address size
	AddressToIsSequencerIndex    []byte // mapping(address => bool)
	AddressToStakedAmountIndex   []byte // mapping(address => uint256)
	AddressToSequencerIndexIndex []byte // mapping(address => uint256)
	StakedAmountIndex            []byte // uint256
}

// Slot definitions for SC storage
var (
	sequencersSlot              = int64(0) // Slot 0
	addressToIsSequencerSlot    = int64(1) // Slot 1
	addressToStakedAmountSlot   = int64(2) // Slot 2
	addressToSequencerIndexSlot = int64(3) // Slot 3
	stakedAmountSlot            = int64(4) // Slot 4
	minNumSequencerSlot         = int64(5) // Slot 5
	maxNumSequencerSlot         = int64(6) // Slot 6
)

const (
	DefaultStakedBalance = "0x8AC7230489E80000" // 10 MATIC
)

// PredeployStakingSC is a helper method for setting up the staking smart contract account,
// using the passed in sequencers as pre-staked sequencers
func PredeployStakingSC(
	sequencers []types.Address,
	params PredeployParams,
) (*chain.GenesisAccount, error) {
	// Set the code for the staking smart contract
	scHex, _ := hex.DecodeHex(stakingContract.StakingMetaData.Bin)
	stakingAccount := &chain.GenesisAccount{
		Code: scHex,
	}

	// Parse the default staked balance value into *big.Int
	val := DefaultStakedBalance
	bigDefaultStakedBalance, err := types.ParseUint256orHex(&val)

	if err != nil {
		return nil, fmt.Errorf("unable to generate DefaultStatkedBalance, %w", err)
	}

	// Generate the empty account storage map
	storageMap := make(map[types.Hash]types.Hash)
	bigTrueValue := big.NewInt(1)
	stakedAmount := big.NewInt(0)
	bigMinNumSequencers := big.NewInt(int64(params.MinSequencerCount))
	bigMaxNumSequencers := big.NewInt(int64(params.MaxSequencerCount))

	for indx, sequencer := range sequencers {
		// Update the total staked amount
		stakedAmount.Add(stakedAmount, bigDefaultStakedBalance)

		// Get the storage indexes
		storageIndexes := getStorageIndexes(sequencer, int64(indx))

		// Set the value for the sequencers array
		storageMap[types.BytesToHash(storageIndexes.SequencersIndex)] =
			types.BytesToHash(
				sequencer.Bytes(),
			)

		// Set the value for the address -> sequencer array index mapping
		storageMap[types.BytesToHash(storageIndexes.AddressToIsSequencerIndex)] =
			types.BytesToHash(bigTrueValue.Bytes())

		// Set the value for the address -> staked amount mapping
		storageMap[types.BytesToHash(storageIndexes.AddressToStakedAmountIndex)] =
			types.StringToHash(hex.EncodeBig(bigDefaultStakedBalance))

		// Set the value for the address -> sequencer index mapping
		storageMap[types.BytesToHash(storageIndexes.AddressToSequencerIndexIndex)] =
			types.StringToHash(hex.EncodeUint64(uint64(indx)))

		// Set the value for the total staked amount
		storageMap[types.BytesToHash(storageIndexes.StakedAmountIndex)] =
			types.BytesToHash(stakedAmount.Bytes())

		// Set the value for the size of the sequencers array
		storageMap[types.BytesToHash(storageIndexes.SequencersArraySizeIndex)] =
			types.StringToHash(hex.EncodeUint64(uint64(indx + 1)))
	}

	// Set the value for the minimum number of sequencers
	storageMap[types.BytesToHash(big.NewInt(minNumSequencerSlot).Bytes())] =
		types.BytesToHash(bigMinNumSequencers.Bytes())

	// Set the value for the maximum number of sequencers
	storageMap[types.BytesToHash(big.NewInt(maxNumSequencerSlot).Bytes())] =
		types.BytesToHash(bigMaxNumSequencers.Bytes())

	// Save the storage map
	stakingAccount.Storage = storageMap

	// Set the Staking SC balance to numSequencers * defaultStakedBalance
	stakingAccount.Balance = stakedAmount

	return stakingAccount, nil
}
