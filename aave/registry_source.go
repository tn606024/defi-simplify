package aave

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	bindaave "github.com/tn606024/defi-simplify/bind/aave"
	binderc20 "github.com/tn606024/defi-simplify/bind/erc20"
)

// RegistryBackend is the minimal read-only Ethereum backend required for
// block-pinned Aave market discovery.
type RegistryBackend interface {
	bind.ContractCaller
	bind.BlockHashContractCaller
	HeaderByNumber(context.Context, *big.Int) (*types.Header, error)
}

type registryBlock struct {
	Number *big.Int
	Hash   common.Hash
}

type listedReserve struct {
	Symbol  string
	Address common.Address
}

type reserveTokenAddresses struct {
	AToken            common.Address
	StableDebtToken   common.Address
	VariableDebtToken common.Address
}

type tokenMetadata struct {
	Name     string
	Symbol   string
	Decimals uint8
}

type registrySource interface {
	HeaderByNumber(context.Context, *big.Int) (*types.Header, error)
	CodeAt(context.Context, common.Address, registryBlock) ([]byte, error)
	PoolAddressesProvider(context.Context, common.Address, registryBlock) (common.Address, error)
	DataProviderAddressesProvider(context.Context, common.Address, registryBlock) (common.Address, error)
	AllReserves(context.Context, common.Address, registryBlock) ([]listedReserve, error)
	ReserveTokenAddresses(context.Context, common.Address, common.Address, registryBlock) (reserveTokenAddresses, error)
	TokenMetadata(context.Context, common.Address, registryBlock) (tokenMetadata, error)
}

type chainRegistrySource struct {
	backend RegistryBackend
}

func (s *chainRegistrySource) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	return s.backend.HeaderByNumber(ctx, number)
}

func (s *chainRegistrySource) CodeAt(
	ctx context.Context,
	address common.Address,
	block registryBlock,
) ([]byte, error) {
	return s.backend.CodeAtHash(ctx, address, block.Hash)
}

func (s *chainRegistrySource) PoolAddressesProvider(
	ctx context.Context,
	poolAddress common.Address,
	block registryBlock,
) (common.Address, error) {
	pool, err := bindaave.NewPoolCaller(poolAddress, s.backend)
	if err != nil {
		return common.Address{}, err
	}
	return pool.ADDRESSESPROVIDER(registryCallOpts(ctx, block))
}

func (s *chainRegistrySource) DataProviderAddressesProvider(
	ctx context.Context,
	dataProviderAddress common.Address,
	block registryBlock,
) (common.Address, error) {
	dataProvider, err := bindaave.NewAaveProtocolDataProviderCaller(dataProviderAddress, s.backend)
	if err != nil {
		return common.Address{}, err
	}
	return dataProvider.ADDRESSESPROVIDER(registryCallOpts(ctx, block))
}

func (s *chainRegistrySource) AllReserves(
	ctx context.Context,
	dataProviderAddress common.Address,
	block registryBlock,
) ([]listedReserve, error) {
	dataProvider, err := bindaave.NewAaveProtocolDataProviderCaller(dataProviderAddress, s.backend)
	if err != nil {
		return nil, err
	}
	listed, err := dataProvider.GetAllReservesTokens(registryCallOpts(ctx, block))
	if err != nil {
		return nil, err
	}
	reserves := make([]listedReserve, len(listed))
	for i, reserve := range listed {
		reserves[i] = listedReserve{Symbol: reserve.Symbol, Address: reserve.TokenAddress}
	}
	return reserves, nil
}

func (s *chainRegistrySource) ReserveTokenAddresses(
	ctx context.Context,
	dataProviderAddress common.Address,
	asset common.Address,
	block registryBlock,
) (reserveTokenAddresses, error) {
	dataProvider, err := bindaave.NewAaveProtocolDataProviderCaller(dataProviderAddress, s.backend)
	if err != nil {
		return reserveTokenAddresses{}, err
	}
	addresses, err := dataProvider.GetReserveTokensAddresses(
		registryCallOpts(ctx, block),
		asset,
	)
	if err != nil {
		return reserveTokenAddresses{}, err
	}
	return reserveTokenAddresses{
		AToken:            addresses.ATokenAddress,
		StableDebtToken:   addresses.StableDebtTokenAddress,
		VariableDebtToken: addresses.VariableDebtTokenAddress,
	}, nil
}

func (s *chainRegistrySource) TokenMetadata(
	ctx context.Context,
	address common.Address,
	block registryBlock,
) (tokenMetadata, error) {
	tokenContract, err := binderc20.NewErc20Caller(address, s.backend)
	if err != nil {
		return tokenMetadata{}, err
	}
	opts := registryCallOpts(ctx, block)
	name, err := tokenContract.Name(opts)
	if err != nil {
		return tokenMetadata{}, err
	}
	symbol, err := tokenContract.Symbol(opts)
	if err != nil {
		return tokenMetadata{}, err
	}
	decimals, err := tokenContract.Decimals(opts)
	if err != nil {
		return tokenMetadata{}, err
	}
	return tokenMetadata{Name: name, Symbol: symbol, Decimals: decimals}, nil
}

func registryCallOpts(ctx context.Context, block registryBlock) *bind.CallOpts {
	return &bind.CallOpts{
		Context:     ctx,
		BlockNumber: cloneRegistryBlockNumber(block.Number),
		BlockHash:   block.Hash,
	}
}

func cloneRegistryBlockNumber(value *big.Int) *big.Int {
	if value == nil {
		return nil
	}
	return new(big.Int).Set(value)
}

func cloneRegistryBlock(block registryBlock) registryBlock {
	return registryBlock{
		Number: cloneRegistryBlockNumber(block.Number),
		Hash:   block.Hash,
	}
}
