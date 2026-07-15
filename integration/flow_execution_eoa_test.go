//go:build integration

package integration

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/shopspring/decimal"
	defi "github.com/tn606024/defi-simplify"
	"github.com/tn606024/defi-simplify/bind/erc20"
	"github.com/tn606024/defi-simplify/config"
	sdkerc20 "github.com/tn606024/defi-simplify/erc20"
)

var _ = Describe("Flow ExecutionEOA integration", func() {
	It("executes ERC20 approval with user EOA allowance ownership on a local fork", func() {
		ctx := context.Background()
		ethClient := baseForkClient(GinkgoT())
		rpcClient := baseForkRPCClient(GinkgoT())
		requireAnvilFork(GinkgoT(), ctx, rpcClient)

		opts, _, user := newForkTransactor(GinkgoT(), ctx, rpcClient)
		spender := common.HexToAddress("0x00000000000000000000000000000000000000bb")

		_, usdcReserve, _ := loadBaseAaveReserves(GinkgoT(), ctx, ethClient)
		usdc := usdcReserve.Underlying().Address()
		assertContractCode(GinkgoT(), ctx, ethClient, usdc, "USDC")

		approveAmount := big.NewInt(1_000_000)
		Expect(fundBaseUSDCFromHolder(ctx, rpcClient, ethClient, user, approveAmount)).To(Succeed())

		token, err := erc20.NewErc20(usdc, ethClient)
		Expect(err).NotTo(HaveOccurred())

		before, err := token.Allowance(nil, user, spender)
		Expect(err).NotTo(HaveOccurred())
		Expect(before.Sign()).To(Equal(0))

		flow := defi.NewFlow(user, defi.WithChain(config.Base)).
			Add(sdkerc20.Approve(usdcReserve.Underlying(), sdkerc20.AddressSpender(spender), decimal.NewFromInt(1)))
		runner := defi.NewRunner(ethClient, opts, config.Base)

		receipt, err := runner.Execute(ctx, flow, defi.ExecutionEOA)
		Expect(err).NotTo(HaveOccurred())
		Expect(receipt.Status).To(Equal(uint64(1)))

		after, err := token.Allowance(nil, user, spender)
		Expect(err).NotTo(HaveOccurred())
		Expect(after.Cmp(approveAmount)).To(Equal(0))
	})
})
