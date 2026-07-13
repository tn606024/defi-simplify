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

		plan, err := defi.NewFlow(account, defi.WithChain(config.Base)).
			Add(ApproveDelegation(config.WETH, delegatee, amount)).
			Add(DelegationWithSig(config.WETH, delegator, delegatee, amount, deadline, v, r, s)).
			Build(ctx, nil)

		Expect(err).NotTo(HaveOccurred())
		debtToken, err := config.WETH.DebtToken()
		Expect(err).NotTo(HaveOccurred())
		debtTokenAddress, err := debtToken.Address(config.Base)
		Expect(err).NotTo(HaveOccurred())
		_, amountWei := coinAddressAndAmount(config.WETH, amount)
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
		originalDebtDecimals := config.CoinDecimals[config.AVDUSDC]
		config.CoinDecimals[config.AVDUSDC] = 18
		DeferCleanup(func() {
			config.CoinDecimals[config.AVDUSDC] = originalDebtDecimals
		})

		ctx := context.Background()
		delegatee := common.HexToAddress("0x00000000000000000000000000000000000000cc")
		amount := decimal.RequireFromString("1.25")
		plan, err := defi.NewFlow(
			common.HexToAddress("0x00000000000000000000000000000000000000aa"),
			defi.WithChain(config.Base),
		).
			Add(ApproveDelegation(config.USDC, delegatee, amount)).
			Build(ctx, nil)

		Expect(err).NotTo(HaveOccurred())
		debtToken, err := config.USDC.DebtToken()
		Expect(err).NotTo(HaveOccurred())
		debtTokenAddress, err := debtToken.Address(config.Base)
		Expect(err).NotTo(HaveOccurred())
		_, amountWei := coinAddressAndAmount(config.USDC, amount)
		Expect(plan.Calls()).To(Equal([]defi.Call{
			mustCall(ctx, contract.BuildApproveDelegationAction(debtTokenAddress, delegatee, amountWei)),
		}))
	})

	It("rejects debt tokens passed as the underlying delegation asset", func() {
		plan, err := defi.NewFlow(
			common.HexToAddress("0x00000000000000000000000000000000000000aa"),
			defi.WithChain(config.Base),
		).
			Add(ApproveDelegation(
				config.AVDWETH,
				common.HexToAddress("0x00000000000000000000000000000000000000cc"),
				decimal.NewFromInt(1),
			)).
			Build(context.Background(), nil)

		Expect(plan).To(BeNil())
		Expect(err).To(MatchError(ContainSubstring("underlying reserve asset")))
	})

	It("allows zero-value delegation to revoke borrow allowance", func() {
		plan, err := defi.NewFlow(
			common.HexToAddress("0x00000000000000000000000000000000000000aa"),
			defi.WithChain(config.Base),
		).
			Add(ApproveDelegation(
				config.WETH,
				common.HexToAddress("0x00000000000000000000000000000000000000cc"),
				decimal.Zero,
			)).
			Build(context.Background(), nil)

		Expect(err).NotTo(HaveOccurred())
		Expect(plan.Steps[0].Expectations).To(HaveLen(1))
	})
})
