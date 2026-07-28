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

// BaseAccountCall is an auto generated low-level Go binding around an user-defined struct.
type BaseAccountCall struct {
	Target common.Address
	Value  *big.Int
	Data   []byte
}

// IDefiSimplify7702AccountBalanceCheckpoint is an auto generated low-level Go binding around an user-defined struct.
type IDefiSimplify7702AccountBalanceCheckpoint struct {
	Token common.Address
	Id    [32]byte
}

// IDefiSimplify7702AccountBalancePatch is an auto generated low-level Go binding around an user-defined struct.
type IDefiSimplify7702AccountBalancePatch struct {
	Token        common.Address
	CheckpointId [32]byte
	Offset       uint32
	Bps          uint16
	Source       uint8
}

// IDefiSimplify7702AccountDynamicCall is an auto generated low-level Go binding around an user-defined struct.
type IDefiSimplify7702AccountDynamicCall struct {
	Target            common.Address
	Value             *big.Int
	Data              []byte
	CheckpointsBefore []IDefiSimplify7702AccountBalanceCheckpoint
	Patches           []IDefiSimplify7702AccountBalancePatch
	ExpectsCallback   bool
}

// PackedUserOperation is an auto generated low-level Go binding around an user-defined struct.
type PackedUserOperation struct {
	Sender             common.Address
	Nonce              *big.Int
	InitCode           []byte
	CallData           []byte
	AccountGasLimits   [32]byte
	PreVerificationGas *big.Int
	GasFees            [32]byte
	PaymasterAndData   []byte
	Signature          []byte
}

