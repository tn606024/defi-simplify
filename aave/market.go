package aave

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/tn606024/defi-simplify/config"
	"github.com/tn606024/defi-simplify/token"
)

var (
	// ErrInvalidMarket is returned when an Aave market definition is incomplete
	// or inconsistent.
	ErrInvalidMarket = errors.New("invalid Aave market")
	// ErrInvalidReserve is returned when an Aave reserve does not contain valid,
	// chain-consistent token roles.
	ErrInvalidReserve = errors.New("invalid Aave reserve")
	// ErrInvalidMarketSnapshot is returned when a snapshot does not represent one
	// immutable market state.
	ErrInvalidMarketSnapshot = errors.New("invalid Aave market snapshot")
	// ErrDuplicateReserve is returned when a snapshot contains the same
	// underlying token more than once.
	ErrDuplicateReserve = errors.New("duplicate Aave reserve")
	// ErrReserveNotFound is returned when an asset is not part of a snapshot.
	ErrReserveNotFound = errors.New("Aave reserve not found")
)

// Market identifies one resolved Aave deployment. Core contract addresses are
// immutable values in this model; future registries own refreshing them.
type Market struct {
	id                     string
	chain                  config.Chain
	pool                   common.Address
	addressesProvider      common.Address
	protocolDataProvider   common.Address
	wrappedTokenGateway    common.Address
	hasWrappedTokenGateway bool
}

// NewMarket creates a validated Aave market. wrappedTokenGateway may be zero
// when the deployment does not expose that optional periphery contract.
func NewMarket(
	id string,
	chain config.Chain,
	pool common.Address,
	addressesProvider common.Address,
	protocolDataProvider common.Address,
	wrappedTokenGateway common.Address,
) (Market, error) {
	market := Market{
		id:                     strings.TrimSpace(id),
		chain:                  chain,
		pool:                   pool,
		addressesProvider:      addressesProvider,
		protocolDataProvider:   protocolDataProvider,
		wrappedTokenGateway:    wrappedTokenGateway,
		hasWrappedTokenGateway: wrappedTokenGateway != (common.Address{}),
	}
	if err := validateMarket(market); err != nil {
		return Market{}, err
	}
	return market, nil
}

// Validate checks whether the market contains a supported chain and all
// required deployment addresses.
func (m Market) Validate() error {
	return validateMarket(m)
}

// ID returns the stable SDK market identifier.
func (m Market) ID() string {
	return m.id
}

// Chain returns the market chain.
func (m Market) Chain() config.Chain {
	return m.chain
}

// Pool returns the resolved Aave Pool address.
func (m Market) Pool() common.Address {
	return m.pool
}

// AddressesProvider returns the market's PoolAddressesProvider.
func (m Market) AddressesProvider() common.Address {
	return m.addressesProvider
}

// ProtocolDataProvider returns the market's AaveProtocolDataProvider.
func (m Market) ProtocolDataProvider() common.Address {
	return m.protocolDataProvider
}

// WrappedTokenGateway returns the optional wrapped-token gateway and whether
// the market definition contains one.
func (m Market) WrappedTokenGateway() (common.Address, bool) {
	return m.wrappedTokenGateway, m.hasWrappedTokenGateway
}

// SameMarket reports whether both values describe the same resolved market,
// including optional periphery addresses.
func (m Market) SameMarket(other Market) bool {
	return m == other
}

// Reserve groups one Aave underlying asset with its market-specific token
// roles. Token role metadata remains independent; callers must not assume
// reserve-token decimals equal underlying decimals.
type Reserve struct {
	market                  Market
	underlying              token.Token
	aToken                  token.Token
	variableDebtToken       token.Token
	stableDebtToken         token.Token
	hasStableDebtTokenValue bool
}

// NewReserve creates a validated reserve. stableDebtToken may be nil when the
// selected Aave market does not expose a usable stable debt token.
func NewReserve(
	market Market,
	underlying token.Token,
	aToken token.Token,
	variableDebtToken token.Token,
	stableDebtToken *token.Token,
) (Reserve, error) {
	reserve := Reserve{
		market:            market,
		underlying:        underlying,
		aToken:            aToken,
		variableDebtToken: variableDebtToken,
	}
	if stableDebtToken != nil {
		reserve.stableDebtToken = *stableDebtToken
		reserve.hasStableDebtTokenValue = true
	}
	if err := validateReserve(reserve); err != nil {
		return Reserve{}, err
	}
	return reserve, nil
}

// Validate checks the reserve's market, token identities, chains, and role
// address uniqueness.
func (r Reserve) Validate() error {
	return validateReserve(r)
}

// Market returns the reserve's resolved Aave market.
func (r Reserve) Market() Market {
	return r.market
}

// Underlying returns the ERC20 asset accepted by Aave Pool operations.
func (r Reserve) Underlying() token.Token {
	return r.underlying
}

