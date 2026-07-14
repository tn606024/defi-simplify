//go:build integration

package integration

import (
	"context"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tn606024/defi-simplify/aave"
	bindaave "github.com/tn606024/defi-simplify/bind/aave"
)

var _ = Describe("Aave deployment manifest", func() {
	It("matches the Base Pool and DataProvider provider relationships", func() {
		ctx := context.Background()
		client := baseForkClient(GinkgoT())

		manifest, err := aave.BaseV3Deployment()
		Expect(err).NotTo(HaveOccurred())
		market := manifest.Market()
		opts := &bind.CallOpts{Context: ctx}

		pool, err := bindaave.NewPoolCaller(market.Pool(), client)
		Expect(err).NotTo(HaveOccurred())
		poolProvider, err := pool.ADDRESSESPROVIDER(opts)
		Expect(err).NotTo(HaveOccurred())
		Expect(poolProvider).To(Equal(market.AddressesProvider()))

		dataProvider, err := bindaave.NewAaveProtocolDataProviderCaller(
			market.ProtocolDataProvider(),
			client,
		)
		Expect(err).NotTo(HaveOccurred())
		dataProviderAddress, err := dataProvider.ADDRESSESPROVIDER(opts)
		Expect(err).NotTo(HaveOccurred())
		Expect(dataProviderAddress).To(Equal(market.AddressesProvider()))

		addresses := []common.Address{
			market.Pool(),
			market.AddressesProvider(),
			market.ProtocolDataProvider(),
		}
		if gateway, ok := market.WrappedTokenGateway(); ok {
			addresses = append(addresses, gateway)
		}
		for _, address := range addresses {
			code, err := client.CodeAt(ctx, address, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(code).NotTo(BeEmpty(), "expected contract code at %s", address.Hex())
		}
	})
})
