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
	txamount "github.com/tn606024/defi-simplify/amount"
	"github.com/tn606024/defi-simplify/assets/base"
	binderc20 "github.com/tn606024/defi-simplify/bind/erc20"
	"github.com/tn606024/defi-simplify/client/account/defisimplify7702"
	"github.com/tn606024/defi-simplify/client/account/eip7702"
	"github.com/tn606024/defi-simplify/config"
	sdkweth "github.com/tn606024/defi-simplify/weth"
)

var _ = Describe("WETH delegated EOA flows", func() {
	var (
		ctx          context.Context
		cancel       context.CancelFunc
		ethClient    *ethclient.Client
		rpcClient    *rpc.Client
		opts         *bind.TransactOpts
		user         common.Address
		manager      *eip7702.Manager
		wethContract *binderc20.Erc20
	)

	BeforeEach(func() {
		ctx, cancel = context.WithTimeout(context.Background(), 60*time.Second)
		DeferCleanup(cancel)

		ethClient = baseForkClient(GinkgoT())
		rpcClient = baseForkRPCClient(GinkgoT())
		requireAnvilFork(GinkgoT(), ctx, rpcClient)

		forkOpts, authorizationKey, forkUser := newForkTransactorWithKey(
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

		wethContract, err = binderc20.NewErc20(base.WETH.Address(), ethClient)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("with Simple7702Account static execution", func() {
		BeforeEach(func() {
			implementation, err := config.Base.Simple7702AccountImplementationAddress()
			Expect(err).NotTo(HaveOccurred())
			assertContractCode(GinkgoT(), ctx, ethClient, implementation, "Simple7702Account")
			delegateForkEOA(ctx, ethClient, manager, user, implementation)
		})

		It("wraps exact native ETH from the delegated EOA", func() {
			value := decimal.RequireFromString("0.2")
			valueWei := decimal.RequireFromString("0.2e18").BigInt()
			beforeWETH, err := wethContract.BalanceOf(&bind.CallOpts{Context: ctx}, user)
			Expect(err).NotTo(HaveOccurred())
			beforeETH, err := ethClient.BalanceAt(ctx, user, nil)
			Expect(err).NotTo(HaveOccurred())

			flow := defi.NewFlow(user, defi.WithChain(config.Base)).
				Add(sdkweth.Wrap(base.WETH, txamount.Exact(value)))

			result, err := defi.NewRunner(ethClient, opts, config.Base).
				ExecuteWithResult(ctx, flow, defi.ExecutionAtomicEOA)

			Expect(err).NotTo(HaveOccurred())
			Expect(result.Receipt.Status).To(Equal(uint64(types.ReceiptStatusSuccessful)))
			deposits := defi.EventsOf[*sdkweth.DepositEvent](result)
			Expect(deposits).To(HaveLen(1))
			Expect(deposits[0].Contract).To(Equal(base.WETH.Address()))
			Expect(deposits[0].Account).To(Equal(user))
			Expect(deposits[0].Amount).To(Equal(valueWei))

			afterWETH, err := wethContract.BalanceOf(&bind.CallOpts{Context: ctx}, user)
			Expect(err).NotTo(HaveOccurred())
			Expect(new(big.Int).Sub(afterWETH, beforeWETH)).To(Equal(valueWei))
			afterETH, err := ethClient.BalanceAt(ctx, user, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(afterETH.Cmp(beforeETH)).To(Equal(-1))
		})

		It("unwraps exact WETH and returns native ETH to the delegated EOA", func() {
			value := decimal.RequireFromString("0.25")
			valueWei := decimal.RequireFromString("0.25e18").BigInt()
			seedWETH(ctx, ethClient, opts, user, value, defi.ExecutionAtomicEOA)

			beforeWETH, err := wethContract.BalanceOf(&bind.CallOpts{Context: ctx}, user)
			Expect(err).NotTo(HaveOccurred())
			beforeETH, err := ethClient.BalanceAt(ctx, user, nil)
			Expect(err).NotTo(HaveOccurred())

			flow := defi.NewFlow(user, defi.WithChain(config.Base)).
				Add(sdkweth.Unwrap(base.WETH, txamount.Exact(value)))

			result, err := defi.NewRunner(ethClient, opts, config.Base).
				ExecuteWithResult(ctx, flow, defi.ExecutionAtomicEOA)

			Expect(err).NotTo(HaveOccurred())
			Expect(result.Receipt.Status).To(Equal(uint64(types.ReceiptStatusSuccessful)))
			withdrawals := defi.EventsOf[*sdkweth.WithdrawalEvent](result)
			Expect(withdrawals).To(HaveLen(1))
			Expect(withdrawals[0].Contract).To(Equal(base.WETH.Address()))
			Expect(withdrawals[0].Account).To(Equal(user))
			Expect(withdrawals[0].Amount).To(Equal(valueWei))

			afterWETH, err := wethContract.BalanceOf(&bind.CallOpts{Context: ctx}, user)
			Expect(err).NotTo(HaveOccurred())
			Expect(new(big.Int).Sub(beforeWETH, afterWETH)).To(Equal(valueWei))
			afterETH, err := ethClient.BalanceAt(ctx, user, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(afterETH.Cmp(beforeETH)).To(Equal(1))
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

		It("unwraps only the WETH gained after a checkpoint", func() {
			preExisting := decimal.RequireFromString("0.3")
			gained := decimal.RequireFromString("0.2")
			gainedWei := decimal.RequireFromString("0.2e18").BigInt()
			seedWETH(ctx, ethClient, opts, user, preExisting, defi.ExecutionEOA)

			beforeWETH, err := wethContract.BalanceOf(&bind.CallOpts{Context: ctx}, user)
			Expect(err).NotTo(HaveOccurred())
			checkpoint := txamount.Checkpoint("before-new-wrap", base.WETH)
			flow := defi.NewFlow(user, defi.WithChain(config.Base)).
				Add(defi.CheckpointBefore(
					sdkweth.Wrap(base.WETH, txamount.Exact(gained)),
					checkpoint,
				)).
				Add(sdkweth.Unwrap(base.WETH, txamount.CheckpointDelta(checkpoint)))

			result, err := defi.NewRunner(ethClient, opts, config.Base).
				ExecuteWithResult(ctx, flow, defi.ExecutionDynamicEOA)

			Expect(err).NotTo(HaveOccurred())
			Expect(result.Receipt.Status).To(Equal(uint64(types.ReceiptStatusSuccessful)))
			deposits := defi.EventsOf[*sdkweth.DepositEvent](result)
			Expect(deposits).To(HaveLen(1))
			Expect(deposits[0].Account).To(Equal(user))
			Expect(deposits[0].Amount).To(Equal(gainedWei))
			withdrawals := defi.EventsOf[*sdkweth.WithdrawalEvent](result)
			Expect(withdrawals).To(HaveLen(1))
			Expect(withdrawals[0].Account).To(Equal(user))
			Expect(withdrawals[0].Amount).To(Equal(gainedWei))

			afterWETH, err := wethContract.BalanceOf(&bind.CallOpts{Context: ctx}, user)
			Expect(err).NotTo(HaveOccurred())
			Expect(afterWETH).To(Equal(beforeWETH))
		})
	})
})

func delegateForkEOA(
	ctx context.Context,
	ethClient *ethclient.Client,
	manager *eip7702.Manager,
	user common.Address,
	implementation common.Address,
) {
	GinkgoHelper()
	delegateTx, err := manager.Delegate(ctx, implementation)
	Expect(err).NotTo(HaveOccurred())
	delegateReceipt, err := bind.WaitMined(ctx, ethClient, delegateTx)
	Expect(err).NotTo(HaveOccurred())
	Expect(delegateReceipt.Status).To(Equal(uint64(types.ReceiptStatusSuccessful)))
	Expect(manager.AssertDelegatedTo(ctx, user, implementation)).To(Succeed())
}

func seedWETH(
	ctx context.Context,
	ethClient *ethclient.Client,
	opts *bind.TransactOpts,
	user common.Address,
	value decimal.Decimal,
	mode defi.ExecutionMode,
) {
	GinkgoHelper()
	flow := defi.NewFlow(user, defi.WithChain(config.Base)).
		Add(sdkweth.Wrap(base.WETH, txamount.Exact(value)))
	result, err := defi.NewRunner(ethClient, opts, config.Base).
		ExecuteWithResult(ctx, flow, mode)
	Expect(err).NotTo(HaveOccurred())
	Expect(result.Receipt.Status).To(Equal(uint64(types.ReceiptStatusSuccessful)))
}
