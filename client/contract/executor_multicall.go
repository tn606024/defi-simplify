package contract

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/tn606024/defi-simplify/bind/multicall"
	"github.com/tn606024/defi-simplify/config"
)

//go:embed abi/multicall/IMulticall3.json
var multicallABI string

type MulticallExecutor struct {
	conn  EthereumClient
	chain config.Chain
	opts  *bind.TransactOpts
}

var _ ActionExecutor = (*MulticallExecutor)(nil)
var _ CallExecutor = (*MulticallExecutor)(nil)

type MulticallAction struct {
	BaseAction
	target common.Address
	calls  []multicall.IMulticall3Call3
}

func NewMulticallExecutor(conn EthereumClient, chain config.Chain, opts *bind.TransactOpts) *MulticallExecutor {
	return &MulticallExecutor{
		conn:  conn,
		chain: chain,
		opts:  opts,
	}
}

// ExecuteActions executes action calls through Multicall3 aggregate3.
// ErrTransactionReverted reports an outer transaction revert; allowed inner
// call failures remain part of a successful aggregate3 transaction.
func (e *MulticallExecutor) ExecuteActions(ctx context.Context, actions []ExecuteAction) (*types.Receipt, error) {
	calls, err := e.ToMulticall3Calls(ctx, actions)
	if err != nil {
		return nil, err
	}
	return e.executeMulticallCalls(ctx, calls)
}

// ExecuteCalls executes neutral calls through Multicall3 aggregate3.
func (e *MulticallExecutor) ExecuteCalls(ctx context.Context, calls []Call) (*types.Receipt, error) {
	multicallCalls, err := e.CallsToMulticall3Calls(calls)
	if err != nil {
		return nil, err
	}
	return e.executeMulticallCalls(ctx, multicallCalls)
}

func (e *MulticallExecutor) executeMulticallCalls(ctx context.Context, calls []multicall.IMulticall3Call3) (*types.Receipt, error) {
	multicallAddress, err := e.chain.MulticallAddress()
	if err != nil {
		return nil, err
	}
	multicallAction := BuildMulticallAction(multicallAddress, calls)
	receipt, err := executeAction(ctx, e.conn, e.opts, multicallAction)
	if err != nil {
		return receipt, err
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return receipt, fmt.Errorf("%w: tx %s", ErrTransactionReverted, receipt.TxHash.Hex())
	}
	return receipt, nil
}

func (e *MulticallExecutor) ExecuteReadActions(ctx context.Context, actions []Action) ([]multicall.IMulticall3Result, error) {
	calls := make([]multicall.IMulticall3Call3, 0, len(actions))
	for _, action := range actions {
		call, err := e.actionToMulticall3Call(ctx, action, false)
		if err != nil {
			log.Printf("Failed to convert action to multicall call: %v", err)
			return nil, err
		}
		calls = append(calls, call)
	}
	multicallAddress, err := e.chain.MulticallAddress()
	if err != nil {
		return nil, err
	}
	multicallAction := BuildMulticallAction(multicallAddress, calls)

	multicallAddr, data, err := multicallAction.ToData(ctx, e.conn, e.opts)
	if err != nil {
		log.Printf("Failed to convert multicall action to data: %v", err)
		return nil, err
	}

	result, err := e.conn.CallContract(ctx, ethereum.CallMsg{
		To:   &multicallAddr,
		Data: data,
	}, nil)

	if err != nil {
		log.Printf("Failed to execute multicall: %v", err)
		return nil, err
	}

	parsed, err := abi.JSON(strings.NewReader(multicallABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse multicall ABI: %v", err)
	}

	var returnData []struct {
		Success    bool   `abi:"success"`
		ReturnData []byte `abi:"returnData"`
	}

	err = parsed.UnpackIntoInterface(&returnData, "aggregate3", result)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack multicall result: %v", err)
	}

	results := make([]multicall.IMulticall3Result, len(returnData))
	for i, data := range returnData {
		results[i] = multicall.IMulticall3Result{
			Success:    data.Success,
			ReturnData: data.ReturnData,
		}
	}

	return results, nil
}

func (e *MulticallExecutor) ToMulticall3Calls(ctx context.Context, actions []ExecuteAction) ([]multicall.IMulticall3Call3, error) {
	calls := make([]multicall.IMulticall3Call3, 0, len(actions))
	for _, action := range actions {
		call, err := e.actionToMulticall3Call(ctx, action, action.AllowFailure())
		if err != nil {
			return nil, err
		}
		calls = append(calls, call)
	}
	return calls, nil
}

// CallsToMulticall3Calls converts neutral calls into Multicall3 aggregate3 calls.
func (e *MulticallExecutor) CallsToMulticall3Calls(calls []Call) ([]multicall.IMulticall3Call3, error) {
	multicallCalls := make([]multicall.IMulticall3Call3, 0, len(calls))
	for _, call := range calls {
		multicallCall, err := callToMulticall3Call(call, false)
		if err != nil {
			return nil, err
		}
		multicallCalls = append(multicallCalls, multicallCall)
	}
	return multicallCalls, nil
}

func (e *MulticallExecutor) actionToMulticall3Call(ctx context.Context, action Action, allowFailure bool) (multicall.IMulticall3Call3, error) {
	call, err := action.ToCall(ctx, e.conn, e.opts)
	if err != nil {
		return multicall.IMulticall3Call3{}, fmt.Errorf("failed to get action call: %w", err)
	}
	return callToMulticall3Call(*call, allowFailure)
}

func callToMulticall3Call(call Call, allowFailure bool) (multicall.IMulticall3Call3, error) {
	if hasValue(call.Value) {
		return multicall.IMulticall3Call3{}, fmt.Errorf("aggregate3 multicall does not support call value for target %s", call.Target.Hex())
	}
	return multicall.IMulticall3Call3{
		Target:       call.Target,
		CallData:     call.Data,
		AllowFailure: allowFailure,
	}, nil
}

func hasValue(value *big.Int) bool {
	return value != nil && value.Sign() != 0
}

func (a *MulticallAction) ToTransaction(ctx context.Context, conn EthereumClient, opt *bind.TransactOpts) (*types.Transaction, error) {
	multicallInterface, err := multicall.NewIMulticall3(a.target, conn)
	if err != nil {
		return nil, err
	}
	return multicallInterface.Aggregate3(opt, a.calls)
}

func (a *MulticallAction) ToData(ctx context.Context, conn EthereumClient, opt *bind.TransactOpts) (common.Address, []byte, error) {
	parsed, err := abi.JSON(strings.NewReader(multicallABI))
	if err != nil {
		return common.Address{}, nil, err
	}
	data, err := parsed.Pack("aggregate3", a.calls)
	if err != nil {
		return common.Address{}, nil, err
	}
	return a.target, data, nil
}
