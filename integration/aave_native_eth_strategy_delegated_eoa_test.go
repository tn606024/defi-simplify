//go:build integration

package integration

import (
	"context"
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
	"github.com/tn606024/defi-simplify/client/account/defisimplify7702"
	"github.com/tn606024/defi-simplify/client/account/eip7702"
	"github.com/tn606024/defi-simplify/config"
	sdkerc20 "github.com/tn606024/defi-simplify/erc20"
	"github.com/tn606024/defi-simplify/strategy"
	sdkweth "github.com/tn606024/defi-simplify/weth"
)

var _ = Describe("Aave native ETH strategy integration", func() {
	var (
		ctx          context.Context
		cancel       context.CancelFunc
		ethClient    *ethclient.Client
		rpcClient    *rpc.Client
		opts         *bind.TransactOpts
		user         common.Address
		manager      *eip7702.Manager
		market       aave.Market
		usdc         aave.Reserve
		wethReserve  aave.Reserve
		wethContract *binderc20.Erc20
	)

	BeforeEach(func() {
		ctx, cancel = context.WithTimeout(context.Background(), 120*time.Second)
		DeferCleanup(cancel)

		ethClient = baseForkClient(GinkgoT())
		rpcClient = baseForkRPCClient(GinkgoT())
		requireAnvilFork(GinkgoT(), ctx, rpcClient)

		forkOpts, _, authorizationKey, forkUser := newForkTransactorWithKey(
			GinkgoT(),
			ctx,
			rpcClient,
		)
		opts = forkOpts
		user = forkUser

		chainID, err := config.Base.ChainID()
		Expect(err).NotTo(HaveOccurred())
		manager, err = eip7702.NewManager(
			ethClient,
			opts,
			authorizationKey,
			big.NewInt(int64(chainID)),
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(manager.AssertClean(ctx, user)).To(Succeed())

		DeferCleanup(func() {
			cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cleanupCancel()

			state, stateErr := manager.State(cleanupCtx, user)
			Expect(stateErr).NotTo(HaveOccurred())
			if state.Clean() {
				return
			}
			opts.GasLimit = 0
			clearTx, clearErr := manager.Clear(cleanupCtx)
			Expect(clearErr).NotTo(HaveOccurred())
			clearReceipt, clearErr := bind.WaitMined(cleanupCtx, ethClient, clearTx)
			Expect(clearErr).NotTo(HaveOccurred())
			Expect(clearReceipt.Status).To(Equal(uint64(types.ReceiptStatusSuccessful)))
			Expect(manager.AssertClean(cleanupCtx, user)).To(Succeed())
		})

		market, usdc, wethReserve = loadBaseAaveReserves(GinkgoT(), ctx, ethClient)
		wethContract, err = binderc20.NewErc20(wethReserve.Underlying().Address(), ethClient)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("with Simple7702Account static execution", func() {
		BeforeEach(func() {
			implementation, err := config.Base.Simple7702AccountImplementationAddress()
			Expect(err).NotTo(HaveOccurred())
			assertContractCode(GinkgoT(), ctx, ethClient, implementation, "Simple7702Account")
			delegateForkEOA(ctx, ethClient, manager, user, implementation)
		})

		It("wraps and supplies native ETH as EOA-owned WETH collateral", func() {
			supplyAmount := decimal.RequireFromString("0.01")
			supplyAmountWei := decimalTokenAmount(wethReserve.Underlying(), supplyAmount)
			beforeETH, err := ethClient.BalanceAt(ctx, user, nil)
			Expect(err).NotTo(HaveOccurred())
			flow, err := strategy.AaveSupplyNativeETH(strategy.AaveSupplyNativeETHParams{
				Account: user,
				Reserve: wethReserve,
				Amount:  supplyAmount,
			})
			Expect(err).NotTo(HaveOccurred())

			result, err := defi.NewRunner(ethClient, opts, config.Base).
				ExecuteWithResult(ctx, flow, defi.ExecutionAtomicEOA)

			Expect(err).NotTo(HaveOccurred())
			Expect(result.Receipt.Status).To(Equal(uint64(types.ReceiptStatusSuccessful)))
			deposits := defi.EventsOf[*sdkweth.DepositEvent](result)
			approvals := defi.EventsOf[*sdkerc20.ApprovalEvent](result)
			supplies := defi.EventsOf[*aave.SupplyEvent](result)
			Expect(deposits).To(HaveLen(1))
			Expect(approvals).To(HaveLen(1))
			Expect(supplies).To(HaveLen(1))
			Expect(deposits[0].Account).To(Equal(user))
			Expect(deposits[0].Amount).To(Equal(supplyAmountWei))
			Expect(approvals[0].Owner).To(Equal(user))
			Expect(approvals[0].Spender).To(Equal(market.Pool()))
			Expect(approvals[0].Amount).To(Equal(supplyAmountWei))
			Expect(supplies[0].Asset).To(Equal(wethReserve.Underlying().Address()))
			Expect(supplies[0].User).To(Equal(user))
			Expect(supplies[0].OnBehalfOf).To(Equal(user))
			Expect(supplies[0].Amount).To(Equal(supplyAmountWei))
			Expect(deposits[0].Metadata.LogIndex).To(BeNumerically("<", approvals[0].Metadata.LogIndex))
			Expect(approvals[0].Metadata.LogIndex).To(BeNumerically("<", supplies[0].Metadata.LogIndex))

			aToken := mustTokenBinding(ethClient, wethReserve.AToken().Address())
			aTokenBalance, err := aToken.BalanceOf(&bind.CallOpts{Context: ctx}, user)
			Expect(err).NotTo(HaveOccurred())
			Expect(aTokenBalance.Sign()).To(Equal(1))
			afterETH, err := ethClient.BalanceAt(ctx, user, nil)
			Expect(err).NotTo(HaveOccurred())
			expectNativeBalanceAfterTransaction(
				beforeETH,
				afterETH,
				new(big.Int).Neg(supplyAmountWei),
				result.Receipt,
			)
		})
	})

	Context("with DefiSimplify7702Account dynamic execution", func() {
		BeforeEach(func() {
			deployment, err := defisimplify7702.DeploymentForChain(config.Base)
			Expect(err).NotTo(HaveOccurred())
			accountDeployment, err := deployment.Contract(defisimplify7702.AccountContract)
			Expect(err).NotTo(HaveOccurred())
			implementation := accountDeployment.RuntimeIdentity
			implementationCode, err := ethClient.CodeAt(ctx, implementation.Address, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(implementation.VerifyRuntimeCode(implementationCode)).To(Succeed())
			delegateForkEOA(ctx, ethClient, manager, user, implementation.Address)
		})

		It("borrows WETH debt as the EOA and returns only the borrow as native ETH", func() {
			collateralAmount := decimal.NewFromInt(10)
			borrowAmount := decimal.RequireFromString("0.0001")
			preExistingWETH := decimal.RequireFromString("0.0002")
			borrowAmountWei := decimalTokenAmount(wethReserve.Underlying(), borrowAmount)
			seedAaveUSDCollateral(
				ctx,
				rpcClient,
				ethClient,
				opts,
				user,
				market,
				usdc,
				collateralAmount,
			)
			seedWETH(ctx, ethClient, opts, user, preExistingWETH, defi.ExecutionEOA)

			beforeWETH, err := wethContract.BalanceOf(&bind.CallOpts{Context: ctx}, user)
			Expect(err).NotTo(HaveOccurred())
			beforeETH, err := ethClient.BalanceAt(ctx, user, nil)
			Expect(err).NotTo(HaveOccurred())
			flow, err := strategy.AaveBorrowNativeETH(strategy.AaveBorrowNativeETHParams{
				Account: user,
				Reserve: wethReserve,
				Amount:  borrowAmount,
			})
			Expect(err).NotTo(HaveOccurred())

			result, err := defi.NewRunner(ethClient, opts, config.Base).
				ExecuteWithResult(ctx, flow, defi.ExecutionDynamicEOA)

			Expect(err).NotTo(HaveOccurred())
			Expect(result.Receipt.Status).To(Equal(uint64(types.ReceiptStatusSuccessful)))
			borrows := defi.EventsOf[*aave.BorrowEvent](result)
			withdrawals := defi.EventsOf[*sdkweth.WithdrawalEvent](result)
			Expect(borrows).To(HaveLen(1))
			Expect(withdrawals).To(HaveLen(1))
			Expect(borrows[0].Asset).To(Equal(wethReserve.Underlying().Address()))
			Expect(borrows[0].User).To(Equal(user))
			Expect(borrows[0].OnBehalfOf).To(Equal(user))
			Expect(borrows[0].Amount).To(Equal(borrowAmountWei))
			Expect(withdrawals[0].Account).To(Equal(user))
			Expect(withdrawals[0].Amount).To(Equal(borrowAmountWei))
			Expect(borrows[0].Metadata.LogIndex).To(BeNumerically("<", withdrawals[0].Metadata.LogIndex))

			afterWETH, err := wethContract.BalanceOf(&bind.CallOpts{Context: ctx}, user)
			Expect(err).NotTo(HaveOccurred())
			Expect(afterWETH).To(Equal(beforeWETH))
			afterETH, err := ethClient.BalanceAt(ctx, user, nil)
			Expect(err).NotTo(HaveOccurred())
			expectNativeBalanceAfterTransaction(beforeETH, afterETH, borrowAmountWei, result.Receipt)
			debtToken := mustTokenBinding(ethClient, wethReserve.VariableDebtToken().Address())
			debtBalance, err := debtToken.BalanceOf(&bind.CallOpts{Context: ctx}, user)
			Expect(err).NotTo(HaveOccurred())
			Expect(debtBalance.Cmp(borrowAmountWei)).To(BeNumerically(">=", 0))
		})

		It("withdraws an exact collateral amount as native ETH without sweeping existing WETH", func() {
			collateralAmount := decimal.RequireFromString("0.01")
			withdrawAmount := decimal.RequireFromString("0.002")
			preExistingWETH := decimal.RequireFromString("0.0002")
			withdrawAmountWei := decimalTokenAmount(wethReserve.Underlying(), withdrawAmount)
			seedAaveWETHCollateral(
				ctx,
				ethClient,
				opts,
				user,
				market,
				wethReserve,
				collateralAmount,
			)
			seedWETH(ctx, ethClient, opts, user, preExistingWETH, defi.ExecutionEOA)

			beforeWETH, err := wethContract.BalanceOf(&bind.CallOpts{Context: ctx}, user)
			Expect(err).NotTo(HaveOccurred())
			beforeETH, err := ethClient.BalanceAt(ctx, user, nil)
			Expect(err).NotTo(HaveOccurred())
			aToken := mustTokenBinding(ethClient, wethReserve.AToken().Address())
			beforeAToken, err := aToken.BalanceOf(&bind.CallOpts{Context: ctx}, user)
			Expect(err).NotTo(HaveOccurred())
			flow, err := strategy.AaveWithdrawNativeETH(strategy.AaveWithdrawNativeETHParams{
				Account: user,
				Reserve: wethReserve,
				Amount:  withdrawAmount,
			})
			Expect(err).NotTo(HaveOccurred())

			result, err := defi.NewRunner(ethClient, opts, config.Base).
				ExecuteWithResult(ctx, flow, defi.ExecutionDynamicEOA)

			Expect(err).NotTo(HaveOccurred())
			Expect(result.Receipt.Status).To(Equal(uint64(types.ReceiptStatusSuccessful)))
			aaveWithdrawals := defi.EventsOf[*aave.WithdrawEvent](result)
			wethWithdrawals := defi.EventsOf[*sdkweth.WithdrawalEvent](result)
			Expect(aaveWithdrawals).To(HaveLen(1))
			Expect(wethWithdrawals).To(HaveLen(1))
			Expect(aaveWithdrawals[0].Asset).To(Equal(wethReserve.Underlying().Address()))
			Expect(aaveWithdrawals[0].User).To(Equal(user))
			Expect(aaveWithdrawals[0].To).To(Equal(user))
			Expect(aaveWithdrawals[0].Amount).To(Equal(withdrawAmountWei))
			Expect(wethWithdrawals[0].Account).To(Equal(user))
			Expect(wethWithdrawals[0].Amount).To(Equal(withdrawAmountWei))
			Expect(aaveWithdrawals[0].Metadata.LogIndex).To(BeNumerically("<", wethWithdrawals[0].Metadata.LogIndex))

			afterWETH, err := wethContract.BalanceOf(&bind.CallOpts{Context: ctx}, user)
			Expect(err).NotTo(HaveOccurred())
			Expect(afterWETH).To(Equal(beforeWETH))
			afterETH, err := ethClient.BalanceAt(ctx, user, nil)
			Expect(err).NotTo(HaveOccurred())
			expectNativeBalanceAfterTransaction(beforeETH, afterETH, withdrawAmountWei, result.Receipt)
			afterAToken, err := aToken.BalanceOf(&bind.CallOpts{Context: ctx}, user)
			Expect(err).NotTo(HaveOccurred())
			Expect(afterAToken.Cmp(beforeAToken)).To(Equal(-1))
			Expect(afterAToken.Sign()).To(Equal(1))
		})

		It("withdraws all collateral as native ETH without sweeping existing WETH", func() {
			collateralAmount := decimal.RequireFromString("0.003")
			preExistingWETH := decimal.RequireFromString("0.0002")
			seedAaveWETHCollateral(
				ctx,
				ethClient,
				opts,
				user,
				market,
				wethReserve,
				collateralAmount,
			)
			seedWETH(ctx, ethClient, opts, user, preExistingWETH, defi.ExecutionEOA)

			beforeWETH, err := wethContract.BalanceOf(&bind.CallOpts{Context: ctx}, user)
			Expect(err).NotTo(HaveOccurred())
			beforeETH, err := ethClient.BalanceAt(ctx, user, nil)
			Expect(err).NotTo(HaveOccurred())
			flow, err := strategy.AaveWithdrawAllNativeETH(strategy.AaveWithdrawAllNativeETHParams{
				Account: user,
				Reserve: wethReserve,
			})
			Expect(err).NotTo(HaveOccurred())

			result, err := defi.NewRunner(ethClient, opts, config.Base).
				ExecuteWithResult(ctx, flow, defi.ExecutionDynamicEOA)

			Expect(err).NotTo(HaveOccurred())
			Expect(result.Receipt.Status).To(Equal(uint64(types.ReceiptStatusSuccessful)))
			aaveWithdrawals := defi.EventsOf[*aave.WithdrawEvent](result)
			wethWithdrawals := defi.EventsOf[*sdkweth.WithdrawalEvent](result)
			Expect(aaveWithdrawals).To(HaveLen(1))
			Expect(wethWithdrawals).To(HaveLen(1))
			Expect(aaveWithdrawals[0].Asset).To(Equal(wethReserve.Underlying().Address()))
			Expect(aaveWithdrawals[0].User).To(Equal(user))
			Expect(aaveWithdrawals[0].To).To(Equal(user))
			Expect(aaveWithdrawals[0].Amount.Sign()).To(Equal(1))
			Expect(wethWithdrawals[0].Account).To(Equal(user))
			Expect(wethWithdrawals[0].Amount).To(Equal(aaveWithdrawals[0].Amount))
			Expect(aaveWithdrawals[0].Metadata.LogIndex).To(BeNumerically("<", wethWithdrawals[0].Metadata.LogIndex))

			afterWETH, err := wethContract.BalanceOf(&bind.CallOpts{Context: ctx}, user)
			Expect(err).NotTo(HaveOccurred())
			Expect(afterWETH).To(Equal(beforeWETH))
			afterETH, err := ethClient.BalanceAt(ctx, user, nil)
			Expect(err).NotTo(HaveOccurred())
			expectNativeBalanceAfterTransaction(
				beforeETH,
				afterETH,
				aaveWithdrawals[0].Amount,
				result.Receipt,
			)
			aToken := mustTokenBinding(ethClient, wethReserve.AToken().Address())
			aTokenBalance, err := aToken.BalanceOf(&bind.CallOpts{Context: ctx}, user)
			Expect(err).NotTo(HaveOccurred())
			Expect(aTokenBalance.Sign()).To(Equal(0))
		})
	})
})

func seedAaveUSDCollateral(
	ctx context.Context,
	rpcClient *rpc.Client,
	ethClient *ethclient.Client,
	opts *bind.TransactOpts,
	user common.Address,
	market aave.Market,
	reserve aave.Reserve,
	value decimal.Decimal,
) {
	GinkgoHelper()
	valueWei := decimalTokenAmount(reserve.Underlying(), value)
	Expect(fundBaseUSDCFromHolder(ctx, rpcClient, ethClient, user, valueWei)).To(Succeed())
	executeDirectForkStep(ctx, ethClient, opts, user, sdkerc20.Approve(
		reserve.Underlying(),
		aave.PoolSpender(market),
		txamount.Exact(value),
	))
	executeDirectForkStep(ctx, ethClient, opts, user, aave.Supply(
		reserve,
		txamount.Exact(value),
	))
}

func seedAaveWETHCollateral(
	ctx context.Context,
	ethClient *ethclient.Client,
	opts *bind.TransactOpts,
	user common.Address,
	market aave.Market,
	reserve aave.Reserve,
	value decimal.Decimal,
) {
	GinkgoHelper()
	executeDirectForkStep(ctx, ethClient, opts, user, sdkweth.Wrap(
		reserve.Underlying().Ref(),
		txamount.Exact(value),
	))
	executeDirectForkStep(ctx, ethClient, opts, user, sdkerc20.Approve(
		reserve.Underlying(),
		aave.PoolSpender(market),
		txamount.Exact(value),
	))
	executeDirectForkStep(ctx, ethClient, opts, user, aave.Supply(
		reserve,
		txamount.Exact(value),
	))
}

func executeDirectForkStep(
	ctx context.Context,
	ethClient *ethclient.Client,
	opts *bind.TransactOpts,
	user common.Address,
	step defi.FlowStep,
) {
	GinkgoHelper()
	flow := defi.NewFlow(user, defi.WithChain(config.Base)).Add(step)
	result, err := defi.NewRunner(ethClient, opts, config.Base).
		ExecuteWithResult(ctx, flow, defi.ExecutionEOA)
	Expect(err).NotTo(HaveOccurred())
	Expect(result.Receipt.Status).To(Equal(uint64(types.ReceiptStatusSuccessful)))
}

func expectNativeBalanceAfterTransaction(
	before *big.Int,
	after *big.Int,
	valueDelta *big.Int,
	receipt *types.Receipt,
) {
	GinkgoHelper()
	Expect(receipt.EffectiveGasPrice).NotTo(BeNil())
	gasCost := new(big.Int).Mul(
		new(big.Int).SetUint64(receipt.GasUsed),
		receipt.EffectiveGasPrice,
	)
	expected := new(big.Int).Add(new(big.Int).Set(before), valueDelta)
	expected.Sub(expected, gasCost)
	Expect(after).To(Equal(expected))
}
