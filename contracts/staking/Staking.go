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
	ABI: "[{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"minNumSequencers\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maxNumSequencers\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Slashed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Staked\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Unstaked\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"AccountStake\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"CurrentSequencers\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"CurrentStakedAmount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"IsSequencer\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"MaxNumSequencers\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"MinNumSequencers\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"STAKING_THRESHOLD\",\"outputs\":[{\"internalType\":\"uint128\",\"name\":\"\",\"type\":\"uint128\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"_addressToIsSequencer\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"_addressToSequencerIndex\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"_addressToStakedAmount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"_maximumNumSequencers\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"_minimumNumSequencers\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"_sequencers\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"_stakedAmount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"slash\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"stake\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"unstake\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"stateMutability\":\"payable\",\"type\":\"receive\"}]",
	Bin: "0x60806040523480156200001157600080fd5b50604051620019e0380380620019e08339818101604052810190620000379190620000d3565b808211156200007d576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016200007490620001a1565b60405180910390fd5b81600581905550806006819055505050620001c3565b600080fd5b6000819050919050565b620000ad8162000098565b8114620000b957600080fd5b50565b600081519050620000cd81620000a2565b92915050565b60008060408385031215620000ed57620000ec62000093565b5b6000620000fd85828601620000bc565b92505060206200011085828601620000bc565b9150509250929050565b600082825260208201905092915050565b7f4d696e2073657175656e63657273206e756d2063616e206e6f7420626520677260008201527f6561746572207468616e206d6178206e756d206f662073657175656e63657273602082015250565b6000620001896040836200011a565b915062000196826200012b565b604082019050919050565b60006020820190508181036000830152620001bc816200017a565b9050919050565b61180d80620001d36000396000f3fe6080604052600436106101025760003560e01c8063586741ee116100955780638d084148116100645780638d0841481461038a57806395cf4632146103b5578063ab5be27d146103e0578063bb5a26ff1461040b578063e387a7ed1461044857610170565b8063586741ee146102a857806372330689146102d35780637b339263146103105780637dceceb81461034d57610170565b80631d7221cc116100d15780631d7221cc146102455780632da25de3146102705780632def6620146102875780633a4b66f11461029e57610170565b80630fb18ccc1461017557806319908df8146101a05780631bec453b146101dd5780631cb9bdca1461020857610170565b36610170576101263373ffffffffffffffffffffffffffffffffffffffff16610473565b15610166576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161015d9061116f565b60405180910390fd5b61016e610496565b005b600080fd5b34801561018157600080fd5b5061018a61056d565b60405161019791906111a8565b60405180910390f35b3480156101ac57600080fd5b506101c760048036038101906101c291906111f4565b610573565b6040516101d49190611262565b60405180910390f35b3480156101e957600080fd5b506101f26105b2565b6040516101ff91906111a8565b60405180910390f35b34801561021457600080fd5b5061022f600480360381019061022a91906112a9565b6105bc565b60405161023c91906111a8565b60405180910390f35b34801561025157600080fd5b5061025a6105d4565b60405161026791906111a8565b60405180910390f35b34801561027c57600080fd5b506102856105de565b005b34801561029357600080fd5b5061029c6106c9565b005b6102a66107b4565b005b3480156102b457600080fd5b506102bd61081d565b6040516102ca91906111a8565b60405180910390f35b3480156102df57600080fd5b506102fa60048036038101906102f591906112a9565b610827565b60405161030791906111a8565b60405180910390f35b34801561031c57600080fd5b50610337600480360381019061033291906112a9565b610870565b60405161034491906112f1565b60405180910390f35b34801561035957600080fd5b50610374600480360381019061036f91906112a9565b610890565b60405161038191906111a8565b60405180910390f35b34801561039657600080fd5b5061039f6108a8565b6040516103ac91906111a8565b60405180910390f35b3480156103c157600080fd5b506103ca6108ae565b6040516103d791906113ca565b60405180910390f35b3480156103ec57600080fd5b506103f561093c565b6040516104029190611417565b60405180910390f35b34801561041757600080fd5b50610432600480360381019061042d91906112a9565b610948565b60405161043f91906112f1565b60405180910390f35b34801561045457600080fd5b5061045d61099e565b60405161046a91906111a8565b60405180910390f35b6000808273ffffffffffffffffffffffffffffffffffffffff163b119050919050565b34600460008282546104a89190611461565b9250508190555034600260003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008282546104fe9190611461565b9250508190555061050e336109a4565b1561051d5761051c33610a1c565b5b3373ffffffffffffffffffffffffffffffffffffffff167f9e71bc8eea02a63969f509818f2dafb9254532904319f9dbda79b67bd34a5f3d3460405161056391906111a8565b60405180910390a2565b60065481565b6000818154811061058357600080fd5b906000526020600020016000915054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b6000600654905090565b60036020528060005260406000206000915090505481565b6000600554905090565b6105fd3373ffffffffffffffffffffffffffffffffffffffff16610473565b1561063d576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016106349061116f565b60405180910390fd5b6000600260003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054116106bf576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016106b690611503565b60405180910390fd5b6106c7610b6b565b565b6106e83373ffffffffffffffffffffffffffffffffffffffff16610473565b15610728576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161071f9061116f565b60405180910390fd5b6000600260003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054116107aa576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016107a190611503565b60405180910390fd5b6107b2610c6c565b565b6107d33373ffffffffffffffffffffffffffffffffffffffff16610473565b15610813576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161080a9061116f565b60405180910390fd5b61081b610496565b565b6000600454905090565b6000600260008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020549050919050565b60016020528060005260406000206000915054906101000a900460ff1681565b60026020528060005260406000206000915090505481565b60055481565b6060600080548060200260200160405190810160405280929190818152602001828054801561093257602002820191906000526020600020905b8160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190600101908083116108e8575b5050505050905090565b670de0b6b3a764000081565b6000600160008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a900460ff169050919050565b60045481565b60006109af82610dbe565b158015610a155750670de0b6b3a76400006fffffffffffffffffffffffffffffffff16600260008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000205410155b9050919050565b60065460008054905010610a65576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610a5c90611595565b60405180910390fd5b60018060008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060006101000a81548160ff021916908315150217905550600080549050600360008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020819055506000819080600181540180825580915050600190039060005260206000200160009091909190916101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555050565b6000600260003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020549050803410610bf1576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610be890611627565b60405180910390fd5b60003482610bff9190611647565b90508060046000828254610c139190611647565b925050819055503373ffffffffffffffffffffffffffffffffffffffff167f4ed05e9673c26d2ed44f7ef6a7f2942df0ee3b5e1e17db4b99f9dcd261a339cd83604051610c6091906111a8565b60405180910390a25050565b6000600260003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000205490506000600260003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020819055508060046000828254610d079190611647565b92505081905550610d1733610dbe565b15610d2657610d2533610e14565b5b3373ffffffffffffffffffffffffffffffffffffffff166108fc829081150290604051600060405180830381858888f19350505050158015610d6c573d6000803e3d6000fd5b503373ffffffffffffffffffffffffffffffffffffffff167f0f5bb82176feb1b5e747e28471aa92156a04d9f3ab9f45f28e2d704232b93f7582604051610db391906111a8565b60405180910390a250565b6000600160008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a900460ff169050919050565b60055460008054905011610e5d576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610e54906116ed565b60405180910390fd5b600080549050600360008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000205410610ee3576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610eda90611759565b60405180910390fd5b6000600360008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054905060006001600080549050610f3b9190611647565b9050808214611029576000808281548110610f5957610f58611779565b5b9060005260206000200160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1690508060008481548110610f9b57610f9a611779565b5b9060005260206000200160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555082600360008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002081905550505b6000600160008573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060006101000a81548160ff0219169083151502179055506000600360008573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000208190555060008054806110d8576110d76117a8565b5b6001900381819060005260206000200160006101000a81549073ffffffffffffffffffffffffffffffffffffffff02191690559055505050565b600082825260208201905092915050565b7f4f6e6c7920454f412063616e2063616c6c2066756e6374696f6e000000000000600082015250565b6000611159601a83611112565b915061116482611123565b602082019050919050565b600060208201905081810360008301526111888161114c565b9050919050565b6000819050919050565b6111a28161118f565b82525050565b60006020820190506111bd6000830184611199565b92915050565b600080fd5b6111d18161118f565b81146111dc57600080fd5b50565b6000813590506111ee816111c8565b92915050565b60006020828403121561120a576112096111c3565b5b6000611218848285016111df565b91505092915050565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b600061124c82611221565b9050919050565b61125c81611241565b82525050565b60006020820190506112776000830184611253565b92915050565b61128681611241565b811461129157600080fd5b50565b6000813590506112a38161127d565b92915050565b6000602082840312156112bf576112be6111c3565b5b60006112cd84828501611294565b91505092915050565b60008115159050919050565b6112eb816112d6565b82525050565b600060208201905061130660008301846112e2565b92915050565b600081519050919050565b600082825260208201905092915050565b6000819050602082019050919050565b61134181611241565b82525050565b60006113538383611338565b60208301905092915050565b6000602082019050919050565b60006113778261130c565b6113818185611317565b935061138c83611328565b8060005b838110156113bd5781516113a48882611347565b97506113af8361135f565b925050600181019050611390565b5085935050505092915050565b600060208201905081810360008301526113e4818461136c565b905092915050565b60006fffffffffffffffffffffffffffffffff82169050919050565b611411816113ec565b82525050565b600060208201905061142c6000830184611408565b92915050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b600061146c8261118f565b91506114778361118f565b9250827fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff038211156114ac576114ab611432565b5b828201905092915050565b7f4f6e6c79207374616b65722063616e2063616c6c2066756e6374696f6e000000600082015250565b60006114ed601d83611112565b91506114f8826114b7565b602082019050919050565b6000602082019050818103600083015261151c816114e0565b9050919050565b7f53657175656e636572207365742068617320726561636865642066756c6c206360008201527f6170616369747900000000000000000000000000000000000000000000000000602082015250565b600061157f602783611112565b915061158a82611523565b604082019050919050565b600060208201905081810360008301526115ae81611572565b9050919050565b7f536c61736820616d6f756e742063616e6e6f742062652067726561746572207460008201527f68616e207374616b656420616d6f756e74000000000000000000000000000000602082015250565b6000611611603183611112565b915061161c826115b5565b604082019050919050565b6000602082019050818103600083015261164081611604565b9050919050565b60006116528261118f565b915061165d8361118f565b9250828210156116705761166f611432565b5b828203905092915050565b7f53657175656e636572732063616e2774206265206c657373207468616e20746860008201527f65206d696e696d756d2072657175697265642073657175656e636572206e756d602082015250565b60006116d7604083611112565b91506116e28261167b565b604082019050919050565b60006020820190508181036000830152611706816116ca565b9050919050565b7f696e646578206f7574206f662072616e67650000000000000000000000000000600082015250565b6000611743601283611112565b915061174e8261170d565b602082019050919050565b6000602082019050818103600083015261177281611736565b9050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052603260045260246000fd5b7f4e487b7100000000000000000000000000000000000000000000000000000000600052603160045260246000fdfea2646970667358221220b012836838963b0939ff14ea8d7dd76b82a05f08b241c9ff4ad58f94393425a364736f6c634300080f0033",
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

