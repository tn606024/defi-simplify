package aave

import (
	"errors"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/tn606024/defi-simplify/config"
	"github.com/tn606024/defi-simplify/token"
)

func TestValidateMarket(t *testing.T) {
	valid := marketFixture(t, "aave-v3-base", config.Base, 1)
	tests := []struct {
		name   string
		mutate func(*Market)
	}{
		{name: "empty ID", mutate: func(m *Market) { m.id = "" }},
		{name: "unsupported chain", mutate: func(m *Market) { m.chain = config.Chain(999) }},
		{name: "zero pool", mutate: func(m *Market) { m.pool = common.Address{} }},
		{name: "zero addresses provider", mutate: func(m *Market) { m.addressesProvider = common.Address{} }},
		{name: "zero data provider", mutate: func(m *Market) { m.protocolDataProvider = common.Address{} }},
		{name: "duplicate required address", mutate: func(m *Market) { m.protocolDataProvider = m.pool }},
		{name: "duplicate gateway", mutate: func(m *Market) { m.wrappedTokenGateway = m.pool }},
		{name: "gateway flag drift", mutate: func(m *Market) { m.hasWrappedTokenGateway = false }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidate := valid
			tt.mutate(&candidate)
			if err := validateMarket(candidate); !errors.Is(err, ErrInvalidMarket) {
				t.Fatalf("validateMarket() error = %v, want errors.Is(ErrInvalidMarket)", err)
			}
		})
	}
}

func TestValidateReserve(t *testing.T) {
	market := marketFixture(t, "aave-v3-base", config.Base, 1)
	valid := reserveFixture(t, market, 20)
	tests := []struct {
		name   string
		mutate func(*Reserve)
	}{
		{name: "invalid market", mutate: func(r *Reserve) { r.market = Market{} }},
		{name: "invalid underlying", mutate: func(r *Reserve) { r.underlying = token.Token{} }},
		{name: "duplicate role address", mutate: func(r *Reserve) { r.variableDebtToken = r.aToken }},
		{name: "stable token flag drift", mutate: func(r *Reserve) {
			r.stableDebtToken = tokenFixture(t, config.Base, 30)
			r.hasStableDebtTokenValue = false
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidate := valid
			tt.mutate(&candidate)
			if err := validateReserve(candidate); !errors.Is(err, ErrInvalidReserve) {
				t.Fatalf("validateReserve() error = %v, want errors.Is(ErrInvalidReserve)", err)
			}
		})
	}
}

func TestIndexReserves(t *testing.T) {
	market := marketFixture(t, "aave-v3-base", config.Base, 1)
	first := reserveFixture(t, market, 20)
	second := reserveFixture(t, market, 30)
	otherMarket := marketFixture(t, "other-market", config.Base, 40)

	tests := []struct {
		name     string
		reserves []Reserve
		wantErr  error
	}{
		{name: "empty"},
		{name: "valid", reserves: []Reserve{first, second}},
		{name: "duplicate underlying", reserves: []Reserve{first, first}, wantErr: ErrDuplicateReserve},
		{name: "wrong market", reserves: []Reserve{first, reserveFixture(t, otherMarket, 50)}, wantErr: ErrInvalidMarketSnapshot},
		{name: "invalid reserve", reserves: []Reserve{{}}, wantErr: ErrInvalidMarketSnapshot},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			indexed, err := indexReserves(market, tt.reserves)
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("indexReserves() error = %v", err)
				}
				if len(indexed) != len(tt.reserves) {
					t.Fatalf("indexReserves() len = %d, want %d", len(indexed), len(tt.reserves))
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("indexReserves() error = %v, want errors.Is(%v)", err, tt.wantErr)
			}
		})
	}
}

func marketFixture(t *testing.T, id string, chain config.Chain, seed byte) Market {
	t.Helper()
	market, err := NewMarket(
		id,
		chain,
		addressFixture(seed),
		addressFixture(seed+1),
		addressFixture(seed+2),
		addressFixture(seed+3),
	)
	if err != nil {
		t.Fatalf("NewMarket() error = %v", err)
	}
	return market
}

func reserveFixture(t *testing.T, market Market, seed byte) Reserve {
	t.Helper()
	reserve, err := NewReserve(
		market,
		tokenFixture(t, market.Chain(), seed),
		tokenFixture(t, market.Chain(), seed+1),
		tokenFixture(t, market.Chain(), seed+2),
		nil,
	)
	if err != nil {
		t.Fatalf("NewReserve() error = %v", err)
	}
	return reserve
}

func tokenFixture(t *testing.T, chain config.Chain, seed byte) token.Token {
	t.Helper()
	ref, err := token.NewRef(chain, addressFixture(seed))
	if err != nil {
		t.Fatalf("token.NewRef() error = %v", err)
	}
	resolved, err := token.New(ref, "TOKEN", "Token", 18)
	if err != nil {
		t.Fatalf("token.New() error = %v", err)
	}
	return resolved
}

func addressFixture(seed byte) common.Address {
	var address common.Address
	address[len(address)-1] = seed
	return address
}
