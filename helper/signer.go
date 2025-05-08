package helper

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

// MsgSigner is responsible for signing EIP-712 messages.
type MsgSigner struct {
	SignEIP712Msg func(msg EIP712Msg) ([]byte, error)
}

// NewMsgSigner creates a new MsgSigner with a given private key.
func NewMsgSigner(key *ecdsa.PrivateKey) *MsgSigner {
	return &MsgSigner{
		SignEIP712Msg: func(msg EIP712Msg) ([]byte, error) {
			sig, err := msg.Sighash()
			if err != nil {
				return nil, err
			}
			return SignTypedData(sig, key)
		},
	}
}

// EIP712Msg represents a message that can be signed using EIP-712.
type EIP712Msg interface {
	Sighash() ([]byte, error)
}

// EIP712Domain represents the domain separator for EIP-712.
type EIP712Domain struct {
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
}

// NewEIP712Domain creates a new EIP712Domain.
func NewEIP712Domain(name, version string, chainId *big.Int, verifyingContract common.Address) *EIP712Domain {
	return &EIP712Domain{
		Name:              name,
		Version:           version,
		ChainId:           chainId,
		VerifyingContract: verifyingContract,
	}
}

// Permit represents the EIP-2612 permit structure.
type Permit struct {
	Owner    common.Address
	Spender  common.Address
	Value    *big.Int
	Nonce    *big.Int
	Deadline *big.Int
}

// NewPermit creates a new Permit.
func NewPermit(owner, spender common.Address, value, nonce, deadline *big.Int) *Permit {
	return &Permit{
		Owner:    owner,
		Spender:  spender,
		Value:    value,
		Nonce:    nonce,
		Deadline: deadline,
	}
}

// PermitEIP712Msg represents a permit message for EIP-712 signing.
type PermitEIP712Msg struct {
	EIP712Domain
	Permit
}

// NewPermitEIP712Msg creates a new PermitEIP712Msg.
func NewPermitEIP712Msg(domain *EIP712Domain, permit *Permit) *PermitEIP712Msg {
	return &PermitEIP712Msg{
		EIP712Domain: *domain,
		Permit:       *permit,
	}
}

// createTypedData creates the TypedData for the permit.
func (p *PermitEIP712Msg) createTypedData() apitypes.TypedData {
	return apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": []apitypes.Type{
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			"Permit": []apitypes.Type{
				{Name: "owner", Type: "address"},
				{Name: "spender", Type: "address"},
				{Name: "value", Type: "uint256"},
				{Name: "nonce", Type: "uint256"},
				{Name: "deadline", Type: "uint256"},
			},
		},
		PrimaryType: "Permit",
		Domain: apitypes.TypedDataDomain{
			Name:              p.EIP712Domain.Name,
			Version:           p.EIP712Domain.Version,
			ChainId:           (*math.HexOrDecimal256)(p.EIP712Domain.ChainId),
			VerifyingContract: p.EIP712Domain.VerifyingContract.Hex(),
		},
		Message: map[string]interface{}{
			"owner":    p.Owner.Hex(),
			"spender":  p.Spender.Hex(),
			"value":    p.Value.String(),
			"nonce":    p.Nonce.String(),
			"deadline": p.Deadline.String(),
		},
	}
}

// Sighash computes the EIP-712 hash of the permit message.
func (p *PermitEIP712Msg) Sighash() ([]byte, error) {
	typedData := p.createTypedData()
	return hashTypedData(typedData)
}

// hashTypedData hashes the TypedData according to EIP-712.
func hashTypedData(typedData apitypes.TypedData) ([]byte, error) {
	domainSeparator, err := typedData.HashStruct("EIP712Domain", typedData.Domain.Map())
	if err != nil {
		return nil, err
	}
	typedDataHash, err := typedData.HashStruct(typedData.PrimaryType, typedData.Message)
	if err != nil {
		return nil, err
	}
	rawData := []byte(fmt.Sprintf("\x19\x01%s%s", string(domainSeparator), string(typedDataHash)))
	sighash := crypto.Keccak256(rawData)
	return sighash, nil
}

// SignTypedData signs the hash with the given private key.
func SignTypedData(hash []byte, privateKey *ecdsa.PrivateKey) ([]byte, error) {
	return crypto.Sign(hash, privateKey)
}

type DelegationWithSig struct {
	Delegatee common.Address
	Value     *big.Int
	Nonce     *big.Int
	Deadline  *big.Int
}

func NewDelegationWithSig(delegatee common.Address, value, nonce, deadline *big.Int) *DelegationWithSig {
	return &DelegationWithSig{
		Delegatee: delegatee,
		Value:     value,
		Nonce:     nonce,
		Deadline:  deadline,
	}
}

type DelegationWithSigEIP712Msg struct {
	EIP712Domain
	DelegationWithSig
}

// NewDelegationWithSigEIP712Msg creates a new DelegationWithSigEIP712Msg.
func NewDelegationWithSigEIP712Msg(domain *EIP712Domain, DelegationWithSig *DelegationWithSig) *DelegationWithSigEIP712Msg {
	return &DelegationWithSigEIP712Msg{
		EIP712Domain:      *domain,
		DelegationWithSig: *DelegationWithSig,
	}
}

// createTypedData creates the TypedData for the DelegationWithSig.
func (d *DelegationWithSigEIP712Msg) createTypedData() apitypes.TypedData {
	return apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": []apitypes.Type{
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			"DelegationWithSig": []apitypes.Type{
				{Name: "delegatee", Type: "address"},
				{Name: "value", Type: "uint256"},
				{Name: "nonce", Type: "uint256"},
				{Name: "deadline", Type: "uint256"},
			},
		},
		PrimaryType: "DelegationWithSig",
		Domain: apitypes.TypedDataDomain{
			Name:              d.EIP712Domain.Name,
			Version:           d.EIP712Domain.Version,
			ChainId:           (*math.HexOrDecimal256)(d.EIP712Domain.ChainId),
			VerifyingContract: d.EIP712Domain.VerifyingContract.Hex(),
		},
		Message: map[string]interface{}{
			"delegatee": d.Delegatee.Hex(),
			"value":     d.Value.String(),
			"nonce":     d.Nonce.String(),
			"deadline":  d.Deadline.String(),
		},
	}
}

// Sighash computes the EIP-712 hash of the DelegationWithSig message.
func (d *DelegationWithSigEIP712Msg) Sighash() ([]byte, error) {
	typedData := d.createTypedData()
	return hashTypedData(typedData)
}

func SignEIP712MsgAndGetVRS(
	signer *MsgSigner,
	msg EIP712Msg,
) (uint8, [32]byte, [32]byte, error) {
	signature, err := signer.SignEIP712Msg(msg)
	if err != nil {
		return 0, [32]byte{}, [32]byte{}, err
	}
	v := uint8(signature[64])
	if v < 27 {
		v += 27
	}
	// Ensure r and s are properly formatted
	r := [32]byte{}
	s := [32]byte{}
	copy(r[:], signature[:32])
	copy(s[:], signature[32:64])
	return v, r, s, nil
}
