package amount

import (
	"errors"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	"github.com/tn606024/defi-simplify/config"
	"github.com/tn606024/defi-simplify/token"
)

func TestSourceValidation(t *testing.T) {
	ref := mustRef(t, config.Base, "0x0000000000000000000000000000000000000001")
	checkpoint := Checkpoint("swap-output", ref)
	tests := []struct {
		name    string
		source  Source
		wantErr error
	}{
		{name: "exact", source: Exact(decimal.RequireFromString("1.25"))},
		{name: "zero exact", source: Exact(decimal.Zero)},
		{name: "current balance", source: CurrentBalance(ref)},
		{name: "checkpoint delta", source: CheckpointDelta(checkpoint)},
		{name: "scaled current balance", source: Scale(CurrentBalance(ref), 5_000)},
		{name: "zero value", source: Source{}, wantErr: ErrInvalidSource},
		{name: "negative exact", source: Exact(decimal.NewFromInt(-1)), wantErr: ErrInvalidSource},
		{name: "scaled exact", source: Scale(Exact(decimal.NewFromInt(1)), 5_000), wantErr: ErrInvalidSource},
		{name: "zero bps", source: Scale(CurrentBalance(ref), 0), wantErr: ErrInvalidSource},
		{name: "excessive bps", source: Scale(CurrentBalance(ref), 10_001), wantErr: ErrInvalidSource},
		{
			name:    "empty checkpoint label",
			source:  CheckpointDelta(Checkpoint("", ref)),
			wantErr: ErrInvalidSource,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.source.Validate()
			if test.wantErr == nil && err != nil {
				t.Fatalf("Validate() error = %v", err)
			}
			if test.wantErr != nil && !errors.Is(err, test.wantErr) {
				t.Fatalf("Validate() error = %v, want errors.Is(%v)", err, test.wantErr)
			}
		})
	}
}

func TestSourceAccessors(t *testing.T) {
	ref := mustRef(t, config.Base, "0x0000000000000000000000000000000000000001")
	checkpoint := Checkpoint("borrow-output", ref)
	source := Scale(CheckpointDelta(checkpoint), 3_750)

	if source.Kind() != KindCheckpointDelta {
		t.Fatalf("Kind() = %d", source.Kind())
	}
	if source.BPS() != 3_750 {
		t.Fatalf("BPS() = %d", source.BPS())
	}
	if actual, ok := source.Token(); !ok || !actual.SameAsset(ref) {
		t.Fatalf("Token() = %v, %v", actual, ok)
	}
	if actual, ok := source.CheckpointRef(); !ok || actual != checkpoint {
		t.Fatalf("CheckpointRef() = %v, %v", actual, ok)
	}
}

func mustRef(t *testing.T, chain config.Chain, address string) token.Ref {
	t.Helper()
	ref, err := token.NewRef(chain, common.HexToAddress(address))
	if err != nil {
		t.Fatal(err)
	}
	return ref
}
