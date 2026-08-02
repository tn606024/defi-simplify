package aave

import (
	"context"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	bindaave "github.com/tn606024/defi-simplify/bind/aave"
)

var _ = Describe("Aave user reserve state", func() {
	var (
		ctx     context.Context
		reserve Reserve
		account common.Address
		backend *aaveUserReserveStateBackend
	)

	BeforeEach(func() {
		ctx = context.Background()
		_, reserve, _ = stepTestReserves()
		account = common.HexToAddress("0x00000000000000000000000000000000000000aa")
		contractABI, err := bindaave.AaveProtocolDataProviderMetaData.GetAbi()
		Expect(err).NotTo(HaveOccurred())
		backend = &aaveUserReserveStateBackend{
			header:   &types.Header{Number: big.NewInt(456), Extra: []byte("aave-state")},
			contract: contractABI,
			values: []interface{}{
				big.NewInt(10_000_000),
				big.NewInt(1),
				big.NewInt(2_000_000),
				big.NewInt(3),
				big.NewInt(4),
				big.NewInt(5),
				big.NewInt(6),
				big.NewInt(7),
				true,
			},
		}
	})

	It("reads the resolved reserve position at one block hash", func() {
		state, err := LoadUserReserveState(ctx, backend, reserve, account)

		Expect(err).NotTo(HaveOccurred())
		Expect(state.Reserve()).To(Equal(reserve))
		Expect(state.Account()).To(Equal(account))
		Expect(state.BlockNumber()).To(Equal(uint64(456)))
		Expect(state.BlockHash()).To(Equal(backend.header.Hash()))
		Expect(state.CurrentATokenBalance()).To(Equal(big.NewInt(10_000_000)))
		Expect(state.CurrentStableDebt()).To(Equal(big.NewInt(1)))
		Expect(state.CurrentVariableDebt()).To(Equal(big.NewInt(2_000_000)))
		Expect(state.PrincipalStableDebt()).To(Equal(big.NewInt(3)))
		Expect(state.ScaledVariableDebt()).To(Equal(big.NewInt(4)))
		Expect(state.StableBorrowRate()).To(Equal(big.NewInt(5)))
		Expect(state.LiquidityRate()).To(Equal(big.NewInt(6)))
		Expect(state.StableRateLastUpdated()).To(Equal(big.NewInt(7)))
		Expect(state.UsageAsCollateralEnabled()).To(BeTrue())
		Expect(backend.target).To(Equal(reserve.Market().ProtocolDataProvider()))
		Expect(backend.blockHash).To(Equal(backend.header.Hash()))
		Expect(backend.method).To(Equal("getUserReserveData"))
		Expect(backend.arguments).To(Equal([]interface{}{reserve.Underlying().Address(), account}))
	})

	It("returns defensive amount copies", func() {
		state, err := LoadUserReserveState(ctx, backend, reserve, account)
		Expect(err).NotTo(HaveOccurred())

		state.CurrentATokenBalance().SetInt64(0)
		state.CurrentVariableDebt().SetInt64(0)
		state.LiquidityRate().SetInt64(0)

		Expect(state.CurrentATokenBalance()).To(Equal(big.NewInt(10_000_000)))
		Expect(state.CurrentVariableDebt()).To(Equal(big.NewInt(2_000_000)))
		Expect(state.LiquidityRate()).To(Equal(big.NewInt(6)))
	})

	DescribeTable("rejects invalid requests before reading the header",
		func(invalidField, message string) {
			var candidateBackend RegistryBackend = backend
			candidateReserve := reserve
			candidateAccount := account
			switch invalidField {
			case "backend":
				candidateBackend = nil
			case "reserve":
				candidateReserve = Reserve{}
			case "account":
				candidateAccount = common.Address{}
			}
			state, err := LoadUserReserveState(ctx, candidateBackend, candidateReserve, candidateAccount)

			Expect(state).To(BeNil())
			Expect(errors.Is(err, ErrUserReserveStateRead)).To(BeTrue())
			Expect(err).To(MatchError(ContainSubstring(message)))
		},
		Entry("nil backend", "backend", "backend is nil"),
		Entry("invalid reserve", "reserve", "invalid Aave reserve"),
		Entry("zero account", "account", "account is zero"),
	)

	It("preserves header RPC failures", func() {
		readErr := errors.New("header RPC failed")
		backend.err = readErr

		state, err := LoadUserReserveState(ctx, backend, reserve, account)

		Expect(state).To(BeNil())
		Expect(errors.Is(err, ErrUserReserveStateRead)).To(BeTrue())
		Expect(errors.Is(err, readErr)).To(BeTrue())
		Expect(err).To(MatchError(ContainSubstring("resolve block")))
	})

	It("preserves protocol data provider failures", func() {
		readErr := errors.New("eth_call failed")
		backend.callErr = readErr

		state, err := LoadUserReserveState(ctx, backend, reserve, account)

		Expect(state).To(BeNil())
		Expect(errors.Is(err, ErrUserReserveStateRead)).To(BeTrue())
		Expect(errors.Is(err, readErr)).To(BeTrue())
		Expect(err).To(MatchError(ContainSubstring("read reserve data")))
	})

	It("rejects a missing block number", func() {
		backend.header = &types.Header{}
		backend.header.Number = nil

		state, err := LoadUserReserveState(ctx, backend, reserve, account)

		Expect(state).To(BeNil())
		Expect(errors.Is(err, ErrUserReserveStateRead)).To(BeTrue())
		Expect(err).To(MatchError(ContainSubstring("header has no uint64 block number")))
	})
})

type aaveUserReserveStateBackend struct {
	header    *types.Header
	contract  *abi.ABI
	values    []interface{}
	err       error
	callErr   error
	target    common.Address
	blockHash common.Hash
	method    string
	arguments []interface{}
}

func (b *aaveUserReserveStateBackend) HeaderByNumber(context.Context, *big.Int) (*types.Header, error) {
	if b.err != nil {
		return nil, b.err
	}
	return b.header, nil
}

func (b *aaveUserReserveStateBackend) CodeAt(context.Context, common.Address, *big.Int) ([]byte, error) {
	return []byte{1}, nil
}

func (b *aaveUserReserveStateBackend) CallContract(context.Context, ethereum.CallMsg, *big.Int) ([]byte, error) {
	return nil, errors.New("unexpected number-pinned contract call")
}

func (b *aaveUserReserveStateBackend) CodeAtHash(context.Context, common.Address, common.Hash) ([]byte, error) {
	return []byte{1}, nil
}

func (b *aaveUserReserveStateBackend) CallContractAtHash(
	_ context.Context,
	call ethereum.CallMsg,
	blockHash common.Hash,
) ([]byte, error) {
	if b.callErr != nil {
		return nil, b.callErr
	}
	method, err := b.contract.MethodById(call.Data)
	if err != nil {
		return nil, err
	}
	arguments, err := method.Inputs.Unpack(call.Data[4:])
	if err != nil {
		return nil, err
	}
	b.target = *call.To
	b.blockHash = blockHash
	b.method = method.Name
	b.arguments = arguments
	return method.Outputs.Pack(b.values...)
}
