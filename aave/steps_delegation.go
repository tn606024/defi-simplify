package aave

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	defi "github.com/tn606024/defi-simplify"
	"github.com/tn606024/defi-simplify/client/contract"
	"github.com/tn606024/defi-simplify/helper"
)

type delegationStepKind uint8

const (
	approveDelegationStep delegationStepKind = iota
	delegationWithSigStep
)

type delegationStep struct {
	name       string
	kind       delegationStepKind
	reserve    Reserve
	capability DelegationCapability
	delegator  common.Address
	delegatee  common.Address
	amount     decimal.Decimal
	signature  eip712Signature
}

// ApproveDelegation lets delegatee borrow an asset against the flow account's position.
func ApproveDelegation(reserve Reserve, delegatee common.Address, amount decimal.Decimal) defi.FlowStep {
	return delegationStep{
		name:      "aave.ApproveDelegation",
		kind:      approveDelegationStep,
		reserve:   reserve,
		delegatee: delegatee,
		amount:    amount,
	}
}

// DelegationWithSig submits a credit-delegation signature from delegator.
// The flow account may be a relayer and does not need to equal delegator.
func DelegationWithSig(
	capability DelegationCapability,
	delegator,
	delegatee common.Address,
	amount decimal.Decimal,
	deadline *big.Int,
	v uint8,
	r, s [32]byte,
) defi.FlowStep {
	return delegationStep{
		name:       "aave.DelegationWithSig",
		kind:       delegationWithSigStep,
		capability: capability,
		delegator:  delegator,
		delegatee:  delegatee,
		amount:     amount,
		signature:  newEIP712Signature(deadline, v, r, s),
	}
}

func (s delegationStep) Build(ctx context.Context, env defi.BuildEnv) (defi.BuiltStep, error) {
	built := defi.BuiltStep{Name: s.name}
	if s.amount.IsNegative() {
		return built, fmt.Errorf("amount must not be negative")
	}
	if s.delegatee == (common.Address{}) {
		return built, fmt.Errorf("delegatee is zero")
	}
	if s.kind == delegationWithSigStep {
		if s.delegator == (common.Address{}) {
			return built, fmt.Errorf("delegator is zero")
		}
		if err := s.signature.validate(); err != nil {
			return built, err
		}
		if err := s.capability.Validate(); err != nil {
			return built, err
		}
		s.reserve = s.capability.Reserve()
	}

	resolved, err := resolveStepReserve(s.reserve, env.Chain)
	if err != nil {
		return built, err
	}
	assetAddress := resolved.underlying.Address()
	debtTokenAddress := resolved.variableDebt.Address()
	amountWei := helper.ToWei(s.amount, resolved.underlying.Decimals())

	var (
		action   defi.Action
		fromUser common.Address
	)
	switch s.kind {
	case approveDelegationStep:
		fromUser = env.Account
		action = contract.BuildApproveDelegationAction(debtTokenAddress, s.delegatee, amountWei)
	case delegationWithSigStep:
		fromUser = s.delegator
		action = contract.BuildDelegationWithSigAction(
			debtTokenAddress,
			s.delegator,
			s.delegatee,
			amountWei,
			s.signature.deadline,
			s.signature.v,
			s.signature.r,
			s.signature.s,
		)
	default:
		return built, fmt.Errorf("unsupported Aave delegation step kind %d", s.kind)
	}

	call, err := action.ToCall(ctx, env.Conn, nil)
	if err != nil {
		return built, err
	}
	if call == nil {
		return built, fmt.Errorf("action returned nil call")
	}
	built.Calls = []defi.Call{*call}
	built.Expectations = []defi.EventExpectation{
		ExpectBorrowAllowanceDelegated(
			debtTokenAddress,
			assetAddress,
			fromUser,
			s.delegatee,
			defi.Exact(amountWei),
		),
	}
	return built, nil
}
