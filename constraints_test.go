package defi

import (
	"errors"
	"math/big"
	"testing"
)

func TestAmountConstraints(t *testing.T) {
	tests := []struct {
		name       string
		constraint AmountConstraint
		actual     *big.Int
		decision   MatchDecision
		wantErr    error
	}{
		{name: "exact match", constraint: Exact(big.NewInt(10)), actual: big.NewInt(10), decision: MatchAccepted},
		{name: "exact mismatch", constraint: Exact(big.NewInt(10)), actual: big.NewInt(9), decision: MatchSkip},
		{name: "positive match", constraint: Positive(), actual: big.NewInt(1), decision: MatchAccepted},
		{name: "positive zero", constraint: Positive(), actual: big.NewInt(0), decision: MatchSkip},
		{name: "at least boundary", constraint: AtLeast(big.NewInt(10)), actual: big.NewInt(10), decision: MatchAccepted},
		{name: "at least mismatch", constraint: AtLeast(big.NewInt(10)), actual: big.NewInt(9), decision: MatchSkip},
		{name: "at most boundary", constraint: AtMost(big.NewInt(10)), actual: big.NewInt(10), decision: MatchAccepted},
		{name: "at most mismatch", constraint: AtMost(big.NewInt(10)), actual: big.NewInt(11), decision: MatchSkip},
		{name: "nil actual", constraint: Exact(big.NewInt(10)), actual: nil, decision: MatchSkip},
		{name: "nil bound", constraint: Exact(nil), actual: big.NewInt(10), decision: MatchSkip, wantErr: ErrInvalidAmountConstraint},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.constraint.Match(test.actual, MatchContext{})
			if test.wantErr != nil {
				if !errors.Is(err, test.wantErr) {
					t.Fatalf("expected error %v, got %v", test.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Decision != test.decision {
				t.Fatalf("expected decision %d, got %d", test.decision, result.Decision)
			}
			if test.decision == MatchSkip && len(result.Mismatches) == 0 {
				t.Fatal("expected mismatch details")
			}
		})
	}
}

func TestMatchAmountConstraintsAggregatesMismatches(t *testing.T) {
	result, err := MatchAmountConstraints(
		"amount",
		big.NewInt(-1),
		MatchContext{},
		Positive(),
		AtLeast(big.NewInt(10)),
		AtMost(big.NewInt(100)),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Decision != MatchSkip {
		t.Fatalf("expected MatchSkip, got %d", result.Decision)
	}
	if len(result.Mismatches) != 2 {
		t.Fatalf("expected 2 mismatches, got %d: %#v", len(result.Mismatches), result.Mismatches)
	}
	for _, mismatch := range result.Mismatches {
		if mismatch.Field != "amount" {
			t.Fatalf("expected amount field, got %q", mismatch.Field)
		}
	}
}

func TestAmountConstraintClonesBound(t *testing.T) {
	expected := big.NewInt(10)
	constraint := Exact(expected)
	expected.SetInt64(20)

	result, err := constraint.Match(big.NewInt(10), MatchContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Decision != MatchAccepted {
		t.Fatalf("expected cloned bound to remain 10, got %#v", result)
	}
}
