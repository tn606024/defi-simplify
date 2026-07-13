package defi

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/ethereum/go-ethereum/core/types"
)

// ValidateExecution validates a mined receipt against plan expectations.
//
// Expectations are processed in step and declaration order. Each expectation
// scans forward from the last accepted log. Accepted logs and all earlier logs
// are unavailable to later expectations, which enforces on-chain emission order
// and consume-once semantics structurally.
func ValidateExecution(plan *ExecutionPlan, receipt *types.Receipt) (*ExecutionResult, error) {
	if plan == nil || len(plan.Steps) == 0 {
		return nil, &ExecutionError{
			Stage: ExecutionStageValidation,
			Err:   fmt.Errorf("%w: plan is nil or empty", ErrInvalidExecutionPlan),
		}
	}
	if receipt == nil {
		return nil, &ExecutionError{
			Stage: ExecutionStageReceipt,
			Err:   fmt.Errorf("%w: receipt is nil", ErrInvalidExecutionReceipt),
		}
	}

	result := newExecutionResult(plan, receipt, SkipPriorValidationFailed)
	if err := validateExecutionReceipt(receipt); err != nil {
		setAllSkipReasons(result, SkipInvalidReceipt)
		return result, &ExecutionError{Stage: ExecutionStageReceipt, Err: err}
	}

	cursor := 0
	for stepIndex, step := range plan.Steps {
		stepResult := &result.Steps[stepIndex]
		if len(step.Expectations) == 0 {
			stepResult.Status = ValidationUnvalidated
			stepResult.SkipReason = ""
			continue
		}

		for expectationIndex, expectation := range step.Expectations {
			expectationResult := &stepResult.Expectations[expectationIndex]
			if expectation == nil || expectation.ExpectationName() == "" {
				return failValidation(
					result,
					stepIndex,
					expectationIndex,
					ExecutionStageValidation,
					nil,
					fmt.Errorf("%w: step %s expectation %d", ErrInvalidEventExpectation, step.ID, expectationIndex+1),
				)
			}

			matched := false
			for logPosition := cursor; logPosition < len(receipt.Logs); logPosition++ {
				receiptLog := receipt.Logs[logPosition]
				if receiptLog == nil || !expectation.IsCandidate(receiptLog) {
					continue
				}
				expectationResult.CandidateCount++

				event, err := expectation.Decode(receiptLog)
				if err != nil || isNilDecodedEvent(event) {
					if err == nil {
						err = fmt.Errorf("decoder returned a nil event")
					}
					decodeErr := fmt.Errorf(
						"%w: decode %s log %d: %v",
						ErrMalformedExecutionEvent,
						expectation.ExpectationName(),
						receiptLog.Index,
						err,
					)
					return failValidation(result, stepIndex, expectationIndex, ExecutionStageDecode, &receiptLog.Index, decodeErr)
				}
				if err := validateDecodedEvent(event, receiptLog); err != nil {
					decodeErr := fmt.Errorf("%w: %v", ErrMalformedExecutionEvent, err)
					return failValidation(result, stepIndex, expectationIndex, ExecutionStageDecode, &receiptLog.Index, decodeErr)
				}

				match, err := expectation.Match(event, MatchContext{StepID: step.ID})
				if err != nil {
					return failValidation(result, stepIndex, expectationIndex, ExecutionStageMatch, &receiptLog.Index, err)
				}
				if err := validateMatchResult(match); err != nil {
					return failValidation(result, stepIndex, expectationIndex, ExecutionStageMatch, &receiptLog.Index, err)
				}
				if match.Decision == MatchSkip {
					if len(match.Mismatches) == 0 {
						match.Mismatches = []FieldMismatch{{
							Field:    "expectation",
							Expected: expectation.ExpectationName(),
							Actual:   "candidate rejected without mismatch details",
						}}
					}
					expectationResult.Mismatches = append(expectationResult.Mismatches, match.Mismatches...)
					continue
				}

				expectationResult.Status = ValidationValidated
				expectationResult.SkipReason = ""
				expectationResult.Event = event
				cursor = logPosition + 1
				matched = true
				break
			}

			if !matched {
				err := expectedEventNotFoundError(expectation, expectationResult)
				return failValidation(result, stepIndex, expectationIndex, ExecutionStageValidation, nil, err)
			}
		}
		stepResult.Status = ValidationValidated
		stepResult.SkipReason = ""
	}
	return result, nil
}

