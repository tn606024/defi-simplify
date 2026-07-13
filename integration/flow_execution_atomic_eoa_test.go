//go:build integration

package integration

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/shopspring/decimal"
	defi "github.com/tn606024/defi-simplify"
	"github.com/tn606024/defi-simplify/bind/erc20"
	"github.com/tn606024/defi-simplify/client/account/eip7702"
	"github.com/tn606024/defi-simplify/config"
	sdkerc20 "github.com/tn606024/defi-simplify/erc20"
)

var _ = Describe("Flow ExecutionAtomicEOA integration", func() {
	It("executes an ERC20 approval batch through a delegated EOA", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()

		ethClient := baseForkClient(GinkgoT())
		rpcClient := baseForkRPCClient(GinkgoT())
		requireAnvilFork(GinkgoT(), ctx, rpcClient)

		opts, _, authorizationKey, user := newForkTransactorWithKey(GinkgoT(), ctx, rpcClient)
		implementation, err := config.Base.Simple7702AccountImplementationAddress()
		Expect(err).NotTo(HaveOccurred())
		assertContractCode(GinkgoT(), ctx, ethClient, implementation, "Simple7702Account")

		chainID, err := config.Base.ChainID()
		Expect(err).NotTo(HaveOccurred())
		manager, err := eip7702.NewManager(ethClient, opts, authorizationKey, big.NewInt(int64(chainID)))
		Expect(err).NotTo(HaveOccurred())
		Expect(manager.AssertClean(ctx, user)).To(Succeed())

		delegateTx, err := manager.DelegateToSimple7702(ctx, config.Base)
		Expect(err).NotTo(HaveOccurred())
		delegateReceipt, err := bind.WaitMined(ctx, ethClient, delegateTx)
		Expect(err).NotTo(HaveOccurred())
		Expect(delegateReceipt.Status).To(Equal(uint64(types.ReceiptStatusSuccessful)))
		Expect(manager.AssertDelegatedTo(ctx, user, implementation)).To(Succeed())

		DeferCleanup(func() {
			cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cleanupCancel()

			clearTx, clearErr := manager.Clear(cleanupCtx)
			Expect(clearErr).NotTo(HaveOccurred())
			clearReceipt, clearErr := bind.WaitMined(cleanupCtx, ethClient, clearTx)
			Expect(clearErr).NotTo(HaveOccurred())
			Expect(clearReceipt.Status).To(Equal(uint64(types.ReceiptStatusSuccessful)))
			Expect(manager.AssertClean(cleanupCtx, user)).To(Succeed())
		})

		usdc, err := config.USDC.Address(config.Base)
		Expect(err).NotTo(HaveOccurred())
		assertContractCode(GinkgoT(), ctx, ethClient, usdc, "USDC")
		token, err := erc20.NewErc20(usdc, ethClient)
		Expect(err).NotTo(HaveOccurred())

		firstSpender := common.HexToAddress("0x00000000000000000000000000000000000000a1")
		secondSpender := common.HexToAddress("0x00000000000000000000000000000000000000b2")
		firstAmount := decimal.NewFromInt(1)
		secondAmount := decimal.NewFromInt(2)
		decimals, err := config.USDC.Decimals()
		Expect(err).NotTo(HaveOccurred())
		firstExpected := firstAmount.Shift(int32(decimals)).BigInt()
		secondExpected := secondAmount.Shift(int32(decimals)).BigInt()

		firstBefore, err := token.Allowance(nil, user, firstSpender)
		Expect(err).NotTo(HaveOccurred())
		Expect(firstBefore.Sign()).To(Equal(0))
		secondBefore, err := token.Allowance(nil, user, secondSpender)
		Expect(err).NotTo(HaveOccurred())
		Expect(secondBefore.Sign()).To(Equal(0))

		flow := defi.NewFlow(user, defi.WithChain(config.Base)).
			Add(sdkerc20.Approve(config.USDC, sdkerc20.AddressSpender(firstSpender), firstAmount)).
			Add(sdkerc20.Approve(config.USDC, sdkerc20.AddressSpender(secondSpender), secondAmount))
		runner := defi.NewRunner(ethClient, opts, config.Base)

		receipt, err := runner.Execute(ctx, flow, defi.ExecutionAtomicEOA)
		Expect(err).NotTo(HaveOccurred())
		Expect(receipt.Status).To(Equal(uint64(types.ReceiptStatusSuccessful)))

		firstAfter, err := token.Allowance(nil, user, firstSpender)
		Expect(err).NotTo(HaveOccurred())
		Expect(firstAfter.Cmp(firstExpected)).To(Equal(0))
		secondAfter, err := token.Allowance(nil, user, secondSpender)
		Expect(err).NotTo(HaveOccurred())
		Expect(secondAfter.Cmp(secondExpected)).To(Equal(0))
	})
})
