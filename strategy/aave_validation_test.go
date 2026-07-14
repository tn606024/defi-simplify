package strategy_test

import (
	"github.com/ethereum/go-ethereum/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/shopspring/decimal"
	"github.com/tn606024/defi-simplify/config"
	"github.com/tn606024/defi-simplify/strategy"
)

var _ = Describe("Aave strategy validation", func() {
	DescribeTable("rejects invalid supply-and-borrow parameters",
		func(mutate func(*strategy.AaveSupplyBorrowParams), expectedError string) {
			params := validSupplyBorrowParams()
			mutate(&params)

			flow, err := strategy.AaveSupplyBorrow(params)

			Expect(flow).To(BeNil())
			Expect(err).To(MatchError(ContainSubstring(expectedError)))
		},
		Entry("zero account", func(p *strategy.AaveSupplyBorrowParams) {
			p.Account = common.Address{}
		}, "account must not be zero"),
		Entry("unsupported chain", func(p *strategy.AaveSupplyBorrowParams) {
			p.Chain = config.Chain(999)
		}, "chain"),
		Entry("unsupported supply asset", func(p *strategy.AaveSupplyBorrowParams) {
			p.SupplyAsset = config.Coin(999)
		}, "supply asset"),
		Entry("zero supply amount", func(p *strategy.AaveSupplyBorrowParams) {
			p.SupplyAmount = decimal.Zero
		}, "supply amount must be positive"),
		Entry("negative supply amount", func(p *strategy.AaveSupplyBorrowParams) {
			p.SupplyAmount = decimal.NewFromInt(-1)
		}, "supply amount must be positive"),
		Entry("unsupported borrow asset", func(p *strategy.AaveSupplyBorrowParams) {
			p.BorrowAsset = config.Coin(999)
		}, "borrow asset"),
		Entry("zero borrow amount", func(p *strategy.AaveSupplyBorrowParams) {
			p.BorrowAmount = decimal.Zero
		}, "borrow amount must be positive"),
		Entry("negative borrow amount", func(p *strategy.AaveSupplyBorrowParams) {
			p.BorrowAmount = decimal.NewFromInt(-1)
		}, "borrow amount must be positive"),
	)

	DescribeTable("rejects invalid close-position parameters",
		func(mutate func(*strategy.AaveClosePositionParams), expectedError string) {
			params := validClosePositionParams()
			mutate(&params)

			flow, err := strategy.AaveClosePosition(params)

			Expect(flow).To(BeNil())
			Expect(err).To(MatchError(ContainSubstring(expectedError)))
		},
		Entry("zero account", func(p *strategy.AaveClosePositionParams) {
			p.Account = common.Address{}
		}, "account must not be zero"),
		Entry("unsupported chain", func(p *strategy.AaveClosePositionParams) {
			p.Chain = config.Chain(999)
		}, "chain"),
		Entry("unsupported debt asset", func(p *strategy.AaveClosePositionParams) {
			p.DebtAsset = config.Coin(999)
		}, "debt asset"),
		Entry("zero allowance", func(p *strategy.AaveClosePositionParams) {
			p.TemporaryRepayAllowance = decimal.Zero
		}, "temporary repay allowance must be positive"),
		Entry("negative allowance", func(p *strategy.AaveClosePositionParams) {
			p.TemporaryRepayAllowance = decimal.NewFromInt(-1)
		}, "temporary repay allowance must be positive"),
		Entry("unsupported collateral asset", func(p *strategy.AaveClosePositionParams) {
			p.CollateralAsset = config.Coin(999)
		}, "collateral asset"),
	)
})

func validSupplyBorrowParams() strategy.AaveSupplyBorrowParams {
	return strategy.AaveSupplyBorrowParams{
		Account:      common.HexToAddress("0x00000000000000000000000000000000000000aa"),
		Chain:        config.Base,
		SupplyAsset:  config.USDC,
		SupplyAmount: decimal.NewFromInt(100),
		BorrowAsset:  config.WETH,
		BorrowAmount: decimal.RequireFromString("0.01"),
	}
}

func validClosePositionParams() strategy.AaveClosePositionParams {
	return strategy.AaveClosePositionParams{
		Account:                 common.HexToAddress("0x00000000000000000000000000000000000000aa"),
		Chain:                   config.Base,
		DebtAsset:               config.USDC,
		TemporaryRepayAllowance: decimal.NewFromInt(102),
		CollateralAsset:         config.WETH,
	}
}
