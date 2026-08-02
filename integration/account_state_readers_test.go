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
	txamount "github.com/tn606024/defi-simplify/amount"
	baseassets "github.com/tn606024/defi-simplify/assets/base"
	"github.com/tn606024/defi-simplify/client/account/eip7702"
	"github.com/tn606024/defi-simplify/config"
	sdkerc20 "github.com/tn606024/defi-simplify/erc20"
	sdkweth "github.com/tn606024/defi-simplify/weth"
)

var _ = Describe("Resolved account state readers", func() {
	It("reads ERC20 and Aave state after an EOA-native position is opened", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()

		ethClient := baseForkClient(GinkgoT())
		rpcClient := baseForkRPCClient(GinkgoT())
		requireAnvilFork(GinkgoT(), ctx, rpcClient)

		opts, authorizationKey, user := newForkTransactorWithKey(GinkgoT(), ctx, rpcClient)
		implementation := loadDefiSimplifyAccountIdentity(GinkgoT(), ctx, ethClient)
		chainID, err := config.Base.ChainID()
		Expect(err).NotTo(HaveOccurred())
		manager, err := eip7702.NewManager(
			ethClient,
			opts,
			authorizationKey,
			big.NewInt(int64(chainID)),
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(manager.AssertClean(ctx, user)).To(Succeed())
		delegateForkEOA(
			GinkgoT(),
			ctx,
			ethClient,
			manager,
			user,
			implementation.Address,
		)

		DeferCleanup(func() {
			cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cleanupCancel()

			opts.GasLimit = 0
			clearTx, clearErr := manager.Clear(cleanupCtx)
			Expect(clearErr).NotTo(HaveOccurred())
			clearReceipt, clearErr := bind.WaitMined(cleanupCtx, ethClient, clearTx)
			Expect(clearErr).NotTo(HaveOccurred())
			Expect(clearReceipt.Status).To(Equal(uint64(types.ReceiptStatusSuccessful)))
			Expect(manager.AssertClean(cleanupCtx, user)).To(Succeed())
		})

		market, usdcReserve, wethReserve := loadBaseAaveReserves(GinkgoT(), ctx, ethClient)
		wethAmount := decimal.RequireFromString("0.1")
		usdcAmount := decimal.NewFromInt(10)
		usdcAmountBaseUnits := usdcAmount.
			Shift(int32(usdcReserve.Underlying().Decimals())).
			BigInt()

		flow := defi.NewFlow(user, defi.WithChain(config.Base)).
			Add(sdkweth.Wrap(baseassets.WETH, txamount.Exact(wethAmount))).
			Add(sdkerc20.Approve(
				wethReserve.Underlying(),
				aave.PoolSpender(market),
				txamount.Exact(wethAmount),
			)).
			Add(aave.Supply(wethReserve, txamount.Exact(wethAmount))).
			Add(aave.Borrow(usdcReserve, txamount.Exact(usdcAmount)))
		result, err := defi.NewRunner(ethClient, opts, config.Base).Execute(ctx, flow)
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Receipt.Status).To(Equal(uint64(types.ReceiptStatusSuccessful)))

		wethState, err := sdkerc20.LoadAccountState(
			ctx,
			ethClient,
			wethReserve.Underlying(),
			user,
			market.Pool(),
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(wethState.Balance()).To(Equal(new(big.Int)))
		Expect(wethState.Allowance()).To(Equal(new(big.Int)))

		usdcState, err := sdkerc20.LoadAccountState(
			ctx,
			ethClient,
			usdcReserve.Underlying(),
			user,
			market.Pool(),
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(usdcState.Balance()).To(Equal(usdcAmountBaseUnits))
		Expect(usdcState.Allowance()).To(Equal(new(big.Int)))

		wethPosition, err := aave.LoadUserReserveState(ctx, ethClient, wethReserve, user)
		Expect(err).NotTo(HaveOccurred())
		Expect(wethPosition.CurrentATokenBalance().Sign()).To(Equal(1))
		Expect(wethPosition.CurrentVariableDebt()).To(Equal(new(big.Int)))
		Expect(wethPosition.UsageAsCollateralEnabled()).To(BeTrue())

		usdcPosition, err := aave.LoadUserReserveState(ctx, ethClient, usdcReserve, user)
		Expect(err).NotTo(HaveOccurred())
		Expect(usdcPosition.CurrentATokenBalance()).To(Equal(new(big.Int)))
		Expect(usdcPosition.CurrentVariableDebt().Sign()).To(Equal(1))
		Expect(usdcPosition.UsageAsCollateralEnabled()).To(BeFalse())

		Expect(wethState.BlockNumber()).To(BeNumerically(">", 0))
		Expect(wethState.BlockHash()).NotTo(BeZero())
		Expect(usdcState.BlockNumber()).To(Equal(wethState.BlockNumber()))
		Expect(usdcState.BlockHash()).To(Equal(wethState.BlockHash()))
		Expect(wethPosition.BlockNumber()).To(Equal(wethState.BlockNumber()))
		Expect(wethPosition.BlockHash()).To(Equal(wethState.BlockHash()))
		Expect(usdcPosition.BlockNumber()).To(Equal(wethState.BlockNumber()))
		Expect(usdcPosition.BlockHash()).To(Equal(wethState.BlockHash()))
	})
})
