package defi

import (
	"context"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/tn606024/defi-simplify/client/account/simple7702"
	"github.com/tn606024/defi-simplify/client/contract"
	"github.com/tn606024/defi-simplify/config"
)

// ErrExecutionAccountMismatch is returned when a Flow is built for an account other than the transaction signer.
var ErrExecutionAccountMismatch = errors.New("flow account does not match transaction signer")

// ExecutionMode describes user-facing execution semantics.
type ExecutionMode string

const (
	// ExecutionEOA executes a one-call Flow as a normal EOA transaction.
	ExecutionEOA ExecutionMode = "eoa"
	// ExecutionAtomicEOA executes a Flow atomically through a delegated EOA.
	ExecutionAtomicEOA ExecutionMode = "atomic_eoa"
)

// Runner executes Flows using user-facing execution modes.
type Runner struct {
	conn  EthereumClient
	opts  *bind.TransactOpts
	chain config.Chain
}

// NewRunner creates a Flow runner for a chain and transaction signer.
func NewRunner(conn EthereumClient, opts *bind.TransactOpts, chain config.Chain) *Runner {
	return &Runner{
		conn:  conn,
		opts:  opts,
		chain: chain,
	}
}

// Execute builds and executes flow using mode.
func (r *Runner) Execute(ctx context.Context, flow *Flow, mode ExecutionMode) (*types.Receipt, error) {
	if r == nil {
		return nil, errors.New("runner is nil")
	}
	if flow != nil && r.opts != nil && flow.account != r.opts.From {
		return nil, fmt.Errorf("%w: flow account %s, signer %s", ErrExecutionAccountMismatch, flow.account.Hex(), r.opts.From.Hex())
	}

	var executor CallExecutor
	switch mode {
	case ExecutionEOA:
		executor = contract.NewDirectExecutor(r.conn, r.opts)
	case ExecutionAtomicEOA:
		implementation, err := r.chain.Simple7702AccountImplementationAddress()
		if err != nil {
			return nil, fmt.Errorf("resolve Simple7702Account implementation: %w", err)
		}
		executor = simple7702.NewExecutor(r.conn, r.opts, implementation)
	default:
		return nil, fmt.Errorf("unsupported execution mode %q", mode)
	}

	return flow.Execute(ctx, r.conn, executor)
}
