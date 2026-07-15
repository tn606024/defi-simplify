// Package token defines protocol-neutral ERC20 token identity and metadata.
package token

import (
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/tn606024/defi-simplify/config"
)

var (
	// ErrInvalidRef is returned when a token reference has an unsupported chain
	// or a zero contract address.
	ErrInvalidRef = errors.New("invalid token reference")
	// ErrInvalidToken is returned when resolved token metadata has an invalid
	// identity.
	ErrInvalidToken = errors.New("invalid token")
)

// Ref identifies a token by chain and contract address without duplicating
// mutable display metadata or protocol-specific relationships.
//
// Named conveniences from packages such as assets/base (for example,
// base.USDC) are Refs. A Ref must be resolved into a Token or a protocol-owned
// model before execution.
type Ref struct {
	chain   config.Chain
	address common.Address
}

// NewRef creates a validated token reference.
func NewRef(chain config.Chain, address common.Address) (Ref, error) {
	if err := validateRef(chain, address); err != nil {
		return Ref{}, err
	}
	return Ref{chain: chain, address: address}, nil
}

// Validate checks whether the reference has a supported chain and non-zero
// address.
func (r Ref) Validate() error {
	return validateRef(r.chain, r.address)
}

// Chain returns the chain that scopes the token address.
func (r Ref) Chain() config.Chain {
	return r.chain
}

// Address returns the token contract address.
func (r Ref) Address() common.Address {
	return r.address
}

// SameAsset reports whether two references identify the same chain and
// contract address.
func (r Ref) SameAsset(other Ref) bool {
	return r == other
}

// Token is resolved protocol-neutral ERC20 metadata. Chain and address are the
// executable identity; symbol and name are display-only metadata.
//
// Token is immutable: all fields are private and all accessors return values.
type Token struct {
	ref      Ref
	symbol   string
	name     string
	decimals uint8
}

// New creates a resolved token value. Symbol and name are intentionally not
// identity fields and may be empty when a token does not expose usable display
// metadata.
func New(ref Ref, symbol string, name string, decimals uint8) (Token, error) {
	if err := ref.Validate(); err != nil {
		return Token{}, fmt.Errorf("%w: %w", ErrInvalidToken, err)
	}
	return Token{
		ref:      ref,
		symbol:   symbol,
		name:     name,
		decimals: decimals,
	}, nil
}

// Validate checks whether the token has a valid executable identity.
func (t Token) Validate() error {
	if err := t.ref.Validate(); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidToken, err)
	}
	return nil
}

// Ref returns the token's chain-scoped identity.
func (t Token) Ref() Ref {
	return t.ref
}

// Chain returns the token's chain.
func (t Token) Chain() config.Chain {
	return t.ref.Chain()
}

// Address returns the token contract address.
func (t Token) Address() common.Address {
	return t.ref.Address()
}

// Symbol returns display-only token metadata.
func (t Token) Symbol() string {
	return t.symbol
}

// Name returns display-only token metadata.
func (t Token) Name() string {
	return t.name
}

// Decimals returns the ERC20 decimal precision used for amount conversion.
func (t Token) Decimals() uint8 {
	return t.decimals
}

// SameAsset reports whether two token values identify the same chain and
// contract address, regardless of their display metadata.
func (t Token) SameAsset(other Token) bool {
	return t.ref.SameAsset(other.ref)
}

func validateRef(chain config.Chain, address common.Address) error {
	if _, err := chain.ChainID(); err != nil {
		return fmt.Errorf("%w: chain %d: %v", ErrInvalidRef, chain, err)
	}
	if address == (common.Address{}) {
		return fmt.Errorf("%w: address is zero", ErrInvalidRef)
	}
	return nil
}
