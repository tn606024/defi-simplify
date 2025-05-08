package contract

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"math/big"
	"time"

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

var _ = Describe("ERC20", func() {
	var (
		mockCtrl   *gomock.Controller
		mockClient *mock.MockEthereumClient
		baseClient *BaseClient
		ctx        context.Context
		privateKey *ecdsa.PrivateKey
		signer     *helper.MsgSigner
		from       common.Address
		client     ERC20Interface
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

		client = NewERC20Client(baseClient)

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

	Describe("Transfer", func() {
		It("should successfully transfer USDC", func() {
			amount := decimal.NewFromFloat(1.0) // 1 USDC
			to := common.HexToAddress("0x456")
			receipt, err := client.Transfer(ctx, config.USDC, to, amount)
			Expect(err).NotTo(HaveOccurred())
			Expect(receipt).NotTo(BeNil())
			Expect(receipt.Status).To(Equal(uint64(1)))
		})
	})

	Describe("Approve", func() {
		It("should successfully approve USDC", func() {
			amount := decimal.NewFromFloat(1.0) // 1 USDC
			spender := common.HexToAddress("0x456")
			receipt, err := client.Approve(ctx, config.USDC, spender, amount)
			Expect(err).NotTo(HaveOccurred())
			Expect(receipt).NotTo(BeNil())
			Expect(receipt.Status).To(Equal(uint64(1)))
		})
	})

	Describe("TransferFrom", func() {
		It("should successfully transfer USDC from another address", func() {
			amount := decimal.NewFromFloat(1.0) // 1 USDC
			from := common.HexToAddress("0x789")
			to := common.HexToAddress("0x456")
			receipt, err := client.TransferFrom(ctx, config.USDC, from, to, amount)
			Expect(err).NotTo(HaveOccurred())
			Expect(receipt).NotTo(BeNil())
			Expect(receipt.Status).To(Equal(uint64(1)))
		})
	})

	Describe("BalanceOf", func() {
		It("should return correct balance", func() {
			expectedBalance := decimal.NewFromInt(1) // 1 USDC
			balance, err := client.BalanceOf(config.Base, config.USDC)
			Expect(err).NotTo(HaveOccurred())
			Expect(balance.String()).To(Equal(expectedBalance.String()))
		})
	})

	Describe("Allowance", func() {
		It("should return correct allowance", func() {
			expectedAllowance := decimal.NewFromInt(1) // 1 USDC
			spender := common.HexToAddress("0x456")
			allowance, err := client.Allowance(ctx, config.USDC, spender)
			Expect(err).NotTo(HaveOccurred())
			Expect(allowance.String()).To(Equal(expectedAllowance.String()))
		})
	})

	Describe("Permit", func() {
		It("should successfully permit USDC", func() {
			amount := decimal.NewFromFloat(1.0) // 1 USDC
			spender := common.HexToAddress("0x456")
			deadline := big.NewInt(time.Now().Add(time.Minute * 10).Unix())
			receipt, err := client.Permit(ctx, config.USDC, spender, amount, deadline)
			Expect(err).NotTo(HaveOccurred())
			Expect(receipt).NotTo(BeNil())
			Expect(receipt.Status).To(Equal(uint64(1)))
		})
	})

	Describe("Nonce", func() {
		It("should return correct nonce", func() {
			expectedNonce := big.NewInt(1)
			nonces, err := client.Nonces(ctx, config.USDC, from)
			Expect(err).NotTo(HaveOccurred())
			Expect(nonces.String()).To(Equal(expectedNonce.String()))
		})
	})

	Describe("ToData", func() {
		Context("TransferAction", func() {
			It("should encode transfer data correctly", func() {
				action := BuildTransferAction(
					config.USDC.Address(config.Base),
					common.HexToAddress("0x456"),
					big.NewInt(1000000),
				)
				address, data, err := action.ToData(ctx, mockClient, baseClient.opts)
				Expect(err).NotTo(HaveOccurred())
				Expect(address).To(Equal(config.USDC.Address(config.Base)))
				Expect(data).NotTo(BeEmpty())
				Expect(len(data)).To(BeNumerically(">", 4))
			})
		})

		Context("ApproveAction", func() {
			It("should encode approve data correctly", func() {
				action := BuildApproveAction(
					config.USDC.Address(config.Base),
					common.HexToAddress("0x456"),
					big.NewInt(1000000),
				)
				address, data, err := action.ToData(ctx, mockClient, baseClient.opts)
				Expect(err).NotTo(HaveOccurred())
				Expect(address).To(Equal(config.USDC.Address(config.Base)))
				Expect(data).NotTo(BeEmpty())
				Expect(len(data)).To(BeNumerically(">", 4))
			})
		})

		Context("TransferFromAction", func() {
			It("should encode transferFrom data correctly", func() {
				action := BuildTransferFromAction(
					config.USDC.Address(config.Base),
					common.HexToAddress("0x789"),
					common.HexToAddress("0x456"),
					big.NewInt(1000000),
				)
				address, data, err := action.ToData(ctx, mockClient, baseClient.opts)
				Expect(err).NotTo(HaveOccurred())
				Expect(address).To(Equal(config.USDC.Address(config.Base)))
				Expect(data).NotTo(BeEmpty())
				Expect(len(data)).To(BeNumerically(">", 4))
			})
		})

		Context("BalanceOfAction", func() {
			It("should encode balanceOf data correctly", func() {
				action := BuildBalanceOfAction(
					config.USDC.Address(config.Base),
					from,
				)
				address, data, err := action.ToData(ctx, mockClient, baseClient.opts)
				Expect(err).NotTo(HaveOccurred())
				Expect(address).To(Equal(config.USDC.Address(config.Base)))
				Expect(data).NotTo(BeEmpty())
				Expect(len(data)).To(BeNumerically(">", 4))
			})
		})

		Context("PermitAction", func() {
			It("should encode permit data correctly", func() {
				action := BuildPermitAction(
					config.USDC.Address(config.Base),
					from,
					common.HexToAddress("0x456"),
					big.NewInt(1000000),
					big.NewInt(time.Now().Add(time.Minute*10).Unix()),
					27,
					[32]byte{},
					[32]byte{},
				)
				address, data, err := action.ToData(ctx, mockClient, baseClient.opts)
				Expect(err).NotTo(HaveOccurred())
				Expect(address).To(Equal(config.USDC.Address(config.Base)))
				Expect(data).NotTo(BeEmpty())
				Expect(len(data)).To(BeNumerically(">", 4))
			})
		})

		Context("NonceAction", func() {
			It("should encode nonce data correctly", func() {
				action := BuildNoncesAction(
					config.USDC.Address(config.Base),
					from,
				)
				address, data, err := action.ToData(ctx, mockClient, baseClient.opts)
				Expect(err).NotTo(HaveOccurred())
				Expect(address).To(Equal(config.USDC.Address(config.Base)))
				Expect(data).NotTo(BeEmpty())
				Expect(len(data)).To(BeNumerically(">", 4))
			})
		})
	})
})
