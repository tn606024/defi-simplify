//go:build integration

package integration

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tn606024/defi-simplify/aave"
	"github.com/tn606024/defi-simplify/assets/base"
)

var _ = Describe("Base asset catalog", func() {
	It("resolves reviewed USDC and WETH identities through the public Aave registry", func() {
		ctx := context.Background()
		client := baseForkClient(GinkgoT())
		market, err := aave.BaseV3Market()
		Expect(err).NotTo(HaveOccurred())
		registry, err := aave.NewRegistry(client, market)
		Expect(err).NotTo(HaveOccurred())

		snapshot, err := registry.Load(ctx)
		Expect(err).NotTo(HaveOccurred())

		usdc, err := snapshot.Reserve(base.USDC)
		Expect(err).NotTo(HaveOccurred())
		Expect(usdc.Underlying().Ref()).To(Equal(base.USDC))
		Expect(usdc.Underlying().Symbol()).To(Equal("USDC"))
		Expect(usdc.Underlying().Decimals()).To(Equal(uint8(6)))

		weth, err := snapshot.Reserve(base.WETH)
		Expect(err).NotTo(HaveOccurred())
		Expect(weth.Underlying().Ref()).To(Equal(base.WETH))
		Expect(weth.Underlying().Symbol()).To(Equal("WETH"))
		Expect(weth.Underlying().Decimals()).To(Equal(uint8(18)))
	})
})
