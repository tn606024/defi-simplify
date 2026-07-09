package eip7702

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type DelegationStatus string

const (
	DelegationStatusClean        DelegationStatus = "clean"
	DelegationStatusDelegated    DelegationStatus = "delegated"
	DelegationStatusContractCode DelegationStatus = "contract_code"
)

type DelegationState struct {
	Account        common.Address
	Status         DelegationStatus
	Implementation common.Address
	Code           []byte
}

type CodeReader interface {
	CodeAt(ctx context.Context, account common.Address, blockNumber *big.Int) ([]byte, error)
}

func DecodeDelegationCode(code []byte) DelegationState {
	state := DelegationState{
		Code: common.CopyBytes(code),
	}
	if len(code) == 0 {
		state.Status = DelegationStatusClean
		return state
	}
	implementation, ok := types.ParseDelegation(code)
	if ok {
		state.Status = DelegationStatusDelegated
		state.Implementation = implementation
		return state
	}
	state.Status = DelegationStatusContractCode
	return state
}

func (s DelegationState) Clean() bool {
	return s.Status == DelegationStatusClean
}

func (s DelegationState) Delegated() bool {
	return s.Status == DelegationStatusDelegated
}

func ReadDelegationState(ctx context.Context, reader CodeReader, account common.Address) (DelegationState, error) {
	code, err := reader.CodeAt(ctx, account, nil)
	if err != nil {
		return DelegationState{}, fmt.Errorf("read code for %s: %w", account.Hex(), err)
	}
	state := DecodeDelegationCode(code)
	state.Account = account
	return state, nil
}

func AssertClean(ctx context.Context, reader CodeReader, account common.Address) error {
	state, err := ReadDelegationState(ctx, reader, account)
	if err != nil {
		return err
	}
	if !state.Clean() {
		return fmt.Errorf("expected %s to be clean, got %s delegated to %s", account.Hex(), state.Status, state.Implementation.Hex())
	}
	return nil
}

func AssertDelegatedTo(ctx context.Context, reader CodeReader, account common.Address, implementation common.Address) error {
	state, err := ReadDelegationState(ctx, reader, account)
	if err != nil {
		return err
	}
	if !state.Delegated() {
		return fmt.Errorf("expected %s to delegate to %s, got %s", account.Hex(), implementation.Hex(), state.Status)
	}
	if state.Implementation != implementation {
		return fmt.Errorf("expected %s to delegate to %s, got %s", account.Hex(), implementation.Hex(), state.Implementation.Hex())
	}
	return nil
}
