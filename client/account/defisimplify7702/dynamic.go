package defisimplify7702

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/tn606024/defi-simplify/client/account/defisimplify7702/bindings"
)

var (
	ErrEmptyDynamicBatch       = errors.New("dynamic batch is empty")
	ErrIncompatibleDynamicCall = errors.New("dynamic call is incompatible with account ABI")
)

// EncodeExecuteBatchDynamic validates and ABI-encodes compiler-produced calls
// for DefiSimplify7702Account.executeBatchDynamic. It validates the account ABI
// shape; Flow.Build owns checkpoint ordering, source-token consistency, and
// zero-placeholder validation.
func EncodeExecuteBatchDynamic(calls []DynamicCall) ([]byte, error) {
	bindingCalls, err := toBindingDynamicCalls(calls)
	if err != nil {
		return nil, err
	}
	parsed, err := bindings.DefiSimplify7702AccountMetaData.GetAbi()
	if err != nil {
		return nil, fmt.Errorf("load DefiSimplify7702Account ABI: %w", err)
	}
	data, err := parsed.Pack("executeBatchDynamic", bindingCalls)
	if err != nil {
		return nil, fmt.Errorf("encode executeBatchDynamic: %w", err)
	}
	return data, nil
}

func toBindingDynamicCalls(
	calls []DynamicCall,
) ([]bindings.IDefiSimplify7702AccountDynamicCall, error) {
	if len(calls) == 0 {
		return nil, ErrEmptyDynamicBatch
	}
	translated := make([]bindings.IDefiSimplify7702AccountDynamicCall, len(calls))
	for callIndex, call := range calls {
		if call.Target == (common.Address{}) {
			return nil, incompatibleCall(callIndex, "target is zero")
		}
		translated[callIndex] = bindings.IDefiSimplify7702AccountDynamicCall{
			Target:          call.Target,
			Value:           cloneBigIntOrZero(call.Value),
			Data:            append([]byte(nil), call.Data...),
			ExpectsCallback: call.ExpectsCallback,
		}
		var previousOffset uint32
		for patchIndex, patch := range call.Patches {
			if err := validateBindingPatch(callIndex, patchIndex, patch, call.Data, previousOffset); err != nil {
				return nil, err
			}
			previousOffset = patch.Offset
			translated[callIndex].Patches = append(
				translated[callIndex].Patches,
				bindings.IDefiSimplify7702AccountBalancePatch{
					Token:        patch.Token,
					CheckpointId: patch.CheckpointID,
					Offset:       patch.Offset,
					Bps:          patch.BPS,
					Source:       uint8(patch.Source),
				},
			)
		}
		for checkpointIndex, checkpoint := range call.CheckpointsBefore {
			if checkpoint.Token == (common.Address{}) {
				return nil, incompatibleCall(callIndex, "checkpoint %d token is zero", checkpointIndex)
			}
			if checkpoint.ID == ([32]byte{}) {
				return nil, incompatibleCall(callIndex, "checkpoint %d ID is zero", checkpointIndex)
			}
			translated[callIndex].CheckpointsBefore = append(
				translated[callIndex].CheckpointsBefore,
				bindings.IDefiSimplify7702AccountBalanceCheckpoint{
					Token: checkpoint.Token,
					Id:    checkpoint.ID,
				},
			)
		}
	}
	return translated, nil
}

func validateBindingPatch(
	callIndex int,
	patchIndex int,
	patch BalancePatch,
	data []byte,
	previousOffset uint32,
) error {
	if patch.Token == (common.Address{}) {
		return incompatibleCall(callIndex, "patch %d token is zero", patchIndex)
	}
	if patch.Offset < 4 || (patch.Offset-4)%32 != 0 || uint64(patch.Offset)+32 > uint64(len(data)) {
		return incompatibleCall(
			callIndex,
			"patch %d offset %d does not identify a complete ABI word in %d bytes",
			patchIndex,
			patch.Offset,
			len(data),
		)
	}
	if patchIndex > 0 && patch.Offset <= previousOffset {
		return incompatibleCall(
			callIndex,
			"patch %d offset %d is not greater than %d",
			patchIndex,
			patch.Offset,
			previousOffset,
		)
	}
	if patch.BPS == 0 || patch.BPS > 10_000 {
		return incompatibleCall(callIndex, "patch %d BPS %d is invalid", patchIndex, patch.BPS)
	}
	switch patch.Source {
	case BalanceSourceCurrentBalance:
		if patch.CheckpointID != ([32]byte{}) {
			return incompatibleCall(callIndex, "patch %d current-balance checkpoint ID is not zero", patchIndex)
		}
	case BalanceSourceCheckpointDelta:
		if patch.CheckpointID == ([32]byte{}) {
			return incompatibleCall(callIndex, "patch %d checkpoint-delta ID is zero", patchIndex)
		}
	default:
		return incompatibleCall(callIndex, "patch %d balance source %d is unsupported", patchIndex, patch.Source)
	}
	return nil
}

func incompatibleCall(callIndex int, format string, args ...interface{}) error {
	return fmt.Errorf(
		"%w: call %d: %s",
		ErrIncompatibleDynamicCall,
		callIndex,
		fmt.Sprintf(format, args...),
	)
}

func cloneBigIntOrZero(value *big.Int) *big.Int {
	if value == nil {
		return new(big.Int)
	}
	return new(big.Int).Set(value)
}
