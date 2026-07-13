//go:build integration

package integration

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/shopspring/decimal"
	defi "github.com/tn606024/defi-simplify"
	"github.com/tn606024/defi-simplify/aave"
	binderc20 "github.com/tn606024/defi-simplify/bind/erc20"
	"github.com/tn606024/defi-simplify/client/account/eip7702"
	sdkcontract "github.com/tn606024/defi-simplify/client/contract"
	"github.com/tn606024/defi-simplify/config"
	sdkerc20 "github.com/tn606024/defi-simplify/erc20"
)

var _ = Describe("Aave Flow ExecutionAtomicEOA integration", func() {
	It("supplies USDC and borrows WETH with an EOA-owned Aave position", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()

		ethClient := baseForkClient(GinkgoT())
		rpcClient := baseForkRPCClient(GinkgoT())
		requireAnvilFork(GinkgoT(), ctx, rpcClient)

		opts, signer, authorizationKey, user := newForkTransactorWithKey(GinkgoT(), ctx, rpcClient)
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
		weth, err := config.WETH.Address(config.Base)
		Expect(err).NotTo(HaveOccurred())
		pool, err := config.Base.AaveV3PoolAddress()
		Expect(err).NotTo(HaveOccurred())
		assertContractCode(GinkgoT(), ctx, ethClient, usdc, "USDC")
		assertContractCode(GinkgoT(), ctx, ethClient, weth, "WETH")
		assertContractCode(GinkgoT(), ctx, ethClient, pool, "Aave V3 Pool")

		client := sdkcontract.NewDefiClient(opts, ethClient, signer, config.Base)
		supplyReserve, err := client.Aave.GetReserveData(ctx, config.USDC)
		Expect(err).NotTo(HaveOccurred())
		borrowReserve, err := client.Aave.GetReserveData(ctx, config.WETH)
		Expect(err).NotTo(HaveOccurred())
		assertContractCode(GinkgoT(), ctx, ethClient, supplyReserve.ATokenAddress, "Base Aave aUSDC")
		assertContractCode(GinkgoT(), ctx, ethClient, borrowReserve.VariableDebtTokenAddress, "Base Aave variable debt WETH")

		supplyToken, err := binderc20.NewErc20(usdc, ethClient)
		Expect(err).NotTo(HaveOccurred())
		borrowToken, err := binderc20.NewErc20(weth, ethClient)
		Expect(err).NotTo(HaveOccurred())
		supplyAToken, err := binderc20.NewErc20(supplyReserve.ATokenAddress, ethClient)
		Expect(err).NotTo(HaveOccurred())
		borrowDebtToken, err := binderc20.NewErc20(borrowReserve.VariableDebtTokenAddress, ethClient)
		Expect(err).NotTo(HaveOccurred())

		supplyAmount := decimal.NewFromInt(10)
		borrowAmount := decimal.NewFromInt(1).Shift(-6)
		supplyDecimals, err := config.USDC.Decimals()
		Expect(err).NotTo(HaveOccurred())
		borrowDecimals, err := config.WETH.Decimals()
		Expect(err).NotTo(HaveOccurred())
		supplyAmountWei := supplyAmount.Shift(int32(supplyDecimals)).BigInt()
		borrowAmountWei := borrowAmount.Shift(int32(borrowDecimals)).BigInt()
		Expect(fundBaseUSDCFromHolder(ctx, rpcClient, ethClient, user, supplyAmountWei)).To(Succeed())

		beforeUserSupply, err := supplyToken.BalanceOf(nil, user)
		Expect(err).NotTo(HaveOccurred())
		beforeUserBorrow, err := borrowToken.BalanceOf(nil, user)
		Expect(err).NotTo(HaveOccurred())
		beforeAToken, err := supplyAToken.BalanceOf(nil, user)
		Expect(err).NotTo(HaveOccurred())
		beforeDebtToken, err := borrowDebtToken.BalanceOf(nil, user)
		Expect(err).NotTo(HaveOccurred())
		beforeImplementationBorrow, err := borrowToken.BalanceOf(nil, implementation)
		Expect(err).NotTo(HaveOccurred())
		beforeSupplyPosition, err := client.Aave.GetUserReserveData(ctx, usdc)
		Expect(err).NotTo(HaveOccurred())
		beforeBorrowPosition, err := client.Aave.GetUserReserveData(ctx, weth)
		Expect(err).NotTo(HaveOccurred())
		beforeAccountData, err := client.Aave.GetUserAccountData(ctx)
		Expect(err).NotTo(HaveOccurred())

		flow := defi.NewFlow(user, defi.WithChain(config.Base)).
			Add(sdkerc20.Approve(config.USDC, aave.PoolSpender(), supplyAmount)).
			Add(aave.Supply(config.USDC, supplyAmount)).
			Add(aave.Borrow(config.WETH, borrowAmount))
		expectedExecution, err := aave.NewExecutionExpectation(
			config.Base,
			user,
			config.USDC,
			supplyAmount,
			config.WETH,
			borrowAmount,
		)
		Expect(err).NotTo(HaveOccurred())
		runner := defi.NewRunner(ethClient, opts, config.Base)

		receipt, err := runner.Execute(ctx, flow, defi.ExecutionAtomicEOA)
		Expect(err).NotTo(HaveOccurred())
		Expect(receipt.Status).To(Equal(uint64(types.ReceiptStatusSuccessful)))
		Expect(manager.AssertDelegatedTo(ctx, user, implementation)).To(Succeed())
		summary, err := aave.ParseExecutionReceipt(receipt, expectedExecution)
		Expect(err).NotTo(HaveOccurred())
		Expect(summary.TransactionHash).To(Equal(receipt.TxHash))
		Expect(summary.Approval.Token).To(Equal(usdc))
		Expect(summary.Approval.Owner).To(Equal(user))
		Expect(summary.Approval.Spender).To(Equal(pool))
		Expect(summary.Approval.Amount.Cmp(supplyAmountWei)).To(Equal(0))
		Expect(summary.Supply.Asset).To(Equal(usdc))
		Expect(summary.Supply.User).To(Equal(user))
		Expect(summary.Supply.OnBehalfOf).To(Equal(user))
		Expect(summary.Supply.Amount.Cmp(supplyAmountWei)).To(Equal(0))
		Expect(summary.Borrow.Asset).To(Equal(weth))
		Expect(summary.Borrow.User).To(Equal(user))
		Expect(summary.Borrow.OnBehalfOf).To(Equal(user))
		Expect(summary.Borrow.Amount.Cmp(borrowAmountWei)).To(Equal(0))
		Expect(summary.Borrow.InterestRateMode).To(Equal(uint8(2)))
		Expect(summary.Borrow.BorrowRate.Sign()).To(Equal(1))
		Expect(summary.Approval.LogIndex).To(BeNumerically("<", summary.Supply.LogIndex))
		Expect(summary.Supply.LogIndex).To(BeNumerically("<", summary.Borrow.LogIndex))

		afterUserSupply, err := supplyToken.BalanceOf(nil, user)
		Expect(err).NotTo(HaveOccurred())
		Expect(new(big.Int).Sub(beforeUserSupply, afterUserSupply).Cmp(supplyAmountWei)).To(Equal(0))

		afterUserBorrow, err := borrowToken.BalanceOf(nil, user)
		Expect(err).NotTo(HaveOccurred())
		Expect(new(big.Int).Sub(afterUserBorrow, beforeUserBorrow).Cmp(borrowAmountWei)).To(Equal(0))

		afterImplementationBorrow, err := borrowToken.BalanceOf(nil, implementation)
		Expect(err).NotTo(HaveOccurred())
		Expect(afterImplementationBorrow.Cmp(beforeImplementationBorrow)).To(Equal(0))

		afterAToken, err := supplyAToken.BalanceOf(nil, user)
		Expect(err).NotTo(HaveOccurred())
		Expect(afterAToken.Cmp(beforeAToken)).To(Equal(1))
		afterDebtToken, err := borrowDebtToken.BalanceOf(nil, user)
		Expect(err).NotTo(HaveOccurred())
		Expect(afterDebtToken.Cmp(beforeDebtToken)).To(Equal(1))

		afterSupplyPosition, err := client.Aave.GetUserReserveData(ctx, usdc)
		Expect(err).NotTo(HaveOccurred())
		Expect(afterSupplyPosition.CurrentATokenBalance.Cmp(beforeSupplyPosition.CurrentATokenBalance)).To(Equal(1))
		Expect(afterSupplyPosition.CurrentATokenBalance.Cmp(afterAToken)).To(Equal(0))
		Expect(afterSupplyPosition.UsageAsCollateralEnabled).To(BeTrue())

		afterBorrowPosition, err := client.Aave.GetUserReserveData(ctx, weth)
		Expect(err).NotTo(HaveOccurred())
		Expect(afterBorrowPosition.CurrentVariableDebt.Cmp(beforeBorrowPosition.CurrentVariableDebt)).To(Equal(1))
		Expect(afterBorrowPosition.CurrentVariableDebt.Cmp(afterDebtToken)).To(Equal(0))

		afterAccountData, err := client.Aave.GetUserAccountData(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(afterAccountData.TotalCollateralBase.Cmp(beforeAccountData.TotalCollateralBase)).To(Equal(1))
		Expect(afterAccountData.TotalDebtBase.Cmp(beforeAccountData.TotalDebtBase)).To(Equal(1))
		Expect(afterAccountData.HealthFactor.Cmp(big.NewInt(1_000_000_000_000_000_000))).To(Equal(1))
	})
})