// DefiSimplify7702AccountMetaData contains all meta data concerning the DefiSimplify7702Account contract.
var DefiSimplify7702AccountMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"contractIEntryPoint\",\"name\":\"anEntryPoint\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"callIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"patchIndex\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"checkpointId\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"current\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"checkpoint\",\"type\":\"uint256\"}],\"name\":\"BalanceBelowCheckpoint\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"outerCallIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"callbackCallIndex\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"target\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"reason\",\"type\":\"bytes\"}],\"name\":\"CallbackDynamicCallFailed\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"callIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"actualState\",\"type\":\"uint8\"}],\"name\":\"CallbackNotAwaiting\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"callIndex\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"target\",\"type\":\"address\"},{\"internalType\":\"uint8\",\"name\":\"actualState\",\"type\":\"uint8\"}],\"name\":\"CallbackNotConsumed\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"callIndex\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"expected\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"actual\",\"type\":\"bytes32\"}],\"name\":\"CallbackOriginMismatch\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"CallbackOutsideDynamicExecution\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"callIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"checkpointIndex\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"id\",\"type\":\"bytes32\"}],\"name\":\"CheckpointAlreadyExists\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"callIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"checkpointIndex\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"reason\",\"type\":\"bytes\"}],\"name\":\"CheckpointBalanceReadFailed\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"callIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"patchIndex\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"id\",\"type\":\"bytes32\"}],\"name\":\"CheckpointNotFound\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"callIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"patchIndex\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"id\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"expected\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"actual\",\"type\":\"address\"}],\"name\":\"CheckpointTokenMismatch\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"target\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"reason\",\"type\":\"bytes\"}],\"name\":\"DynamicCallFailed\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DynamicExecutionReentered\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ECDSAInvalidSignature\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"length\",\"type\":\"uint256\"}],\"name\":\"ECDSAInvalidSignatureLength\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"ECDSAInvalidSignatureS\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"EmptyDynamicBatch\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"error\",\"type\":\"bytes\"}],\"name\":\"ExecuteError\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"callIndex\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"asset\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"pool\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"reason\",\"type\":\"bytes\"}],\"name\":\"FlashLoanAllowanceReadFailed\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"callIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"actual\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maximum\",\"type\":\"uint256\"}],\"name\":\"FlashLoanPremiumTooHigh\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"callIndex\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"asset\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"pool\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"reason\",\"type\":\"bytes\"}],\"name\":\"FlashLoanRepaymentApprovalFailed\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"callIndex\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"asset\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"actual\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"required\",\"type\":\"uint256\"}],\"name\":\"FlashLoanRepaymentBalanceInsufficient\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"callIndex\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"asset\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"reason\",\"type\":\"bytes\"}],\"name\":\"FlashLoanRepaymentBalanceReadFailed\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"callIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"patchIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"bps\",\"type\":\"uint256\"}],\"name\":\"InvalidBps\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"callIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"checkpointIndex\",\"type\":\"uint256\"}],\"name\":\"InvalidCheckpointId\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"callIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"checkpointIndex\",\"type\":\"uint256\"}],\"name\":\"InvalidCheckpointToken\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"callIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"patchIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"offset\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"dataLength\",\"type\":\"uint256\"}],\"name\":\"InvalidPatchOffset\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"callIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"patchIndex\",\"type\":\"uint256\"}],\"name\":\"InvalidPatchToken\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"callIndex\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"target\",\"type\":\"address\"}],\"name\":\"InvalidTarget\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"firstCallIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"secondCallIndex\",\"type\":\"uint256\"}],\"name\":\"MultipleExpectedCallbacks\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"outerCallIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"callbackCallIndex\",\"type\":\"uint256\"}],\"name\":\"NestedCallbackNotSupported\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"msgSender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"entity\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"entryPoint\",\"type\":\"address\"}],\"name\":\"NotFromEntryPoint\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"callIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"patchIndex\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"reason\",\"type\":\"bytes\"}],\"name\":\"PatchBalanceReadFailed\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"callIndex\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"asset\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"pool\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"remaining\",\"type\":\"uint256\"}],\"name\":\"ResidualFlashLoanAllowance\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"callIndex\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"expected\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"actual\",\"type\":\"address\"}],\"name\":\"UnexpectedCallbackInitiator\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"callIndex\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"expected\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"actual\",\"type\":\"address\"}],\"name\":\"UnexpectedCallbackSender\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"callIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"patchIndex\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"id\",\"type\":\"bytes32\"}],\"name\":\"UnexpectedCheckpointId\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"callIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"patchIndex\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"previous\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"current\",\"type\":\"uint256\"}],\"name\":\"UnsortedPatchOffset\",\"type\":\"error\"},{\"stateMutability\":\"payable\",\"type\":\"fallback\"},{\"inputs\":[],\"name\":\"entryPoint\",\"outputs\":[{\"internalType\":\"contractIEntryPoint\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"target\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"execute\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"target\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"internalType\":\"structBaseAccount.Call[]\",\"name\":\"calls\",\"type\":\"tuple[]\"}],\"name\":\"executeBatch\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"target\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"id\",\"type\":\"bytes32\"}],\"internalType\":\"structIDefiSimplify7702Account.BalanceCheckpoint[]\",\"name\":\"checkpointsBefore\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"checkpointId\",\"type\":\"bytes32\"},{\"internalType\":\"uint32\",\"name\":\"offset\",\"type\":\"uint32\"},{\"internalType\":\"uint16\",\"name\":\"bps\",\"type\":\"uint16\"},{\"internalType\":\"enumIDefiSimplify7702Account.BalanceSource\",\"name\":\"source\",\"type\":\"uint8\"}],\"internalType\":\"structIDefiSimplify7702Account.BalancePatch[]\",\"name\":\"patches\",\"type\":\"tuple[]\"},{\"internalType\":\"bool\",\"name\":\"expectsCallback\",\"type\":\"bool\"}],\"internalType\":\"structIDefiSimplify7702Account.DynamicCall[]\",\"name\":\"calls\",\"type\":\"tuple[]\"}],\"name\":\"executeBatchDynamic\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"asset\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"premium\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"initiator\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"params\",\"type\":\"bytes\"}],\"name\":\"executeOperation\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"success\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getNonce\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"hash\",\"type\":\"bytes32\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"}],\"name\":\"isValidSignature\",\"outputs\":[{\"internalType\":\"bytes4\",\"name\":\"magicValue\",\"type\":\"bytes4\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256[]\",\"name\":\"\",\"type\":\"uint256[]\"},{\"internalType\":\"uint256[]\",\"name\":\"\",\"type\":\"uint256[]\"},{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"name\":\"onERC1155BatchReceived\",\"outputs\":[{\"internalType\":\"bytes4\",\"name\":\"\",\"type\":\"bytes4\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"name\":\"onERC1155Received\",\"outputs\":[{\"internalType\":\"bytes4\",\"name\":\"\",\"type\":\"bytes4\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"name\":\"onERC721Received\",\"outputs\":[{\"internalType\":\"bytes4\",\"name\":\"\",\"type\":\"bytes4\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes4\",\"name\":\"id\",\"type\":\"bytes4\"}],\"name\":\"supportsInterface\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"initCode\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"callData\",\"type\":\"bytes\"},{\"internalType\":\"bytes32\",\"name\":\"accountGasLimits\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"preVerificationGas\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"gasFees\",\"type\":\"bytes32\"},{\"internalType\":\"bytes\",\"name\":\"paymasterAndData\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"}],\"internalType\":\"structPackedUserOperation\",\"name\":\"userOp\",\"type\":\"tuple\"},{\"internalType\":\"bytes32\",\"name\":\"userOpHash\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"missingAccountFunds\",\"type\":\"uint256\"}],\"name\":\"validateUserOp\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"validationData\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"stateMutability\":\"payable\",\"type\":\"receive\"}]",
}

