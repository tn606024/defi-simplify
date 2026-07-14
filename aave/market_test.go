package aave

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tn606024/defi-simplify/config"
	"github.com/tn606024/defi-simplify/token"
)

var _ = Describe("Aave market models", func() {
	var (
		market        Market
		underlyingRef token.Ref
		underlying    token.Token
		aToken        token.Token
		variableDebt  token.Token
		stableDebt    token.Token
	)

	BeforeEach(func() {
		var err error
		market, err = NewMarket(
			"aave-v3-base",
			config.Base,
			common.HexToAddress("0x1000000000000000000000000000000000000001"),
			common.HexToAddress("0x1000000000000000000000000000000000000002"),
			common.HexToAddress("0x1000000000000000000000000000000000000003"),
			common.HexToAddress("0x1000000000000000000000000000000000000004"),
		)
		Expect(err).NotTo(HaveOccurred())

		underlyingRef = mustTokenRef("0x2000000000000000000000000000000000000001")
		underlying = mustToken(underlyingRef, "USDC", "USD Coin", 6)
		aToken = mustToken(
			mustTokenRef("0x2000000000000000000000000000000000000002"),
			"aBasUSDC",
			"Aave Base USDC",
			6,
		)
		variableDebt = mustToken(
			mustTokenRef("0x2000000000000000000000000000000000000003"),
			"variableDebtBasUSDC",
			"Aave Base Variable Debt USDC",
			6,
		)
		stableDebt = mustToken(
			mustTokenRef("0x2000000000000000000000000000000000000004"),
			"stableDebtBasUSDC",
			"Aave Base Stable Debt USDC",
			6,
		)
	})

	It("represents a resolved market as an immutable value", func() {
		Expect(market.Validate()).To(Succeed())
		Expect(market.ID()).To(Equal("aave-v3-base"))
		Expect(market.Chain()).To(Equal(config.Base))
		Expect(market.Pool()).NotTo(Equal(common.Address{}))
		Expect(market.AddressesProvider()).NotTo(Equal(common.Address{}))
		Expect(market.ProtocolDataProvider()).NotTo(Equal(common.Address{}))
		gateway, ok := market.WrappedTokenGateway()
		Expect(ok).To(BeTrue())
		Expect(gateway).NotTo(Equal(common.Address{}))
	})

	It("supports markets without an optional wrapped-token gateway", func() {
		withoutGateway, err := NewMarket(
			"aave-v3-base",
			config.Base,
			market.Pool(),
			market.AddressesProvider(),
			market.ProtocolDataProvider(),
			common.Address{},
		)
		Expect(err).NotTo(HaveOccurred())
		gateway, ok := withoutGateway.WrappedTokenGateway()
		Expect(ok).To(BeFalse())
		Expect(gateway).To(Equal(common.Address{}))
	})

	It("rejects malformed public market definitions", func() {
		_, err := NewMarket(
			"aave-v3-base",
			config.Base,
			common.Address{},
			market.AddressesProvider(),
			market.ProtocolDataProvider(),
			common.Address{},
		)
		Expect(errors.Is(err, ErrInvalidMarket)).To(BeTrue())
	})

	It("groups token roles without assuming their decimal precision matches", func() {
		differentDecimals := mustToken(variableDebt.Ref(), variableDebt.Symbol(), variableDebt.Name(), 18)
		reserve, err := NewReserve(market, underlying, aToken, differentDecimals, &stableDebt)
		Expect(err).NotTo(HaveOccurred())

		Expect(reserve.Validate()).To(Succeed())
		Expect(reserve.Market()).To(Equal(market))
		Expect(reserve.Underlying()).To(Equal(underlying))
		Expect(reserve.AToken()).To(Equal(aToken))
		Expect(reserve.VariableDebtToken().Decimals()).To(Equal(uint8(18)))
		resolvedStableDebt, ok := reserve.StableDebtToken()
		Expect(ok).To(BeTrue())
		Expect(resolvedStableDebt).To(Equal(stableDebt))
	})

	It("supports a reserve without a stable debt token", func() {
		reserve, err := NewReserve(market, underlying, aToken, variableDebt, nil)
		Expect(err).NotTo(HaveOccurred())
		resolvedStableDebt, ok := reserve.StableDebtToken()
		Expect(ok).To(BeFalse())
		Expect(resolvedStableDebt).To(Equal(token.Token{}))
	})

	It("rejects duplicate public reserve token roles", func() {
		_, err := NewReserve(market, underlying, aToken, aToken, nil)
		Expect(errors.Is(err, ErrInvalidReserve)).To(BeTrue())
	})

	It("indexes reserves by chain-scoped underlying identity", func() {
		usdc, err := NewReserve(market, underlying, aToken, variableDebt, &stableDebt)
		Expect(err).NotTo(HaveOccurred())
		weth := mustReserve(
			market,
			"0x3000000000000000000000000000000000000001",
			"0x3000000000000000000000000000000000000002",
			"0x3000000000000000000000000000000000000003",
		)
		input := []Reserve{weth, usdc}

		snapshot, err := NewMarketSnapshot(
			market,
			12345,
			common.HexToHash("0x1234"),
			input,
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(snapshot.Validate()).To(Succeed())
		Expect(snapshot.Market()).To(Equal(market))
		Expect(snapshot.BlockNumber()).To(Equal(uint64(12345)))
		Expect(snapshot.BlockHash()).To(Equal(common.HexToHash("0x1234")))
		Expect(snapshot.Len()).To(Equal(2))

		byRef, err := snapshot.Reserve(underlyingRef)
		Expect(err).NotTo(HaveOccurred())
		Expect(byRef).To(Equal(usdc))
		byAddress, err := snapshot.ReserveByAddress(underlying.Address())
		Expect(err).NotTo(HaveOccurred())
		Expect(byAddress).To(Equal(usdc))

		ordered := snapshot.Reserves()
		Expect(ordered).To(HaveLen(2))
		Expect(ordered[0].Underlying().Address()).To(Equal(underlying.Address()))
		Expect(ordered[1].Underlying().Address()).To(Equal(weth.Underlying().Address()))
	})

	It("does not retain or expose a mutable reserve collection", func() {
		original := mustReserve(
			market,
			"0x3000000000000000000000000000000000000001",
			"0x3000000000000000000000000000000000000002",
			"0x3000000000000000000000000000000000000003",
		)
		replacement := mustReserve(
			market,
			"0x4000000000000000000000000000000000000001",
			"0x4000000000000000000000000000000000000002",
			"0x4000000000000000000000000000000000000003",
		)
		input := []Reserve{original}
		snapshot, err := NewMarketSnapshot(market, 12345, common.HexToHash("0x1234"), input)
		Expect(err).NotTo(HaveOccurred())

		input[0] = replacement
		returned := snapshot.Reserves()
		returned[0] = replacement

		resolved, err := snapshot.ReserveByAddress(original.Underlying().Address())
		Expect(err).NotTo(HaveOccurred())
		Expect(resolved).To(Equal(original))
		_, err = snapshot.ReserveByAddress(replacement.Underlying().Address())
		Expect(errors.Is(err, ErrReserveNotFound)).To(BeTrue())
	})

	It("rejects duplicate reserves through the public snapshot constructor", func() {
		reserve := mustReserve(
			market,
			"0x3000000000000000000000000000000000000001",
			"0x3000000000000000000000000000000000000002",
			"0x3000000000000000000000000000000000000003",
		)
		_, err := NewMarketSnapshot(
			market,
			12345,
			common.HexToHash("0x1234"),
			[]Reserve{reserve, reserve},
		)
		Expect(errors.Is(err, ErrInvalidMarketSnapshot)).To(BeTrue())
		Expect(errors.Is(err, ErrDuplicateReserve)).To(BeTrue())
	})
})

func mustTokenRef(address string) token.Ref {
	ref, err := token.NewRef(config.Base, common.HexToAddress(address))
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	return ref
}

func mustToken(ref token.Ref, symbol string, name string, decimals uint8) token.Token {
	resolved, err := token.New(ref, symbol, name, decimals)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	return resolved
}

func mustReserve(market Market, underlying string, aToken string, variableDebt string) Reserve {
	reserve, err := NewReserve(
		market,
		mustToken(mustTokenRef(underlying), "UNDERLYING", "Underlying", 18),
		mustToken(mustTokenRef(aToken), "ATOKEN", "A Token", 18),
		mustToken(mustTokenRef(variableDebt), "VARIABLE_DEBT", "Variable Debt", 18),
		nil,
	)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	return reserve
}
