//go:build integration

package integration

import (
	"context"
	"math/big"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tn606024/defi-simplify/config"
)

var _ = Describe("Base fork smoke", func() {
	It("connects to Base and verifies configured contracts have code", func() {
		ctx := context.Background()
		client := baseForkClient(GinkgoT())

		chainID, err := client.ChainID(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(chainID).To(Equal(big.NewInt(8453)))

		pool, err := config.Base.AaveV3PoolAddress()
		Expect(err).NotTo(HaveOccurred())
		assertContractCode(GinkgoT(), ctx, client, pool, "Aave V3 Pool")

		usdc, err := config.USDC.Address(config.Base)
		Expect(err).NotTo(HaveOccurred())
		assertContractCode(GinkgoT(), ctx, client, usdc, "USDC")
	})
})
