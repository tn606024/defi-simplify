package defisimplify7702_test

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tn606024/defi-simplify/client/account/defisimplify7702"
	"github.com/tn606024/defi-simplify/client/account/defisimplify7702/bindings"
	"github.com/tn606024/defi-simplify/client/account/eip7702"
	"github.com/tn606024/defi-simplify/client/contract"
	"github.com/tn606024/defi-simplify/client/contract/mock"
	"go.uber.org/mock/gomock"
)

func TestDynamicExecutorExecutesThroughReviewedDelegation(t *testing.T) {
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)
	client := mock.NewMockEthereumClient(mockCtrl)
	opts, user := newDynamicExecutorTransactor(t)
	implementation := common.HexToAddress("0x2000000000000000000000000000000000000000")
	calls := []defisimplify7702.DynamicCall{{
		Target: common.HexToAddress("0x3000000000000000000000000000000000000000"),
		Data:   []byte{0x01, 0x02},
	}}

	client.EXPECT().
		PendingCodeAt(ctx, user).
		Return(types.AddressToDelegation(implementation), nil).
		Times(2)
	client.EXPECT().
		PendingNonceAt(ctx, user).
		Return(uint64(3), nil).
		Times(2)
	client.EXPECT().
		SuggestGasPrice(ctx).
		Return(big.NewInt(1_000_000_000), nil).
		Times(2)
	client.EXPECT().
		EstimateGas(ctx, gomock.Any()).
		DoAndReturn(func(_ context.Context, msg ethereum.CallMsg) (uint64, error) {
			if msg.From != user {
				t.Fatalf("unexpected gas estimation sender: got %s want %s", msg.From.Hex(), user.Hex())
			}
			if msg.To == nil || *msg.To != user {
				t.Fatalf("dynamic transaction must target delegated EOA: got %v want %s", msg.To, user.Hex())
			}
			wantData, err := defisimplify7702.EncodeExecuteBatchDynamic(calls)
			if err != nil {
				t.Fatalf("encode expected dynamic batch: %v", err)
			}
			if string(msg.Data) != string(wantData) {
				t.Fatalf("unexpected dynamic calldata: got %x want %x", msg.Data, wantData)
			}
			return uint64(120_000), nil
		}).
		Times(2)
	client.EXPECT().
		SendTransaction(ctx, gomock.Any()).
		DoAndReturn(func(_ context.Context, tx *types.Transaction) error {
			if tx.To() == nil || *tx.To() != user {
				t.Fatalf("dynamic transaction must target delegated EOA: got %v want %s", tx.To(), user.Hex())
			}
			if tx.Value().Sign() != 0 {
				t.Fatalf("outer self-call value must be zero, got %s", tx.Value())
			}
			return nil
		}).
		Times(2)
	receipt := &types.Receipt{
		Status:      types.ReceiptStatusSuccessful,
		BlockNumber: big.NewInt(42),
	}
	client.EXPECT().
		TransactionReceipt(ctx, gomock.Any()).
		Return(receipt, nil).
		Times(2)

	executor := defisimplify7702.NewExecutor(client, opts, implementation)
	result, err := executor.ExecuteBatchDynamicWithResult(ctx, calls)

	if err != nil {
		t.Fatalf("execute dynamic calls: %v", err)
	}
	if result == nil || result.Receipt != receipt {
		t.Fatalf("unexpected execution result: %+v", result)
	}
	if result.Account != user || result.Implementation != implementation || result.CallCount != 1 {
		t.Fatalf("unexpected execution metadata: %+v", result)
	}

	second, err := executor.ExecuteBatchDynamicWithResult(ctx, calls)
	if err != nil {
		t.Fatalf("execute dynamic calls again: %v", err)
	}
	if second == nil || second.Receipt != receipt {
		t.Fatalf("unexpected second execution result: %+v", second)
	}
}

func TestDynamicExecutorRejectsMalformedCallsBeforeChainReads(t *testing.T) {
	opts, _ := newDynamicExecutorTransactor(t)
	executor := defisimplify7702.NewExecutor(
		mock.NewMockEthereumClient(gomock.NewController(t)),
		opts,
		common.HexToAddress("0x2000000000000000000000000000000000000000"),
	)

	result, err := executor.ExecuteBatchDynamicWithResult(
		context.Background(),
		[]defisimplify7702.DynamicCall{{}},
	)

	if result != nil {
		t.Fatalf("expected nil result for malformed calls, got %+v", result)
	}
	if !errors.Is(err, defisimplify7702.ErrIncompatibleDynamicCall) {
		t.Fatalf("expected incompatible dynamic call error, got %v", err)
	}
}

