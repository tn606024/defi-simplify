package config

import (
	"strings"
	"testing"
)

func TestChainConfigErrors(t *testing.T) {
	unsupportedChain := Chain(999)

	if _, err := unsupportedChain.ChainID(); err == nil || !strings.Contains(err.Error(), "unsupported chain id") {
		t.Fatalf("expected unsupported chain id error, got %v", err)
	}

	if _, err := unsupportedChain.GasTokenDecimals(); err == nil || !strings.Contains(err.Error(), "unsupported gas token decimals") {
		t.Fatalf("expected unsupported gas token decimals error, got %v", err)
	}

	if _, err := unsupportedChain.AaveV3PoolAddress(); err == nil || !strings.Contains(err.Error(), "unsupported Aave V3 pool address") {
		t.Fatalf("expected unsupported Aave pool address error, got %v", err)
	}

	if _, err := unsupportedChain.WrappedTokenGatewayV3Address(); err == nil || !strings.Contains(err.Error(), "unsupported wrapped token gateway address") {
		t.Fatalf("expected unsupported wrapped token gateway address error, got %v", err)
	}

	if _, err := unsupportedChain.AaveProtocolDataProviderAddress(); err == nil || !strings.Contains(err.Error(), "unsupported Aave protocol data provider address") {
		t.Fatalf("expected unsupported protocol data provider address error, got %v", err)
	}

	if _, err := unsupportedChain.MulticallAddress(); err == nil || !strings.Contains(err.Error(), "unsupported multicall address") {
		t.Fatalf("expected unsupported multicall address error, got %v", err)
	}

	if _, err := unsupportedChain.Simple7702AccountImplementationAddress(); err == nil || !strings.Contains(err.Error(), "unsupported Simple7702Account implementation address") {
		t.Fatalf("expected unsupported Simple7702Account implementation address error, got %v", err)
	}
}

func TestBaseSimple7702AccountImplementationAddress(t *testing.T) {
	address, err := Base.Simple7702AccountImplementationAddress()
	if err != nil {
		t.Fatalf("expected Base Simple7702Account implementation address, got %v", err)
	}

	if got, want := address.Hex(), "0xa625961dcb8a01c75DBeA172F58181FC5C711dA4"; got != want {
		t.Fatalf("unexpected Base Simple7702Account implementation address: got %s want %s", got, want)
	}
}
