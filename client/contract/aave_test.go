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

var _ = Describe("AaveV3", func() {
	var (
		mockCtrl   *gomock.Controller
		mockClient *mock.MockEthereumClient
		baseClient *BaseClient
		ctx        context.Context
		privateKey *ecdsa.PrivateKey
		signer     *helper.MsgSigner
		from       common.Address
		aaveClient AaveV3Interface
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

		baseClient = &BaseClient{
			conn:   mockClient,
			chain:  config.Base,
			opts:   auth,
			signer: signer,
		}

		aaveClient = NewAaveV3Client(baseClient)

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
				if len(msg.Data) >= 4 && bytes.Equal(msg.Data[:4], []byte{0x35, 0xea, 0x6a, 0x75}) { // getReserveData(address) selector
					return common.Hex2Bytes("100000000000000000000003e800e4e1c0000aa22b0003e8850629041e781d4c0000000000000000000000000000000000000000038461e1720ad3b5c0b25e6d0000000000000000000000000000000000000000001c4f7d19ac492dc5f3921a0000000000000000000000000000000000000000039ef96deb142fb06faf641700000000000000000000000000000000000000000027e0491ba08e6ae733ab4e0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000006826e6f300000000000000000000000000000000000000000000000000000000000000040000000000000000000000004e65fe4dba92790696d040ac24aa414708f5c0ab000000000000000000000000aed3b56fea82e809665f02acbcdec0816c75f4d900000000000000000000000059dca05b6c26dbd64b5381374aaac5cd05644c2800000000000000000000000086ab1c62a8bf868e1b3e1ab87d587aba6fbcbdc50000000000000000000000000000000000000000000000000000000004023a8c00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"), nil
				}
				if len(msg.Data) >= 4 && bytes.Equal(msg.Data[:4], []byte{0xbf, 0x92, 0x85, 0x7c}) { // getUserAccountData(address) selector
					return common.Hex2Bytes("000000000000000000000000000000000000000000000000000000000be8304600000000000000000000000000000000000000000000000000000000000000630000000000000000000000000000000000000000000000000000000008ee23d20000000000000000000000000000000000000000000000000000000000001e780000000000000000000000000000000000000000000000000000000000001d4c000000000000000000000000000000000000000000014d4a151219f63fb33122"), nil
				}
				// For other calls, return a properly encoded uint256 of 0
				return common.Hex2Bytes("0000000000000000000000000000000000000000000000000000000000000000"), nil
			}).
			AnyTimes()
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Describe("Supply", func() {
		It("should successfully supply USDC", func() {
			amount := decimal.NewFromFloat(1.0) // 1 USDC
			receipt, err := aaveClient.Supply(ctx, config.USDC, amount)
			Expect(err).NotTo(HaveOccurred())
			Expect(receipt).NotTo(BeNil())
			Expect(receipt.Status).To(Equal(uint64(1)))
		})
	})

	Describe("Withdraw", func() {
		It("should successfully withdraw USDC", func() {
			amount := decimal.NewFromFloat(1.0) // 1 USDC
			receipt, err := aaveClient.Withdraw(ctx, config.USDC, amount)
			Expect(err).NotTo(HaveOccurred())
			Expect(receipt).NotTo(BeNil())
			Expect(receipt.Status).To(Equal(uint64(1)))
		})
	})

	Describe("Borrow", func() {
		It("should successfully borrow USDC", func() {
			amount := decimal.NewFromFloat(1.0) // 1 USDC
			receipt, err := aaveClient.Borrow(ctx, config.USDC, amount)
			Expect(err).NotTo(HaveOccurred())
			Expect(receipt).NotTo(BeNil())
			Expect(receipt.Status).To(Equal(uint64(1)))
		})
	})

	Describe("BorrowETH", func() {
		It("should successfully borrow ETH", func() {
			amount := decimal.NewFromFloat(1.0) // 1 ETH
			receipt, err := aaveClient.BorrowETH(ctx, amount)
			Expect(err).NotTo(HaveOccurred())
			Expect(receipt).NotTo(BeNil())
			Expect(receipt.Status).To(Equal(uint64(1)))
		})
	})

	Describe("Repay", func() {
		It("should successfully repay USDC", func() {
			amount := decimal.NewFromFloat(1.0) // 1 USDC
			receipt, err := aaveClient.Repay(ctx, config.USDC, amount)
			Expect(err).NotTo(HaveOccurred())
			Expect(receipt).NotTo(BeNil())
			Expect(receipt.Status).To(Equal(uint64(1)))
		})
	})

	Describe("DepositETH", func() {
		It("should successfully deposit ETH", func() {
			amount := decimal.NewFromFloat(1.0) // 1 ETH
			receipt, err := aaveClient.DepositETH(ctx, amount)
			Expect(err).NotTo(HaveOccurred())
			Expect(receipt).NotTo(BeNil())
			Expect(receipt.Status).To(Equal(uint64(1)))
		})
	})

	Describe("WithdrawETH", func() {
		It("should successfully withdraw ETH", func() {
			amount := decimal.NewFromFloat(1.0) // 1 ETH
			receipt, err := aaveClient.WithdrawETH(ctx, amount)
			Expect(err).NotTo(HaveOccurred())
			Expect(receipt).NotTo(BeNil())
			Expect(receipt.Status).To(Equal(uint64(1)))
		})
	})

	Describe("ApproveDelegation", func() {
		It("should successfully approve delegation", func() {
			amount := decimal.NewFromFloat(1.0) // 1 USDC
			delegatee := common.HexToAddress("0x1234567890123456789012345678901234567890")
			receipt, err := aaveClient.ApproveDelegation(ctx, config.USDC, delegatee, amount)
			Expect(err).NotTo(HaveOccurred())
			Expect(receipt).NotTo(BeNil())
			Expect(receipt.Status).To(Equal(uint64(1)))
		})
	})

	Describe("DelegationWithSig", func() {
		It("should successfully delegate with signature", func() {
			amount := decimal.NewFromFloat(1.0) // 1 USDC
			delegatee := common.HexToAddress("0x1234567890123456789012345678901234567890")
			receipt, err := aaveClient.DelegationWithSig(ctx, config.USDC, delegatee, amount)
			Expect(err).NotTo(HaveOccurred())
			Expect(receipt).NotTo(BeNil())
			Expect(receipt.Status).To(Equal(uint64(1)))
		})
	})

	Describe("GetReserveData", func() {
		It("should successfully get reserve data", func() {
			data, err := aaveClient.GetReserveData(ctx, config.USDC)
			Expect(err).NotTo(HaveOccurred())
			Expect(data).NotTo(BeNil())
		})
	})
})
