//go:build integration

package integration

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/shopspring/decimal"
	defi "github.com/tn606024/defi-simplify"
	"github.com/tn606024/defi-simplify/aave"
	txamount "github.com/tn606024/defi-simplify/amount"
	binderc20 "github.com/tn606024/defi-simplify/bind/erc20"
	"github.com/tn606024/defi-simplify/client/account/eip7702"
	"github.com/tn606024/defi-simplify/config"
	sdkerc20 "github.com/tn606024/defi-simplify/erc20"
	"github.com/tn606024/defi-simplify/token"
	sdkweth "github.com/tn606024/defi-simplify/weth"
)

var _ = Describe("Extended Aave FlowStep integration", func() {
	var (
		ctx              context.Context
		cancel           context.CancelFunc
		ethClient        *ethclient.Client
		rpcClient        *rpc.Client
		opts             *bind.TransactOpts
		authorizationKey *ecdsa.PrivateKey
		user             common.Address
		manager          *eip7702.Manager
		implementation   common.Address
	)

	BeforeEach(func() {
		ctx, cancel = context.WithTimeout(context.Background(), 120*time.Second)
		DeferCleanup(cancel)

		ethClient = baseForkClient(GinkgoT())
		rpcClient = baseForkRPCClient(GinkgoT())
		requireAnvilFork(GinkgoT(), ctx, rpcClient)
		opts, authorizationKey, user = newForkTransactorWithKey(GinkgoT(), ctx, rpcClient)

		implementation = loadDefiSimplifyAccountIdentity(GinkgoT(), ctx, ethClient).Address
		chainID, err := config.Base.ChainID()
		Expect(err).NotTo(HaveOccurred())
		manager, err = eip7702.NewManager(ethClient, opts, authorizationKey, big.NewInt(int64(chainID)))
		Expect(err).NotTo(HaveOccurred())
		Expect(manager.AssertClean(ctx, user)).To(Succeed())

		delegateTx, err := manager.Delegate(ctx, implementation)
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
	})

	It("executes approve, supply, repay, and withdraw as one EOA-native flow", func() {
		market, usdc, weth := loadBaseAaveReserves(GinkgoT(), ctx, ethClient)
		supplyAmount := decimal.NewFromInt(10)
		borrowAmount := decimal.NewFromInt(1).Shift(-6)
		withdrawAmount := decimal.NewFromInt(1)
		supplyAmountWei := decimalTokenAmount(usdc.Underlying(), supplyAmount)
		Expect(fundBaseUSDCFromHolder(ctx, rpcClient, ethClient, user, supplyAmountWei)).To(Succeed())

		flow := defi.NewFlow(user, defi.WithChain(config.Base)).
			Add(sdkerc20.Approve(usdc.Underlying(), aave.PoolSpender(market), txamount.Exact(supplyAmount))).
			Add(aave.Supply(usdc, txamount.Exact(supplyAmount))).
			Add(aave.Borrow(weth, txamount.Exact(borrowAmount))).
			Add(sdkerc20.Approve(weth.Underlying(), aave.PoolSpender(market), txamount.Exact(borrowAmount))).
			Add(aave.Repay(weth, txamount.Exact(borrowAmount))).
			Add(aave.Withdraw(usdc, txamount.Exact(withdrawAmount)))

		result, err := defi.NewRunner(ethClient, opts, config.Base).
			Execute(ctx, flow)

		Expect(err).NotTo(HaveOccurred())
		Expect(result.Receipt.Status).To(Equal(uint64(types.ReceiptStatusSuccessful)))
		Expect(defi.EventsOf[*aave.SupplyEvent](result)).To(HaveLen(1))
		Expect(defi.EventsOf[*aave.BorrowEvent](result)).To(HaveLen(1))
		Expect(defi.EventsOf[*aave.RepayEvent](result)).To(HaveLen(1))
		Expect(defi.EventsOf[*aave.WithdrawEvent](result)).To(HaveLen(1))
	})

	It("repays and withdraws entire positions using Aave sentinel amounts", func() {
		market, usdc, weth := loadBaseAaveReserves(GinkgoT(), ctx, ethClient)
		depositAmount := decimal.RequireFromString("0.01")
		borrowAmount := decimal.NewFromInt(1)
		repaymentApproval := decimal.NewFromInt(2)
		depositAmountWei := decimalTokenAmount(weth.Underlying(), depositAmount)
		borrowAmountWei := decimalTokenAmount(usdc.Underlying(), borrowAmount)
		repaymentApprovalWei := decimalTokenAmount(usdc.Underlying(), repaymentApproval)
		maxAmount := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))
		Expect(fundBaseUSDCFromHolder(ctx, rpcClient, ethClient, user, borrowAmountWei)).To(Succeed())

		flow := defi.NewFlow(user, defi.WithChain(config.Base)).
			Add(sdkweth.Wrap(weth.Underlying().Ref(), txamount.Exact(depositAmount))).
			Add(sdkerc20.Approve(
				weth.Underlying(),
				aave.PoolSpender(market),
				txamount.Exact(depositAmount),
			)).
			Add(aave.Supply(weth, txamount.Exact(depositAmount))).
			Add(aave.Borrow(usdc, txamount.Exact(borrowAmount))).
			Add(sdkerc20.Approve(usdc.Underlying(), aave.PoolSpender(market), txamount.Exact(repaymentApproval))).
			Add(aave.RepayAll(usdc)).
			Add(aave.WithdrawAll(weth))

		result, err := defi.NewRunner(ethClient, opts, config.Base).
			Execute(ctx, flow)

		Expect(err).NotTo(HaveOccurred())
		Expect(result.Receipt.Status).To(Equal(uint64(types.ReceiptStatusSuccessful)))
		repayments := defi.EventsOf[*aave.RepayEvent](result)
		withdrawals := defi.EventsOf[*aave.WithdrawEvent](result)
		Expect(repayments).To(HaveLen(1))
		Expect(repayments[0].Asset).To(Equal(usdc.Underlying().Address()))
		Expect(repayments[0].User).To(Equal(user))
		Expect(repayments[0].Repayer).To(Equal(user))
		Expect(repayments[0].Amount.Cmp(borrowAmountWei)).To(BeNumerically(">=", 0))
		Expect(repayments[0].Amount.Cmp(repaymentApprovalWei)).To(BeNumerically("<", 0))
		Expect(repayments[0].Amount).NotTo(Equal(maxAmount))
		Expect(withdrawals).To(HaveLen(1))
		Expect(withdrawals[0].Asset).To(Equal(weth.Underlying().Address()))
		Expect(withdrawals[0].User).To(Equal(user))
		Expect(withdrawals[0].To).To(Equal(user))
		Expect(withdrawals[0].Amount.Sign()).To(Equal(1))
		Expect(withdrawals[0].Amount.Cmp(depositAmountWei)).To(BeNumerically("<=", 0))
		Expect(withdrawals[0].Amount).NotTo(Equal(maxAmount))
		Expect(repayments[0].Metadata.LogIndex).To(BeNumerically("<", withdrawals[0].Metadata.LogIndex))

		debtTokenBinding, err := binderc20.NewErc20(usdc.VariableDebtToken().Address(), ethClient)
		Expect(err).NotTo(HaveOccurred())
		debtBalance, err := debtTokenBinding.BalanceOf(&bind.CallOpts{Context: ctx}, user)
		Expect(err).NotTo(HaveOccurred())
		Expect(debtBalance.Sign()).To(Equal(0))

		aTokenBinding, err := binderc20.NewErc20(weth.AToken().Address(), ethClient)
		Expect(err).NotTo(HaveOccurred())
		aTokenBalance, err := aTokenBinding.BalanceOf(&bind.CallOpts{Context: ctx}, user)
		Expect(err).NotTo(HaveOccurred())
		Expect(aTokenBalance.Sign()).To(Equal(0))
	})
})

func decimalTokenAmount(asset token.Token, amount decimal.Decimal) *big.Int {
	return amount.Shift(int32(asset.Decimals())).BigInt()
}
