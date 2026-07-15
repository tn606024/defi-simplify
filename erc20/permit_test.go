package erc20

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tn606024/defi-simplify/config"
	"github.com/tn606024/defi-simplify/token"
)

var _ = Describe("ERC20 permit capabilities", func() {
	It("builds an explicit EIP-712 domain from a resolved token", func() {
		asset := newPermitTestToken("USD Coin", 6)

		capability, err := NewPermitCapability(asset, "2")
		Expect(err).NotTo(HaveOccurred())
		Expect(capability.Token()).To(Equal(asset))
		Expect(capability.Version()).To(Equal("2"))

		domain, err := capability.Domain()
		Expect(err).NotTo(HaveOccurred())
		Expect(domain.Name).To(Equal("USD Coin"))
		Expect(domain.Version).To(Equal("2"))
		Expect(domain.ChainId).To(Equal(big.NewInt(8453)))
		Expect(domain.VerifyingContract).To(Equal(asset.Address()))
	})

	DescribeTable("rejects incomplete capabilities",
		func(asset token.Token, version string, expectedError string) {
			capability, err := NewPermitCapability(asset, version)

			Expect(capability).To(Equal(PermitCapability{}))
			Expect(err).To(MatchError(ContainSubstring(expectedError)))
		},
		Entry("invalid token", token.Token{}, "1", "invalid permit token"),
		Entry("missing token name", newPermitTestToken("", 6), "1", "token name is empty"),
		Entry("missing version", newPermitTestToken("USD Coin", 6), "", "domain version is empty"),
	)
})

func newPermitTestToken(name string, decimals uint8) token.Token {
	GinkgoHelper()
	ref, err := token.NewRef(
		config.Base,
		common.HexToAddress("0x0000000000000000000000000000000000000101"),
	)
	Expect(err).NotTo(HaveOccurred())
	asset, err := token.New(ref, "TEST", name, decimals)
	Expect(err).NotTo(HaveOccurred())
	return asset
}
