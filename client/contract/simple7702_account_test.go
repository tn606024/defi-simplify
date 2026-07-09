package contract

import (
	"context"
	"math/big"
	"reflect"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Simple7702Account", func() {
	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()
	})

	It("encodes execute calldata for a single delegated account call", func() {
		account := common.HexToAddress("0x1000000000000000000000000000000000000000")
		target := common.HexToAddress("0x2000000000000000000000000000000000000000")
		value := big.NewInt(123)
		callData := []byte{0xa9, 0x05, 0x9c, 0xbb}

		action := BuildSimple7702AccountExecuteAction(account, target, value, callData)
		address, data, err := action.ToData(ctx, nil, nil)

		Expect(err).NotTo(HaveOccurred())
		Expect(address).To(Equal(account))

		methodName, values := unpackSimple7702AccountCalldata(data)
		Expect(methodName).To(Equal("execute"))
		Expect(values).To(HaveLen(3))
		Expect(values[0]).To(Equal(target))
		Expect(values[1].(*big.Int)).To(Equal(value))
		Expect(values[2]).To(Equal(callData))
	})

	It("encodes executeBatch calldata from neutral calls", func() {
		account := common.HexToAddress("0x1000000000000000000000000000000000000000")
		firstTarget := common.HexToAddress("0x2000000000000000000000000000000000000000")
		secondTarget := common.HexToAddress("0x3000000000000000000000000000000000000000")
		calls := []Call{
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

		action := BuildSimple7702AccountExecuteBatchAction(account, calls)
		address, data, err := action.ToData(ctx, nil, nil)

		Expect(err).NotTo(HaveOccurred())
		Expect(address).To(Equal(account))

		methodName, values := unpackSimple7702AccountCalldata(data)
		Expect(methodName).To(Equal("executeBatch"))
		Expect(values).To(HaveLen(1))

		decodedCalls := reflect.ValueOf(values[0])
		Expect(decodedCalls.Kind()).To(Equal(reflect.Slice))
		Expect(decodedCalls.Len()).To(Equal(2))

		assertDecodedSimple7702AccountCall(decodedCalls.Index(0), firstTarget, big.NewInt(0), []byte{0x01, 0x02})
		assertDecodedSimple7702AccountCall(decodedCalls.Index(1), secondTarget, big.NewInt(456), []byte{0x03, 0x04})
	})
})

func unpackSimple7702AccountCalldata(data []byte) (string, []interface{}) {
	GinkgoHelper()

	parsed, err := abi.JSON(strings.NewReader(simple7702AccountABI))
	Expect(err).NotTo(HaveOccurred())

	method, err := parsed.MethodById(data[:4])
	Expect(err).NotTo(HaveOccurred())

	values, err := method.Inputs.Unpack(data[4:])
	Expect(err).NotTo(HaveOccurred())

	return method.Name, values
}

func assertDecodedSimple7702AccountCall(decodedCall reflect.Value, target common.Address, value *big.Int, data []byte) {
	GinkgoHelper()

	Expect(decodedCall.FieldByName("Target").Interface()).To(Equal(target))
	Expect(decodedCall.FieldByName("Value").Interface().(*big.Int).Cmp(value)).To(Equal(0))
	Expect(decodedCall.FieldByName("Data").Interface()).To(Equal(data))
}
