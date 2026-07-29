package strategy_test

import (
	"github.com/ethereum/go-ethereum/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/shopspring/decimal"
	"github.com/tn606024/defi-simplify/aave"
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
		Entry("invalid supply reserve", func(p *strategy.AaveSupplyBorrowParams) {
			p.SupplyReserve = aave.Reserve{}
		}, "supply reserve"),
		Entry("zero supply amount", func(p *strategy.AaveSupplyBorrowParams) {
			p.SupplyAmount = decimal.Zero
		}, "supply amount must be positive"),
		Entry("negative supply amount", func(p *strategy.AaveSupplyBorrowParams) {
			p.SupplyAmount = decimal.NewFromInt(-1)
		}, "supply amount must be positive"),
		Entry("invalid borrow reserve", func(p *strategy.AaveSupplyBorrowParams) {
			p.BorrowReserve = aave.Reserve{}
		}, "borrow reserve"),
		Entry("reserves from different markets", func(p *strategy.AaveSupplyBorrowParams) {
			p.BorrowReserve = strategyReserveFromOtherMarket()
		}, "must belong to the same Aave market"),
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
		Entry("invalid debt reserve", func(p *strategy.AaveClosePositionParams) {
			p.DebtReserve = aave.Reserve{}
		}, "debt reserve"),
		Entry("zero allowance", func(p *strategy.AaveClosePositionParams) {
			p.TemporaryRepayAllowance = decimal.Zero
		}, "temporary repay allowance must be positive"),
		Entry("negative allowance", func(p *strategy.AaveClosePositionParams) {
			p.TemporaryRepayAllowance = decimal.NewFromInt(-1)
		}, "temporary repay allowance must be positive"),
		Entry("invalid collateral reserve", func(p *strategy.AaveClosePositionParams) {
			p.CollateralReserve = aave.Reserve{}
		}, "collateral reserve"),
		Entry("reserves from different markets", func(p *strategy.AaveClosePositionParams) {
			p.CollateralReserve = strategyReserveFromOtherMarket()
		}, "must belong to the same Aave market"),
	)

	DescribeTable("rejects invalid native supply parameters",
		func(mutate func(*strategy.AaveSupplyNativeETHParams), expectedError string) {
			params := validSupplyNativeETHParams()
			mutate(&params)

			flow, err := strategy.AaveSupplyNativeETH(params)

			Expect(flow).To(BeNil())
			Expect(err).To(MatchError(ContainSubstring(expectedError)))
		},
		Entry("zero account", func(p *strategy.AaveSupplyNativeETHParams) {
			p.Account = common.Address{}
		}, "account must not be zero"),
		Entry("invalid reserve", func(p *strategy.AaveSupplyNativeETHParams) {
			p.Reserve = aave.Reserve{}
		}, "wrapped native reserve"),
		Entry("zero amount", func(p *strategy.AaveSupplyNativeETHParams) {
			p.Amount = decimal.Zero
		}, "supply amount must be positive"),
		Entry("negative amount", func(p *strategy.AaveSupplyNativeETHParams) {
			p.Amount = decimal.NewFromInt(-1)
		}, "supply amount must be positive"),
	)

	DescribeTable("rejects invalid native borrow parameters",
		func(mutate func(*strategy.AaveBorrowNativeETHParams), expectedError string) {
			params := validBorrowNativeETHParams()
			mutate(&params)

			flow, err := strategy.AaveBorrowNativeETH(params)

			Expect(flow).To(BeNil())
			Expect(err).To(MatchError(ContainSubstring(expectedError)))
		},
		Entry("zero account", func(p *strategy.AaveBorrowNativeETHParams) {
			p.Account = common.Address{}
		}, "account must not be zero"),
		Entry("invalid reserve", func(p *strategy.AaveBorrowNativeETHParams) {
			p.Reserve = aave.Reserve{}
		}, "wrapped native reserve"),
		Entry("zero amount", func(p *strategy.AaveBorrowNativeETHParams) {
			p.Amount = decimal.Zero
		}, "borrow amount must be positive"),
		Entry("negative amount", func(p *strategy.AaveBorrowNativeETHParams) {
			p.Amount = decimal.NewFromInt(-1)
		}, "borrow amount must be positive"),
	)

	DescribeTable("rejects invalid exact native withdrawal parameters",
		func(mutate func(*strategy.AaveWithdrawNativeETHParams), expectedError string) {
			params := validWithdrawNativeETHParams()
			mutate(&params)

			flow, err := strategy.AaveWithdrawNativeETH(params)

			Expect(flow).To(BeNil())
			Expect(err).To(MatchError(ContainSubstring(expectedError)))
		},
		Entry("zero account", func(p *strategy.AaveWithdrawNativeETHParams) {
			p.Account = common.Address{}
		}, "account must not be zero"),
		Entry("invalid reserve", func(p *strategy.AaveWithdrawNativeETHParams) {
			p.Reserve = aave.Reserve{}
		}, "wrapped native reserve"),
		Entry("zero amount", func(p *strategy.AaveWithdrawNativeETHParams) {
			p.Amount = decimal.Zero
		}, "withdraw amount must be positive"),
		Entry("negative amount", func(p *strategy.AaveWithdrawNativeETHParams) {
			p.Amount = decimal.NewFromInt(-1)
		}, "withdraw amount must be positive"),
	)

	DescribeTable("rejects invalid full native withdrawal parameters",
		func(mutate func(*strategy.AaveWithdrawAllNativeETHParams), expectedError string) {
			params := validWithdrawAllNativeETHParams()
			mutate(&params)

			flow, err := strategy.AaveWithdrawAllNativeETH(params)

			Expect(flow).To(BeNil())
			Expect(err).To(MatchError(ContainSubstring(expectedError)))
		},
		Entry("zero account", func(p *strategy.AaveWithdrawAllNativeETHParams) {
			p.Account = common.Address{}
		}, "account must not be zero"),
		Entry("invalid reserve", func(p *strategy.AaveWithdrawAllNativeETHParams) {
			p.Reserve = aave.Reserve{}
		}, "wrapped native reserve"),
	)
})

