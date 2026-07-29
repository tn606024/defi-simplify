package contract

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/shopspring/decimal"
	"github.com/tn606024/defi-simplify/client/contract/mock"
	"github.com/tn606024/defi-simplify/config"
	"go.uber.org/mock/gomock"
)

var _ = Describe("Base", func() {
	var (
		mockClient    *mock.MockEthereumClient
		baseClient    *BaseClient
		ctx           context.Context
		privateKey    *ecdsa.PrivateKey
		from          common.Address
		receiptStatus uint64
	)

	BeforeEach(func() {
		ctx = context.Background()
		receiptStatus = types.ReceiptStatusSuccessful
		mockClient = mock.NewMockEthereumClient(gomock.NewController(GinkgoT()))

		var err error
		privateKey, err = crypto.GenerateKey()
		Expect(err).NotTo(HaveOccurred())
		from = crypto.PubkeyToAddress(privateKey.PublicKey)
		auth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(1))
		Expect(err).NotTo(HaveOccurred())
		auth.From = from

		baseClient = NewBaseClient(mockClient, config.Base, auth)

		mockClient.EXPECT().
			PendingNonceAt(gomock.Any(), gomock.Any()).
			Return(uint64(1), nil).
			AnyTimes()
		mockClient.EXPECT().
			HeaderByNumber(gomock.Any(), gomock.Any()).
			Return(&types.Header{Number: big.NewInt(1)}, nil).
			AnyTimes()
		mockClient.EXPECT().
			PendingCodeAt(gomock.Any(), gomock.Any()).
			Return([]byte{0x01}, nil).
			AnyTimes()
		mockClient.EXPECT().
			SuggestGasPrice(gomock.Any()).
			Return(big.NewInt(1_000_000_000), nil).
			AnyTimes()
		mockClient.EXPECT().
			EstimateGas(gomock.Any(), gomock.Any()).
			Return(uint64(21_000), nil).
			AnyTimes()
		mockClient.EXPECT().
			SendTransaction(gomock.Any(), gomock.Any()).
			Return(nil).
			AnyTimes()
		mockClient.EXPECT().
			TransactionReceipt(gomock.Any(), gomock.Any()).
			DoAndReturn(func(context.Context, common.Hash) (*types.Receipt, error) {
				return &types.Receipt{Status: receiptStatus}, nil
			}).
			AnyTimes()
	})

	It("keeps chain and signer context for legacy protocol clients", func() {
		Expect(baseClient.chain).To(Equal(config.Base))
		Expect(baseClient.opts.From).To(Equal(from))
	})

	It("converts amounts between human units and wei", func() {
		client := &BaseClientWithConverter{BaseClient: baseClient}

		Expect(client.FromWei(big.NewInt(1_000_000), 6).String()).To(Equal("1"))
		Expect(client.ToWei(decimal.NewFromInt(1), 6)).To(Equal(big.NewInt(1_000_000)))
	})

	It("executes individual legacy actions without a batch backend", func() {
		action := BuildTransferAction(
			common.HexToAddress("0x123"),
			common.HexToAddress("0x456"),
			big.NewInt(1_000_000),
		)

		receipt, err := executeAction(ctx, mockClient, baseClient.opts, action)

		Expect(err).NotTo(HaveOccurred())
		Expect(receipt.Status).To(Equal(uint64(types.ReceiptStatusSuccessful)))
	})

	It("encodes Actions into neutral Calls", func() {
		token := common.HexToAddress("0x123")
		action := BuildTransferAction(token, common.HexToAddress("0x456"), big.NewInt(1_000_000))

		call, err := action.ToCall(ctx, mockClient, baseClient.opts)

		Expect(err).NotTo(HaveOccurred())
		Expect(call.Target).To(Equal(token))
		Expect(call.Value.Sign()).To(BeZero())
		Expect(call.Data).NotTo(BeEmpty())
	})

	Describe("SendCall", func() {
		It("submits one low-level call with native value", func() {
			receipt, err := SendCall(ctx, mockClient, baseClient.opts, Call{
				Target: common.HexToAddress("0x123"),
				Value:  big.NewInt(123),
				Data:   []byte{0x01, 0x02},
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(receipt.Status).To(Equal(uint64(types.ReceiptStatusSuccessful)))
		})

		It("returns a mined reverted receipt with the transaction sentinel", func() {
			receiptStatus = types.ReceiptStatusFailed

			receipt, err := SendCall(ctx, mockClient, baseClient.opts, Call{
				Target: common.HexToAddress("0x123"),
				Data:   []byte{0x01, 0x02},
			})

			Expect(receipt).NotTo(BeNil())
			Expect(receipt.Status).To(Equal(uint64(types.ReceiptStatusFailed)))
			Expect(errors.Is(err, ErrTransactionReverted)).To(BeTrue())
		})
	})
})
