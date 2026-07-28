package aave

import (
	"fmt"
	"math/big"

	defi "github.com/tn606024/defi-simplify"
	"github.com/tn606024/defi-simplify/amount"
	"github.com/tn606024/defi-simplify/helper"
	"github.com/tn606024/defi-simplify/token"
)

func resolveExactReserveAmount(
	source amount.Source,
	asset token.Token,
	requirePositive bool,
) (*big.Int, error) {
	value, exact, err := resolveReserveExactValue(source, asset, requirePositive)
	if err != nil {
		return nil, err
	}
	if !exact {
		return nil, fmt.Errorf("resolve amount: runtime amount source is not supported for this Aave step")
	}
	return value, nil
}

func resolvePatchableReserveAmount(
	source amount.Source,
	asset token.Token,
	offset uint32,
	requirePositive bool,
) (*big.Int, *defi.CalldataPatch, bool, error) {
	value, exact, err := resolveReserveExactValue(source, asset, requirePositive)
	if err != nil {
		return nil, nil, false, err
	}
	if exact {
		return value, nil, true, nil
	}
	sourceToken, _ := source.Token()
	if !sourceToken.SameAsset(asset.Ref()) {
		return nil, nil, false, fmt.Errorf(
			"resolve amount: source token %s does not match Aave asset %s",
			sourceToken.Address().Hex(),
			asset.Address().Hex(),
		)
	}
	return new(big.Int), &defi.CalldataPatch{Source: source, Offset: offset}, false, nil
}

func resolveReserveExactValue(
	source amount.Source,
	asset token.Token,
	requirePositive bool,
) (*big.Int, bool, error) {
	if err := source.Validate(); err != nil {
		return nil, false, fmt.Errorf("resolve amount: %w", err)
	}
	if exact, ok := source.ExactValue(); ok {
		if requirePositive && !exact.IsPositive() {
			return nil, false, fmt.Errorf("amount must be positive")
		}
		value := helper.ToWei(exact, asset.Decimals())
		return value, true, nil
	}
	return nil, false, nil
}

func exactOrPositiveAmountConstraint(exact bool, value *big.Int) defi.AmountConstraint {
	if exact {
		return defi.Exact(value)
	}
	return defi.Positive()
}
