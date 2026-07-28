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

// FlowAssertionsMetaData contains all meta data concerning the FlowAssertions contract.
var FlowAssertionsMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"pool\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"reason\",\"type\":\"bytes\"}],\"name\":\"AaveV3AccountDataReadFailed\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"pool\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"actual\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minimum\",\"type\":\"uint256\"}],\"name\":\"AaveV3HealthFactorTooLow\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"reason\",\"type\":\"bytes\"}],\"name\":\"AssertionBalanceReadFailed\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"id\",\"type\":\"bytes32\"}],\"name\":\"AssertionCheckpointAlreadyExists\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"id\",\"type\":\"bytes32\"}],\"name\":\"AssertionCheckpointNotFound\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"id\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"expected\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"actual\",\"type\":\"address\"}],\"name\":\"AssertionCheckpointTokenMismatch\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"actual\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minimum\",\"type\":\"uint256\"}],\"name\":\"BalanceBelowMinimum\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"id\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"actualDelta\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maximumDelta\",\"type\":\"uint256\"}],\"name\":\"BalanceDecreaseTooLarge\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"id\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"actualDelta\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minimumDelta\",\"type\":\"uint256\"}],\"name\":\"BalanceIncreaseTooSmall\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"id\",\"type\":\"bytes32\"}],\"name\":\"InvalidAssertionCheckpointId\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"}],\"name\":\"InvalidAssertionToken\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"pool\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"minimumHealthFactor\",\"type\":\"uint256\"}],\"name\":\"assertAaveV3HealthFactorAtLeast\",\"outputs\":[],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"minimum\",\"type\":\"uint256\"}],\"name\":\"assertBalanceAtLeast\",\"outputs\":[],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"checkpointId\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"maximumDelta\",\"type\":\"uint256\"}],\"name\":\"assertBalanceDecreaseAtMost\",\"outputs\":[],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"checkpointId\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"minimumDelta\",\"type\":\"uint256\"}],\"name\":\"assertBalanceIncreaseAtLeast\",\"outputs\":[],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"checkpointId\",\"type\":\"bytes32\"}],\"name\":\"snapshotBalance\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// FlowAssertionsABI is the input ABI used to generate the binding from.
// Deprecated: Use FlowAssertionsMetaData.ABI instead.
var FlowAssertionsABI = FlowAssertionsMetaData.ABI

// FlowAssertions is an auto generated Go binding around an Ethereum contract.
type FlowAssertions struct {
	FlowAssertionsCaller     // Read-only binding to the contract
	FlowAssertionsTransactor // Write-only binding to the contract
	FlowAssertionsFilterer   // Log filterer for contract events
}

// FlowAssertionsCaller is an auto generated read-only Go binding around an Ethereum contract.
type FlowAssertionsCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// FlowAssertionsTransactor is an auto generated write-only Go binding around an Ethereum contract.
type FlowAssertionsTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// FlowAssertionsFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type FlowAssertionsFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// FlowAssertionsSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type FlowAssertionsSession struct {
	Contract     *FlowAssertions   // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// FlowAssertionsCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type FlowAssertionsCallerSession struct {
	Contract *FlowAssertionsCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts         // Call options to use throughout this session
}

// FlowAssertionsTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type FlowAssertionsTransactorSession struct {
	Contract     *FlowAssertionsTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// FlowAssertionsRaw is an auto generated low-level Go binding around an Ethereum contract.
type FlowAssertionsRaw struct {
	Contract *FlowAssertions // Generic contract binding to access the raw methods on
}

// FlowAssertionsCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type FlowAssertionsCallerRaw struct {
	Contract *FlowAssertionsCaller // Generic read-only contract binding to access the raw methods on
}

// FlowAssertionsTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type FlowAssertionsTransactorRaw struct {
	Contract *FlowAssertionsTransactor // Generic write-only contract binding to access the raw methods on
}

