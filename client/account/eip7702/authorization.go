package eip7702

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/holiman/uint256"
	"github.com/tn606024/defi-simplify/config"
)

// BuildUnsignedAuthorization builds an unsigned EIP-7702 authorization tuple.
// Use the zero address as implementation to clear an existing delegation.
func BuildUnsignedAuthorization(chainID *big.Int, implementation common.Address, nonce uint64) (types.SetCodeAuthorization, error) {
	chainIDUint, err := uint256FromBig("chain ID", chainID)
	if err != nil {
		return types.SetCodeAuthorization{}, err
	}

	return types.SetCodeAuthorization{
		ChainID: *chainIDUint,
		Address: implementation,
		Nonce:   nonce,
	}, nil
}

// SignAuthorization signs an EIP-7702 authorization tuple for implementation.
func SignAuthorization(key *ecdsa.PrivateKey, chainID *big.Int, implementation common.Address, nonce uint64) (types.SetCodeAuthorization, error) {
	if key == nil {
		return types.SetCodeAuthorization{}, errors.New("authorization key is nil")
	}

	auth, err := BuildUnsignedAuthorization(chainID, implementation, nonce)
	if err != nil {
		return types.SetCodeAuthorization{}, err
	}
	return types.SignSetCode(key, auth)
}

// SignClearAuthorization signs an authorization that clears an existing delegation.
func SignClearAuthorization(key *ecdsa.PrivateKey, chainID *big.Int, nonce uint64) (types.SetCodeAuthorization, error) {
	return SignAuthorization(key, chainID, common.Address{}, nonce)
}

// SignSimple7702Authorization signs an authorization to the configured
// Simple7702Account implementation for chain.
func SignSimple7702Authorization(key *ecdsa.PrivateKey, chain config.Chain, nonce uint64) (types.SetCodeAuthorization, error) {
	implementation, err := chain.Simple7702AccountImplementationAddress()
	if err != nil {
		return types.SetCodeAuthorization{}, err
	}
	chainID, err := chain.ChainID()
	if err != nil {
		return types.SetCodeAuthorization{}, err
	}
	return SignAuthorization(key, big.NewInt(int64(chainID)), implementation, nonce)
}

func uint256FromBig(label string, value *big.Int) (*uint256.Int, error) {
	if value == nil {
		return nil, fmt.Errorf("%s is nil", label)
	}
	if value.Sign() < 0 {
		return nil, fmt.Errorf("%s must be non-negative", label)
	}
	result, overflow := uint256.FromBig(value)
	if overflow {
		return nil, fmt.Errorf("%s must fit uint256", label)
	}
	return result, nil
}
