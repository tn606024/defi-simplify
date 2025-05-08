// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package aave

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

// WrappedTokenGatewayV3MetaData contains all meta data concerning the WrappedTokenGatewayV3 contract.
var WrappedTokenGatewayV3MetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"weth\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"contractIPool\",\"name\":\"pool\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"stateMutability\":\"payable\",\"type\":\"fallback\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint16\",\"name\":\"referralCode\",\"type\":\"uint16\"}],\"name\":\"borrowETH\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"onBehalfOf\",\"type\":\"address\"},{\"internalType\":\"uint16\",\"name\":\"referralCode\",\"type\":\"uint16\"}],\"name\":\"depositETH\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"emergencyEtherTransfer\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"emergencyTokenTransfer\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getWETHAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"rateMode\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"onBehalfOf\",\"type\":\"address\"}],\"name\":\"repayETH\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"}],\"name\":\"withdrawETH\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"deadline\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"permitV\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"permitR\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"permitS\",\"type\":\"bytes32\"}],\"name\":\"withdrawETHWithPermit\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"stateMutability\":\"payable\",\"type\":\"receive\"}]",
}

// WrappedTokenGatewayV3ABI is the input ABI used to generate the binding from.
// Deprecated: Use WrappedTokenGatewayV3MetaData.ABI instead.
var WrappedTokenGatewayV3ABI = WrappedTokenGatewayV3MetaData.ABI

// WrappedTokenGatewayV3 is an auto generated Go binding around an Ethereum contract.
type WrappedTokenGatewayV3 struct {
	WrappedTokenGatewayV3Caller     // Read-only binding to the contract
	WrappedTokenGatewayV3Transactor // Write-only binding to the contract
	WrappedTokenGatewayV3Filterer   // Log filterer for contract events
}

// WrappedTokenGatewayV3Caller is an auto generated read-only Go binding around an Ethereum contract.
type WrappedTokenGatewayV3Caller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// WrappedTokenGatewayV3Transactor is an auto generated write-only Go binding around an Ethereum contract.
type WrappedTokenGatewayV3Transactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// WrappedTokenGatewayV3Filterer is an auto generated log filtering Go binding around an Ethereum contract events.
type WrappedTokenGatewayV3Filterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// WrappedTokenGatewayV3Session is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type WrappedTokenGatewayV3Session struct {
	Contract     *WrappedTokenGatewayV3 // Generic contract binding to set the session for
	CallOpts     bind.CallOpts          // Call options to use throughout this session
	TransactOpts bind.TransactOpts      // Transaction auth options to use throughout this session
}

// WrappedTokenGatewayV3CallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type WrappedTokenGatewayV3CallerSession struct {
	Contract *WrappedTokenGatewayV3Caller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                // Call options to use throughout this session
}

// WrappedTokenGatewayV3TransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type WrappedTokenGatewayV3TransactorSession struct {
	Contract     *WrappedTokenGatewayV3Transactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                // Transaction auth options to use throughout this session
}

// WrappedTokenGatewayV3Raw is an auto generated low-level Go binding around an Ethereum contract.
type WrappedTokenGatewayV3Raw struct {
	Contract *WrappedTokenGatewayV3 // Generic contract binding to access the raw methods on
}

// WrappedTokenGatewayV3CallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type WrappedTokenGatewayV3CallerRaw struct {
	Contract *WrappedTokenGatewayV3Caller // Generic read-only contract binding to access the raw methods on
}

// WrappedTokenGatewayV3TransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type WrappedTokenGatewayV3TransactorRaw struct {
	Contract *WrappedTokenGatewayV3Transactor // Generic write-only contract binding to access the raw methods on
}

