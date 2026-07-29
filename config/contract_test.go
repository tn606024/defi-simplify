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

	if _, err := unsupportedChain.AaveProtocolDataProviderAddress(); err == nil || !strings.Contains(err.Error(), "unsupported Aave protocol data provider address") {
		t.Fatalf("expected unsupported protocol data provider address error, got %v", err)
	}

}
