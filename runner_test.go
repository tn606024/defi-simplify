package defi

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
			Return(&types.Receipt{Status: 1}, nil).
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
