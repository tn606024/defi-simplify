package aave

import (
	"fmt"

	"github.com/tn606024/defi-simplify/config"
	"github.com/tn606024/defi-simplify/token"
)

type resolvedStepReserve struct {
	market       Market
	underlying   token.Token
	aToken       token.Token
	variableDebt token.Token
}

func resolveStepReserve(reserve Reserve, chain config.Chain) (resolvedStepReserve, error) {
	if err := reserve.Validate(); err != nil {
		return resolvedStepReserve{}, fmt.Errorf("resolve Aave reserve: %w", err)
	}
	market := reserve.Market()
	if market.Chain() != chain {
		return resolvedStepReserve{}, fmt.Errorf(
			"reserve market chain %d does not match flow chain %d",
			market.Chain(),
			chain,
		)
	}
	return resolvedStepReserve{
		market:       market,
		underlying:   reserve.Underlying(),
		aToken:       reserve.AToken(),
		variableDebt: reserve.VariableDebtToken(),
	}, nil
}
