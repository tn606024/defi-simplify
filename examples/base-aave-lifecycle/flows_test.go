package main

import (
	"bytes"
	"context"

	"github.com/ethereum/go-ethereum/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/shopspring/decimal"
	defi "github.com/tn606024/defi-simplify"
	"github.com/tn606024/defi-simplify/aave"
	"github.com/tn606024/defi-simplify/assets/base"
	"github.com/tn606024/defi-simplify/config"
	"github.com/tn606024/defi-simplify/token"
)

var _ = Describe("Base Aave lifecycle flows", func() {
	var (
		account     common.Address
		market      aave.Market
		wethReserve aave.Reserve
		usdcReserve aave.Reserve
	)

	BeforeEach(func() {
		account = common.HexToAddress("0x1000000000000000000000000000000000000001")
		market, wethReserve, usdcReserve = lifecycleTestReserves()
	})

	It("builds the open phase as a static ordered Flow", func() {
		flow := buildOpenFlow(
			account,
			market,
			wethReserve,
			usdcReserve,
			decimal.RequireFromString("0.1"),
			decimal.NewFromInt(10),
		)

		plan, err := flow.Build(context.Background(), nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(plan.Kind()).To(Equal(defi.PlanStatic))
		Expect(plan.Account()).To(Equal(account))
		Expect(stepNames(plan)).To(Equal([]string{
			"weth.Wrap",
			"erc20.Approve",
			"aave.Supply",
			"aave.Borrow",
		}))
	})

	It("builds cleanup as a dynamic Flow scoped to the withdrawn WETH delta", func() {
		flow := buildCloseFlow(
			account,
			market,
			wethReserve,
			usdcReserve,
			decimal.RequireFromString("10.1"),
		)

		plan, err := flow.Build(context.Background(), nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(plan.Kind()).To(Equal(defi.PlanDynamic))
		Expect(stepNames(plan)).To(Equal([]string{
			"erc20.Approve",
			"aave.RepayAll",
			"erc20.Approve",
			"aave.WithdrawAll",
			"weth.Unwrap",
		}))
		calls, err := plan.DynamicCalls()
		Expect(err).NotTo(HaveOccurred())
		Expect(calls[3].CheckpointsBefore).To(HaveLen(1))
		Expect(calls[3].CheckpointsBefore[0].Token).To(Equal(base.WETH.Address()))
		Expect(calls[4].Patches).To(HaveLen(1))
		Expect(calls[4].Patches[0].Token).To(Equal(base.WETH.Address()))
		Expect(calls[4].Patches[0].Source).To(Equal(defi.BalanceSourceCheckpointDelta))
		Expect(calls[4].Patches[0].CheckpointID).To(Equal(calls[3].CheckpointsBefore[0].ID))
	})

	It("prints reviewable plan targets, values, selectors, checkpoints, and patches", func() {
		flow := buildCloseFlow(
			account,
			market,
			wethReserve,
			usdcReserve,
			decimal.RequireFromString("10.1"),
		)
		plan, err := flow.Build(context.Background(), nil)
		Expect(err).NotTo(HaveOccurred())
		var output bytes.Buffer
		app := &application{out: &output}

		app.printPlan(plan)

		Expect(output.String()).To(ContainSubstring("plan account=" + account.Hex() + " kind=dynamic"))
		Expect(output.String()).To(ContainSubstring("plan.step id=aave.RepayAll#1"))
		Expect(output.String()).To(ContainSubstring("target=" + market.Pool().Hex()))
		Expect(output.String()).To(ContainSubstring("selector=0x"))
		Expect(output.String()).To(ContainSubstring("label=before-close-withdraw"))
		Expect(output.String()).To(ContainSubstring("source=checkpoint_delta"))
		Expect(output.String()).To(ContainSubstring("token=" + base.WETH.Address().Hex()))
	})
})

func stepNames(plan *defi.ExecutionPlan) []string {
	steps := plan.Steps()
	names := make([]string, len(steps))
	for i, step := range steps {
		names[i] = step.Name
	}
	return names
}

func lifecycleTestReserves() (aave.Market, aave.Reserve, aave.Reserve) {
	GinkgoHelper()
	market, err := aave.NewMarket(
		"aave-v3-base",
		config.Base,
		common.HexToAddress("0x2000000000000000000000000000000000000001"),
		common.HexToAddress("0x2000000000000000000000000000000000000002"),
		common.HexToAddress("0x2000000000000000000000000000000000000003"),
	)
	Expect(err).NotTo(HaveOccurred())
	wethReserve := lifecycleTestReserve(
		market,
		mustLifecycleToken(base.WETH, "WETH", "Wrapped Ether", 18),
		"0x3000000000000000000000000000000000000001",
		"0x3000000000000000000000000000000000000002",
	)
	usdcReserve := lifecycleTestReserve(
		market,
		mustLifecycleToken(base.USDC, "USDC", "USD Coin", 6),
		"0x4000000000000000000000000000000000000001",
		"0x4000000000000000000000000000000000000002",
	)
	return market, wethReserve, usdcReserve
}

func lifecycleTestReserve(
	market aave.Market,
	underlying token.Token,
	aTokenAddress string,
	debtTokenAddress string,
) aave.Reserve {
	GinkgoHelper()
	aToken := mustLifecycleToken(
		mustLifecycleRef(aTokenAddress),
		"a"+underlying.Symbol(),
		"Aave "+underlying.Name(),
		underlying.Decimals(),
	)
	debtToken := mustLifecycleToken(
		mustLifecycleRef(debtTokenAddress),
		"variableDebt"+underlying.Symbol(),
		"Aave Variable Debt "+underlying.Name(),
		underlying.Decimals(),
	)
	reserve, err := aave.NewReserve(market, underlying, aToken, debtToken, nil)
	Expect(err).NotTo(HaveOccurred())
	return reserve
}

func mustLifecycleRef(address string) token.Ref {
	ref, err := token.NewRef(config.Base, common.HexToAddress(address))
	Expect(err).NotTo(HaveOccurred())
	return ref
}

func mustLifecycleToken(ref token.Ref, symbol, name string, decimals uint8) token.Token {
	resolved, err := token.New(ref, symbol, name, decimals)
	Expect(err).NotTo(HaveOccurred())
	return resolved
}