// Slash is a paid mutator transaction binding the contract method 0x2da25de3.
//
// Solidity: function slash() returns()
func (_Staking *StakingTransactor) Slash(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Staking.contract.Transact(opts, "slash")
}

// Slash is a paid mutator transaction binding the contract method 0x2da25de3.
//
// Solidity: function slash() returns()
func (_Staking *StakingSession) Slash() (*types.Transaction, error) {
	return _Staking.Contract.Slash(&_Staking.TransactOpts)
}

// Slash is a paid mutator transaction binding the contract method 0x2da25de3.
//
// Solidity: function slash() returns()
func (_Staking *StakingTransactorSession) Slash() (*types.Transaction, error) {
	return _Staking.Contract.Slash(&_Staking.TransactOpts)
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

// StakingSlashedIterator is returned from FilterSlashed and is used to iterate over the raw logs and unpacked data for Slashed events raised by the Staking contract.
type StakingSlashedIterator struct {
	Event *StakingSlashed // Event containing the contract specifics and raw log

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
func (it *StakingSlashedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(StakingSlashed)
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
		it.Event = new(StakingSlashed)
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
func (it *StakingSlashedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *StakingSlashedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// StakingSlashed represents a Slashed event raised by the Staking contract.
type StakingSlashed struct {
	Account common.Address
	Amount  *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterSlashed is a free log retrieval operation binding the contract event 0x4ed05e9673c26d2ed44f7ef6a7f2942df0ee3b5e1e17db4b99f9dcd261a339cd.
//
// Solidity: event Slashed(address indexed account, uint256 amount)
func (_Staking *StakingFilterer) FilterSlashed(opts *bind.FilterOpts, account []common.Address) (*StakingSlashedIterator, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _Staking.contract.FilterLogs(opts, "Slashed", accountRule)
	if err != nil {
		return nil, err
	}
	return &StakingSlashedIterator{contract: _Staking.contract, event: "Slashed", logs: logs, sub: sub}, nil
}

// WatchSlashed is a free log subscription operation binding the contract event 0x4ed05e9673c26d2ed44f7ef6a7f2942df0ee3b5e1e17db4b99f9dcd261a339cd.
//
// Solidity: event Slashed(address indexed account, uint256 amount)
func (_Staking *StakingFilterer) WatchSlashed(opts *bind.WatchOpts, sink chan<- *StakingSlashed, account []common.Address) (event.Subscription, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _Staking.contract.WatchLogs(opts, "Slashed", accountRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(StakingSlashed)
				if err := _Staking.contract.UnpackLog(event, "Slashed", log); err != nil {
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

// ParseSlashed is a log parse operation binding the contract event 0x4ed05e9673c26d2ed44f7ef6a7f2942df0ee3b5e1e17db4b99f9dcd261a339cd.
//
// Solidity: event Slashed(address indexed account, uint256 amount)
func (_Staking *StakingFilterer) ParseSlashed(log types.Log) (*StakingSlashed, error) {
	event := new(StakingSlashed)
	if err := _Staking.contract.UnpackLog(event, "Slashed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
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
