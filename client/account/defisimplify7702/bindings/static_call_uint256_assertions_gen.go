// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package bindings

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

// StaticCallUint256AssertionsMetaData contains all meta data concerning the StaticCallUint256Assertions contract.
var StaticCallUint256AssertionsMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"target\",\"type\":\"address\"},{\"internalType\":\"bytes4\",\"name\":\"selector\",\"type\":\"bytes4\"},{\"internalType\":\"uint256\",\"name\":\"accountOffset\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"reason\",\"type\":\"bytes\"}],\"name\":\"AssertionStaticCallFailed\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"offset\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"dataLength\",\"type\":\"uint256\"}],\"name\":\"InvalidAssertionAccountOffset\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"dataLength\",\"type\":\"uint256\"}],\"name\":\"InvalidAssertionCallData\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"offset\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"returnDataLength\",\"type\":\"uint256\"}],\"name\":\"InvalidAssertionReturnOffset\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"target\",\"type\":\"address\"}],\"name\":\"InvalidAssertionTarget\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"target\",\"type\":\"address\"},{\"internalType\":\"bytes4\",\"name\":\"selector\",\"type\":\"bytes4\"},{\"internalType\":\"uint256\",\"name\":\"accountOffset\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"returnOffset\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"actual\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maximum\",\"type\":\"uint256\"}],\"name\":\"StaticCallUint256AboveMaximum\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"target\",\"type\":\"address\"},{\"internalType\":\"bytes4\",\"name\":\"selector\",\"type\":\"bytes4\"},{\"internalType\":\"uint256\",\"name\":\"accountOffset\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"returnOffset\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"actual\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minimum\",\"type\":\"uint256\"}],\"name\":\"StaticCallUint256BelowMinimum\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"target\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"internalType\":\"uint32\",\"name\":\"accountOffset\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"returnOffset\",\"type\":\"uint32\"},{\"internalType\":\"uint256\",\"name\":\"minimum\",\"type\":\"uint256\"}],\"name\":\"assertStaticCallUint256AtLeast\",\"outputs\":[],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"target\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"internalType\":\"uint32\",\"name\":\"accountOffset\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"returnOffset\",\"type\":\"uint32\"},{\"internalType\":\"uint256\",\"name\":\"maximum\",\"type\":\"uint256\"}],\"name\":\"assertStaticCallUint256AtMost\",\"outputs\":[],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// StaticCallUint256AssertionsABI is the input ABI used to generate the binding from.
// Deprecated: Use StaticCallUint256AssertionsMetaData.ABI instead.
var StaticCallUint256AssertionsABI = StaticCallUint256AssertionsMetaData.ABI

// StaticCallUint256Assertions is an auto generated Go binding around an Ethereum contract.
type StaticCallUint256Assertions struct {
	StaticCallUint256AssertionsCaller     // Read-only binding to the contract
	StaticCallUint256AssertionsTransactor // Write-only binding to the contract
	StaticCallUint256AssertionsFilterer   // Log filterer for contract events
}

// StaticCallUint256AssertionsCaller is an auto generated read-only Go binding around an Ethereum contract.
type StaticCallUint256AssertionsCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// StaticCallUint256AssertionsTransactor is an auto generated write-only Go binding around an Ethereum contract.
type StaticCallUint256AssertionsTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// StaticCallUint256AssertionsFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type StaticCallUint256AssertionsFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// StaticCallUint256AssertionsSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type StaticCallUint256AssertionsSession struct {
	Contract     *StaticCallUint256Assertions // Generic contract binding to set the session for
	CallOpts     bind.CallOpts                // Call options to use throughout this session
	TransactOpts bind.TransactOpts            // Transaction auth options to use throughout this session
}

// StaticCallUint256AssertionsCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type StaticCallUint256AssertionsCallerSession struct {
	Contract *StaticCallUint256AssertionsCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                      // Call options to use throughout this session
}

// StaticCallUint256AssertionsTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type StaticCallUint256AssertionsTransactorSession struct {
	Contract     *StaticCallUint256AssertionsTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                      // Transaction auth options to use throughout this session
}

