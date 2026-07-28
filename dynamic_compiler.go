package defi

import (
	"errors"
	"fmt"
	"math/big"
	"sort"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tn606024/defi-simplify/amount"
	"github.com/tn606024/defi-simplify/config"
)

var (
	ErrInvalidDynamicCall        = errors.New("invalid dynamic call")
	ErrInvalidCheckpointGraph    = errors.New("invalid checkpoint graph")
	ErrInvalidCalldataPatch      = errors.New("invalid calldata patch")
	ErrDynamicPlaceholderNotZero = errors.New("dynamic calldata placeholder is not zero")
)

func compileExecutionPlan(
	account common.Address,
	chain config.Chain,
	steps []BuiltStep,
) (*ExecutionPlan, error) {
	plan := &ExecutionPlan{
		account: account,
		steps:   cloneBuiltSteps(steps),
	}
	if !hasDynamicMetadata(plan.steps) {
		plan.kind = PlanStatic
		for _, step := range plan.steps {
			for _, planned := range step.Calls {
				plan.staticCalls = append(plan.staticCalls, cloneCall(planned.Call))
			}
		}
		return plan, nil
	}

	chainID, err := chain.ChainID()
	if err != nil {
		return nil, fmt.Errorf("compile dynamic plan chain: %w", err)
	}
	plan.kind = PlanDynamic
	declared := make(map[amount.CheckpointRef]BalanceCheckpoint)
	for _, step := range plan.steps {
		for callIndex, planned := range step.Calls {
			dynamic, err := compileDynamicCall(
				chainID,
				account,
				step.ID,
				callIndex,
				planned,
				declared,
			)
			if err != nil {
				return nil, fmt.Errorf("compile %s call %d: %w", step.ID, callIndex, err)
			}
			plan.dynamicCalls = append(plan.dynamicCalls, dynamic)
		}
	}
	return plan, nil
}

func compileDynamicCall(
	chainID int,
	account common.Address,
	stepID StepID,
	callIndex int,
	planned PlannedCall,
	declared map[amount.CheckpointRef]BalanceCheckpoint,
) (DynamicCall, error) {
	call := cloneCall(planned.Call)
	if call.Target == (common.Address{}) {
		return DynamicCall{}, fmt.Errorf("%w: target is zero", ErrInvalidDynamicCall)
	}
	dynamic := DynamicCall{
		Target:          call.Target,
		Value:           cloneBigIntOrZero(call.Value),
		Data:            append([]byte(nil), call.Data...),
		ExpectsCallback: planned.ExpectsCallback,
	}

	patches := append([]CalldataPatch(nil), planned.Patches...)
	sort.Slice(patches, func(i, j int) bool { return patches[i].Offset < patches[j].Offset })
	for patchIndex, patch := range patches {
		materialized, err := compileBalancePatch(chainID, patch, dynamic.Data, declared)
		if err != nil {
			return DynamicCall{}, fmt.Errorf("patch %d: %w", patchIndex, err)
		}
		if patchIndex > 0 && materialized.Offset == dynamic.Patches[patchIndex-1].Offset {
			return DynamicCall{}, fmt.Errorf(
				"%w: duplicate offset %d",
				ErrInvalidCalldataPatch,
				materialized.Offset,
			)
		}
		dynamic.Patches = append(dynamic.Patches, materialized)
	}

	for checkpointIndex, declaration := range planned.CheckpointsBefore {
		if err := declaration.Ref.Validate(); err != nil {
			return DynamicCall{}, fmt.Errorf("%w: %w", ErrInvalidCheckpointGraph, err)
		}
		refChainID, err := declaration.Ref.Token().Chain().ChainID()
		if err != nil || refChainID != chainID {
			return DynamicCall{}, fmt.Errorf(
				"%w: checkpoint %q token chain does not match flow chain ID %d",
				ErrInvalidCheckpointGraph,
				declaration.Ref.Label(),
				chainID,
			)
		}
		if _, exists := declared[declaration.Ref]; exists {
			return DynamicCall{}, fmt.Errorf(
				"%w: checkpoint %q for %s is declared more than once",
				ErrInvalidCheckpointGraph,
				declaration.Ref.Label(),
				declaration.Ref.Token().Address().Hex(),
			)
		}
		id, err := deriveCheckpointID(
			chainID,
			account,
			stepID,
			callIndex,
			checkpointIndex,
			declaration.Ref,
		)
		if err != nil {
			return DynamicCall{}, err
		}
		checkpoint := BalanceCheckpoint{
			Token: declaration.Ref.Token().Address(),
			ID:    id,
		}
		declared[declaration.Ref] = checkpoint
		dynamic.CheckpointsBefore = append(dynamic.CheckpointsBefore, checkpoint)
	}
	return dynamic, nil
}

