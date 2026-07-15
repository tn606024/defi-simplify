package aave

import (
	"context"
	"fmt"
	"math/big"

	"github.com/shopspring/decimal"
	defi "github.com/tn606024/defi-simplify"
	"github.com/tn606024/defi-simplify/client/contract"
	"github.com/tn606024/defi-simplify/erc20"
	"github.com/tn606024/defi-simplify/helper"
)

type poolStepKind uint8

const (
	supplyStep poolStepKind = iota
	supplyWithPermitStep
	borrowStep
	withdrawStep
	withdrawAllStep
	repayStep
	repayAllStep
	repayWithPermitStep
)

type poolStep struct {
	name      string
	kind      poolStepKind
	reserve   Reserve
	permit    erc20.PermitCapability
	amount    decimal.Decimal
	signature eip712Signature
}

// Supply builds an Aave supply call using the flow account as onBehalfOf.
func Supply(reserve Reserve, amount decimal.Decimal) defi.FlowStep {
	return poolStep{name: "aave.Supply", kind: supplyStep, reserve: reserve, amount: amount}
}

// SupplyWithPermit builds an Aave supplyWithPermit call for the flow account.
func SupplyWithPermit(
	reserve Reserve,
	permit erc20.PermitCapability,
	amount decimal.Decimal,
	deadline *big.Int,
	v uint8,
	r, s [32]byte,
) defi.FlowStep {
	return poolStep{
		name:      "aave.SupplyWithPermit",
		kind:      supplyWithPermitStep,
		reserve:   reserve,
		permit:    permit,
		amount:    amount,
		signature: newEIP712Signature(deadline, v, r, s),
	}
}

// Borrow builds an Aave variable-rate borrow call using the flow account as onBehalfOf.
func Borrow(reserve Reserve, amount decimal.Decimal) defi.FlowStep {
	return poolStep{name: "aave.Borrow", kind: borrowStep, reserve: reserve, amount: amount}
}

// Withdraw builds an Aave withdraw call that sends the asset to the flow account.
func Withdraw(reserve Reserve, amount decimal.Decimal) defi.FlowStep {
	return poolStep{name: "aave.Withdraw", kind: withdrawStep, reserve: reserve, amount: amount}
}

// WithdrawAll builds an Aave withdraw call using the protocol's uint256.max sentinel.
// The Pool emits the actual withdrawn amount, not the sentinel value.
func WithdrawAll(reserve Reserve) defi.FlowStep {
	return poolStep{name: "aave.WithdrawAll", kind: withdrawAllStep, reserve: reserve}
}

// Repay builds an Aave variable-debt repayment call for the flow account.
func Repay(reserve Reserve, amount decimal.Decimal) defi.FlowStep {
	return poolStep{name: "aave.Repay", kind: repayStep, reserve: reserve, amount: amount}
}

// RepayAll builds an Aave variable-debt repayment using the protocol's uint256.max sentinel.
// The Pool emits the actual repaid amount, not the sentinel value.
func RepayAll(reserve Reserve) defi.FlowStep {
	return poolStep{name: "aave.RepayAll", kind: repayAllStep, reserve: reserve}
}

// RepayWithPermit builds an Aave variable-debt repayment using an asset permit signed by the flow account.
func RepayWithPermit(
	reserve Reserve,
	permit erc20.PermitCapability,
	amount decimal.Decimal,
	deadline *big.Int,
	v uint8,
	r, s [32]byte,
) defi.FlowStep {
	return poolStep{
		name:      "aave.RepayWithPermit",
		kind:      repayWithPermitStep,
		reserve:   reserve,
		permit:    permit,
		amount:    amount,
		signature: newEIP712Signature(deadline, v, r, s),
	}
}

