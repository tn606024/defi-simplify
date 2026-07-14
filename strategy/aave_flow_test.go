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
		flow, err := strategy.AaveSupplyBorrow(strategy.AaveSupplyBorrowParams{
			Account:      account,
			Chain:        config.Base,
			SupplyAsset:  config.USDC,
			SupplyAmount: supplyAmount,
			BorrowAsset:  config.WETH,
			BorrowAmount: borrowAmount,
		})
		Expect(err).NotTo(HaveOccurred())

		actual, err := flow.Build(ctx, nil)
		Expect(err).NotTo(HaveOccurred())
		expected, err := defi.NewFlow(account, defi.WithChain(config.Base)).
			Add(aave.ApproveSupply(config.USDC, supplyAmount)).
			Add(aave.Supply(config.USDC, supplyAmount)).
			Add(aave.Borrow(config.WETH, borrowAmount)).
			Build(ctx, nil)
		Expect(err).NotTo(HaveOccurred())

		expectEquivalentPlans(actual, expected)
	})

	It("builds a single-reserve close exactly like manual Flow composition", func() {
		allowance := decimal.NewFromInt(102)
		flow, err := strategy.AaveClosePosition(strategy.AaveClosePositionParams{
			Account:                 account,
			Chain:                   config.Base,
			DebtAsset:               config.USDC,
			TemporaryRepayAllowance: allowance,
			CollateralAsset:         config.WETH,
		})
		Expect(err).NotTo(HaveOccurred())

		actual, err := flow.Build(ctx, nil)
		Expect(err).NotTo(HaveOccurred())
		expected, err := defi.NewFlow(account, defi.WithChain(config.Base)).
			Add(erc20.Approve(config.USDC, aave.PoolSpender(), allowance)).
			Add(aave.RepayAll(config.USDC)).
			Add(erc20.Approve(config.USDC, aave.PoolSpender(), decimal.Zero)).
			Add(aave.WithdrawAll(config.WETH)).
			Build(ctx, nil)
		Expect(err).NotTo(HaveOccurred())

		expectEquivalentPlans(actual, expected)
		Expect(actual.Steps).To(HaveLen(4))
		Expect(actual.Steps[2].Name).To(Equal("erc20.Approve"))
		pool, err := config.Base.AaveV3PoolAddress()
		Expect(err).NotTo(HaveOccurred())
		debtAsset, err := config.USDC.Address(config.Base)
		Expect(err).NotTo(HaveOccurred())
		zeroApproval, err := contract.BuildApproveAction(debtAsset, pool, big.NewInt(0)).ToCall(ctx, nil, nil)
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
