package erc20

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/tn606024/defi-simplify/helper"
	"github.com/tn606024/defi-simplify/token"
)

// PermitCapability identifies a resolved token whose EIP-2612-style permit
// domain has been reviewed explicitly. Capability construction never probes
// nonces() or infers support from token metadata.
type PermitCapability struct {
	token   token.Token
	version string
}

// NewPermitCapability creates an explicit permit capability for one resolved
// token and its reviewed EIP-712 domain version.
func NewPermitCapability(asset token.Token, version string) (PermitCapability, error) {
	capability := PermitCapability{token: asset, version: strings.TrimSpace(version)}
	if err := capability.Validate(); err != nil {
		return PermitCapability{}, err
	}
	return capability, nil
}

// Validate checks the resolved token identity and the explicit domain fields
// required for permit signing.
func (c PermitCapability) Validate() error {
	if err := c.token.Validate(); err != nil {
		return fmt.Errorf("invalid permit token: %w", err)
	}
	if strings.TrimSpace(c.token.Name()) == "" {
		return fmt.Errorf("permit token name is empty")
	}
	if strings.TrimSpace(c.version) == "" {
		return fmt.Errorf("permit domain version is empty")
	}
	return nil
}

// Token returns the resolved token governed by this permit capability.
func (c PermitCapability) Token() token.Token {
	return c.token
}

// Version returns the reviewed EIP-712 domain version.
func (c PermitCapability) Version() string {
	return c.version
}

// Domain returns the complete EIP-712 domain used to sign permits for the
// resolved token.
func (c PermitCapability) Domain() (*helper.EIP712Domain, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	chainID, err := c.token.Chain().ChainID()
	if err != nil {
		return nil, fmt.Errorf("resolve permit chain ID: %w", err)
	}
	return helper.NewEIP712Domain(
		c.token.Name(),
		c.version,
		big.NewInt(int64(chainID)),
		c.token.Address(),
	), nil
}
