package main

import (
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	"github.com/tn606024/defi-simplify/client/account/eip7702"
)

func TestValidatePhasePreconditions(t *testing.T) {
	implementation := common.HexToAddress("0x1000000000000000000000000000000000000001")
	tests := []struct {
		name    string
		config  commandConfig
		mutate  func(*accountSnapshot)
		wantErr string
	}{
		{name: "inspect", config: commandConfig{phase: phaseInspect}},
		{name: "delegate clean", config: commandConfig{phase: phaseDelegate}, mutate: cleanDelegation},
		{name: "delegate refuses installed code", config: commandConfig{phase: phaseDelegate}, wantErr: "requires a clean EOA"},
		{name: "open", config: openTestConfig()},
		{name: "open requires delegation", config: openTestConfig(), mutate: cleanDelegation, wantErr: "must delegate"},
		{name: "open refuses collateral", config: openTestConfig(), mutate: func(state *accountSnapshot) { state.wethPosition.aTokenBalance = big.NewInt(1) }, wantErr: "zero WETH and USDC"},
		{name: "open checks native balance", config: openTestConfig(), mutate: func(state *accountSnapshot) { state.nativeBalance = big.NewInt(1) }, wantErr: "native balance"},
		{name: "close", config: closeTestConfig(), mutate: openPosition},
		{name: "close requires a position", config: closeTestConfig(), wantErr: "requires WETH collateral"},
		{name: "close refuses stable debt", config: closeTestConfig(), mutate: func(state *accountSnapshot) { openPosition(state); state.usdcPosition.stableDebt = big.NewInt(1) }, wantErr: "isolated WETH-collateral"},
		{name: "close checks repay bound", config: commandConfig{phase: phaseClose, usdcRepayLimit: decimal.NewFromInt(9)}, mutate: openPosition, wantErr: "below current variable debt"},
		{name: "close checks token balance", config: closeTestConfig(), mutate: func(state *accountSnapshot) { openPosition(state); state.usdcToken.balance = big.NewInt(10_000_000) }, wantErr: "cannot cover configured repay limit"},
		{name: "clear", config: commandConfig{phase: phaseClear}},
		{name: "clear refuses debt", config: commandConfig{phase: phaseClear}, mutate: openPosition, wantErr: "zero targeted collateral and debt"},
		{name: "clear refuses allowance", config: commandConfig{phase: phaseClear}, mutate: func(state *accountSnapshot) { state.usdcToken.allowance = big.NewInt(1) }, wantErr: "zero Aave Pool allowance"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state := emptyAccountSnapshot(implementation)
			if test.mutate != nil {
				test.mutate(state)
			}
			err := validatePhasePreconditions(test.config, state, implementation)
			if test.wantErr == "" {
				if err != nil {
					t.Fatalf("validatePhasePreconditions() error = %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("validatePhasePreconditions() error = %v, want substring %q", err, test.wantErr)
			}
		})
	}
}

func emptyAccountSnapshot(implementation common.Address) *accountSnapshot {
	return &accountSnapshot{
		delegation: eip7702.DelegationState{
			Status:         eip7702.DelegationStatusDelegated,
			Implementation: implementation,
		},
		nativeBalance: big.NewInt(1_000_000_000_000_000_000),
		wethDecimals:  18,
		usdcDecimals:  6,
		wethToken: tokenStateSnapshot{
			balance:   new(big.Int),
			allowance: new(big.Int),
		},
		usdcToken: tokenStateSnapshot{
			balance:   big.NewInt(11_000_000),
			allowance: new(big.Int),
		},
		wethPosition: reserveStateSnapshot{
			aTokenBalance: new(big.Int),
			stableDebt:    new(big.Int),
			variableDebt:  new(big.Int),
		},
		usdcPosition: reserveStateSnapshot{
			aTokenBalance: new(big.Int),
			stableDebt:    new(big.Int),
			variableDebt:  new(big.Int),
		},
	}
}

func cleanDelegation(state *accountSnapshot) {
	state.delegation = eip7702.DelegationState{Status: eip7702.DelegationStatusClean}
}

func openPosition(state *accountSnapshot) {
	state.wethPosition.aTokenBalance = big.NewInt(100_000_000_000_000_000)
	state.usdcPosition.variableDebt = big.NewInt(10_000_000)
}

func openTestConfig() commandConfig {
	return commandConfig{
		phase:            phaseOpen,
		wethSupplyAmount: decimal.RequireFromString("0.1"),
		usdcBorrowAmount: decimal.NewFromInt(10),
	}
}

func closeTestConfig() commandConfig {
	return commandConfig{
		phase:          phaseClose,
		usdcRepayLimit: decimal.RequireFromString("10.1"),
	}
}
