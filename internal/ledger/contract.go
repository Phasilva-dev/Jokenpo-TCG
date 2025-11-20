// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package ledger

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
	_ = abi.ConvertType
)

// LedgerMetaData contains all meta data concerning the Ledger contract.
var LedgerMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"roomId\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"winnerId\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"loserId\",\"type\":\"string\"}],\"name\":\"AuditMatch\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"playerId\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"string[]\",\"name\":\"cardIds\",\"type\":\"string[]\"}],\"name\":\"AuditPackOpened\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"fromPlayer\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"toPlayer\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"cardId\",\"type\":\"string\"}],\"name\":\"AuditTrade\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"gameServerAuthority\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_playerId\",\"type\":\"string\"}],\"name\":\"getPlayerAssets\",\"outputs\":[{\"internalType\":\"string[]\",\"name\":\"\",\"type\":\"string[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_roomId\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"_winnerId\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"_loserId\",\"type\":\"string\"}],\"name\":\"logMatchResult\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_playerId\",\"type\":\"string\"},{\"internalType\":\"string[]\",\"name\":\"_cardIds\",\"type\":\"string[]\"}],\"name\":\"logPackOpening\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_fromPlayer\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"_toPlayer\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"_cardId\",\"type\":\"string\"}],\"name\":\"logTrade\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x6080604052348015600e575f5ffd5b50335f5f6101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055506115568061005b5f395ff3fe608060405234801561000f575f5ffd5b5060043610610055575f3560e01c80630ada582d146100595780631502cd0c1461007757806357da55ce146100a75780637908708b146100c3578063cf27ed0c146100df575b5f5ffd5b6100616100fb565b60405161006e9190610891565b60405180910390f35b610091600480360381019061008c91906109f7565b61011f565b60405161009e9190610b59565b60405180910390f35b6100c160048036038101906100bc9190610b79565b610211565b005b6100dd60048036038101906100d89190610b79565b610382565b005b6100f960048036038101906100f49190610cff565b610452565b005b5f5f9054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b60606001826040516101319190610daf565b9081526020016040518091039020805480602002602001604051908101604052809291908181526020015f905b82821015610206578382905f5260205f2001805461017b90610df2565b80601f01602080910402602001604051908101604052809291908181526020018280546101a790610df2565b80156101f25780601f106101c9576101008083540402835291602001916101f2565b820191905f5260205f20905b8154815290600101906020018083116101d557829003601f168201915b50505050508152602001906001019061015e565b505050509050919050565b5f5f9054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff161461029f576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161029690610ea2565b60405180910390fd5b6102a983826105a3565b6102e8576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016102df90610f30565b60405180910390fd5b6102f283826106f1565b6001826040516103029190610daf565b908152602001604051809103902081908060018154018082558091505060019003905f5260205f20015f90919091909150908161033f91906110f7565b507fcb6a9427f5732496720fa2f6427b1bc9a407a78d57f02a411a4f459a1d97c5c842848484604051610375949392919061120d565b60405180910390a1505050565b5f5f9054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610410576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161040790610ea2565b60405180910390fd5b7f459166290fcb68519a7a83e9074a5eddb1c5872f6494632588302fe07ab3ac6f42848484604051610445949392919061120d565b60405180910390a1505050565b5f5f9054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16146104e0576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016104d790610ea2565b60405180910390fd5b5f5f90505b8151811015610563576001836040516104fe9190610daf565b908152602001604051809103902082828151811061051f5761051e611265565b5b6020026020010151908060018154018082558091505060019003905f5260205f20015f90919091909150908161055591906110f7565b5080806001019150506104e5565b507f1e2592092e270aa65505d82cfc0297cf9860cc6bc5501cf9544edf647b891be142838360405161059793929190611292565b60405180910390a15050565b5f5f6001846040516105b59190610daf565b9081526020016040518091039020805480602002602001604051908101604052809291908181526020015f905b8282101561068a578382905f5260205f200180546105ff90610df2565b80601f016020809104026020016040519081016040528092919081815260200182805461062b90610df2565b80156106765780601f1061064d57610100808354040283529160200191610676565b820191905f5260205f20905b81548152906001019060200180831161065957829003601f168201915b5050505050815260200190600101906105e2565b5050505090505f5f90505b81518110156106e55783805190602001208282815181106106b9576106b8611265565b5b602002602001015180519060200120036106d8576001925050506106eb565b8080600101915050610695565b505f9150505b92915050565b5f6001836040516107029190610daf565b908152602001604051809103902090505f5f90505b81805490508110156107f357828051906020012082828154811061073e5761073d611265565b5b905f5260205f20016040516107539190611371565b6040518091039020036107e657816001838054905061077291906113b4565b8154811061078357610782611265565b5b905f5260205f200182828154811061079e5761079d611265565b5b905f5260205f200190816107b2919061140e565b50818054806107c4576107c36114f3565b5b600190038181905f5260205f20015f6107dd91906107fa565b905550506107f6565b8080600101915050610717565b50505b5050565b50805461080690610df2565b5f825580601f106108175750610834565b601f0160209004905f5260205f20908101906108339190610837565b5b50565b5b8082111561084e575f815f905550600101610838565b5090565b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f61087b82610852565b9050919050565b61088b81610871565b82525050565b5f6020820190506108a45f830184610882565b92915050565b5f604051905090565b5f5ffd5b5f5ffd5b5f5ffd5b5f5ffd5b5f601f19601f8301169050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b610909826108c3565b810181811067ffffffffffffffff82111715610928576109276108d3565b5b80604052505050565b5f61093a6108aa565b90506109468282610900565b919050565b5f67ffffffffffffffff821115610965576109646108d3565b5b61096e826108c3565b9050602081019050919050565b828183375f83830152505050565b5f61099b6109968461094b565b610931565b9050828152602081018484840111156109b7576109b66108bf565b5b6109c284828561097b565b509392505050565b5f82601f8301126109de576109dd6108bb565b5b81356109ee848260208601610989565b91505092915050565b5f60208284031215610a0c57610a0b6108b3565b5b5f82013567ffffffffffffffff811115610a2957610a286108b7565b5b610a35848285016109ca565b91505092915050565b5f81519050919050565b5f82825260208201905092915050565b5f819050602082019050919050565b5f81519050919050565b5f82825260208201905092915050565b8281835e5f83830152505050565b5f610a9982610a67565b610aa38185610a71565b9350610ab3818560208601610a81565b610abc816108c3565b840191505092915050565b5f610ad28383610a8f565b905092915050565b5f602082019050919050565b5f610af082610a3e565b610afa8185610a48565b935083602082028501610b0c85610a58565b805f5b85811015610b475784840389528151610b288582610ac7565b9450610b3383610ada565b925060208a01995050600181019050610b0f565b50829750879550505050505092915050565b5f6020820190508181035f830152610b718184610ae6565b905092915050565b5f5f5f60608486031215610b9057610b8f6108b3565b5b5f84013567ffffffffffffffff811115610bad57610bac6108b7565b5b610bb9868287016109ca565b935050602084013567ffffffffffffffff811115610bda57610bd96108b7565b5b610be6868287016109ca565b925050604084013567ffffffffffffffff811115610c0757610c066108b7565b5b610c13868287016109ca565b9150509250925092565b5f67ffffffffffffffff821115610c3757610c366108d3565b5b602082029050602081019050919050565b5f5ffd5b5f610c5e610c5984610c1d565b610931565b90508083825260208201905060208402830185811115610c8157610c80610c48565b5b835b81811015610cc857803567ffffffffffffffff811115610ca657610ca56108bb565b5b808601610cb389826109ca565b85526020850194505050602081019050610c83565b5050509392505050565b5f82601f830112610ce657610ce56108bb565b5b8135610cf6848260208601610c4c565b91505092915050565b5f5f60408385031215610d1557610d146108b3565b5b5f83013567ffffffffffffffff811115610d3257610d316108b7565b5b610d3e858286016109ca565b925050602083013567ffffffffffffffff811115610d5f57610d5e6108b7565b5b610d6b85828601610cd2565b9150509250929050565b5f81905092915050565b5f610d8982610a67565b610d938185610d75565b9350610da3818560208601610a81565b80840191505092915050565b5f610dba8284610d7f565b915081905092915050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602260045260245ffd5b5f6002820490506001821680610e0957607f821691505b602082108103610e1c57610e1b610dc5565b5b50919050565b5f82825260208201905092915050565b7f41636573736f206e656761646f3a204170656e6173206f2047616d65205365725f8201527f76657220706f646520726567697374726172206c6f67732e0000000000000000602082015250565b5f610e8c603883610e22565b9150610e9782610e32565b604082019050919050565b5f6020820190508181035f830152610eb981610e80565b9050919050565b7f4572726f2064652041756469746f7269613a204f206a6f6761646f72206465205f8201527f6f726967656d206e616f20706f73737569206f20617469766f2e000000000000602082015250565b5f610f1a603a83610e22565b9150610f2582610ec0565b604082019050919050565b5f6020820190508181035f830152610f4781610f0e565b9050919050565b5f819050815f5260205f209050919050565b5f6020601f8301049050919050565b5f82821b905092915050565b5f60088302610faa7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff82610f6f565b610fb48683610f6f565b95508019841693508086168417925050509392505050565b5f819050919050565b5f819050919050565b5f610ff8610ff3610fee84610fcc565b610fd5565b610fcc565b9050919050565b5f819050919050565b61101183610fde565b61102561101d82610fff565b848454610f7b565b825550505050565b5f5f905090565b61103c61102d565b611047818484611008565b505050565b5b8181101561106a5761105f5f82611034565b60018101905061104d565b5050565b601f8211156110af5761108081610f4e565b61108984610f60565b81016020851015611098578190505b6110ac6110a485610f60565b83018261104c565b50505b505050565b5f82821c905092915050565b5f6110cf5f19846008026110b4565b1980831691505092915050565b5f6110e783836110c0565b9150826002028217905092915050565b61110082610a67565b67ffffffffffffffff811115611119576111186108d3565b5b6111238254610df2565b61112e82828561106e565b5f60209050601f83116001811461115f575f841561114d578287015190505b61115785826110dc565b8655506111be565b601f19841661116d86610f4e565b5f5b828110156111945784890151825560018201915060208501945060208101905061116f565b868310156111b157848901516111ad601f8916826110c0565b8355505b6001600288020188555050505b505050505050565b6111cf81610fcc565b82525050565b5f6111df82610a67565b6111e98185610e22565b93506111f9818560208601610a81565b611202816108c3565b840191505092915050565b5f6080820190506112205f8301876111c6565b818103602083015261123281866111d5565b9050818103604083015261124681856111d5565b9050818103606083015261125a81846111d5565b905095945050505050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52603260045260245ffd5b5f6060820190506112a55f8301866111c6565b81810360208301526112b781856111d5565b905081810360408301526112cb8184610ae6565b9050949350505050565b5f81905092915050565b5f819050815f5260205f209050919050565b5f81546112fd81610df2565b61130781866112d5565b9450600182165f8114611321576001811461133657611368565b60ff1983168652811515820286019350611368565b61133f856112df565b5f5b8381101561136057815481890152600182019150602081019050611341565b838801955050505b50505092915050565b5f61137c82846112f1565b915081905092915050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52601160045260245ffd5b5f6113be82610fcc565b91506113c983610fcc565b92508282039050818111156113e1576113e0611387565b5b92915050565b5f815490506113f581610df2565b9050919050565b5f819050815f5260205f209050919050565b81810361141c5750506114f1565b611425826113e7565b67ffffffffffffffff81111561143e5761143d6108d3565b5b6114488254610df2565b61145382828561106e565b5f601f831160018114611480575f841561146e578287015490505b61147885826110dc565b8655506114ea565b601f19841661148e876113fc565b965061149986610f4e565b5f5b828110156114c05784890154825560018201915060018501945060208101905061149b565b868310156114dd57848901546114d9601f8916826110c0565b8355505b6001600288020188555050505b5050505050505b565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52603160045260245ffdfea2646970667358221220e77b505c556e21eac73ff24c2752e664684e9b9c5b9d0c7a5a829454e39fbde264736f6c634300081e0033",
}

