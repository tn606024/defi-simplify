package token_test

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tn606024/defi-simplify/config"
	"github.com/tn606024/defi-simplify/token"
)

var _ = Describe("Token identity", func() {
	var address common.Address

	BeforeEach(func() {
		address = common.HexToAddress("0x1111111111111111111111111111111111111111")
	})

	It("creates a chain-scoped reference", func() {
		ref, err := token.NewRef(config.Base, address)
		Expect(err).NotTo(HaveOccurred())
		Expect(ref.Validate()).To(Succeed())
		Expect(ref.Chain()).To(Equal(config.Base))
		Expect(ref.Address()).To(Equal(address))
	})

	It("rejects an invalid reference", func() {
		_, err := token.NewRef(config.Base, common.Address{})
		Expect(errors.Is(err, token.ErrInvalidRef)).To(BeTrue())
	})

	It("keeps executable identity separate from display metadata", func() {
		ref, err := token.NewRef(config.Base, address)
		Expect(err).NotTo(HaveOccurred())

		first, err := token.New(ref, "USDC", "USD Coin", 6)
		Expect(err).NotTo(HaveOccurred())
		second, err := token.New(ref, "renamed", "different display name", 18)
		Expect(err).NotTo(HaveOccurred())

		Expect(first.Validate()).To(Succeed())
		Expect(first.Ref()).To(Equal(ref))
		Expect(first.Chain()).To(Equal(config.Base))
		Expect(first.Address()).To(Equal(address))
		Expect(first.Symbol()).To(Equal("USDC"))
		Expect(first.Name()).To(Equal("USD Coin"))
		Expect(first.Decimals()).To(Equal(uint8(6)))
		Expect(first.SameAsset(second)).To(BeTrue())
	})

	It("allows missing display metadata without weakening identity", func() {
		ref, err := token.NewRef(config.Base, address)
		Expect(err).NotTo(HaveOccurred())

		resolved, err := token.New(ref, "", "", 0)
		Expect(err).NotTo(HaveOccurred())
		Expect(resolved.Validate()).To(Succeed())
		Expect(resolved.Symbol()).To(BeEmpty())
		Expect(resolved.Name()).To(BeEmpty())
		Expect(resolved.Decimals()).To(BeZero())
	})

	It("rejects resolved metadata with a zero-value identity", func() {
		_, err := token.New(token.Ref{}, "USDC", "USD Coin", 6)
		Expect(errors.Is(err, token.ErrInvalidToken)).To(BeTrue())
	})
})