// AToken returns the reserve's interest-bearing token.
func (r Reserve) AToken() token.Token {
	return r.aToken
}

// VariableDebtToken returns the reserve's variable-rate debt token.
func (r Reserve) VariableDebtToken() token.Token {
	return r.variableDebtToken
}

// StableDebtToken returns the optional stable-rate debt token and whether it is
// present in this market snapshot.
func (r Reserve) StableDebtToken() (token.Token, bool) {
	return r.stableDebtToken, r.hasStableDebtTokenValue
}

// MarketSnapshot is an immutable address-indexed view of one Aave market at a
// specific block. It owns its reserve index and never exposes the backing map.
type MarketSnapshot struct {
	market      Market
	blockNumber uint64
	blockHash   common.Hash
	reserves    map[common.Address]Reserve
}

// NewMarketSnapshot creates an immutable market snapshot. An empty reserve set
// is valid so the model can represent a newly initialized market; discovery
// policy may impose stronger requirements.
func NewMarketSnapshot(
	market Market,
	blockNumber uint64,
	blockHash common.Hash,
	reserves []Reserve,
) (*MarketSnapshot, error) {
	if err := market.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidMarketSnapshot, err)
	}
	if blockHash == (common.Hash{}) {
		return nil, fmt.Errorf("%w: block hash is zero", ErrInvalidMarketSnapshot)
	}

	reserveIndex, err := indexReserves(market, reserves)
	if err != nil {
		return nil, err
	}
	return &MarketSnapshot{
		market:      market,
		blockNumber: blockNumber,
		blockHash:   blockHash,
		reserves:    reserveIndex,
	}, nil
}

// Validate checks the snapshot's market, block identity, reserve membership,
// and reserve index.
func (s *MarketSnapshot) Validate() error {
	if s == nil {
		return fmt.Errorf("%w: snapshot is nil", ErrInvalidMarketSnapshot)
	}
	if err := s.market.Validate(); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidMarketSnapshot, err)
	}
	if s.blockHash == (common.Hash{}) {
		return fmt.Errorf("%w: block hash is zero", ErrInvalidMarketSnapshot)
	}
	reserves := make([]Reserve, 0, len(s.reserves))
	for address, reserve := range s.reserves {
		if reserve.Underlying().Address() != address {
			return fmt.Errorf(
				"%w: reserve index key %s does not match underlying %s",
				ErrInvalidMarketSnapshot,
				address.Hex(),
				reserve.Underlying().Address().Hex(),
			)
		}
		reserves = append(reserves, reserve)
	}
	_, err := indexReserves(s.market, reserves)
	return err
}

// Market returns the resolved market represented by the snapshot.
func (s *MarketSnapshot) Market() Market {
	if s == nil {
		return Market{}
	}
	return s.market
}

// BlockNumber returns the block number used for every snapshot read.
func (s *MarketSnapshot) BlockNumber() uint64 {
	if s == nil {
		return 0
	}
	return s.blockNumber
}

// BlockHash returns the block hash used for every snapshot read.
func (s *MarketSnapshot) BlockHash() common.Hash {
	if s == nil {
		return common.Hash{}
	}
	return s.blockHash
}

// Len returns the number of reserves in the snapshot.
func (s *MarketSnapshot) Len() int {
	if s == nil {
		return 0
	}
	return len(s.reserves)
}

// Reserves returns a new deterministic slice ordered by underlying address.
// Mutating the returned slice cannot affect the snapshot.
func (s *MarketSnapshot) Reserves() []Reserve {
	if s == nil {
		return nil
	}
	reserves := make([]Reserve, 0, len(s.reserves))
	for _, reserve := range s.reserves {
		reserves = append(reserves, reserve)
	}
	sort.Slice(reserves, func(i, j int) bool {
		return bytes.Compare(
			reserves[i].Underlying().Address().Bytes(),
			reserves[j].Underlying().Address().Bytes(),
		) < 0
	})
	return reserves
}

// Reserve resolves a reserve from a chain-scoped token reference.
func (s *MarketSnapshot) Reserve(ref token.Ref) (Reserve, error) {
	if s == nil {
		return Reserve{}, fmt.Errorf("%w: snapshot is nil", ErrInvalidMarketSnapshot)
	}
	if err := ref.Validate(); err != nil {
		return Reserve{}, err
	}
	if ref.Chain() != s.market.Chain() {
		return Reserve{}, fmt.Errorf(
			"%w: reference chain %d does not match market chain %d",
			ErrReserveNotFound,
			ref.Chain(),
			s.market.Chain(),
		)
	}
	reserve, ok := s.reserves[ref.Address()]
	if !ok {
		return Reserve{}, fmt.Errorf("%w: %s", ErrReserveNotFound, ref.Address().Hex())
	}
	return reserve, nil
}

