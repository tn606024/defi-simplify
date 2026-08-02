//go:build integration

package integration

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tn606024/defi-simplify/aave"
	baseassets "github.com/tn606024/defi-simplify/assets/base"
	"github.com/tn606024/defi-simplify/client/account/defisimplify7702"
	"github.com/tn606024/defi-simplify/config"
)

var _ = Describe("SDK initialization", func() {
	It("verifies the account deployment and loads the Base Aave snapshot", func() {
		ctx := context.Background()
		ethClient := baseForkClient(GinkgoT())

		implementation, err := defisimplify7702.ResolveAndVerifyAccountDeployment(
			ctx,
			ethClient,
			config.Base,
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(implementation.Address).NotTo(Equal(common.Address{}))
		Expect(implementation.RuntimeCodeHash).NotTo(Equal(common.Hash{}))
		Expect(implementation.Version).To(Equal("v1.1.0"))

		snapshot, err := aave.LoadBaseV3Snapshot(ctx, ethClient)
		Expect(err).NotTo(HaveOccurred())
		Expect(snapshot.Validate()).To(Succeed())
		Expect(snapshot.Market().Chain()).To(Equal(config.Base))
		Expect(snapshot.Market().Pool()).NotTo(Equal(common.Address{}))
		Expect(snapshot.BlockNumber()).NotTo(BeZero())
		Expect(snapshot.BlockHash()).NotTo(Equal(common.Hash{}))

		usdcReserve, err := snapshot.Reserve(baseassets.USDC)
		Expect(err).NotTo(HaveOccurred())
		Expect(usdcReserve.Underlying().Ref()).To(Equal(baseassets.USDC))
		Expect(usdcReserve.Underlying().Decimals()).To(Equal(uint8(6)))

		wethReserve, err := snapshot.Reserve(baseassets.WETH)
		Expect(err).NotTo(HaveOccurred())
		Expect(wethReserve.Underlying().Ref()).To(Equal(baseassets.WETH))
		Expect(wethReserve.Underlying().Decimals()).To(Equal(uint8(18)))
	})
})
