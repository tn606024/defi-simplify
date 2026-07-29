package aave

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/tn606024/defi-simplify/config"
	"github.com/tn606024/defi-simplify/erc20"
)

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
