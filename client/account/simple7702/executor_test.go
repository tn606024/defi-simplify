package simple7702_test

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tn606024/defi-simplify/client/account/eip7702"
	"github.com/tn606024/defi-simplify/client/account/simple7702"
	"github.com/tn606024/defi-simplify/client/contract"
	"github.com/tn606024/defi-simplify/client/contract/mock"
	"go.uber.org/mock/gomock"
)

func TestExecutorExecutesNeutralCallsThroughDelegatedEOA(t *testing.T) {
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)
	client := mock.NewMockEthereumClient(mockCtrl)
	opts, user := newExecutorTransactor(t)
	implementation := common.HexToAddress("0x2000000000000000000000000000000000000000")
	calls := []contract.Call{
		{Target: common.HexToAddress("0x3000000000000000000000000000000000000000"), Data: []byte{0x01, 0x02}},
		{Target: common.HexToAddress("0x4000000000000000000000000000000000000000"), Value: big.NewInt(7), Data: []byte{0x03, 0x04}},
	}

	client.EXPECT().
		PendingCodeAt(ctx, user).
		Return(types.AddressToDelegation(implementation), nil)
	client.EXPECT().
		PendingNonceAt(ctx, user).
		Return(uint64(3), nil)
	client.EXPECT().
		SuggestGasPrice(ctx).
		Return(big.NewInt(1_000_000_000), nil)
	client.EXPECT().
		EstimateGas(ctx, gomock.Any()).
		DoAndReturn(func(_ context.Context, msg ethereum.CallMsg) (uint64, error) {
			if msg.From != user {
				t.Fatalf("unexpected gas estimation sender: got %s want %s", msg.From.Hex(), user.Hex())
			}
			if msg.To == nil || *msg.To != user {
				t.Fatalf("batch transaction must target delegated EOA: got %v want %s", msg.To, user.Hex())
			}
			wantData, err := simple7702.EncodeExecuteBatch(calls)
			if err != nil {
				t.Fatalf("encode expected batch: %v", err)
			}
			if string(msg.Data) != string(wantData) {
				t.Fatalf("unexpected batch calldata: got %x want %x", msg.Data, wantData)
			}
			return uint64(120_000), nil
		})
	client.EXPECT().
		SendTransaction(ctx, gomock.Any()).
		DoAndReturn(func(_ context.Context, tx *types.Transaction) error {
			if tx.To() == nil || *tx.To() != user {
				t.Fatalf("batch transaction must target delegated EOA: got %v want %s", tx.To(), user.Hex())
			}
			if tx.Value().Sign() != 0 {
				t.Fatalf("outer self-call value must be zero, got %s", tx.Value())
			}
			return nil
		})
	receipt := &types.Receipt{Status: types.ReceiptStatusSuccessful}
	client.EXPECT().
		TransactionReceipt(ctx, gomock.Any()).
		Return(receipt, nil)

	executor := simple7702.NewExecutor(client, opts, implementation)
	result, err := executor.ExecuteCallsWithResult(ctx, calls)
	if err != nil {
		t.Fatalf("execute calls: %v", err)
	}
	if result.Receipt != receipt {
		t.Fatal("unexpected execution receipt")
	}
	if result.Account != user {
		t.Fatalf("unexpected execution account: got %s want %s", result.Account.Hex(), user.Hex())
	}
	if result.Implementation != implementation {
		t.Fatalf("unexpected implementation: got %s want %s", result.Implementation.Hex(), implementation.Hex())
	}
	if result.CallCount != len(calls) {
		t.Fatalf("unexpected call count: got %d want %d", result.CallCount, len(calls))
	}
}

func TestExecutorRejectsUnexpectedDelegation(t *testing.T) {
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)
	client := mock.NewMockEthereumClient(mockCtrl)
	opts, user := newExecutorTransactor(t)
	expected := common.HexToAddress("0x2000000000000000000000000000000000000000")
	actual := common.HexToAddress("0x3000000000000000000000000000000000000000")

	client.EXPECT().
		PendingCodeAt(ctx, user).
		Return(types.AddressToDelegation(actual), nil)

	executor := simple7702.NewExecutor(client, opts, expected)
	receipt, err := executor.ExecuteCalls(ctx, []contract.Call{{
		Target: common.HexToAddress("0x4000000000000000000000000000000000000000"),
	}})
	if receipt != nil {
		t.Fatal("unexpected receipt for invalid delegation")
	}
	if err == nil {
		t.Fatal("expected invalid delegation error")
	}
	if !errors.Is(err, eip7702.ErrUnexpectedDelegation) {
		t.Fatalf("expected unexpected delegation error, got %v", err)
	}
	if !strings.Contains(err.Error(), expected.Hex()) || !strings.Contains(err.Error(), actual.Hex()) {
		t.Fatalf("expected delegation error to include expected and actual implementations, got %v", err)
	}
}

