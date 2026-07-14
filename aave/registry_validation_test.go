package aave

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/tn606024/defi-simplify/config"
)

func TestBlockPinnedSnapshotLoaderValidation(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*fakeRegistrySource, Market)
		wantErr error
	}{
		{name: "valid snapshot"},
		{
			name: "Pool provider mismatch",
			mutate: func(source *fakeRegistrySource, _ Market) {
				source.poolProvider = common.HexToAddress("0x9000000000000000000000000000000000000001")
			},
			wantErr: ErrMarketRelationshipMismatch,
		},
		{
			name: "DataProvider provider mismatch",
			mutate: func(source *fakeRegistrySource, _ Market) {
				source.dataProvider = common.HexToAddress("0x9000000000000000000000000000000000000002")
			},
			wantErr: ErrMarketRelationshipMismatch,
		},
		{
			name: "missing contract code",
			mutate: func(source *fakeRegistrySource, market Market) {
				source.code[market.Pool()] = nil
			},
			wantErr: ErrMissingContractCode,
		},
		{
			name: "empty token symbol",
			mutate: func(source *fakeRegistrySource, _ Market) {
				metadata := source.metadata[source.listed[0].Address]
				metadata.Symbol = " "
				source.metadata[source.listed[0].Address] = metadata
			},
			wantErr: ErrMalformedTokenMetadata,
		},
		{
			name: "zero required reserve token",
			mutate: func(source *fakeRegistrySource, _ Market) {
				addresses := source.reserveTokens[source.listed[0].Address]
				addresses.AToken = common.Address{}
				source.reserveTokens[source.listed[0].Address] = addresses
			},
			wantErr: ErrRegistryDiscovery,
		},
		{
			name: "duplicate role address across reserves",
			mutate: func(source *fakeRegistrySource, _ Market) {
				secondUnderlying := common.HexToAddress("0x3000000000000000000000000000000000000011")
				firstTokens := source.reserveTokens[source.listed[0].Address]
				source.listed = append(source.listed, listedReserve{Symbol: "SECOND", Address: secondUnderlying})
				source.reserveTokens[secondUnderlying] = reserveTokenAddresses{
					AToken:            firstTokens.AToken,
					VariableDebtToken: common.HexToAddress("0x3000000000000000000000000000000000000013"),
				}
				source.addToken(secondUnderlying, "Second", "SECOND", 18)
				source.addToken(
					common.HexToAddress("0x3000000000000000000000000000000000000013"),
					"Variable Debt Second",
					"variableDebtSECOND",
					18,
				)
			},
			wantErr: ErrRegistryDiscovery,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			market := registryTestMarket(t)
			source := newFakeRegistrySource(market)
			if test.mutate != nil {
				test.mutate(source, market)
			}
			loader := &blockPinnedSnapshotLoader{source: source}
			snapshot, err := loader.Load(context.Background(), market)

			if test.wantErr != nil {
				if !errors.Is(err, test.wantErr) {
					t.Fatalf("Load() error = %v, want %v", err, test.wantErr)
				}
				if !errors.Is(err, ErrRegistryDiscovery) {
					t.Fatalf("Load() error = %v, want ErrRegistryDiscovery", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}
			if snapshot.BlockNumber() != source.header.Number.Uint64() {
				t.Fatalf("snapshot block = %d, want %d", snapshot.BlockNumber(), source.header.Number.Uint64())
			}
			if snapshot.BlockHash() != source.header.Hash() {
				t.Fatalf("snapshot hash = %s, want %s", snapshot.BlockHash(), source.header.Hash())
			}
			if snapshot.Len() != 1 {
				t.Fatalf("snapshot reserve count = %d, want 1", snapshot.Len())
			}
			for i, block := range source.readBlocks {
				if block == nil || block.Cmp(source.header.Number) != 0 {
					t.Fatalf("read block %d = %v, want %s", i, block, source.header.Number)
				}
			}
		})
	}
}

type fakeRegistrySource struct {
	header          *types.Header
	poolProvider    common.Address
	dataProvider    common.Address
	listed          []listedReserve
	reserveTokens   map[common.Address]reserveTokenAddresses
	metadata        map[common.Address]tokenMetadata
	code            map[common.Address][]byte
	readBlocks      []*big.Int
	headerRequested *big.Int
}

