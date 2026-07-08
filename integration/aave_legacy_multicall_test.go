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

var _ = Describe("Legacy Multicall Aave integration", func() {
	It("supplies and borrows Base Aave USDC through the legacy Multicall flow", func() {
		ctx := context.Background()
		ethClient := baseForkClient(GinkgoT())
		rpcClient := baseForkRPCClient(GinkgoT())
		requireAnvilFork(GinkgoT(), ctx, rpcClient)

		client, user := newForkDefiClient(GinkgoT(), ctx, rpcClient, ethClient)

		usdc, err := config.USDC.Address(config.Base)
		Expect(err).NotTo(HaveOccurred())
		assertContractCode(GinkgoT(), ctx, ethClient, usdc, "USDC")

		multicall, err := config.Base.MulticallAddress()
		Expect(err).NotTo(HaveOccurred())
		assertContractCode(GinkgoT(), ctx, ethClient, multicall, "Multicall3")

		reserveData, err := client.Aave.GetReserveData(ctx, config.USDC)
		Expect(err).NotTo(HaveOccurred())
		assertContractCode(GinkgoT(), ctx, ethClient, reserveData.ATokenAddress, "Base Aave aUSDC")
		assertContractCode(GinkgoT(), ctx, ethClient, reserveData.VariableDebtTokenAddress, "Base Aave variable debt USDC")

		token, err := erc20.NewErc20(usdc, ethClient)
		Expect(err).NotTo(HaveOccurred())
		aToken, err := erc20.NewErc20(reserveData.ATokenAddress, ethClient)
		Expect(err).NotTo(HaveOccurred())
		variableDebtToken, err := erc20.NewErc20(reserveData.VariableDebtTokenAddress, ethClient)
		Expect(err).NotTo(HaveOccurred())

		supplyAmount := decimal.NewFromInt(2)
		borrowAmount := decimal.NewFromInt(1)
		supplyAmountWei := big.NewInt(2_000_000)
		borrowAmountWei := big.NewInt(1_000_000)
		Expect(fundBaseUSDCFromHolder(ctx, rpcClient, ethClient, user, supplyAmountWei)).To(Succeed())

		beforeUserUSDC, err := token.BalanceOf(nil, user)
		Expect(err).NotTo(HaveOccurred())
		beforeMulticallUSDC, err := token.BalanceOf(nil, multicall)
		Expect(err).NotTo(HaveOccurred())
		beforeAToken, err := aToken.BalanceOf(nil, user)
		Expect(err).NotTo(HaveOccurred())
		beforeVariableDebt, err := variableDebtToken.BalanceOf(nil, user)
		Expect(err).NotTo(HaveOccurred())
		beforeReserveData, err := client.Aave.GetUserReserveData(ctx, usdc)
		Expect(err).NotTo(HaveOccurred())

		receipt, err := client.LegacyMulticallSupplyAndBorrowAaveV3Coin(ctx, config.USDC, supplyAmount, borrowAmount)
		Expect(err).NotTo(HaveOccurred())
		Expect(receipt.Status).To(Equal(types.ReceiptStatusSuccessful))

		assertLegacyMulticallAaveStateChange(
			GinkgoT(),
			ctx,
			client,
			token,
			aToken,
			variableDebtToken,
			user,
			multicall,
			usdc,
			supplyAmountWei,
			borrowAmountWei,
			beforeUserUSDC,
			beforeMulticallUSDC,
			beforeAToken,
			beforeVariableDebt,
			beforeReserveData,
		)
	})
})

func assertLegacyMulticallAaveStateChange(
	t testHelper,
	ctx context.Context,
	client *sdkcontract.DefiClient,
	token *erc20.Erc20,
	aToken *erc20.Erc20,
	variableDebtToken *erc20.Erc20,
	user common.Address,
	multicall common.Address,
	asset common.Address,
	supplyAmountWei *big.Int,
	borrowAmountWei *big.Int,
	beforeUserUSDC *big.Int,
	beforeMulticallUSDC *big.Int,
	beforeAToken *big.Int,
	beforeVariableDebt *big.Int,
	beforeReserveData *sdkcontract.DataTypesUserReserveData,
) {
	t.Helper()

	// Legacy Multicall is a comparison baseline: user signs permissions, but
	// token movement and Aave calls pass through Multicall before settling back.
	afterUserUSDC, err := token.BalanceOf(nil, user)
	Expect(err).NotTo(HaveOccurred())
	expectedUserUSDC := new(big.Int).Sub(beforeUserUSDC, supplyAmountWei)
	expectedUserUSDC.Add(expectedUserUSDC, borrowAmountWei)
	Expect(afterUserUSDC.Cmp(expectedUserUSDC)).To(Equal(0))

	afterMulticallUSDC, err := token.BalanceOf(nil, multicall)
	Expect(err).NotTo(HaveOccurred())
	Expect(afterMulticallUSDC.Cmp(beforeMulticallUSDC)).To(Equal(0))

	afterAToken, err := aToken.BalanceOf(nil, user)
	Expect(err).NotTo(HaveOccurred())
	aTokenDelta := new(big.Int).Sub(afterAToken, beforeAToken)
	Expect(aTokenDelta.Sign()).To(Equal(1))

	afterVariableDebt, err := variableDebtToken.BalanceOf(nil, user)
	Expect(err).NotTo(HaveOccurred())
	variableDebtDelta := new(big.Int).Sub(afterVariableDebt, beforeVariableDebt)
	Expect(variableDebtDelta.Sign()).To(Equal(1))

	afterReserveData, err := client.Aave.GetUserReserveData(ctx, asset)
	Expect(err).NotTo(HaveOccurred())
	Expect(afterReserveData.CurrentATokenBalance.Cmp(beforeReserveData.CurrentATokenBalance)).To(Equal(1))
	Expect(afterReserveData.CurrentATokenBalance.Cmp(afterAToken)).To(Equal(0))
	Expect(afterReserveData.CurrentVariableDebt.Cmp(beforeReserveData.CurrentVariableDebt)).To(Equal(1))
	Expect(afterReserveData.CurrentVariableDebt.Cmp(afterVariableDebt)).To(Equal(0))
	Expect(afterReserveData.UsageAsCollateralEnabled).To(BeTrue())
}