// NewFlowAssertions creates a new instance of FlowAssertions, bound to a specific deployed contract.
func NewFlowAssertions(address common.Address, backend bind.ContractBackend) (*FlowAssertions, error) {
	contract, err := bindFlowAssertions(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &FlowAssertions{FlowAssertionsCaller: FlowAssertionsCaller{contract: contract}, FlowAssertionsTransactor: FlowAssertionsTransactor{contract: contract}, FlowAssertionsFilterer: FlowAssertionsFilterer{contract: contract}}, nil
}

// NewFlowAssertionsCaller creates a new read-only instance of FlowAssertions, bound to a specific deployed contract.
func NewFlowAssertionsCaller(address common.Address, caller bind.ContractCaller) (*FlowAssertionsCaller, error) {
	contract, err := bindFlowAssertions(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &FlowAssertionsCaller{contract: contract}, nil
}

// NewFlowAssertionsTransactor creates a new write-only instance of FlowAssertions, bound to a specific deployed contract.
func NewFlowAssertionsTransactor(address common.Address, transactor bind.ContractTransactor) (*FlowAssertionsTransactor, error) {
	contract, err := bindFlowAssertions(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &FlowAssertionsTransactor{contract: contract}, nil
}

// NewFlowAssertionsFilterer creates a new log filterer instance of FlowAssertions, bound to a specific deployed contract.
func NewFlowAssertionsFilterer(address common.Address, filterer bind.ContractFilterer) (*FlowAssertionsFilterer, error) {
	contract, err := bindFlowAssertions(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &FlowAssertionsFilterer{contract: contract}, nil
}

// bindFlowAssertions binds a generic wrapper to an already deployed contract.
func bindFlowAssertions(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := FlowAssertionsMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_FlowAssertions *FlowAssertionsRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _FlowAssertions.Contract.FlowAssertionsCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_FlowAssertions *FlowAssertionsRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _FlowAssertions.Contract.FlowAssertionsTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_FlowAssertions *FlowAssertionsRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _FlowAssertions.Contract.FlowAssertionsTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_FlowAssertions *FlowAssertionsCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _FlowAssertions.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_FlowAssertions *FlowAssertionsTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _FlowAssertions.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_FlowAssertions *FlowAssertionsTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _FlowAssertions.Contract.contract.Transact(opts, method, params...)
}

// AssertAaveV3HealthFactorAtLeast is a free data retrieval call binding the contract method 0x4c91c4c7.
//
// Solidity: function assertAaveV3HealthFactorAtLeast(address pool, uint256 minimumHealthFactor) view returns()
func (_FlowAssertions *FlowAssertionsCaller) AssertAaveV3HealthFactorAtLeast(opts *bind.CallOpts, pool common.Address, minimumHealthFactor *big.Int) error {
	var out []interface{}
	err := _FlowAssertions.contract.Call(opts, &out, "assertAaveV3HealthFactorAtLeast", pool, minimumHealthFactor)

	if err != nil {
		return err
	}

	return err

}

// AssertAaveV3HealthFactorAtLeast is a free data retrieval call binding the contract method 0x4c91c4c7.
//
// Solidity: function assertAaveV3HealthFactorAtLeast(address pool, uint256 minimumHealthFactor) view returns()
func (_FlowAssertions *FlowAssertionsSession) AssertAaveV3HealthFactorAtLeast(pool common.Address, minimumHealthFactor *big.Int) error {
	return _FlowAssertions.Contract.AssertAaveV3HealthFactorAtLeast(&_FlowAssertions.CallOpts, pool, minimumHealthFactor)
}

// AssertAaveV3HealthFactorAtLeast is a free data retrieval call binding the contract method 0x4c91c4c7.
//
// Solidity: function assertAaveV3HealthFactorAtLeast(address pool, uint256 minimumHealthFactor) view returns()
func (_FlowAssertions *FlowAssertionsCallerSession) AssertAaveV3HealthFactorAtLeast(pool common.Address, minimumHealthFactor *big.Int) error {
	return _FlowAssertions.Contract.AssertAaveV3HealthFactorAtLeast(&_FlowAssertions.CallOpts, pool, minimumHealthFactor)
}

// AssertBalanceAtLeast is a free data retrieval call binding the contract method 0xebaf413b.
//
// Solidity: function assertBalanceAtLeast(address token, uint256 minimum) view returns()
func (_FlowAssertions *FlowAssertionsCaller) AssertBalanceAtLeast(opts *bind.CallOpts, token common.Address, minimum *big.Int) error {
	var out []interface{}
	err := _FlowAssertions.contract.Call(opts, &out, "assertBalanceAtLeast", token, minimum)

	if err != nil {
		return err
	}

	return err

}

// AssertBalanceAtLeast is a free data retrieval call binding the contract method 0xebaf413b.
//
// Solidity: function assertBalanceAtLeast(address token, uint256 minimum) view returns()
func (_FlowAssertions *FlowAssertionsSession) AssertBalanceAtLeast(token common.Address, minimum *big.Int) error {
	return _FlowAssertions.Contract.AssertBalanceAtLeast(&_FlowAssertions.CallOpts, token, minimum)
}

// AssertBalanceAtLeast is a free data retrieval call binding the contract method 0xebaf413b.
//
// Solidity: function assertBalanceAtLeast(address token, uint256 minimum) view returns()
func (_FlowAssertions *FlowAssertionsCallerSession) AssertBalanceAtLeast(token common.Address, minimum *big.Int) error {
	return _FlowAssertions.Contract.AssertBalanceAtLeast(&_FlowAssertions.CallOpts, token, minimum)
}

// AssertBalanceDecreaseAtMost is a free data retrieval call binding the contract method 0x834a785c.
//
// Solidity: function assertBalanceDecreaseAtMost(address token, bytes32 checkpointId, uint256 maximumDelta) view returns()
func (_FlowAssertions *FlowAssertionsCaller) AssertBalanceDecreaseAtMost(opts *bind.CallOpts, token common.Address, checkpointId [32]byte, maximumDelta *big.Int) error {
	var out []interface{}
	err := _FlowAssertions.contract.Call(opts, &out, "assertBalanceDecreaseAtMost", token, checkpointId, maximumDelta)

	if err != nil {
		return err
	}

	return err

}

// AssertBalanceDecreaseAtMost is a free data retrieval call binding the contract method 0x834a785c.
//
// Solidity: function assertBalanceDecreaseAtMost(address token, bytes32 checkpointId, uint256 maximumDelta) view returns()
func (_FlowAssertions *FlowAssertionsSession) AssertBalanceDecreaseAtMost(token common.Address, checkpointId [32]byte, maximumDelta *big.Int) error {
	return _FlowAssertions.Contract.AssertBalanceDecreaseAtMost(&_FlowAssertions.CallOpts, token, checkpointId, maximumDelta)
}

// AssertBalanceDecreaseAtMost is a free data retrieval call binding the contract method 0x834a785c.
//
// Solidity: function assertBalanceDecreaseAtMost(address token, bytes32 checkpointId, uint256 maximumDelta) view returns()
func (_FlowAssertions *FlowAssertionsCallerSession) AssertBalanceDecreaseAtMost(token common.Address, checkpointId [32]byte, maximumDelta *big.Int) error {
	return _FlowAssertions.Contract.AssertBalanceDecreaseAtMost(&_FlowAssertions.CallOpts, token, checkpointId, maximumDelta)
}

// AssertBalanceIncreaseAtLeast is a free data retrieval call binding the contract method 0xcf2c62bd.
//
// Solidity: function assertBalanceIncreaseAtLeast(address token, bytes32 checkpointId, uint256 minimumDelta) view returns()
func (_FlowAssertions *FlowAssertionsCaller) AssertBalanceIncreaseAtLeast(opts *bind.CallOpts, token common.Address, checkpointId [32]byte, minimumDelta *big.Int) error {
	var out []interface{}
	err := _FlowAssertions.contract.Call(opts, &out, "assertBalanceIncreaseAtLeast", token, checkpointId, minimumDelta)

	if err != nil {
		return err
	}

	return err

}

// AssertBalanceIncreaseAtLeast is a free data retrieval call binding the contract method 0xcf2c62bd.
//
// Solidity: function assertBalanceIncreaseAtLeast(address token, bytes32 checkpointId, uint256 minimumDelta) view returns()
func (_FlowAssertions *FlowAssertionsSession) AssertBalanceIncreaseAtLeast(token common.Address, checkpointId [32]byte, minimumDelta *big.Int) error {
	return _FlowAssertions.Contract.AssertBalanceIncreaseAtLeast(&_FlowAssertions.CallOpts, token, checkpointId, minimumDelta)
}

// AssertBalanceIncreaseAtLeast is a free data retrieval call binding the contract method 0xcf2c62bd.
//
// Solidity: function assertBalanceIncreaseAtLeast(address token, bytes32 checkpointId, uint256 minimumDelta) view returns()
func (_FlowAssertions *FlowAssertionsCallerSession) AssertBalanceIncreaseAtLeast(token common.Address, checkpointId [32]byte, minimumDelta *big.Int) error {
	return _FlowAssertions.Contract.AssertBalanceIncreaseAtLeast(&_FlowAssertions.CallOpts, token, checkpointId, minimumDelta)
}

// SnapshotBalance is a paid mutator transaction binding the contract method 0xf44bec62.
//
// Solidity: function snapshotBalance(address token, bytes32 checkpointId) returns()
func (_FlowAssertions *FlowAssertionsTransactor) SnapshotBalance(opts *bind.TransactOpts, token common.Address, checkpointId [32]byte) (*types.Transaction, error) {
	return _FlowAssertions.contract.Transact(opts, "snapshotBalance", token, checkpointId)
}

// SnapshotBalance is a paid mutator transaction binding the contract method 0xf44bec62.
//
// Solidity: function snapshotBalance(address token, bytes32 checkpointId) returns()
func (_FlowAssertions *FlowAssertionsSession) SnapshotBalance(token common.Address, checkpointId [32]byte) (*types.Transaction, error) {
	return _FlowAssertions.Contract.SnapshotBalance(&_FlowAssertions.TransactOpts, token, checkpointId)
}

// SnapshotBalance is a paid mutator transaction binding the contract method 0xf44bec62.
//
// Solidity: function snapshotBalance(address token, bytes32 checkpointId) returns()
func (_FlowAssertions *FlowAssertionsTransactorSession) SnapshotBalance(token common.Address, checkpointId [32]byte) (*types.Transaction, error) {
	return _FlowAssertions.Contract.SnapshotBalance(&_FlowAssertions.TransactOpts, token, checkpointId)
}
