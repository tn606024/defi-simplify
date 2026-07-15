// Package strategy provides opinionated Flow compositions for common DeFi
// workflows. Strategy builders only construct flows; callers retain ownership
// of execution, signing, and submission.
package strategy

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	defi "github.com/tn606024/defi-simplify"
	"github.com/tn606024/defi-simplify/aave"
	"github.com/tn606024/defi-simplify/erc20"
)

// AaveSupplyBorrowParams configures a static Aave supply-and-borrow flow.
type AaveSupplyBorrowParams struct {
	Account       common.Address
	SupplyReserve aave.Reserve
	SupplyAmount  decimal.Decimal
	BorrowReserve aave.Reserve
	BorrowAmount  decimal.Decimal
}

// AaveSupplyBorrow builds ApproveSupply -> Supply -> Borrow with exact amounts.
func AaveSupplyBorrow(params AaveSupplyBorrowParams) (*defi.Flow, error) {
	market, err := validateReserves(
		params.Account,
		"supply reserve",
		params.SupplyReserve,
		"borrow reserve",
		params.BorrowReserve,
	)
	if err != nil {
		return nil, err
	}
	if !params.SupplyAmount.IsPositive() {
		return nil, fmt.Errorf("supply amount must be positive")
	}
	if !params.BorrowAmount.IsPositive() {
		return nil, fmt.Errorf("borrow amount must be positive")
	}

	return defi.NewFlow(params.Account, defi.WithChain(market.Chain())).
		Add(aave.ApproveSupply(params.SupplyReserve, params.SupplyAmount)).
		Add(aave.Supply(params.SupplyReserve, params.SupplyAmount)).
		Add(aave.Borrow(params.BorrowReserve, params.BorrowAmount)), nil
}

// AaveClosePositionParams configures a static close flow for one Aave variable
// debt reserve and one collateral reserve.
type AaveClosePositionParams struct {
	Account                 common.Address
	DebtReserve             aave.Reserve
	TemporaryRepayAllowance decimal.Decimal
	CollateralReserve       aave.Reserve
}

// AaveClosePosition builds Approve -> RepayAll -> revoke approval ->
// WithdrawAll. TemporaryRepayAllowance is an upper allowance bound; the caller
// must hold enough of DebtAsset to cover the actual debt at execution time.
// DebtAsset must support replacing an existing ERC20 allowance directly.
func AaveClosePosition(params AaveClosePositionParams) (*defi.Flow, error) {
	market, err := validateReserves(
		params.Account,
		"debt reserve",
		params.DebtReserve,
		"collateral reserve",
		params.CollateralReserve,
	)
	if err != nil {
		return nil, err
	}
	if !params.TemporaryRepayAllowance.IsPositive() {
		return nil, fmt.Errorf("temporary repay allowance must be positive")
	}

	return defi.NewFlow(params.Account, defi.WithChain(market.Chain())).
		Add(erc20.Approve(
			params.DebtReserve.Underlying(),
			aave.PoolSpender(market),
			params.TemporaryRepayAllowance,
		)).
		Add(aave.RepayAll(params.DebtReserve)).
		Add(erc20.Approve(params.DebtReserve.Underlying(), aave.PoolSpender(market), decimal.Zero)).
		Add(aave.WithdrawAll(params.CollateralReserve)), nil
}

func validateReserves(
	account common.Address,
	firstLabel string,
	first aave.Reserve,
	secondLabel string,
	second aave.Reserve,
) (aave.Market, error) {
	if account == (common.Address{}) {
		return aave.Market{}, fmt.Errorf("account must not be zero")
	}
	if err := first.Validate(); err != nil {
		return aave.Market{}, fmt.Errorf("%s: %w", firstLabel, err)
	}
	if err := second.Validate(); err != nil {
		return aave.Market{}, fmt.Errorf("%s: %w", secondLabel, err)
	}
	if !first.Market().SameMarket(second.Market()) {
		return aave.Market{}, fmt.Errorf("%s and %s must belong to the same Aave market", firstLabel, secondLabel)
	}
	return first.Market(), nil
}
