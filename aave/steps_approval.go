package aave

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	defi "github.com/tn606024/defi-simplify"
	"github.com/tn606024/defi-simplify/amount"
	"github.com/tn606024/defi-simplify/config"
	"github.com/tn606024/defi-simplify/erc20"
)

type approveSupplyStep struct {
	reserve Reserve
	amount  amount.Source
}

// PoolSpender resolves one reviewed Aave market's Pool as an ERC20 spender.
func PoolSpender(market Market) erc20.Spender {
	return erc20.SpenderFunc(func(chain config.Chain) (common.Address, error) {
		if err := market.Validate(); err != nil {
			return common.Address{}, err
		}
		if market.Chain() != chain {
			return common.Address{}, fmt.Errorf(
				"Aave market chain %d does not match flow chain %d",
				market.Chain(),
				chain,
			)
		}
		return market.Pool(), nil
	})
}

// ApproveSupply builds an ERC20 approval call for supplying a token into Aave.
func ApproveSupply(reserve Reserve, value amount.Source) defi.FlowStep {
	return approveSupplyStep{
		reserve: reserve,
		amount:  value,
	}
}

func (s approveSupplyStep) Build(ctx context.Context, env defi.BuildEnv) (defi.BuiltStep, error) {
	resolved, err := resolveStepReserve(s.reserve, env.Chain)
	if err != nil {
		return defi.BuiltStep{Name: "aave.ApproveSupply"}, err
	}
	built, err := erc20.Approve(
		resolved.underlying,
		PoolSpender(resolved.market),
		s.amount,
	).Build(ctx, env)
	built.Name = "aave.ApproveSupply"
	return built, err
}
