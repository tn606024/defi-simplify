package defi

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
	"github.com/tn606024/defi-simplify/amount"
	"github.com/tn606024/defi-simplify/client/account/defisimplify7702"
	"github.com/tn606024/defi-simplify/client/account/defisimplify7702/bindings"
	"github.com/tn606024/defi-simplify/client/contract"
	"github.com/tn606024/defi-simplify/client/contract/mock"
	"github.com/tn606024/defi-simplify/config"
	"github.com/tn606024/defi-simplify/token"
	"go.uber.org/mock/gomock"
)

var _ = Describe("Runner", func() {
	var (
		ctx            context.Context
		mockClient     *mock.MockEthereumClient
		privateKey     *ecdsa.PrivateKey
		opts           *bind.TransactOpts
		user           common.Address
		implementation common.Address
		receipt        *types.Receipt
		sentTxs        int
		sentTo         common.Address
		sentData       []byte
	)

	BeforeEach(func() {
		ctx = context.Background()
		mockClient = mock.NewMockEthereumClient(gomock.NewController(GinkgoT()))

		var err error
		privateKey, err = crypto.GenerateKey()
		Expect(err).NotTo(HaveOccurred())
		user = crypto.PubkeyToAddress(privateKey.PublicKey)
		opts, err = bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(1))
		Expect(err).NotTo(HaveOccurred())
		opts.From = user

		deployment, err := defisimplify7702.DeploymentForChain(config.Base)
		Expect(err).NotTo(HaveOccurred())
		accountDeployment, err := deployment.Contract(defisimplify7702.AccountContract)
		Expect(err).NotTo(HaveOccurred())
		implementation = accountDeployment.Address

		receipt = &types.Receipt{
			Status:      types.ReceiptStatusSuccessful,
			TxHash:      common.HexToHash("0x1234"),
			BlockNumber: big.NewInt(42),
		}
		sentTxs = 0
		sentTo = common.Address{}
		sentData = nil

		mockClient.EXPECT().
			PendingCodeAt(gomock.Any(), user).
			Return(types.AddressToDelegation(implementation), nil).
			AnyTimes()
		mockClient.EXPECT().
			PendingNonceAt(gomock.Any(), user).
			Return(uint64(1), nil).
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
			DoAndReturn(func(_ context.Context, tx *types.Transaction) error {
				sentTxs++
				if tx.To() != nil {
					sentTo = *tx.To()
				}
				sentData = append([]byte(nil), tx.Data()...)
				return nil
			}).
			AnyTimes()
		mockClient.EXPECT().
			TransactionReceipt(gomock.Any(), gomock.Any()).
			Return(receipt, nil).
			AnyTimes()
		mockClient.EXPECT().
			CallContract(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil, nil).
			AnyTimes()
	})

	It("automatically dispatches static plans to inherited executeBatch", func() {
		flow := NewFlow(user, WithChain(config.Base)).
			Add(&fakeFlowStep{
				name: "custom.Static",
				calls: []Call{
					{
						Target: common.HexToAddress("0x0000000000000000000000000000000000000010"),
						Data:   []byte{0x01},
					},
					{
						Target: common.HexToAddress("0x0000000000000000000000000000000000000020"),
						Value:  big.NewInt(7),
						Data:   []byte{0x02},
					},
				},
			})

		result, err := NewRunner(mockClient, opts, config.Base).Execute(ctx, flow)

		Expect(err).NotTo(HaveOccurred())
		Expect(result).NotTo(BeNil())
		Expect(result.Receipt).To(Equal(receipt))
		Expect(result.Steps).To(HaveLen(1))
		Expect(result.Steps[0].Status).To(Equal(ValidationUnvalidated))
		Expect(sentTxs).To(Equal(1))
		Expect(sentTo).To(Equal(user))
		Expect(sentData[:4]).To(Equal(accountMethodSelector("executeBatch")))
	})

	It("automatically dispatches runtime plans to executeBatchDynamic", func() {
		asset, err := token.NewRef(
			config.Base,
			common.HexToAddress("0x0000000000000000000000000000000000000101"),
		)
		Expect(err).NotTo(HaveOccurred())
		flow := NewFlow(user, WithChain(config.Base)).
			Add(&fakeFlowStep{
				name: "custom.Dynamic",
				plannedCalls: []PlannedCall{{
					Call: Call{
						Target: common.HexToAddress("0x0000000000000000000000000000000000000010"),
						Data:   calldataWithWords(1),
					},
					Patches: []CalldataPatch{{
						Source: amount.CurrentBalance(asset),
						Offset: 4,
					}},
				}},
			})

		result, err := NewRunner(mockClient, opts, config.Base).Execute(ctx, flow)

		Expect(err).NotTo(HaveOccurred())
		Expect(result).NotTo(BeNil())
		Expect(result.Receipt).To(Equal(receipt))
		Expect(sentTxs).To(Equal(1))
		Expect(sentTo).To(Equal(user))
		Expect(sentData[:4]).To(Equal(accountMethodSelector("executeBatchDynamic")))
	})

	It("rejects ambiguous semantic validation before sending a transaction", func() {
		emitter := common.HexToAddress("0x0000000000000000000000000000000000000010")
		topic := common.HexToHash("0x1234")
		flow := NewFlow(user, WithChain(config.Base)).
			Add(&fakeFlowStep{
				name:  "custom.Unvalidated",
				calls: []Call{{Target: emitter, Data: []byte{0x01}}},
			}).
			Add(&fakeFlowStep{
				name:  "custom.Validated",
				calls: []Call{{Target: emitter, Data: []byte{0x02}}},
				expectations: []EventExpectation{
					&fakeEventExpectation{name: "Expected", emitter: emitter, topic: topic, expected: "value"},
				},
			})

		result, err := NewRunner(mockClient, opts, config.Base).Execute(ctx, flow)

		Expect(result).To(BeNil())
		Expect(errors.Is(err, ErrInvalidExecutionPlan)).To(BeTrue())
		Expect(sentTxs).To(BeZero())
		var executionErr *ExecutionError
		Expect(errors.As(err, &executionErr)).To(BeTrue())
		Expect(executionErr.Stage).To(Equal(ExecutionStageValidation))
	})

	It("returns the mined receipt when semantic validation fails", func() {
		emitter := common.HexToAddress("0x0000000000000000000000000000000000000010")
		topic := common.HexToHash("0x1234")
		flow := NewFlow(user, WithChain(config.Base)).
			Add(&fakeFlowStep{
				name:  "custom.Validated",
				calls: []Call{{Target: emitter, Data: []byte{0x01}}},
				expectations: []EventExpectation{
					&fakeEventExpectation{name: "Expected", emitter: emitter, topic: topic, expected: "value"},
				},
			})

		result, err := NewRunner(mockClient, opts, config.Base).Execute(ctx, flow)

		Expect(result).NotTo(BeNil())
		Expect(result.Receipt).To(Equal(receipt))
		Expect(errors.Is(err, ErrExpectedEventNotFound)).To(BeTrue())
		Expect(result.Steps[0].Status).To(Equal(ValidationFailed))
	})

	It("returns the reverted receipt and preserves the transaction sentinel", func() {
		receipt.Status = types.ReceiptStatusFailed
		flow := NewFlow(user, WithChain(config.Base)).
			Add(&fakeFlowStep{
				name:  "custom.Static",
				calls: []Call{{Target: common.HexToAddress("0x0000000000000000000000000000000000000010")}},
			})

		result, err := NewRunner(mockClient, opts, config.Base).Execute(ctx, flow)

		Expect(result).NotTo(BeNil())
		Expect(result.Receipt).To(Equal(receipt))
		Expect(errors.Is(err, contract.ErrTransactionReverted)).To(BeTrue())
		Expect(result.Steps[0].Status).To(Equal(ValidationSkipped))
		Expect(result.Steps[0].SkipReason).To(Equal(SkipExecutionFailed))
		var executionErr *ExecutionError
		Expect(errors.As(err, &executionErr)).To(BeTrue())
		Expect(executionErr.Stage).To(Equal(ExecutionStageTransaction))
	})

	It("rejects a flow account that does not match the transaction signer", func() {
		flowAccount := common.HexToAddress("0x00000000000000000000000000000000000000ff")
		flow := NewFlow(flowAccount, WithChain(config.Base)).
			Add(&fakeFlowStep{
				name:  "custom.Static",
				calls: []Call{{Target: common.HexToAddress("0x0000000000000000000000000000000000000010")}},
			})

		result, err := NewRunner(mockClient, opts, config.Base).Execute(ctx, flow)

		Expect(result).To(BeNil())
		Expect(errors.Is(err, ErrExecutionAccountMismatch)).To(BeTrue())
		Expect(err.Error()).To(ContainSubstring(flowAccount.Hex()))
		Expect(err.Error()).To(ContainSubstring(user.Hex()))
		Expect(sentTxs).To(BeZero())
	})
})

func accountMethodSelector(name string) []byte {
	parsed, err := bindings.DefiSimplify7702AccountMetaData.GetAbi()
	Expect(err).NotTo(HaveOccurred())
	method, ok := parsed.Methods[name]
	Expect(ok).To(BeTrue())
	return append([]byte(nil), method.ID...)
}
