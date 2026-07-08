package aave

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	defi "github.com/tn606024/defi-simplify"
	"github.com/tn606024/defi-simplify/client/contract"
	"github.com/tn606024/defi-simplify/config"
	"github.com/tn606024/defi-simplify/erc20"
	"github.com/tn606024/defi-simplify/helper"
)

type stepKind int

const (
	supplyStep stepKind = iota
	borrowStep
)

type step struct {
	name   string
	kind   stepKind
	coin   config.Coin
	amount decimal.Decimal
}

type approveSupplyStep struct {
	coin   config.Coin
	amount decimal.Decimal
}

// PoolSpender resolves the configured chain's Aave V3 Pool as an ERC20 spender.
func PoolSpender() erc20.Spender {
	return erc20.SpenderFunc(func(chain config.Chain) (common.Address, error) {
		return chain.AaveV3PoolAddress()
	})
}

// ApproveSupply builds an ERC20 approval call for supplying a token into Aave.
func ApproveSupply(coin config.Coin, amount decimal.Decimal) defi.FlowStep {
	return approveSupplyStep{
		coin:   coin,
		amount: amount,
	}
}

func (s approveSupplyStep) FlowStepName() string {
	return "aave.ApproveSupply"
}

func (s approveSupplyStep) BuildCalls(ctx context.Context, env defi.BuildEnv) ([]defi.Call, error) {
	return erc20.Approve(s.coin, PoolSpender(), s.amount).BuildCalls(ctx, env)
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
