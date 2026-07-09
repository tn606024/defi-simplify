package simple7702_test

import (
	"context"
	"math/big"
	"reflect"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/tn606024/defi-simplify/client/account/simple7702"
	"github.com/tn606024/defi-simplify/client/contract"
)

func TestBuildExecuteActionEncodesSingleDelegatedAccountCall(t *testing.T) {
	account := common.HexToAddress("0x1000000000000000000000000000000000000000")
	target := common.HexToAddress("0x2000000000000000000000000000000000000000")
	value := big.NewInt(123)
	callData := []byte{0xa9, 0x05, 0x9c, 0xbb}

	action := simple7702.BuildExecuteAction(account, target, value, callData)
	address, data, err := action.ToData(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("ToData returned error: %v", err)
	}
	if address != account {
		t.Fatalf("unexpected account target: got %s want %s", address.Hex(), account.Hex())
	}

	methodName, values := unpackCalldata(t, data)
	if methodName != "execute" {
		t.Fatalf("unexpected method: got %s want execute", methodName)
	}
	if len(values) != 3 {
		t.Fatalf("unexpected value count: got %d want 3", len(values))
	}
	if values[0] != target {
		t.Fatalf("unexpected call target: got %v want %v", values[0], target)
	}
	if values[1].(*big.Int).Cmp(value) != 0 {
		t.Fatalf("unexpected call value: got %s want %s", values[1].(*big.Int), value)
	}
	if !reflect.DeepEqual(values[2], callData) {
		t.Fatalf("unexpected calldata: got %x want %x", values[2], callData)
	}
}

func TestBuildExecuteBatchActionEncodesNeutralCalls(t *testing.T) {
	account := common.HexToAddress("0x1000000000000000000000000000000000000000")
	firstTarget := common.HexToAddress("0x2000000000000000000000000000000000000000")
	secondTarget := common.HexToAddress("0x3000000000000000000000000000000000000000")
	calls := []contract.Call{
		{
			Target: firstTarget,
			Value:  nil,
			Data:   []byte{0x01, 0x02},
		},
		{
			Target: secondTarget,
			Value:  big.NewInt(456),
			Data:   []byte{0x03, 0x04},
		},
	}

	action := simple7702.BuildExecuteBatchAction(account, calls)
	address, data, err := action.ToData(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("ToData returned error: %v", err)
	}
	if address != account {
		t.Fatalf("unexpected account target: got %s want %s", address.Hex(), account.Hex())
	}

	methodName, values := unpackCalldata(t, data)
	if methodName != "executeBatch" {
		t.Fatalf("unexpected method: got %s want executeBatch", methodName)
	}
	if len(values) != 1 {
		t.Fatalf("unexpected value count: got %d want 1", len(values))
	}

	decodedCalls := reflect.ValueOf(values[0])
	if decodedCalls.Kind() != reflect.Slice {
		t.Fatalf("unexpected decoded calls kind: got %s want slice", decodedCalls.Kind())
	}
	if decodedCalls.Len() != 2 {
		t.Fatalf("unexpected decoded call count: got %d want 2", decodedCalls.Len())
	}

	assertDecodedCall(t, decodedCalls.Index(0), firstTarget, big.NewInt(0), []byte{0x01, 0x02})
	assertDecodedCall(t, decodedCalls.Index(1), secondTarget, big.NewInt(456), []byte{0x03, 0x04})
}

func unpackCalldata(t *testing.T, data []byte) (string, []interface{}) {
	t.Helper()

	parsed, err := abi.JSON(strings.NewReader(simple7702.ABIJSON))
	if err != nil {
		t.Fatalf("parse ABI: %v", err)
	}

	method, err := parsed.MethodById(data[:4])
	if err != nil {
		t.Fatalf("method by id: %v", err)
	}

	values, err := method.Inputs.Unpack(data[4:])
	if err != nil {
		t.Fatalf("unpack inputs: %v", err)
	}

	return method.Name, values
}

func assertDecodedCall(t *testing.T, decodedCall reflect.Value, target common.Address, value *big.Int, data []byte) {
	t.Helper()

	if got := decodedCall.FieldByName("Target").Interface(); got != target {
		t.Fatalf("unexpected target: got %v want %v", got, target)
	}
	if got := decodedCall.FieldByName("Value").Interface().(*big.Int); got.Cmp(value) != 0 {
		t.Fatalf("unexpected value: got %s want %s", got, value)
	}
	if got := decodedCall.FieldByName("Data").Interface(); !reflect.DeepEqual(got, data) {
		t.Fatalf("unexpected data: got %x want %x", got, data)
	}
}
