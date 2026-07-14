package aave

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/tn606024/defi-simplify/token"
)

var (
	// ErrInvalidRegistry is returned when a registry has no usable backend or
	// market definition.
	ErrInvalidRegistry = errors.New("invalid Aave registry")
	// ErrRegistryDiscovery is returned when one block-pinned market discovery
	// cannot produce a complete validated snapshot.
	ErrRegistryDiscovery = errors.New("Aave registry discovery failed")
	// ErrMarketRelationshipMismatch is returned when the Pool or DataProvider
	// does not belong to the market's trusted PoolAddressesProvider.
	ErrMarketRelationshipMismatch = errors.New("Aave market relationship mismatch")
	// ErrMissingContractCode is returned when a required market or reserve-token
	// address has no code at the snapshot block.
	ErrMissingContractCode = errors.New("Aave registry address has no contract code")
	// ErrMalformedTokenMetadata is returned when required ERC20 display metadata
	// cannot be resolved into a usable token value.
	ErrMalformedTokenMetadata = errors.New("malformed ERC20 metadata")
)

// Registry explicitly loads and caches immutable snapshots for one trusted
// Aave market. Load reuses the first successful snapshot; Refresh is the only
// operation that reads a newer block.
type Registry struct {
	market Market
	loader snapshotLoader

	mu       sync.Mutex
	snapshot *MarketSnapshot
}

type snapshotLoader interface {
	Load(context.Context, Market) (*MarketSnapshot, error)
}

// NewRegistry creates a market-scoped Aave registry. It performs no RPC reads;
// callers explicitly choose when to Load or Refresh the first snapshot.
func NewRegistry(backend RegistryBackend, market Market) (*Registry, error) {
	if backend == nil {
		return nil, fmt.Errorf("%w: backend is nil", ErrInvalidRegistry)
	}
	return newRegistry(market, &blockPinnedSnapshotLoader{
		source: &chainRegistrySource{backend: backend},
	})
}

func newRegistry(market Market, loader snapshotLoader) (*Registry, error) {
	if err := market.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidRegistry, err)
	}
	if loader == nil {
		return nil, fmt.Errorf("%w: snapshot loader is nil", ErrInvalidRegistry)
	}
	return &Registry{market: market, loader: loader}, nil
}

// Market returns the immutable market trust anchors used by this registry.
func (r *Registry) Market() Market {
	if r == nil {
		return Market{}
	}
	return r.market
}

// Load returns the cached snapshot, loading it from one pinned block on the
// first successful call. It never refreshes a populated cache implicitly.
func (r *Registry) Load(ctx context.Context) (*MarketSnapshot, error) {
	if r == nil {
		return nil, fmt.Errorf("%w: registry is nil", ErrInvalidRegistry)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.snapshot != nil {
		return r.snapshot, nil
	}
	return r.refreshLocked(ctx)
}

// Refresh explicitly loads a new block-pinned snapshot and replaces the cache
// only after discovery and validation succeed. A failed refresh leaves the
// previous snapshot available through Load.
func (r *Registry) Refresh(ctx context.Context) (*MarketSnapshot, error) {
	if r == nil {
		return nil, fmt.Errorf("%w: registry is nil", ErrInvalidRegistry)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.refreshLocked(ctx)
}

func (r *Registry) refreshLocked(ctx context.Context) (*MarketSnapshot, error) {
	snapshot, err := r.loader.Load(ctx, r.market)
	if err != nil {
		return nil, err
	}
	if snapshot == nil {
		return nil, fmt.Errorf("%w: loader returned a nil snapshot", ErrRegistryDiscovery)
	}
	if err := snapshot.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrRegistryDiscovery, err)
	}
	if !snapshot.Market().SameMarket(r.market) {
		return nil, fmt.Errorf(
			"%w: snapshot market %q does not match registry market %q",
			ErrRegistryDiscovery,
			snapshot.Market().ID(),
			r.market.ID(),
		)
	}
	r.snapshot = snapshot
	return snapshot, nil
}

