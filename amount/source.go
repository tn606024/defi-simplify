// Package amount defines protocol-neutral transaction amount intent.
package amount

import (
	"errors"
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
	"github.com/tn606024/defi-simplify/token"
)

const FullBPS uint16 = 10_000

var (
	ErrInvalidSource     = errors.New("invalid amount source")
	ErrInvalidCheckpoint = errors.New("invalid amount checkpoint")
)

// Kind identifies how an amount is resolved.
type Kind uint8

const (
	KindExact Kind = iota + 1
	KindCurrentBalance
	KindCheckpointDelta
)

// CheckpointRef names one token-balance snapshot within a flow.
type CheckpointRef struct {
	label string
	token token.Ref
}

// Checkpoint creates a named token-balance checkpoint reference.
func Checkpoint(label string, asset token.Ref) CheckpointRef {
	return CheckpointRef{label: label, token: asset}
}

// Validate checks the checkpoint label and token identity.
func (r CheckpointRef) Validate() error {
	if strings.TrimSpace(r.label) == "" {
		return fmt.Errorf("%w: label is empty", ErrInvalidCheckpoint)
	}
	if err := r.token.Validate(); err != nil {
		return fmt.Errorf("%w: token: %w", ErrInvalidCheckpoint, err)
	}
	return nil
}

// Label returns the human-readable checkpoint label used in diagnostics.
func (r CheckpointRef) Label() string {
	return r.label
}

// Token returns the checkpoint's chain-scoped token identity.
func (r CheckpointRef) Token() token.Ref {
	return r.token
}

// Source describes either an exact off-chain amount or an on-chain token
// balance source. The zero value is invalid.
type Source struct {
	kind       Kind
	exact      decimal.Decimal
	token      token.Ref
	checkpoint CheckpointRef
	bps        uint16
}

// Exact creates a statically resolvable human-unit amount.
func Exact(value decimal.Decimal) Source {
	return Source{kind: KindExact, exact: cloneDecimal(value)}
}

// CurrentBalance resolves to the delegated account's token balance immediately
// before the patched call.
func CurrentBalance(asset token.Ref) Source {
	return Source{kind: KindCurrentBalance, token: asset, bps: FullBPS}
}

// CheckpointDelta resolves to the current token balance minus an explicitly
// declared earlier checkpoint.
func CheckpointDelta(checkpoint CheckpointRef) Source {
	return Source{
		kind:       KindCheckpointDelta,
		token:      checkpoint.Token(),
		checkpoint: checkpoint,
		bps:        FullBPS,
	}
}

// Scale applies a basis-point ratio to a runtime balance source. Exact amounts
// cannot be scaled through this API; invalid combinations fail during Build.
func Scale(source Source, bps uint16) Source {
	scaled := source
	scaled.bps = bps
	return scaled
}

// Validate checks the source shape without applying protocol-specific amount
// rules such as whether zero is accepted.
func (s Source) Validate() error {
	switch s.kind {
	case KindExact:
		if s.exact.IsNegative() {
			return fmt.Errorf("%w: exact amount must not be negative", ErrInvalidSource)
		}
		if s.bps != 0 {
			return fmt.Errorf("%w: exact amount cannot be scaled", ErrInvalidSource)
		}
	case KindCurrentBalance:
		if err := s.token.Validate(); err != nil {
			return fmt.Errorf("%w: token: %w", ErrInvalidSource, err)
		}
		if s.bps == 0 || s.bps > FullBPS {
			return fmt.Errorf("%w: bps must be between 1 and %d", ErrInvalidSource, FullBPS)
		}
	case KindCheckpointDelta:
		if err := s.checkpoint.Validate(); err != nil {
			return fmt.Errorf("%w: %w", ErrInvalidSource, err)
		}
		if !s.token.SameAsset(s.checkpoint.Token()) {
			return fmt.Errorf("%w: checkpoint token does not match source token", ErrInvalidSource)
		}
		if s.bps == 0 || s.bps > FullBPS {
			return fmt.Errorf("%w: bps must be between 1 and %d", ErrInvalidSource, FullBPS)
		}
	default:
		return fmt.Errorf("%w: unsupported kind %d", ErrInvalidSource, s.kind)
	}
	return nil
}

// Kind returns the source resolution kind.
func (s Source) Kind() Kind {
	return s.kind
}

// ExactValue returns a defensive copy of the exact amount.
func (s Source) ExactValue() (decimal.Decimal, bool) {
	if s.kind != KindExact {
		return decimal.Decimal{}, false
	}
	return cloneDecimal(s.exact), true
}

// Token returns the runtime source token.
func (s Source) Token() (token.Ref, bool) {
	if s.kind != KindCurrentBalance && s.kind != KindCheckpointDelta {
		return token.Ref{}, false
	}
	return s.token, true
}

// CheckpointRef returns the checkpoint consumed by a delta source.
func (s Source) CheckpointRef() (CheckpointRef, bool) {
	if s.kind != KindCheckpointDelta {
		return CheckpointRef{}, false
	}
	return s.checkpoint, true
}

// BPS returns the runtime source ratio in basis points.
func (s Source) BPS() uint16 {
	return s.bps
}

func cloneDecimal(value decimal.Decimal) decimal.Decimal {
	return decimal.NewFromBigInt(value.Coefficient(), value.Exponent())
}