func (s poolStep) Build(ctx context.Context, env defi.BuildEnv) (defi.BuiltStep, error) {
	built := defi.BuiltStep{Name: s.name}
	if !s.usesMaxAmount() && !s.amount.IsPositive() {
		return built, fmt.Errorf("amount must be positive")
	}
	resolved, err := resolveStepReserve(s.reserve, env.Chain)
	if err != nil {
		return built, err
	}
	if s.kind == supplyWithPermitStep || s.kind == repayWithPermitStep {
		if err := s.signature.validate(); err != nil {
			return built, err
		}
		if err := validatePermitCapability(s.permit, resolved.underlying); err != nil {
			return built, fmt.Errorf("resolve asset permit capability: %w", err)
		}
	}

	poolAddress := resolved.market.Pool()
	coinAddress := resolved.underlying.Address()
	var amountWei *big.Int
	if s.usesMaxAmount() {
		amountWei = newUint256Max()
	} else {
		amountWei = helper.ToWei(s.amount, resolved.underlying.Decimals())
	}

	var (
		action      defi.Action
		expectation defi.EventExpectation
	)
	switch s.kind {
	case supplyStep:
		action = contract.BuildSupplyAction(poolAddress, coinAddress, amountWei, env.Account)
		expectation = ExpectSupply(poolAddress, coinAddress, env.Account, env.Account, defi.Exact(amountWei))
	case supplyWithPermitStep:
		action = contract.BuildSupplyWithPermitAction(
			poolAddress,
			coinAddress,
			amountWei,
			env.Account,
			0,
			s.signature.deadline,
			s.signature.v,
			s.signature.r,
			s.signature.s,
		)
		expectation = ExpectSupply(poolAddress, coinAddress, env.Account, env.Account, defi.Exact(amountWei))
	case borrowStep:
		action = contract.BuildBorrowAction(poolAddress, coinAddress, amountWei, env.Account)
		expectation = ExpectBorrow(
			poolAddress,
			coinAddress,
			env.Account,
			env.Account,
			VariableInterestRateMode,
			defi.Exact(amountWei),
		)
	case withdrawStep:
		action = contract.BuildWithdrawAction(poolAddress, coinAddress, amountWei, env.Account)
		expectation = ExpectWithdraw(poolAddress, coinAddress, env.Account, env.Account, defi.Exact(amountWei))
	case withdrawAllStep:
		action = contract.BuildWithdrawAction(poolAddress, coinAddress, amountWei, env.Account)
		expectation = ExpectWithdraw(poolAddress, coinAddress, env.Account, env.Account, defi.Positive())
	case repayStep:
		action = contract.BuildRepayAction(poolAddress, coinAddress, amountWei, env.Account)
		expectation = ExpectRepay(
			poolAddress,
			coinAddress,
			env.Account,
			env.Account,
			false,
			defi.Positive(),
			defi.AtMost(amountWei),
		)
	case repayAllStep:
		action = contract.BuildRepayAction(poolAddress, coinAddress, amountWei, env.Account)
		expectation = ExpectRepay(
			poolAddress,
			coinAddress,
			env.Account,
			env.Account,
			false,
			defi.Positive(),
		)
	case repayWithPermitStep:
		action = contract.BuildRepayWithPermitAction(
			poolAddress,
			coinAddress,
			amountWei,
			env.Account,
			s.signature.deadline,
			s.signature.v,
			s.signature.r,
			s.signature.s,
		)
		expectation = ExpectRepay(
			poolAddress,
			coinAddress,
			env.Account,
			env.Account,
			false,
			defi.Positive(),
			defi.AtMost(amountWei),
		)
	default:
		return built, fmt.Errorf("unsupported Aave Pool step kind %d", s.kind)
	}

	call, err := action.ToCall(ctx, env.Conn, nil)
	if err != nil {
		return built, err
	}
	if call == nil {
		return built, fmt.Errorf("action returned nil call")
	}
	built.Calls = []defi.Call{*call}
	built.Expectations = []defi.EventExpectation{expectation}
	return built, nil
}

func (s poolStep) usesMaxAmount() bool {
	return s.kind == withdrawAllStep || s.kind == repayAllStep
}

func newUint256Max() *big.Int {
	max := new(big.Int).Lsh(big.NewInt(1), 256)
	return max.Sub(max, big.NewInt(1))
}
