package contract

import (
	"context"
	"crypto/ecdsa"
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
	"github.com/tn606024/defi-simplify/helper"
	"go.uber.org/mock/gomock"
)

type recordingActionExecutor struct {
	actions []ExecuteAction
	receipt *types.Receipt
	err     error
}

func (e *recordingActionExecutor) ExecuteActions(ctx context.Context, actions []ExecuteAction) (*types.Receipt, error) {
	if e.err != nil {
		return nil, e.err
	}
	e.actions = actions
	return e.receipt, nil
}

var _ = Describe("Base", func() {
	var (
		mockCtrl   *gomock.Controller
		mockClient *mock.MockEthereumClient
		baseClient *BaseClient
		ctx        context.Context
		privateKey *ecdsa.PrivateKey
		signer     *helper.MsgSigner
		from       common.Address
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
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Describe("BaseClient", func() {
		Context("when creating a new client", func() {
			It("should have correct chain configuration", func() {
				Expect(baseClient.chain).To(Equal(config.Base))
			})

			It("should have correct from address", func() {
				Expect(baseClient.opts.From).To(Equal(from))
			})

			It("should have signer configured", func() {
				Expect(baseClient.signer).NotTo(BeNil())
			})
		})

		It("should execute tx actions through the configured action executor", func() {
			action := BuildTransferAction(common.HexToAddress("0x123"), common.HexToAddress("0x456"), big.NewInt(1000000))
			executor := &recordingActionExecutor{
				receipt: &types.Receipt{Status: 1},
			}

			baseClient.SetActionExecutor(executor)
			receipt, err := baseClient.ExecuteTxActions(ctx, []ExecuteAction{
				NewExecuteAction(action, true),
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(receipt).To(Equal(executor.receipt))
			Expect(executor.actions).To(HaveLen(1))
			Expect(executor.actions[0].AllowFailure()).To(BeTrue())
		})
	})

	Describe("BaseClientWithConverter", func() {
		var client *BaseClientWithConverter

		BeforeEach(func() {
			client = &BaseClientWithConverter{
				BaseClient: baseClient,
			}
		})

		It("should convert wei to amount correctly", func() {
			wei := big.NewInt(1000000) // 1 USDC (6 decimals)
			amount := client.FromWei(wei, 6)
			Expect(amount.String()).To(Equal("1"))
		})

		It("should convert amount to wei correctly", func() {
			amount := decimal.NewFromFloat(1.0) // 1 USDC
			wei := client.ToWei(amount, 6)
			Expect(wei).To(Equal(big.NewInt(1000000)))
		})
	})

	Describe("executeAction", func() {
		Context("when executing a simple action", func() {
			It("should execute successfully", func() {
				action := BuildTransferAction(common.HexToAddress("0x123"), common.HexToAddress("0x456"), big.NewInt(1000000))

				receipt, err := executeAction(ctx, mockClient, baseClient.opts, action)
				Expect(err).NotTo(HaveOccurred())
				Expect(receipt).NotTo(BeNil())
				Expect(receipt.Status).To(Equal(uint64(1)))
			})
		})
	})

	Describe("BaseAction", func() {
		Context("when creating a new action", func() {
			It("should have ToDataFunc set", func() {
				action := BuildTransferAction(common.HexToAddress("0x123"), common.HexToAddress("0x456"), big.NewInt(1000000))

				Expect(action.ToDataFunc).NotTo(BeNil())
			})
		})

		It("should convert an action into a neutral call", func() {
			token := common.HexToAddress("0x123")
			action := BuildTransferAction(token, common.HexToAddress("0x456"), big.NewInt(1000000))

			call, err := action.ToCall(ctx, mockClient, baseClient.opts)

			Expect(err).NotTo(HaveOccurred())
			Expect(call.Target).To(Equal(token))
			Expect(call.Value.Sign()).To(Equal(0))
			Expect(call.Data).NotTo(BeEmpty())
		})
	})

	Describe("MulticallExecutor", func() {
		It("should convert execute actions into multicall calls outside the action abstraction", func() {
			token := common.HexToAddress("0x123")
			action := BuildTransferAction(token, common.HexToAddress("0x456"), big.NewInt(1000000))
			executor := NewMulticallExecutor(mockClient, config.Base, baseClient.opts)

			calls, err := executor.ToMulticall3Calls(ctx, []ExecuteAction{
				NewExecuteAction(action, true),
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(calls).To(HaveLen(1))
			Expect(calls[0].Target).To(Equal(token))
			Expect(calls[0].AllowFailure).To(BeTrue())
			Expect(calls[0].CallData).NotTo(BeEmpty())
		})
	})
})
