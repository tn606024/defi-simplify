package aave

import (
	"context"
	"fmt"
	"math/big"

	defi "github.com/tn606024/defi-simplify"
	"github.com/tn606024/defi-simplify/amount"
	"github.com/tn606024/defi-simplify/client/contract"
)

const poolAmountOffset uint32 = 36

type poolStepKind uint8

const (
	supplyStep poolStepKind = iota
	borrowStep
	withdrawStep
	withdrawAllStep
	repayStep
	repayAllStep
)

type poolStep struct {
	name    string
	kind    poolStepKind
	reserve Reserve
	amount  amount.Source
}

// Supply builds an Aave supply call using the flow account as onBehalfOf.
func Supply(reserve Reserve, value amount.Source) defi.FlowStep {
	return poolStep{name: "aave.Supply", kind: supplyStep, reserve: reserve, amount: value}
}

// Borrow builds an Aave variable-rate borrow call using the flow account as onBehalfOf.
func Borrow(reserve Reserve, value amount.Source) defi.FlowStep {
	return poolStep{name: "aave.Borrow", kind: borrowStep, reserve: reserve, amount: value}
}

// Withdraw builds an Aave withdraw call that sends the asset to the flow account.
func Withdraw(reserve Reserve, value amount.Source) defi.FlowStep {
	return poolStep{name: "aave.Withdraw", kind: withdrawStep, reserve: reserve, amount: value}
}

// WithdrawAll builds an Aave withdraw call using the protocol's uint256.max sentinel.
// The Pool emits the actual withdrawn amount, not the sentinel value.
func WithdrawAll(reserve Reserve) defi.FlowStep {
	return poolStep{name: "aave.WithdrawAll", kind: withdrawAllStep, reserve: reserve}
}

// Repay builds an Aave variable-debt repayment call for the flow account.
func Repay(reserve Reserve, value amount.Source) defi.FlowStep {
	return poolStep{name: "aave.Repay", kind: repayStep, reserve: reserve, amount: value}
}

// RepayAll builds an Aave variable-debt repayment using the protocol's uint256.max sentinel.
// The Pool emits the actual repaid amount, not the sentinel value.
func RepayAll(reserve Reserve) defi.FlowStep {
	return poolStep{name: "aave.RepayAll", kind: repayAllStep, reserve: reserve}
}

func (s poolStep) Build(ctx context.Context, env defi.BuildEnv) (defi.BuiltStep, error) {
	built := defi.BuiltStep{Name: s.name}
	resolved, err := resolveStepReserve(s.reserve, env.Chain)
	if err != nil {
		return built, err
	}
	poolAddress := resolved.market.Pool()
	coinAddress := resolved.underlying.Address()
	var (
		amountWei *big.Int
		patch     *defi.CalldataPatch
		exact     bool
	)
	if s.usesMaxAmount() {
		amountWei = newUint256Max()
		exact = true
	} else {
		amountWei, patch, exact, err = resolvePatchableReserveAmount(
			s.amount,
			resolved.underlying,
			poolAmountOffset,
			true,
		)
	}
	if err != nil {
		return built, err
	}

	var (
		action      defi.Action
		expectation defi.EventExpectation
	)
	switch s.kind {
	case supplyStep:
		action = contract.BuildSupplyAction(poolAddress, coinAddress, amountWei, env.Account)
		expectation = ExpectSupply(
			poolAddress,
			coinAddress,
			env.Account,
			env.Account,
			exactOrPositiveAmountConstraint(exact, amountWei),
		)
	case borrowStep:
		action = contract.BuildBorrowAction(poolAddress, coinAddress, amountWei, env.Account)
		expectation = ExpectBorrow(
			poolAddress,
			coinAddress,
			env.Account,
			env.Account,
			VariableInterestRateMode,
			exactOrPositiveAmountConstraint(exact, amountWei),
		)
	case withdrawStep:
		action = contract.BuildWithdrawAction(poolAddress, coinAddress, amountWei, env.Account)
		expectation = ExpectWithdraw(
			poolAddress,
			coinAddress,
			env.Account,
			env.Account,
			exactOrPositiveAmountConstraint(exact, amountWei),
		)
	case withdrawAllStep:
		action = contract.BuildWithdrawAction(poolAddress, coinAddress, amountWei, env.Account)
		expectation = ExpectWithdraw(poolAddress, coinAddress, env.Account, env.Account, defi.Positive())
	case repayStep:
		action = contract.BuildRepayAction(poolAddress, coinAddress, amountWei, env.Account)
		constraints := []defi.AmountConstraint{defi.Positive()}
		if exact {
			constraints = append(constraints, defi.AtMost(amountWei))
		}
		expectation = ExpectRepay(poolAddress, coinAddress, env.Account, env.Account, false, constraints...)
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
	planned := defi.PlannedCall{Call: *call}
	if patch != nil {
		planned.Patches = []defi.CalldataPatch{*patch}
	}
	built.Calls = []defi.PlannedCall{planned}
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
