package weth

import (
	"context"
	"fmt"
	"math/big"

	defi "github.com/tn606024/defi-simplify"
	"github.com/tn606024/defi-simplify/amount"
	"github.com/tn606024/defi-simplify/helper"
	"github.com/tn606024/defi-simplify/token"
	"github.com/tn606024/defi-simplify/weth/bindings"
)

const withdrawAmountOffset uint32 = 4

type stepKind uint8

const (
	wrapStep stepKind = iota
	unwrapStep
)

type step struct {
	name   string
	kind   stepKind
	asset  token.Ref
	amount amount.Source
}

// Wrap builds a payable WETH deposit call. Wrap accepts exact amounts because
// runtime native-value patching is not supported by the dynamic account.
func Wrap(asset token.Ref, value amount.Source) defi.FlowStep {
	return step{
		name:   "weth.Wrap",
		kind:   wrapStep,
		asset:  asset,
		amount: value,
	}
}

// Unwrap builds a WETH withdrawal call. Exact, current-balance, and
// checkpoint-delta amount sources are supported.
func Unwrap(asset token.Ref, value amount.Source) defi.FlowStep {
	return step{
		name:   "weth.Unwrap",
		kind:   unwrapStep,
		asset:  asset,
		amount: value,
	}
}

func (s step) Build(_ context.Context, env defi.BuildEnv) (defi.BuiltStep, error) {
	built := defi.BuiltStep{Name: s.name}
	if err := s.validateAsset(env); err != nil {
		return built, err
	}
	decimals, err := env.Chain.GasTokenDecimals()
	if err != nil {
		return built, fmt.Errorf("resolve native token decimals: %w", err)
	}

	parsed, err := bindings.WETHMetaData.GetAbi()
	if err != nil {
		return built, fmt.Errorf("load WETH ABI: %w", err)
	}

	switch s.kind {
	case wrapStep:
		amountWei, err := resolveExactAmount(s.amount, decimals)
		if err != nil {
			return built, fmt.Errorf("resolve wrap amount: %w", err)
		}
		data, err := parsed.Pack("deposit")
		if err != nil {
			return built, fmt.Errorf("encode WETH deposit: %w", err)
		}
		built.Calls = []defi.PlannedCall{{
			Call: defi.Call{
				Target: s.asset.Address(),
				Value:  amountWei,
				Data:   data,
			},
		}}
		built.Expectations = []defi.EventExpectation{
			ExpectDeposit(s.asset.Address(), env.Account, defi.Exact(amountWei)),
		}
	case unwrapStep:
		amountWei, patch, exact, err := resolveUnwrapAmount(s.amount, s.asset, decimals)
		if err != nil {
			return built, fmt.Errorf("resolve unwrap amount: %w", err)
		}
		data, err := parsed.Pack("withdraw", amountWei)
		if err != nil {
			return built, fmt.Errorf("encode WETH withdrawal: %w", err)
		}
		planned := defi.PlannedCall{
			Call: defi.Call{
				Target: s.asset.Address(),
				Value:  new(big.Int),
				Data:   data,
			},
		}
		if patch != nil {
			planned.Patches = []defi.CalldataPatch{*patch}
		}
		var constraint defi.AmountConstraint = defi.Positive()
		if exact {
			constraint = defi.Exact(amountWei)
		}
		built.Calls = []defi.PlannedCall{planned}
		built.Expectations = []defi.EventExpectation{
			ExpectWithdrawal(s.asset.Address(), env.Account, constraint),
		}
	default:
		return built, fmt.Errorf("unsupported WETH step kind %d", s.kind)
	}
	return built, nil
}

func (s step) validateAsset(env defi.BuildEnv) error {
	if err := s.asset.Validate(); err != nil {
		return fmt.Errorf("resolve WETH contract: %w", err)
	}
	if s.asset.Chain() != env.Chain {
		return fmt.Errorf(
			"WETH chain %d does not match flow chain %d",
			s.asset.Chain(),
			env.Chain,
		)
	}
	return nil
}

func resolveExactAmount(source amount.Source, decimals uint8) (*big.Int, error) {
	value, exact, err := resolveExactValue(source, decimals)
	if err != nil {
		return nil, err
	}
	if !exact {
		return nil, fmt.Errorf("runtime amount source is not supported")
	}
	return value, nil
}

func resolveUnwrapAmount(
	source amount.Source,
	asset token.Ref,
	decimals uint8,
) (*big.Int, *defi.CalldataPatch, bool, error) {
	value, exact, err := resolveExactValue(source, decimals)
	if err != nil {
		return nil, nil, false, err
	}
	if exact {
		return value, nil, true, nil
	}
	sourceToken, _ := source.Token()
	if !sourceToken.SameAsset(asset) {
		return nil, nil, false, fmt.Errorf(
			"source token %s does not match WETH contract %s",
			sourceToken.Address().Hex(),
			asset.Address().Hex(),
		)
	}
	return new(big.Int), &defi.CalldataPatch{
		Source: source,
		Offset: withdrawAmountOffset,
	}, false, nil
}

func resolveExactValue(
	source amount.Source,
	decimals uint8,
) (*big.Int, bool, error) {
	if err := source.Validate(); err != nil {
		return nil, false, err
	}
	exact, ok := source.ExactValue()
	if !ok {
		return nil, false, nil
	}
	if !exact.IsPositive() {
		return nil, false, fmt.Errorf("amount must be positive")
	}
	value := helper.ToWei(exact, decimals)
	if value.Sign() == 0 {
		return nil, false, fmt.Errorf("amount rounds to zero at %d decimals", decimals)
	}
	return value, true, nil
}
