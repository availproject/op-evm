// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package fraud

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// FraudMetaData contains all meta data concerning the Fraud contract.
var FraudMetaData = &bind.MetaData{
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"counter\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"get\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"i\",\"type\":\"uint256\"}],\"name\":\"set\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x608060405234801561001057600080fd5b5061002d61002261003260201b60201c565b61003a60201b60201c565b6100fe565b600033905090565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff169050816000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055508173ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff167f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e060405160405180910390a35050565b6105b38061010d6000396000f3fe608060405234801561001057600080fd5b50600436106100625760003560e01c806360fe47b11461006757806361bc221a146100835780636d4ce63c146100a1578063715018a6146100bf5780638da5cb5b146100c9578063f2fde38b146100e7575b600080fd5b610081600480360381019061007c9190610362565b610103565b005b61008b61010d565b604051610098919061039e565b60405180910390f35b6100a9610113565b6040516100b6919061039e565b60405180910390f35b6100c761011d565b005b6100d1610131565b6040516100de91906103fa565b60405180910390f35b61010160048036038101906100fc9190610441565b61015a565b005b8060018190555050565b60015481565b6000600154905090565b6101256101dd565b61012f600061025b565b565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff16905090565b6101626101dd565b600073ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff16036101d1576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016101c8906104f1565b60405180910390fd5b6101da8161025b565b50565b6101e561031f565b73ffffffffffffffffffffffffffffffffffffffff16610203610131565b73ffffffffffffffffffffffffffffffffffffffff1614610259576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016102509061055d565b60405180910390fd5b565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff169050816000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055508173ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff167f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e060405160405180910390a35050565b600033905090565b600080fd5b6000819050919050565b61033f8161032c565b811461034a57600080fd5b50565b60008135905061035c81610336565b92915050565b60006020828403121561037857610377610327565b5b60006103868482850161034d565b91505092915050565b6103988161032c565b82525050565b60006020820190506103b3600083018461038f565b92915050565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b60006103e4826103b9565b9050919050565b6103f4816103d9565b82525050565b600060208201905061040f60008301846103eb565b92915050565b61041e816103d9565b811461042957600080fd5b50565b60008135905061043b81610415565b92915050565b60006020828403121561045757610456610327565b5b60006104658482850161042c565b91505092915050565b600082825260208201905092915050565b7f4f776e61626c653a206e6577206f776e657220697320746865207a65726f206160008201527f6464726573730000000000000000000000000000000000000000000000000000602082015250565b60006104db60268361046e565b91506104e68261047f565b604082019050919050565b6000602082019050818103600083015261050a816104ce565b9050919050565b7f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572600082015250565b600061054760208361046e565b915061055282610511565b602082019050919050565b600060208201905081810360008301526105768161053a565b905091905056fea2646970667358221220d2099a16454dcfc10998b83c33f54e075ef01f139abba3d86d4882072a248d4064736f6c634300080f0033",
}

// FraudABI is the input ABI used to generate the binding from.
// Deprecated: Use FraudMetaData.ABI instead.
var FraudABI = FraudMetaData.ABI

// FraudBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use FraudMetaData.Bin instead.
var FraudBin = FraudMetaData.Bin

// DeployFraud deploys a new Ethereum contract, binding an instance of Fraud to it.
func DeployFraud(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Fraud, error) {
	parsed, err := FraudMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(FraudBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Fraud{FraudCaller: FraudCaller{contract: contract}, FraudTransactor: FraudTransactor{contract: contract}, FraudFilterer: FraudFilterer{contract: contract}}, nil
}

// Fraud is an auto generated Go binding around an Ethereum contract.
type Fraud struct {
	FraudCaller     // Read-only binding to the contract
	FraudTransactor // Write-only binding to the contract
	FraudFilterer   // Log filterer for contract events
}

// FraudCaller is an auto generated read-only Go binding around an Ethereum contract.
type FraudCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// FraudTransactor is an auto generated write-only Go binding around an Ethereum contract.
type FraudTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// FraudFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type FraudFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// FraudSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type FraudSession struct {
	Contract     *Fraud            // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// FraudCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type FraudCallerSession struct {
	Contract *FraudCaller  // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// FraudTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type FraudTransactorSession struct {
	Contract     *FraudTransactor  // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// FraudRaw is an auto generated low-level Go binding around an Ethereum contract.
type FraudRaw struct {
	Contract *Fraud // Generic contract binding to access the raw methods on
}

// FraudCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type FraudCallerRaw struct {
	Contract *FraudCaller // Generic read-only contract binding to access the raw methods on
}

// FraudTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type FraudTransactorRaw struct {
	Contract *FraudTransactor // Generic write-only contract binding to access the raw methods on
}