func validSupplyBorrowParams() strategy.AaveSupplyBorrowParams {
	_, usdc, weth := strategyTestReserves()
	return strategy.AaveSupplyBorrowParams{
		Account:       common.HexToAddress("0x00000000000000000000000000000000000000aa"),
		SupplyReserve: usdc,
		SupplyAmount:  decimal.NewFromInt(100),
		BorrowReserve: weth,
		BorrowAmount:  decimal.RequireFromString("0.01"),
	}
}

func validClosePositionParams() strategy.AaveClosePositionParams {
	_, usdc, weth := strategyTestReserves()
	return strategy.AaveClosePositionParams{
		Account:                 common.HexToAddress("0x00000000000000000000000000000000000000aa"),
		DebtReserve:             usdc,
		TemporaryRepayAllowance: decimal.NewFromInt(102),
		CollateralReserve:       weth,
	}
}

func validSupplyNativeETHParams() strategy.AaveSupplyNativeETHParams {
	_, _, weth := strategyTestReserves()
	return strategy.AaveSupplyNativeETHParams{
		Account: common.HexToAddress("0x00000000000000000000000000000000000000aa"),
		Reserve: weth,
		Amount:  decimal.RequireFromString("0.1"),
	}
}

func validBorrowNativeETHParams() strategy.AaveBorrowNativeETHParams {
	_, _, weth := strategyTestReserves()
	return strategy.AaveBorrowNativeETHParams{
		Account: common.HexToAddress("0x00000000000000000000000000000000000000aa"),
		Reserve: weth,
		Amount:  decimal.RequireFromString("0.01"),
	}
}

func validWithdrawNativeETHParams() strategy.AaveWithdrawNativeETHParams {
	_, _, weth := strategyTestReserves()
	return strategy.AaveWithdrawNativeETHParams{
		Account: common.HexToAddress("0x00000000000000000000000000000000000000aa"),
		Reserve: weth,
		Amount:  decimal.RequireFromString("0.05"),
	}
}

func validWithdrawAllNativeETHParams() strategy.AaveWithdrawAllNativeETHParams {
	_, _, weth := strategyTestReserves()
	return strategy.AaveWithdrawAllNativeETHParams{
		Account: common.HexToAddress("0x00000000000000000000000000000000000000aa"),
		Reserve: weth,
	}
}

func strategyReserveFromOtherMarket() aave.Reserve {
	market, err := aave.NewMarket(
		"aave-v3-base-other",
		config.Base,
		common.HexToAddress("0x4000000000000000000000000000000000000001"),
		common.HexToAddress("0x4000000000000000000000000000000000000002"),
		common.HexToAddress("0x4000000000000000000000000000000000000003"),
		common.HexToAddress("0x4000000000000000000000000000000000000004"),
	)
	Expect(err).NotTo(HaveOccurred())
	return strategyTestReserve(market, "5", "WETH", "Wrapped Ether", 18)
}
