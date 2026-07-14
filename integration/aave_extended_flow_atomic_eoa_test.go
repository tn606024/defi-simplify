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
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/shopspring/decimal"
	defi "github.com/tn606024/defi-simplify"
	"github.com/tn606024/defi-simplify/aave"
	bindaave "github.com/tn606024/defi-simplify/bind/aave"
	binderc20 "github.com/tn606024/defi-simplify/bind/erc20"
	"github.com/tn606024/defi-simplify/client/account/eip7702"
	"github.com/tn606024/defi-simplify/config"
	sdkerc20 "github.com/tn606024/defi-simplify/erc20"
	"github.com/tn606024/defi-simplify/helper"
)

var _ = Describe("Extended Aave FlowStep integration", func() {
	var (
		ctx              context.Context
		cancel           context.CancelFunc
		ethClient        *ethclient.Client
		rpcClient        *rpc.Client
		opts             *bind.TransactOpts
		signer           *helper.MsgSigner
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
		opts, signer, authorizationKey, user = newForkTransactorWithKey(GinkgoT(), ctx, rpcClient)

		var err error
		implementation, err = config.Base.Simple7702AccountImplementationAddress()
		Expect(err).NotTo(HaveOccurred())
		chainID, err := config.Base.ChainID()
		Expect(err).NotTo(HaveOccurred())
		manager, err = eip7702.NewManager(ethClient, opts, authorizationKey, big.NewInt(int64(chainID)))
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
	})

	It("executes permit supply, repay, and withdraw as one EOA-native flow", func() {
		pool, err := config.Base.AaveV3PoolAddress()
		Expect(err).NotTo(HaveOccurred())
		supplyAmount := decimal.NewFromInt(10)
		borrowAmount := decimal.NewFromInt(1).Shift(-6)
		withdrawAmount := decimal.NewFromInt(1)
		supplyAmountWei := decimalAmount(config.USDC, supplyAmount)
		Expect(fundBaseUSDCFromHolder(ctx, rpcClient, ethClient, user, supplyAmountWei)).To(Succeed())

		deadline := big.NewInt(time.Now().Add(10 * time.Minute).Unix())
		v, r, s, err := signPermit(ctx, ethClient, config.USDC, user, pool, supplyAmountWei, deadline, signer)
		Expect(err).NotTo(HaveOccurred())

		flow := defi.NewFlow(user, defi.WithChain(config.Base)).
			Add(aave.SupplyWithPermit(config.USDC, supplyAmount, deadline, v, r, s)).
			Add(aave.Borrow(config.WETH, borrowAmount)).
			Add(sdkerc20.Approve(config.WETH, aave.PoolSpender(), borrowAmount)).
			Add(aave.Repay(config.WETH, borrowAmount)).
			Add(aave.Withdraw(config.USDC, withdrawAmount))

		result, err := defi.NewRunner(ethClient, opts, config.Base).
			ExecuteWithResult(ctx, flow, defi.ExecutionAtomicEOA)

		Expect(err).NotTo(HaveOccurred())
		Expect(result.Receipt.Status).To(Equal(uint64(types.ReceiptStatusSuccessful)))
		Expect(defi.EventsOf[*aave.SupplyEvent](result)).To(HaveLen(1))
		Expect(defi.EventsOf[*aave.BorrowEvent](result)).To(HaveLen(1))
		Expect(defi.EventsOf[*aave.RepayEvent](result)).To(HaveLen(1))
		Expect(defi.EventsOf[*aave.WithdrawEvent](result)).To(HaveLen(1))
	})

	It("repays and withdraws entire positions using Aave sentinel amounts", func() {
		depositAmount := decimal.RequireFromString("0.01")
		borrowAmount := decimal.NewFromInt(1)
		repaymentApproval := decimal.NewFromInt(2)
		depositAmountWei := decimalAmount(config.WETH, depositAmount)
		borrowAmountWei := decimalAmount(config.USDC, borrowAmount)
		repaymentApprovalWei := decimalAmount(config.USDC, repaymentApproval)
		maxAmount := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))
		Expect(fundBaseUSDCFromHolder(ctx, rpcClient, ethClient, user, borrowAmountWei)).To(Succeed())

		flow := defi.NewFlow(user, defi.WithChain(config.Base)).
			Add(aave.DepositETH(depositAmount)).
			Add(aave.Borrow(config.USDC, borrowAmount)).
			Add(sdkerc20.Approve(config.USDC, aave.PoolSpender(), repaymentApproval)).
			Add(aave.RepayAll(config.USDC)).
			Add(aave.WithdrawAll(config.WETH))

		result, err := defi.NewRunner(ethClient, opts, config.Base).
			ExecuteWithResult(ctx, flow, defi.ExecutionAtomicEOA)

		Expect(err).NotTo(HaveOccurred())
		Expect(result.Receipt.Status).To(Equal(uint64(types.ReceiptStatusSuccessful)))
		repayments := defi.EventsOf[*aave.RepayEvent](result)
		withdrawals := defi.EventsOf[*aave.WithdrawEvent](result)
		Expect(repayments).To(HaveLen(1))
		Expect(repayments[0].Asset).To(Equal(mustCoinAddress(config.USDC)))
		Expect(repayments[0].User).To(Equal(user))
		Expect(repayments[0].Repayer).To(Equal(user))
		Expect(repayments[0].Amount.Cmp(borrowAmountWei)).To(BeNumerically(">=", 0))
		Expect(repayments[0].Amount.Cmp(repaymentApprovalWei)).To(BeNumerically("<", 0))
		Expect(repayments[0].Amount).NotTo(Equal(maxAmount))
		Expect(withdrawals).To(HaveLen(1))
		Expect(withdrawals[0].Asset).To(Equal(mustCoinAddress(config.WETH)))
		Expect(withdrawals[0].User).To(Equal(user))
		Expect(withdrawals[0].To).To(Equal(user))
		Expect(withdrawals[0].Amount.Sign()).To(Equal(1))
		Expect(withdrawals[0].Amount.Cmp(depositAmountWei)).To(BeNumerically("<=", 0))
		Expect(withdrawals[0].Amount).NotTo(Equal(maxAmount))
		Expect(repayments[0].Metadata.LogIndex).To(BeNumerically("<", withdrawals[0].Metadata.LogIndex))

		debtToken, err := config.USDC.DebtToken()
		Expect(err).NotTo(HaveOccurred())
		debtTokenBinding, err := binderc20.NewErc20(mustCoinAddress(debtToken), ethClient)
		Expect(err).NotTo(HaveOccurred())
		debtBalance, err := debtTokenBinding.BalanceOf(&bind.CallOpts{Context: ctx}, user)
		Expect(err).NotTo(HaveOccurred())
		Expect(debtBalance.Sign()).To(Equal(0))

		aToken, err := config.WETH.AToken()
		Expect(err).NotTo(HaveOccurred())
		aTokenBinding, err := binderc20.NewErc20(mustCoinAddress(aToken), ethClient)
		Expect(err).NotTo(HaveOccurred())
		aTokenBalance, err := aTokenBinding.BalanceOf(&bind.CallOpts{Context: ctx}, user)
		Expect(err).NotTo(HaveOccurred())
		Expect(aTokenBalance.Sign()).To(Equal(0))
	})

	It("executes native deposit, permit repayment, and permit withdrawal", func() {
		pool, err := config.Base.AaveV3PoolAddress()
		Expect(err).NotTo(HaveOccurred())
		gateway, err := config.Base.WrappedTokenGatewayV3Address()
		Expect(err).NotTo(HaveOccurred())
		collateralAmount := decimal.RequireFromString("0.02")
		withdrawAmount := decimal.RequireFromString("0.01")
		repayAmount := decimal.NewFromInt(1)
		withdrawAmountWei := decimalAmount(config.AWETH, withdrawAmount)
		repayAmountWei := decimalAmount(config.USDC, repayAmount)
		deadline := big.NewInt(time.Now().Add(10 * time.Minute).Unix())

		repayV, repayR, repayS, err := signPermit(
			ctx,
			ethClient,
			config.USDC,
			user,
			pool,
			repayAmountWei,
			deadline,
			signer,
		)
		Expect(err).NotTo(HaveOccurred())
		withdrawV, withdrawR, withdrawS, err := signPermit(
			ctx,
			ethClient,
			config.AWETH,
			user,
			gateway,
			withdrawAmountWei,
			deadline,
			signer,
		)
		Expect(err).NotTo(HaveOccurred())

		flow := defi.NewFlow(user, defi.WithChain(config.Base)).
			Add(aave.DepositETH(collateralAmount)).
			Add(aave.Borrow(config.USDC, repayAmount)).
			Add(aave.RepayWithPermit(config.USDC, repayAmount, deadline, repayV, repayR, repayS)).
			Add(aave.WithdrawETHWithPermit(
				withdrawAmount,
				deadline,
				withdrawV,
				withdrawR,
				withdrawS,
			))

		result, err := defi.NewRunner(ethClient, opts, config.Base).
			ExecuteWithResult(ctx, flow, defi.ExecutionAtomicEOA)

		Expect(err).NotTo(HaveOccurred())
		Expect(result.Receipt.Status).To(Equal(uint64(types.ReceiptStatusSuccessful)))
		supplies := defi.EventsOf[*aave.SupplyEvent](result)
		repayments := defi.EventsOf[*aave.RepayEvent](result)
		withdrawals := defi.EventsOf[*aave.WithdrawEvent](result)
		Expect(supplies).To(HaveLen(1))
		Expect(supplies[0].User).To(Equal(gateway))
		Expect(supplies[0].OnBehalfOf).To(Equal(user))
		Expect(repayments).To(HaveLen(1))
		Expect(withdrawals).To(HaveLen(1))
		Expect(withdrawals[0].User).To(Equal(gateway))
		Expect(withdrawals[0].To).To(Equal(gateway))
	})

	It("borrows native ETH through delegated credit and performs a plain gateway withdrawal", func() {
		gateway, err := config.Base.WrappedTokenGatewayV3Address()
		Expect(err).NotTo(HaveOccurred())
		supplyAmount := decimal.NewFromInt(10)
		borrowAmount := decimal.RequireFromString("0.0001")
		depositAmount := decimal.RequireFromString("0.001")
		withdrawAmount := decimal.RequireFromString("0.0005")
		Expect(fundBaseUSDCFromHolder(
			ctx,
			rpcClient,
			ethClient,
			user,
			decimalAmount(config.USDC, supplyAmount),
		)).To(Succeed())

		flow := defi.NewFlow(user, defi.WithChain(config.Base)).
			Add(sdkerc20.Approve(config.USDC, aave.PoolSpender(), supplyAmount)).
			Add(aave.Supply(config.USDC, supplyAmount)).
			Add(aave.ApproveDelegation(config.WETH, gateway, borrowAmount)).
			Add(aave.BorrowETH(borrowAmount)).
			Add(aave.DepositETH(depositAmount)).
			Add(sdkerc20.Approve(config.AWETH, aave.GatewaySpender(), withdrawAmount)).
			Add(aave.WithdrawETH(withdrawAmount))

		result, err := defi.NewRunner(ethClient, opts, config.Base).
			ExecuteWithResult(ctx, flow, defi.ExecutionAtomicEOA)

		Expect(err).NotTo(HaveOccurred())
		Expect(result.Receipt.Status).To(Equal(uint64(types.ReceiptStatusSuccessful)))
		delegations := defi.EventsOf[*aave.BorrowAllowanceDelegatedEvent](result)
		borrows := defi.EventsOf[*aave.BorrowEvent](result)
		withdrawals := defi.EventsOf[*aave.WithdrawEvent](result)
		Expect(delegations).To(HaveLen(1))
		Expect(delegations[0].FromUser).To(Equal(user))
		Expect(delegations[0].ToUser).To(Equal(gateway))
		Expect(borrows).To(HaveLen(1))
		Expect(borrows[0].User).To(Equal(gateway))
		Expect(borrows[0].OnBehalfOf).To(Equal(user))
		Expect(withdrawals).To(HaveLen(1))
	})

	It("submits DelegationWithSig from a relaying flow account", func() {
		delegatorKey, err := crypto.GenerateKey()
		Expect(err).NotTo(HaveOccurred())
		delegator := crypto.PubkeyToAddress(delegatorKey.PublicKey)
		delegatorSigner := helper.NewMsgSigner(delegatorKey)
		delegatee, err := config.Base.WrappedTokenGatewayV3Address()
		Expect(err).NotTo(HaveOccurred())
		amount := decimal.RequireFromString("0.005")
		amountWei := decimalAmount(config.WETH, amount)
		deadline := big.NewInt(time.Now().Add(10 * time.Minute).Unix())
		v, r, s, err := signDelegation(
			ctx,
			ethClient,
			config.WETH,
			delegator,
			delegatee,
			amountWei,
			deadline,
			delegatorSigner,
		)
		Expect(err).NotTo(HaveOccurred())

		flow := defi.NewFlow(user, defi.WithChain(config.Base)).
			Add(aave.DelegationWithSig(config.WETH, delegator, delegatee, amount, deadline, v, r, s))
		result, err := defi.NewRunner(ethClient, opts, config.Base).
			ExecuteWithResult(ctx, flow, defi.ExecutionAtomicEOA)

		Expect(err).NotTo(HaveOccurred())
		Expect(result.Receipt.Status).To(Equal(uint64(types.ReceiptStatusSuccessful)))
		delegations := defi.EventsOf[*aave.BorrowAllowanceDelegatedEvent](result)
		Expect(delegations).To(HaveLen(1))
		Expect(delegations[0].FromUser).To(Equal(delegator))
		Expect(delegations[0].ToUser).To(Equal(delegatee))

		debtTokenCoin, err := config.WETH.DebtToken()
		Expect(err).NotTo(HaveOccurred())
		debtTokenAddress, err := debtTokenCoin.Address(config.Base)
		Expect(err).NotTo(HaveOccurred())
		debtToken, err := bindaave.NewDebtTokenBase(debtTokenAddress, ethClient)
		Expect(err).NotTo(HaveOccurred())
		allowance, err := debtToken.BorrowAllowance(&bind.CallOpts{Context: ctx}, delegator, delegatee)
		Expect(err).NotTo(HaveOccurred())
		Expect(allowance).To(Equal(amountWei))
	})
})

func decimalAmount(coin config.Coin, amount decimal.Decimal) *big.Int {
	decimals, err := coin.Decimals()
	Expect(err).NotTo(HaveOccurred())
	return amount.Shift(int32(decimals)).BigInt()
}

func mustCoinAddress(coin config.Coin) common.Address {
	address, err := coin.Address(config.Base)
	Expect(err).NotTo(HaveOccurred())
	return address
}
