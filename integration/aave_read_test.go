//go:build integration

package integration

import (
	"context"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	sdkcontract "github.com/tn606024/defi-simplify/client/contract"
	"github.com/tn606024/defi-simplify/config"
)

var _ = Describe("Aave read integration", func() {
	It("reads Base Aave reserve and user account data through the SDK", func() {
		ctx := context.Background()
		ethClient := baseForkClient(GinkgoT())
		client := newForkAaveReadClient(GinkgoT(), ethClient)

		pool, err := config.Base.AaveV3PoolAddress()
		Expect(err).NotTo(HaveOccurred())
		assertContractCode(GinkgoT(), ctx, ethClient, pool, "Aave V3 Pool")

		protocolDataProvider, err := config.Base.AaveProtocolDataProviderAddress()
		Expect(err).NotTo(HaveOccurred())
		assertContractCode(GinkgoT(), ctx, ethClient, protocolDataProvider, "Aave Protocol Data Provider")

		usdc, err := config.USDC.Address(config.Base)
		Expect(err).NotTo(HaveOccurred())

		reserveData, err := client.Aave.GetReserveData(ctx, config.USDC)
		Expect(err).NotTo(HaveOccurred())
		Expect(reserveData).NotTo(BeNil())
		Expect(reserveData.LiquidityIndex).NotTo(BeNil())
		Expect(reserveData.LiquidityIndex.Sign()).To(BeNumerically(">", 0))
		Expect(reserveData.VariableBorrowIndex).NotTo(BeNil())
		Expect(reserveData.VariableBorrowIndex.Sign()).To(BeNumerically(">", 0))
		Expect(reserveData.LastUpdateTimestamp).NotTo(BeNil())
		Expect(reserveData.LastUpdateTimestamp.Sign()).To(BeNumerically(">", 0))
		Expect(reserveData.ATokenAddress).NotTo(Equal(common.Address{}))
		Expect(reserveData.VariableDebtTokenAddress).NotTo(Equal(common.Address{}))

		userAccountData, err := client.Aave.GetUserAccountData(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(userAccountData).NotTo(BeNil())
		Expect(userAccountData.TotalCollateralBase).NotTo(BeNil())
		Expect(userAccountData.TotalDebtBase).NotTo(BeNil())
		Expect(userAccountData.AvailableBorrowsBase).NotTo(BeNil())
		Expect(userAccountData.CurrentLiquidationThreshold).NotTo(BeNil())
		Expect(userAccountData.Ltv).NotTo(BeNil())
		Expect(userAccountData.HealthFactor).NotTo(BeNil())

		reserves, err := client.Aave.GetAllReservesTokens(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(reserves).NotTo(BeEmpty())

		foundUSDC := false
		for _, reserve := range reserves {
			if reserve.TokenAddress == usdc {
				foundUSDC = true
				Expect(reserve.Symbol).NotTo(BeEmpty())
				break
			}
		}
		Expect(foundUSDC).To(BeTrue())
	})
})

func newForkAaveReadClient(t testHelper, ethClient *ethclient.Client) *sdkcontract.DefiClient {
	t.Helper()

	return sdkcontract.NewDefiClient(
		&bind.TransactOpts{From: forkTestUser},
		ethClient,
		nil,
		config.Base,
	)
}
