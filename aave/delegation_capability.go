package aave

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/tn606024/defi-simplify/helper"
)

// DelegationCapability identifies a resolved reserve whose variable debt
// token supports Aave credit-delegation signatures with an explicitly reviewed
// EIP-712 domain version.
type DelegationCapability struct {
	reserve Reserve
	version string
}

// NewDelegationCapability creates an explicit signature capability for one
// reserve's variable debt token.
func NewDelegationCapability(reserve Reserve, version string) (DelegationCapability, error) {
	capability := DelegationCapability{reserve: reserve, version: strings.TrimSpace(version)}
	if err := capability.Validate(); err != nil {
		return DelegationCapability{}, err
	}
	return capability, nil
}

// Validate checks the reserve and the debt-token EIP-712 domain fields.
func (c DelegationCapability) Validate() error {
	if err := c.reserve.Validate(); err != nil {
		return fmt.Errorf("invalid delegation reserve: %w", err)
	}
	if strings.TrimSpace(c.reserve.VariableDebtToken().Name()) == "" {
		return fmt.Errorf("variable debt token name is empty")
	}
	if strings.TrimSpace(c.version) == "" {
		return fmt.Errorf("delegation domain version is empty")
	}
	return nil
}

// Reserve returns the resolved reserve governed by this capability.
func (c DelegationCapability) Reserve() Reserve {
	return c.reserve
}

// Version returns the reviewed variable-debt-token EIP-712 domain version.
func (c DelegationCapability) Version() string {
	return c.version
}

// Domain returns the complete EIP-712 domain used for credit-delegation
// signatures.
func (c DelegationCapability) Domain() (*helper.EIP712Domain, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	debtToken := c.reserve.VariableDebtToken()
	chainID, err := debtToken.Chain().ChainID()
	if err != nil {
		return nil, fmt.Errorf("resolve delegation chain ID: %w", err)
	}
	return helper.NewEIP712Domain(
		debtToken.Name(),
		c.version,
		big.NewInt(int64(chainID)),
		debtToken.Address(),
	), nil
}
