package aave

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/shopspring/decimal"
	defi "github.com/tn606024/defi-simplify"
	"github.com/tn606024/defi-simplify/client/contract"
	"github.com/tn606024/defi-simplify/config"
)

var _ = Describe("Aave credit-delegation Flow steps", func() {
	It("builds account-owned and relayed delegation calls", func() {
		ctx := context.Background()
		account := common.HexToAddress("0x00000000000000000000000000000000000000aa")
		delegator := common.HexToAddress("0x00000000000000000000000000000000000000bb")
		delegatee := common.HexToAddress("0x00000000000000000000000000000000000000cc")
		amount := decimal.RequireFromString("0.25")
		deadline := big.NewInt(2_000_000_000)
		v := uint8(27)
		var r, s [32]byte
		r[0] = 1
		s[0] = 2
		_, _, weth := stepTestReserves()
		capability := testDelegationCapability(weth, "1")

		plan, err := defi.NewFlow(account, defi.WithChain(config.Base)).
			Add(ApproveDelegation(weth, delegatee, amount)).
			Add(DelegationWithSig(capability, delegator, delegatee, amount, deadline, v, r, s)).
			Build(ctx, nil)

		Expect(err).NotTo(HaveOccurred())
		debtTokenAddress := weth.VariableDebtToken().Address()
		amountWei := reserveAmount(weth, amount)
		Expect(plan.Calls()).To(Equal([]defi.Call{
			mustCall(ctx, contract.BuildApproveDelegationAction(debtTokenAddress, delegatee, amountWei)),
			mustCall(ctx, contract.BuildDelegationWithSigAction(
				debtTokenAddress,
				delegator,
				delegatee,
				amountWei,
				deadline,
				v,
				r,
				s,
			)),
		}))
		Expect(plan.Steps[0].Expectations[0].ExpectationName()).To(Equal("aave.BorrowAllowanceDelegated"))
		Expect(plan.Steps[1].Expectations[0].ExpectationName()).To(Equal("aave.BorrowAllowanceDelegated"))
	})

	It("scales delegation amounts with the underlying asset decimals", func() {
		ctx := context.Background()
		delegatee := common.HexToAddress("0x00000000000000000000000000000000000000cc")
		amount := decimal.RequireFromString("1.25")
		market, usdc, _ := stepTestReserves()
		differentDebtDecimals := mustToken(
			usdc.VariableDebtToken().Ref(),
			usdc.VariableDebtToken().Symbol(),
			usdc.VariableDebtToken().Name(),
			18,
		)
		usdc, err := NewReserve(market, usdc.Underlying(), usdc.AToken(), differentDebtDecimals, nil)
		Expect(err).NotTo(HaveOccurred())
		plan, err := defi.NewFlow(
			common.HexToAddress("0x00000000000000000000000000000000000000aa"),
			defi.WithChain(config.Base),
		).
			Add(ApproveDelegation(usdc, delegatee, amount)).
			Build(ctx, nil)

		Expect(err).NotTo(HaveOccurred())
		debtTokenAddress := usdc.VariableDebtToken().Address()
		amountWei := reserveAmount(usdc, amount)
		Expect(plan.Calls()).To(Equal([]defi.Call{
			mustCall(ctx, contract.BuildApproveDelegationAction(debtTokenAddress, delegatee, amountWei)),
		}))
	})

	It("rejects unresolved delegation reserves", func() {
		plan, err := defi.NewFlow(
			common.HexToAddress("0x00000000000000000000000000000000000000aa"),
			defi.WithChain(config.Base),
		).
			Add(ApproveDelegation(
				Reserve{},
				common.HexToAddress("0x00000000000000000000000000000000000000cc"),
				decimal.NewFromInt(1),
			)).
			Build(context.Background(), nil)

		Expect(plan).To(BeNil())
		Expect(err).To(MatchError(ContainSubstring("invalid Aave reserve")))
	})

	It("allows zero-value delegation to revoke borrow allowance", func() {
		_, _, weth := stepTestReserves()
		plan, err := defi.NewFlow(
			common.HexToAddress("0x00000000000000000000000000000000000000aa"),
			defi.WithChain(config.Base),
		).
			Add(ApproveDelegation(
				weth,
				common.HexToAddress("0x00000000000000000000000000000000000000cc"),
				decimal.Zero,
			)).
			Build(context.Background(), nil)

		Expect(err).NotTo(HaveOccurred())
		Expect(plan.Steps[0].Expectations).To(HaveLen(1))
	})
})