// DefiSimplify7702AccountABI is the input ABI used to generate the binding from.
// Deprecated: Use DefiSimplify7702AccountMetaData.ABI instead.
var DefiSimplify7702AccountABI = DefiSimplify7702AccountMetaData.ABI

// DefiSimplify7702Account is an auto generated Go binding around an Ethereum contract.
type DefiSimplify7702Account struct {
	DefiSimplify7702AccountCaller     // Read-only binding to the contract
	DefiSimplify7702AccountTransactor // Write-only binding to the contract
	DefiSimplify7702AccountFilterer   // Log filterer for contract events
}

// DefiSimplify7702AccountCaller is an auto generated read-only Go binding around an Ethereum contract.
type DefiSimplify7702AccountCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DefiSimplify7702AccountTransactor is an auto generated write-only Go binding around an Ethereum contract.
type DefiSimplify7702AccountTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DefiSimplify7702AccountFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type DefiSimplify7702AccountFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DefiSimplify7702AccountSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type DefiSimplify7702AccountSession struct {
	Contract     *DefiSimplify7702Account // Generic contract binding to set the session for
	CallOpts     bind.CallOpts            // Call options to use throughout this session
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// DefiSimplify7702AccountCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type DefiSimplify7702AccountCallerSession struct {
	Contract *DefiSimplify7702AccountCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                  // Call options to use throughout this session
}

// DefiSimplify7702AccountTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type DefiSimplify7702AccountTransactorSession struct {
	Contract     *DefiSimplify7702AccountTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                  // Transaction auth options to use throughout this session
}

// DefiSimplify7702AccountRaw is an auto generated low-level Go binding around an Ethereum contract.
type DefiSimplify7702AccountRaw struct {
	Contract *DefiSimplify7702Account // Generic contract binding to access the raw methods on
}

// DefiSimplify7702AccountCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type DefiSimplify7702AccountCallerRaw struct {
	Contract *DefiSimplify7702AccountCaller // Generic read-only contract binding to access the raw methods on
}

// DefiSimplify7702AccountTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type DefiSimplify7702AccountTransactorRaw struct {
	Contract *DefiSimplify7702AccountTransactor // Generic write-only contract binding to access the raw methods on
}

