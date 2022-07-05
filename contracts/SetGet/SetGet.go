// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package setget

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

// SetgetMetaData contains all meta data concerning the Setget contract.
var SetgetMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"name\":\"counter\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"get\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"i\",\"type\":\"uint256\"}],\"name\":\"set\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// SetgetABI is the input ABI used to generate the binding from.
// Deprecated: Use SetgetMetaData.ABI instead.
var SetgetABI = SetgetMetaData.ABI

// Setget is an auto generated Go binding around an Ethereum contract.
type Setget struct {
	SetgetCaller     // Read-only binding to the contract
	SetgetTransactor // Write-only binding to the contract
	SetgetFilterer   // Log filterer for contract events
}

// SetgetCaller is an auto generated read-only Go binding around an Ethereum contract.
type SetgetCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SetgetTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SetgetTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SetgetFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SetgetFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SetgetSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SetgetSession struct {
	Contract     *Setget           // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SetgetCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SetgetCallerSession struct {
	Contract *SetgetCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// SetgetTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SetgetTransactorSession struct {
	Contract     *SetgetTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SetgetRaw is an auto generated low-level Go binding around an Ethereum contract.
type SetgetRaw struct {
	Contract *Setget // Generic contract binding to access the raw methods on
}

// SetgetCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SetgetCallerRaw struct {
	Contract *SetgetCaller // Generic read-only contract binding to access the raw methods on
}

// SetgetTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SetgetTransactorRaw struct {
	Contract *SetgetTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSetget creates a new instance of Setget, bound to a specific deployed contract.
func NewSetget(address common.Address, backend bind.ContractBackend) (*Setget, error) {
	contract, err := bindSetget(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Setget{SetgetCaller: SetgetCaller{contract: contract}, SetgetTransactor: SetgetTransactor{contract: contract}, SetgetFilterer: SetgetFilterer{contract: contract}}, nil
}

// NewSetgetCaller creates a new read-only instance of Setget, bound to a specific deployed contract.
func NewSetgetCaller(address common.Address, caller bind.ContractCaller) (*SetgetCaller, error) {
	contract, err := bindSetget(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SetgetCaller{contract: contract}, nil
}

// NewSetgetTransactor creates a new write-only instance of Setget, bound to a specific deployed contract.
func NewSetgetTransactor(address common.Address, transactor bind.ContractTransactor) (*SetgetTransactor, error) {
	contract, err := bindSetget(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SetgetTransactor{contract: contract}, nil
}

// NewSetgetFilterer creates a new log filterer instance of Setget, bound to a specific deployed contract.
func NewSetgetFilterer(address common.Address, filterer bind.ContractFilterer) (*SetgetFilterer, error) {
	contract, err := bindSetget(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SetgetFilterer{contract: contract}, nil
}

// bindSetget binds a generic wrapper to an already deployed contract.
func bindSetget(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(SetgetABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Setget *SetgetRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Setget.Contract.SetgetCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Setget *SetgetRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Setget.Contract.SetgetTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Setget *SetgetRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Setget.Contract.SetgetTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Setget *SetgetCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Setget.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Setget *SetgetTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Setget.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Setget *SetgetTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Setget.Contract.contract.Transact(opts, method, params...)
}

// Counter is a free data retrieval call binding the contract method 0x61bc221a.
//
// Solidity: function counter() view returns(uint256)
func (_Setget *SetgetCaller) Counter(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Setget.contract.Call(opts, &out, "counter")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Counter is a free data retrieval call binding the contract method 0x61bc221a.
//
// Solidity: function counter() view returns(uint256)
func (_Setget *SetgetSession) Counter() (*big.Int, error) {
	return _Setget.Contract.Counter(&_Setget.CallOpts)
}

// Counter is a free data retrieval call binding the contract method 0x61bc221a.
//
// Solidity: function counter() view returns(uint256)
func (_Setget *SetgetCallerSession) Counter() (*big.Int, error) {
	return _Setget.Contract.Counter(&_Setget.CallOpts)
}

// Get is a free data retrieval call binding the contract method 0x6d4ce63c.
//
// Solidity: function get() view returns(uint256)
func (_Setget *SetgetCaller) Get(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Setget.contract.Call(opts, &out, "get")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Get is a free data retrieval call binding the contract method 0x6d4ce63c.
//
// Solidity: function get() view returns(uint256)
func (_Setget *SetgetSession) Get() (*big.Int, error) {
	return _Setget.Contract.Get(&_Setget.CallOpts)
}

// Get is a free data retrieval call binding the contract method 0x6d4ce63c.
//
// Solidity: function get() view returns(uint256)
func (_Setget *SetgetCallerSession) Get() (*big.Int, error) {
	return _Setget.Contract.Get(&_Setget.CallOpts)
}

// Set is a paid mutator transaction binding the contract method 0x60fe47b1.
//
// Solidity: function set(uint256 i) returns()
func (_Setget *SetgetTransactor) Set(opts *bind.TransactOpts, i *big.Int) (*types.Transaction, error) {
	return _Setget.contract.Transact(opts, "set", i)
}

// Set is a paid mutator transaction binding the contract method 0x60fe47b1.
//
// Solidity: function set(uint256 i) returns()
func (_Setget *SetgetSession) Set(i *big.Int) (*types.Transaction, error) {
	return _Setget.Contract.Set(&_Setget.TransactOpts, i)
}

// Set is a paid mutator transaction binding the contract method 0x60fe47b1.
//
// Solidity: function set(uint256 i) returns()
func (_Setget *SetgetTransactorSession) Set(i *big.Int) (*types.Transaction, error) {
	return _Setget.Contract.Set(&_Setget.TransactOpts, i)
}