// StaticCallUint256AssertionsRaw is an auto generated low-level Go binding around an Ethereum contract.
type StaticCallUint256AssertionsRaw struct {
	Contract *StaticCallUint256Assertions // Generic contract binding to access the raw methods on
}

// StaticCallUint256AssertionsCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type StaticCallUint256AssertionsCallerRaw struct {
	Contract *StaticCallUint256AssertionsCaller // Generic read-only contract binding to access the raw methods on
}

// StaticCallUint256AssertionsTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type StaticCallUint256AssertionsTransactorRaw struct {
	Contract *StaticCallUint256AssertionsTransactor // Generic write-only contract binding to access the raw methods on
}

// NewStaticCallUint256Assertions creates a new instance of StaticCallUint256Assertions, bound to a specific deployed contract.
func NewStaticCallUint256Assertions(address common.Address, backend bind.ContractBackend) (*StaticCallUint256Assertions, error) {
	contract, err := bindStaticCallUint256Assertions(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &StaticCallUint256Assertions{StaticCallUint256AssertionsCaller: StaticCallUint256AssertionsCaller{contract: contract}, StaticCallUint256AssertionsTransactor: StaticCallUint256AssertionsTransactor{contract: contract}, StaticCallUint256AssertionsFilterer: StaticCallUint256AssertionsFilterer{contract: contract}}, nil
}

// NewStaticCallUint256AssertionsCaller creates a new read-only instance of StaticCallUint256Assertions, bound to a specific deployed contract.
func NewStaticCallUint256AssertionsCaller(address common.Address, caller bind.ContractCaller) (*StaticCallUint256AssertionsCaller, error) {
	contract, err := bindStaticCallUint256Assertions(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &StaticCallUint256AssertionsCaller{contract: contract}, nil
}

// NewStaticCallUint256AssertionsTransactor creates a new write-only instance of StaticCallUint256Assertions, bound to a specific deployed contract.
func NewStaticCallUint256AssertionsTransactor(address common.Address, transactor bind.ContractTransactor) (*StaticCallUint256AssertionsTransactor, error) {
	contract, err := bindStaticCallUint256Assertions(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &StaticCallUint256AssertionsTransactor{contract: contract}, nil
}

// NewStaticCallUint256AssertionsFilterer creates a new log filterer instance of StaticCallUint256Assertions, bound to a specific deployed contract.
func NewStaticCallUint256AssertionsFilterer(address common.Address, filterer bind.ContractFilterer) (*StaticCallUint256AssertionsFilterer, error) {
	contract, err := bindStaticCallUint256Assertions(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &StaticCallUint256AssertionsFilterer{contract: contract}, nil
}

// bindStaticCallUint256Assertions binds a generic wrapper to an already deployed contract.
func bindStaticCallUint256Assertions(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := StaticCallUint256AssertionsMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_StaticCallUint256Assertions *StaticCallUint256AssertionsRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _StaticCallUint256Assertions.Contract.StaticCallUint256AssertionsCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_StaticCallUint256Assertions *StaticCallUint256AssertionsRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _StaticCallUint256Assertions.Contract.StaticCallUint256AssertionsTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_StaticCallUint256Assertions *StaticCallUint256AssertionsRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _StaticCallUint256Assertions.Contract.StaticCallUint256AssertionsTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_StaticCallUint256Assertions *StaticCallUint256AssertionsCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _StaticCallUint256Assertions.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_StaticCallUint256Assertions *StaticCallUint256AssertionsTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _StaticCallUint256Assertions.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_StaticCallUint256Assertions *StaticCallUint256AssertionsTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _StaticCallUint256Assertions.Contract.contract.Transact(opts, method, params...)
}

// AssertStaticCallUint256AtLeast is a free data retrieval call binding the contract method 0x27d5c024.
//
// Solidity: function assertStaticCallUint256AtLeast(address target, bytes data, uint32 accountOffset, uint32 returnOffset, uint256 minimum) view returns()
func (_StaticCallUint256Assertions *StaticCallUint256AssertionsCaller) AssertStaticCallUint256AtLeast(opts *bind.CallOpts, target common.Address, data []byte, accountOffset uint32, returnOffset uint32, minimum *big.Int) error {
	var out []interface{}
	err := _StaticCallUint256Assertions.contract.Call(opts, &out, "assertStaticCallUint256AtLeast", target, data, accountOffset, returnOffset, minimum)

	if err != nil {
		return err
	}

	return err

}

// AssertStaticCallUint256AtLeast is a free data retrieval call binding the contract method 0x27d5c024.
//
// Solidity: function assertStaticCallUint256AtLeast(address target, bytes data, uint32 accountOffset, uint32 returnOffset, uint256 minimum) view returns()
func (_StaticCallUint256Assertions *StaticCallUint256AssertionsSession) AssertStaticCallUint256AtLeast(target common.Address, data []byte, accountOffset uint32, returnOffset uint32, minimum *big.Int) error {
	return _StaticCallUint256Assertions.Contract.AssertStaticCallUint256AtLeast(&_StaticCallUint256Assertions.CallOpts, target, data, accountOffset, returnOffset, minimum)
}

// AssertStaticCallUint256AtLeast is a free data retrieval call binding the contract method 0x27d5c024.
//
// Solidity: function assertStaticCallUint256AtLeast(address target, bytes data, uint32 accountOffset, uint32 returnOffset, uint256 minimum) view returns()
func (_StaticCallUint256Assertions *StaticCallUint256AssertionsCallerSession) AssertStaticCallUint256AtLeast(target common.Address, data []byte, accountOffset uint32, returnOffset uint32, minimum *big.Int) error {
	return _StaticCallUint256Assertions.Contract.AssertStaticCallUint256AtLeast(&_StaticCallUint256Assertions.CallOpts, target, data, accountOffset, returnOffset, minimum)
}

// AssertStaticCallUint256AtMost is a free data retrieval call binding the contract method 0x6ac43f19.
//
// Solidity: function assertStaticCallUint256AtMost(address target, bytes data, uint32 accountOffset, uint32 returnOffset, uint256 maximum) view returns()
func (_StaticCallUint256Assertions *StaticCallUint256AssertionsCaller) AssertStaticCallUint256AtMost(opts *bind.CallOpts, target common.Address, data []byte, accountOffset uint32, returnOffset uint32, maximum *big.Int) error {
	var out []interface{}
	err := _StaticCallUint256Assertions.contract.Call(opts, &out, "assertStaticCallUint256AtMost", target, data, accountOffset, returnOffset, maximum)

	if err != nil {
		return err
	}

	return err

}

// AssertStaticCallUint256AtMost is a free data retrieval call binding the contract method 0x6ac43f19.
//
// Solidity: function assertStaticCallUint256AtMost(address target, bytes data, uint32 accountOffset, uint32 returnOffset, uint256 maximum) view returns()
func (_StaticCallUint256Assertions *StaticCallUint256AssertionsSession) AssertStaticCallUint256AtMost(target common.Address, data []byte, accountOffset uint32, returnOffset uint32, maximum *big.Int) error {
	return _StaticCallUint256Assertions.Contract.AssertStaticCallUint256AtMost(&_StaticCallUint256Assertions.CallOpts, target, data, accountOffset, returnOffset, maximum)
}

// AssertStaticCallUint256AtMost is a free data retrieval call binding the contract method 0x6ac43f19.
//
// Solidity: function assertStaticCallUint256AtMost(address target, bytes data, uint32 accountOffset, uint32 returnOffset, uint256 maximum) view returns()
func (_StaticCallUint256Assertions *StaticCallUint256AssertionsCallerSession) AssertStaticCallUint256AtMost(target common.Address, data []byte, accountOffset uint32, returnOffset uint32, maximum *big.Int) error {
	return _StaticCallUint256Assertions.Contract.AssertStaticCallUint256AtMost(&_StaticCallUint256Assertions.CallOpts, target, data, accountOffset, returnOffset, maximum)
}
