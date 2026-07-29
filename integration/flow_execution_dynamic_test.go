//go:build integration

package integration

import (
	"context"
	"errors"
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
	binderc20 "github.com/tn606024/defi-simplify/bind/erc20"
	"github.com/tn606024/defi-simplify/client/account/defisimplify7702"
	"github.com/tn606024/defi-simplify/client/account/eip7702"
	"github.com/tn606024/defi-simplify/client/contract"
	"github.com/tn606024/defi-simplify/config"
	sdkerc20 "github.com/tn606024/defi-simplify/erc20"
	"github.com/tn606024/defi-simplify/token"
)

var _ = Describe("Dynamic Flow execution integration", func() {
	var (
		ctx            context.Context
		cancel         context.CancelFunc
		ethClient      *ethclient.Client
		rpcClient      *rpc.Client
		opts           *bind.TransactOpts
		user           common.Address
		manager        *eip7702.Manager
		implementation defisimplify7702.RuntimeIdentity
		usdc           token.Token
		usdcContract   *binderc20.Erc20
	)

	BeforeEach(func() {
		ctx, cancel = context.WithTimeout(context.Background(), 60*time.Second)
		DeferCleanup(cancel)

		ethClient = baseForkClient(GinkgoT())
		rpcClient = baseForkRPCClient(GinkgoT())
		requireAnvilFork(GinkgoT(), ctx, rpcClient)

		forkOpts, authorizationPrivateKey, forkUser := newForkTransactorWithKey(
			GinkgoT(),
			ctx,
			rpcClient,
		)
		opts = forkOpts
		user = forkUser

		deployment, err := defisimplify7702.DeploymentForChain(config.Base)
		Expect(err).NotTo(HaveOccurred())
		accountDeployment, err := deployment.Contract(defisimplify7702.AccountContract)
		Expect(err).NotTo(HaveOccurred())
		implementation = accountDeployment.RuntimeIdentity
		implementationCode, err := ethClient.CodeAt(ctx, implementation.Address, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(implementation.VerifyRuntimeCode(implementationCode)).To(Succeed())

		chainID, err := config.Base.ChainID()
		Expect(err).NotTo(HaveOccurred())
		manager, err = eip7702.NewManager(
			ethClient,
			opts,
			authorizationPrivateKey,
			big.NewInt(int64(chainID)),
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(manager.AssertClean(ctx, user)).To(Succeed())

		delegateTx, err := manager.Delegate(ctx, implementation.Address)
		Expect(err).NotTo(HaveOccurred())
		delegateReceipt, err := bind.WaitMined(ctx, ethClient, delegateTx)
		Expect(err).NotTo(HaveOccurred())
		Expect(delegateReceipt.Status).To(Equal(uint64(types.ReceiptStatusSuccessful)))
		Expect(manager.AssertDelegatedTo(ctx, user, implementation.Address)).To(Succeed())

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

		_, usdcReserve, _ := loadBaseAaveReserves(GinkgoT(), ctx, ethClient)
		usdc = usdcReserve.Underlying()
		usdcContract, err = binderc20.NewErc20(usdc.Address(), ethClient)
		Expect(err).NotTo(HaveOccurred())
	})

	It("executes current-balance calls and mixed exact calls as the EOA", func() {
		funded := decimal.NewFromInt(5).Shift(int32(usdc.Decimals())).BigInt()
		Expect(fundBaseUSDCFromHolder(ctx, rpcClient, ethClient, user, funded)).To(Succeed())

		firstSpender := common.HexToAddress("0x00000000000000000000000000000000000000d1")
		single := defi.NewFlow(user, defi.WithChain(config.Base)).
			Add(sdkerc20.Approve(
				usdc,
				sdkerc20.AddressSpender(firstSpender),
				txamount.CurrentBalance(usdc.Ref()),
			))
		runner := defi.NewRunner(ethClient, opts, config.Base)

		result, err := runner.Execute(ctx, single)
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Receipt.Status).To(Equal(uint64(types.ReceiptStatusSuccessful)))
		firstAllowance, err := usdcContract.Allowance(nil, user, firstSpender)
		Expect(err).NotTo(HaveOccurred())
		Expect(firstAllowance).To(Equal(funded))

		exactSpender := common.HexToAddress("0x00000000000000000000000000000000000000d2")
		dynamicSpender := common.HexToAddress("0x00000000000000000000000000000000000000d3")
		mixed := defi.NewFlow(user, defi.WithChain(config.Base)).
			Add(sdkerc20.Approve(
				usdc,
				sdkerc20.AddressSpender(exactSpender),
				txamount.Exact(decimal.NewFromInt(2)),
			)).
			Add(sdkerc20.Approve(
				usdc,
				sdkerc20.AddressSpender(dynamicSpender),
				txamount.CurrentBalance(usdc.Ref()),
			))

		result, err = runner.Execute(ctx, mixed)
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Receipt.Status).To(Equal(uint64(types.ReceiptStatusSuccessful)))
		approvals := defi.EventsOf[*sdkerc20.ApprovalEvent](result)
		Expect(approvals).To(HaveLen(2))
		Expect(approvals[0].Owner).To(Equal(user))
		Expect(approvals[0].Spender).To(Equal(exactSpender))
		Expect(approvals[1].Owner).To(Equal(user))
		Expect(approvals[1].Spender).To(Equal(dynamicSpender))
		Expect(approvals[1].Amount).To(Equal(funded))
	})

	It("patches a later call with the balance gained after a checkpoint", func() {
		initial := decimal.NewFromInt(1).Shift(int32(usdc.Decimals())).BigInt()
		incomingDecimal := decimal.NewFromInt(3)
		incoming := incomingDecimal.Shift(int32(usdc.Decimals())).BigInt()
		Expect(fundBaseUSDCFromHolder(ctx, rpcClient, ethClient, user, initial)).To(Succeed())
		Expect(approveBaseUSDCFromHolder(ctx, rpcClient, ethClient, user, incoming)).To(Succeed())

		spender := common.HexToAddress("0x00000000000000000000000000000000000000e1")
		checkpoint := txamount.Checkpoint("holder-transfer", usdc.Ref())
		flow := defi.NewFlow(user, defi.WithChain(config.Base)).
			Add(defi.CheckpointBefore(
				sdkerc20.TransferFrom(
					usdc,
					baseUSDCFunder,
					user,
					txamount.Exact(incomingDecimal),
				),
				checkpoint,
			)).
			Add(sdkerc20.Approve(
				usdc,
				sdkerc20.AddressSpender(spender),
				txamount.CheckpointDelta(checkpoint),
			))

		result, err := defi.NewRunner(ethClient, opts, config.Base).
			Execute(ctx, flow)

		Expect(err).NotTo(HaveOccurred())
		Expect(result.Receipt.Status).To(Equal(uint64(types.ReceiptStatusSuccessful)))
		transfers := defi.EventsOf[*sdkerc20.TransferEvent](result)
		Expect(transfers).To(HaveLen(1))
		Expect(transfers[0].From).To(Equal(baseUSDCFunder))
		Expect(transfers[0].To).To(Equal(user))
		Expect(transfers[0].Amount).To(Equal(incoming))
		approvals := defi.EventsOf[*sdkerc20.ApprovalEvent](result)
		Expect(approvals).To(HaveLen(1))
		Expect(approvals[0].Owner).To(Equal(user))
		Expect(approvals[0].Amount).To(Equal(incoming))

		allowance, err := usdcContract.Allowance(nil, user, spender)
		Expect(err).NotTo(HaveOccurred())
		Expect(allowance).To(Equal(incoming))
	})

	It("attributes a failed call, rolls back the batch, and rejects cleared delegation", func() {
		fundedDecimal := decimal.NewFromInt(1)
		funded := fundedDecimal.Shift(int32(usdc.Decimals())).BigInt()
		Expect(fundBaseUSDCFromHolder(ctx, rpcClient, ethClient, user, funded)).To(Succeed())

		spender := common.HexToAddress("0x00000000000000000000000000000000000000f1")
		recipient := common.HexToAddress("0x00000000000000000000000000000000000000f2")
		dynamicSpender := common.HexToAddress("0x00000000000000000000000000000000000000f3")
		beforeAllowance, err := usdcContract.Allowance(nil, user, spender)
		Expect(err).NotTo(HaveOccurred())
		beforeRecipient, err := usdcContract.BalanceOf(nil, recipient)
		Expect(err).NotTo(HaveOccurred())

		flow := defi.NewFlow(user, defi.WithChain(config.Base)).
			Add(sdkerc20.Approve(
				usdc,
				sdkerc20.AddressSpender(spender),
				txamount.Exact(fundedDecimal),
			)).
			Add(sdkerc20.Transfer(
				usdc,
				recipient,
				txamount.Exact(decimal.NewFromInt(100)),
			)).
			Add(sdkerc20.Approve(
				usdc,
				sdkerc20.AddressSpender(dynamicSpender),
				txamount.CurrentBalance(usdc.Ref()),
			))

		opts.GasLimit = 1_500_000
		result, err := defi.NewRunner(ethClient, opts, config.Base).
			Execute(ctx, flow)

		Expect(result).NotTo(BeNil())
		Expect(result.Receipt.Status).To(Equal(uint64(types.ReceiptStatusFailed)))
		Expect(errors.Is(err, contract.ErrTransactionReverted)).To(BeTrue())
		var accountErr *defisimplify7702.ContractError
		Expect(errors.As(err, &accountErr)).To(BeTrue())
		Expect(accountErr.Name).To(Equal("DynamicCallFailed"))
		Expect(accountErr.CallIndex).To(Equal(big.NewInt(1)))
		Expect(result.Steps).To(HaveLen(3))
		for _, step := range result.Steps {
			Expect(step.Status).To(Equal(defi.ValidationSkipped))
			Expect(step.SkipReason).To(Equal(defi.SkipExecutionFailed))
		}

		afterAllowance, err := usdcContract.Allowance(nil, user, spender)
		Expect(err).NotTo(HaveOccurred())
		Expect(afterAllowance).To(Equal(beforeAllowance))
		afterRecipient, err := usdcContract.BalanceOf(nil, recipient)
		Expect(err).NotTo(HaveOccurred())
		Expect(afterRecipient).To(Equal(beforeRecipient))

		opts.GasLimit = 0
		clearTx, err := manager.Clear(ctx)
		Expect(err).NotTo(HaveOccurred())
		clearReceipt, err := bind.WaitMined(ctx, ethClient, clearTx)
		Expect(err).NotTo(HaveOccurred())
		Expect(clearReceipt.Status).To(Equal(uint64(types.ReceiptStatusSuccessful)))
		Expect(manager.AssertClean(ctx, user)).To(Succeed())
		nonceBefore, err := ethClient.PendingNonceAt(ctx, user)
		Expect(err).NotTo(HaveOccurred())

		clearedFlow := defi.NewFlow(user, defi.WithChain(config.Base)).
			Add(sdkerc20.Approve(
				usdc,
				sdkerc20.AddressSpender(spender),
				txamount.CurrentBalance(usdc.Ref()),
			))
		clearedResult, err := defi.NewRunner(ethClient, opts, config.Base).
			Execute(ctx, clearedFlow)
		Expect(clearedResult).To(BeNil())
		Expect(errors.Is(err, eip7702.ErrUnexpectedDelegation)).To(BeTrue())
		nonceAfter, err := ethClient.PendingNonceAt(ctx, user)
		Expect(err).NotTo(HaveOccurred())
		Expect(nonceAfter).To(Equal(nonceBefore))
	})
})
