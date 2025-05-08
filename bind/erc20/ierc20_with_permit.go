// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package erc20

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

// IErc20WithPermitMetaData contains all meta data concerning the IErc20WithPermit contract.
var IErc20WithPermitMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"name\":\"DOMAIN_SEPARATOR\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"nonces\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deadline\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"permit\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// IErc20WithPermitABI is the input ABI used to generate the binding from.
// Deprecated: Use IErc20WithPermitMetaData.ABI instead.
var IErc20WithPermitABI = IErc20WithPermitMetaData.ABI

// IErc20WithPermit is an auto generated Go binding around an Ethereum contract.
type IErc20WithPermit struct {
	IErc20WithPermitCaller     // Read-only binding to the contract
	IErc20WithPermitTransactor // Write-only binding to the contract
	IErc20WithPermitFilterer   // Log filterer for contract events
}

// IErc20WithPermitCaller is an auto generated read-only Go binding around an Ethereum contract.
type IErc20WithPermitCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IErc20WithPermitTransactor is an auto generated write-only Go binding around an Ethereum contract.
type IErc20WithPermitTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IErc20WithPermitFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type IErc20WithPermitFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IErc20WithPermitSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type IErc20WithPermitSession struct {
	Contract     *IErc20WithPermit // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// IErc20WithPermitCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type IErc20WithPermitCallerSession struct {
	Contract *IErc20WithPermitCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts           // Call options to use throughout this session
}

// IErc20WithPermitTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type IErc20WithPermitTransactorSession struct {
	Contract     *IErc20WithPermitTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts           // Transaction auth options to use throughout this session
}

// IErc20WithPermitRaw is an auto generated low-level Go binding around an Ethereum contract.
type IErc20WithPermitRaw struct {
	Contract *IErc20WithPermit // Generic contract binding to access the raw methods on
}

// IErc20WithPermitCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type IErc20WithPermitCallerRaw struct {
	Contract *IErc20WithPermitCaller // Generic read-only contract binding to access the raw methods on
}

// IErc20WithPermitTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type IErc20WithPermitTransactorRaw struct {
	Contract *IErc20WithPermitTransactor // Generic write-only contract binding to access the raw methods on
}