func TestExecutorReturnsMetadataForRevertedBatch(t *testing.T) {
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)
	client := mock.NewMockEthereumClient(mockCtrl)
	opts, user := newExecutorTransactor(t)
	implementation := common.HexToAddress("0x2000000000000000000000000000000000000000")
	calls := []contract.Call{{Target: common.HexToAddress("0x3000000000000000000000000000000000000000")}}

	client.EXPECT().PendingCodeAt(ctx, user).Return(types.AddressToDelegation(implementation), nil)
	client.EXPECT().PendingNonceAt(ctx, user).Return(uint64(3), nil)
	client.EXPECT().SuggestGasPrice(ctx).Return(big.NewInt(1_000_000_000), nil)
	client.EXPECT().EstimateGas(ctx, gomock.Any()).Return(uint64(120_000), nil)
	client.EXPECT().SendTransaction(ctx, gomock.Any()).Return(nil)
	receipt := &types.Receipt{Status: types.ReceiptStatusFailed}
	client.EXPECT().TransactionReceipt(ctx, gomock.Any()).Return(receipt, nil)

	executor := simple7702.NewExecutor(client, opts, implementation)
	result, err := executor.ExecuteCallsWithResult(ctx, calls)

	if result == nil {
		t.Fatal("expected execution metadata for reverted batch")
	}
	if result.Receipt != receipt {
		t.Fatal("expected reverted receipt in execution result")
	}
	if result.Account != user || result.Implementation != implementation || result.CallCount != 1 {
		t.Fatalf("unexpected reverted execution metadata: %+v", result)
	}
	if !errors.Is(err, contract.ErrTransactionReverted) {
		t.Fatalf("expected transaction reverted error, got %v", err)
	}
}

func TestExecutorReturnsNilResultWhenSubmissionFails(t *testing.T) {
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)
	client := mock.NewMockEthereumClient(mockCtrl)
	opts, user := newExecutorTransactor(t)
	implementation := common.HexToAddress("0x2000000000000000000000000000000000000000")
	calls := []contract.Call{{Target: common.HexToAddress("0x3000000000000000000000000000000000000000")}}
	estimationErr := errors.New("gas estimation failed")

	client.EXPECT().PendingCodeAt(ctx, user).Return(types.AddressToDelegation(implementation), nil)
	client.EXPECT().PendingNonceAt(ctx, user).Return(uint64(3), nil)
	client.EXPECT().SuggestGasPrice(ctx).Return(big.NewInt(1_000_000_000), nil)
	client.EXPECT().EstimateGas(ctx, gomock.Any()).Return(uint64(0), estimationErr)

	executor := simple7702.NewExecutor(client, opts, implementation)
	result, err := executor.ExecuteCallsWithResult(ctx, calls)

	if result != nil {
		t.Fatalf("expected nil result before transaction submission, got %+v", result)
	}
	if !errors.Is(err, estimationErr) {
		t.Fatalf("expected gas estimation error, got %v", err)
	}
}

func TestExecutorRejectsClearedPendingDelegation(t *testing.T) {
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)
	client := mock.NewMockEthereumClient(mockCtrl)
	opts, user := newExecutorTransactor(t)
	implementation := common.HexToAddress("0x2000000000000000000000000000000000000000")

	client.EXPECT().PendingCodeAt(ctx, user).Return(nil, nil)

	executor := simple7702.NewExecutor(client, opts, implementation)
	receipt, err := executor.ExecuteCalls(ctx, []contract.Call{{
		Target: common.HexToAddress("0x3000000000000000000000000000000000000000"),
	}})

	if receipt != nil {
		t.Fatal("unexpected receipt after pending delegation was cleared")
	}
	if err == nil {
		t.Fatal("expected pending delegation error")
	}
	if !errors.Is(err, eip7702.ErrUnexpectedDelegation) {
		t.Fatalf("expected unexpected delegation error, got %v", err)
	}
}

func TestExecutorRejectsEmptyBatch(t *testing.T) {
	opts, _ := newExecutorTransactor(t)
	executor := simple7702.NewExecutor(nil, opts, common.HexToAddress("0x2000000000000000000000000000000000000000"))

	receipt, err := executor.ExecuteCalls(context.Background(), nil)
	if receipt != nil {
		t.Fatal("unexpected receipt for empty batch")
	}
	if !errors.Is(err, simple7702.ErrEmptyBatch) {
		t.Fatalf("expected empty batch error, got %v", err)
	}
}

func TestNilExecutorErrorPrecedesEmptyBatch(t *testing.T) {
	var executor *simple7702.Executor

	receipt, err := executor.ExecuteCalls(context.Background(), nil)
	if receipt != nil {
		t.Fatal("unexpected receipt from nil executor")
	}
	if err == nil || !strings.Contains(err.Error(), "executor is nil") {
		t.Fatalf("expected nil executor error, got %v", err)
	}
	if errors.Is(err, simple7702.ErrEmptyBatch) {
		t.Fatalf("nil executor error must not be hidden by empty batch: %v", err)
	}
}

func newExecutorTransactor(t *testing.T) (*bind.TransactOpts, common.Address) {
	t.Helper()
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	return transactorForKey(t, key)
}

func transactorForKey(t *testing.T, key *ecdsa.PrivateKey) (*bind.TransactOpts, common.Address) {
	t.Helper()
	opts, err := bind.NewKeyedTransactorWithChainID(key, big.NewInt(8453))
	if err != nil {
		t.Fatalf("create transactor: %v", err)
	}
	return opts, crypto.PubkeyToAddress(key.PublicKey)
}
