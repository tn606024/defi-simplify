package defi

import (
	"context"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/tn606024/defi-simplify/client/contract"
	"github.com/tn606024/defi-simplify/config"
)

var (
	ErrEmptyFlow       = errors.New("empty flow")
	ErrMissingChain    = errors.New("flow chain is required")
	ErrInvalidAccount  = errors.New("flow account is zero")
	ErrMissingExecutor = errors.New("flow executor is required")
)

// Call is the neutral contract call model shared by flow builders and executors.
type Call = contract.Call

// Action is an existing protocol action that can encode itself into a Call.
type Action = contract.Action

// EthereumClient is the client interface needed by steps that build calls from contract state.
type EthereumClient = contract.EthereumClient

// CallExecutor executes Flow-built calls.
type CallExecutor = contract.CallExecutor

// FlowStep builds one or more neutral calls using the shared flow build context.
type FlowStep interface {
	BuildCalls(ctx context.Context, env BuildEnv) ([]Call, error)
}

// BuildEnv contains shared context passed to every step during Flow.Build.
type BuildEnv struct {
	Account common.Address
	Chain   config.Chain
	Conn    EthereumClient
}

// Flow is a static ordered composition of DeFi steps.
type Flow struct {
	account  common.Address
	chain    config.Chain
	chainSet bool
	steps    []FlowStep
}

// FlowOption configures a Flow.
type FlowOption func(*Flow)

// WithChain configures the chain context used when building flow steps.
func WithChain(chain config.Chain) FlowOption {
	return func(f *Flow) {
		f.chain = chain
		f.chainSet = true
	}
}

// NewFlow creates an empty static flow for account.
func NewFlow(account common.Address, opts ...FlowOption) *Flow {
	flow := &Flow{
		account: account,
	}
	for _, opt := range opts {
		opt(flow)
	}
	return flow
}

// Add appends a step and returns the flow for fluent composition.
func (f *Flow) Add(step FlowStep) *Flow {
	f.steps = append(f.steps, step)
	return f
}

// Build converts all flow steps into an ordered list of neutral calls.
func (f *Flow) Build(ctx context.Context, conn EthereumClient) ([]Call, error) {
	if f == nil {
		return nil, errors.New("flow is nil")
	}
	if f.account == (common.Address{}) {
		return nil, ErrInvalidAccount
	}
	if !f.chainSet {
		return nil, ErrMissingChain
	}
	if _, err := f.chain.Name(); err != nil {
		return nil, fmt.Errorf("flow chain: %w", err)
	}
	if len(f.steps) == 0 {
		return nil, ErrEmptyFlow
	}

	env := BuildEnv{
		Account: f.account,
		Chain:   f.chain,
		Conn:    conn,
	}
	calls := make([]Call, 0, len(f.steps))
	for i, step := range f.steps {
		if step == nil {
			return nil, fmt.Errorf("build flow step %d: nil step", i+1)
		}
		stepCalls, err := step.BuildCalls(ctx, env)
		if err != nil {
			return nil, fmt.Errorf("build flow step %d %s: %w", i+1, flowStepName(step), err)
		}
		calls = append(calls, stepCalls...)
	}
	return calls, nil
}

// Execute builds the flow and executes the resulting calls through executor.
func (f *Flow) Execute(ctx context.Context, conn EthereumClient, executor CallExecutor) (*types.Receipt, error) {
	if executor == nil {
		return nil, ErrMissingExecutor
	}
	calls, err := f.Build(ctx, conn)
	if err != nil {
		return nil, err
	}
	return executor.ExecuteCalls(ctx, calls)
}

type namedFlowStep interface {
	FlowStepName() string
}

func flowStepName(step FlowStep) string {
	if named, ok := step.(namedFlowStep); ok {
		if name := named.FlowStepName(); name != "" {
			return name
		}
	}
	return fmt.Sprintf("%T", step)
}

type actionFlowStep struct {
	name   string
	action Action
}

// ActionStep adapts an existing Action into a FlowStep.
func ActionStep(name string, action Action) FlowStep {
	return &actionFlowStep{
		name:   name,
		action: action,
	}
}

func (s *actionFlowStep) FlowStepName() string {
	if s.name != "" {
		return s.name
	}
	return "ActionStep"
}

func (s *actionFlowStep) BuildCalls(ctx context.Context, env BuildEnv) ([]Call, error) {
	if s.action == nil {
		return nil, errors.New("action is nil")
	}
	call, err := s.action.ToCall(ctx, env.Conn, nil)
	if err != nil {
		return nil, err
	}
	if call == nil {
		return nil, errors.New("action returned nil call")
	}
	return []Call{*call}, nil
}