// NewWrappedTokenGatewayV3 creates a new instance of WrappedTokenGatewayV3, bound to a specific deployed contract.
func NewWrappedTokenGatewayV3(address common.Address, backend bind.ContractBackend) (*WrappedTokenGatewayV3, error) {
	contract, err := bindWrappedTokenGatewayV3(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &WrappedTokenGatewayV3{WrappedTokenGatewayV3Caller: WrappedTokenGatewayV3Caller{contract: contract}, WrappedTokenGatewayV3Transactor: WrappedTokenGatewayV3Transactor{contract: contract}, WrappedTokenGatewayV3Filterer: WrappedTokenGatewayV3Filterer{contract: contract}}, nil
}

// NewWrappedTokenGatewayV3Caller creates a new read-only instance of WrappedTokenGatewayV3, bound to a specific deployed contract.
func NewWrappedTokenGatewayV3Caller(address common.Address, caller bind.ContractCaller) (*WrappedTokenGatewayV3Caller, error) {
	contract, err := bindWrappedTokenGatewayV3(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &WrappedTokenGatewayV3Caller{contract: contract}, nil
}

// NewWrappedTokenGatewayV3Transactor creates a new write-only instance of WrappedTokenGatewayV3, bound to a specific deployed contract.
func NewWrappedTokenGatewayV3Transactor(address common.Address, transactor bind.ContractTransactor) (*WrappedTokenGatewayV3Transactor, error) {
	contract, err := bindWrappedTokenGatewayV3(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &WrappedTokenGatewayV3Transactor{contract: contract}, nil
}

// NewWrappedTokenGatewayV3Filterer creates a new log filterer instance of WrappedTokenGatewayV3, bound to a specific deployed contract.
func NewWrappedTokenGatewayV3Filterer(address common.Address, filterer bind.ContractFilterer) (*WrappedTokenGatewayV3Filterer, error) {
	contract, err := bindWrappedTokenGatewayV3(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &WrappedTokenGatewayV3Filterer{contract: contract}, nil
}

// bindWrappedTokenGatewayV3 binds a generic wrapper to an already deployed contract.
func bindWrappedTokenGatewayV3(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(WrappedTokenGatewayV3ABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_WrappedTokenGatewayV3 *WrappedTokenGatewayV3Raw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _WrappedTokenGatewayV3.Contract.WrappedTokenGatewayV3Caller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_WrappedTokenGatewayV3 *WrappedTokenGatewayV3Raw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _WrappedTokenGatewayV3.Contract.WrappedTokenGatewayV3Transactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_WrappedTokenGatewayV3 *WrappedTokenGatewayV3Raw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _WrappedTokenGatewayV3.Contract.WrappedTokenGatewayV3Transactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_WrappedTokenGatewayV3 *WrappedTokenGatewayV3CallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _WrappedTokenGatewayV3.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_WrappedTokenGatewayV3 *WrappedTokenGatewayV3TransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _WrappedTokenGatewayV3.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_WrappedTokenGatewayV3 *WrappedTokenGatewayV3TransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _WrappedTokenGatewayV3.Contract.contract.Transact(opts, method, params...)
}

// GetWETHAddress is a free data retrieval call binding the contract method 0xaffa8817.
//
// Solidity: function getWETHAddress() view returns(address)
func (_WrappedTokenGatewayV3 *WrappedTokenGatewayV3Caller) GetWETHAddress(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _WrappedTokenGatewayV3.contract.Call(opts, &out, "getWETHAddress")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetWETHAddress is a free data retrieval call binding the contract method 0xaffa8817.
//
// Solidity: function getWETHAddress() view returns(address)
func (_WrappedTokenGatewayV3 *WrappedTokenGatewayV3Session) GetWETHAddress() (common.Address, error) {
	return _WrappedTokenGatewayV3.Contract.GetWETHAddress(&_WrappedTokenGatewayV3.CallOpts)
}

// GetWETHAddress is a free data retrieval call binding the contract method 0xaffa8817.
//
// Solidity: function getWETHAddress() view returns(address)
func (_WrappedTokenGatewayV3 *WrappedTokenGatewayV3CallerSession) GetWETHAddress() (common.Address, error) {
	return _WrappedTokenGatewayV3.Contract.GetWETHAddress(&_WrappedTokenGatewayV3.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_WrappedTokenGatewayV3 *WrappedTokenGatewayV3Caller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _WrappedTokenGatewayV3.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_WrappedTokenGatewayV3 *WrappedTokenGatewayV3Session) Owner() (common.Address, error) {
	return _WrappedTokenGatewayV3.Contract.Owner(&_WrappedTokenGatewayV3.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_WrappedTokenGatewayV3 *WrappedTokenGatewayV3CallerSession) Owner() (common.Address, error) {
	return _WrappedTokenGatewayV3.Contract.Owner(&_WrappedTokenGatewayV3.CallOpts)
}

// BorrowETH is a paid mutator transaction binding the contract method 0xe74f7b85.
//
// Solidity: function borrowETH(address , uint256 amount, uint16 referralCode) returns()
func (_WrappedTokenGatewayV3 *WrappedTokenGatewayV3Transactor) BorrowETH(opts *bind.TransactOpts, arg0 common.Address, amount *big.Int, referralCode uint16) (*types.Transaction, error) {
	return _WrappedTokenGatewayV3.contract.Transact(opts, "borrowETH", arg0, amount, referralCode)
}

// BorrowETH is a paid mutator transaction binding the contract method 0xe74f7b85.
//
// Solidity: function borrowETH(address , uint256 amount, uint16 referralCode) returns()
func (_WrappedTokenGatewayV3 *WrappedTokenGatewayV3Session) BorrowETH(arg0 common.Address, amount *big.Int, referralCode uint16) (*types.Transaction, error) {
	return _WrappedTokenGatewayV3.Contract.BorrowETH(&_WrappedTokenGatewayV3.TransactOpts, arg0, amount, referralCode)
}

// BorrowETH is a paid mutator transaction binding the contract method 0xe74f7b85.
//
// Solidity: function borrowETH(address , uint256 amount, uint16 referralCode) returns()
func (_WrappedTokenGatewayV3 *WrappedTokenGatewayV3TransactorSession) BorrowETH(arg0 common.Address, amount *big.Int, referralCode uint16) (*types.Transaction, error) {
	return _WrappedTokenGatewayV3.Contract.BorrowETH(&_WrappedTokenGatewayV3.TransactOpts, arg0, amount, referralCode)
}

// DepositETH is a paid mutator transaction binding the contract method 0x474cf53d.
//
// Solidity: function depositETH(address , address onBehalfOf, uint16 referralCode) payable returns()
func (_WrappedTokenGatewayV3 *WrappedTokenGatewayV3Transactor) DepositETH(opts *bind.TransactOpts, arg0 common.Address, onBehalfOf common.Address, referralCode uint16) (*types.Transaction, error) {
	return _WrappedTokenGatewayV3.contract.Transact(opts, "depositETH", arg0, onBehalfOf, referralCode)
}

// DepositETH is a paid mutator transaction binding the contract method 0x474cf53d.
//
// Solidity: function depositETH(address , address onBehalfOf, uint16 referralCode) payable returns()
func (_WrappedTokenGatewayV3 *WrappedTokenGatewayV3Session) DepositETH(arg0 common.Address, onBehalfOf common.Address, referralCode uint16) (*types.Transaction, error) {
	return _WrappedTokenGatewayV3.Contract.DepositETH(&_WrappedTokenGatewayV3.TransactOpts, arg0, onBehalfOf, referralCode)
}

// DepositETH is a paid mutator transaction binding the contract method 0x474cf53d.
//
// Solidity: function depositETH(address , address onBehalfOf, uint16 referralCode) payable returns()
func (_WrappedTokenGatewayV3 *WrappedTokenGatewayV3TransactorSession) DepositETH(arg0 common.Address, onBehalfOf common.Address, referralCode uint16) (*types.Transaction, error) {
	return _WrappedTokenGatewayV3.Contract.DepositETH(&_WrappedTokenGatewayV3.TransactOpts, arg0, onBehalfOf, referralCode)
}

// EmergencyEtherTransfer is a paid mutator transaction binding the contract method 0xeed88b8d.
//
// Solidity: function emergencyEtherTransfer(address to, uint256 amount) returns()
func (_WrappedTokenGatewayV3 *WrappedTokenGatewayV3Transactor) EmergencyEtherTransfer(opts *bind.TransactOpts, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _WrappedTokenGatewayV3.contract.Transact(opts, "emergencyEtherTransfer", to, amount)
}

// EmergencyEtherTransfer is a paid mutator transaction binding the contract method 0xeed88b8d.
//
// Solidity: function emergencyEtherTransfer(address to, uint256 amount) returns()
func (_WrappedTokenGatewayV3 *WrappedTokenGatewayV3Session) EmergencyEtherTransfer(to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _WrappedTokenGatewayV3.Contract.EmergencyEtherTransfer(&_WrappedTokenGatewayV3.TransactOpts, to, amount)
}

// EmergencyEtherTransfer is a paid mutator transaction binding the contract method 0xeed88b8d.
//
// Solidity: function emergencyEtherTransfer(address to, uint256 amount) returns()
func (_WrappedTokenGatewayV3 *WrappedTokenGatewayV3TransactorSession) EmergencyEtherTransfer(to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _WrappedTokenGatewayV3.Contract.EmergencyEtherTransfer(&_WrappedTokenGatewayV3.TransactOpts, to, amount)
}

// EmergencyTokenTransfer is a paid mutator transaction binding the contract method 0xa3d5b255.
//
// Solidity: function emergencyTokenTransfer(address token, address to, uint256 amount) returns()
func (_WrappedTokenGatewayV3 *WrappedTokenGatewayV3Transactor) EmergencyTokenTransfer(opts *bind.TransactOpts, token common.Address, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _WrappedTokenGatewayV3.contract.Transact(opts, "emergencyTokenTransfer", token, to, amount)
}

// EmergencyTokenTransfer is a paid mutator transaction binding the contract method 0xa3d5b255.
//
// Solidity: function emergencyTokenTransfer(address token, address to, uint256 amount) returns()
func (_WrappedTokenGatewayV3 *WrappedTokenGatewayV3Session) EmergencyTokenTransfer(token common.Address, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _WrappedTokenGatewayV3.Contract.EmergencyTokenTransfer(&_WrappedTokenGatewayV3.TransactOpts, token, to, amount)
}

// EmergencyTokenTransfer is a paid mutator transaction binding the contract method 0xa3d5b255.
//
// Solidity: function emergencyTokenTransfer(address token, address to, uint256 amount) returns()
func (_WrappedTokenGatewayV3 *WrappedTokenGatewayV3TransactorSession) EmergencyTokenTransfer(token common.Address, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return _WrappedTokenGatewayV3.Contract.EmergencyTokenTransfer(&_WrappedTokenGatewayV3.TransactOpts, token, to, amount)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_WrappedTokenGatewayV3 *WrappedTokenGatewayV3Transactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _WrappedTokenGatewayV3.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_WrappedTokenGatewayV3 *WrappedTokenGatewayV3Session) RenounceOwnership() (*types.Transaction, error) {
	return _WrappedTokenGatewayV3.Contract.RenounceOwnership(&_WrappedTokenGatewayV3.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_WrappedTokenGatewayV3 *WrappedTokenGatewayV3TransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _WrappedTokenGatewayV3.Contract.RenounceOwnership(&_WrappedTokenGatewayV3.TransactOpts)
}

// RepayETH is a paid mutator transaction binding the contract method 0x02c5fcf8.
//
// Solidity: function repayETH(address , uint256 amount, uint256 rateMode, address onBehalfOf) payable returns()
func (_WrappedTokenGatewayV3 *WrappedTokenGatewayV3Transactor) RepayETH(opts *bind.TransactOpts, arg0 common.Address, amount *big.Int, rateMode *big.Int, onBehalfOf common.Address) (*types.Transaction, error) {
	return _WrappedTokenGatewayV3.contract.Transact(opts, "repayETH", arg0, amount, rateMode, onBehalfOf)
}

// RepayETH is a paid mutator transaction binding the contract method 0x02c5fcf8.
//
// Solidity: function repayETH(address , uint256 amount, uint256 rateMode, address onBehalfOf) payable returns()
func (_WrappedTokenGatewayV3 *WrappedTokenGatewayV3Session) RepayETH(arg0 common.Address, amount *big.Int, rateMode *big.Int, onBehalfOf common.Address) (*types.Transaction, error) {
	return _WrappedTokenGatewayV3.Contract.RepayETH(&_WrappedTokenGatewayV3.TransactOpts, arg0, amount, rateMode, onBehalfOf)
}

// RepayETH is a paid mutator transaction binding the contract method 0x02c5fcf8.
//
// Solidity: function repayETH(address , uint256 amount, uint256 rateMode, address onBehalfOf) payable returns()
func (_WrappedTokenGatewayV3 *WrappedTokenGatewayV3TransactorSession) RepayETH(arg0 common.Address, amount *big.Int, rateMode *big.Int, onBehalfOf common.Address) (*types.Transaction, error) {
	return _WrappedTokenGatewayV3.Contract.RepayETH(&_WrappedTokenGatewayV3.TransactOpts, arg0, amount, rateMode, onBehalfOf)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_WrappedTokenGatewayV3 *WrappedTokenGatewayV3Transactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _WrappedTokenGatewayV3.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_WrappedTokenGatewayV3 *WrappedTokenGatewayV3Session) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _WrappedTokenGatewayV3.Contract.TransferOwnership(&_WrappedTokenGatewayV3.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_WrappedTokenGatewayV3 *WrappedTokenGatewayV3TransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _WrappedTokenGatewayV3.Contract.TransferOwnership(&_WrappedTokenGatewayV3.TransactOpts, newOwner)
}

// WithdrawETH is a paid mutator transaction binding the contract method 0x80500d20.
//
// Solidity: function withdrawETH(address , uint256 amount, address to) returns()
func (_WrappedTokenGatewayV3 *WrappedTokenGatewayV3Transactor) WithdrawETH(opts *bind.TransactOpts, arg0 common.Address, amount *big.Int, to common.Address) (*types.Transaction, error) {
	return _WrappedTokenGatewayV3.contract.Transact(opts, "withdrawETH", arg0, amount, to)
}

// WithdrawETH is a paid mutator transaction binding the contract method 0x80500d20.
//
// Solidity: function withdrawETH(address , uint256 amount, address to) returns()
func (_WrappedTokenGatewayV3 *WrappedTokenGatewayV3Session) WithdrawETH(arg0 common.Address, amount *big.Int, to common.Address) (*types.Transaction, error) {
	return _WrappedTokenGatewayV3.Contract.WithdrawETH(&_WrappedTokenGatewayV3.TransactOpts, arg0, amount, to)
}

// WithdrawETH is a paid mutator transaction binding the contract method 0x80500d20.
//
// Solidity: function withdrawETH(address , uint256 amount, address to) returns()
func (_WrappedTokenGatewayV3 *WrappedTokenGatewayV3TransactorSession) WithdrawETH(arg0 common.Address, amount *big.Int, to common.Address) (*types.Transaction, error) {
	return _WrappedTokenGatewayV3.Contract.WithdrawETH(&_WrappedTokenGatewayV3.TransactOpts, arg0, amount, to)
}

// WithdrawETHWithPermit is a paid mutator transaction binding the contract method 0xd4c40b6c.
//
// Solidity: function withdrawETHWithPermit(address , uint256 amount, address to, uint256 deadline, uint8 permitV, bytes32 permitR, bytes32 permitS) returns()
func (_WrappedTokenGatewayV3 *WrappedTokenGatewayV3Transactor) WithdrawETHWithPermit(opts *bind.TransactOpts, arg0 common.Address, amount *big.Int, to common.Address, deadline *big.Int, permitV uint8, permitR [32]byte, permitS [32]byte) (*types.Transaction, error) {
	return _WrappedTokenGatewayV3.contract.Transact(opts, "withdrawETHWithPermit", arg0, amount, to, deadline, permitV, permitR, permitS)
}

// WithdrawETHWithPermit is a paid mutator transaction binding the contract method 0xd4c40b6c.
//
// Solidity: function withdrawETHWithPermit(address , uint256 amount, address to, uint256 deadline, uint8 permitV, bytes32 permitR, bytes32 permitS) returns()
func (_WrappedTokenGatewayV3 *WrappedTokenGatewayV3Session) WithdrawETHWithPermit(arg0 common.Address, amount *big.Int, to common.Address, deadline *big.Int, permitV uint8, permitR [32]byte, permitS [32]byte) (*types.Transaction, error) {
	return _WrappedTokenGatewayV3.Contract.WithdrawETHWithPermit(&_WrappedTokenGatewayV3.TransactOpts, arg0, amount, to, deadline, permitV, permitR, permitS)
}

// WithdrawETHWithPermit is a paid mutator transaction binding the contract method 0xd4c40b6c.
//
// Solidity: function withdrawETHWithPermit(address , uint256 amount, address to, uint256 deadline, uint8 permitV, bytes32 permitR, bytes32 permitS) returns()
func (_WrappedTokenGatewayV3 *WrappedTokenGatewayV3TransactorSession) WithdrawETHWithPermit(arg0 common.Address, amount *big.Int, to common.Address, deadline *big.Int, permitV uint8, permitR [32]byte, permitS [32]byte) (*types.Transaction, error) {
	return _WrappedTokenGatewayV3.Contract.WithdrawETHWithPermit(&_WrappedTokenGatewayV3.TransactOpts, arg0, amount, to, deadline, permitV, permitR, permitS)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() payable returns()
func (_WrappedTokenGatewayV3 *WrappedTokenGatewayV3Transactor) Fallback(opts *bind.TransactOpts, calldata []byte) (*types.Transaction, error) {
	return _WrappedTokenGatewayV3.contract.RawTransact(opts, calldata)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() payable returns()
func (_WrappedTokenGatewayV3 *WrappedTokenGatewayV3Session) Fallback(calldata []byte) (*types.Transaction, error) {
	return _WrappedTokenGatewayV3.Contract.Fallback(&_WrappedTokenGatewayV3.TransactOpts, calldata)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() payable returns()
func (_WrappedTokenGatewayV3 *WrappedTokenGatewayV3TransactorSession) Fallback(calldata []byte) (*types.Transaction, error) {
	return _WrappedTokenGatewayV3.Contract.Fallback(&_WrappedTokenGatewayV3.TransactOpts, calldata)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_WrappedTokenGatewayV3 *WrappedTokenGatewayV3Transactor) Receive(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _WrappedTokenGatewayV3.contract.RawTransact(opts, nil) // calldata is disallowed for receive function
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_WrappedTokenGatewayV3 *WrappedTokenGatewayV3Session) Receive() (*types.Transaction, error) {
	return _WrappedTokenGatewayV3.Contract.Receive(&_WrappedTokenGatewayV3.TransactOpts)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_WrappedTokenGatewayV3 *WrappedTokenGatewayV3TransactorSession) Receive() (*types.Transaction, error) {
	return _WrappedTokenGatewayV3.Contract.Receive(&_WrappedTokenGatewayV3.TransactOpts)
}

// WrappedTokenGatewayV3OwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the WrappedTokenGatewayV3 contract.
type WrappedTokenGatewayV3OwnershipTransferredIterator struct {
	Event *WrappedTokenGatewayV3OwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *WrappedTokenGatewayV3OwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(WrappedTokenGatewayV3OwnershipTransferred)
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
		it.Event = new(WrappedTokenGatewayV3OwnershipTransferred)
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
func (it *WrappedTokenGatewayV3OwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *WrappedTokenGatewayV3OwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// WrappedTokenGatewayV3OwnershipTransferred represents a OwnershipTransferred event raised by the WrappedTokenGatewayV3 contract.
type WrappedTokenGatewayV3OwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_WrappedTokenGatewayV3 *WrappedTokenGatewayV3Filterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*WrappedTokenGatewayV3OwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _WrappedTokenGatewayV3.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &WrappedTokenGatewayV3OwnershipTransferredIterator{contract: _WrappedTokenGatewayV3.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_WrappedTokenGatewayV3 *WrappedTokenGatewayV3Filterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *WrappedTokenGatewayV3OwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _WrappedTokenGatewayV3.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(WrappedTokenGatewayV3OwnershipTransferred)
				if err := _WrappedTokenGatewayV3.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_WrappedTokenGatewayV3 *WrappedTokenGatewayV3Filterer) ParseOwnershipTransferred(log types.Log) (*WrappedTokenGatewayV3OwnershipTransferred, error) {
	event := new(WrappedTokenGatewayV3OwnershipTransferred)
	if err := _WrappedTokenGatewayV3.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
