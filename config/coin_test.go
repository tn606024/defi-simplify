package config

import (
	"strings"
	"testing"
)

func TestCoinConfigErrors(t *testing.T) {
	unsupportedCoin := Coin(999)

	if _, err := unsupportedCoin.Address(Base); err == nil || !strings.Contains(err.Error(), "unsupported coin address") {
		t.Fatalf("expected unsupported coin address error, got %v", err)
	}

	if _, err := unsupportedCoin.Decimals(); err == nil || !strings.Contains(err.Error(), "unsupported coin decimals") {
		t.Fatalf("expected unsupported coin decimals error, got %v", err)
	}

	if _, err := unsupportedCoin.Name(Base); err == nil || !strings.Contains(err.Error(), "unsupported coin name") {
		t.Fatalf("expected unsupported coin name error, got %v", err)
	}

	if _, err := unsupportedCoin.PermitVersion(Base); err == nil || !strings.Contains(err.Error(), "unsupported permit version") {
		t.Fatalf("expected unsupported permit version error, got %v", err)
	}

	if _, err := unsupportedCoin.PermitDomain(Base); err == nil || !strings.Contains(err.Error(), "unsupported coin name") {
		t.Fatalf("expected permit domain to fail on missing coin config, got %v", err)
	}
}

func TestAaveTokenConfigErrors(t *testing.T) {
	if _, err := GHO.AToken(); err == nil || !strings.Contains(err.Error(), "unsupported aToken") {
		t.Fatalf("expected unsupported aToken error, got %v", err)
	}

	if _, err := GHO.DebtToken(); err == nil || !strings.Contains(err.Error(), "unsupported debt token") {
		t.Fatalf("expected unsupported debt token error, got %v", err)
	}
}