type blockPinnedSnapshotLoader struct {
	source registrySource
}

func (l *blockPinnedSnapshotLoader) Load(ctx context.Context, market Market) (*MarketSnapshot, error) {
	if l == nil || l.source == nil {
		return nil, fmt.Errorf("%w: discovery source is nil", ErrRegistryDiscovery)
	}
	header, err := l.source.HeaderByNumber(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: resolve snapshot block: %w", ErrRegistryDiscovery, err)
	}
	if header == nil || header.Number == nil || !header.Number.IsUint64() {
		return nil, fmt.Errorf("%w: snapshot header has no uint64 block number", ErrRegistryDiscovery)
	}
	block := registryBlock{Number: new(big.Int).Set(header.Number), Hash: header.Hash()}
	if block.Hash == (common.Hash{}) {
		return nil, fmt.Errorf("%w: snapshot block hash is zero", ErrRegistryDiscovery)
	}

	if err := l.validateMarketContracts(ctx, market, block); err != nil {
		return nil, err
	}
	if err := l.validateProviderRelationships(ctx, market, block); err != nil {
		return nil, err
	}

	listed, err := l.source.AllReserves(ctx, market.ProtocolDataProvider(), block)
	if err != nil {
		return nil, fmt.Errorf("%w: list reserves: %w", ErrRegistryDiscovery, err)
	}
	if len(listed) == 0 {
		return nil, fmt.Errorf("%w: market returned no reserves", ErrRegistryDiscovery)
	}

	reserves := make([]Reserve, 0, len(listed))
	seenRoles := make(map[common.Address]string, len(listed)*3)
	for i, listedReserve := range listed {
		reserve, err := l.loadReserve(ctx, market, block, listedReserve, seenRoles)
		if err != nil {
			return nil, fmt.Errorf(
				"%w: reserve %d (%s): %w",
				ErrRegistryDiscovery,
				i+1,
				listedReserve.Address.Hex(),
				err,
			)
		}
		reserves = append(reserves, reserve)
	}

	snapshot, err := NewMarketSnapshot(market, block.Number.Uint64(), block.Hash, reserves)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrRegistryDiscovery, err)
	}
	return snapshot, nil
}

func (l *blockPinnedSnapshotLoader) validateMarketContracts(
	ctx context.Context,
	market Market,
	block registryBlock,
) error {
	addresses := []struct {
		name    string
		address common.Address
	}{
		{name: "Pool", address: market.Pool()},
		{name: "PoolAddressesProvider", address: market.AddressesProvider()},
		{name: "AaveProtocolDataProvider", address: market.ProtocolDataProvider()},
	}
	if gateway, ok := market.WrappedTokenGateway(); ok {
		addresses = append(addresses, struct {
			name    string
			address common.Address
		}{name: "WrappedTokenGateway", address: gateway})
	}
	for _, contract := range addresses {
		if err := l.requireCode(ctx, contract.name, contract.address, block); err != nil {
			return err
		}
	}
	return nil
}

func (l *blockPinnedSnapshotLoader) validateProviderRelationships(
	ctx context.Context,
	market Market,
	block registryBlock,
) error {
	poolProvider, err := l.source.PoolAddressesProvider(ctx, market.Pool(), block)
	if err != nil {
		return fmt.Errorf("%w: read Pool addresses provider: %w", ErrRegistryDiscovery, err)
	}
	if poolProvider != market.AddressesProvider() {
		return fmt.Errorf(
			"%w: %w: Pool returned %s, expected %s",
			ErrRegistryDiscovery,
			ErrMarketRelationshipMismatch,
			poolProvider.Hex(),
			market.AddressesProvider().Hex(),
		)
	}

	dataProvider, err := l.source.DataProviderAddressesProvider(
		ctx,
		market.ProtocolDataProvider(),
		block,
	)
	if err != nil {
		return fmt.Errorf("%w: read DataProvider addresses provider: %w", ErrRegistryDiscovery, err)
	}
	if dataProvider != market.AddressesProvider() {
		return fmt.Errorf(
			"%w: %w: DataProvider returned %s, expected %s",
			ErrRegistryDiscovery,
			ErrMarketRelationshipMismatch,
			dataProvider.Hex(),
			market.AddressesProvider().Hex(),
		)
	}
	return nil
}

