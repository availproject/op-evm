// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package staking

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

// StakingMetaData contains all meta data concerning the Staking contract.
var StakingMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"minNumSequencers\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maxNumSequencers\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Staked\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Unstaked\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"AccountStake\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"CurrentSequencers\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"CurrentStakedAmount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"IsSequencer\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"MaxNumSequencers\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"MinNumSequencers\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"STAKING_THRESHOLD\",\"outputs\":[{\"internalType\":\"uint128\",\"name\":\"\",\"type\":\"uint128\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"_addressToIsSequencer\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"_addressToSequencerIndex\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"_addressToStakedAmount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"_maximumNumSequencers\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"_minimumNumSequencers\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"_sequencers\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"_stakedAmount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"stake\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"unstake\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"stateMutability\":\"payable\",\"type\":\"receive\"}]",
	Bin: "0x60806040523480156200001157600080fd5b5060405162001740380380620017408339818101604052810190620000379190620000d3565b808211156200007d576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016200007490620001a1565b60405180910390fd5b81600581905550806006819055505050620001c3565b600080fd5b6000819050919050565b620000ad8162000098565b8114620000b957600080fd5b50565b600081519050620000cd81620000a2565b92915050565b60008060408385031215620000ed57620000ec62000093565b5b6000620000fd85828601620000bc565b92505060206200011085828601620000bc565b9150509250929050565b600082825260208201905092915050565b7f4d696e2073657175656e63657273206e756d2063616e206e6f7420626520677260008201527f6561746572207468616e206d6178206e756d206f662073657175656e63657273602082015250565b6000620001896040836200011a565b915062000196826200012b565b604082019050919050565b60006020820190508181036000830152620001bc816200017a565b9050919050565b61156d80620001d36000396000f3fe6080604052600436106100f75760003560e01c8063723306891161008a57806395cf46321161005957806395cf463214610393578063ab5be27d146103be578063bb5a26ff146103e9578063e387a7ed1461042657610165565b806372330689146102b15780637b339263146102ee5780637dceceb81461032b5780638d0841481461036857610165565b80631d7221cc116100c65780631d7221cc1461023a5780632def6620146102655780633a4b66f11461027c578063586741ee1461028657610165565b80630fb18ccc1461016a57806319908df8146101955780631bec453b146101d25780631cb9bdca146101fd57610165565b366101655761011b3373ffffffffffffffffffffffffffffffffffffffff16610451565b1561015b576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161015290610f61565b60405180910390fd5b610163610474565b005b600080fd5b34801561017657600080fd5b5061017f61054b565b60405161018c9190610f9a565b60405180910390f35b3480156101a157600080fd5b506101bc60048036038101906101b79190610fe6565b610551565b6040516101c99190611054565b60405180910390f35b3480156101de57600080fd5b506101e7610590565b6040516101f49190610f9a565b60405180910390f35b34801561020957600080fd5b50610224600480360381019061021f919061109b565b61059a565b6040516102319190610f9a565b60405180910390f35b34801561024657600080fd5b5061024f6105b2565b60405161025c9190610f9a565b60405180910390f35b34801561027157600080fd5b5061027a6105bc565b005b6102846106a7565b005b34801561029257600080fd5b5061029b610710565b6040516102a89190610f9a565b60405180910390f35b3480156102bd57600080fd5b506102d860048036038101906102d3919061109b565b61071a565b6040516102e59190610f9a565b60405180910390f35b3480156102fa57600080fd5b506103156004803603810190610310919061109b565b610763565b60405161032291906110e3565b60405180910390f35b34801561033757600080fd5b50610352600480360381019061034d919061109b565b610783565b60405161035f9190610f9a565b60405180910390f35b34801561037457600080fd5b5061037d61079b565b60405161038a9190610f9a565b60405180910390f35b34801561039f57600080fd5b506103a86107a1565b6040516103b591906111bc565b60405180910390f35b3480156103ca57600080fd5b506103d361082f565b6040516103e09190611209565b60405180910390f35b3480156103f557600080fd5b50610410600480360381019061040b919061109b565b61083b565b60405161041d91906110e3565b60405180910390f35b34801561043257600080fd5b5061043b610891565b6040516104489190610f9a565b60405180910390f35b6000808273ffffffffffffffffffffffffffffffffffffffff163b119050919050565b34600460008282546104869190611253565b9250508190555034600260003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008282546104dc9190611253565b925050819055506104ec33610897565b156104fb576104fa3361090f565b5b3373ffffffffffffffffffffffffffffffffffffffff167f9e71bc8eea02a63969f509818f2dafb9254532904319f9dbda79b67bd34a5f3d346040516105419190610f9a565b60405180910390a2565b60065481565b6000818154811061056157600080fd5b906000526020600020016000915054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b6000600654905090565b60036020528060005260406000206000915090505481565b6000600554905090565b6105db3373ffffffffffffffffffffffffffffffffffffffff16610451565b1561061b576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161061290610f61565b60405180910390fd5b6000600260003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020541161069d576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610694906112f5565b60405180910390fd5b6106a5610a5e565b565b6106c63373ffffffffffffffffffffffffffffffffffffffff16610451565b15610706576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016106fd90610f61565b60405180910390fd5b61070e610474565b565b6000600454905090565b6000600260008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020549050919050565b60016020528060005260406000206000915054906101000a900460ff1681565b60026020528060005260406000206000915090505481565b60055481565b6060600080548060200260200160405190810160405280929190818152602001828054801561082557602002820191906000526020600020905b8160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190600101908083116107db575b5050505050905090565b670de0b6b3a764000081565b6000600160008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a900460ff169050919050565b60045481565b60006108a282610bb0565b1580156109085750670de0b6b3a76400006fffffffffffffffffffffffffffffffff16600260008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000205410155b9050919050565b60065460008054905010610958576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161094f90611387565b60405180910390fd5b60018060008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060006101000a81548160ff021916908315150217905550600080549050600360008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020819055506000819080600181540180825580915050600190039060005260206000200160009091909190916101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555050565b6000600260003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000205490506000600260003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020819055508060046000828254610af991906113a7565b92505081905550610b0933610bb0565b15610b1857610b1733610c06565b5b3373ffffffffffffffffffffffffffffffffffffffff166108fc829081150290604051600060405180830381858888f19350505050158015610b5e573d6000803e3d6000fd5b503373ffffffffffffffffffffffffffffffffffffffff167f0f5bb82176feb1b5e747e28471aa92156a04d9f3ab9f45f28e2d704232b93f7582604051610ba59190610f9a565b60405180910390a250565b6000600160008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a900460ff169050919050565b60055460008054905011610c4f576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610c469061144d565b60405180910390fd5b600080549050600360008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000205410610cd5576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610ccc906114b9565b60405180910390fd5b6000600360008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054905060006001600080549050610d2d91906113a7565b9050808214610e1b576000808281548110610d4b57610d4a6114d9565b5b9060005260206000200160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1690508060008481548110610d8d57610d8c6114d9565b5b9060005260206000200160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555082600360008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002081905550505b6000600160008573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060006101000a81548160ff0219169083151502179055506000600360008573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020819055506000805480610eca57610ec9611508565b5b6001900381819060005260206000200160006101000a81549073ffffffffffffffffffffffffffffffffffffffff02191690559055505050565b600082825260208201905092915050565b7f4f6e6c7920454f412063616e2063616c6c2066756e6374696f6e000000000000600082015250565b6000610f4b601a83610f04565b9150610f5682610f15565b602082019050919050565b60006020820190508181036000830152610f7a81610f3e565b9050919050565b6000819050919050565b610f9481610f81565b82525050565b6000602082019050610faf6000830184610f8b565b92915050565b600080fd5b610fc381610f81565b8114610fce57600080fd5b50565b600081359050610fe081610fba565b92915050565b600060208284031215610ffc57610ffb610fb5565b5b600061100a84828501610fd1565b91505092915050565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b600061103e82611013565b9050919050565b61104e81611033565b82525050565b60006020820190506110696000830184611045565b92915050565b61107881611033565b811461108357600080fd5b50565b6000813590506110958161106f565b92915050565b6000602082840312156110b1576110b0610fb5565b5b60006110bf84828501611086565b91505092915050565b60008115159050919050565b6110dd816110c8565b82525050565b60006020820190506110f860008301846110d4565b92915050565b600081519050919050565b600082825260208201905092915050565b6000819050602082019050919050565b61113381611033565b82525050565b6000611145838361112a565b60208301905092915050565b6000602082019050919050565b6000611169826110fe565b6111738185611109565b935061117e8361111a565b8060005b838110156111af5781516111968882611139565b97506111a183611151565b925050600181019050611182565b5085935050505092915050565b600060208201905081810360008301526111d6818461115e565b905092915050565b60006fffffffffffffffffffffffffffffffff82169050919050565b611203816111de565b82525050565b600060208201905061121e60008301846111fa565b92915050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b600061125e82610f81565b915061126983610f81565b9250827fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff0382111561129e5761129d611224565b5b828201905092915050565b7f4f6e6c79207374616b65722063616e2063616c6c2066756e6374696f6e000000600082015250565b60006112df601d83610f04565b91506112ea826112a9565b602082019050919050565b6000602082019050818103600083015261130e816112d2565b9050919050565b7f53657175656e636572207365742068617320726561636865642066756c6c206360008201527f6170616369747900000000000000000000000000000000000000000000000000602082015250565b6000611371602783610f04565b915061137c82611315565b604082019050919050565b600060208201905081810360008301526113a081611364565b9050919050565b60006113b282610f81565b91506113bd83610f81565b9250828210156113d0576113cf611224565b5b828203905092915050565b7f53657175656e636572732063616e2774206265206c657373207468616e20746860008201527f65206d696e696d756d2072657175697265642073657175656e636572206e756d602082015250565b6000611437604083610f04565b9150611442826113db565b604082019050919050565b600060208201905081810360008301526114668161142a565b9050919050565b7f696e646578206f7574206f662072616e67650000000000000000000000000000600082015250565b60006114a3601283610f04565b91506114ae8261146d565b602082019050919050565b600060208201905081810360008301526114d281611496565b9050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052603260045260246000fd5b7f4e487b7100000000000000000000000000000000000000000000000000000000600052603160045260246000fdfea2646970667358221220aad55fdb40ea9b9aa19a82e4045143b02946c81e396e8673a86a6c84cb30b43364736f6c634300080f0033",
}

