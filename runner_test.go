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
	"github.com/tn606024/defi-simplify/client/contract"
	"github.com/tn606024/defi-simplify/client/contract/mock"
	"github.com/tn606024/defi-simplify/config"
	"go.uber.org/mock/gomock"
)

var _ = Describe("Runner", func() {
	var (
		ctx        context.Context
		mockCtrl   *gomock.Controller
		mockClient *mock.MockEthereumClient
		privateKey *ecdsa.PrivateKey
		opts       *bind.TransactOpts
		user       common.Address
		receipt    *types.Receipt
	)

	BeforeEach(func() {
		ctx = context.Background()
		mockCtrl = gomock.NewController(GinkgoT())
		mockClient = mock.NewMockEthereumClient(mockCtrl)

		var err error
		privateKey, err = crypto.GenerateKey()
		Expect(err).NotTo(HaveOccurred())
		user = crypto.PubkeyToAddress(privateKey.PublicKey)

		opts, err = bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(1))
		Expect(err).NotTo(HaveOccurred())
		opts.From = user
		receipt = &types.Receipt{
			Status:      types.ReceiptStatusSuccessful,
			TxHash:      common.HexToHash("0x1234"),
			BlockNumber: big.NewInt(42),
		}

		mockClient.EXPECT().
			PendingNonceAt(gomock.Any(), user).
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
			Return(receipt, nil).
			AnyTimes()
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	It("executes a one-call flow through ExecutionEOA", func() {
		flow := NewFlow(user, WithChain(config.Base)).
			Add(&fakeFlowStep{
				name: "custom.Step",
				calls: []Call{{
					Target: common.HexToAddress("0x0000000000000000000000000000000000000010"),
					Value:  big.NewInt(0),
					Data:   []byte{0x01, 0x02},
				}},
			})
		runner := NewRunner(mockClient, opts, config.Base)

		receipt, err := runner.Execute(ctx, flow, ExecutionEOA)

		Expect(err).NotTo(HaveOccurred())
		Expect(receipt).NotTo(BeNil())
		Expect(receipt.Status).To(Equal(uint64(1)))
	})

	It("returns an unvalidated result for a successful step without expectations", func() {
		flow := NewFlow(user, WithChain(config.Base)).
			Add(&fakeFlowStep{
				name: "custom.Step",
				calls: []Call{{
					Target: common.HexToAddress("0x0000000000000000000000000000000000000010"),
					Value:  big.NewInt(0),
					Data:   []byte{0x01},
				}},
			})
		runner := NewRunner(mockClient, opts, config.Base)

		result, err := runner.ExecuteWithResult(ctx, flow, ExecutionEOA)

		Expect(err).NotTo(HaveOccurred())
		Expect(result.Receipt).To(Equal(receipt))
		Expect(result.Steps).To(HaveLen(1))
		Expect(result.Steps[0].Status).To(Equal(ValidationUnvalidated))
	})

	It("returns the mined receipt when semantic validation fails", func() {
		emitter := common.HexToAddress("0x0000000000000000000000000000000000000010")
		topic := common.HexToHash("0x1234")
		flow := NewFlow(user, WithChain(config.Base)).
			Add(&fakeFlowStep{
				name:  "custom.Step",
				calls: []Call{{Target: emitter, Value: big.NewInt(0), Data: []byte{0x01}}},
				expectations: []EventExpectation{
					&fakeEventExpectation{name: "Expected", emitter: emitter, topic: topic, expected: "value"},
				},
			})
		runner := NewRunner(mockClient, opts, config.Base)

		result, err := runner.ExecuteWithResult(ctx, flow, ExecutionEOA)

		Expect(result).NotTo(BeNil())
		Expect(result.Receipt).To(Equal(receipt))
		Expect(errors.Is(err, ErrExpectedEventNotFound)).To(BeTrue())
		Expect(result.Steps[0].Status).To(Equal(ValidationFailed))
	})

	It("returns the reverted receipt and preserves the executor sentinel", func() {
		receipt.Status = types.ReceiptStatusFailed
		flow := NewFlow(user, WithChain(config.Base)).
			Add(&fakeFlowStep{
				name:  "custom.Step",
				calls: []Call{{Target: common.HexToAddress("0x0000000000000000000000000000000000000010")}},
			})
		runner := NewRunner(mockClient, opts, config.Base)

		result, err := runner.ExecuteWithResult(ctx, flow, ExecutionEOA)

		Expect(result).NotTo(BeNil())
		Expect(result.Receipt).To(Equal(receipt))
		Expect(errors.Is(err, contract.ErrTransactionReverted)).To(BeTrue())
		Expect(result.Steps[0].Status).To(Equal(ValidationSkipped))
		Expect(result.Steps[0].SkipReason).To(Equal(SkipExecutionFailed))
		var executionErr *ExecutionError
		Expect(errors.As(err, &executionErr)).To(BeTrue())
		Expect(executionErr.Stage).To(Equal(ExecutionStageTransaction))
	})

	It("rejects multi-call flows through ExecutionEOA", func() {
		flow := NewFlow(user, WithChain(config.Base)).
			Add(&fakeFlowStep{
				name: "custom.MultiStep",
				calls: []Call{
					{Target: common.HexToAddress("0x0000000000000000000000000000000000000010"), Value: big.NewInt(0), Data: []byte{0x01}},
					{Target: common.HexToAddress("0x0000000000000000000000000000000000000020"), Value: big.NewInt(0), Data: []byte{0x02}},
				},
			})
		runner := NewRunner(mockClient, opts, config.Base)

		receipt, err := runner.Execute(ctx, flow, ExecutionEOA)

		Expect(receipt).To(BeNil())
		Expect(err).To(MatchError(ContainSubstring("direct executor requires exactly one call")))
	})

	It("executes a multi-call flow through ExecutionAtomicEOA", func() {
		implementation, err := config.Base.Simple7702AccountImplementationAddress()
		Expect(err).NotTo(HaveOccurred())
		mockClient.EXPECT().
			PendingCodeAt(ctx, user).
			Return(types.AddressToDelegation(implementation), nil)

		flow := NewFlow(user, WithChain(config.Base)).
			Add(&fakeFlowStep{
				name: "custom.MultiStep",
				calls: []Call{
					{Target: common.HexToAddress("0x0000000000000000000000000000000000000010"), Value: big.NewInt(0), Data: []byte{0x01}},
					{Target: common.HexToAddress("0x0000000000000000000000000000000000000020"), Value: big.NewInt(0), Data: []byte{0x02}},
				},
			})
		runner := NewRunner(mockClient, opts, config.Base)

		receipt, err := runner.Execute(ctx, flow, ExecutionAtomicEOA)

		Expect(err).NotTo(HaveOccurred())
		Expect(receipt).NotTo(BeNil())
		Expect(receipt.Status).To(Equal(uint64(1)))
	})

	It("rejects a flow account that does not match the transaction signer", func() {
		flowAccount := common.HexToAddress("0x00000000000000000000000000000000000000ff")
		flow := NewFlow(flowAccount, WithChain(config.Base)).
			Add(&fakeFlowStep{
				name:  "custom.Step",
				calls: []Call{{Target: common.HexToAddress("0x0000000000000000000000000000000000000010")}},
			})
		runner := NewRunner(mockClient, opts, config.Base)

		receipt, err := runner.Execute(ctx, flow, ExecutionAtomicEOA)

		Expect(receipt).To(BeNil())
		Expect(errors.Is(err, ErrExecutionAccountMismatch)).To(BeTrue())
		Expect(err.Error()).To(ContainSubstring(flowAccount.Hex()))
		Expect(err.Error()).To(ContainSubstring(user.Hex()))
	})

	It("rejects unsupported execution modes", func() {
		flow := NewFlow(user, WithChain(config.Base)).
			Add(&fakeFlowStep{
				name:  "custom.Step",
				calls: []Call{{Target: common.HexToAddress("0x0000000000000000000000000000000000000010")}},
			})
		runner := NewRunner(mockClient, opts, config.Base)

		receipt, err := runner.Execute(ctx, flow, ExecutionMode("unsupported"))

		Expect(receipt).To(BeNil())
		Expect(err).To(MatchError(ContainSubstring("unsupported execution mode")))
	})
})
