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
			Add(SupplyWithPermit(usdc, permit, amount, deadline, v, r, s)).
			Add(Withdraw(usdc, amount)).
			Add(Repay(usdc, amount)).
			Add(RepayWithPermit(usdc, permit, amount, deadline, v, r, s)).
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
		Expect(plan.Steps).To(HaveLen(4))
		Expect(plan.Steps[0].Expectations[0].ExpectationName()).To(Equal("aave.Supply"))
		Expect(plan.Steps[1].Expectations[0].ExpectationName()).To(Equal("aave.Withdraw"))
		Expect(plan.Steps[2].Expectations[0].ExpectationName()).To(Equal("aave.Repay"))
		Expect(plan.Steps[3].Expectations[0].ExpectationName()).To(Equal("aave.Repay"))
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
		Expect(plan.Steps[0].Name).To(Equal("aave.RepayAll"))
		Expect(plan.Steps[0].Expectations[0].ExpectationName()).To(Equal("aave.Repay"))
		Expect(plan.Steps[1].Name).To(Equal("aave.WithdrawAll"))
		Expect(plan.Steps[1].Expectations[0].ExpectationName()).To(Equal("aave.Withdraw"))
	})

	It("rejects missing permit signature deadlines", func() {
		_, usdc, _ := stepTestReserves()
		permit := testPermitCapability(usdc.Underlying(), "2")
		plan, err := defi.NewFlow(
			common.HexToAddress("0x00000000000000000000000000000000000000aa"),
			defi.WithChain(config.Base),
		).
			Add(SupplyWithPermit(usdc, permit, decimal.NewFromInt(1), nil, 0, [32]byte{}, [32]byte{})).
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
				decimal.NewFromInt(1),
				big.NewInt(2_000_000_000),
				27,
				r,
				s,
			)).
			Build(context.Background(), nil)

		Expect(plan).To(BeNil())
		Expect(err).To(MatchError(ContainSubstring("does not match expected token")))
	})
})
