package aave

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"
	defi "github.com/tn606024/defi-simplify"
	"github.com/tn606024/defi-simplify/client/contract"
	"github.com/tn606024/defi-simplify/config"
	"github.com/tn606024/defi-simplify/helper"
)

type stepKind int

const (
	approveStep stepKind = iota
	supplyStep
	borrowStep
)

type step struct {
	name   string
	kind   stepKind
	coin   config.Coin
	amount decimal.Decimal
}

// Approve builds an ERC20 approval call for the configured chain's Aave Pool.
func Approve(coin config.Coin, amount decimal.Decimal) defi.FlowStep {
	return step{
		name:   "aave.Approve",
		kind:   approveStep,
		coin:   coin,
		amount: amount,
	}
}

// ApproveSupply is a readability alias for approving a token before Aave supply.
func ApproveSupply(coin config.Coin, amount decimal.Decimal) defi.FlowStep {
	return Approve(coin, amount)
}

// Supply builds an Aave supply call using the flow account as onBehalfOf.
func Supply(coin config.Coin, amount decimal.Decimal) defi.FlowStep {
	return step{
		name:   "aave.Supply",
		kind:   supplyStep,
		coin:   coin,
		amount: amount,
	}
}

// Borrow builds an Aave variable-rate borrow call using the flow account as onBehalfOf.
func Borrow(coin config.Coin, amount decimal.Decimal) defi.FlowStep {
	return step{
		name:   "aave.Borrow",
		kind:   borrowStep,
		coin:   coin,
		amount: amount,
	}
}

func (s step) FlowStepName() string {
	return s.name
}

func (s step) BuildCalls(ctx context.Context, env defi.BuildEnv) ([]defi.Call, error) {
	if !s.amount.IsPositive() {
		return nil, fmt.Errorf("amount must be positive")
	}

	poolAddress, err := env.Chain.AaveV3PoolAddress()
	if err != nil {
		return nil, fmt.Errorf("resolve Aave pool: %w", err)
	}
	coinAddress, err := s.coin.Address(env.Chain)
	if err != nil {
		return nil, fmt.Errorf("resolve asset: %w", err)
	}
	decimals, err := s.coin.Decimals()
	if err != nil {
		return nil, fmt.Errorf("resolve asset decimals: %w", err)
	}
	amountWei := helper.ToWei(s.amount, decimals)

	var action defi.Action
	switch s.kind {
	case approveStep:
		action = contract.BuildApproveAction(coinAddress, poolAddress, amountWei)
	case supplyStep:
		action = contract.BuildSupplyAction(poolAddress, coinAddress, amountWei, env.Account)
	case borrowStep:
		action = contract.BuildBorrowAction(poolAddress, coinAddress, amountWei, env.Account)
	default:
		return nil, fmt.Errorf("unsupported Aave step kind %d", s.kind)
	}

	call, err := action.ToCall(ctx, env.Conn, nil)
	if err != nil {
		return nil, err
	}
	if call == nil {
		return nil, fmt.Errorf("action returned nil call")
	}
	return []defi.Call{*call}, nil
}
