package strategy

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	defi "github.com/tn606024/defi-simplify"
	"github.com/tn606024/defi-simplify/aave"
	"github.com/tn606024/defi-simplify/amount"
	"github.com/tn606024/defi-simplify/erc20"
	"github.com/tn606024/defi-simplify/weth"
)

// AaveSupplyNativeETHParams configures an exact native-ETH supply through the
// Aave wrapped-native reserve.
type AaveSupplyNativeETHParams struct {
	Account common.Address
	Reserve aave.Reserve
	Amount  decimal.Decimal
}

// AaveSupplyNativeETH builds Wrap -> Approve -> Supply.
func AaveSupplyNativeETH(params AaveSupplyNativeETHParams) (*defi.Flow, error) {
	market, err := validateNativeETHReserve(params.Account, params.Reserve)
	if err != nil {
		return nil, err
	}
	if !params.Amount.IsPositive() {
		return nil, fmt.Errorf("supply amount must be positive")
	}

	exact := amount.Exact(params.Amount)
	return defi.NewFlow(params.Account, defi.WithChain(market.Chain())).
		Add(weth.Wrap(params.Reserve.Underlying().Ref(), exact)).
		Add(erc20.Approve(
			params.Reserve.Underlying(),
			aave.PoolSpender(market),
			exact,
		)).
		Add(aave.Supply(params.Reserve, exact)), nil
}

// AaveBorrowNativeETHParams configures an exact wrapped-native Aave borrow
// whose newly borrowed WETH is returned as native ETH.
type AaveBorrowNativeETHParams struct {
	Account common.Address
	Reserve aave.Reserve
	Amount  decimal.Decimal
}

// AaveBorrowNativeETH builds checkpoint -> Borrow -> Unwrap(checkpoint delta).
func AaveBorrowNativeETH(params AaveBorrowNativeETHParams) (*defi.Flow, error) {
	market, err := validateNativeETHReserve(params.Account, params.Reserve)
	if err != nil {
		return nil, err
	}
	if !params.Amount.IsPositive() {
		return nil, fmt.Errorf("borrow amount must be positive")
	}

	asset := params.Reserve.Underlying().Ref()
	checkpoint := amount.Checkpoint("before-native-borrow", asset)
	return defi.NewFlow(params.Account, defi.WithChain(market.Chain())).
		Add(defi.CheckpointBefore(
			aave.Borrow(params.Reserve, amount.Exact(params.Amount)),
			checkpoint,
		)).
		Add(weth.Unwrap(asset, amount.CheckpointDelta(checkpoint))), nil
}

// AaveWithdrawNativeETHParams configures an exact wrapped-native Aave
// collateral withdrawal returned as native ETH.
type AaveWithdrawNativeETHParams struct {
	Account common.Address
	Reserve aave.Reserve
	Amount  decimal.Decimal
}

// AaveWithdrawNativeETH builds checkpoint -> Withdraw ->
// Unwrap(checkpoint delta).
func AaveWithdrawNativeETH(params AaveWithdrawNativeETHParams) (*defi.Flow, error) {
	market, err := validateNativeETHReserve(params.Account, params.Reserve)
	if err != nil {
		return nil, err
	}
	if !params.Amount.IsPositive() {
		return nil, fmt.Errorf("withdraw amount must be positive")
	}

	return buildAaveWithdrawNativeETH(
		params.Account,
		market,
		params.Reserve,
		aave.Withdraw(params.Reserve, amount.Exact(params.Amount)),
	), nil
}

// AaveWithdrawAllNativeETHParams configures a full wrapped-native Aave
// collateral withdrawal returned as native ETH.
type AaveWithdrawAllNativeETHParams struct {
	Account common.Address
	Reserve aave.Reserve
}

// AaveWithdrawAllNativeETH builds checkpoint -> WithdrawAll ->
// Unwrap(checkpoint delta).
func AaveWithdrawAllNativeETH(params AaveWithdrawAllNativeETHParams) (*defi.Flow, error) {
	market, err := validateNativeETHReserve(params.Account, params.Reserve)
	if err != nil {
		return nil, err
	}
	return buildAaveWithdrawNativeETH(
		params.Account,
		market,
		params.Reserve,
		aave.WithdrawAll(params.Reserve),
	), nil
}

func buildAaveWithdrawNativeETH(
	account common.Address,
	market aave.Market,
	reserve aave.Reserve,
	withdraw defi.FlowStep,
) *defi.Flow {
	asset := reserve.Underlying().Ref()
	checkpoint := amount.Checkpoint("before-native-withdraw", asset)
	return defi.NewFlow(account, defi.WithChain(market.Chain())).
		Add(defi.CheckpointBefore(withdraw, checkpoint)).
		Add(weth.Unwrap(asset, amount.CheckpointDelta(checkpoint)))
}

func validateNativeETHReserve(account common.Address, reserve aave.Reserve) (aave.Market, error) {
	if account == (common.Address{}) {
		return aave.Market{}, fmt.Errorf("account must not be zero")
	}
	if err := reserve.Validate(); err != nil {
		return aave.Market{}, fmt.Errorf("wrapped native reserve: %w", err)
	}
	return reserve.Market(), nil
}
