package defisimplify7702_test

import (
	"context"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tn606024/defi-simplify/client/account/defisimplify7702"
	"github.com/tn606024/defi-simplify/client/account/eip7702"
	"github.com/tn606024/defi-simplify/client/contract"
	"github.com/tn606024/defi-simplify/client/contract/mock"
	"go.uber.org/mock/gomock"
)

var _ = Describe("Static account execution", func() {
	It("encodes and submits inherited executeBatch through the delegated EOA", func() {
		ctx := context.Background()
		client := mock.NewMockEthereumClient(gomock.NewController(GinkgoT()))
		opts, user := newDynamicExecutorTransactor(GinkgoT())
		implementation := common.HexToAddress("0x2000000000000000000000000000000000000000")
		calls := []contract.Call{{
			Target: common.HexToAddress("0x3000000000000000000000000000000000000000"),
			Value:  big.NewInt(7),
			Data:   []byte{0x01, 0x02},
		}}
		expectedData, err := defisimplify7702.EncodeExecuteBatch(calls)
		Expect(err).NotTo(HaveOccurred())

		client.EXPECT().
			PendingCodeAt(ctx, user).
			Return(types.AddressToDelegation(implementation), nil)
		client.EXPECT().PendingNonceAt(ctx, user).Return(uint64(3), nil)
		client.EXPECT().SuggestGasPrice(ctx).Return(big.NewInt(1_000_000_000), nil)
		client.EXPECT().
			EstimateGas(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, msg ethereum.CallMsg) (uint64, error) {
				Expect(msg.From).To(Equal(user))
				Expect(msg.To).NotTo(BeNil())
				Expect(*msg.To).To(Equal(user))
				Expect(msg.Value.Sign()).To(BeZero())
				Expect(msg.Data).To(Equal(expectedData))
				return uint64(120_000), nil
			})
		client.EXPECT().
			SendTransaction(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, tx *types.Transaction) error {
				Expect(tx.To()).NotTo(BeNil())
				Expect(*tx.To()).To(Equal(user))
				Expect(tx.Value().Sign()).To(BeZero())
				Expect(tx.Data()).To(Equal(expectedData))
				return nil
			})
		receipt := &types.Receipt{
			Status:      types.ReceiptStatusSuccessful,
			BlockNumber: big.NewInt(42),
		}
		client.EXPECT().TransactionReceipt(ctx, gomock.Any()).Return(receipt, nil)

		result, err := defisimplify7702.NewExecutor(client, opts, implementation).
			ExecuteBatchWithResult(ctx, calls)

		Expect(err).NotTo(HaveOccurred())
		Expect(result).NotTo(BeNil())
		Expect(result.Receipt).To(BeIdenticalTo(receipt))
		Expect(result.Account).To(Equal(user))
		Expect(result.Implementation).To(Equal(implementation))
		Expect(result.CallCount).To(Equal(1))
	})

	It("rejects malformed static calls before chain access", func() {
		client := mock.NewMockEthereumClient(gomock.NewController(GinkgoT()))
		opts, _ := newDynamicExecutorTransactor(GinkgoT())
		executor := defisimplify7702.NewExecutor(
			client,
			opts,
			common.HexToAddress("0x2000000000000000000000000000000000000000"),
		)

		result, err := executor.ExecuteBatchWithResult(
			context.Background(),
			[]contract.Call{{}},
		)

		Expect(result).To(BeNil())
		Expect(errors.Is(err, defisimplify7702.ErrIncompatibleStaticCall)).To(BeTrue())
	})

	It("rejects static execution when the pending delegation target changed", func() {
		ctx := context.Background()
		client := mock.NewMockEthereumClient(gomock.NewController(GinkgoT()))
		opts, user := newDynamicExecutorTransactor(GinkgoT())
		implementation := common.HexToAddress("0x2000000000000000000000000000000000000000")
		client.EXPECT().PendingCodeAt(ctx, user).Return(nil, nil)

		result, err := defisimplify7702.NewExecutor(client, opts, implementation).
			ExecuteBatchWithResult(ctx, []contract.Call{{
				Target: common.HexToAddress("0x3000000000000000000000000000000000000000"),
			}})

		Expect(result).To(BeNil())
		Expect(errors.Is(err, eip7702.ErrUnexpectedDelegation)).To(BeTrue())
	})
})
