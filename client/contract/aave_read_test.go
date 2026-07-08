package contract

import (
	"bytes"
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tn606024/defi-simplify/client/contract/mock"
	"github.com/tn606024/defi-simplify/config"
	"go.uber.org/mock/gomock"
)

var _ = Describe("Aave read helpers", func() {
	It("queries user reserve data with asset before user in calldata", func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(GinkgoT())
		defer mockCtrl.Finish()

		mockClient := mock.NewMockEthereumClient(mockCtrl)
		user := common.HexToAddress("0x1000000000000000000000000000000000000001")
		asset := common.HexToAddress("0x2000000000000000000000000000000000000002")
		protocolDataProvider := mustAddress(config.Base.AaveProtocolDataProviderAddress())
		zeroUserReserveData := make([]byte, 32*9)

		mockClient.EXPECT().
			CallContract(gomock.Any(), gomock.Any(), (*big.Int)(nil)).
			DoAndReturn(func(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
				Expect(msg.To).NotTo(BeNil())
				Expect(*msg.To).To(Equal(protocolDataProvider))
				Expect(msg.Data[:4]).To(Equal([]byte{0x28, 0xdd, 0x2d, 0x01}))
				Expect(bytes.Equal(msg.Data[4+12:4+32], asset.Bytes())).To(BeTrue())
				Expect(bytes.Equal(msg.Data[4+32+12:4+64], user.Bytes())).To(BeTrue())
				return zeroUserReserveData, nil
			})

		baseClient := &BaseClient{
			conn:  mockClient,
			chain: config.Base,
			opts:  &bind.TransactOpts{From: user},
		}
		aaveClient := NewAaveV3Client(baseClient)

		data, err := aaveClient.GetUserReserveData(ctx, asset)

		Expect(err).NotTo(HaveOccurred())
		Expect(data).NotTo(BeNil())
	})
})
