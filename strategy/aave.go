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
	"github.com/tn606024/defi-simplify/config"
	"github.com/tn606024/defi-simplify/erc20"
)

// AaveSupplyBorrowParams configures a static Aave supply-and-borrow flow.
type AaveSupplyBorrowParams struct {
	Account      common.Address
	Chain        config.Chain
	SupplyAsset  config.Coin
	SupplyAmount decimal.Decimal
	BorrowAsset  config.Coin
	BorrowAmount decimal.Decimal
}

// AaveSupplyBorrow builds ApproveSupply -> Supply -> Borrow with exact amounts.
func AaveSupplyBorrow(params AaveSupplyBorrowParams) (*defi.Flow, error) {
	if err := validateAccountAndChain(params.Account, params.Chain); err != nil {
		return nil, err
	}
	if err := validateAsset("supply asset", params.SupplyAsset, params.Chain); err != nil {
		return nil, err
	}
	if !params.SupplyAmount.IsPositive() {
		return nil, fmt.Errorf("supply amount must be positive")
	}
	if err := validateAsset("borrow asset", params.BorrowAsset, params.Chain); err != nil {
		return nil, err
	}
	if !params.BorrowAmount.IsPositive() {
		return nil, fmt.Errorf("borrow amount must be positive")
	}

	return defi.NewFlow(params.Account, defi.WithChain(params.Chain)).
		Add(aave.ApproveSupply(params.SupplyAsset, params.SupplyAmount)).
		Add(aave.Supply(params.SupplyAsset, params.SupplyAmount)).
		Add(aave.Borrow(params.BorrowAsset, params.BorrowAmount)), nil
}

// AaveClosePositionParams configures a static close flow for one Aave variable
// debt reserve and one collateral reserve.
type AaveClosePositionParams struct {
	Account                 common.Address
	Chain                   config.Chain
	DebtAsset               config.Coin
	TemporaryRepayAllowance decimal.Decimal
	CollateralAsset         config.Coin
}

// AaveClosePosition builds Approve -> RepayAll -> revoke approval ->
// WithdrawAll. TemporaryRepayAllowance is an upper allowance bound; the caller
// must hold enough of DebtAsset to cover the actual debt at execution time.
// DebtAsset must support replacing an existing ERC20 allowance directly.
func AaveClosePosition(params AaveClosePositionParams) (*defi.Flow, error) {
	if err := validateAccountAndChain(params.Account, params.Chain); err != nil {
		return nil, err
	}
	if err := validateAsset("debt asset", params.DebtAsset, params.Chain); err != nil {
		return nil, err
	}
	if !params.TemporaryRepayAllowance.IsPositive() {
		return nil, fmt.Errorf("temporary repay allowance must be positive")
	}
	if err := validateAsset("collateral asset", params.CollateralAsset, params.Chain); err != nil {
		return nil, err
	}

	return defi.NewFlow(params.Account, defi.WithChain(params.Chain)).
		Add(erc20.Approve(params.DebtAsset, aave.PoolSpender(), params.TemporaryRepayAllowance)).
		Add(aave.RepayAll(params.DebtAsset)).
		Add(erc20.Approve(params.DebtAsset, aave.PoolSpender(), decimal.Zero)).
		Add(aave.WithdrawAll(params.CollateralAsset)), nil
}

func validateAccountAndChain(account common.Address, chain config.Chain) error {
	if account == (common.Address{}) {
		return fmt.Errorf("account must not be zero")
	}
	if _, err := chain.Name(); err != nil {
		return fmt.Errorf("chain: %w", err)
	}
	return nil
}

func validateAsset(label string, asset config.Coin, chain config.Chain) error {
	if _, err := asset.Address(chain); err != nil {
		return fmt.Errorf("%s: %w", label, err)
	}
	if _, err := asset.Decimals(); err != nil {
		return fmt.Errorf("%s: %w", label, err)
	}
	return nil
}
