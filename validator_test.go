package defi

import (
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type fakeDecodedEvent struct {
	metadata EventMetadata
	value    string
}

func (e *fakeDecodedEvent) EventMetadata() EventMetadata {
	return e.metadata
}

type fakeEventExpectation struct {
	name      string
	emitter   common.Address
	topic     common.Hash
	decodeErr error
	matchErr  error
	expected  string
}

func (e *fakeEventExpectation) ExpectationName() string {
	return e.name
}

func (e *fakeEventExpectation) IsCandidate(log *types.Log) bool {
	return log != nil && log.Address == e.emitter && len(log.Topics) != 0 && log.Topics[0] == e.topic
}

func (e *fakeEventExpectation) Decode(log *types.Log) (DecodedEvent, error) {
	if e.decodeErr != nil {
		return nil, e.decodeErr
	}
	return &fakeDecodedEvent{
		metadata: EventMetadata{Protocol: "fake", Name: e.name, Emitter: log.Address, LogIndex: log.Index},
		value:    string(log.Data),
	}, nil
}

func (e *fakeEventExpectation) Match(event DecodedEvent, _ MatchContext) (MatchResult, error) {
	if e.matchErr != nil {
		return MatchResult{}, e.matchErr
	}
	decoded := event.(*fakeDecodedEvent)
	if decoded.value != e.expected {
		return MatchResult{
			Decision: MatchSkip,
			Mismatches: []FieldMismatch{{
				Field: "value", Expected: e.expected, Actual: decoded.value,
			}},
		}, nil
	}
	return MatchResult{Decision: MatchAccepted}, nil
}

var _ = Describe("Execution validation", func() {
	var (
		emitter common.Address
		topic   common.Hash
		receipt *types.Receipt
	)

	BeforeEach(func() {
		emitter = common.HexToAddress("0x00000000000000000000000000000000000000aa")
		topic = common.HexToHash("0x1234")
		receipt = &types.Receipt{
			Status:      types.ReceiptStatusSuccessful,
			TxHash:      common.HexToHash("0xabcd"),
			BlockNumber: big.NewInt(42),
		}
	})

	It("skips mismatched candidates and consumes accepted logs in plan order", func() {
		first := &fakeEventExpectation{name: "First", emitter: emitter, topic: topic, expected: "first"}
		second := &fakeEventExpectation{name: "Second", emitter: emitter, topic: topic, expected: "second"}
		plan := &ExecutionPlan{Steps: []BuiltStep{
			{ID: "step#1", Name: "step", Expectations: []EventExpectation{first}},
			{ID: "step#2", Name: "step", Expectations: []EventExpectation{second}},
		}}
		receipt.Logs = []*types.Log{
			{Address: emitter, Topics: []common.Hash{topic}, Data: []byte("wrong"), Index: 1},
			{Address: emitter, Topics: []common.Hash{topic}, Data: []byte("first"), Index: 2},
			{Address: emitter, Topics: []common.Hash{topic}, Data: []byte("second"), Index: 3},
		}

		result, err := ValidateExecution(plan, receipt)

		Expect(err).NotTo(HaveOccurred())
		Expect(result.Receipt).To(Equal(receipt))
		Expect(result.Steps[0].Status).To(Equal(ValidationValidated))
		Expect(result.Steps[0].Expectations[0].CandidateCount).To(Equal(2))
		Expect(result.Steps[0].Expectations[0].Mismatches).To(HaveLen(1))
		Expect(result.Steps[1].Status).To(Equal(ValidationValidated))
		Expect(result.Steps[1].Expectations[0].CandidateCount).To(Equal(1))
		Expect(EventsOf[*fakeDecodedEvent](result)).To(HaveLen(2))
	})

	It("fails when expectation declaration order disagrees with emission order", func() {
		plan := &ExecutionPlan{Steps: []BuiltStep{{
			ID:   "step#1",
			Name: "step",
			Expectations: []EventExpectation{
				&fakeEventExpectation{name: "Second", emitter: emitter, topic: topic, expected: "second"},
				&fakeEventExpectation{name: "First", emitter: emitter, topic: topic, expected: "first"},
			},
		}}}
		receipt.Logs = []*types.Log{
			{Address: emitter, Topics: []common.Hash{topic}, Data: []byte("first"), Index: 1},
			{Address: emitter, Topics: []common.Hash{topic}, Data: []byte("second"), Index: 2},
		}

		result, err := ValidateExecution(plan, receipt)

		Expect(errors.Is(err, ErrExpectedEventNotFound)).To(BeTrue())
		Expect(result.Steps[0].Status).To(Equal(ValidationFailed))
		Expect(result.Steps[0].Expectations[0].Status).To(Equal(ValidationValidated))
		Expect(result.Steps[0].Expectations[1].Status).To(Equal(ValidationFailed))
	})

	It("returns a partial result and hard error for malformed candidate logs", func() {
		plan := &ExecutionPlan{Steps: []BuiltStep{
			{ID: "decode#1", Name: "decode", Expectations: []EventExpectation{
				&fakeEventExpectation{name: "Malformed", emitter: emitter, topic: topic, decodeErr: errors.New("bad ABI")},
			}},
			{ID: "later#1", Name: "later", Expectations: []EventExpectation{
				&fakeEventExpectation{name: "Later", emitter: emitter, topic: topic, expected: "later"},
			}},
		}}
		receipt.Logs = []*types.Log{{Address: emitter, Topics: []common.Hash{topic}, Index: 7}}

		result, err := ValidateExecution(plan, receipt)

		Expect(result).NotTo(BeNil())
		Expect(result.Receipt).To(Equal(receipt))
		Expect(errors.Is(err, ErrMalformedExecutionEvent)).To(BeTrue())
		Expect(result.Steps[0].Status).To(Equal(ValidationFailed))
		Expect(result.Steps[1].Status).To(Equal(ValidationSkipped))
		Expect(result.Steps[1].SkipReason).To(Equal(SkipPriorValidationFailed))
		var executionErr *ExecutionError
		Expect(errors.As(err, &executionErr)).To(BeTrue())
		Expect(executionErr.Stage).To(Equal(ExecutionStageDecode))
		Expect(*executionErr.LogIndex).To(Equal(uint(7)))
	})

	It("unwraps hard Match errors without treating them as mismatches", func() {
		boom := errors.New("matcher invariant failed")
		plan := &ExecutionPlan{Steps: []BuiltStep{{
			ID: "match#1", Name: "match", Expectations: []EventExpectation{
				&fakeEventExpectation{name: "Match", emitter: emitter, topic: topic, matchErr: boom},
			},
		}}}
		receipt.Logs = []*types.Log{{Address: emitter, Topics: []common.Hash{topic}, Index: 1}}

		result, err := ValidateExecution(plan, receipt)

		Expect(result).NotTo(BeNil())
		Expect(errors.Is(err, boom)).To(BeTrue())
		var executionErr *ExecutionError
		Expect(errors.As(err, &executionErr)).To(BeTrue())
		Expect(executionErr.Stage).To(Equal(ExecutionStageMatch))
	})

	It("marks steps without expectations as unvalidated", func() {
		plan := &ExecutionPlan{Steps: []BuiltStep{{ID: "escape#1", Name: "escape"}}}

		result, err := ValidateExecution(plan, receipt)

		Expect(err).NotTo(HaveOccurred())
		Expect(result.Steps[0].Status).To(Equal(ValidationUnvalidated))
		Expect(result.Steps[0].SkipReason).To(BeEmpty())
	})

	It("preserves a failed mined receipt and marks all steps skipped", func() {
		plan := &ExecutionPlan{Steps: []BuiltStep{{
			ID: "step#1", Name: "step", Expectations: []EventExpectation{
				&fakeEventExpectation{name: "Expected", emitter: emitter, topic: topic, expected: "value"},
			},
		}}}
		receipt.Status = types.ReceiptStatusFailed

		result, err := ValidateExecution(plan, receipt)

		Expect(result.Receipt).To(Equal(receipt))
		Expect(errors.Is(err, ErrInvalidExecutionReceipt)).To(BeTrue())
		Expect(result.Steps[0].Status).To(Equal(ValidationSkipped))
		Expect(result.Steps[0].SkipReason).To(Equal(SkipInvalidReceipt))
	})
})
