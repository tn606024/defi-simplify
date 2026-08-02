package erc20

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
	binderc20 "github.com/tn606024/defi-simplify/bind/erc20"
	"github.com/tn606024/defi-simplify/token"
)

var _ = Describe("ERC20 account state", func() {
	var (
		ctx     context.Context
		asset   token.Token
		account common.Address
		spender common.Address
		backend *erc20StateBackend
	)

	BeforeEach(func() {
		ctx = context.Background()
		asset = resolvedTestToken(
			"0x0000000000000000000000000000000000000101",
			"USDC",
			"USD Coin",
			6,
		)
		account = common.HexToAddress("0x00000000000000000000000000000000000000aa")
		spender = common.HexToAddress("0x00000000000000000000000000000000000000bb")
		contractABI, err := binderc20.Erc20MetaData.GetAbi()
		Expect(err).NotTo(HaveOccurred())
		backend = &erc20StateBackend{
			header:    &types.Header{Number: big.NewInt(123), Extra: []byte("erc20-state")},
			contract:  contractABI,
			balance:   big.NewInt(1_500_000),
			allowance: big.NewInt(750_000),
		}
	})

	It("reads balance and allowance at one block hash", func() {
		state, err := LoadAccountState(ctx, backend, asset, account, spender)

		Expect(err).NotTo(HaveOccurred())
		Expect(state.Asset()).To(Equal(asset))
		Expect(state.Account()).To(Equal(account))
		Expect(state.Spender()).To(Equal(spender))
		Expect(state.BlockNumber()).To(Equal(uint64(123)))
		Expect(state.BlockHash()).To(Equal(backend.header.Hash()))
		Expect(state.Balance()).To(Equal(big.NewInt(1_500_000)))
		Expect(state.Allowance()).To(Equal(big.NewInt(750_000)))
		Expect(backend.calls).To(HaveLen(2))
		for _, call := range backend.calls {
			Expect(call.target).To(Equal(asset.Address()))
			Expect(call.blockHash).To(Equal(backend.header.Hash()))
		}
		Expect(backend.calls[0].method).To(Equal("balanceOf"))
		Expect(backend.calls[0].arguments).To(Equal([]interface{}{account}))
		Expect(backend.calls[1].method).To(Equal("allowance"))
		Expect(backend.calls[1].arguments).To(Equal([]interface{}{account, spender}))
	})

	It("returns defensive amount copies", func() {
		state, err := LoadAccountState(ctx, backend, asset, account, spender)
		Expect(err).NotTo(HaveOccurred())

		state.Balance().SetInt64(0)
		state.Allowance().SetInt64(0)

		Expect(state.Balance()).To(Equal(big.NewInt(1_500_000)))
		Expect(state.Allowance()).To(Equal(big.NewInt(750_000)))
	})

	DescribeTable("rejects invalid requests before reading the header",
		func(invalidField, message string) {
			var candidateBackend StateBackend = backend
			candidateAsset := asset
			candidateAccount := account
			candidateSpender := spender
			switch invalidField {
			case "backend":
				candidateBackend = nil
			case "asset":
				candidateAsset = token.Token{}
			case "account":
				candidateAccount = common.Address{}
			case "spender":
				candidateSpender = common.Address{}
			}
			state, err := LoadAccountState(
				ctx,
				candidateBackend,
				candidateAsset,
				candidateAccount,
				candidateSpender,
			)

			Expect(state).To(BeNil())
			Expect(errors.Is(err, ErrAccountStateRead)).To(BeTrue())
			Expect(err).To(MatchError(ContainSubstring(message)))
		},
		Entry("nil backend", "backend", "backend is nil"),
		Entry("invalid token", "asset", "invalid token"),
		Entry("zero account", "account", "account is zero"),
		Entry("zero spender", "spender", "spender is zero"),
	)

	It("preserves header RPC failures", func() {
		readErr := errors.New("header RPC failed")
		backend.err = readErr

		state, err := LoadAccountState(ctx, backend, asset, account, spender)

		Expect(state).To(BeNil())
		Expect(errors.Is(err, ErrAccountStateRead)).To(BeTrue())
		Expect(errors.Is(err, readErr)).To(BeTrue())
		Expect(err).To(MatchError(ContainSubstring("resolve block")))
	})

	It("preserves contract call failures", func() {
		readErr := errors.New("eth_call failed")
		backend.callErr = readErr

		state, err := LoadAccountState(ctx, backend, asset, account, spender)

		Expect(state).To(BeNil())
		Expect(errors.Is(err, ErrAccountStateRead)).To(BeTrue())
		Expect(errors.Is(err, readErr)).To(BeTrue())
		Expect(err).To(MatchError(ContainSubstring("read balance")))
	})

	It("rejects a missing block number", func() {
		backend.header = &types.Header{}
		backend.header.Number = nil

		state, err := LoadAccountState(ctx, backend, asset, account, spender)

		Expect(state).To(BeNil())
		Expect(errors.Is(err, ErrAccountStateRead)).To(BeTrue())
		Expect(err).To(MatchError(ContainSubstring("header has no uint64 block number")))
	})
})

type erc20StateCall struct {
	target    common.Address
	blockHash common.Hash
	method    string
	arguments []interface{}
}

type erc20StateBackend struct {
	header    *types.Header
	contract  *abi.ABI
	balance   *big.Int
	allowance *big.Int
	err       error
	callErr   error
	calls     []erc20StateCall
}

func (b *erc20StateBackend) HeaderByNumber(context.Context, *big.Int) (*types.Header, error) {
	if b.err != nil {
		return nil, b.err
	}
	return b.header, nil
}

func (b *erc20StateBackend) CodeAt(context.Context, common.Address, *big.Int) ([]byte, error) {
	return []byte{1}, nil
}

func (b *erc20StateBackend) CallContract(context.Context, ethereum.CallMsg, *big.Int) ([]byte, error) {
	return nil, errors.New("unexpected number-pinned contract call")
}

func (b *erc20StateBackend) CodeAtHash(context.Context, common.Address, common.Hash) ([]byte, error) {
	return []byte{1}, nil
}

func (b *erc20StateBackend) CallContractAtHash(
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
	b.calls = append(b.calls, erc20StateCall{
		target:    *call.To,
		blockHash: blockHash,
		method:    method.Name,
		arguments: arguments,
	})
	switch method.Name {
	case "balanceOf":
		return method.Outputs.Pack(b.balance)
	case "allowance":
		return method.Outputs.Pack(b.allowance)
	default:
		return nil, errors.New("unexpected ERC20 method")
	}
}
