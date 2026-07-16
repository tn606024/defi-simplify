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

		market, supplyReserve, borrowReserve := loadBaseAaveReserves(GinkgoT(), ctx, ethClient)
		usdc := supplyReserve.Underlying().Address()
		weth := borrowReserve.Underlying().Address()
		pool := market.Pool()
		assertContractCode(GinkgoT(), ctx, ethClient, usdc, "USDC")
		assertContractCode(GinkgoT(), ctx, ethClient, weth, "WETH")
		assertContractCode(GinkgoT(), ctx, ethClient, pool, "Aave V3 Pool")

		client := sdkcontract.NewDefiClient(opts, ethClient, signer, config.Base)
		assertContractCode(GinkgoT(), ctx, ethClient, supplyReserve.AToken().Address(), "Base Aave aUSDC")
		assertContractCode(GinkgoT(), ctx, ethClient, borrowReserve.VariableDebtToken().Address(), "Base Aave variable debt WETH")

		supplyToken, err := binderc20.NewErc20(usdc, ethClient)
		Expect(err).NotTo(HaveOccurred())
		borrowToken, err := binderc20.NewErc20(weth, ethClient)
		Expect(err).NotTo(HaveOccurred())
		supplyAToken, err := binderc20.NewErc20(supplyReserve.AToken().Address(), ethClient)
		Expect(err).NotTo(HaveOccurred())
		borrowDebtToken, err := binderc20.NewErc20(borrowReserve.VariableDebtToken().Address(), ethClient)
		Expect(err).NotTo(HaveOccurred())

		supplyAmount := decimal.NewFromInt(10)
		borrowAmount := decimal.NewFromInt(1).Shift(-6)
		supplyDecimals := supplyReserve.Underlying().Decimals()
		borrowDecimals := borrowReserve.Underlying().Decimals()
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
			Add(sdkerc20.Approve(supplyReserve.Underlying(), aave.PoolSpender(market), supplyAmount)).
			Add(aave.Supply(supplyReserve, supplyAmount)).
			Add(aave.Borrow(borrowReserve, borrowAmount))
		runner := defi.NewRunner(ethClient, opts, config.Base)

		execution, err := runner.ExecuteWithResult(ctx, flow, defi.ExecutionAtomicEOA)
		Expect(err).NotTo(HaveOccurred())
		Expect(execution).NotTo(BeNil())
		receipt := execution.Receipt
		Expect(receipt.Status).To(Equal(uint64(types.ReceiptStatusSuccessful)))
		Expect(manager.AssertDelegatedTo(ctx, user, implementation)).To(Succeed())
		approvals := defi.EventsOf[*sdkerc20.ApprovalEvent](execution)
		supplies := defi.EventsOf[*aave.SupplyEvent](execution)
		borrows := defi.EventsOf[*aave.BorrowEvent](execution)
		Expect(approvals).To(HaveLen(1))
		Expect(supplies).To(HaveLen(1))
		Expect(borrows).To(HaveLen(1))
		approval := approvals[0]
		supply := supplies[0]
		borrow := borrows[0]
		Expect(approval.Token).To(Equal(usdc))
		Expect(approval.Owner).To(Equal(user))
		Expect(approval.Spender).To(Equal(pool))
		Expect(approval.Amount.Cmp(supplyAmountWei)).To(Equal(0))
		Expect(supply.Asset).To(Equal(usdc))
		Expect(supply.User).To(Equal(user))
		Expect(supply.OnBehalfOf).To(Equal(user))
		Expect(supply.Amount.Cmp(supplyAmountWei)).To(Equal(0))
		Expect(borrow.Asset).To(Equal(weth))
		Expect(borrow.User).To(Equal(user))
		Expect(borrow.OnBehalfOf).To(Equal(user))
		Expect(borrow.Amount.Cmp(borrowAmountWei)).To(Equal(0))
		Expect(borrow.InterestRateMode).To(Equal(aave.VariableInterestRateMode))
		Expect(borrow.BorrowRate.Sign()).To(Equal(1))
		Expect(approval.Metadata.LogIndex).To(BeNumerically("<", supply.Metadata.LogIndex))
		Expect(supply.Metadata.LogIndex).To(BeNumerically("<", borrow.Metadata.LogIndex))

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
