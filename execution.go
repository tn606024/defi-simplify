package defi

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// StepID identifies one built occurrence of a named Flow step.
type StepID string

// BuiltStep contains the calls and semantic expectations produced from the same
// resolved Flow step data. Flow.Build owns ID assignment; step implementations
// must set Name and leave ID empty.
type BuiltStep struct {
	ID           StepID
	Name         string
	Calls        []Call
	Expectations []EventExpectation
}

// ExecutionPlan is the ordered, executor-neutral result of building a Flow.
type ExecutionPlan struct {
	Account common.Address
	Steps   []BuiltStep
}

// Calls returns the plan's calls in step order.
func (p *ExecutionPlan) Calls() []Call {
	if p == nil {
		return nil
	}

	callCount := 0
	for _, step := range p.Steps {
		callCount += len(step.Calls)
	}
	calls := make([]Call, 0, callCount)
	for _, step := range p.Steps {
		for _, call := range step.Calls {
			calls = append(calls, cloneCall(call))
		}
	}
	return calls
}

// EventMetadata identifies a decoded protocol event without depending on its
// protocol-specific fields.
type EventMetadata struct {
	Protocol string
	Name     string
	Emitter  common.Address
	LogIndex uint
}

// DecodedEvent is implemented by protocol-specific event result types.
type DecodedEvent interface {
	EventMetadata() EventMetadata
}

// MatchDecision describes whether a decoded candidate satisfies an
// expectation. The zero value is MatchSkip so matching fails closed.
type MatchDecision uint8

const (
	// MatchSkip rejects the current candidate without aborting receipt scanning.
	MatchSkip MatchDecision = iota
	// MatchAccepted accepts and consumes the current candidate log.
	MatchAccepted
)

// FieldMismatch explains one expected field that did not match its actual
// decoded value.
type FieldMismatch struct {
	Field    string
	Expected string
	Actual   string
}

// MatchResult is shared by event expectations and value constraints. At the
// constraint level MatchSkip means the value is unsatisfied; the enclosing
// event expectation decides whether receipt scanning continues.
type MatchResult struct {
	Decision   MatchDecision
	Mismatches []FieldMismatch
}

// MatchContext provides execution-plan context to expectation and constraint
// matching. It intentionally has no cross-step value resolution in Phase 1.
type MatchContext struct {
	StepID StepID
}

// EventExpectation decodes and validates one expected protocol event.
//
// IsCandidate must only identify logs by stable source data such as emitter and
// topic. Decode errors and Match errors are hard failures. Ordinary decoded
// field mismatches must return MatchSkip with mismatch details and a nil error.
// Within one BuiltStep, expectations must be declared in the same order as the
// corresponding events are emitted on-chain; the validator scans forward and
// never reuses an earlier or consumed log.
type EventExpectation interface {
	ExpectationName() string
	IsCandidate(log *types.Log) bool
	Decode(log *types.Log) (DecodedEvent, error)
	Match(event DecodedEvent, ctx MatchContext) (MatchResult, error)
}

// AmountConstraint validates a decoded event amount. Implementations must
// return MatchSkip rather than panic when actual is nil.
type AmountConstraint interface {
	Describe() string
	Match(actual *big.Int, ctx MatchContext) (MatchResult, error)
}

func cloneCall(call Call) Call {
	cloned := call
	if call.Value != nil {
		cloned.Value = new(big.Int).Set(call.Value)
	}
	cloned.Data = append([]byte(nil), call.Data...)
	return cloned
}
