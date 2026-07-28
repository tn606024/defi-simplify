package defi

import (
	"errors"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/tn606024/defi-simplify/amount"
	"github.com/tn606024/defi-simplify/config"
	"github.com/tn606024/defi-simplify/token"
)

func TestCompileBalancePatchOffsetValidation(t *testing.T) {
	ref, err := token.NewRef(
		config.Base,
		common.HexToAddress("0x0000000000000000000000000000000000000001"),
	)
	if err != nil {
		t.Fatal(err)
	}
	source := amount.CurrentBalance(ref)
	tests := []struct {
		name   string
		offset uint32
		length int
		wantOK bool
	}{
		{name: "first ABI word", offset: 4, length: 36, wantOK: true},
		{name: "second ABI word", offset: 36, length: 68, wantOK: true},
		{name: "selector overlap", offset: 3, length: 68},
		{name: "unaligned", offset: 5, length: 68},
		{name: "incomplete word", offset: 36, length: 67},
		{name: "outside calldata", offset: 68, length: 68},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := compileBalancePatch(
				8453,
				CalldataPatch{Source: source, Offset: test.offset},
				make([]byte, test.length),
				map[amount.CheckpointRef]BalanceCheckpoint{},
			)
			if test.wantOK && err != nil {
				t.Fatalf("compileBalancePatch() error = %v", err)
			}
			if !test.wantOK && !errors.Is(err, ErrInvalidCalldataPatch) {
				t.Fatalf("compileBalancePatch() error = %v, want ErrInvalidCalldataPatch", err)
			}
		})
	}
}

func TestDeriveCheckpointID(t *testing.T) {
	ref, err := token.NewRef(
		config.Base,
		common.HexToAddress("0x0000000000000000000000000000000000000001"),
	)
	if err != nil {
		t.Fatal(err)
	}
	account := common.HexToAddress("0x00000000000000000000000000000000000000aa")
	baseRef := amount.Checkpoint("swap-output", ref)
	base, err := deriveCheckpointID(8453, account, "swap#1", 0, 0, baseRef)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name            string
		account         common.Address
		stepID          StepID
		callIndex       int
		checkpointIndex int
		ref             amount.CheckpointRef
		wantSame        bool
	}{
		{
			name:     "same inputs",
			account:  account,
			stepID:   "swap#1",
			ref:      baseRef,
			wantSame: true,
		},
		{
			name:    "different account",
			account: common.HexToAddress("0x00000000000000000000000000000000000000bb"),
			stepID:  "swap#1",
			ref:     baseRef,
		},
		{
			name:    "different step",
			account: account,
			stepID:  "swap#2",
			ref:     baseRef,
		},
		{
			name:      "different call",
			account:   account,
			stepID:    "swap#1",
			callIndex: 1,
			ref:       baseRef,
		},
		{
			name:            "different checkpoint position",
			account:         account,
			stepID:          "swap#1",
			checkpointIndex: 1,
			ref:             baseRef,
		},
		{
			name:    "different label",
			account: account,
			stepID:  "swap#1",
			ref:     amount.Checkpoint("other-output", ref),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := deriveCheckpointID(
				8453,
				test.account,
				test.stepID,
				test.callIndex,
				test.checkpointIndex,
				test.ref,
			)
			if err != nil {
				t.Fatal(err)
			}
			if (actual == base) != test.wantSame {
				t.Fatalf("deriveCheckpointID() = %x, base = %x, wantSame = %t", actual, base, test.wantSame)
			}
		})
	}
}
