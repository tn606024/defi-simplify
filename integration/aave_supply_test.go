//go:build integration

package integration

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/shopspring/decimal"
	"github.com/tn606024/defi-simplify/bind/erc20"
	sdkcontract "github.com/tn606024/defi-simplify/client/contract"
	"github.com/tn606024/defi-simplify/config"
)

var _ = Describe("Aave supply integration", func() {
	It("supplies user USDC into Base Aave V3 on a local fork", func() {
		ctx := context.Background()
		ethClient := baseForkClient(GinkgoT())
		rpcClient := baseForkRPCClient(GinkgoT())
		requireAnvilFork(GinkgoT(), ctx, rpcClient)

		client, user := newForkDefiClient(GinkgoT(), ctx, rpcClient, ethClient)

		usdc, err := config.USDC.Address(config.Base)
		Expect(err).NotTo(HaveOccurred())
		assertContractCode(GinkgoT(), ctx, ethClient, usdc, "USDC")

		pool, err := config.Base.AaveV3PoolAddress()
		Expect(err).NotTo(HaveOccurred())
		assertContractCode(GinkgoT(), ctx, ethClient, pool, "Aave V3 Pool")

		reserveData, err := client.Aave.GetReserveData(ctx, config.USDC)
		Expect(err).NotTo(HaveOccurred())
		assertContractCode(GinkgoT(), ctx, ethClient, reserveData.ATokenAddress, "Base Aave aUSDC")

		supplyAmount := decimal.NewFromInt(1)
		supplyAmountWei := big.NewInt(1_000_000)
		Expect(fundBaseUSDCFromHolder(ctx, rpcClient, ethClient, user, supplyAmountWei)).To(Succeed())

		token, err := erc20.NewErc20(usdc, ethClient)
		Expect(err).NotTo(HaveOccurred())
		aToken, err := erc20.NewErc20(reserveData.ATokenAddress, ethClient)
		Expect(err).NotTo(HaveOccurred())

		beforeUSDC, err := token.BalanceOf(nil, user)
		Expect(err).NotTo(HaveOccurred())
		beforeAToken, err := aToken.BalanceOf(nil, user)
		Expect(err).NotTo(HaveOccurred())
		beforeAccountData, err := client.Aave.GetUserAccountData(ctx)
		Expect(err).NotTo(HaveOccurred())
		beforeReserveData, err := client.Aave.GetUserReserveData(ctx, usdc)
		Expect(err).NotTo(HaveOccurred())

		approveReceipt, err := client.ERC20.Approve(ctx, config.USDC, pool, supplyAmount)
		Expect(err).NotTo(HaveOccurred())
		Expect(approveReceipt.Status).To(Equal(types.ReceiptStatusSuccessful))

		supplyReceipt, err := client.Aave.Supply(ctx, config.USDC, supplyAmount)
		Expect(err).NotTo(HaveOccurred())
		Expect(supplyReceipt.Status).To(Equal(types.ReceiptStatusSuccessful))

		assertAaveSupplyStateChange(
			GinkgoT(),
			ctx,
			client,
			token,
			aToken,
			user,
			usdc,
			supplyAmountWei,
			beforeUSDC,
			beforeAToken,
			beforeAccountData,
			beforeReserveData,
		)
	})
})

func assertAaveSupplyStateChange(
	t testHelper,
	ctx context.Context,
	client *sdkcontract.DefiClient,
	token *erc20.Erc20,
	aToken *erc20.Erc20,
	user common.Address,
	asset common.Address,
	supplyAmountWei *big.Int,
	beforeUSDC *big.Int,
	beforeAToken *big.Int,
	beforeAccountData *sdkcontract.DataTypesUserAccountData,
	beforeReserveData *sdkcontract.DataTypesUserReserveData,
) {
	t.Helper()

	afterUSDC, err := token.BalanceOf(nil, user)
	Expect(err).NotTo(HaveOccurred())
	usdcSpent := new(big.Int).Sub(beforeUSDC, afterUSDC)
	Expect(usdcSpent.Cmp(supplyAmountWei)).To(Equal(0))

	afterAToken, err := aToken.BalanceOf(nil, user)
	Expect(err).NotTo(HaveOccurred())
	aTokenDelta := new(big.Int).Sub(afterAToken, beforeAToken)
	Expect(aTokenDelta.Sign()).To(Equal(1))

	afterReserveData, err := client.Aave.GetUserReserveData(ctx, asset)
	Expect(err).NotTo(HaveOccurred())
	Expect(afterReserveData.CurrentATokenBalance.Cmp(beforeReserveData.CurrentATokenBalance)).To(Equal(1))
	Expect(afterReserveData.CurrentATokenBalance.Cmp(afterAToken)).To(Equal(0))
	Expect(afterReserveData.UsageAsCollateralEnabled).To(BeTrue())

	afterAccountData, err := client.Aave.GetUserAccountData(ctx)
	Expect(err).NotTo(HaveOccurred())
	Expect(afterAccountData.TotalCollateralBase.Cmp(beforeAccountData.TotalCollateralBase)).To(Equal(1))
	Expect(afterAccountData.AvailableBorrowsBase.Cmp(beforeAccountData.AvailableBorrowsBase)).To(Equal(1))
	Expect(afterAccountData.TotalDebtBase.Cmp(beforeAccountData.TotalDebtBase)).To(Equal(0))
}