// LedgerABI is the input ABI used to generate the binding from.
// Deprecated: Use LedgerMetaData.ABI instead.
var LedgerABI = LedgerMetaData.ABI

// LedgerBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use LedgerMetaData.Bin instead.
var LedgerBin = LedgerMetaData.Bin

// DeployLedger deploys a new Ethereum contract, binding an instance of Ledger to it.
func DeployLedger(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Ledger, error) {
	parsed, err := LedgerMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(LedgerBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Ledger{LedgerCaller: LedgerCaller{contract: contract}, LedgerTransactor: LedgerTransactor{contract: contract}, LedgerFilterer: LedgerFilterer{contract: contract}}, nil
}

// Ledger is an auto generated Go binding around an Ethereum contract.
type Ledger struct {
	LedgerCaller     // Read-only binding to the contract
	LedgerTransactor // Write-only binding to the contract
	LedgerFilterer   // Log filterer for contract events
}

// LedgerCaller is an auto generated read-only Go binding around an Ethereum contract.
type LedgerCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// LedgerTransactor is an auto generated write-only Go binding around an Ethereum contract.
type LedgerTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// LedgerFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type LedgerFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// LedgerSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type LedgerSession struct {
	Contract     *Ledger           // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// LedgerCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type LedgerCallerSession struct {
	Contract *LedgerCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// LedgerTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type LedgerTransactorSession struct {
	Contract     *LedgerTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// LedgerRaw is an auto generated low-level Go binding around an Ethereum contract.
type LedgerRaw struct {
	Contract *Ledger // Generic contract binding to access the raw methods on
}

// LedgerCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type LedgerCallerRaw struct {
	Contract *LedgerCaller // Generic read-only contract binding to access the raw methods on
}

// LedgerTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type LedgerTransactorRaw struct {
	Contract *LedgerTransactor // Generic write-only contract binding to access the raw methods on
}

// NewLedger creates a new instance of Ledger, bound to a specific deployed contract.
func NewLedger(address common.Address, backend bind.ContractBackend) (*Ledger, error) {
	contract, err := bindLedger(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Ledger{LedgerCaller: LedgerCaller{contract: contract}, LedgerTransactor: LedgerTransactor{contract: contract}, LedgerFilterer: LedgerFilterer{contract: contract}}, nil
}

// NewLedgerCaller creates a new read-only instance of Ledger, bound to a specific deployed contract.
func NewLedgerCaller(address common.Address, caller bind.ContractCaller) (*LedgerCaller, error) {
	contract, err := bindLedger(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &LedgerCaller{contract: contract}, nil
}

// NewLedgerTransactor creates a new write-only instance of Ledger, bound to a specific deployed contract.
func NewLedgerTransactor(address common.Address, transactor bind.ContractTransactor) (*LedgerTransactor, error) {
	contract, err := bindLedger(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &LedgerTransactor{contract: contract}, nil
}

// NewLedgerFilterer creates a new log filterer instance of Ledger, bound to a specific deployed contract.
func NewLedgerFilterer(address common.Address, filterer bind.ContractFilterer) (*LedgerFilterer, error) {
	contract, err := bindLedger(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &LedgerFilterer{contract: contract}, nil
}

// bindLedger binds a generic wrapper to an already deployed contract.
func bindLedger(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := LedgerMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Ledger *LedgerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Ledger.Contract.LedgerCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Ledger *LedgerRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Ledger.Contract.LedgerTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Ledger *LedgerRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Ledger.Contract.LedgerTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Ledger *LedgerCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Ledger.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Ledger *LedgerTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Ledger.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Ledger *LedgerTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Ledger.Contract.contract.Transact(opts, method, params...)
}

// GameServerAuthority is a free data retrieval call binding the contract method 0x0ada582d.
//
// Solidity: function gameServerAuthority() view returns(address)
func (_Ledger *LedgerCaller) GameServerAuthority(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Ledger.contract.Call(opts, &out, "gameServerAuthority")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GameServerAuthority is a free data retrieval call binding the contract method 0x0ada582d.
//
// Solidity: function gameServerAuthority() view returns(address)
func (_Ledger *LedgerSession) GameServerAuthority() (common.Address, error) {
	return _Ledger.Contract.GameServerAuthority(&_Ledger.CallOpts)
}

// GameServerAuthority is a free data retrieval call binding the contract method 0x0ada582d.
//
// Solidity: function gameServerAuthority() view returns(address)
func (_Ledger *LedgerCallerSession) GameServerAuthority() (common.Address, error) {
	return _Ledger.Contract.GameServerAuthority(&_Ledger.CallOpts)
}

// GetPlayerAssets is a free data retrieval call binding the contract method 0x1502cd0c.
//
// Solidity: function getPlayerAssets(string _playerId) view returns(string[])
func (_Ledger *LedgerCaller) GetPlayerAssets(opts *bind.CallOpts, _playerId string) ([]string, error) {
	var out []interface{}
	err := _Ledger.contract.Call(opts, &out, "getPlayerAssets", _playerId)

	if err != nil {
		return *new([]string), err
	}

	out0 := *abi.ConvertType(out[0], new([]string)).(*[]string)

	return out0, err

}

// GetPlayerAssets is a free data retrieval call binding the contract method 0x1502cd0c.
//
// Solidity: function getPlayerAssets(string _playerId) view returns(string[])
func (_Ledger *LedgerSession) GetPlayerAssets(_playerId string) ([]string, error) {
	return _Ledger.Contract.GetPlayerAssets(&_Ledger.CallOpts, _playerId)
}

// GetPlayerAssets is a free data retrieval call binding the contract method 0x1502cd0c.
//
// Solidity: function getPlayerAssets(string _playerId) view returns(string[])
func (_Ledger *LedgerCallerSession) GetPlayerAssets(_playerId string) ([]string, error) {
	return _Ledger.Contract.GetPlayerAssets(&_Ledger.CallOpts, _playerId)
}

// LogMatchResult is a paid mutator transaction binding the contract method 0x7908708b.
//
// Solidity: function logMatchResult(string _roomId, string _winnerId, string _loserId) returns()
func (_Ledger *LedgerTransactor) LogMatchResult(opts *bind.TransactOpts, _roomId string, _winnerId string, _loserId string) (*types.Transaction, error) {
	return _Ledger.contract.Transact(opts, "logMatchResult", _roomId, _winnerId, _loserId)
}

// LogMatchResult is a paid mutator transaction binding the contract method 0x7908708b.
//
// Solidity: function logMatchResult(string _roomId, string _winnerId, string _loserId) returns()
func (_Ledger *LedgerSession) LogMatchResult(_roomId string, _winnerId string, _loserId string) (*types.Transaction, error) {
	return _Ledger.Contract.LogMatchResult(&_Ledger.TransactOpts, _roomId, _winnerId, _loserId)
}

// LogMatchResult is a paid mutator transaction binding the contract method 0x7908708b.
//
// Solidity: function logMatchResult(string _roomId, string _winnerId, string _loserId) returns()
func (_Ledger *LedgerTransactorSession) LogMatchResult(_roomId string, _winnerId string, _loserId string) (*types.Transaction, error) {
	return _Ledger.Contract.LogMatchResult(&_Ledger.TransactOpts, _roomId, _winnerId, _loserId)
}

// LogPackOpening is a paid mutator transaction binding the contract method 0xcf27ed0c.
//
// Solidity: function logPackOpening(string _playerId, string[] _cardIds) returns()
func (_Ledger *LedgerTransactor) LogPackOpening(opts *bind.TransactOpts, _playerId string, _cardIds []string) (*types.Transaction, error) {
	return _Ledger.contract.Transact(opts, "logPackOpening", _playerId, _cardIds)
}

// LogPackOpening is a paid mutator transaction binding the contract method 0xcf27ed0c.
//
// Solidity: function logPackOpening(string _playerId, string[] _cardIds) returns()
func (_Ledger *LedgerSession) LogPackOpening(_playerId string, _cardIds []string) (*types.Transaction, error) {
	return _Ledger.Contract.LogPackOpening(&_Ledger.TransactOpts, _playerId, _cardIds)
}

// LogPackOpening is a paid mutator transaction binding the contract method 0xcf27ed0c.
//
// Solidity: function logPackOpening(string _playerId, string[] _cardIds) returns()
func (_Ledger *LedgerTransactorSession) LogPackOpening(_playerId string, _cardIds []string) (*types.Transaction, error) {
	return _Ledger.Contract.LogPackOpening(&_Ledger.TransactOpts, _playerId, _cardIds)
}

// LogTrade is a paid mutator transaction binding the contract method 0x57da55ce.
//
// Solidity: function logTrade(string _fromPlayer, string _toPlayer, string _cardId) returns()
func (_Ledger *LedgerTransactor) LogTrade(opts *bind.TransactOpts, _fromPlayer string, _toPlayer string, _cardId string) (*types.Transaction, error) {
	return _Ledger.contract.Transact(opts, "logTrade", _fromPlayer, _toPlayer, _cardId)
}

// LogTrade is a paid mutator transaction binding the contract method 0x57da55ce.
//
// Solidity: function logTrade(string _fromPlayer, string _toPlayer, string _cardId) returns()
func (_Ledger *LedgerSession) LogTrade(_fromPlayer string, _toPlayer string, _cardId string) (*types.Transaction, error) {
	return _Ledger.Contract.LogTrade(&_Ledger.TransactOpts, _fromPlayer, _toPlayer, _cardId)
}

// LogTrade is a paid mutator transaction binding the contract method 0x57da55ce.
//
// Solidity: function logTrade(string _fromPlayer, string _toPlayer, string _cardId) returns()
func (_Ledger *LedgerTransactorSession) LogTrade(_fromPlayer string, _toPlayer string, _cardId string) (*types.Transaction, error) {
	return _Ledger.Contract.LogTrade(&_Ledger.TransactOpts, _fromPlayer, _toPlayer, _cardId)
}

// LedgerAuditMatchIterator is returned from FilterAuditMatch and is used to iterate over the raw logs and unpacked data for AuditMatch events raised by the Ledger contract.
type LedgerAuditMatchIterator struct {
	Event *LedgerAuditMatch // Event containing the contract specifics and raw log

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
func (it *LedgerAuditMatchIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(LedgerAuditMatch)
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
		it.Event = new(LedgerAuditMatch)
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
func (it *LedgerAuditMatchIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *LedgerAuditMatchIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// LedgerAuditMatch represents a AuditMatch event raised by the Ledger contract.
type LedgerAuditMatch struct {
	Timestamp *big.Int
	RoomId    string
	WinnerId  string
	LoserId   string
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterAuditMatch is a free log retrieval operation binding the contract event 0x459166290fcb68519a7a83e9074a5eddb1c5872f6494632588302fe07ab3ac6f.
//
// Solidity: event AuditMatch(uint256 timestamp, string roomId, string winnerId, string loserId)
func (_Ledger *LedgerFilterer) FilterAuditMatch(opts *bind.FilterOpts) (*LedgerAuditMatchIterator, error) {

	logs, sub, err := _Ledger.contract.FilterLogs(opts, "AuditMatch")
	if err != nil {
		return nil, err
	}
	return &LedgerAuditMatchIterator{contract: _Ledger.contract, event: "AuditMatch", logs: logs, sub: sub}, nil
}

// WatchAuditMatch is a free log subscription operation binding the contract event 0x459166290fcb68519a7a83e9074a5eddb1c5872f6494632588302fe07ab3ac6f.
//
// Solidity: event AuditMatch(uint256 timestamp, string roomId, string winnerId, string loserId)
func (_Ledger *LedgerFilterer) WatchAuditMatch(opts *bind.WatchOpts, sink chan<- *LedgerAuditMatch) (event.Subscription, error) {

	logs, sub, err := _Ledger.contract.WatchLogs(opts, "AuditMatch")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(LedgerAuditMatch)
				if err := _Ledger.contract.UnpackLog(event, "AuditMatch", log); err != nil {
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

// ParseAuditMatch is a log parse operation binding the contract event 0x459166290fcb68519a7a83e9074a5eddb1c5872f6494632588302fe07ab3ac6f.
//
// Solidity: event AuditMatch(uint256 timestamp, string roomId, string winnerId, string loserId)
func (_Ledger *LedgerFilterer) ParseAuditMatch(log types.Log) (*LedgerAuditMatch, error) {
	event := new(LedgerAuditMatch)
	if err := _Ledger.contract.UnpackLog(event, "AuditMatch", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// LedgerAuditPackOpenedIterator is returned from FilterAuditPackOpened and is used to iterate over the raw logs and unpacked data for AuditPackOpened events raised by the Ledger contract.
type LedgerAuditPackOpenedIterator struct {
	Event *LedgerAuditPackOpened // Event containing the contract specifics and raw log

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
func (it *LedgerAuditPackOpenedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(LedgerAuditPackOpened)
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
		it.Event = new(LedgerAuditPackOpened)
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
func (it *LedgerAuditPackOpenedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *LedgerAuditPackOpenedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// LedgerAuditPackOpened represents a AuditPackOpened event raised by the Ledger contract.
type LedgerAuditPackOpened struct {
	Timestamp *big.Int
	PlayerId  string
	CardIds   []string
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterAuditPackOpened is a free log retrieval operation binding the contract event 0x1e2592092e270aa65505d82cfc0297cf9860cc6bc5501cf9544edf647b891be1.
//
// Solidity: event AuditPackOpened(uint256 timestamp, string playerId, string[] cardIds)
func (_Ledger *LedgerFilterer) FilterAuditPackOpened(opts *bind.FilterOpts) (*LedgerAuditPackOpenedIterator, error) {

	logs, sub, err := _Ledger.contract.FilterLogs(opts, "AuditPackOpened")
	if err != nil {
		return nil, err
	}
	return &LedgerAuditPackOpenedIterator{contract: _Ledger.contract, event: "AuditPackOpened", logs: logs, sub: sub}, nil
}

// WatchAuditPackOpened is a free log subscription operation binding the contract event 0x1e2592092e270aa65505d82cfc0297cf9860cc6bc5501cf9544edf647b891be1.
//
// Solidity: event AuditPackOpened(uint256 timestamp, string playerId, string[] cardIds)
func (_Ledger *LedgerFilterer) WatchAuditPackOpened(opts *bind.WatchOpts, sink chan<- *LedgerAuditPackOpened) (event.Subscription, error) {

	logs, sub, err := _Ledger.contract.WatchLogs(opts, "AuditPackOpened")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(LedgerAuditPackOpened)
				if err := _Ledger.contract.UnpackLog(event, "AuditPackOpened", log); err != nil {
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

// ParseAuditPackOpened is a log parse operation binding the contract event 0x1e2592092e270aa65505d82cfc0297cf9860cc6bc5501cf9544edf647b891be1.
//
// Solidity: event AuditPackOpened(uint256 timestamp, string playerId, string[] cardIds)
func (_Ledger *LedgerFilterer) ParseAuditPackOpened(log types.Log) (*LedgerAuditPackOpened, error) {
	event := new(LedgerAuditPackOpened)
	if err := _Ledger.contract.UnpackLog(event, "AuditPackOpened", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// LedgerAuditTradeIterator is returned from FilterAuditTrade and is used to iterate over the raw logs and unpacked data for AuditTrade events raised by the Ledger contract.
type LedgerAuditTradeIterator struct {
	Event *LedgerAuditTrade // Event containing the contract specifics and raw log

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
func (it *LedgerAuditTradeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(LedgerAuditTrade)
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
		it.Event = new(LedgerAuditTrade)
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
func (it *LedgerAuditTradeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *LedgerAuditTradeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// LedgerAuditTrade represents a AuditTrade event raised by the Ledger contract.
type LedgerAuditTrade struct {
	Timestamp  *big.Int
	FromPlayer string
	ToPlayer   string
	CardId     string
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterAuditTrade is a free log retrieval operation binding the contract event 0xcb6a9427f5732496720fa2f6427b1bc9a407a78d57f02a411a4f459a1d97c5c8.
//
// Solidity: event AuditTrade(uint256 timestamp, string fromPlayer, string toPlayer, string cardId)
func (_Ledger *LedgerFilterer) FilterAuditTrade(opts *bind.FilterOpts) (*LedgerAuditTradeIterator, error) {

	logs, sub, err := _Ledger.contract.FilterLogs(opts, "AuditTrade")
	if err != nil {
		return nil, err
	}
	return &LedgerAuditTradeIterator{contract: _Ledger.contract, event: "AuditTrade", logs: logs, sub: sub}, nil
}

// WatchAuditTrade is a free log subscription operation binding the contract event 0xcb6a9427f5732496720fa2f6427b1bc9a407a78d57f02a411a4f459a1d97c5c8.
//
// Solidity: event AuditTrade(uint256 timestamp, string fromPlayer, string toPlayer, string cardId)
func (_Ledger *LedgerFilterer) WatchAuditTrade(opts *bind.WatchOpts, sink chan<- *LedgerAuditTrade) (event.Subscription, error) {

	logs, sub, err := _Ledger.contract.WatchLogs(opts, "AuditTrade")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(LedgerAuditTrade)
				if err := _Ledger.contract.UnpackLog(event, "AuditTrade", log); err != nil {
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

// ParseAuditTrade is a log parse operation binding the contract event 0xcb6a9427f5732496720fa2f6427b1bc9a407a78d57f02a411a4f459a1d97c5c8.
//
// Solidity: event AuditTrade(uint256 timestamp, string fromPlayer, string toPlayer, string cardId)
func (_Ledger *LedgerFilterer) ParseAuditTrade(log types.Log) (*LedgerAuditTrade, error) {
	event := new(LedgerAuditTrade)
	if err := _Ledger.contract.UnpackLog(event, "AuditTrade", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