// NewDefiSimplify7702Account creates a new instance of DefiSimplify7702Account, bound to a specific deployed contract.
func NewDefiSimplify7702Account(address common.Address, backend bind.ContractBackend) (*DefiSimplify7702Account, error) {
	contract, err := bindDefiSimplify7702Account(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &DefiSimplify7702Account{DefiSimplify7702AccountCaller: DefiSimplify7702AccountCaller{contract: contract}, DefiSimplify7702AccountTransactor: DefiSimplify7702AccountTransactor{contract: contract}, DefiSimplify7702AccountFilterer: DefiSimplify7702AccountFilterer{contract: contract}}, nil
}

// NewDefiSimplify7702AccountCaller creates a new read-only instance of DefiSimplify7702Account, bound to a specific deployed contract.
func NewDefiSimplify7702AccountCaller(address common.Address, caller bind.ContractCaller) (*DefiSimplify7702AccountCaller, error) {
	contract, err := bindDefiSimplify7702Account(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &DefiSimplify7702AccountCaller{contract: contract}, nil
}

// NewDefiSimplify7702AccountTransactor creates a new write-only instance of DefiSimplify7702Account, bound to a specific deployed contract.
func NewDefiSimplify7702AccountTransactor(address common.Address, transactor bind.ContractTransactor) (*DefiSimplify7702AccountTransactor, error) {
	contract, err := bindDefiSimplify7702Account(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &DefiSimplify7702AccountTransactor{contract: contract}, nil
}

// NewDefiSimplify7702AccountFilterer creates a new log filterer instance of DefiSimplify7702Account, bound to a specific deployed contract.
func NewDefiSimplify7702AccountFilterer(address common.Address, filterer bind.ContractFilterer) (*DefiSimplify7702AccountFilterer, error) {
	contract, err := bindDefiSimplify7702Account(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &DefiSimplify7702AccountFilterer{contract: contract}, nil
}

// bindDefiSimplify7702Account binds a generic wrapper to an already deployed contract.
func bindDefiSimplify7702Account(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := DefiSimplify7702AccountMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_DefiSimplify7702Account *DefiSimplify7702AccountRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _DefiSimplify7702Account.Contract.DefiSimplify7702AccountCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_DefiSimplify7702Account *DefiSimplify7702AccountRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _DefiSimplify7702Account.Contract.DefiSimplify7702AccountTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_DefiSimplify7702Account *DefiSimplify7702AccountRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _DefiSimplify7702Account.Contract.DefiSimplify7702AccountTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_DefiSimplify7702Account *DefiSimplify7702AccountCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _DefiSimplify7702Account.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_DefiSimplify7702Account *DefiSimplify7702AccountTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _DefiSimplify7702Account.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_DefiSimplify7702Account *DefiSimplify7702AccountTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _DefiSimplify7702Account.Contract.contract.Transact(opts, method, params...)
}

// EntryPoint is a free data retrieval call binding the contract method 0xb0d691fe.
//
// Solidity: function entryPoint() view returns(address)
func (_DefiSimplify7702Account *DefiSimplify7702AccountCaller) EntryPoint(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _DefiSimplify7702Account.contract.Call(opts, &out, "entryPoint")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// EntryPoint is a free data retrieval call binding the contract method 0xb0d691fe.
//
// Solidity: function entryPoint() view returns(address)
func (_DefiSimplify7702Account *DefiSimplify7702AccountSession) EntryPoint() (common.Address, error) {
	return _DefiSimplify7702Account.Contract.EntryPoint(&_DefiSimplify7702Account.CallOpts)
}

// EntryPoint is a free data retrieval call binding the contract method 0xb0d691fe.
//
// Solidity: function entryPoint() view returns(address)
func (_DefiSimplify7702Account *DefiSimplify7702AccountCallerSession) EntryPoint() (common.Address, error) {
	return _DefiSimplify7702Account.Contract.EntryPoint(&_DefiSimplify7702Account.CallOpts)
}

// GetNonce is a free data retrieval call binding the contract method 0xd087d288.
//
// Solidity: function getNonce() view returns(uint256)
func (_DefiSimplify7702Account *DefiSimplify7702AccountCaller) GetNonce(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _DefiSimplify7702Account.contract.Call(opts, &out, "getNonce")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetNonce is a free data retrieval call binding the contract method 0xd087d288.
//
// Solidity: function getNonce() view returns(uint256)
func (_DefiSimplify7702Account *DefiSimplify7702AccountSession) GetNonce() (*big.Int, error) {
	return _DefiSimplify7702Account.Contract.GetNonce(&_DefiSimplify7702Account.CallOpts)
}

// GetNonce is a free data retrieval call binding the contract method 0xd087d288.
//
// Solidity: function getNonce() view returns(uint256)
func (_DefiSimplify7702Account *DefiSimplify7702AccountCallerSession) GetNonce() (*big.Int, error) {
	return _DefiSimplify7702Account.Contract.GetNonce(&_DefiSimplify7702Account.CallOpts)
}

// IsValidSignature is a free data retrieval call binding the contract method 0x1626ba7e.
//
// Solidity: function isValidSignature(bytes32 hash, bytes signature) view returns(bytes4 magicValue)
func (_DefiSimplify7702Account *DefiSimplify7702AccountCaller) IsValidSignature(opts *bind.CallOpts, hash [32]byte, signature []byte) ([4]byte, error) {
	var out []interface{}
	err := _DefiSimplify7702Account.contract.Call(opts, &out, "isValidSignature", hash, signature)

	if err != nil {
		return *new([4]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([4]byte)).(*[4]byte)

	return out0, err

}

// IsValidSignature is a free data retrieval call binding the contract method 0x1626ba7e.
//
// Solidity: function isValidSignature(bytes32 hash, bytes signature) view returns(bytes4 magicValue)
func (_DefiSimplify7702Account *DefiSimplify7702AccountSession) IsValidSignature(hash [32]byte, signature []byte) ([4]byte, error) {
	return _DefiSimplify7702Account.Contract.IsValidSignature(&_DefiSimplify7702Account.CallOpts, hash, signature)
}

// IsValidSignature is a free data retrieval call binding the contract method 0x1626ba7e.
//
// Solidity: function isValidSignature(bytes32 hash, bytes signature) view returns(bytes4 magicValue)
func (_DefiSimplify7702Account *DefiSimplify7702AccountCallerSession) IsValidSignature(hash [32]byte, signature []byte) ([4]byte, error) {
	return _DefiSimplify7702Account.Contract.IsValidSignature(&_DefiSimplify7702Account.CallOpts, hash, signature)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 id) pure returns(bool)
func (_DefiSimplify7702Account *DefiSimplify7702AccountCaller) SupportsInterface(opts *bind.CallOpts, id [4]byte) (bool, error) {
	var out []interface{}
	err := _DefiSimplify7702Account.contract.Call(opts, &out, "supportsInterface", id)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 id) pure returns(bool)
func (_DefiSimplify7702Account *DefiSimplify7702AccountSession) SupportsInterface(id [4]byte) (bool, error) {
	return _DefiSimplify7702Account.Contract.SupportsInterface(&_DefiSimplify7702Account.CallOpts, id)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 id) pure returns(bool)
func (_DefiSimplify7702Account *DefiSimplify7702AccountCallerSession) SupportsInterface(id [4]byte) (bool, error) {
	return _DefiSimplify7702Account.Contract.SupportsInterface(&_DefiSimplify7702Account.CallOpts, id)
}

// Execute is a paid mutator transaction binding the contract method 0xb61d27f6.
//
// Solidity: function execute(address target, uint256 value, bytes data) returns()
func (_DefiSimplify7702Account *DefiSimplify7702AccountTransactor) Execute(opts *bind.TransactOpts, target common.Address, value *big.Int, data []byte) (*types.Transaction, error) {
	return _DefiSimplify7702Account.contract.Transact(opts, "execute", target, value, data)
}

// Execute is a paid mutator transaction binding the contract method 0xb61d27f6.
//
// Solidity: function execute(address target, uint256 value, bytes data) returns()
func (_DefiSimplify7702Account *DefiSimplify7702AccountSession) Execute(target common.Address, value *big.Int, data []byte) (*types.Transaction, error) {
	return _DefiSimplify7702Account.Contract.Execute(&_DefiSimplify7702Account.TransactOpts, target, value, data)
}

// Execute is a paid mutator transaction binding the contract method 0xb61d27f6.
//
// Solidity: function execute(address target, uint256 value, bytes data) returns()
func (_DefiSimplify7702Account *DefiSimplify7702AccountTransactorSession) Execute(target common.Address, value *big.Int, data []byte) (*types.Transaction, error) {
	return _DefiSimplify7702Account.Contract.Execute(&_DefiSimplify7702Account.TransactOpts, target, value, data)
}

// ExecuteBatch is a paid mutator transaction binding the contract method 0x34fcd5be.
//
// Solidity: function executeBatch((address,uint256,bytes)[] calls) returns()
func (_DefiSimplify7702Account *DefiSimplify7702AccountTransactor) ExecuteBatch(opts *bind.TransactOpts, calls []BaseAccountCall) (*types.Transaction, error) {
	return _DefiSimplify7702Account.contract.Transact(opts, "executeBatch", calls)
}

// ExecuteBatch is a paid mutator transaction binding the contract method 0x34fcd5be.
//
// Solidity: function executeBatch((address,uint256,bytes)[] calls) returns()
func (_DefiSimplify7702Account *DefiSimplify7702AccountSession) ExecuteBatch(calls []BaseAccountCall) (*types.Transaction, error) {
	return _DefiSimplify7702Account.Contract.ExecuteBatch(&_DefiSimplify7702Account.TransactOpts, calls)
}

// ExecuteBatch is a paid mutator transaction binding the contract method 0x34fcd5be.
//
// Solidity: function executeBatch((address,uint256,bytes)[] calls) returns()
func (_DefiSimplify7702Account *DefiSimplify7702AccountTransactorSession) ExecuteBatch(calls []BaseAccountCall) (*types.Transaction, error) {
	return _DefiSimplify7702Account.Contract.ExecuteBatch(&_DefiSimplify7702Account.TransactOpts, calls)
}

// ExecuteBatchDynamic is a paid mutator transaction binding the contract method 0xecadebe3.
//
// Solidity: function executeBatchDynamic((address,uint256,bytes,(address,bytes32)[],(address,bytes32,uint32,uint16,uint8)[],bool)[] calls) payable returns()
func (_DefiSimplify7702Account *DefiSimplify7702AccountTransactor) ExecuteBatchDynamic(opts *bind.TransactOpts, calls []IDefiSimplify7702AccountDynamicCall) (*types.Transaction, error) {
	return _DefiSimplify7702Account.contract.Transact(opts, "executeBatchDynamic", calls)
}

// ExecuteBatchDynamic is a paid mutator transaction binding the contract method 0xecadebe3.
//
// Solidity: function executeBatchDynamic((address,uint256,bytes,(address,bytes32)[],(address,bytes32,uint32,uint16,uint8)[],bool)[] calls) payable returns()
func (_DefiSimplify7702Account *DefiSimplify7702AccountSession) ExecuteBatchDynamic(calls []IDefiSimplify7702AccountDynamicCall) (*types.Transaction, error) {
	return _DefiSimplify7702Account.Contract.ExecuteBatchDynamic(&_DefiSimplify7702Account.TransactOpts, calls)
}

// ExecuteBatchDynamic is a paid mutator transaction binding the contract method 0xecadebe3.
//
// Solidity: function executeBatchDynamic((address,uint256,bytes,(address,bytes32)[],(address,bytes32,uint32,uint16,uint8)[],bool)[] calls) payable returns()
func (_DefiSimplify7702Account *DefiSimplify7702AccountTransactorSession) ExecuteBatchDynamic(calls []IDefiSimplify7702AccountDynamicCall) (*types.Transaction, error) {
	return _DefiSimplify7702Account.Contract.ExecuteBatchDynamic(&_DefiSimplify7702Account.TransactOpts, calls)
}

// ExecuteOperation is a paid mutator transaction binding the contract method 0x1b11d0ff.
//
// Solidity: function executeOperation(address asset, uint256 amount, uint256 premium, address initiator, bytes params) returns(bool success)
func (_DefiSimplify7702Account *DefiSimplify7702AccountTransactor) ExecuteOperation(opts *bind.TransactOpts, asset common.Address, amount *big.Int, premium *big.Int, initiator common.Address, params []byte) (*types.Transaction, error) {
	return _DefiSimplify7702Account.contract.Transact(opts, "executeOperation", asset, amount, premium, initiator, params)
}

// ExecuteOperation is a paid mutator transaction binding the contract method 0x1b11d0ff.
//
// Solidity: function executeOperation(address asset, uint256 amount, uint256 premium, address initiator, bytes params) returns(bool success)
func (_DefiSimplify7702Account *DefiSimplify7702AccountSession) ExecuteOperation(asset common.Address, amount *big.Int, premium *big.Int, initiator common.Address, params []byte) (*types.Transaction, error) {
	return _DefiSimplify7702Account.Contract.ExecuteOperation(&_DefiSimplify7702Account.TransactOpts, asset, amount, premium, initiator, params)
}

// ExecuteOperation is a paid mutator transaction binding the contract method 0x1b11d0ff.
//
// Solidity: function executeOperation(address asset, uint256 amount, uint256 premium, address initiator, bytes params) returns(bool success)
func (_DefiSimplify7702Account *DefiSimplify7702AccountTransactorSession) ExecuteOperation(asset common.Address, amount *big.Int, premium *big.Int, initiator common.Address, params []byte) (*types.Transaction, error) {
	return _DefiSimplify7702Account.Contract.ExecuteOperation(&_DefiSimplify7702Account.TransactOpts, asset, amount, premium, initiator, params)
}

// OnERC1155BatchReceived is a paid mutator transaction binding the contract method 0xbc197c81.
//
// Solidity: function onERC1155BatchReceived(address , address , uint256[] , uint256[] , bytes ) returns(bytes4)
func (_DefiSimplify7702Account *DefiSimplify7702AccountTransactor) OnERC1155BatchReceived(opts *bind.TransactOpts, arg0 common.Address, arg1 common.Address, arg2 []*big.Int, arg3 []*big.Int, arg4 []byte) (*types.Transaction, error) {
	return _DefiSimplify7702Account.contract.Transact(opts, "onERC1155BatchReceived", arg0, arg1, arg2, arg3, arg4)
}

// OnERC1155BatchReceived is a paid mutator transaction binding the contract method 0xbc197c81.
//
// Solidity: function onERC1155BatchReceived(address , address , uint256[] , uint256[] , bytes ) returns(bytes4)
func (_DefiSimplify7702Account *DefiSimplify7702AccountSession) OnERC1155BatchReceived(arg0 common.Address, arg1 common.Address, arg2 []*big.Int, arg3 []*big.Int, arg4 []byte) (*types.Transaction, error) {
	return _DefiSimplify7702Account.Contract.OnERC1155BatchReceived(&_DefiSimplify7702Account.TransactOpts, arg0, arg1, arg2, arg3, arg4)
}

// OnERC1155BatchReceived is a paid mutator transaction binding the contract method 0xbc197c81.
//
// Solidity: function onERC1155BatchReceived(address , address , uint256[] , uint256[] , bytes ) returns(bytes4)
func (_DefiSimplify7702Account *DefiSimplify7702AccountTransactorSession) OnERC1155BatchReceived(arg0 common.Address, arg1 common.Address, arg2 []*big.Int, arg3 []*big.Int, arg4 []byte) (*types.Transaction, error) {
	return _DefiSimplify7702Account.Contract.OnERC1155BatchReceived(&_DefiSimplify7702Account.TransactOpts, arg0, arg1, arg2, arg3, arg4)
}

// OnERC1155Received is a paid mutator transaction binding the contract method 0xf23a6e61.
//
// Solidity: function onERC1155Received(address , address , uint256 , uint256 , bytes ) returns(bytes4)
func (_DefiSimplify7702Account *DefiSimplify7702AccountTransactor) OnERC1155Received(opts *bind.TransactOpts, arg0 common.Address, arg1 common.Address, arg2 *big.Int, arg3 *big.Int, arg4 []byte) (*types.Transaction, error) {
	return _DefiSimplify7702Account.contract.Transact(opts, "onERC1155Received", arg0, arg1, arg2, arg3, arg4)
}

// OnERC1155Received is a paid mutator transaction binding the contract method 0xf23a6e61.
//
// Solidity: function onERC1155Received(address , address , uint256 , uint256 , bytes ) returns(bytes4)
func (_DefiSimplify7702Account *DefiSimplify7702AccountSession) OnERC1155Received(arg0 common.Address, arg1 common.Address, arg2 *big.Int, arg3 *big.Int, arg4 []byte) (*types.Transaction, error) {
	return _DefiSimplify7702Account.Contract.OnERC1155Received(&_DefiSimplify7702Account.TransactOpts, arg0, arg1, arg2, arg3, arg4)
}

// OnERC1155Received is a paid mutator transaction binding the contract method 0xf23a6e61.
//
// Solidity: function onERC1155Received(address , address , uint256 , uint256 , bytes ) returns(bytes4)
func (_DefiSimplify7702Account *DefiSimplify7702AccountTransactorSession) OnERC1155Received(arg0 common.Address, arg1 common.Address, arg2 *big.Int, arg3 *big.Int, arg4 []byte) (*types.Transaction, error) {
	return _DefiSimplify7702Account.Contract.OnERC1155Received(&_DefiSimplify7702Account.TransactOpts, arg0, arg1, arg2, arg3, arg4)
}

// OnERC721Received is a paid mutator transaction binding the contract method 0x150b7a02.
//
// Solidity: function onERC721Received(address , address , uint256 , bytes ) returns(bytes4)
func (_DefiSimplify7702Account *DefiSimplify7702AccountTransactor) OnERC721Received(opts *bind.TransactOpts, arg0 common.Address, arg1 common.Address, arg2 *big.Int, arg3 []byte) (*types.Transaction, error) {
	return _DefiSimplify7702Account.contract.Transact(opts, "onERC721Received", arg0, arg1, arg2, arg3)
}

// OnERC721Received is a paid mutator transaction binding the contract method 0x150b7a02.
//
// Solidity: function onERC721Received(address , address , uint256 , bytes ) returns(bytes4)
func (_DefiSimplify7702Account *DefiSimplify7702AccountSession) OnERC721Received(arg0 common.Address, arg1 common.Address, arg2 *big.Int, arg3 []byte) (*types.Transaction, error) {
	return _DefiSimplify7702Account.Contract.OnERC721Received(&_DefiSimplify7702Account.TransactOpts, arg0, arg1, arg2, arg3)
}

// OnERC721Received is a paid mutator transaction binding the contract method 0x150b7a02.
//
// Solidity: function onERC721Received(address , address , uint256 , bytes ) returns(bytes4)
func (_DefiSimplify7702Account *DefiSimplify7702AccountTransactorSession) OnERC721Received(arg0 common.Address, arg1 common.Address, arg2 *big.Int, arg3 []byte) (*types.Transaction, error) {
	return _DefiSimplify7702Account.Contract.OnERC721Received(&_DefiSimplify7702Account.TransactOpts, arg0, arg1, arg2, arg3)
}

// ValidateUserOp is a paid mutator transaction binding the contract method 0x19822f7c.
//
// Solidity: function validateUserOp((address,uint256,bytes,bytes,bytes32,uint256,bytes32,bytes,bytes) userOp, bytes32 userOpHash, uint256 missingAccountFunds) returns(uint256 validationData)
func (_DefiSimplify7702Account *DefiSimplify7702AccountTransactor) ValidateUserOp(opts *bind.TransactOpts, userOp PackedUserOperation, userOpHash [32]byte, missingAccountFunds *big.Int) (*types.Transaction, error) {
	return _DefiSimplify7702Account.contract.Transact(opts, "validateUserOp", userOp, userOpHash, missingAccountFunds)
}

// ValidateUserOp is a paid mutator transaction binding the contract method 0x19822f7c.
//
// Solidity: function validateUserOp((address,uint256,bytes,bytes,bytes32,uint256,bytes32,bytes,bytes) userOp, bytes32 userOpHash, uint256 missingAccountFunds) returns(uint256 validationData)
func (_DefiSimplify7702Account *DefiSimplify7702AccountSession) ValidateUserOp(userOp PackedUserOperation, userOpHash [32]byte, missingAccountFunds *big.Int) (*types.Transaction, error) {
	return _DefiSimplify7702Account.Contract.ValidateUserOp(&_DefiSimplify7702Account.TransactOpts, userOp, userOpHash, missingAccountFunds)
}

// ValidateUserOp is a paid mutator transaction binding the contract method 0x19822f7c.
//
// Solidity: function validateUserOp((address,uint256,bytes,bytes,bytes32,uint256,bytes32,bytes,bytes) userOp, bytes32 userOpHash, uint256 missingAccountFunds) returns(uint256 validationData)
func (_DefiSimplify7702Account *DefiSimplify7702AccountTransactorSession) ValidateUserOp(userOp PackedUserOperation, userOpHash [32]byte, missingAccountFunds *big.Int) (*types.Transaction, error) {
	return _DefiSimplify7702Account.Contract.ValidateUserOp(&_DefiSimplify7702Account.TransactOpts, userOp, userOpHash, missingAccountFunds)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() payable returns()
func (_DefiSimplify7702Account *DefiSimplify7702AccountTransactor) Fallback(opts *bind.TransactOpts, calldata []byte) (*types.Transaction, error) {
	return _DefiSimplify7702Account.contract.RawTransact(opts, calldata)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() payable returns()
func (_DefiSimplify7702Account *DefiSimplify7702AccountSession) Fallback(calldata []byte) (*types.Transaction, error) {
	return _DefiSimplify7702Account.Contract.Fallback(&_DefiSimplify7702Account.TransactOpts, calldata)
}

// Fallback is a paid mutator transaction binding the contract fallback function.
//
// Solidity: fallback() payable returns()
func (_DefiSimplify7702Account *DefiSimplify7702AccountTransactorSession) Fallback(calldata []byte) (*types.Transaction, error) {
	return _DefiSimplify7702Account.Contract.Fallback(&_DefiSimplify7702Account.TransactOpts, calldata)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_DefiSimplify7702Account *DefiSimplify7702AccountTransactor) Receive(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _DefiSimplify7702Account.contract.RawTransact(opts, nil) // calldata is disallowed for receive function
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_DefiSimplify7702Account *DefiSimplify7702AccountSession) Receive() (*types.Transaction, error) {
	return _DefiSimplify7702Account.Contract.Receive(&_DefiSimplify7702Account.TransactOpts)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_DefiSimplify7702Account *DefiSimplify7702AccountTransactorSession) Receive() (*types.Transaction, error) {
	return _DefiSimplify7702Account.Contract.Receive(&_DefiSimplify7702Account.TransactOpts)
}