func newFakeRegistrySource(market Market) *fakeRegistrySource {
	underlying := common.HexToAddress("0x3000000000000000000000000000000000000001")
	aToken := common.HexToAddress("0x3000000000000000000000000000000000000002")
	variableDebt := common.HexToAddress("0x3000000000000000000000000000000000000003")
	source := &fakeRegistrySource{
		header:       &types.Header{Number: big.NewInt(12345)},
		poolProvider: market.AddressesProvider(),
		dataProvider: market.AddressesProvider(),
		listed:       []listedReserve{{Symbol: "USDC", Address: underlying}},
		reserveTokens: map[common.Address]reserveTokenAddresses{
			underlying: {AToken: aToken, VariableDebtToken: variableDebt},
		},
		metadata: make(map[common.Address]tokenMetadata),
		code:     make(map[common.Address][]byte),
	}
	for _, address := range []common.Address{
		market.Pool(),
		market.AddressesProvider(),
		market.ProtocolDataProvider(),
	} {
		source.code[address] = []byte{0x01}
	}
	if gateway, ok := market.WrappedTokenGateway(); ok {
		source.code[gateway] = []byte{0x01}
	}
	source.addToken(underlying, "USD Coin", "USDC", 6)
	source.addToken(aToken, "Aave Base USDC", "aBasUSDC", 6)
	source.addToken(variableDebt, "Aave Base Variable Debt USDC", "variableDebtBasUSDC", 6)
	return source
}

func (s *fakeRegistrySource) addToken(address common.Address, name string, symbol string, decimals uint8) {
	s.metadata[address] = tokenMetadata{Name: name, Symbol: symbol, Decimals: decimals}
	s.code[address] = []byte{0x01}
}

func (s *fakeRegistrySource) HeaderByNumber(_ context.Context, number *big.Int) (*types.Header, error) {
	s.headerRequested = cloneRegistryBlockNumber(number)
	return s.header, nil
}

func (s *fakeRegistrySource) CodeAt(_ context.Context, address common.Address, block *big.Int) ([]byte, error) {
	s.recordBlock(block)
	return s.code[address], nil
}

func (s *fakeRegistrySource) PoolAddressesProvider(_ context.Context, _ common.Address, block *big.Int) (common.Address, error) {
	s.recordBlock(block)
	return s.poolProvider, nil
}

func (s *fakeRegistrySource) DataProviderAddressesProvider(_ context.Context, _ common.Address, block *big.Int) (common.Address, error) {
	s.recordBlock(block)
	return s.dataProvider, nil
}

func (s *fakeRegistrySource) AllReserves(_ context.Context, _ common.Address, block *big.Int) ([]listedReserve, error) {
	s.recordBlock(block)
	return append([]listedReserve(nil), s.listed...), nil
}

func (s *fakeRegistrySource) ReserveTokenAddresses(
	_ context.Context,
	_ common.Address,
	asset common.Address,
	block *big.Int,
) (reserveTokenAddresses, error) {
	s.recordBlock(block)
	return s.reserveTokens[asset], nil
}

func (s *fakeRegistrySource) TokenMetadata(_ context.Context, address common.Address, block *big.Int) (tokenMetadata, error) {
	s.recordBlock(block)
	return s.metadata[address], nil
}

func (s *fakeRegistrySource) recordBlock(block *big.Int) {
	s.readBlocks = append(s.readBlocks, cloneRegistryBlockNumber(block))
}

func registryTestMarket(t *testing.T) Market {
	t.Helper()
	market, err := NewMarket(
		"aave-v3-base",
		config.Base,
		common.HexToAddress("0x1000000000000000000000000000000000000001"),
		common.HexToAddress("0x1000000000000000000000000000000000000002"),
		common.HexToAddress("0x1000000000000000000000000000000000000003"),
		common.HexToAddress("0x1000000000000000000000000000000000000004"),
	)
	if err != nil {
		t.Fatalf("NewMarket() error = %v", err)
	}
	return market
}
