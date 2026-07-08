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
	"github.com/tn606024/defi-simplify/erc20"
	"github.com/tn606024/defi-simplify/helper"
)

var _ = Describe("Aave Flow steps", func() {
	var (
		ctx  context.Context
		user common.Address
	)

	BeforeEach(func() {
		ctx = context.Background()
		user = common.HexToAddress("0x00000000000000000000000000000000000000aa")
	})

	It("builds approve, supply, and borrow calls matching the low-level action builders", func() {
		supplyAmount := decimal.RequireFromString("100.5")
		borrowAmount := decimal.RequireFromString("0.01")

		calls, err := defi.NewFlow(user, defi.WithChain(config.Base)).
			Add(erc20.Approve(config.USDC, PoolSpender(), supplyAmount)).
			Add(Supply(config.USDC, supplyAmount)).
			Add(Borrow(config.WETH, borrowAmount)).
			Build(ctx, nil)

		Expect(err).NotTo(HaveOccurred())
		Expect(calls).To(Equal([]defi.Call{
			expectedApproveCall(ctx, config.USDC, supplyAmount),
			expectedSupplyCall(ctx, user, config.USDC, supplyAmount),
			expectedBorrowCall(ctx, user, config.WETH, borrowAmount),
		}))
	})

	It("builds ApproveSupply as the Aave pool approval helper", func() {
		amount := decimal.RequireFromString("42")

		calls, err := defi.NewFlow(user, defi.WithChain(config.Base)).
			Add(ApproveSupply(config.USDC, amount)).
			Build(ctx, nil)

		Expect(err).NotTo(HaveOccurred())
		Expect(calls).To(Equal([]defi.Call{
			expectedApproveCall(ctx, config.USDC, amount),
		}))
	})

	It("returns a useful error for unsupported assets", func() {
		calls, err := defi.NewFlow(user, defi.WithChain(config.Base)).
			Add(Supply(config.Coin(9999), decimal.NewFromInt(1))).
			Build(ctx, nil)

		Expect(calls).To(BeNil())
		Expect(err).To(MatchError(ContainSubstring("build flow step 1 aave.Supply")))
		Expect(err).To(MatchError(ContainSubstring("unsupported coin")))
	})

	It("rejects non-positive amounts", func() {
		calls, err := defi.NewFlow(user, defi.WithChain(config.Base)).
			Add(Borrow(config.WETH, decimal.Zero)).
			Build(ctx, nil)

		Expect(calls).To(BeNil())
		Expect(err).To(MatchError(ContainSubstring("amount must be positive")))
	})
})

func expectedApproveCall(ctx context.Context, coin config.Coin, amount decimal.Decimal) defi.Call {
	poolAddress, err := config.Base.AaveV3PoolAddress()
	Expect(err).NotTo(HaveOccurred())
	coinAddress, amountWei := coinAddressAndAmount(coin, amount)
	return mustCall(ctx, contract.BuildApproveAction(coinAddress, poolAddress, amountWei))
}

func expectedSupplyCall(ctx context.Context, user common.Address, coin config.Coin, amount decimal.Decimal) defi.Call {
	poolAddress, err := config.Base.AaveV3PoolAddress()
	Expect(err).NotTo(HaveOccurred())
	coinAddress, amountWei := coinAddressAndAmount(coin, amount)
	return mustCall(ctx, contract.BuildSupplyAction(poolAddress, coinAddress, amountWei, user))
}

func expectedBorrowCall(ctx context.Context, user common.Address, coin config.Coin, amount decimal.Decimal) defi.Call {
	poolAddress, err := config.Base.AaveV3PoolAddress()
	Expect(err).NotTo(HaveOccurred())
	coinAddress, amountWei := coinAddressAndAmount(coin, amount)
	return mustCall(ctx, contract.BuildBorrowAction(poolAddress, coinAddress, amountWei, user))
}

func coinAddressAndAmount(coin config.Coin, amount decimal.Decimal) (common.Address, *big.Int) {
	coinAddress, err := coin.Address(config.Base)
	Expect(err).NotTo(HaveOccurred())
	decimals, err := coin.Decimals()
	Expect(err).NotTo(HaveOccurred())
	return coinAddress, helper.ToWei(amount, decimals)
}

func mustCall(ctx context.Context, action defi.Action) defi.Call {
	call, err := action.ToCall(ctx, nil, nil)
	Expect(err).NotTo(HaveOccurred())
	Expect(call).NotTo(BeNil())
	return *call
}