// StakingABI is the input ABI used to generate the binding from.
// Deprecated: Use StakingMetaData.ABI instead.
var StakingABI = StakingMetaData.ABI

// StakingBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use StakingMetaData.Bin instead.
var StakingBin = StakingMetaData.Bin

// DeployStaking deploys a new Ethereum contract, binding an instance of Staking to it.
func DeployStaking(auth *bind.TransactOpts, backend bind.ContractBackend, minNumSequencers *big.Int, maxNumSequencers *big.Int) (common.Address, *types.Transaction, *Staking, error) {
	parsed, err := StakingMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(StakingBin), backend, minNumSequencers, maxNumSequencers)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Staking{StakingCaller: StakingCaller{contract: contract}, StakingTransactor: StakingTransactor{contract: contract}, StakingFilterer: StakingFilterer{contract: contract}}, nil
}

// Staking is an auto generated Go binding around an Ethereum contract.
type Staking struct {
	StakingCaller     // Read-only binding to the contract
	StakingTransactor // Write-only binding to the contract
	StakingFilterer   // Log filterer for contract events
}

// StakingCaller is an auto generated read-only Go binding around an Ethereum contract.
type StakingCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// StakingTransactor is an auto generated write-only Go binding around an Ethereum contract.
type StakingTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// StakingFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type StakingFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// StakingSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type StakingSession struct {
	Contract     *Staking          // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// StakingCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type StakingCallerSession struct {
	Contract *StakingCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts  // Call options to use throughout this session
}

// StakingTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type StakingTransactorSession struct {
	Contract     *StakingTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// StakingRaw is an auto generated low-level Go binding around an Ethereum contract.
type StakingRaw struct {
	Contract *Staking // Generic contract binding to access the raw methods on
}

// StakingCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type StakingCallerRaw struct {
	Contract *StakingCaller // Generic read-only contract binding to access the raw methods on
}

// StakingTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type StakingTransactorRaw struct {
	Contract *StakingTransactor // Generic write-only contract binding to access the raw methods on
}

// NewStaking creates a new instance of Staking, bound to a specific deployed contract.
func NewStaking(address common.Address, backend bind.ContractBackend) (*Staking, error) {
	contract, err := bindStaking(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Staking{StakingCaller: StakingCaller{contract: contract}, StakingTransactor: StakingTransactor{contract: contract}, StakingFilterer: StakingFilterer{contract: contract}}, nil
}

// NewStakingCaller creates a new read-only instance of Staking, bound to a specific deployed contract.
func NewStakingCaller(address common.Address, caller bind.ContractCaller) (*StakingCaller, error) {
	contract, err := bindStaking(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &StakingCaller{contract: contract}, nil
}

// NewStakingTransactor creates a new write-only instance of Staking, bound to a specific deployed contract.
func NewStakingTransactor(address common.Address, transactor bind.ContractTransactor) (*StakingTransactor, error) {
	contract, err := bindStaking(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &StakingTransactor{contract: contract}, nil
}

// NewStakingFilterer creates a new log filterer instance of Staking, bound to a specific deployed contract.
func NewStakingFilterer(address common.Address, filterer bind.ContractFilterer) (*StakingFilterer, error) {
	contract, err := bindStaking(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &StakingFilterer{contract: contract}, nil
}

// bindStaking binds a generic wrapper to an already deployed contract.
func bindStaking(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(StakingABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Staking *StakingRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Staking.Contract.StakingCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Staking *StakingRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Staking.Contract.StakingTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Staking *StakingRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Staking.Contract.StakingTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Staking *StakingCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Staking.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Staking *StakingTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Staking.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Staking *StakingTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Staking.Contract.contract.Transact(opts, method, params...)
}

// AccountStake is a free data retrieval call binding the contract method 0x72330689.
//
// Solidity: function AccountStake(address addr) view returns(uint256)
func (_Staking *StakingCaller) AccountStake(opts *bind.CallOpts, addr common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Staking.contract.Call(opts, &out, "AccountStake", addr)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// AccountStake is a free data retrieval call binding the contract method 0x72330689.
//
// Solidity: function AccountStake(address addr) view returns(uint256)
func (_Staking *StakingSession) AccountStake(addr common.Address) (*big.Int, error) {
	return _Staking.Contract.AccountStake(&_Staking.CallOpts, addr)
}

// AccountStake is a free data retrieval call binding the contract method 0x72330689.
//
// Solidity: function AccountStake(address addr) view returns(uint256)
func (_Staking *StakingCallerSession) AccountStake(addr common.Address) (*big.Int, error) {
	return _Staking.Contract.AccountStake(&_Staking.CallOpts, addr)
}

// CurrentSequencers is a free data retrieval call binding the contract method 0x95cf4632.
//
// Solidity: function CurrentSequencers() view returns(address[])
func (_Staking *StakingCaller) CurrentSequencers(opts *bind.CallOpts) ([]common.Address, error) {
	var out []interface{}
	err := _Staking.contract.Call(opts, &out, "CurrentSequencers")

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// CurrentSequencers is a free data retrieval call binding the contract method 0x95cf4632.
//
// Solidity: function CurrentSequencers() view returns(address[])
func (_Staking *StakingSession) CurrentSequencers() ([]common.Address, error) {
	return _Staking.Contract.CurrentSequencers(&_Staking.CallOpts)
}

// CurrentSequencers is a free data retrieval call binding the contract method 0x95cf4632.
//
// Solidity: function CurrentSequencers() view returns(address[])
func (_Staking *StakingCallerSession) CurrentSequencers() ([]common.Address, error) {
	return _Staking.Contract.CurrentSequencers(&_Staking.CallOpts)
}

// CurrentStakedAmount is a free data retrieval call binding the contract method 0x586741ee.
//
// Solidity: function CurrentStakedAmount() view returns(uint256)
func (_Staking *StakingCaller) CurrentStakedAmount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Staking.contract.Call(opts, &out, "CurrentStakedAmount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// CurrentStakedAmount is a free data retrieval call binding the contract method 0x586741ee.
//
// Solidity: function CurrentStakedAmount() view returns(uint256)
func (_Staking *StakingSession) CurrentStakedAmount() (*big.Int, error) {
	return _Staking.Contract.CurrentStakedAmount(&_Staking.CallOpts)
}

// CurrentStakedAmount is a free data retrieval call binding the contract method 0x586741ee.
//
// Solidity: function CurrentStakedAmount() view returns(uint256)
func (_Staking *StakingCallerSession) CurrentStakedAmount() (*big.Int, error) {
	return _Staking.Contract.CurrentStakedAmount(&_Staking.CallOpts)
}

// IsSequencer is a free data retrieval call binding the contract method 0xbb5a26ff.
//
// Solidity: function IsSequencer(address addr) view returns(bool)
func (_Staking *StakingCaller) IsSequencer(opts *bind.CallOpts, addr common.Address) (bool, error) {
	var out []interface{}
	err := _Staking.contract.Call(opts, &out, "IsSequencer", addr)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsSequencer is a free data retrieval call binding the contract method 0xbb5a26ff.
//
// Solidity: function IsSequencer(address addr) view returns(bool)
func (_Staking *StakingSession) IsSequencer(addr common.Address) (bool, error) {
	return _Staking.Contract.IsSequencer(&_Staking.CallOpts, addr)
}

// IsSequencer is a free data retrieval call binding the contract method 0xbb5a26ff.
//
// Solidity: function IsSequencer(address addr) view returns(bool)
func (_Staking *StakingCallerSession) IsSequencer(addr common.Address) (bool, error) {
	return _Staking.Contract.IsSequencer(&_Staking.CallOpts, addr)
}

// MaxNumSequencers is a free data retrieval call binding the contract method 0x1bec453b.
//
// Solidity: function MaxNumSequencers() view returns(uint256)
func (_Staking *StakingCaller) MaxNumSequencers(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Staking.contract.Call(opts, &out, "MaxNumSequencers")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MaxNumSequencers is a free data retrieval call binding the contract method 0x1bec453b.
//
// Solidity: function MaxNumSequencers() view returns(uint256)
func (_Staking *StakingSession) MaxNumSequencers() (*big.Int, error) {
	return _Staking.Contract.MaxNumSequencers(&_Staking.CallOpts)
}

// MaxNumSequencers is a free data retrieval call binding the contract method 0x1bec453b.
//
// Solidity: function MaxNumSequencers() view returns(uint256)
func (_Staking *StakingCallerSession) MaxNumSequencers() (*big.Int, error) {
	return _Staking.Contract.MaxNumSequencers(&_Staking.CallOpts)
}

// MinNumSequencers is a free data retrieval call binding the contract method 0x1d7221cc.
//
// Solidity: function MinNumSequencers() view returns(uint256)
func (_Staking *StakingCaller) MinNumSequencers(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Staking.contract.Call(opts, &out, "MinNumSequencers")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MinNumSequencers is a free data retrieval call binding the contract method 0x1d7221cc.
//
// Solidity: function MinNumSequencers() view returns(uint256)
func (_Staking *StakingSession) MinNumSequencers() (*big.Int, error) {
	return _Staking.Contract.MinNumSequencers(&_Staking.CallOpts)
}

// MinNumSequencers is a free data retrieval call binding the contract method 0x1d7221cc.
//
// Solidity: function MinNumSequencers() view returns(uint256)
func (_Staking *StakingCallerSession) MinNumSequencers() (*big.Int, error) {
	return _Staking.Contract.MinNumSequencers(&_Staking.CallOpts)
}

// STAKINGTHRESHOLD is a free data retrieval call binding the contract method 0xab5be27d.
//
// Solidity: function STAKING_THRESHOLD() view returns(uint128)
func (_Staking *StakingCaller) STAKINGTHRESHOLD(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Staking.contract.Call(opts, &out, "STAKING_THRESHOLD")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// STAKINGTHRESHOLD is a free data retrieval call binding the contract method 0xab5be27d.
//
// Solidity: function STAKING_THRESHOLD() view returns(uint128)
func (_Staking *StakingSession) STAKINGTHRESHOLD() (*big.Int, error) {
	return _Staking.Contract.STAKINGTHRESHOLD(&_Staking.CallOpts)
}

// STAKINGTHRESHOLD is a free data retrieval call binding the contract method 0xab5be27d.
//
// Solidity: function STAKING_THRESHOLD() view returns(uint128)
func (_Staking *StakingCallerSession) STAKINGTHRESHOLD() (*big.Int, error) {
	return _Staking.Contract.STAKINGTHRESHOLD(&_Staking.CallOpts)
}

// AddressToIsSequencer is a free data retrieval call binding the contract method 0x7b339263.
//
// Solidity: function _addressToIsSequencer(address ) view returns(bool)
func (_Staking *StakingCaller) AddressToIsSequencer(opts *bind.CallOpts, arg0 common.Address) (bool, error) {
	var out []interface{}
	err := _Staking.contract.Call(opts, &out, "_addressToIsSequencer", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// AddressToIsSequencer is a free data retrieval call binding the contract method 0x7b339263.
//
// Solidity: function _addressToIsSequencer(address ) view returns(bool)
func (_Staking *StakingSession) AddressToIsSequencer(arg0 common.Address) (bool, error) {
	return _Staking.Contract.AddressToIsSequencer(&_Staking.CallOpts, arg0)
}

// AddressToIsSequencer is a free data retrieval call binding the contract method 0x7b339263.
//
// Solidity: function _addressToIsSequencer(address ) view returns(bool)
func (_Staking *StakingCallerSession) AddressToIsSequencer(arg0 common.Address) (bool, error) {
	return _Staking.Contract.AddressToIsSequencer(&_Staking.CallOpts, arg0)
}

// AddressToSequencerIndex is a free data retrieval call binding the contract method 0x1cb9bdca.
//
// Solidity: function _addressToSequencerIndex(address ) view returns(uint256)
func (_Staking *StakingCaller) AddressToSequencerIndex(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Staking.contract.Call(opts, &out, "_addressToSequencerIndex", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// AddressToSequencerIndex is a free data retrieval call binding the contract method 0x1cb9bdca.
//
// Solidity: function _addressToSequencerIndex(address ) view returns(uint256)
func (_Staking *StakingSession) AddressToSequencerIndex(arg0 common.Address) (*big.Int, error) {
	return _Staking.Contract.AddressToSequencerIndex(&_Staking.CallOpts, arg0)
}

// AddressToSequencerIndex is a free data retrieval call binding the contract method 0x1cb9bdca.
//
// Solidity: function _addressToSequencerIndex(address ) view returns(uint256)
func (_Staking *StakingCallerSession) AddressToSequencerIndex(arg0 common.Address) (*big.Int, error) {
	return _Staking.Contract.AddressToSequencerIndex(&_Staking.CallOpts, arg0)
}

// AddressToStakedAmount is a free data retrieval call binding the contract method 0x7dceceb8.
//
// Solidity: function _addressToStakedAmount(address ) view returns(uint256)
func (_Staking *StakingCaller) AddressToStakedAmount(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Staking.contract.Call(opts, &out, "_addressToStakedAmount", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// AddressToStakedAmount is a free data retrieval call binding the contract method 0x7dceceb8.
//
// Solidity: function _addressToStakedAmount(address ) view returns(uint256)
func (_Staking *StakingSession) AddressToStakedAmount(arg0 common.Address) (*big.Int, error) {
	return _Staking.Contract.AddressToStakedAmount(&_Staking.CallOpts, arg0)
}

// AddressToStakedAmount is a free data retrieval call binding the contract method 0x7dceceb8.
//
// Solidity: function _addressToStakedAmount(address ) view returns(uint256)
func (_Staking *StakingCallerSession) AddressToStakedAmount(arg0 common.Address) (*big.Int, error) {
	return _Staking.Contract.AddressToStakedAmount(&_Staking.CallOpts, arg0)
}

// MaximumNumSequencers is a free data retrieval call binding the contract method 0x0fb18ccc.
//
// Solidity: function _maximumNumSequencers() view returns(uint256)
func (_Staking *StakingCaller) MaximumNumSequencers(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Staking.contract.Call(opts, &out, "_maximumNumSequencers")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MaximumNumSequencers is a free data retrieval call binding the contract method 0x0fb18ccc.
//
// Solidity: function _maximumNumSequencers() view returns(uint256)
func (_Staking *StakingSession) MaximumNumSequencers() (*big.Int, error) {
	return _Staking.Contract.MaximumNumSequencers(&_Staking.CallOpts)
}

// MaximumNumSequencers is a free data retrieval call binding the contract method 0x0fb18ccc.
//
// Solidity: function _maximumNumSequencers() view returns(uint256)
func (_Staking *StakingCallerSession) MaximumNumSequencers() (*big.Int, error) {
	return _Staking.Contract.MaximumNumSequencers(&_Staking.CallOpts)
}

// MinimumNumSequencers is a free data retrieval call binding the contract method 0x8d084148.
//
// Solidity: function _minimumNumSequencers() view returns(uint256)
func (_Staking *StakingCaller) MinimumNumSequencers(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Staking.contract.Call(opts, &out, "_minimumNumSequencers")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MinimumNumSequencers is a free data retrieval call binding the contract method 0x8d084148.
//
// Solidity: function _minimumNumSequencers() view returns(uint256)
func (_Staking *StakingSession) MinimumNumSequencers() (*big.Int, error) {
	return _Staking.Contract.MinimumNumSequencers(&_Staking.CallOpts)
}

// MinimumNumSequencers is a free data retrieval call binding the contract method 0x8d084148.
//
// Solidity: function _minimumNumSequencers() view returns(uint256)
func (_Staking *StakingCallerSession) MinimumNumSequencers() (*big.Int, error) {
	return _Staking.Contract.MinimumNumSequencers(&_Staking.CallOpts)
}

// Sequencers is a free data retrieval call binding the contract method 0x19908df8.
//
// Solidity: function _sequencers(uint256 ) view returns(address)
func (_Staking *StakingCaller) Sequencers(opts *bind.CallOpts, arg0 *big.Int) (common.Address, error) {
	var out []interface{}
	err := _Staking.contract.Call(opts, &out, "_sequencers", arg0)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Sequencers is a free data retrieval call binding the contract method 0x19908df8.
//
// Solidity: function _sequencers(uint256 ) view returns(address)
func (_Staking *StakingSession) Sequencers(arg0 *big.Int) (common.Address, error) {
	return _Staking.Contract.Sequencers(&_Staking.CallOpts, arg0)
}

// Sequencers is a free data retrieval call binding the contract method 0x19908df8.
//
// Solidity: function _sequencers(uint256 ) view returns(address)
func (_Staking *StakingCallerSession) Sequencers(arg0 *big.Int) (common.Address, error) {
	return _Staking.Contract.Sequencers(&_Staking.CallOpts, arg0)
}

// StakedAmount is a free data retrieval call binding the contract method 0xe387a7ed.
//
// Solidity: function _stakedAmount() view returns(uint256)
func (_Staking *StakingCaller) StakedAmount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Staking.contract.Call(opts, &out, "_stakedAmount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// StakedAmount is a free data retrieval call binding the contract method 0xe387a7ed.
//
// Solidity: function _stakedAmount() view returns(uint256)
func (_Staking *StakingSession) StakedAmount() (*big.Int, error) {
	return _Staking.Contract.StakedAmount(&_Staking.CallOpts)
}

// StakedAmount is a free data retrieval call binding the contract method 0xe387a7ed.
//
// Solidity: function _stakedAmount() view returns(uint256)
func (_Staking *StakingCallerSession) StakedAmount() (*big.Int, error) {
	return _Staking.Contract.StakedAmount(&_Staking.CallOpts)
}

// Stake is a paid mutator transaction binding the contract method 0x3a4b66f1.
//
// Solidity: function stake() payable returns()
func (_Staking *StakingTransactor) Stake(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Staking.contract.Transact(opts, "stake")
}

// Stake is a paid mutator transaction binding the contract method 0x3a4b66f1.
//
// Solidity: function stake() payable returns()
func (_Staking *StakingSession) Stake() (*types.Transaction, error) {
	return _Staking.Contract.Stake(&_Staking.TransactOpts)
}

// Stake is a paid mutator transaction binding the contract method 0x3a4b66f1.
//
// Solidity: function stake() payable returns()
func (_Staking *StakingTransactorSession) Stake() (*types.Transaction, error) {
	return _Staking.Contract.Stake(&_Staking.TransactOpts)
}

// Unstake is a paid mutator transaction binding the contract method 0x2def6620.
//
// Solidity: function unstake() returns()
func (_Staking *StakingTransactor) Unstake(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Staking.contract.Transact(opts, "unstake")
}

// Unstake is a paid mutator transaction binding the contract method 0x2def6620.
//
// Solidity: function unstake() returns()
func (_Staking *StakingSession) Unstake() (*types.Transaction, error) {
	return _Staking.Contract.Unstake(&_Staking.TransactOpts)
}

// Unstake is a paid mutator transaction binding the contract method 0x2def6620.
//
// Solidity: function unstake() returns()
func (_Staking *StakingTransactorSession) Unstake() (*types.Transaction, error) {
	return _Staking.Contract.Unstake(&_Staking.TransactOpts)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_Staking *StakingTransactor) Receive(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Staking.contract.RawTransact(opts, nil) // calldata is disallowed for receive function
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_Staking *StakingSession) Receive() (*types.Transaction, error) {
	return _Staking.Contract.Receive(&_Staking.TransactOpts)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_Staking *StakingTransactorSession) Receive() (*types.Transaction, error) {
	return _Staking.Contract.Receive(&_Staking.TransactOpts)
}

// StakingStakedIterator is returned from FilterStaked and is used to iterate over the raw logs and unpacked data for Staked events raised by the Staking contract.
type StakingStakedIterator struct {
	Event *StakingStaked // Event containing the contract specifics and raw log

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
func (it *StakingStakedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(StakingStaked)
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
		it.Event = new(StakingStaked)
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
func (it *StakingStakedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *StakingStakedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// StakingStaked represents a Staked event raised by the Staking contract.
type StakingStaked struct {
	Account common.Address
	Amount  *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterStaked is a free log retrieval operation binding the contract event 0x9e71bc8eea02a63969f509818f2dafb9254532904319f9dbda79b67bd34a5f3d.
//
// Solidity: event Staked(address indexed account, uint256 amount)
func (_Staking *StakingFilterer) FilterStaked(opts *bind.FilterOpts, account []common.Address) (*StakingStakedIterator, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _Staking.contract.FilterLogs(opts, "Staked", accountRule)
	if err != nil {
		return nil, err
	}
	return &StakingStakedIterator{contract: _Staking.contract, event: "Staked", logs: logs, sub: sub}, nil
}

// WatchStaked is a free log subscription operation binding the contract event 0x9e71bc8eea02a63969f509818f2dafb9254532904319f9dbda79b67bd34a5f3d.
//
// Solidity: event Staked(address indexed account, uint256 amount)
func (_Staking *StakingFilterer) WatchStaked(opts *bind.WatchOpts, sink chan<- *StakingStaked, account []common.Address) (event.Subscription, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _Staking.contract.WatchLogs(opts, "Staked", accountRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(StakingStaked)
				if err := _Staking.contract.UnpackLog(event, "Staked", log); err != nil {
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

// ParseStaked is a log parse operation binding the contract event 0x9e71bc8eea02a63969f509818f2dafb9254532904319f9dbda79b67bd34a5f3d.
//
// Solidity: event Staked(address indexed account, uint256 amount)
func (_Staking *StakingFilterer) ParseStaked(log types.Log) (*StakingStaked, error) {
	event := new(StakingStaked)
	if err := _Staking.contract.UnpackLog(event, "Staked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// StakingUnstakedIterator is returned from FilterUnstaked and is used to iterate over the raw logs and unpacked data for Unstaked events raised by the Staking contract.
type StakingUnstakedIterator struct {
	Event *StakingUnstaked // Event containing the contract specifics and raw log

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
func (it *StakingUnstakedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(StakingUnstaked)
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
		it.Event = new(StakingUnstaked)
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
func (it *StakingUnstakedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *StakingUnstakedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// StakingUnstaked represents a Unstaked event raised by the Staking contract.
type StakingUnstaked struct {
	Account common.Address
	Amount  *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterUnstaked is a free log retrieval operation binding the contract event 0x0f5bb82176feb1b5e747e28471aa92156a04d9f3ab9f45f28e2d704232b93f75.
//
// Solidity: event Unstaked(address indexed account, uint256 amount)
func (_Staking *StakingFilterer) FilterUnstaked(opts *bind.FilterOpts, account []common.Address) (*StakingUnstakedIterator, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _Staking.contract.FilterLogs(opts, "Unstaked", accountRule)
	if err != nil {
		return nil, err
	}
	return &StakingUnstakedIterator{contract: _Staking.contract, event: "Unstaked", logs: logs, sub: sub}, nil
}

// WatchUnstaked is a free log subscription operation binding the contract event 0x0f5bb82176feb1b5e747e28471aa92156a04d9f3ab9f45f28e2d704232b93f75.
//
// Solidity: event Unstaked(address indexed account, uint256 amount)
func (_Staking *StakingFilterer) WatchUnstaked(opts *bind.WatchOpts, sink chan<- *StakingUnstaked, account []common.Address) (event.Subscription, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _Staking.contract.WatchLogs(opts, "Unstaked", accountRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(StakingUnstaked)
				if err := _Staking.contract.UnpackLog(event, "Unstaked", log); err != nil {
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

// ParseUnstaked is a log parse operation binding the contract event 0x0f5bb82176feb1b5e747e28471aa92156a04d9f3ab9f45f28e2d704232b93f75.
//
// Solidity: event Unstaked(address indexed account, uint256 amount)
func (_Staking *StakingFilterer) ParseUnstaked(log types.Log) (*StakingUnstaked, error) {
	event := new(StakingUnstaked)
	if err := _Staking.contract.UnpackLog(event, "Unstaked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