func validateExecutionReceipt(receipt *types.Receipt) error {
	if receipt.Status != types.ReceiptStatusSuccessful {
		return fmt.Errorf("%w: transaction %s has status %d", ErrInvalidExecutionReceipt, receipt.TxHash.Hex(), receipt.Status)
	}
	if receipt.BlockNumber == nil {
		return fmt.Errorf("%w: transaction %s has no block number", ErrInvalidExecutionReceipt, receipt.TxHash.Hex())
	}
	return nil
}

func validateDecodedEvent(event DecodedEvent, receiptLog *types.Log) error {
	metadata := event.EventMetadata()
	if metadata.Protocol == "" || metadata.Name == "" {
		return fmt.Errorf("decoded event metadata is incomplete")
	}
	if metadata.Emitter != receiptLog.Address {
		return fmt.Errorf("decoded event emitter %s does not match log emitter %s", metadata.Emitter.Hex(), receiptLog.Address.Hex())
	}
	if metadata.LogIndex != receiptLog.Index {
		return fmt.Errorf("decoded event log index %d does not match receipt log index %d", metadata.LogIndex, receiptLog.Index)
	}
	return nil
}

func validateMatchResult(result MatchResult) error {
	switch result.Decision {
	case MatchSkip:
		return nil
	case MatchAccepted:
		if len(result.Mismatches) != 0 {
			return fmt.Errorf("%w: accepted result contains mismatches", ErrInvalidMatchResult)
		}
		return nil
	default:
		return fmt.Errorf("%w: unknown decision %d", ErrInvalidMatchResult, result.Decision)
	}
}

func failValidation(
	result *ExecutionResult,
	stepIndex int,
	expectationIndex int,
	stage ExecutionStage,
	logIndex *uint,
	err error,
) (*ExecutionResult, error) {
	step := &result.Steps[stepIndex]
	step.Status = ValidationFailed
	step.SkipReason = ""
	expectation := &step.Expectations[expectationIndex]
	expectation.Status = ValidationFailed
	expectation.SkipReason = ""

	return result, &ExecutionError{
		Stage:       stage,
		StepID:      step.ID,
		Expectation: expectation.Name,
		LogIndex:    cloneUint(logIndex),
		Err:         err,
	}
}

func expectedEventNotFoundError(expectation EventExpectation, result *ExpectationResult) error {
	detail := ""
	if len(result.Mismatches) != 0 {
		parts := make([]string, 0, len(result.Mismatches))
		for _, mismatch := range result.Mismatches {
			parts = append(parts, fmt.Sprintf("%s expected %s, got %s", mismatch.Field, mismatch.Expected, mismatch.Actual))
		}
		detail = ": " + strings.Join(parts, "; ")
	}
	return fmt.Errorf(
		"%w: %s matched 0 of %d candidate logs%s",
		ErrExpectedEventNotFound,
		expectation.ExpectationName(),
		result.CandidateCount,
		detail,
	)
}

func setAllSkipReasons(result *ExecutionResult, reason SkipReason) {
	for i := range result.Steps {
		result.Steps[i].SkipReason = reason
		for j := range result.Steps[i].Expectations {
			result.Steps[i].Expectations[j].SkipReason = reason
		}
	}
}

func isNilDecodedEvent(event DecodedEvent) bool {
	if event == nil {
		return true
	}
	value := reflect.ValueOf(event)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

func cloneUint(value *uint) *uint {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
