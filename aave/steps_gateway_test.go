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
	"github.com/tn606024/defi-simplify/helper"
)

var _ = Describe("Aave WrappedTokenGateway Flow steps", func() {
	It("builds native ETH calls with gateway-aware event expectations", func() {
		ctx := context.Background()
		account := common.HexToAddress("0x00000000000000000000000000000000000000aa")
		amount := decimal.RequireFromString("0.01")
		deadline := big.NewInt(2_000_000_000)
		v := uint8(28)
		var r, s [32]byte
		r[0] = 1
		s[0] = 2

		plan, err := defi.NewFlow(account, defi.WithChain(config.Base)).
			Add(DepositETH(amount)).
			Add(BorrowETH(amount)).
			Add(WithdrawETH(amount)).
			Add(WithdrawETHWithPermit(amount, deadline, v, r, s)).
			Build(ctx, nil)

		Expect(err).NotTo(HaveOccurred())
		gateway, err := config.Base.WrappedTokenGatewayV3Address()
		Expect(err).NotTo(HaveOccurred())
		pool, err := config.Base.AaveV3PoolAddress()
		Expect(err).NotTo(HaveOccurred())
		decimals, err := config.Base.GasTokenDecimals()
		Expect(err).NotTo(HaveOccurred())
		amountWei := helper.ToWei(amount, decimals)
		Expect(plan.Calls()).To(Equal([]defi.Call{
			mustCall(ctx, contract.BuildDepositETHAction(gateway, pool, account, 0, amountWei)),
			mustCall(ctx, contract.BuildBorrowETHAction(gateway, pool, amountWei)),
			mustCall(ctx, contract.BuildWithdrawETHAction(gateway, pool, amountWei, account)),
			mustCall(ctx, contract.BuildWithdrawETHWithPermitAction(gateway, pool, amountWei, account, deadline, v, r, s)),
		}))
		Expect(plan.Calls()[0].Value).To(Equal(amountWei))
		Expect(plan.Steps[0].Expectations[0].ExpectationName()).To(Equal("aave.Supply"))
		Expect(plan.Steps[1].Expectations[0].ExpectationName()).To(Equal("aave.Borrow"))
		Expect(plan.Steps[2].Expectations[0].ExpectationName()).To(Equal("aave.Withdraw"))
		Expect(plan.Steps[3].Expectations[0].ExpectationName()).To(Equal("aave.Withdraw"))
		Expect(GatewaySpender().Address(config.Base)).To(Equal(gateway))
	})
})
