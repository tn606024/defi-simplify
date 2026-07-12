package simple7702

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/tn606024/defi-simplify/client/account/eip7702"
	"github.com/tn606024/defi-simplify/client/contract"
)

var ErrEmptyBatch = errors.New("Simple7702Account executor requires at least one call")

// ExecutionResult describes a completed delegated-account batch execution.
type ExecutionResult struct {
	Receipt        *types.Receipt
	Account        common.Address
	Implementation common.Address
	CallCount      int
}

// Executor executes neutral calls through an EOA delegated to Simple7702Account.
type Executor struct {
	conn           contract.EthereumClient
	opts           *bind.TransactOpts
	implementation common.Address
}

var _ contract.CallExecutor = (*Executor)(nil)

// NewExecutor creates an executor for opts.From and the expected delegated implementation.
func NewExecutor(conn contract.EthereumClient, opts *bind.TransactOpts, implementation common.Address) *Executor {
	return &Executor{
		conn:           conn,
		opts:           opts,
		implementation: implementation,
	}
}

// ExecuteCalls executes calls atomically through Simple7702Account.executeBatch.
func (e *Executor) ExecuteCalls(ctx context.Context, calls []contract.Call) (*types.Receipt, error) {
	result, err := e.ExecuteCallsWithResult(ctx, calls)
	if result == nil {
		return nil, err
	}
	return result.Receipt, err
}

// ExecuteCallsWithResult executes calls and returns delegated-account metadata.
//
// Delegation is checked against pending state before submission. This preflight
// check cannot eliminate the EIP-7702 lifecycle race between validation and
// transaction inclusion; callers must coordinate delegation changes per EOA.
func (e *Executor) ExecuteCallsWithResult(ctx context.Context, calls []contract.Call) (*ExecutionResult, error) {
	if e == nil {
		return nil, errors.New("Simple7702Account executor is nil")
	}
	if len(calls) == 0 {
		return nil, ErrEmptyBatch
	}
	if e.conn == nil {
		return nil, errors.New("ethereum client is nil")
	}
	if e.opts == nil {
		return nil, errors.New("transaction options are nil")
	}
	if e.opts.Signer == nil {
		return nil, errors.New("transaction signer is nil")
	}
	if e.opts.From == (common.Address{}) {
		return nil, errors.New("delegated EOA is zero")
	}
	if e.implementation == (common.Address{}) {
		return nil, errors.New("Simple7702Account implementation is zero")
	}
	if err := eip7702.AssertPendingDelegatedTo(ctx, e.conn, e.opts.From, e.implementation); err != nil {
		return nil, fmt.Errorf("verify Simple7702Account delegation: %w", err)
	}

	data, err := EncodeExecuteBatch(calls)
	if err != nil {
		return nil, fmt.Errorf("encode Simple7702Account batch: %w", err)
	}
	batchCall := contract.Call{
		Target: e.opts.From,
		Value:  big.NewInt(0),
		Data:   data,
	}
	receipt, err := contract.NewDirectExecutor(e.conn, e.opts).ExecuteCalls(ctx, []contract.Call{batchCall})
	result := &ExecutionResult{
		Receipt:        receipt,
		Account:        e.opts.From,
		Implementation: e.implementation,
		CallCount:      len(calls),
	}
	if err != nil {
		return result, fmt.Errorf("execute Simple7702Account batch: %w", err)
	}
	return result, nil
}
