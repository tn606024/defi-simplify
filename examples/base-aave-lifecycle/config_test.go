package main

import (
	"strings"
	"testing"
)

func TestParseCommandConfig(t *testing.T) {
	environment := func(name string) string {
		switch name {
		case baseRPCURLEnv:
			return "http://127.0.0.1:8545"
		case privateKeyEnv:
			return strings.Repeat("1", 64)
		default:
			return ""
		}
	}
	tests := []struct {
		name      string
		args      []string
		getenv    func(string) string
		wantPhase lifecyclePhase
		broadcast bool
		wantErr   string
	}{
		{name: "inspect", getenv: environment, wantPhase: phaseInspect},
		{name: "delegate dry run", args: []string{"--delegate"}, getenv: environment, wantPhase: phaseDelegate},
		{name: "delegate broadcast", args: []string{"--delegate", "--broadcast"}, getenv: environment, wantPhase: phaseDelegate, broadcast: true},
		{name: "open", args: []string{"--open", "--weth-supply", "0.1", "--usdc-borrow", "10"}, getenv: environment, wantPhase: phaseOpen},
		{name: "close", args: []string{"--close", "--usdc-repay-limit", "10.1"}, getenv: environment, wantPhase: phaseClose},
		{name: "multiple phases", args: []string{"--open", "--close"}, getenv: environment, wantErr: "select exactly one"},
		{name: "broadcast without phase", args: []string{"--broadcast"}, getenv: environment, wantErr: "requires one lifecycle phase"},
		{name: "missing open amount", args: []string{"--open", "--weth-supply", "0.1"}, getenv: environment, wantErr: "--usdc-borrow is required"},
		{name: "zero amount", args: []string{"--open", "--weth-supply", "0", "--usdc-borrow", "10"}, getenv: environment, wantErr: "must be positive"},
		{name: "excess USDC precision", args: []string{"--open", "--weth-supply", "0.1", "--usdc-borrow", "0.0000001"}, getenv: environment, wantErr: "exceeds 6 decimal places"},
		{name: "open flag on close", args: []string{"--close", "--usdc-repay-limit", "10", "--weth-supply", "0.1"}, getenv: environment, wantErr: "only valid with --open"},
		{name: "amount without flow", args: []string{"--delegate", "--weth-supply", "0.1"}, getenv: environment, wantErr: "amount flags require"},
		{name: "missing environment", getenv: func(string) string { return "" }, wantErr: "BASE_RPC_URL is required"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config, err := parseCommandConfig(test.args, test.getenv)
			if test.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), test.wantErr) {
					t.Fatalf("parseCommandConfig() error = %v, want substring %q", err, test.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseCommandConfig() error = %v", err)
			}
			if config.phase != test.wantPhase {
				t.Fatalf("phase = %q, want %q", config.phase, test.wantPhase)
			}
			if config.broadcast != test.broadcast {
				t.Fatalf("broadcast = %t, want %t", config.broadcast, test.broadcast)
			}
		})
	}
}