// NewIErc20WithPermit creates a new instance of IErc20WithPermit, bound to a specific deployed contract.
func NewIErc20WithPermit(address common.Address, backend bind.ContractBackend) (*IErc20WithPermit, error) {
	contract, err := bindIErc20WithPermit(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &IErc20WithPermit{IErc20WithPermitCaller: IErc20WithPermitCaller{contract: contract}, IErc20WithPermitTransactor: IErc20WithPermitTransactor{contract: contract}, IErc20WithPermitFilterer: IErc20WithPermitFilterer{contract: contract}}, nil
}

// NewIErc20WithPermitCaller creates a new read-only instance of IErc20WithPermit, bound to a specific deployed contract.
func NewIErc20WithPermitCaller(address common.Address, caller bind.ContractCaller) (*IErc20WithPermitCaller, error) {
	contract, err := bindIErc20WithPermit(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &IErc20WithPermitCaller{contract: contract}, nil
}

// NewIErc20WithPermitTransactor creates a new write-only instance of IErc20WithPermit, bound to a specific deployed contract.
func NewIErc20WithPermitTransactor(address common.Address, transactor bind.ContractTransactor) (*IErc20WithPermitTransactor, error) {
	contract, err := bindIErc20WithPermit(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &IErc20WithPermitTransactor{contract: contract}, nil
}

// NewIErc20WithPermitFilterer creates a new log filterer instance of IErc20WithPermit, bound to a specific deployed contract.
func NewIErc20WithPermitFilterer(address common.Address, filterer bind.ContractFilterer) (*IErc20WithPermitFilterer, error) {
	contract, err := bindIErc20WithPermit(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &IErc20WithPermitFilterer{contract: contract}, nil
}

// bindIErc20WithPermit binds a generic wrapper to an already deployed contract.
func bindIErc20WithPermit(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(IErc20WithPermitABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_IErc20WithPermit *IErc20WithPermitRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _IErc20WithPermit.Contract.IErc20WithPermitCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_IErc20WithPermit *IErc20WithPermitRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IErc20WithPermit.Contract.IErc20WithPermitTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_IErc20WithPermit *IErc20WithPermitRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _IErc20WithPermit.Contract.IErc20WithPermitTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_IErc20WithPermit *IErc20WithPermitCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _IErc20WithPermit.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_IErc20WithPermit *IErc20WithPermitTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IErc20WithPermit.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_IErc20WithPermit *IErc20WithPermitTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _IErc20WithPermit.Contract.contract.Transact(opts, method, params...)
}

// DOMAINSEPARATOR is a free data retrieval call binding the contract method 0x3644e515.
//
// Solidity: function DOMAIN_SEPARATOR() view returns(bytes32)
func (_IErc20WithPermit *IErc20WithPermitCaller) DOMAINSEPARATOR(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _IErc20WithPermit.contract.Call(opts, &out, "DOMAIN_SEPARATOR")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DOMAINSEPARATOR is a free data retrieval call binding the contract method 0x3644e515.
//
// Solidity: function DOMAIN_SEPARATOR() view returns(bytes32)
func (_IErc20WithPermit *IErc20WithPermitSession) DOMAINSEPARATOR() ([32]byte, error) {
	return _IErc20WithPermit.Contract.DOMAINSEPARATOR(&_IErc20WithPermit.CallOpts)
}

// DOMAINSEPARATOR is a free data retrieval call binding the contract method 0x3644e515.
//
// Solidity: function DOMAIN_SEPARATOR() view returns(bytes32)
func (_IErc20WithPermit *IErc20WithPermitCallerSession) DOMAINSEPARATOR() ([32]byte, error) {
	return _IErc20WithPermit.Contract.DOMAINSEPARATOR(&_IErc20WithPermit.CallOpts)
}

// Nonces is a free data retrieval call binding the contract method 0x7ecebe00.
//
// Solidity: function nonces(address owner) view returns(uint256)
func (_IErc20WithPermit *IErc20WithPermitCaller) Nonces(opts *bind.CallOpts, owner common.Address) (*big.Int, error) {
	var out []interface{}
	err := _IErc20WithPermit.contract.Call(opts, &out, "nonces", owner)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Nonces is a free data retrieval call binding the contract method 0x7ecebe00.
//
// Solidity: function nonces(address owner) view returns(uint256)
func (_IErc20WithPermit *IErc20WithPermitSession) Nonces(owner common.Address) (*big.Int, error) {
	return _IErc20WithPermit.Contract.Nonces(&_IErc20WithPermit.CallOpts, owner)
}

// Nonces is a free data retrieval call binding the contract method 0x7ecebe00.
//
// Solidity: function nonces(address owner) view returns(uint256)
func (_IErc20WithPermit *IErc20WithPermitCallerSession) Nonces(owner common.Address) (*big.Int, error) {
	return _IErc20WithPermit.Contract.Nonces(&_IErc20WithPermit.CallOpts, owner)
}

// Permit is a paid mutator transaction binding the contract method 0xd505accf.
//
// Solidity: function permit(address owner, address spender, uint256 value, uint256 deadline, uint8 v, bytes32 r, bytes32 s) returns()
func (_IErc20WithPermit *IErc20WithPermitTransactor) Permit(opts *bind.TransactOpts, owner common.Address, spender common.Address, value *big.Int, deadline *big.Int, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _IErc20WithPermit.contract.Transact(opts, "permit", owner, spender, value, deadline, v, r, s)
}

// Permit is a paid mutator transaction binding the contract method 0xd505accf.
//
// Solidity: function permit(address owner, address spender, uint256 value, uint256 deadline, uint8 v, bytes32 r, bytes32 s) returns()
func (_IErc20WithPermit *IErc20WithPermitSession) Permit(owner common.Address, spender common.Address, value *big.Int, deadline *big.Int, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _IErc20WithPermit.Contract.Permit(&_IErc20WithPermit.TransactOpts, owner, spender, value, deadline, v, r, s)
}

// Permit is a paid mutator transaction binding the contract method 0xd505accf.
//
// Solidity: function permit(address owner, address spender, uint256 value, uint256 deadline, uint8 v, bytes32 r, bytes32 s) returns()
func (_IErc20WithPermit *IErc20WithPermitTransactorSession) Permit(owner common.Address, spender common.Address, value *big.Int, deadline *big.Int, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _IErc20WithPermit.Contract.Permit(&_IErc20WithPermit.TransactOpts, owner, spender, value, deadline, v, r, s)
}
