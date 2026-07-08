//go:build integration

package integration

import (
	"context"
	"math/big"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/shopspring/decimal"
	"github.com/tn606024/defi-simplify/bind/erc20"
	"github.com/tn606024/defi-simplify/config"
)

var _ = Describe("ERC20 approve integration", func() {
	It("approves the Base Aave V3 Pool to spend user USDC on a local fork", func() {
		ctx := context.Background()
		ethClient := baseForkClient(GinkgoT())
		rpcClient := baseForkRPCClient(GinkgoT())
		requireAnvilFork(GinkgoT(), ctx, rpcClient)

		client, user := newForkDefiClient(GinkgoT(), ctx, rpcClient, ethClient)

		usdc, err := config.USDC.Address(config.Base)
		Expect(err).NotTo(HaveOccurred())
		assertContractCode(GinkgoT(), ctx, ethClient, usdc, "USDC")

		pool, err := config.Base.AaveV3PoolAddress()
		Expect(err).NotTo(HaveOccurred())

		approveAmount := big.NewInt(1_000_000)
		Expect(fundBaseUSDCFromHolder(ctx, rpcClient, ethClient, user, approveAmount)).To(Succeed())

		token, err := erc20.NewErc20(usdc, ethClient)
		Expect(err).NotTo(HaveOccurred())

		balance, err := token.BalanceOf(nil, user)
		Expect(err).NotTo(HaveOccurred())
		Expect(balance.Cmp(approveAmount)).To(BeNumerically(">=", 0))

		before, err := token.Allowance(nil, user, pool)
		Expect(err).NotTo(HaveOccurred())
		Expect(before.Sign()).To(Equal(0))

		receipt, err := client.ERC20.Approve(ctx, config.USDC, pool, decimal.NewFromInt(1))
		Expect(err).NotTo(HaveOccurred())
		Expect(receipt.Status).To(Equal(uint64(1)))

		after, err := token.Allowance(nil, user, pool)
		Expect(err).NotTo(HaveOccurred())
		Expect(after.Cmp(approveAmount)).To(Equal(0))
	})
})
