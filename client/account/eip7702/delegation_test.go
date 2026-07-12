package eip7702_test

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/tn606024/defi-simplify/client/account/eip7702"
)

func TestDecodeDelegationCode(t *testing.T) {
	implementation := common.HexToAddress("0x1000000000000000000000000000000000000000")

	tests := []struct {
		name          string
		code          []byte
		wantStatus    eip7702.DelegationStatus
		wantDelegated bool
		wantClean     bool
		wantAddress   common.Address
	}{
		{
			name:          "empty code is clean EOA",
			code:          nil,
			wantStatus:    eip7702.DelegationStatusClean,
			wantDelegated: false,
			wantClean:     true,
		},
		{
			name:          "7702 indicator decodes implementation",
			code:          types.AddressToDelegation(implementation),
			wantStatus:    eip7702.DelegationStatusDelegated,
			wantDelegated: true,
			wantAddress:   implementation,
		},
		{
			name:          "other code is not a 7702 delegation",
			code:          []byte{0x60, 0x00, 0x60, 0x00},
			wantStatus:    eip7702.DelegationStatusContractCode,
			wantDelegated: false,
			wantClean:     false,
		},
		{
			name:          "short delegation prefix is treated as other code",
			code:          []byte{0xef, 0x01, 0x00},
			wantStatus:    eip7702.DelegationStatusContractCode,
			wantDelegated: false,
			wantClean:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := eip7702.DecodeDelegationCode(tt.code)

			if state.Status != tt.wantStatus {
				t.Fatalf("unexpected status: got %s want %s", state.Status, tt.wantStatus)
			}
			if state.Delegated() != tt.wantDelegated {
				t.Fatalf("unexpected delegated flag: got %t want %t", state.Delegated(), tt.wantDelegated)
			}
			if state.Clean() != tt.wantClean {
				t.Fatalf("unexpected clean flag: got %t want %t", state.Clean(), tt.wantClean)
			}
			if state.Implementation != tt.wantAddress {
				t.Fatalf("unexpected implementation: got %s want %s", state.Implementation.Hex(), tt.wantAddress.Hex())
			}
		})
	}
}

func TestReadAndAssertDelegationState(t *testing.T) {
	account := common.HexToAddress("0x1000000000000000000000000000000000000000")
	implementation := common.HexToAddress("0x2000000000000000000000000000000000000000")
	reader := &fakeCodeReader{
		code: map[common.Address][]byte{
			account: types.AddressToDelegation(implementation),
		},
	}

	state, err := eip7702.ReadDelegationState(context.Background(), reader, account)
	if err != nil {
		t.Fatalf("read delegation state: %v", err)
	}
	if state.Account != account {
		t.Fatalf("unexpected account: got %s want %s", state.Account.Hex(), account.Hex())
	}
	if state.Implementation != implementation {
		t.Fatalf("unexpected implementation: got %s want %s", state.Implementation.Hex(), implementation.Hex())
	}

	if err := eip7702.AssertDelegatedTo(context.Background(), reader, account, implementation); err != nil {
		t.Fatalf("expected delegated assertion to pass: %v", err)
	}
	if err := eip7702.AssertClean(context.Background(), reader, account); err == nil {
		t.Fatal("expected clean assertion to fail for delegated account")
	}
}

func TestAssertClean(t *testing.T) {
	account := common.HexToAddress("0x1000000000000000000000000000000000000000")
	reader := &fakeCodeReader{
		code: map[common.Address][]byte{
			account: nil,
		},
	}

	if err := eip7702.AssertClean(context.Background(), reader, account); err != nil {
		t.Fatalf("expected clean assertion to pass: %v", err)
	}
	if err := eip7702.AssertDelegatedTo(context.Background(), reader, account, common.Address{}); err == nil {
		t.Fatal("expected delegated assertion to fail for clean account")
	}
}

func TestReadDelegationStateWrapsCodeReadError(t *testing.T) {
	account := common.HexToAddress("0x1000000000000000000000000000000000000000")
	wantErr := errors.New("boom")
	reader := &fakeCodeReader{err: wantErr}

	_, err := eip7702.ReadDelegationState(context.Background(), reader, account)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected wrapped code read error, got %v", err)
	}
}

type fakeCodeReader struct {
	code map[common.Address][]byte
	err  error
}

func (r *fakeCodeReader) CodeAt(ctx context.Context, account common.Address, blockNumber *big.Int) ([]byte, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.code[account], nil
}
