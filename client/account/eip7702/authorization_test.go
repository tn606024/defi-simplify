package eip7702_test

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tn606024/defi-simplify/client/account/eip7702"
	"github.com/tn606024/defi-simplify/config"
)

func TestSignAuthorizationRecoversAuthority(t *testing.T) {
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	authority := crypto.PubkeyToAddress(key.PublicKey)
	implementation := common.HexToAddress("0x1000000000000000000000000000000000000000")

	auth, err := eip7702.SignAuthorization(key, big.NewInt(8453), implementation, 7)
	if err != nil {
		t.Fatalf("sign authorization: %v", err)
	}

	if auth.Address != implementation {
		t.Fatalf("unexpected implementation: got %s want %s", auth.Address.Hex(), implementation.Hex())
	}
	if auth.Nonce != 7 {
		t.Fatalf("unexpected nonce: got %d want 7", auth.Nonce)
	}
	if auth.ChainID.CmpBig(big.NewInt(8453)) != 0 {
		t.Fatalf("unexpected chain id: got %s want 8453", auth.ChainID.String())
	}

	gotAuthority, err := auth.Authority()
	if err != nil {
		t.Fatalf("recover authority: %v", err)
	}
	if gotAuthority != authority {
		t.Fatalf("unexpected authority: got %s want %s", gotAuthority.Hex(), authority.Hex())
	}
}

func TestSignClearAuthorizationUsesZeroAddress(t *testing.T) {
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	auth, err := eip7702.SignClearAuthorization(key, big.NewInt(8453), 9)
	if err != nil {
		t.Fatalf("sign clear authorization: %v", err)
	}

	if auth.Address != (common.Address{}) {
		t.Fatalf("clear authorization should target zero address, got %s", auth.Address.Hex())
	}
	if auth.Nonce != 9 {
		t.Fatalf("unexpected nonce: got %d want 9", auth.Nonce)
	}
}

func TestSignSimple7702AuthorizationUsesConfiguredImplementation(t *testing.T) {
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	want, err := config.Base.Simple7702AccountImplementationAddress()
	if err != nil {
		t.Fatalf("simple7702 config: %v", err)
	}

	auth, err := eip7702.SignSimple7702Authorization(key, config.Base, 3)
	if err != nil {
		t.Fatalf("sign simple7702 authorization: %v", err)
	}

	if auth.Address != want {
		t.Fatalf("unexpected Simple7702Account implementation: got %s want %s", auth.Address.Hex(), want.Hex())
	}
	if auth.Nonce != 3 {
		t.Fatalf("unexpected nonce: got %d want 3", auth.Nonce)
	}
}

func TestBuildUnsignedAuthorizationRejectsInvalidChainID(t *testing.T) {
	if _, err := eip7702.BuildUnsignedAuthorization(nil, common.Address{}, 0); err == nil {
		t.Fatal("expected nil chain id error")
	}
	if _, err := eip7702.BuildUnsignedAuthorization(big.NewInt(-1), common.Address{}, 0); err == nil {
		t.Fatal("expected negative chain id error")
	}
}

func TestAuthorizationHelpersReturnGethSetCodeAuthorization(t *testing.T) {
	auth, err := eip7702.BuildUnsignedAuthorization(big.NewInt(8453), common.Address{}, 1)
	if err != nil {
		t.Fatalf("build authorization: %v", err)
	}

	var _ types.SetCodeAuthorization = auth
}