func TestDynamicExecutorRejectsClearedPendingDelegation(t *testing.T) {
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)
	client := mock.NewMockEthereumClient(mockCtrl)
	opts, user := newDynamicExecutorTransactor(t)
	implementation := common.HexToAddress("0x2000000000000000000000000000000000000000")

	client.EXPECT().PendingCodeAt(ctx, user).Return(nil, nil)

	executor := defisimplify7702.NewExecutor(client, opts, implementation)
	result, err := executor.ExecuteBatchDynamicWithResult(ctx, validDynamicCalls())

	if result != nil {
		t.Fatalf("expected nil result after delegation clear, got %+v", result)
	}
	if !errors.Is(err, eip7702.ErrUnexpectedDelegation) {
		t.Fatalf("expected unexpected delegation error, got %v", err)
	}
}

func TestDynamicExecutorDecodesMinedCallFailureAndPreservesReceipt(t *testing.T) {
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)
	client := mock.NewMockEthereumClient(mockCtrl)
	opts, user := newDynamicExecutorTransactor(t)
	opts.GasLimit = 500_000
	implementation := common.HexToAddress("0x2000000000000000000000000000000000000000")
	calls := validDynamicCalls()
	revertData := dynamicCallFailedData(t, big.NewInt(1), calls[0].Target, []byte{0xde, 0xad})

	client.EXPECT().
		PendingCodeAt(ctx, user).
		Return(types.AddressToDelegation(implementation), nil)
	client.EXPECT().PendingNonceAt(ctx, user).Return(uint64(3), nil)
	client.EXPECT().SuggestGasPrice(ctx).Return(big.NewInt(1_000_000_000), nil)
	client.EXPECT().SendTransaction(ctx, gomock.Any()).Return(nil)
	receipt := &types.Receipt{
		Status:      types.ReceiptStatusFailed,
		BlockNumber: big.NewInt(42),
	}
	client.EXPECT().TransactionReceipt(ctx, gomock.Any()).Return(receipt, nil)
	client.EXPECT().
		CallContract(ctx, gomock.Any(), receipt.BlockNumber).
		Return(nil, revertRPCError{data: hexutil.Encode(revertData)})

	executor := defisimplify7702.NewExecutor(client, opts, implementation)
	result, err := executor.ExecuteBatchDynamicWithResult(ctx, calls)

	if result == nil || result.Receipt != receipt {
		t.Fatalf("expected mined receipt, got %+v", result)
	}
	if !errors.Is(err, contract.ErrTransactionReverted) {
		t.Fatalf("expected transaction reverted sentinel, got %v", err)
	}
	var decoded *defisimplify7702.ContractError
	if !errors.As(err, &decoded) {
		t.Fatalf("expected decoded account error, got %v", err)
	}
	if decoded.Name != "DynamicCallFailed" || decoded.CallIndex.Cmp(big.NewInt(1)) != 0 {
		t.Fatalf("unexpected decoded account error: %+v", decoded)
	}
}

func validDynamicCalls() []defisimplify7702.DynamicCall {
	return []defisimplify7702.DynamicCall{{
		Target: common.HexToAddress("0x3000000000000000000000000000000000000000"),
		Data:   []byte{0x01},
	}}
}

type executorTestHelper interface {
	Helper()
	Fatalf(format string, args ...interface{})
}

func newDynamicExecutorTransactor(t executorTestHelper) (*bind.TransactOpts, common.Address) {
	t.Helper()
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	opts, err := bind.NewKeyedTransactorWithChainID(key, big.NewInt(8453))
	if err != nil {
		t.Fatalf("create transactor: %v", err)
	}
	return opts, crypto.PubkeyToAddress(key.PublicKey)
}

func dynamicCallFailedData(
	t *testing.T,
	index *big.Int,
	target common.Address,
	reason []byte,
) []byte {
	t.Helper()
	parsed, err := bindings.DefiSimplify7702AccountMetaData.GetAbi()
	if err != nil {
		t.Fatalf("load account ABI: %v", err)
	}
	definition := parsed.Errors["DynamicCallFailed"]
	arguments, err := definition.Inputs.Pack(index, target, reason)
	if err != nil {
		t.Fatalf("encode DynamicCallFailed: %v", err)
	}
	return append(append([]byte(nil), definition.ID[:4]...), arguments...)
}

type revertRPCError struct {
	data string
}

func (e revertRPCError) Error() string          { return "execution reverted" }
func (e revertRPCError) ErrorCode() int         { return 3 }
func (e revertRPCError) ErrorData() interface{} { return e.data }