// ReserveByAddress resolves a reserve by an address scoped to the snapshot's
// market chain.
func (s *MarketSnapshot) ReserveByAddress(address common.Address) (Reserve, error) {
	if s == nil {
		return Reserve{}, fmt.Errorf("%w: snapshot is nil", ErrInvalidMarketSnapshot)
	}
	ref, err := token.NewRef(s.market.Chain(), address)
	if err != nil {
		return Reserve{}, err
	}
	return s.Reserve(ref)
}

func validateMarket(market Market) error {
	if market.id == "" {
		return fmt.Errorf("%w: ID is empty", ErrInvalidMarket)
	}
	if _, err := market.chain.ChainID(); err != nil {
		return fmt.Errorf("%w: chain %d: %v", ErrInvalidMarket, market.chain, err)
	}
	required := []struct {
		name    string
		address common.Address
	}{
		{name: "pool", address: market.pool},
		{name: "addresses provider", address: market.addressesProvider},
		{name: "protocol data provider", address: market.protocolDataProvider},
	}
	seen := make(map[common.Address]string, len(required)+1)
	for _, field := range required {
		if field.address == (common.Address{}) {
			return fmt.Errorf("%w: %s address is zero", ErrInvalidMarket, field.name)
		}
		if previous, ok := seen[field.address]; ok {
			return fmt.Errorf(
				"%w: %s and %s use the same address %s",
				ErrInvalidMarket,
				previous,
				field.name,
				field.address.Hex(),
			)
		}
		seen[field.address] = field.name
	}
	if market.hasWrappedTokenGateway {
		if market.wrappedTokenGateway == (common.Address{}) {
			return fmt.Errorf("%w: wrapped token gateway presence has a zero address", ErrInvalidMarket)
		}
		if previous, ok := seen[market.wrappedTokenGateway]; ok {
			return fmt.Errorf(
				"%w: %s and wrapped token gateway use the same address %s",
				ErrInvalidMarket,
				previous,
				market.wrappedTokenGateway.Hex(),
			)
		}
	} else if market.wrappedTokenGateway != (common.Address{}) {
		return fmt.Errorf("%w: wrapped token gateway address is present without presence flag", ErrInvalidMarket)
	}
	return nil
}

func validateReserve(reserve Reserve) error {
	if err := reserve.market.Validate(); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidReserve, err)
	}
	roles := []struct {
		name  string
		token token.Token
	}{
		{name: "underlying", token: reserve.underlying},
		{name: "aToken", token: reserve.aToken},
		{name: "variable debt token", token: reserve.variableDebtToken},
	}
	if reserve.hasStableDebtTokenValue {
		roles = append(roles, struct {
			name  string
			token token.Token
		}{name: "stable debt token", token: reserve.stableDebtToken})
	} else if err := reserve.stableDebtToken.Validate(); err == nil {
		return fmt.Errorf("%w: stable debt token value is present without presence flag", ErrInvalidReserve)
	}

	seen := make(map[common.Address]string, len(roles))
	for _, role := range roles {
		if err := role.token.Validate(); err != nil {
			return fmt.Errorf("%w: %s: %w", ErrInvalidReserve, role.name, err)
		}
		if role.token.Chain() != reserve.market.Chain() {
			return fmt.Errorf(
				"%w: %s chain %d does not match market chain %d",
				ErrInvalidReserve,
				role.name,
				role.token.Chain(),
				reserve.market.Chain(),
			)
		}
		if previous, ok := seen[role.token.Address()]; ok {
			return fmt.Errorf(
				"%w: %s and %s use the same address %s",
				ErrInvalidReserve,
				previous,
				role.name,
				role.token.Address().Hex(),
			)
		}
		seen[role.token.Address()] = role.name
	}
	return nil
}

func indexReserves(market Market, reserves []Reserve) (map[common.Address]Reserve, error) {
	indexed := make(map[common.Address]Reserve, len(reserves))
	for i, reserve := range reserves {
		if err := reserve.Validate(); err != nil {
			return nil, fmt.Errorf("%w: reserve %d: %w", ErrInvalidMarketSnapshot, i+1, err)
		}
		if !reserve.Market().SameMarket(market) {
			return nil, fmt.Errorf(
				"%w: reserve %d belongs to market %q, expected %q",
				ErrInvalidMarketSnapshot,
				i+1,
				reserve.Market().ID(),
				market.ID(),
			)
		}
		underlying := reserve.Underlying().Address()
		if _, ok := indexed[underlying]; ok {
			return nil, fmt.Errorf(
				"%w: %w: %s",
				ErrInvalidMarketSnapshot,
				ErrDuplicateReserve,
				underlying.Hex(),
			)
		}
		indexed[underlying] = reserve
	}
	return indexed, nil
}
