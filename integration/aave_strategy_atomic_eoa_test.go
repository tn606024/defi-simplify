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
	binderc20 "github.com/tn606024/defi-simplify/bind/erc20"
	"github.com/tn606024/defi-simplify/client/account/eip7702"
	"github.com/tn606024/defi-simplify/config"
	sdkerc20 "github.com/tn606024/defi-simplify/erc20"
	"github.com/tn606024/defi-simplify/strategy"
)

var _ = Describe("Aave strategy integration", func() {
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
		opts, _, authorizationKey, user = newForkTransactorWithKey(GinkgoT(), ctx, rpcClient)

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

	It("opens an exact Aave supply and borrow position", func() {
		supplyAmount := decimal.NewFromInt(10)
		borrowAmount := decimal.NewFromInt(1).Shift(-6)
		supplyAmountWei := decimalAmount(config.USDC, supplyAmount)
		borrowAmountWei := decimalAmount(config.WETH, borrowAmount)
		Expect(fundBaseUSDCFromHolder(ctx, rpcClient, ethClient, user, supplyAmountWei)).To(Succeed())

		flow, err := strategy.AaveSupplyBorrow(strategy.AaveSupplyBorrowParams{
			Account:      user,
			Chain:        config.Base,
			SupplyAsset:  config.USDC,
			SupplyAmount: supplyAmount,
			BorrowAsset:  config.WETH,
			BorrowAmount: borrowAmount,
		})
		Expect(err).NotTo(HaveOccurred())

		result, err := defi.NewRunner(ethClient, opts, config.Base).
			ExecuteWithResult(ctx, flow, defi.ExecutionAtomicEOA)

		Expect(err).NotTo(HaveOccurred())
		Expect(result.Receipt.Status).To(Equal(uint64(types.ReceiptStatusSuccessful)))
		approvals := defi.EventsOf[*sdkerc20.ApprovalEvent](result)
		supplies := defi.EventsOf[*aave.SupplyEvent](result)
		borrows := defi.EventsOf[*aave.BorrowEvent](result)
		Expect(approvals).To(HaveLen(1))
		Expect(supplies).To(HaveLen(1))
		Expect(borrows).To(HaveLen(1))

		pool := mustAavePoolAddress()
		Expect(approvals[0].Token).To(Equal(mustCoinAddress(config.USDC)))
		Expect(approvals[0].Owner).To(Equal(user))
		Expect(approvals[0].Spender).To(Equal(pool))
		Expect(approvals[0].Amount).To(Equal(supplyAmountWei))
		Expect(supplies[0].Asset).To(Equal(mustCoinAddress(config.USDC)))
		Expect(supplies[0].User).To(Equal(user))
		Expect(supplies[0].OnBehalfOf).To(Equal(user))
		Expect(supplies[0].Amount).To(Equal(supplyAmountWei))
		Expect(borrows[0].Asset).To(Equal(mustCoinAddress(config.WETH)))
		Expect(borrows[0].User).To(Equal(user))
		Expect(borrows[0].OnBehalfOf).To(Equal(user))
		Expect(borrows[0].Amount).To(Equal(borrowAmountWei))
		Expect(borrows[0].InterestRateMode).To(Equal(aave.VariableInterestRateMode))
		Expect(approvals[0].Metadata.LogIndex).To(BeNumerically("<", supplies[0].Metadata.LogIndex))
		Expect(supplies[0].Metadata.LogIndex).To(BeNumerically("<", borrows[0].Metadata.LogIndex))

		aToken := mustTokenBinding(ethClient, config.USDC, func(asset config.Coin) (config.Coin, error) {
			return asset.AToken()
		})
		debtToken := mustTokenBinding(ethClient, config.WETH, func(asset config.Coin) (config.Coin, error) {
			return asset.DebtToken()
		})
		aTokenBalance, err := aToken.BalanceOf(&bind.CallOpts{Context: ctx}, user)
		Expect(err).NotTo(HaveOccurred())
		Expect(aTokenBalance.Sign()).To(Equal(1))
		debtBalance, err := debtToken.BalanceOf(&bind.CallOpts{Context: ctx}, user)
		Expect(err).NotTo(HaveOccurred())
		Expect(debtBalance.Sign()).To(Equal(1))
	})

	It("closes one Aave debt and collateral reserve pair", func() {
		collateralAmount := decimal.RequireFromString("0.01")
		borrowAmount := decimal.NewFromInt(1)
		temporaryAllowance := decimal.NewFromInt(2)
		borrowAmountWei := decimalAmount(config.USDC, borrowAmount)
		temporaryAllowanceWei := decimalAmount(config.USDC, temporaryAllowance)
		Expect(fundBaseUSDCFromHolder(ctx, rpcClient, ethClient, user, borrowAmountWei)).To(Succeed())

		setupFlow := defi.NewFlow(user, defi.WithChain(config.Base)).
			Add(aave.DepositETH(collateralAmount)).
			Add(aave.Borrow(config.USDC, borrowAmount))
		setupResult, err := defi.NewRunner(ethClient, opts, config.Base).
			ExecuteWithResult(ctx, setupFlow, defi.ExecutionAtomicEOA)
		Expect(err).NotTo(HaveOccurred())
		Expect(setupResult.Receipt.Status).To(Equal(uint64(types.ReceiptStatusSuccessful)))

		flow, err := strategy.AaveClosePosition(strategy.AaveClosePositionParams{
			Account:                 user,
			Chain:                   config.Base,
			DebtAsset:               config.USDC,
			TemporaryRepayAllowance: temporaryAllowance,
			CollateralAsset:         config.WETH,
		})
		Expect(err).NotTo(HaveOccurred())

		result, err := defi.NewRunner(ethClient, opts, config.Base).
			ExecuteWithResult(ctx, flow, defi.ExecutionAtomicEOA)

		Expect(err).NotTo(HaveOccurred())
		Expect(result.Receipt.Status).To(Equal(uint64(types.ReceiptStatusSuccessful)))
		approvals := defi.EventsOf[*sdkerc20.ApprovalEvent](result)
		repayments := defi.EventsOf[*aave.RepayEvent](result)
		withdrawals := defi.EventsOf[*aave.WithdrawEvent](result)
		Expect(approvals).To(HaveLen(2))
		Expect(repayments).To(HaveLen(1))
		Expect(withdrawals).To(HaveLen(1))

		pool := mustAavePoolAddress()
		Expect(approvals[0].Token).To(Equal(mustCoinAddress(config.USDC)))
		Expect(approvals[0].Owner).To(Equal(user))
		Expect(approvals[0].Spender).To(Equal(pool))
		Expect(approvals[0].Amount).To(Equal(temporaryAllowanceWei))
		Expect(repayments[0].Asset).To(Equal(mustCoinAddress(config.USDC)))
		Expect(repayments[0].User).To(Equal(user))
		Expect(repayments[0].Repayer).To(Equal(user))
		Expect(repayments[0].Amount.Cmp(borrowAmountWei)).To(BeNumerically(">=", 0))
		Expect(repayments[0].Amount.Cmp(temporaryAllowanceWei)).To(BeNumerically("<", 0))
		Expect(approvals[1].Token).To(Equal(mustCoinAddress(config.USDC)))
		Expect(approvals[1].Owner).To(Equal(user))
		Expect(approvals[1].Spender).To(Equal(pool))
		Expect(approvals[1].Amount.Sign()).To(Equal(0))
		Expect(withdrawals[0].Asset).To(Equal(mustCoinAddress(config.WETH)))
		Expect(withdrawals[0].User).To(Equal(user))
		Expect(withdrawals[0].To).To(Equal(user))
		Expect(withdrawals[0].Amount.Sign()).To(Equal(1))
		Expect(approvals[0].Metadata.LogIndex).To(BeNumerically("<", repayments[0].Metadata.LogIndex))
		Expect(repayments[0].Metadata.LogIndex).To(BeNumerically("<", approvals[1].Metadata.LogIndex))
		Expect(approvals[1].Metadata.LogIndex).To(BeNumerically("<", withdrawals[0].Metadata.LogIndex))

		debtToken := mustTokenBinding(ethClient, config.USDC, func(asset config.Coin) (config.Coin, error) {
			return asset.DebtToken()
		})
		aToken := mustTokenBinding(ethClient, config.WETH, func(asset config.Coin) (config.Coin, error) {
			return asset.AToken()
		})
		debtBalance, err := debtToken.BalanceOf(&bind.CallOpts{Context: ctx}, user)
		Expect(err).NotTo(HaveOccurred())
		Expect(debtBalance.Sign()).To(Equal(0))
		aTokenBalance, err := aToken.BalanceOf(&bind.CallOpts{Context: ctx}, user)
		Expect(err).NotTo(HaveOccurred())
		Expect(aTokenBalance.Sign()).To(Equal(0))
		debtAsset, err := binderc20.NewErc20(mustCoinAddress(config.USDC), ethClient)
		Expect(err).NotTo(HaveOccurred())
		allowance, err := debtAsset.Allowance(&bind.CallOpts{Context: ctx}, user, pool)
		Expect(err).NotTo(HaveOccurred())
		Expect(allowance.Sign()).To(Equal(0))
	})

})

func mustAavePoolAddress() common.Address {
	pool, err := config.Base.AaveV3PoolAddress()
	Expect(err).NotTo(HaveOccurred())
	return pool
}

func mustTokenBinding(
	client *ethclient.Client,
	asset config.Coin,
	resolve func(config.Coin) (config.Coin, error),
) *binderc20.Erc20 {
	token, err := resolve(asset)
	Expect(err).NotTo(HaveOccurred())
	binding, err := binderc20.NewErc20(mustCoinAddress(token), client)
	Expect(err).NotTo(HaveOccurred())
	return binding
}
