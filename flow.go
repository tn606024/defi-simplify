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

// FlowStep builds one named step from shared Flow context.
type FlowStep interface {
	Build(ctx context.Context, env BuildEnv) (BuiltStep, error)
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

// Build compiles all Flow steps into an ordered execution plan.
func (f *Flow) Build(ctx context.Context, conn EthereumClient) (*ExecutionPlan, error) {
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
	plan := &ExecutionPlan{
		Account: f.account,
		Steps:   make([]BuiltStep, 0, len(f.steps)),
	}
	nameCounts := make(map[string]int, len(f.steps))
	for i, step := range f.steps {
		if step == nil {
			return nil, fmt.Errorf("build flow step %d: nil step", i+1)
		}
		built, err := step.Build(ctx, env)
		if err != nil {
			name := built.Name
			if name == "" {
				name = fmt.Sprintf("%T", step)
			}
			return nil, fmt.Errorf("build flow step %d %s: %w", i+1, name, err)
		}
		if built.Name == "" {
			return nil, fmt.Errorf("build flow step %d: step name is required", i+1)
		}
		if len(built.Calls) == 0 {
			return nil, fmt.Errorf("build flow step %d %s: step returned no calls", i+1, built.Name)
		}
		nameCounts[built.Name]++
		built.ID = StepID(fmt.Sprintf("%s#%d", built.Name, nameCounts[built.Name]))
		built.Calls = cloneCalls(built.Calls)
		built.Expectations = append([]EventExpectation(nil), built.Expectations...)
		plan.Steps = append(plan.Steps, built)
	}
	return plan, nil
}

// Execute builds the flow and executes the resulting calls through executor.
func (f *Flow) Execute(ctx context.Context, conn EthereumClient, executor CallExecutor) (*types.Receipt, error) {
	if executor == nil {
		return nil, ErrMissingExecutor
	}
	plan, err := f.Build(ctx, conn)
	if err != nil {
		return nil, err
	}
	return executor.ExecuteCalls(ctx, plan.Calls())
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

func (s *actionFlowStep) stepName() string {
	if s.name != "" {
		return s.name
	}
	return "ActionStep"
}

func (s *actionFlowStep) Build(ctx context.Context, env BuildEnv) (BuiltStep, error) {
	built := BuiltStep{Name: s.stepName()}
	if s.action == nil {
		return built, errors.New("action is nil")
	}
	call, err := s.action.ToCall(ctx, env.Conn, nil)
	if err != nil {
		return built, err
	}
	if call == nil {
		return built, errors.New("action returned nil call")
	}
	built.Calls = []Call{*call}
	return built, nil
}

func cloneCalls(calls []Call) []Call {
	cloned := make([]Call, len(calls))
	for i, call := range calls {
		cloned[i] = cloneCall(call)
	}
	return cloned
}
