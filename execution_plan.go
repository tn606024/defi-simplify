package defi

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/tn606024/defi-simplify/amount"
)

var ErrPlanKindMismatch = errors.New("execution plan kind is incompatible with accessor")

// CheckpointDeclaration attaches a named token-balance checkpoint immediately
// before a planned call.
type CheckpointDeclaration struct {
	Ref amount.CheckpointRef
}

// CalldataPatch replaces one protocol-owned ABI uint256 word with a runtime
// amount immediately before execution.
type CalldataPatch struct {
	Source amount.Source
	Offset uint32
}

// PlannedCall combines one neutral call with optional dynamic execution
// metadata. Protocol packages own all patch offsets they place here.
type PlannedCall struct {
	Call              Call
	CheckpointsBefore []CheckpointDeclaration
	Patches           []CalldataPatch
	ExpectsCallback   bool
}

// PlanKind identifies the execution capability required by a built plan.
type PlanKind uint8

const (
	PlanStatic PlanKind = iota + 1
	PlanDynamic
)

// BalanceSource mirrors the dynamic account contract's runtime balance source.
type BalanceSource uint8

const (
	BalanceSourceCurrentBalance BalanceSource = iota
	BalanceSourceCheckpointDelta
)

// BalanceCheckpoint is the materialized contract checkpoint value.
type BalanceCheckpoint struct {
	Token common.Address
	ID    [32]byte
}

// BalancePatch is the materialized contract calldata patch value.
type BalancePatch struct {
	Token        common.Address
	CheckpointID [32]byte
	Offset       uint32
	BPS          uint16
	Source       BalanceSource
}

// DynamicCall is the protocol-neutral wire model for executeBatchDynamic.
type DynamicCall struct {
	Target            common.Address
	Value             *big.Int
	Data              []byte
	CheckpointsBefore []BalanceCheckpoint
	Patches           []BalancePatch
	ExpectsCallback   bool
}

// ExecutionPlan is the ordered, executor-neutral result of building a Flow.
// Its contents are immutable to callers and exposed through defensive-copy
// accessors. Account is the semantic caller identity for steps that derive
// owner, sender, or onBehalfOf values from BuildEnv.Account.
type ExecutionPlan struct {
	account      common.Address
	steps        []BuiltStep
	kind         PlanKind
	staticCalls  []Call
	dynamicCalls []DynamicCall
}

// Account returns the protocol-visible caller identity.
func (p *ExecutionPlan) Account() common.Address {
	if p == nil {
		return common.Address{}
	}
	return p.account
}

// Steps returns a defensive copy of the built steps.
func (p *ExecutionPlan) Steps() []BuiltStep {
	if p == nil {
		return nil
	}
	return cloneBuiltSteps(p.steps)
}

// Kind returns the execution capability required by the plan.
func (p *ExecutionPlan) Kind() PlanKind {
	if p == nil {
		return 0
	}
	return p.kind
}

// StaticCalls returns calls only when the plan has no runtime metadata.
func (p *ExecutionPlan) StaticCalls() ([]Call, error) {
	if p == nil {
		return nil, fmt.Errorf("%w: plan is nil", ErrPlanKindMismatch)
	}
	if p.kind != PlanStatic {
		return nil, fmt.Errorf("%w: expected static, got %d", ErrPlanKindMismatch, p.kind)
	}
	return cloneCalls(p.staticCalls), nil
}

// DynamicCalls returns materialized calls only when dynamic execution is
// required. Static calls within a dynamic plan are represented with empty
// checkpoint and patch metadata.
func (p *ExecutionPlan) DynamicCalls() ([]DynamicCall, error) {
	if p == nil {
		return nil, fmt.Errorf("%w: plan is nil", ErrPlanKindMismatch)
	}
	if p.kind != PlanDynamic {
		return nil, fmt.Errorf("%w: expected dynamic, got %d", ErrPlanKindMismatch, p.kind)
	}
	return cloneDynamicCalls(p.dynamicCalls), nil
}

// Calls returns static plan calls in step order.
//
// Deprecated: use StaticCalls, which reports an error for dynamic plans.
func (p *ExecutionPlan) Calls() []Call {
	calls, err := p.StaticCalls()
	if err != nil {
		return nil
	}
	return calls
}

func cloneDynamicCalls(calls []DynamicCall) []DynamicCall {
	cloned := make([]DynamicCall, len(calls))
	for i, call := range calls {
		cloned[i] = call
		cloned[i].Value = cloneBigIntOrZero(call.Value)
		cloned[i].Data = append([]byte(nil), call.Data...)
		cloned[i].CheckpointsBefore = append([]BalanceCheckpoint(nil), call.CheckpointsBefore...)
		cloned[i].Patches = append([]BalancePatch(nil), call.Patches...)
	}
	return cloned
}
