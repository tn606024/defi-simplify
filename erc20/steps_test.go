package erc20

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

var _ = Describe("ERC20 Flow steps", func() {
	var (
		ctx     context.Context
		account common.Address
		spender common.Address
		to      common.Address
	)

	BeforeEach(func() {
		ctx = context.Background()
		account = common.HexToAddress("0x00000000000000000000000000000000000000aa")
		spender = common.HexToAddress("0x00000000000000000000000000000000000000bb")
		to = common.HexToAddress("0x00000000000000000000000000000000000000cc")
	})

	It("builds approve calls matching the low-level action builder", func() {
		amount := decimal.RequireFromString("100.5")

		plan, err := defi.NewFlow(account, defi.WithChain(config.Base)).
			Add(Approve(config.USDC, AddressSpender(spender), amount)).
			Build(ctx, nil)

		Expect(err).NotTo(HaveOccurred())
		Expect(plan.Calls()).To(Equal([]defi.Call{
			expectedApproveCall(ctx, config.USDC, spender, amount),
		}))
	})

	It("builds transfer calls matching the low-level action builder", func() {
		amount := decimal.RequireFromString("0.125")

		plan, err := defi.NewFlow(account, defi.WithChain(config.Base)).
			Add(Transfer(config.WETH, to, amount)).
			Build(ctx, nil)

		Expect(err).NotTo(HaveOccurred())
		Expect(plan.Calls()).To(Equal([]defi.Call{
			expectedTransferCall(ctx, config.WETH, to, amount),
		}))
	})

	It("builds transferFrom calls matching the low-level action builder", func() {
		amount := decimal.RequireFromString("2.25")

		plan, err := defi.NewFlow(account, defi.WithChain(config.Base)).
			Add(TransferFrom(config.USDC, account, to, amount)).
			Build(ctx, nil)

		Expect(err).NotTo(HaveOccurred())
		Expect(plan.Calls()).To(Equal([]defi.Call{
			expectedTransferFromCall(ctx, config.USDC, account, to, amount),
		}))
	})

	It("builds permit calls matching the low-level action builder", func() {
		amount := decimal.RequireFromString("12.34")
		deadline := big.NewInt(1893456000)
		v := uint8(27)
		var r [32]byte
		var s [32]byte
		copy(r[:], common.Hex2Bytes("1111111111111111111111111111111111111111111111111111111111111111"))
		copy(s[:], common.Hex2Bytes("2222222222222222222222222222222222222222222222222222222222222222"))

		plan, err := defi.NewFlow(account, defi.WithChain(config.Base)).
			Add(Permit(config.USDC, account, AddressSpender(spender), amount, deadline, v, r, s)).
			Build(ctx, nil)

		Expect(err).NotTo(HaveOccurred())
		Expect(plan.Calls()).To(Equal([]defi.Call{
			expectedPermitCall(ctx, config.USDC, account, spender, amount, deadline, v, r, s),
		}))
	})

	It("returns a useful error for unsupported token config", func() {
		plan, err := defi.NewFlow(account, defi.WithChain(config.Base)).
			Add(Transfer(config.Coin(9999), to, decimal.NewFromInt(1))).
			Build(ctx, nil)

		Expect(plan).To(BeNil())
		Expect(err).To(MatchError(ContainSubstring("build flow step 1 erc20.Transfer")))
		Expect(err).To(MatchError(ContainSubstring("unsupported coin")))
	})

	It("rejects negative amounts", func() {
		plan, err := defi.NewFlow(account, defi.WithChain(config.Base)).
			Add(Approve(config.USDC, AddressSpender(spender), decimal.NewFromInt(-1))).
			Build(ctx, nil)

		Expect(plan).To(BeNil())
		Expect(err).To(MatchError(ContainSubstring("amount must not be negative")))
	})

	It("rejects missing spender resolvers", func() {
		plan, err := defi.NewFlow(account, defi.WithChain(config.Base)).
			Add(Approve(config.USDC, nil, decimal.NewFromInt(1))).
			Build(ctx, nil)

		Expect(plan).To(BeNil())
		Expect(err).To(MatchError(ContainSubstring("spender is nil")))
	})
})

func expectedApproveCall(ctx context.Context, coin config.Coin, spender common.Address, amount decimal.Decimal) defi.Call {
	coinAddress, amountWei := coinAddressAndAmount(coin, amount)
	return mustCall(ctx, contract.BuildApproveAction(coinAddress, spender, amountWei))
}

func expectedTransferCall(ctx context.Context, coin config.Coin, to common.Address, amount decimal.Decimal) defi.Call {
	coinAddress, amountWei := coinAddressAndAmount(coin, amount)
	return mustCall(ctx, contract.BuildTransferAction(coinAddress, to, amountWei))
}

func expectedTransferFromCall(ctx context.Context, coin config.Coin, from common.Address, to common.Address, amount decimal.Decimal) defi.Call {
	coinAddress, amountWei := coinAddressAndAmount(coin, amount)
	return mustCall(ctx, contract.BuildTransferFromAction(coinAddress, from, to, amountWei))
}

func expectedPermitCall(ctx context.Context, coin config.Coin, owner common.Address, spender common.Address, amount decimal.Decimal, deadline *big.Int, v uint8, r [32]byte, s [32]byte) defi.Call {
	coinAddress, amountWei := coinAddressAndAmount(coin, amount)
	return mustCall(ctx, contract.BuildPermitAction(coinAddress, owner, spender, amountWei, deadline, v, r, s))
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
