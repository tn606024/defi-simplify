package contract

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/shopspring/decimal"
	"github.com/tn606024/defi-simplify/client/contract/mock"
	"github.com/tn606024/defi-simplify/config"
	"github.com/tn606024/defi-simplify/helper"
	"go.uber.org/mock/gomock"
)

var _ = Describe("Defi", func() {
	var (
		mockCtrl   *gomock.Controller
		mockClient *mock.MockEthereumClient
		ctx        context.Context
		privateKey *ecdsa.PrivateKey
		signer     *helper.MsgSigner
		from       common.Address
		defiClient *DefiClient
	)

	BeforeEach(func() {
		ctx = context.Background()
		mockCtrl = gomock.NewController(GinkgoT())
		mockClient = mock.NewMockEthereumClient(mockCtrl)

		// Generate a test private key
		var err error
		privateKey, err = crypto.GenerateKey()
		Expect(err).NotTo(HaveOccurred())

		// Get the address from the private key
		from = crypto.PubkeyToAddress(privateKey.PublicKey)

		// Create signer
		signer = helper.NewMsgSigner(privateKey)

		// Create auth with the private key
		auth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(1))
		Expect(err).NotTo(HaveOccurred())
		auth.From = from

		defiClient = NewDefiClient(auth, mockClient, signer, config.Base)

		mockClient.EXPECT().
			HeaderByNumber(gomock.Any(), gomock.Any()).
			Return(&types.Header{Number: big.NewInt(1)}, nil).
			AnyTimes()

		mockClient.EXPECT().
			PendingCodeAt(gomock.Any(), gomock.Any()).
			Return([]byte{0x00, 0x00}, nil).
			AnyTimes()

		mockClient.EXPECT().
			PendingNonceAt(gomock.Any(), gomock.Any()).
			Return(uint64(1), nil).
			AnyTimes()

		mockClient.EXPECT().
			SuggestGasPrice(gomock.Any()).
			Return(big.NewInt(1000000000), nil).
			AnyTimes()

		mockClient.EXPECT().
			EstimateGas(gomock.Any(), gomock.Any()).
			Return(uint64(21000), nil).
			AnyTimes()

		mockClient.EXPECT().
			SendTransaction(gomock.Any(), gomock.Any()).
			Return(nil).
			AnyTimes()

		mockClient.EXPECT().
			TransactionReceipt(gomock.Any(), gomock.Any()).
			Return(&types.Receipt{Status: 1}, nil).
			AnyTimes()

		mockClient.EXPECT().
			CodeAt(gomock.Any(), gomock.Any(), gomock.Any()).
			Return([]byte{0x00, 0x01}, nil).
			AnyTimes()

		mockClient.EXPECT().
			CallContract(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
				// For nonce calls, return a properly encoded uint256 of 1
				if len(msg.Data) >= 4 && bytes.Equal(msg.Data[:4], []byte{0x7e, 0xce, 0xbe, 0x00}) { // nonces(address) selector
					return common.Hex2Bytes("0000000000000000000000000000000000000000000000000000000000000001"), nil
				}
				// For allowance calls, return a properly encoded uint256 of 1000000
				if len(msg.Data) >= 4 && bytes.Equal(msg.Data[:4], []byte{0xdd, 0x62, 0xed, 0x3e}) { // allowance(address,address) selector
					return common.Hex2Bytes("00000000000000000000000000000000000000000000000000000000000f4240"), nil
				}
				// For balanceOf calls, return a properly encoded uint256 of 1000000
				if len(msg.Data) >= 4 && bytes.Equal(msg.Data[:4], []byte{0x70, 0xa0, 0x82, 0x31}) { // balanceOf(address) selector
					return common.Hex2Bytes("00000000000000000000000000000000000000000000000000000000000f4240"), nil
				}
				// For other calls, return a properly encoded uint256 of 0
				return common.Hex2Bytes("0000000000000000000000000000000000000000000000000000000000000000"), nil
			}).
			AnyTimes()
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Describe("SupplyAndBorrowAaveV3Coin", func() {
		It("should successfully supply and borrow USDC", func() {
			supplyAmount := decimal.NewFromFloat(1.0) // 1 USDC
			borrowAmount := decimal.NewFromFloat(0.5) // 0.5 USDC
			receipt, err := defiClient.SupplyAndBorrowAaveV3Coin(ctx, config.USDC, supplyAmount, borrowAmount)
			Expect(err).NotTo(HaveOccurred())
			Expect(receipt).NotTo(BeNil())
			Expect(receipt.Status).To(Equal(uint64(1)))
		})
	})
})
