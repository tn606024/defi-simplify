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
	"github.com/tn606024/defi-simplify/token"
)

var _ = Describe("ERC20 Flow steps", func() {
	var (
		ctx     context.Context
		account common.Address
		spender common.Address
		to      common.Address
		usdc    token.Token
		weth    token.Token
	)

	BeforeEach(func() {
		ctx = context.Background()
		account = common.HexToAddress("0x00000000000000000000000000000000000000aa")
		spender = common.HexToAddress("0x00000000000000000000000000000000000000bb")
		to = common.HexToAddress("0x00000000000000000000000000000000000000cc")
		usdc = resolvedTestToken(
			"0x0000000000000000000000000000000000000101",
			"USDC",
			"USD Coin",
			6,
		)
		weth = resolvedTestToken(
			"0x0000000000000000000000000000000000000102",
			"WETH",
			"Wrapped Ether",
			18,
		)
	})

	It("builds approve calls matching the low-level action builder", func() {
		amount := decimal.RequireFromString("100.5")

		plan, err := defi.NewFlow(account, defi.WithChain(config.Base)).
			Add(Approve(usdc, AddressSpender(spender), amount)).
			Build(ctx, nil)

		Expect(err).NotTo(HaveOccurred())
		Expect(plan.Calls()).To(Equal([]defi.Call{
			expectedApproveCall(ctx, usdc, spender, amount),
		}))
		Expect(plan.Steps[0].Expectations[0].ExpectationName()).To(Equal("erc20.Approval"))
	})

	It("builds transfer calls matching the low-level action builder", func() {
		amount := decimal.RequireFromString("0.125")

		plan, err := defi.NewFlow(account, defi.WithChain(config.Base)).
			Add(Transfer(weth, to, amount)).
			Build(ctx, nil)

		Expect(err).NotTo(HaveOccurred())
		Expect(plan.Calls()).To(Equal([]defi.Call{
			expectedTransferCall(ctx, weth, to, amount),
		}))
		Expect(plan.Steps[0].Expectations[0].ExpectationName()).To(Equal("erc20.Transfer"))
	})

	It("builds transferFrom calls matching the low-level action builder", func() {
		amount := decimal.RequireFromString("2.25")

		plan, err := defi.NewFlow(account, defi.WithChain(config.Base)).
			Add(TransferFrom(usdc, account, to, amount)).
			Build(ctx, nil)

		Expect(err).NotTo(HaveOccurred())
		Expect(plan.Calls()).To(Equal([]defi.Call{
			expectedTransferFromCall(ctx, usdc, account, to, amount),
		}))
		Expect(plan.Steps[0].Expectations[0].ExpectationName()).To(Equal("erc20.Transfer"))
	})

	It("builds permit calls matching the low-level action builder", func() {
		amount := decimal.RequireFromString("12.34")
		deadline := big.NewInt(1893456000)
		v := uint8(27)
		var r [32]byte
		var s [32]byte
		copy(r[:], common.Hex2Bytes("1111111111111111111111111111111111111111111111111111111111111111"))
		copy(s[:], common.Hex2Bytes("2222222222222222222222222222222222222222222222222222222222222222"))

		permit, err := NewPermitCapability(usdc, "2")
		Expect(err).NotTo(HaveOccurred())
		plan, err := defi.NewFlow(account, defi.WithChain(config.Base)).
			Add(Permit(permit, account, AddressSpender(spender), amount, deadline, v, r, s)).
			Build(ctx, nil)

		Expect(err).NotTo(HaveOccurred())
		Expect(plan.Calls()).To(Equal([]defi.Call{
			expectedPermitCall(ctx, usdc, account, spender, amount, deadline, v, r, s),
		}))
		Expect(plan.Steps[0].Expectations[0].ExpectationName()).To(Equal("erc20.Approval"))
	})

	It("returns a useful error for an unresolved token", func() {
		plan, err := defi.NewFlow(account, defi.WithChain(config.Base)).
			Add(Transfer(token.Token{}, to, decimal.NewFromInt(1))).
			Build(ctx, nil)

		Expect(plan).To(BeNil())
		Expect(err).To(MatchError(ContainSubstring("build flow step 1 erc20.Transfer")))
		Expect(err).To(MatchError(ContainSubstring("invalid token")))
	})

	It("rejects negative amounts", func() {
		plan, err := defi.NewFlow(account, defi.WithChain(config.Base)).
			Add(Approve(usdc, AddressSpender(spender), decimal.NewFromInt(-1))).
			Build(ctx, nil)

		Expect(plan).To(BeNil())
		Expect(err).To(MatchError(ContainSubstring("amount must not be negative")))
	})

	It("rejects missing spender resolvers", func() {
		plan, err := defi.NewFlow(account, defi.WithChain(config.Base)).
			Add(Approve(usdc, nil, decimal.NewFromInt(1))).
			Build(ctx, nil)

		Expect(plan).To(BeNil())
		Expect(err).To(MatchError(ContainSubstring("spender is nil")))
	})
})

func expectedApproveCall(ctx context.Context, asset token.Token, spender common.Address, amount decimal.Decimal) defi.Call {
	return mustCall(ctx, contract.BuildApproveAction(asset.Address(), spender, tokenAmount(asset, amount)))
}

func expectedTransferCall(ctx context.Context, asset token.Token, to common.Address, amount decimal.Decimal) defi.Call {
	return mustCall(ctx, contract.BuildTransferAction(asset.Address(), to, tokenAmount(asset, amount)))
}

func expectedTransferFromCall(ctx context.Context, asset token.Token, from common.Address, to common.Address, amount decimal.Decimal) defi.Call {
	return mustCall(ctx, contract.BuildTransferFromAction(asset.Address(), from, to, tokenAmount(asset, amount)))
}

func expectedPermitCall(ctx context.Context, asset token.Token, owner common.Address, spender common.Address, amount decimal.Decimal, deadline *big.Int, v uint8, r [32]byte, s [32]byte) defi.Call {
	return mustCall(ctx, contract.BuildPermitAction(asset.Address(), owner, spender, tokenAmount(asset, amount), deadline, v, r, s))
}

func tokenAmount(asset token.Token, amount decimal.Decimal) *big.Int {
	return helper.ToWei(amount, asset.Decimals())
}

func resolvedTestToken(address, symbol, name string, decimals uint8) token.Token {
	GinkgoHelper()
	ref, err := token.NewRef(config.Base, common.HexToAddress(address))
	Expect(err).NotTo(HaveOccurred())
	asset, err := token.New(ref, symbol, name, decimals)
	Expect(err).NotTo(HaveOccurred())
	return asset
}

func mustCall(ctx context.Context, action defi.Action) defi.Call {
	call, err := action.ToCall(ctx, nil, nil)
	Expect(err).NotTo(HaveOccurred())
	Expect(call).NotTo(BeNil())
	return *call
}
