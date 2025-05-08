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

// DebtTokenBaseMetaData contains all meta data concerning the DebtTokenBase contract.
var DebtTokenBaseMetaData = &bind.MetaData{
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"fromUser\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"toUser\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"asset\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"BorrowAllowanceDelegated\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"DELEGATION_WITH_SIG_TYPEHASH\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"DOMAIN_SEPARATOR\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"EIP712_REVISION\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"delegatee\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"approveDelegation\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"fromUser\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"toUser\",\"type\":\"address\"}],\"name\":\"borrowAllowance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"delegatee\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deadline\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"delegationWithSig\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"nonces\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// DebtTokenBaseABI is the input ABI used to generate the binding from.
// Deprecated: Use DebtTokenBaseMetaData.ABI instead.
var DebtTokenBaseABI = DebtTokenBaseMetaData.ABI

// DebtTokenBase is an auto generated Go binding around an Ethereum contract.
type DebtTokenBase struct {
	DebtTokenBaseCaller     // Read-only binding to the contract
	DebtTokenBaseTransactor // Write-only binding to the contract
	DebtTokenBaseFilterer   // Log filterer for contract events
}

// DebtTokenBaseCaller is an auto generated read-only Go binding around an Ethereum contract.
type DebtTokenBaseCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DebtTokenBaseTransactor is an auto generated write-only Go binding around an Ethereum contract.
type DebtTokenBaseTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DebtTokenBaseFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type DebtTokenBaseFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DebtTokenBaseSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type DebtTokenBaseSession struct {
	Contract     *DebtTokenBase    // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// DebtTokenBaseCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type DebtTokenBaseCallerSession struct {
	Contract *DebtTokenBaseCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts        // Call options to use throughout this session
}

// DebtTokenBaseTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type DebtTokenBaseTransactorSession struct {
	Contract     *DebtTokenBaseTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// DebtTokenBaseRaw is an auto generated low-level Go binding around an Ethereum contract.
type DebtTokenBaseRaw struct {
	Contract *DebtTokenBase // Generic contract binding to access the raw methods on
}

// DebtTokenBaseCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type DebtTokenBaseCallerRaw struct {
	Contract *DebtTokenBaseCaller // Generic read-only contract binding to access the raw methods on
}

// DebtTokenBaseTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type DebtTokenBaseTransactorRaw struct {
	Contract *DebtTokenBaseTransactor // Generic write-only contract binding to access the raw methods on
}

