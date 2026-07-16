package aave

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tn606024/defi-simplify/config"
)

var _ = Describe("Aave delegation capabilities", func() {
	It("builds the variable debt token EIP-712 domain", func() {
		reserve := testDelegationReserve()

		capability, err := NewDelegationCapability(reserve, "1")
		Expect(err).NotTo(HaveOccurred())
		Expect(capability.Reserve()).To(Equal(reserve))
		Expect(capability.Version()).To(Equal("1"))

		domain, err := capability.Domain()
		Expect(err).NotTo(HaveOccurred())
		Expect(domain.Name).To(Equal(reserve.VariableDebtToken().Name()))
		Expect(domain.Version).To(Equal("1"))
		Expect(domain.ChainId).To(Equal(big.NewInt(8453)))
		Expect(domain.VerifyingContract).To(Equal(reserve.VariableDebtToken().Address()))
	})

	It("rejects incomplete delegation capabilities", func() {
		capability, err := NewDelegationCapability(Reserve{}, "1")
		Expect(capability).To(Equal(DelegationCapability{}))
		Expect(err).To(MatchError(ContainSubstring("invalid delegation reserve")))

		capability, err = NewDelegationCapability(testDelegationReserve(), " ")
		Expect(capability).To(Equal(DelegationCapability{}))
		Expect(err).To(MatchError(ContainSubstring("domain version is empty")))
	})
})

func testDelegationReserve() Reserve {
	GinkgoHelper()
	market, err := NewMarket(
		"aave-v3-base",
		config.Base,
		common.HexToAddress("0x1000000000000000000000000000000000000001"),
		common.HexToAddress("0x1000000000000000000000000000000000000002"),
		common.HexToAddress("0x1000000000000000000000000000000000000003"),
		common.HexToAddress("0x1000000000000000000000000000000000000004"),
	)
	Expect(err).NotTo(HaveOccurred())
	return mustReserve(
		market,
		"0x2000000000000000000000000000000000000001",
		"0x2000000000000000000000000000000000000002",
		"0x2000000000000000000000000000000000000003",
	)
}
