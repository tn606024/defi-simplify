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
)

var _ = Describe("Aave Pool write Flow steps", func() {
	It("builds permit, withdraw, and repay calls with semantic expectations", func() {
		ctx := context.Background()
		account := common.HexToAddress("0x00000000000000000000000000000000000000aa")
		amount := decimal.RequireFromString("12.5")
		deadline := big.NewInt(2_000_000_000)
		v := uint8(28)
		var r, s [32]byte
		r[0] = 1
		s[0] = 2

		_, usdc, _ := stepTestReserves()
		permit := testPermitCapability(usdc.Underlying(), "2")
		plan, err := defi.NewFlow(account, defi.WithChain(config.Base)).
			Add(SupplyWithPermit(usdc, permit, txamount.Exact(amount), deadline, v, r, s)).
			Add(Withdraw(usdc, txamount.Exact(amount))).
			Add(Repay(usdc, txamount.Exact(amount))).
			Add(RepayWithPermit(usdc, permit, txamount.Exact(amount), deadline, v, r, s)).
			Build(ctx, nil)

		Expect(err).NotTo(HaveOccurred())
		pool := usdc.Market().Pool()
		asset := usdc.Underlying().Address()
		amountWei := reserveAmount(usdc, amount)
		Expect(plan.Calls()).To(Equal([]defi.Call{
			mustCall(ctx, contract.BuildSupplyWithPermitAction(pool, asset, amountWei, account, 0, deadline, v, r, s)),
			mustCall(ctx, contract.BuildWithdrawAction(pool, asset, amountWei, account)),
			mustCall(ctx, contract.BuildRepayAction(pool, asset, amountWei, account)),
			mustCall(ctx, contract.BuildRepayWithPermitAction(pool, asset, amountWei, account, deadline, v, r, s)),
		}))
		steps := plan.Steps()
		Expect(steps).To(HaveLen(4))
		Expect(steps[0].Expectations[0].ExpectationName()).To(Equal("aave.Supply"))
		Expect(steps[1].Expectations[0].ExpectationName()).To(Equal("aave.Withdraw"))
		Expect(steps[2].Expectations[0].ExpectationName()).To(Equal("aave.Repay"))
		Expect(steps[3].Expectations[0].ExpectationName()).To(Equal("aave.Repay"))
	})

	It("builds full-position calls with the uint256.max sentinel", func() {
		ctx := context.Background()
		account := common.HexToAddress("0x00000000000000000000000000000000000000aa")
		maxAmount := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))
		_, usdc, _ := stepTestReserves()

		plan, err := defi.NewFlow(account, defi.WithChain(config.Base)).
			Add(RepayAll(usdc)).
			Add(WithdrawAll(usdc)).
			Build(ctx, nil)

		Expect(err).NotTo(HaveOccurred())
		pool := usdc.Market().Pool()
		asset := usdc.Underlying().Address()
		Expect(plan.Calls()).To(Equal([]defi.Call{
			mustCall(ctx, contract.BuildRepayAction(pool, asset, maxAmount, account)),
			mustCall(ctx, contract.BuildWithdrawAction(pool, asset, maxAmount, account)),
		}))
		steps := plan.Steps()
		Expect(steps[0].Name).To(Equal("aave.RepayAll"))
		Expect(steps[0].Expectations[0].ExpectationName()).To(Equal("aave.Repay"))
		Expect(steps[1].Name).To(Equal("aave.WithdrawAll"))
		Expect(steps[1].Expectations[0].ExpectationName()).To(Equal("aave.Withdraw"))
	})

	It("rejects missing permit signature deadlines", func() {
		_, usdc, _ := stepTestReserves()
		permit := testPermitCapability(usdc.Underlying(), "2")
		plan, err := defi.NewFlow(
			common.HexToAddress("0x00000000000000000000000000000000000000aa"),
			defi.WithChain(config.Base),
		).
			Add(SupplyWithPermit(usdc, permit, txamount.Exact(decimal.NewFromInt(1)), nil, 0, [32]byte{}, [32]byte{})).
			Build(context.Background(), nil)

		Expect(plan).To(BeNil())
		Expect(err).To(MatchError(ContainSubstring("signature deadline is nil")))
	})

	It("rejects a permit capability for a different asset", func() {
		var r, s [32]byte
		r[0] = 1
		s[0] = 2

		_, usdc, weth := stepTestReserves()
		wethPermit := testPermitCapability(weth.Underlying(), "1")
		plan, err := defi.NewFlow(
			common.HexToAddress("0x00000000000000000000000000000000000000aa"),
			defi.WithChain(config.Base),
		).
			Add(SupplyWithPermit(
				usdc,
				wethPermit,
				txamount.Exact(decimal.NewFromInt(1)),
				big.NewInt(2_000_000_000),
				27,
				r,
				s,
			)).
			Build(context.Background(), nil)

		Expect(plan).To(BeNil())
		Expect(err).To(MatchError(ContainSubstring("does not match expected token")))
	})

	DescribeTable(
		"compiles runtime balances into supported Aave Pool amount words",
		func(kind poolStepKind) {
			account := common.HexToAddress("0x00000000000000000000000000000000000000aa")
			_, usdc, _ := stepTestReserves()
			source := txamount.CurrentBalance(usdc.Underlying().Ref())
			var flowStep defi.FlowStep
			switch kind {
			case supplyStep:
				flowStep = Supply(usdc, source)
			case borrowStep:
				flowStep = Borrow(usdc, source)
			case withdrawStep:
				flowStep = Withdraw(usdc, source)
			case repayStep:
				flowStep = Repay(usdc, source)
			default:
				Fail("unsupported test step")
			}

			plan, err := defi.NewFlow(account, defi.WithChain(config.Base)).
				Add(flowStep).
				Build(context.Background(), nil)

			Expect(err).NotTo(HaveOccurred())
			Expect(plan.Kind()).To(Equal(defi.PlanDynamic))
			calls, err := plan.DynamicCalls()
			Expect(err).NotTo(HaveOccurred())
			Expect(calls).To(HaveLen(1))
			Expect(calls[0].Patches).To(Equal([]defi.BalancePatch{{
				Token:  usdc.Underlying().Address(),
				Offset: poolAmountOffset,
				BPS:    txamount.FullBPS,
				Source: defi.BalanceSourceCurrentBalance,
			}}))
		},
		Entry("supply", supplyStep),
		Entry("borrow", borrowStep),
		Entry("withdraw", withdrawStep),
		Entry("repay", repayStep),
	)

	It("rejects runtime sources for permit-backed Pool calls", func() {
		account := common.HexToAddress("0x00000000000000000000000000000000000000aa")
		_, usdc, _ := stepTestReserves()
		permit := testPermitCapability(usdc.Underlying(), "2")
		plan, err := defi.NewFlow(account, defi.WithChain(config.Base)).
			Add(SupplyWithPermit(
				usdc,
				permit,
				txamount.CurrentBalance(usdc.Underlying().Ref()),
				big.NewInt(2_000_000_000),
				27,
				[32]byte{1},
				[32]byte{2},
			)).
			Build(context.Background(), nil)

		Expect(plan).To(BeNil())
		Expect(err).To(MatchError(ContainSubstring("runtime amount source is not supported")))
	})
})
