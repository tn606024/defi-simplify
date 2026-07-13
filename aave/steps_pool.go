package aave

import (
	"context"
	"fmt"
	"math/big"

	"github.com/shopspring/decimal"
	defi "github.com/tn606024/defi-simplify"
	"github.com/tn606024/defi-simplify/client/contract"
	"github.com/tn606024/defi-simplify/config"
	"github.com/tn606024/defi-simplify/helper"
)

type poolStepKind uint8

const (
	supplyStep poolStepKind = iota
	supplyWithPermitStep
	borrowStep
	withdrawStep
	repayStep
	repayWithPermitStep
)

type poolStep struct {
	name      string
	kind      poolStepKind
	coin      config.Coin
	amount    decimal.Decimal
	signature eip712Signature
}

// Supply builds an Aave supply call using the flow account as onBehalfOf.
func Supply(coin config.Coin, amount decimal.Decimal) defi.FlowStep {
	return poolStep{name: "aave.Supply", kind: supplyStep, coin: coin, amount: amount}
}

// SupplyWithPermit builds an Aave supplyWithPermit call for the flow account.
func SupplyWithPermit(
	coin config.Coin,
	amount decimal.Decimal,
	deadline *big.Int,
	v uint8,
	r, s [32]byte,
) defi.FlowStep {
	return poolStep{
		name:      "aave.SupplyWithPermit",
		kind:      supplyWithPermitStep,
		coin:      coin,
		amount:    amount,
		signature: newEIP712Signature(deadline, v, r, s),
	}
}

// Borrow builds an Aave variable-rate borrow call using the flow account as onBehalfOf.
func Borrow(coin config.Coin, amount decimal.Decimal) defi.FlowStep {
	return poolStep{name: "aave.Borrow", kind: borrowStep, coin: coin, amount: amount}
}

// Withdraw builds an Aave withdraw call that sends the asset to the flow account.
func Withdraw(coin config.Coin, amount decimal.Decimal) defi.FlowStep {
	return poolStep{name: "aave.Withdraw", kind: withdrawStep, coin: coin, amount: amount}
}

// Repay builds an Aave variable-debt repayment call for the flow account.
func Repay(coin config.Coin, amount decimal.Decimal) defi.FlowStep {
	return poolStep{name: "aave.Repay", kind: repayStep, coin: coin, amount: amount}
}

// RepayWithPermit builds an Aave variable-debt repayment using an asset permit signed by the flow account.
func RepayWithPermit(
	coin config.Coin,
	amount decimal.Decimal,
	deadline *big.Int,
	v uint8,
	r, s [32]byte,
) defi.FlowStep {
	return poolStep{
		name:      "aave.RepayWithPermit",
		kind:      repayWithPermitStep,
		coin:      coin,
		amount:    amount,
		signature: newEIP712Signature(deadline, v, r, s),
	}
}

func (s poolStep) Build(ctx context.Context, env defi.BuildEnv) (defi.BuiltStep, error) {
	built := defi.BuiltStep{Name: s.name}
	if !s.amount.IsPositive() {
		return built, fmt.Errorf("amount must be positive")
	}
	if s.kind == supplyWithPermitStep || s.kind == repayWithPermitStep {
		if err := s.signature.validate(); err != nil {
			return built, err
		}
		permitSupported, err := s.coin.PermitSupported(env.Chain)
		if err != nil {
			return built, fmt.Errorf("resolve asset permit support: %w", err)
		}
		if !permitSupported {
			return built, fmt.Errorf("asset does not support permit on chain %d", env.Chain)
		}
	}

	poolAddress, err := env.Chain.AaveV3PoolAddress()
	if err != nil {
		return built, fmt.Errorf("resolve Aave pool: %w", err)
	}
	coinAddress, err := s.coin.Address(env.Chain)
	if err != nil {
		return built, fmt.Errorf("resolve asset: %w", err)
	}
	decimals, err := s.coin.Decimals()
	if err != nil {
		return built, fmt.Errorf("resolve asset decimals: %w", err)
	}
	amountWei := helper.ToWei(s.amount, decimals)

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
