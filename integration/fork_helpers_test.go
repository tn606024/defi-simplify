//go:build integration

package integration

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tn606024/defi-simplify/bind/erc20"
	"github.com/tn606024/defi-simplify/config"
)

var forkTestUser = common.HexToAddress("0x1000000000000000000000000000000000000001")

var _ = Describe("Fork helpers", func() {
	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()
	})

	It("sets ETH balance for a test user", func() {
		ethClient := baseForkClient(GinkgoT())
		rpcClient := baseForkRPCClient(GinkgoT())
		requireAnvilFork(GinkgoT(), ctx, rpcClient)

		balance := new(big.Int).Mul(big.NewInt(2), big.NewInt(1_000_000_000_000_000_000))
		Expect(setForkETHBalance(ctx, rpcClient, forkTestUser, balance)).To(Succeed())

		actual, err := ethClient.BalanceAt(ctx, forkTestUser, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(actual.Cmp(balance)).To(Equal(0))
	})

	It("starts and stops account impersonation", func() {
		rpcClient := baseForkRPCClient(GinkgoT())
		requireAnvilFork(GinkgoT(), ctx, rpcClient)

		Expect(impersonateForkAccount(ctx, rpcClient, baseUSDCFunder)).To(Succeed())
		Expect(stopImpersonatingForkAccount(ctx, rpcClient, baseUSDCFunder)).To(Succeed())
	})

	It("funds a test user with USDC from an impersonated Base holder", func() {
		ethClient := baseForkClient(GinkgoT())
		rpcClient := baseForkRPCClient(GinkgoT())
		requireAnvilFork(GinkgoT(), ctx, rpcClient)

		usdc, err := config.USDC.Address(config.Base)
		Expect(err).NotTo(HaveOccurred())
		token, err := erc20.NewErc20(usdc, ethClient)
		Expect(err).NotTo(HaveOccurred())

		before, err := token.BalanceOf(nil, forkTestUser)
		Expect(err).NotTo(HaveOccurred())

		amount := big.NewInt(1_000_000)
		Expect(fundBaseUSDCFromHolder(ctx, rpcClient, ethClient, forkTestUser, amount)).To(Succeed())

		after, err := token.BalanceOf(nil, forkTestUser)
		Expect(err).NotTo(HaveOccurred())
		delta := new(big.Int).Sub(after, before)
		Expect(delta.Cmp(amount)).To(Equal(0))
	})
})
