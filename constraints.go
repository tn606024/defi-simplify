package defi

import (
	"errors"
	"fmt"
	"math/big"
)

// ErrInvalidAmountConstraint is returned when a constraint is nil or was
// configured without a required bound.
var ErrInvalidAmountConstraint = errors.New("invalid amount constraint")

type amountConstraintKind uint8

const (
	exactAmountConstraint amountConstraintKind = iota
	positiveAmountConstraint
	atLeastAmountConstraint
	atMostAmountConstraint
)

type amountConstraint struct {
	kind     amountConstraintKind
	expected *big.Int
}

// Exact requires an amount to equal expected.
func Exact(expected *big.Int) AmountConstraint {
	return &amountConstraint{kind: exactAmountConstraint, expected: cloneBigInt(expected)}
}

// Positive requires an amount to be greater than zero.
func Positive() AmountConstraint {
	return &amountConstraint{kind: positiveAmountConstraint}
}

// AtLeast requires an amount to be greater than or equal to minimum.
func AtLeast(minimum *big.Int) AmountConstraint {
	return &amountConstraint{kind: atLeastAmountConstraint, expected: cloneBigInt(minimum)}
}

// AtMost requires an amount to be less than or equal to maximum.
func AtMost(maximum *big.Int) AmountConstraint {
	return &amountConstraint{kind: atMostAmountConstraint, expected: cloneBigInt(maximum)}
}

func (c *amountConstraint) Describe() string {
	if c == nil {
		return "<nil constraint>"
	}
	switch c.kind {
	case exactAmountConstraint:
		return formatBigInt(c.expected)
	case positiveAmountConstraint:
		return "> 0"
	case atLeastAmountConstraint:
		return ">= " + formatBigInt(c.expected)
	case atMostAmountConstraint:
		return "<= " + formatBigInt(c.expected)
	default:
		return fmt.Sprintf("unknown constraint %d", c.kind)
	}
}

func (c *amountConstraint) Match(actual *big.Int, _ MatchContext) (MatchResult, error) {
	if c == nil {
		return MatchResult{}, fmt.Errorf("%w: constraint is nil", ErrInvalidAmountConstraint)
	}
	if c.kind != positiveAmountConstraint && c.expected == nil {
		return MatchResult{}, fmt.Errorf("%w: %s bound is nil", ErrInvalidAmountConstraint, c.Describe())
	}
	if actual == nil {
		return constraintMismatch(c.Describe(), actual), nil
	}

	matched := false
	switch c.kind {
	case exactAmountConstraint:
		matched = actual.Cmp(c.expected) == 0
	case positiveAmountConstraint:
		matched = actual.Sign() > 0
	case atLeastAmountConstraint:
		matched = actual.Cmp(c.expected) >= 0
	case atMostAmountConstraint:
		matched = actual.Cmp(c.expected) <= 0
	default:
		return MatchResult{}, fmt.Errorf("%w: unknown constraint kind %d", ErrInvalidAmountConstraint, c.kind)
	}
	if matched {
		return MatchResult{Decision: MatchAccepted}, nil
	}
	return constraintMismatch(c.Describe(), actual), nil
}

// MatchAmountConstraints evaluates every amount constraint and aggregates all
// ordinary mismatches. A constraint error remains a hard failure and aborts
// evaluation immediately.
func MatchAmountConstraints(field string, actual *big.Int, ctx MatchContext, constraints ...AmountConstraint) (MatchResult, error) {
	result := MatchResult{Decision: MatchAccepted}
	for _, constraint := range constraints {
		if constraint == nil {
			return MatchResult{}, fmt.Errorf("%w: constraint is nil", ErrInvalidAmountConstraint)
		}
		matched, err := constraint.Match(actual, ctx)
		if err != nil {
			return MatchResult{}, err
		}
		if err := validateMatchResult(matched); err != nil {
			return MatchResult{}, err
		}
		if matched.Decision == MatchSkip {
			result.Decision = MatchSkip
			if len(matched.Mismatches) == 0 {
				matched.Mismatches = []FieldMismatch{{
					Expected: constraint.Describe(),
					Actual:   formatBigInt(actual),
				}}
			}
			for _, mismatch := range matched.Mismatches {
				if mismatch.Field == "" {
					mismatch.Field = field
				}
				result.Mismatches = append(result.Mismatches, mismatch)
			}
		}
	}
	return result, nil
}

func constraintMismatch(expected string, actual *big.Int) MatchResult {
	return MatchResult{
		Decision: MatchSkip,
		Mismatches: []FieldMismatch{{
			Expected: expected,
			Actual:   formatBigInt(actual),
		}},
	}
}

func formatBigInt(value *big.Int) string {
	if value == nil {
		return "<nil>"
	}
	return value.String()
}

func cloneBigInt(value *big.Int) *big.Int {
	if value == nil {
		return nil
	}
	return new(big.Int).Set(value)
}
