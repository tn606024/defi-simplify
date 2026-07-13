package defi

import (
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/core/types"
)

var (
	// ErrInvalidExecutionPlan is returned when validation receives no usable plan.
	ErrInvalidExecutionPlan = errors.New("invalid execution plan")
	// ErrInvalidExecutionReceipt is returned when a receipt is nil, incomplete, or unsuccessful.
	ErrInvalidExecutionReceipt = errors.New("invalid execution receipt")
	// ErrInvalidEventExpectation is returned for nil or unnamed expectations.
	ErrInvalidEventExpectation = errors.New("invalid event expectation")
	// ErrMalformedExecutionEvent is returned when a candidate log cannot be decoded.
	ErrMalformedExecutionEvent = errors.New("malformed execution event")
	// ErrExpectedEventNotFound is returned when no candidate satisfies an expectation.
	ErrExpectedEventNotFound = errors.New("expected execution event not found")
	// ErrInvalidMatchResult is returned when an expectation or constraint violates the matching contract.
	ErrInvalidMatchResult = errors.New("invalid match result")
)

// ExecutionStage identifies where execution or semantic validation failed.
type ExecutionStage string

const (
	ExecutionStageTransaction ExecutionStage = "transaction"
	ExecutionStageReceipt     ExecutionStage = "receipt"
	ExecutionStageDecode      ExecutionStage = "decode"
	ExecutionStageMatch       ExecutionStage = "match"
	ExecutionStageValidation  ExecutionStage = "validation"
)

// ValidationStatus describes the semantic validation state of a step or expectation.
type ValidationStatus string

const (
	ValidationValidated   ValidationStatus = "validated"
	ValidationUnvalidated ValidationStatus = "unvalidated"
	ValidationFailed      ValidationStatus = "failed"
	ValidationSkipped     ValidationStatus = "skipped"
)

// SkipReason explains why validation was not attempted.
type SkipReason string

const (
	SkipExecutionFailed       SkipReason = "execution_failed"
	SkipInvalidReceipt        SkipReason = "invalid_receipt"
	SkipPriorValidationFailed SkipReason = "prior_validation_failed"
)

// ExpectationResult reports matching details for one expected event.
type ExpectationResult struct {
	Name           string
	Status         ValidationStatus
	SkipReason     SkipReason
	CandidateCount int
	Mismatches     []FieldMismatch
	Event          DecodedEvent
}

// StepResult reports semantic validation for one built step.
type StepResult struct {
	ID           StepID
	Name         string
	Status       ValidationStatus
	SkipReason   SkipReason
	Expectations []ExpectationResult
}

// ExecutionResult preserves the mined receipt and all available step-level
// semantic validation results, including partial results on failure.
type ExecutionResult struct {
	Receipt *types.Receipt
	Steps   []StepResult
}

// EventsOf returns every validated event assignable to T in execution order.
func EventsOf[T DecodedEvent](result *ExecutionResult) []T {
	if result == nil {
		return nil
	}
	events := make([]T, 0)
	for _, step := range result.Steps {
		for _, expectation := range step.Expectations {
			if event, ok := expectation.Event.(T); ok {
				events = append(events, event)
			}
		}
	}
	return events
}

// ExecutionError wraps an execution or validation failure with stage and
// partial-result location metadata.
type ExecutionError struct {
	Stage       ExecutionStage
	StepID      StepID
	Expectation string
	LogIndex    *uint
	Err         error
}

func (e *ExecutionError) Error() string {
	if e == nil {
		return "<nil>"
	}
	location := ""
	if e.StepID != "" {
		location += " step " + string(e.StepID)
	}
	if e.Expectation != "" {
		location += " expectation " + e.Expectation
	}
	if e.LogIndex != nil {
		location += fmt.Sprintf(" log %d", *e.LogIndex)
	}
	if e.Err == nil {
		return fmt.Sprintf("execution %s failed%s", e.Stage, location)
	}
	return fmt.Sprintf("execution %s failed%s: %v", e.Stage, location, e.Err)
}

// Unwrap preserves sentinel and typed errors for errors.Is and errors.As.
func (e *ExecutionError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func newExecutionResult(plan *ExecutionPlan, receipt *types.Receipt, reason SkipReason) *ExecutionResult {
	result := &ExecutionResult{Receipt: receipt}
	if plan == nil {
		return result
	}
	result.Steps = make([]StepResult, len(plan.Steps))
	for i, step := range plan.Steps {
		result.Steps[i] = StepResult{
			ID:         step.ID,
			Name:       step.Name,
			Status:     ValidationSkipped,
			SkipReason: reason,
		}
		result.Steps[i].Expectations = make([]ExpectationResult, len(step.Expectations))
		for j, expectation := range step.Expectations {
			name := "<nil>"
			if expectation != nil {
				name = expectation.ExpectationName()
				if name == "" {
					name = "<unnamed>"
				}
			}
			result.Steps[i].Expectations[j] = ExpectationResult{
				Name:       name,
				Status:     ValidationSkipped,
				SkipReason: reason,
			}
		}
	}
	return result
}