func (l *blockPinnedSnapshotLoader) loadReserve(
	ctx context.Context,
	market Market,
	block registryBlock,
	listed listedReserve,
	seenRoles map[common.Address]string,
) (Reserve, error) {
	if listed.Address == (common.Address{}) {
		return Reserve{}, fmt.Errorf("underlying address is zero")
	}
	addresses, err := l.source.ReserveTokenAddresses(
		ctx,
		market.ProtocolDataProvider(),
		listed.Address,
		block,
	)
	if err != nil {
		return Reserve{}, fmt.Errorf("read reserve token addresses: %w", err)
	}

	roles := []struct {
		name     string
		address  common.Address
		optional bool
	}{
		{name: "underlying", address: listed.Address},
		{name: "aToken", address: addresses.AToken},
		{name: "variable debt token", address: addresses.VariableDebtToken},
		{name: "stable debt token", address: addresses.StableDebtToken, optional: true},
	}
	resolved := make(map[string]token.Token, len(roles))
	for _, role := range roles {
		if role.optional && role.address == (common.Address{}) {
			continue
		}
		if role.address == (common.Address{}) {
			return Reserve{}, fmt.Errorf("%s address is zero", role.name)
		}
		if previous, ok := seenRoles[role.address]; ok {
			return Reserve{}, fmt.Errorf(
				"%s and %s use the same address %s",
				previous,
				role.name,
				role.address.Hex(),
			)
		}
		if err := l.requireCode(ctx, role.name, role.address, block); err != nil {
			return Reserve{}, err
		}
		metadata, err := l.source.TokenMetadata(ctx, role.address, block)
		if err != nil {
			return Reserve{}, fmt.Errorf("read %s metadata: %w", role.name, err)
		}
		resolvedToken, err := resolvedToken(market, role.name, role.address, metadata)
		if err != nil {
			return Reserve{}, err
		}
		seenRoles[role.address] = role.name
		resolved[role.name] = resolvedToken
	}

	var stableDebt *token.Token
	if value, ok := resolved["stable debt token"]; ok {
		stableDebt = &value
	}
	return NewReserve(
		market,
		resolved["underlying"],
		resolved["aToken"],
		resolved["variable debt token"],
		stableDebt,
	)
}

func (l *blockPinnedSnapshotLoader) requireCode(
	ctx context.Context,
	name string,
	address common.Address,
	block registryBlock,
) error {
	code, err := l.source.CodeAt(ctx, address, block)
	if err != nil {
		return fmt.Errorf("%w: read %s code at %s: %w", ErrRegistryDiscovery, name, address.Hex(), err)
	}
	if len(code) == 0 {
		return fmt.Errorf(
			"%w: %w: %s at %s",
			ErrRegistryDiscovery,
			ErrMissingContractCode,
			name,
			address.Hex(),
		)
	}
	return nil
}

func resolvedToken(
	market Market,
	role string,
	address common.Address,
	metadata tokenMetadata,
) (token.Token, error) {
	if strings.TrimSpace(metadata.Symbol) == "" {
		return token.Token{}, fmt.Errorf("%w: %s symbol is empty", ErrMalformedTokenMetadata, role)
	}
	if strings.TrimSpace(metadata.Name) == "" {
		return token.Token{}, fmt.Errorf("%w: %s name is empty", ErrMalformedTokenMetadata, role)
	}
	ref, err := token.NewRef(market.Chain(), address)
	if err != nil {
		return token.Token{}, err
	}
	resolved, err := token.New(ref, metadata.Symbol, metadata.Name, metadata.Decimals)
	if err != nil {
		return token.Token{}, err
	}
	return resolved, nil
}
