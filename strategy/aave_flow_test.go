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
	"github.com/tn606024/defi-simplify/client/contract"
	"github.com/tn606024/defi-simplify/config"
	"github.com/tn606024/defi-simplify/erc20"
	"github.com/tn606024/defi-simplify/strategy"
	"github.com/tn606024/defi-simplify/token"
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
			Add(aave.ApproveSupply(usdc, supplyAmount)).
			Add(aave.Supply(usdc, supplyAmount)).
			Add(aave.Borrow(weth, borrowAmount)).
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
			Add(erc20.Approve(usdc.Underlying(), aave.PoolSpender(market), allowance)).
			Add(aave.RepayAll(usdc)).
			Add(erc20.Approve(usdc.Underlying(), aave.PoolSpender(market), decimal.Zero)).
			Add(aave.WithdrawAll(weth)).
			Build(ctx, nil)
		Expect(err).NotTo(HaveOccurred())

		expectEquivalentPlans(actual, expected)
		Expect(actual.Steps).To(HaveLen(4))
		Expect(actual.Steps[2].Name).To(Equal("erc20.Approve"))
		zeroApproval, err := contract.BuildApproveAction(
			usdc.Underlying().Address(),
			market.Pool(),
			big.NewInt(0),
		).ToCall(ctx, nil, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(actual.Steps[2].Calls).To(Equal([]defi.Call{*zeroApproval}))
	})
})

func expectEquivalentPlans(actual, expected *defi.ExecutionPlan) {
	GinkgoHelper()
	Expect(actual.Account).To(Equal(expected.Account))
	Expect(actual.Calls()).To(Equal(expected.Calls()))
	Expect(actual.Steps).To(HaveLen(len(expected.Steps)))
	for i := range expected.Steps {
		Expect(actual.Steps[i].ID).To(Equal(expected.Steps[i].ID))
		Expect(actual.Steps[i].Name).To(Equal(expected.Steps[i].Name))
		Expect(actual.Steps[i].Expectations).To(Equal(expected.Steps[i].Expectations))
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