func compileBalancePatch(
	chainID int,
	patch CalldataPatch,
	data []byte,
	declared map[amount.CheckpointRef]BalanceCheckpoint,
) (BalancePatch, error) {
	if err := patch.Source.Validate(); err != nil {
		return BalancePatch{}, fmt.Errorf("%w: %w", ErrInvalidCalldataPatch, err)
	}
	if patch.Offset < 4 || (patch.Offset-4)%32 != 0 || uint64(patch.Offset)+32 > uint64(len(data)) {
		return BalancePatch{}, fmt.Errorf(
			"%w: offset %d does not identify a complete ABI word in %d bytes",
			ErrInvalidCalldataPatch,
			patch.Offset,
			len(data),
		)
	}
	for _, value := range data[patch.Offset : patch.Offset+32] {
		if value != 0 {
			return BalancePatch{}, fmt.Errorf(
				"%w: offset %d: %w",
				ErrInvalidCalldataPatch,
				patch.Offset,
				ErrDynamicPlaceholderNotZero,
			)
		}
	}
	sourceToken, ok := patch.Source.Token()
	if !ok {
		return BalancePatch{}, fmt.Errorf("%w: exact amounts cannot create patches", ErrInvalidCalldataPatch)
	}
	sourceChainID, err := sourceToken.Chain().ChainID()
	if err != nil || sourceChainID != chainID {
		return BalancePatch{}, fmt.Errorf(
			"%w: source token chain does not match flow chain ID %d",
			ErrInvalidCalldataPatch,
			chainID,
		)
	}
	materialized := BalancePatch{
		Token:  sourceToken.Address(),
		Offset: patch.Offset,
		BPS:    patch.Source.BPS(),
	}
	switch patch.Source.Kind() {
	case amount.KindCurrentBalance:
		materialized.Source = BalanceSourceCurrentBalance
	case amount.KindCheckpointDelta:
		ref, _ := patch.Source.CheckpointRef()
		checkpoint, exists := declared[ref]
		if !exists {
			return BalancePatch{}, fmt.Errorf(
				"%w: checkpoint %q for %s must be declared by an earlier call",
				ErrInvalidCheckpointGraph,
				ref.Label(),
				ref.Token().Address().Hex(),
			)
		}
		materialized.Source = BalanceSourceCheckpointDelta
		materialized.CheckpointID = checkpoint.ID
	default:
		return BalancePatch{}, fmt.Errorf(
			"%w: unsupported source kind %d",
			ErrInvalidCalldataPatch,
			patch.Source.Kind(),
		)
	}
	return materialized, nil
}

func deriveCheckpointID(
	chainID int,
	account common.Address,
	stepID StepID,
	callIndex int,
	checkpointIndex int,
	ref amount.CheckpointRef,
) ([32]byte, error) {
	bytes32Type, err := abi.NewType("bytes32", "", nil)
	if err != nil {
		return [32]byte{}, err
	}
	uint256Type, err := abi.NewType("uint256", "", nil)
	if err != nil {
		return [32]byte{}, err
	}
	addressType, err := abi.NewType("address", "", nil)
	if err != nil {
		return [32]byte{}, err
	}
	encoded, err := (abi.Arguments{
		{Type: bytes32Type},
		{Type: uint256Type},
		{Type: addressType},
		{Type: bytes32Type},
		{Type: uint256Type},
		{Type: uint256Type},
		{Type: addressType},
		{Type: bytes32Type},
	}).Pack(
		crypto.Keccak256Hash([]byte("defi-simplify.checkpoint.v1")),
		big.NewInt(int64(chainID)),
		account,
		crypto.Keccak256Hash([]byte(stepID)),
		big.NewInt(int64(callIndex)),
		big.NewInt(int64(checkpointIndex)),
		ref.Token().Address(),
		crypto.Keccak256Hash([]byte(ref.Label())),
	)
	if err != nil {
		return [32]byte{}, fmt.Errorf("encode checkpoint ID: %w", err)
	}
	return crypto.Keccak256Hash(encoded), nil
}

func hasDynamicMetadata(steps []BuiltStep) bool {
	for _, step := range steps {
		for _, call := range step.Calls {
			if len(call.CheckpointsBefore) != 0 || len(call.Patches) != 0 || call.ExpectsCallback {
				return true
			}
		}
	}
	return false
}
