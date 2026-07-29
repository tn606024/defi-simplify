package aave

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/shopspring/decimal"
	defi "github.com/tn606024/defi-simplify"
	txamount "github.com/tn606024/defi-simplify/amount"
	"github.com/tn606024/defi-simplify/client/contract"
	"github.com/tn606024/defi-simplify/config"
	"github.com/tn606024/defi-simplify/erc20"
	"github.com/tn606024/defi-simplify/helper"
)

var _ = Describe("Aave Flow steps", func() {
	var (
		ctx    context.Context
		user   common.Address
		market Market
		usdc   Reserve
		weth   Reserve
	)

	BeforeEach(func() {
		ctx = context.Background()
		user = common.HexToAddress("0x00000000000000000000000000000000000000aa")
		market, usdc, weth = stepTestReserves()
	})

	It("builds approve, supply, and borrow calls matching the low-level action builders", func() {
		supplyAmount := decimal.RequireFromString("100.5")
		borrowAmount := decimal.RequireFromString("0.01")

		plan, err := defi.NewFlow(user, defi.WithChain(config.Base)).
			Add(erc20.Approve(usdc.Underlying(), PoolSpender(market), txamount.Exact(supplyAmount))).
			Add(Supply(usdc, txamount.Exact(supplyAmount))).
			Add(Borrow(weth, txamount.Exact(borrowAmount))).
			Build(ctx, nil)

		Expect(err).NotTo(HaveOccurred())
		Expect(plan.Calls()).To(Equal([]defi.Call{
			expectedAaveApproveCall(ctx, usdc, supplyAmount),
			expectedSupplyCall(ctx, user, usdc, supplyAmount),
			expectedBorrowCall(ctx, user, weth, borrowAmount),
		}))
		steps := plan.Steps()
		Expect(steps[0].Expectations[0].ExpectationName()).To(Equal("erc20.Approval"))
		Expect(steps[1].Expectations[0].ExpectationName()).To(Equal("aave.Supply"))
		Expect(steps[2].Expectations[0].ExpectationName()).To(Equal("aave.Borrow"))
	})

	It("returns a useful error for an unresolved reserve", func() {
		plan, err := defi.NewFlow(user, defi.WithChain(config.Base)).
			Add(Supply(Reserve{}, txamount.Exact(decimal.NewFromInt(1)))).
			Build(ctx, nil)

		Expect(plan).To(BeNil())
		Expect(err).To(MatchError(ContainSubstring("build flow step 1 aave.Supply")))
		Expect(err).To(MatchError(ContainSubstring("invalid Aave reserve")))
	})

	It("rejects non-positive amounts", func() {
		plan, err := defi.NewFlow(user, defi.WithChain(config.Base)).
			Add(Borrow(weth, txamount.Exact(decimal.Zero))).
			Build(ctx, nil)

		Expect(plan).To(BeNil())
		Expect(err).To(MatchError(ContainSubstring("amount must be positive")))
	})
})

func expectedAaveApproveCall(ctx context.Context, reserve Reserve, amount decimal.Decimal) defi.Call {
	return mustCall(ctx, contract.BuildApproveAction(
		reserve.Underlying().Address(),
		reserve.Market().Pool(),
		reserveAmount(reserve, amount),
	))
}

func expectedSupplyCall(ctx context.Context, user common.Address, reserve Reserve, amount decimal.Decimal) defi.Call {
	return mustCall(ctx, contract.BuildSupplyAction(
		reserve.Market().Pool(),
		reserve.Underlying().Address(),
		reserveAmount(reserve, amount),
		user,
	))
}

func expectedBorrowCall(ctx context.Context, user common.Address, reserve Reserve, amount decimal.Decimal) defi.Call {
	return mustCall(ctx, contract.BuildBorrowAction(
		reserve.Market().Pool(),
		reserve.Underlying().Address(),
		reserveAmount(reserve, amount),
		user,
	))
}

func reserveAmount(reserve Reserve, amount decimal.Decimal) *big.Int {
	return helper.ToWei(amount, reserve.Underlying().Decimals())
}

func mustCall(ctx context.Context, action defi.Action) defi.Call {
	call, err := action.ToCall(ctx, nil, nil)
	Expect(err).NotTo(HaveOccurred())
	Expect(call).NotTo(BeNil())
	return *call
}

func stepTestReserves() (Market, Reserve, Reserve) {
	GinkgoHelper()
	market, err := NewMarket(
		"aave-v3-base",
		config.Base,
		common.HexToAddress("0x1000000000000000000000000000000000000001"),
		common.HexToAddress("0x1000000000000000000000000000000000000002"),
		common.HexToAddress("0x1000000000000000000000000000000000000003"),
	)
	Expect(err).NotTo(HaveOccurred())
	usdc := stepTestReserve(
		market,
		"0x2000000000000000000000000000000000000001",
		"0x2000000000000000000000000000000000000002",
		"0x2000000000000000000000000000000000000003",
		"USDC",
		"USD Coin",
		6,
	)
	weth := stepTestReserve(
		market,
		"0x3000000000000000000000000000000000000001",
		"0x3000000000000000000000000000000000000002",
		"0x3000000000000000000000000000000000000003",
		"WETH",
		"Wrapped Ether",
		18,
	)
	return market, usdc, weth
}

func stepTestReserve(
	market Market,
	underlyingAddress string,
	aTokenAddress string,
	variableDebtAddress string,
	symbol string,
	name string,
	decimals uint8,
) Reserve {
	GinkgoHelper()
	underlying := mustToken(mustTokenRef(underlyingAddress), symbol, name, decimals)
	aToken := mustToken(mustTokenRef(aTokenAddress), "a"+symbol, "Aave "+name, decimals)
	variableDebt := mustToken(
		mustTokenRef(variableDebtAddress),
		"variableDebt"+symbol,
		"Aave Variable Debt "+name,
		decimals,
	)
	reserve, err := NewReserve(market, underlying, aToken, variableDebt, nil)
	Expect(err).NotTo(HaveOccurred())
	return reserve
}
