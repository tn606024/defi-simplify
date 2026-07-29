package strategy_test

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/shopspring/decimal"
	defi "github.com/tn606024/defi-simplify"
	"github.com/tn606024/defi-simplify/aave"
	txamount "github.com/tn606024/defi-simplify/amount"
	"github.com/tn606024/defi-simplify/client/contract"
	"github.com/tn606024/defi-simplify/config"
	"github.com/tn606024/defi-simplify/erc20"
	"github.com/tn606024/defi-simplify/strategy"
	"github.com/tn606024/defi-simplify/token"
	sdkweth "github.com/tn606024/defi-simplify/weth"
)

var _ = Describe("Aave strategies", func() {
	var (
		ctx     context.Context
		account common.Address
	)

	BeforeEach(func() {
		ctx = context.Background()
		account = common.HexToAddress("0x00000000000000000000000000000000000000aa")
	})

	It("builds supply and borrow exactly like manual Flow composition", func() {
		supplyAmount := decimal.NewFromInt(100)
		borrowAmount := decimal.RequireFromString("0.01")
		market, usdc, weth := strategyTestReserves()
		flow, err := strategy.AaveSupplyBorrow(strategy.AaveSupplyBorrowParams{
			Account:       account,
			SupplyReserve: usdc,
			SupplyAmount:  supplyAmount,
			BorrowReserve: weth,
			BorrowAmount:  borrowAmount,
		})
		Expect(err).NotTo(HaveOccurred())

		actual, err := flow.Build(ctx, nil)
		Expect(err).NotTo(HaveOccurred())
		expected, err := defi.NewFlow(account, defi.WithChain(market.Chain())).
			Add(aave.ApproveSupply(usdc, txamount.Exact(supplyAmount))).
			Add(aave.Supply(usdc, txamount.Exact(supplyAmount))).
			Add(aave.Borrow(weth, txamount.Exact(borrowAmount))).
			Build(ctx, nil)
		Expect(err).NotTo(HaveOccurred())

		expectEquivalentPlans(actual, expected)
	})

	It("builds a single-reserve close exactly like manual Flow composition", func() {
		allowance := decimal.NewFromInt(102)
		market, usdc, weth := strategyTestReserves()
		flow, err := strategy.AaveClosePosition(strategy.AaveClosePositionParams{
			Account:                 account,
			DebtReserve:             usdc,
			TemporaryRepayAllowance: allowance,
			CollateralReserve:       weth,
		})
		Expect(err).NotTo(HaveOccurred())

		actual, err := flow.Build(ctx, nil)
		Expect(err).NotTo(HaveOccurred())
		expected, err := defi.NewFlow(account, defi.WithChain(market.Chain())).
			Add(erc20.Approve(usdc.Underlying(), aave.PoolSpender(market), txamount.Exact(allowance))).
			Add(aave.RepayAll(usdc)).
			Add(erc20.Approve(usdc.Underlying(), aave.PoolSpender(market), txamount.Exact(decimal.Zero))).
			Add(aave.WithdrawAll(weth)).
			Build(ctx, nil)
		Expect(err).NotTo(HaveOccurred())

		expectEquivalentPlans(actual, expected)
		steps := actual.Steps()
		Expect(steps).To(HaveLen(4))
		Expect(steps[2].Name).To(Equal("erc20.Approve"))
		zeroApproval, err := contract.BuildApproveAction(
			usdc.Underlying().Address(),
			market.Pool(),
			big.NewInt(0),
		).ToCall(ctx, nil, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(steps[2].Calls).To(Equal([]defi.PlannedCall{{Call: *zeroApproval}}))
	})

	It("builds native ETH supply exactly like manual Flow composition", func() {
		supplyAmount := decimal.RequireFromString("0.25")
		market, _, weth := strategyTestReserves()
		flow, err := strategy.AaveSupplyNativeETH(strategy.AaveSupplyNativeETHParams{
			Account: account,
			Reserve: weth,
			Amount:  supplyAmount,
		})
		Expect(err).NotTo(HaveOccurred())

		actual, err := flow.Build(ctx, nil)
		Expect(err).NotTo(HaveOccurred())
		exact := txamount.Exact(supplyAmount)
		expected, err := defi.NewFlow(account, defi.WithChain(market.Chain())).
			Add(sdkweth.Wrap(weth.Underlying().Ref(), exact)).
			Add(erc20.Approve(weth.Underlying(), aave.PoolSpender(market), exact)).
			Add(aave.Supply(weth, exact)).
			Build(ctx, nil)
		Expect(err).NotTo(HaveOccurred())

		Expect(actual.Kind()).To(Equal(defi.PlanStatic))
		expectEquivalentPlans(actual, expected)
	})

	It("builds native ETH borrow exactly like manual checkpoint composition", func() {
		borrowAmount := decimal.RequireFromString("0.01")
		market, _, weth := strategyTestReserves()
		flow, err := strategy.AaveBorrowNativeETH(strategy.AaveBorrowNativeETHParams{
			Account: account,
			Reserve: weth,
			Amount:  borrowAmount,
		})
		Expect(err).NotTo(HaveOccurred())

		actual, err := flow.Build(ctx, nil)
		Expect(err).NotTo(HaveOccurred())
		checkpoint := txamount.Checkpoint("before-native-borrow", weth.Underlying().Ref())
		expected, err := defi.NewFlow(account, defi.WithChain(market.Chain())).
			Add(defi.CheckpointBefore(
				aave.Borrow(weth, txamount.Exact(borrowAmount)),
				checkpoint,
			)).
			Add(sdkweth.Unwrap(weth.Underlying().Ref(), txamount.CheckpointDelta(checkpoint))).
			Build(ctx, nil)
		Expect(err).NotTo(HaveOccurred())

		Expect(actual.Kind()).To(Equal(defi.PlanDynamic))
		expectEquivalentPlans(actual, expected)
	})

	It("builds exact native ETH withdrawal exactly like manual checkpoint composition", func() {
		withdrawAmount := decimal.RequireFromString("0.05")
		market, _, weth := strategyTestReserves()
		flow, err := strategy.AaveWithdrawNativeETH(strategy.AaveWithdrawNativeETHParams{
			Account: account,
			Reserve: weth,
			Amount:  withdrawAmount,
		})
		Expect(err).NotTo(HaveOccurred())

		actual, err := flow.Build(ctx, nil)
		Expect(err).NotTo(HaveOccurred())
		checkpoint := txamount.Checkpoint("before-native-withdraw", weth.Underlying().Ref())
		expected, err := defi.NewFlow(account, defi.WithChain(market.Chain())).
			Add(defi.CheckpointBefore(
				aave.Withdraw(weth, txamount.Exact(withdrawAmount)),
				checkpoint,
			)).
			Add(sdkweth.Unwrap(weth.Underlying().Ref(), txamount.CheckpointDelta(checkpoint))).
			Build(ctx, nil)
		Expect(err).NotTo(HaveOccurred())

		Expect(actual.Kind()).To(Equal(defi.PlanDynamic))
		expectEquivalentPlans(actual, expected)
	})

	It("builds full native ETH withdrawal exactly like manual checkpoint composition", func() {
		market, _, weth := strategyTestReserves()
		flow, err := strategy.AaveWithdrawAllNativeETH(strategy.AaveWithdrawAllNativeETHParams{
			Account: account,
			Reserve: weth,
		})
		Expect(err).NotTo(HaveOccurred())

		actual, err := flow.Build(ctx, nil)
		Expect(err).NotTo(HaveOccurred())
		checkpoint := txamount.Checkpoint("before-native-withdraw", weth.Underlying().Ref())
		expected, err := defi.NewFlow(account, defi.WithChain(market.Chain())).
			Add(defi.CheckpointBefore(aave.WithdrawAll(weth), checkpoint)).
			Add(sdkweth.Unwrap(weth.Underlying().Ref(), txamount.CheckpointDelta(checkpoint))).
			Build(ctx, nil)
		Expect(err).NotTo(HaveOccurred())

		Expect(actual.Kind()).To(Equal(defi.PlanDynamic))
		expectEquivalentPlans(actual, expected)
	})
})

func expectEquivalentPlans(actual, expected *defi.ExecutionPlan) {
	GinkgoHelper()
	Expect(actual.Account()).To(Equal(expected.Account()))
	Expect(actual.Kind()).To(Equal(expected.Kind()))
	switch actual.Kind() {
	case defi.PlanStatic:
		actualCalls, err := actual.StaticCalls()
		Expect(err).NotTo(HaveOccurred())
		expectedCalls, err := expected.StaticCalls()
		Expect(err).NotTo(HaveOccurred())
		Expect(actualCalls).To(Equal(expectedCalls))
	case defi.PlanDynamic:
		actualCalls, err := actual.DynamicCalls()
		Expect(err).NotTo(HaveOccurred())
		expectedCalls, err := expected.DynamicCalls()
		Expect(err).NotTo(HaveOccurred())
		Expect(actualCalls).To(Equal(expectedCalls))
	default:
		Fail("unexpected execution plan kind")
	}
	actualSteps := actual.Steps()
	expectedSteps := expected.Steps()
	Expect(actualSteps).To(HaveLen(len(expectedSteps)))
	for i := range expectedSteps {
		Expect(actualSteps[i].ID).To(Equal(expectedSteps[i].ID))
		Expect(actualSteps[i].Name).To(Equal(expectedSteps[i].Name))
		Expect(actualSteps[i].Expectations).To(Equal(expectedSteps[i].Expectations))
	}
}

func strategyTestReserves() (aave.Market, aave.Reserve, aave.Reserve) {
	GinkgoHelper()
	market, err := aave.NewMarket(
		"aave-v3-base",
		config.Base,
		common.HexToAddress("0x1000000000000000000000000000000000000001"),
		common.HexToAddress("0x1000000000000000000000000000000000000002"),
		common.HexToAddress("0x1000000000000000000000000000000000000003"),
		common.HexToAddress("0x1000000000000000000000000000000000000004"),
	)
	Expect(err).NotTo(HaveOccurred())
	return market,
		strategyTestReserve(market, "2", "USDC", "USD Coin", 6),
		strategyTestReserve(market, "3", "WETH", "Wrapped Ether", 18)
}

func strategyTestReserve(
	market aave.Market,
	prefix string,
	symbol string,
	name string,
	decimals uint8,
) aave.Reserve {
	GinkgoHelper()
	underlying := strategyTestToken("0x"+prefix+"000000000000000000000000000000000000001", symbol, name, decimals)
	aToken := strategyTestToken("0x"+prefix+"000000000000000000000000000000000000002", "a"+symbol, "Aave "+name, decimals)
	variableDebt := strategyTestToken(
		"0x"+prefix+"000000000000000000000000000000000000003",
		"variableDebt"+symbol,
		"Aave Variable Debt "+name,
		decimals,
	)
	reserve, err := aave.NewReserve(market, underlying, aToken, variableDebt, nil)
	Expect(err).NotTo(HaveOccurred())
	return reserve
}

func strategyTestToken(address, symbol, name string, decimals uint8) token.Token {
	GinkgoHelper()
	ref, err := token.NewRef(config.Base, common.HexToAddress(address))
	Expect(err).NotTo(HaveOccurred())
	asset, err := token.New(ref, symbol, name, decimals)
	Expect(err).NotTo(HaveOccurred())
	return asset
}