// NewFraud creates a new instance of Fraud, bound to a specific deployed contract.
func NewFraud(address common.Address, backend bind.ContractBackend) (*Fraud, error) {
	contract, err := bindFraud(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Fraud{FraudCaller: FraudCaller{contract: contract}, FraudTransactor: FraudTransactor{contract: contract}, FraudFilterer: FraudFilterer{contract: contract}}, nil
}

// NewFraudCaller creates a new read-only instance of Fraud, bound to a specific deployed contract.
func NewFraudCaller(address common.Address, caller bind.ContractCaller) (*FraudCaller, error) {
	contract, err := bindFraud(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &FraudCaller{contract: contract}, nil
}

// NewFraudTransactor creates a new write-only instance of Fraud, bound to a specific deployed contract.
func NewFraudTransactor(address common.Address, transactor bind.ContractTransactor) (*FraudTransactor, error) {
	contract, err := bindFraud(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &FraudTransactor{contract: contract}, nil
}

// NewFraudFilterer creates a new log filterer instance of Fraud, bound to a specific deployed contract.
func NewFraudFilterer(address common.Address, filterer bind.ContractFilterer) (*FraudFilterer, error) {
	contract, err := bindFraud(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &FraudFilterer{contract: contract}, nil
}

// bindFraud binds a generic wrapper to an already deployed contract.
func bindFraud(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(FraudABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Fraud *FraudRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Fraud.Contract.FraudCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Fraud *FraudRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Fraud.Contract.FraudTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Fraud *FraudRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Fraud.Contract.FraudTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Fraud *FraudCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Fraud.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Fraud *FraudTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Fraud.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Fraud *FraudTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Fraud.Contract.contract.Transact(opts, method, params...)
}

// Counter is a free data retrieval call binding the contract method 0x61bc221a.
//
// Solidity: function counter() view returns(uint256)
func (_Fraud *FraudCaller) Counter(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Fraud.contract.Call(opts, &out, "counter")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Counter is a free data retrieval call binding the contract method 0x61bc221a.
//
// Solidity: function counter() view returns(uint256)
func (_Fraud *FraudSession) Counter() (*big.Int, error) {
	return _Fraud.Contract.Counter(&_Fraud.CallOpts)
}

// Counter is a free data retrieval call binding the contract method 0x61bc221a.
//
// Solidity: function counter() view returns(uint256)
func (_Fraud *FraudCallerSession) Counter() (*big.Int, error) {
	return _Fraud.Contract.Counter(&_Fraud.CallOpts)
}

// Get is a free data retrieval call binding the contract method 0x6d4ce63c.
//
// Solidity: function get() view returns(uint256)
func (_Fraud *FraudCaller) Get(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Fraud.contract.Call(opts, &out, "get")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Get is a free data retrieval call binding the contract method 0x6d4ce63c.
//
// Solidity: function get() view returns(uint256)
func (_Fraud *FraudSession) Get() (*big.Int, error) {
	return _Fraud.Contract.Get(&_Fraud.CallOpts)
}

// Get is a free data retrieval call binding the contract method 0x6d4ce63c.
//
// Solidity: function get() view returns(uint256)
func (_Fraud *FraudCallerSession) Get() (*big.Int, error) {
	return _Fraud.Contract.Get(&_Fraud.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Fraud *FraudCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Fraud.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Fraud *FraudSession) Owner() (common.Address, error) {
	return _Fraud.Contract.Owner(&_Fraud.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Fraud *FraudCallerSession) Owner() (common.Address, error) {
	return _Fraud.Contract.Owner(&_Fraud.CallOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Fraud *FraudTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Fraud.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Fraud *FraudSession) RenounceOwnership() (*types.Transaction, error) {
	return _Fraud.Contract.RenounceOwnership(&_Fraud.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Fraud *FraudTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _Fraud.Contract.RenounceOwnership(&_Fraud.TransactOpts)
}

// Set is a paid mutator transaction binding the contract method 0x60fe47b1.
//
// Solidity: function set(uint256 i) returns()
func (_Fraud *FraudTransactor) Set(opts *bind.TransactOpts, i *big.Int) (*types.Transaction, error) {
	return _Fraud.contract.Transact(opts, "set", i)
}

// Set is a paid mutator transaction binding the contract method 0x60fe47b1.
//
// Solidity: function set(uint256 i) returns()
func (_Fraud *FraudSession) Set(i *big.Int) (*types.Transaction, error) {
	return _Fraud.Contract.Set(&_Fraud.TransactOpts, i)
}

// Set is a paid mutator transaction binding the contract method 0x60fe47b1.
//
// Solidity: function set(uint256 i) returns()
func (_Fraud *FraudTransactorSession) Set(i *big.Int) (*types.Transaction, error) {
	return _Fraud.Contract.Set(&_Fraud.TransactOpts, i)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Fraud *FraudTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _Fraud.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Fraud *FraudSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Fraud.Contract.TransferOwnership(&_Fraud.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Fraud *FraudTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Fraud.Contract.TransferOwnership(&_Fraud.TransactOpts, newOwner)
}

// FraudOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the Fraud contract.
type FraudOwnershipTransferredIterator struct {
	Event *FraudOwnershipTransferred // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *FraudOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(FraudOwnershipTransferred)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(FraudOwnershipTransferred)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *FraudOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *FraudOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// FraudOwnershipTransferred represents a OwnershipTransferred event raised by the Fraud contract.
type FraudOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Fraud *FraudFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*FraudOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Fraud.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &FraudOwnershipTransferredIterator{contract: _Fraud.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Fraud *FraudFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *FraudOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Fraud.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(FraudOwnershipTransferred)
				if err := _Fraud.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Fraud *FraudFilterer) ParseOwnershipTransferred(log types.Log) (*FraudOwnershipTransferred, error) {
	event := new(FraudOwnershipTransferred)
	if err := _Fraud.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