// NewDebtTokenBase creates a new instance of DebtTokenBase, bound to a specific deployed contract.
func NewDebtTokenBase(address common.Address, backend bind.ContractBackend) (*DebtTokenBase, error) {
	contract, err := bindDebtTokenBase(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &DebtTokenBase{DebtTokenBaseCaller: DebtTokenBaseCaller{contract: contract}, DebtTokenBaseTransactor: DebtTokenBaseTransactor{contract: contract}, DebtTokenBaseFilterer: DebtTokenBaseFilterer{contract: contract}}, nil
}

// NewDebtTokenBaseCaller creates a new read-only instance of DebtTokenBase, bound to a specific deployed contract.
func NewDebtTokenBaseCaller(address common.Address, caller bind.ContractCaller) (*DebtTokenBaseCaller, error) {
	contract, err := bindDebtTokenBase(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &DebtTokenBaseCaller{contract: contract}, nil
}

// NewDebtTokenBaseTransactor creates a new write-only instance of DebtTokenBase, bound to a specific deployed contract.
func NewDebtTokenBaseTransactor(address common.Address, transactor bind.ContractTransactor) (*DebtTokenBaseTransactor, error) {
	contract, err := bindDebtTokenBase(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &DebtTokenBaseTransactor{contract: contract}, nil
}

// NewDebtTokenBaseFilterer creates a new log filterer instance of DebtTokenBase, bound to a specific deployed contract.
func NewDebtTokenBaseFilterer(address common.Address, filterer bind.ContractFilterer) (*DebtTokenBaseFilterer, error) {
	contract, err := bindDebtTokenBase(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &DebtTokenBaseFilterer{contract: contract}, nil
}

// bindDebtTokenBase binds a generic wrapper to an already deployed contract.
func bindDebtTokenBase(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(DebtTokenBaseABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_DebtTokenBase *DebtTokenBaseRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _DebtTokenBase.Contract.DebtTokenBaseCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_DebtTokenBase *DebtTokenBaseRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _DebtTokenBase.Contract.DebtTokenBaseTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_DebtTokenBase *DebtTokenBaseRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _DebtTokenBase.Contract.DebtTokenBaseTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_DebtTokenBase *DebtTokenBaseCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _DebtTokenBase.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_DebtTokenBase *DebtTokenBaseTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _DebtTokenBase.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_DebtTokenBase *DebtTokenBaseTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _DebtTokenBase.Contract.contract.Transact(opts, method, params...)
}

// DELEGATIONWITHSIGTYPEHASH is a free data retrieval call binding the contract method 0xf3bfc738.
//
// Solidity: function DELEGATION_WITH_SIG_TYPEHASH() view returns(bytes32)
func (_DebtTokenBase *DebtTokenBaseCaller) DELEGATIONWITHSIGTYPEHASH(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _DebtTokenBase.contract.Call(opts, &out, "DELEGATION_WITH_SIG_TYPEHASH")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DELEGATIONWITHSIGTYPEHASH is a free data retrieval call binding the contract method 0xf3bfc738.
//
// Solidity: function DELEGATION_WITH_SIG_TYPEHASH() view returns(bytes32)
func (_DebtTokenBase *DebtTokenBaseSession) DELEGATIONWITHSIGTYPEHASH() ([32]byte, error) {
	return _DebtTokenBase.Contract.DELEGATIONWITHSIGTYPEHASH(&_DebtTokenBase.CallOpts)
}

// DELEGATIONWITHSIGTYPEHASH is a free data retrieval call binding the contract method 0xf3bfc738.
//
// Solidity: function DELEGATION_WITH_SIG_TYPEHASH() view returns(bytes32)
func (_DebtTokenBase *DebtTokenBaseCallerSession) DELEGATIONWITHSIGTYPEHASH() ([32]byte, error) {
	return _DebtTokenBase.Contract.DELEGATIONWITHSIGTYPEHASH(&_DebtTokenBase.CallOpts)
}

// DOMAINSEPARATOR is a free data retrieval call binding the contract method 0x3644e515.
//
// Solidity: function DOMAIN_SEPARATOR() view returns(bytes32)
func (_DebtTokenBase *DebtTokenBaseCaller) DOMAINSEPARATOR(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _DebtTokenBase.contract.Call(opts, &out, "DOMAIN_SEPARATOR")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DOMAINSEPARATOR is a free data retrieval call binding the contract method 0x3644e515.
//
// Solidity: function DOMAIN_SEPARATOR() view returns(bytes32)
func (_DebtTokenBase *DebtTokenBaseSession) DOMAINSEPARATOR() ([32]byte, error) {
	return _DebtTokenBase.Contract.DOMAINSEPARATOR(&_DebtTokenBase.CallOpts)
}

// DOMAINSEPARATOR is a free data retrieval call binding the contract method 0x3644e515.
//
// Solidity: function DOMAIN_SEPARATOR() view returns(bytes32)
func (_DebtTokenBase *DebtTokenBaseCallerSession) DOMAINSEPARATOR() ([32]byte, error) {
	return _DebtTokenBase.Contract.DOMAINSEPARATOR(&_DebtTokenBase.CallOpts)
}

// EIP712REVISION is a free data retrieval call binding the contract method 0x78160376.
//
// Solidity: function EIP712_REVISION() view returns(bytes)
func (_DebtTokenBase *DebtTokenBaseCaller) EIP712REVISION(opts *bind.CallOpts) ([]byte, error) {
	var out []interface{}
	err := _DebtTokenBase.contract.Call(opts, &out, "EIP712_REVISION")

	if err != nil {
		return *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([]byte)).(*[]byte)

	return out0, err

}

// EIP712REVISION is a free data retrieval call binding the contract method 0x78160376.
//
// Solidity: function EIP712_REVISION() view returns(bytes)
func (_DebtTokenBase *DebtTokenBaseSession) EIP712REVISION() ([]byte, error) {
	return _DebtTokenBase.Contract.EIP712REVISION(&_DebtTokenBase.CallOpts)
}

// EIP712REVISION is a free data retrieval call binding the contract method 0x78160376.
//
// Solidity: function EIP712_REVISION() view returns(bytes)
func (_DebtTokenBase *DebtTokenBaseCallerSession) EIP712REVISION() ([]byte, error) {
	return _DebtTokenBase.Contract.EIP712REVISION(&_DebtTokenBase.CallOpts)
}

// BorrowAllowance is a free data retrieval call binding the contract method 0x6bd76d24.
//
// Solidity: function borrowAllowance(address fromUser, address toUser) view returns(uint256)
func (_DebtTokenBase *DebtTokenBaseCaller) BorrowAllowance(opts *bind.CallOpts, fromUser common.Address, toUser common.Address) (*big.Int, error) {
	var out []interface{}
	err := _DebtTokenBase.contract.Call(opts, &out, "borrowAllowance", fromUser, toUser)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BorrowAllowance is a free data retrieval call binding the contract method 0x6bd76d24.
//
// Solidity: function borrowAllowance(address fromUser, address toUser) view returns(uint256)
func (_DebtTokenBase *DebtTokenBaseSession) BorrowAllowance(fromUser common.Address, toUser common.Address) (*big.Int, error) {
	return _DebtTokenBase.Contract.BorrowAllowance(&_DebtTokenBase.CallOpts, fromUser, toUser)
}

// BorrowAllowance is a free data retrieval call binding the contract method 0x6bd76d24.
//
// Solidity: function borrowAllowance(address fromUser, address toUser) view returns(uint256)
func (_DebtTokenBase *DebtTokenBaseCallerSession) BorrowAllowance(fromUser common.Address, toUser common.Address) (*big.Int, error) {
	return _DebtTokenBase.Contract.BorrowAllowance(&_DebtTokenBase.CallOpts, fromUser, toUser)
}

// Nonces is a free data retrieval call binding the contract method 0x7ecebe00.
//
// Solidity: function nonces(address owner) view returns(uint256)
func (_DebtTokenBase *DebtTokenBaseCaller) Nonces(opts *bind.CallOpts, owner common.Address) (*big.Int, error) {
	var out []interface{}
	err := _DebtTokenBase.contract.Call(opts, &out, "nonces", owner)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Nonces is a free data retrieval call binding the contract method 0x7ecebe00.
//
// Solidity: function nonces(address owner) view returns(uint256)
func (_DebtTokenBase *DebtTokenBaseSession) Nonces(owner common.Address) (*big.Int, error) {
	return _DebtTokenBase.Contract.Nonces(&_DebtTokenBase.CallOpts, owner)
}

// Nonces is a free data retrieval call binding the contract method 0x7ecebe00.
//
// Solidity: function nonces(address owner) view returns(uint256)
func (_DebtTokenBase *DebtTokenBaseCallerSession) Nonces(owner common.Address) (*big.Int, error) {
	return _DebtTokenBase.Contract.Nonces(&_DebtTokenBase.CallOpts, owner)
}

// ApproveDelegation is a paid mutator transaction binding the contract method 0xc04a8a10.
//
// Solidity: function approveDelegation(address delegatee, uint256 amount) returns()
func (_DebtTokenBase *DebtTokenBaseTransactor) ApproveDelegation(opts *bind.TransactOpts, delegatee common.Address, amount *big.Int) (*types.Transaction, error) {
	return _DebtTokenBase.contract.Transact(opts, "approveDelegation", delegatee, amount)
}

// ApproveDelegation is a paid mutator transaction binding the contract method 0xc04a8a10.
//
// Solidity: function approveDelegation(address delegatee, uint256 amount) returns()
func (_DebtTokenBase *DebtTokenBaseSession) ApproveDelegation(delegatee common.Address, amount *big.Int) (*types.Transaction, error) {
	return _DebtTokenBase.Contract.ApproveDelegation(&_DebtTokenBase.TransactOpts, delegatee, amount)
}

// ApproveDelegation is a paid mutator transaction binding the contract method 0xc04a8a10.
//
// Solidity: function approveDelegation(address delegatee, uint256 amount) returns()
func (_DebtTokenBase *DebtTokenBaseTransactorSession) ApproveDelegation(delegatee common.Address, amount *big.Int) (*types.Transaction, error) {
	return _DebtTokenBase.Contract.ApproveDelegation(&_DebtTokenBase.TransactOpts, delegatee, amount)
}

// DelegationWithSig is a paid mutator transaction binding the contract method 0x0b52d558.
//
// Solidity: function delegationWithSig(address delegator, address delegatee, uint256 value, uint256 deadline, uint8 v, bytes32 r, bytes32 s) returns()
func (_DebtTokenBase *DebtTokenBaseTransactor) DelegationWithSig(opts *bind.TransactOpts, delegator common.Address, delegatee common.Address, value *big.Int, deadline *big.Int, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _DebtTokenBase.contract.Transact(opts, "delegationWithSig", delegator, delegatee, value, deadline, v, r, s)
}

// DelegationWithSig is a paid mutator transaction binding the contract method 0x0b52d558.
//
// Solidity: function delegationWithSig(address delegator, address delegatee, uint256 value, uint256 deadline, uint8 v, bytes32 r, bytes32 s) returns()
func (_DebtTokenBase *DebtTokenBaseSession) DelegationWithSig(delegator common.Address, delegatee common.Address, value *big.Int, deadline *big.Int, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _DebtTokenBase.Contract.DelegationWithSig(&_DebtTokenBase.TransactOpts, delegator, delegatee, value, deadline, v, r, s)
}

// DelegationWithSig is a paid mutator transaction binding the contract method 0x0b52d558.
//
// Solidity: function delegationWithSig(address delegator, address delegatee, uint256 value, uint256 deadline, uint8 v, bytes32 r, bytes32 s) returns()
func (_DebtTokenBase *DebtTokenBaseTransactorSession) DelegationWithSig(delegator common.Address, delegatee common.Address, value *big.Int, deadline *big.Int, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _DebtTokenBase.Contract.DelegationWithSig(&_DebtTokenBase.TransactOpts, delegator, delegatee, value, deadline, v, r, s)
}

// DebtTokenBaseBorrowAllowanceDelegatedIterator is returned from FilterBorrowAllowanceDelegated and is used to iterate over the raw logs and unpacked data for BorrowAllowanceDelegated events raised by the DebtTokenBase contract.
type DebtTokenBaseBorrowAllowanceDelegatedIterator struct {
	Event *DebtTokenBaseBorrowAllowanceDelegated // Event containing the contract specifics and raw log

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
func (it *DebtTokenBaseBorrowAllowanceDelegatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DebtTokenBaseBorrowAllowanceDelegated)
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
		it.Event = new(DebtTokenBaseBorrowAllowanceDelegated)
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
func (it *DebtTokenBaseBorrowAllowanceDelegatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DebtTokenBaseBorrowAllowanceDelegatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DebtTokenBaseBorrowAllowanceDelegated represents a BorrowAllowanceDelegated event raised by the DebtTokenBase contract.
type DebtTokenBaseBorrowAllowanceDelegated struct {
	FromUser common.Address
	ToUser   common.Address
	Asset    common.Address
	Amount   *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterBorrowAllowanceDelegated is a free log retrieval operation binding the contract event 0xda919360433220e13b51e8c211e490d148e61a3bd53de8c097194e458b97f3e1.
//
// Solidity: event BorrowAllowanceDelegated(address indexed fromUser, address indexed toUser, address indexed asset, uint256 amount)
func (_DebtTokenBase *DebtTokenBaseFilterer) FilterBorrowAllowanceDelegated(opts *bind.FilterOpts, fromUser []common.Address, toUser []common.Address, asset []common.Address) (*DebtTokenBaseBorrowAllowanceDelegatedIterator, error) {

	var fromUserRule []interface{}
	for _, fromUserItem := range fromUser {
		fromUserRule = append(fromUserRule, fromUserItem)
	}
	var toUserRule []interface{}
	for _, toUserItem := range toUser {
		toUserRule = append(toUserRule, toUserItem)
	}
	var assetRule []interface{}
	for _, assetItem := range asset {
		assetRule = append(assetRule, assetItem)
	}

	logs, sub, err := _DebtTokenBase.contract.FilterLogs(opts, "BorrowAllowanceDelegated", fromUserRule, toUserRule, assetRule)
	if err != nil {
		return nil, err
	}
	return &DebtTokenBaseBorrowAllowanceDelegatedIterator{contract: _DebtTokenBase.contract, event: "BorrowAllowanceDelegated", logs: logs, sub: sub}, nil
}

// WatchBorrowAllowanceDelegated is a free log subscription operation binding the contract event 0xda919360433220e13b51e8c211e490d148e61a3bd53de8c097194e458b97f3e1.
//
// Solidity: event BorrowAllowanceDelegated(address indexed fromUser, address indexed toUser, address indexed asset, uint256 amount)
func (_DebtTokenBase *DebtTokenBaseFilterer) WatchBorrowAllowanceDelegated(opts *bind.WatchOpts, sink chan<- *DebtTokenBaseBorrowAllowanceDelegated, fromUser []common.Address, toUser []common.Address, asset []common.Address) (event.Subscription, error) {

	var fromUserRule []interface{}
	for _, fromUserItem := range fromUser {
		fromUserRule = append(fromUserRule, fromUserItem)
	}
	var toUserRule []interface{}
	for _, toUserItem := range toUser {
		toUserRule = append(toUserRule, toUserItem)
	}
	var assetRule []interface{}
	for _, assetItem := range asset {
		assetRule = append(assetRule, assetItem)
	}

	logs, sub, err := _DebtTokenBase.contract.WatchLogs(opts, "BorrowAllowanceDelegated", fromUserRule, toUserRule, assetRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DebtTokenBaseBorrowAllowanceDelegated)
				if err := _DebtTokenBase.contract.UnpackLog(event, "BorrowAllowanceDelegated", log); err != nil {
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

// ParseBorrowAllowanceDelegated is a log parse operation binding the contract event 0xda919360433220e13b51e8c211e490d148e61a3bd53de8c097194e458b97f3e1.
//
// Solidity: event BorrowAllowanceDelegated(address indexed fromUser, address indexed toUser, address indexed asset, uint256 amount)
func (_DebtTokenBase *DebtTokenBaseFilterer) ParseBorrowAllowanceDelegated(log types.Log) (*DebtTokenBaseBorrowAllowanceDelegated, error) {
	event := new(DebtTokenBaseBorrowAllowanceDelegated)
	if err := _DebtTokenBase.contract.UnpackLog(event, "BorrowAllowanceDelegated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
