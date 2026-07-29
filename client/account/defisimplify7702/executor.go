package defisimplify7702

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/tn606024/defi-simplify/client/account/eip7702"
	"github.com/tn606024/defi-simplify/client/contract"
)

// ExecutionResult describes a completed delegated-account execution.
type ExecutionResult struct {
	Receipt        *types.Receipt
	Account        common.Address
	Implementation common.Address
	CallCount      int
}

// Executor executes calls through an EOA delegated to a reviewed
// DefiSimplify7702Account deployment.
type Executor struct {
	conn           contract.EthereumClient
	opts           *bind.TransactOpts
	implementation common.Address
}

// NewExecutor creates an executor for opts.From and implementation.
func NewExecutor(
	conn contract.EthereumClient,
	opts *bind.TransactOpts,
	implementation common.Address,
) *Executor {
	return &Executor{
		conn:           conn,
		opts:           opts,
		implementation: implementation,
	}
}

// ExecuteBatch executes exact calls through inherited executeBatch.
func (e *Executor) ExecuteBatch(ctx context.Context, calls []contract.Call) (*types.Receipt, error) {
	result, err := e.ExecuteBatchWithResult(ctx, calls)
	if result == nil {
		return nil, err
	}
	return result.Receipt, err
}

// ExecuteBatchWithResult validates and executes inherited executeBatch.
func (e *Executor) ExecuteBatchWithResult(
	ctx context.Context,
	calls []contract.Call,
) (*ExecutionResult, error) {
	data, err := EncodeExecuteBatch(calls)
	if err != nil {
		return nil, fmt.Errorf("encode DefiSimplify7702Account static batch: %w", err)
	}
	return e.executeEncodedBatch(ctx, data, len(calls), "static")
}

// ExecuteBatchDynamic executes runtime-dependent calls through
// executeBatchDynamic.
func (e *Executor) ExecuteBatchDynamic(ctx context.Context, calls []DynamicCall) (*types.Receipt, error) {
	result, err := e.ExecuteBatchDynamicWithResult(ctx, calls)
	if result == nil {
		return nil, err
	}
	return result.Receipt, err
}

// ExecuteBatchDynamicWithResult validates and executes executeBatchDynamic.
func (e *Executor) ExecuteBatchDynamicWithResult(
	ctx context.Context,
	calls []DynamicCall,
) (*ExecutionResult, error) {
	data, err := EncodeExecuteBatchDynamic(calls)
	if err != nil {
		return nil, fmt.Errorf("encode DefiSimplify7702Account dynamic batch: %w", err)
	}
	return e.executeEncodedBatch(ctx, data, len(calls), "dynamic")
}

// ExecuteCalls executes dynamic calls and returns their mined receipt.
//
// Deprecated: use ExecuteBatchDynamic.
func (e *Executor) ExecuteCalls(ctx context.Context, calls []DynamicCall) (*types.Receipt, error) {
	return e.ExecuteBatchDynamic(ctx, calls)
}

// ExecuteCallsWithResult validates and executes executeBatchDynamic.
//
// Deprecated: use ExecuteBatchDynamicWithResult.
func (e *Executor) ExecuteCallsWithResult(
	ctx context.Context,
	calls []DynamicCall,
) (*ExecutionResult, error) {
	return e.ExecuteBatchDynamicWithResult(ctx, calls)
}

// executeEncodedBatch owns the common delegated-account preflight, outer
// self-call submission, receipt preservation, and revert decoding path.
//
// The account ABI shape is checked before chain access. The implementation
// address is a reviewed, pinned deployment selected by the caller. The pending
// delegation target is checked before every submission because it can change
// between executions. That pending-state check cannot eliminate the EIP-7702
// lifecycle race between preflight and inclusion; callers must coordinate
// delegation changes for each EOA. Delegation remains installed until
// explicitly replaced or cleared, even when execution reverts.
//
// The outer self-call carries zero value. Inner call values are paid from the
// delegated EOA's existing native-token balance.
func (e *Executor) executeEncodedBatch(
	ctx context.Context,
	data []byte,
	callCount int,
	batchKind string,
) (*ExecutionResult, error) {
	if e == nil {
		return nil, errors.New("DefiSimplify7702Account executor is nil")
	}
	if err := e.validateConfiguration(); err != nil {
		return nil, err
	}

	if err := eip7702.AssertPendingDelegatedTo(
		ctx,
		e.conn,
		e.opts.From,
		e.implementation,
	); err != nil {
		return nil, fmt.Errorf("verify DefiSimplify7702Account delegation: %w", err)
	}

	batchCall := contract.Call{
		Target: e.opts.From,
		Value:  new(big.Int),
		Data:   data,
	}
	receipt, executionErr := contract.NewDirectExecutor(e.conn, e.opts).
		ExecuteCalls(ctx, []contract.Call{batchCall})
	if receipt == nil {
		if executionErr == nil {
			return nil, fmt.Errorf(
				"execute DefiSimplify7702Account %s batch: missing transaction receipt",
				batchKind,
			)
		}
		return nil, fmt.Errorf(
			"execute DefiSimplify7702Account %s batch: %w",
			batchKind,
			withDecodedContractError(executionErr),
		)
	}

	result := &ExecutionResult{
		Receipt:        receipt,
		Account:        e.opts.From,
		Implementation: e.implementation,
		CallCount:      callCount,
	}
	if executionErr != nil {
		executionErr = e.withReplayedContractError(ctx, batchCall, receipt, executionErr)
		return result, fmt.Errorf(
			"execute DefiSimplify7702Account %s batch: %w",
			batchKind,
			executionErr,
		)
	}
	return result, nil
}

func (e *Executor) validateConfiguration() error {
	if e.conn == nil {
		return errors.New("ethereum client is nil")
	}
	if e.opts == nil {
		return errors.New("transaction options are nil")
	}
	if e.opts.Signer == nil {
		return errors.New("transaction signer is nil")
	}
	if e.opts.From == (common.Address{}) {
		return errors.New("delegated EOA is zero")
	}
	if e.implementation == (common.Address{}) {
		return errors.New("DefiSimplify7702Account implementation is zero")
	}
	return nil
}

func (e *Executor) withReplayedContractError(
	ctx context.Context,
	call contract.Call,
	receipt *types.Receipt,
	executionErr error,
) error {
	if receipt == nil || receipt.BlockNumber == nil {
		return executionErr
	}
	_, replayErr := e.conn.CallContract(ctx, ethereum.CallMsg{
		From:  e.opts.From,
		To:    &call.Target,
		Value: call.Value,
		Data:  call.Data,
	}, receipt.BlockNumber)
	if replayErr == nil {
		return executionErr
	}
	return errors.Join(executionErr, decodedContractError(replayErr))
}

func withDecodedContractError(err error) error {
	decoded := decodedContractError(err)
	if decoded == nil {
		return err
	}
	return errors.Join(err, decoded)
}

func decodedContractError(err error) error {
	data, ok := ethclient.RevertErrorData(err)
	if !ok {
		return nil
	}
	decoded, decodeErr := DecodeContractError(data)
	if decodeErr != nil {
		return nil
	}
	return decoded
}
