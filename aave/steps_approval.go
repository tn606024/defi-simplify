package aave

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	defi "github.com/tn606024/defi-simplify"
	"github.com/tn606024/defi-simplify/config"
	"github.com/tn606024/defi-simplify/erc20"
)

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

func (s approveSupplyStep) Build(ctx context.Context, env defi.BuildEnv) (defi.BuiltStep, error) {
	built, err := erc20.Approve(s.coin, PoolSpender(), s.amount).Build(ctx, env)
	built.Name = "aave.ApproveSupply"
	return built, err
}
