package strategy_test

import (
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	"github.com/tn606024/defi-simplify/config"
	"github.com/tn606024/defi-simplify/strategy"
)

func TestAaveSupplyBorrowValidation(t *testing.T) {
	valid := strategy.AaveSupplyBorrowParams{
		Account:      common.HexToAddress("0x00000000000000000000000000000000000000aa"),
		Chain:        config.Base,
		SupplyAsset:  config.USDC,
		SupplyAmount: decimal.NewFromInt(100),
		BorrowAsset:  config.WETH,
		BorrowAmount: decimal.RequireFromString("0.01"),
	}

	tests := []struct {
		name    string
		mutate  func(*strategy.AaveSupplyBorrowParams)
		wantErr string
	}{
		{name: "zero account", mutate: func(p *strategy.AaveSupplyBorrowParams) { p.Account = common.Address{} }, wantErr: "account must not be zero"},
		{name: "unsupported chain", mutate: func(p *strategy.AaveSupplyBorrowParams) { p.Chain = config.Chain(999) }, wantErr: "chain"},
		{name: "unsupported supply asset", mutate: func(p *strategy.AaveSupplyBorrowParams) { p.SupplyAsset = config.Coin(999) }, wantErr: "supply asset"},
		{name: "zero supply amount", mutate: func(p *strategy.AaveSupplyBorrowParams) { p.SupplyAmount = decimal.Zero }, wantErr: "supply amount must be positive"},
		{name: "negative supply amount", mutate: func(p *strategy.AaveSupplyBorrowParams) { p.SupplyAmount = decimal.NewFromInt(-1) }, wantErr: "supply amount must be positive"},
		{name: "unsupported borrow asset", mutate: func(p *strategy.AaveSupplyBorrowParams) { p.BorrowAsset = config.Coin(999) }, wantErr: "borrow asset"},
		{name: "zero borrow amount", mutate: func(p *strategy.AaveSupplyBorrowParams) { p.BorrowAmount = decimal.Zero }, wantErr: "borrow amount must be positive"},
		{name: "negative borrow amount", mutate: func(p *strategy.AaveSupplyBorrowParams) { p.BorrowAmount = decimal.NewFromInt(-1) }, wantErr: "borrow amount must be positive"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := valid
			tt.mutate(&params)

			flow, err := strategy.AaveSupplyBorrow(params)

			if flow != nil {
				t.Fatalf("expected nil flow, got %#v", flow)
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestAaveClosePositionValidation(t *testing.T) {
	valid := strategy.AaveClosePositionParams{
		Account:                 common.HexToAddress("0x00000000000000000000000000000000000000aa"),
		Chain:                   config.Base,
		DebtAsset:               config.USDC,
		TemporaryRepayAllowance: decimal.NewFromInt(102),
		CollateralAsset:         config.WETH,
	}

	tests := []struct {
		name    string
		mutate  func(*strategy.AaveClosePositionParams)
		wantErr string
	}{
		{name: "zero account", mutate: func(p *strategy.AaveClosePositionParams) { p.Account = common.Address{} }, wantErr: "account must not be zero"},
		{name: "unsupported chain", mutate: func(p *strategy.AaveClosePositionParams) { p.Chain = config.Chain(999) }, wantErr: "chain"},
		{name: "unsupported debt asset", mutate: func(p *strategy.AaveClosePositionParams) { p.DebtAsset = config.Coin(999) }, wantErr: "debt asset"},
		{name: "zero allowance", mutate: func(p *strategy.AaveClosePositionParams) { p.TemporaryRepayAllowance = decimal.Zero }, wantErr: "temporary repay allowance must be positive"},
		{name: "negative allowance", mutate: func(p *strategy.AaveClosePositionParams) { p.TemporaryRepayAllowance = decimal.NewFromInt(-1) }, wantErr: "temporary repay allowance must be positive"},
		{name: "unsupported collateral asset", mutate: func(p *strategy.AaveClosePositionParams) { p.CollateralAsset = config.Coin(999) }, wantErr: "collateral asset"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := valid
			tt.mutate(&params)

			flow, err := strategy.AaveClosePosition(params)

			if flow != nil {
				t.Fatalf("expected nil flow, got %#v", flow)
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
			}
		})
	}
}
