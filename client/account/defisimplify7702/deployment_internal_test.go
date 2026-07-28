package defisimplify7702

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/tn606024/defi-simplify/config"
)

func TestParseDeploymentRejectsInvalidRequiredFields(t *testing.T) {
	t.Parallel()

	valid := validDeploymentJSON()
	tests := []struct {
		name    string
		old     string
		replace string
	}{
		{
			name:    "wrong chain",
			old:     `"chainId": 8453`,
			replace: `"chainId": 1`,
		},
		{
			name:    "unsupported schema",
			old:     `"schemaVersion": 1`,
			replace: `"schemaVersion": 2`,
		},
		{
			name:    "missing required contract",
			old:     `"DefiSimplify7702Account"`,
			replace: `"OtherAccount"`,
		},
		{
			name:    "invalid address",
			old:     `"address": "0x0000000000000000000000000000000000000001"`,
			replace: `"address": "invalid"`,
		},
		{
			name:    "invalid runtime hash",
			old:     `"runtimeCodeHash": "0x1111111111111111111111111111111111111111111111111111111111111111"`,
			replace: `"runtimeCodeHash": "0x01"`,
		},
		{
			name:    "wrong ABI path",
			old:     `"abiPath": "abi/DefiSimplify7702Account.json"`,
			replace: `"abiPath": "abi/Other.json"`,
		},
		{
			name:    "missing security notice",
			old:     `"notice": "unsafe"`,
			replace: `"notice": ""`,
		},
		{
			name:    "missing EntryPoint version",
			old:     `"version": "v0.9.0"`,
			replace: `"version": ""`,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			data := strings.Replace(valid, test.old, test.replace, 1)
			if data == valid {
				t.Fatalf("fixture did not contain %q", test.old)
			}
			_, err := parseDeployment(config.Base, 8453, []byte(data))
			if !errors.Is(err, ErrInvalidDeployment) {
				t.Fatalf("parseDeployment() error = %v, want ErrInvalidDeployment", err)
			}
		})
	}
}

func TestParseDeploymentAllowsUnrelatedManifestFields(t *testing.T) {
	t.Parallel()

	data := strings.Replace(validDeploymentJSON(), `"schemaVersion": 1`, `"schemaVersion": 1, "build": {"optimizer": true}`, 1)
	if _, err := parseDeployment(config.Base, 8453, []byte(data)); err != nil {
		t.Fatalf("parseDeployment() error = %v", err)
	}
}

func validDeploymentJSON() string {
	artifacts := ""
	for i, contract := range []Contract{
		AccountContract,
		FlowAssertionsContract,
		StaticCallUint256AssertionsContract,
	} {
		if i > 0 {
			artifacts += ","
		}
		artifacts += fmt.Sprintf(
			`%q: {"contractName": %q, "address": "0x0000000000000000000000000000000000000001", "runtimeCodeHash": "0x1111111111111111111111111111111111111111111111111111111111111111", "abiPath": %q}`,
			contract,
			contract,
			contractABIPaths[contract],
		)
	}
	return fmt.Sprintf(`{
		"schemaVersion": 1,
		"network": {"chainId": 8453, "name": "Base"},
		"entryPoint": {
			"address": "0x0000000000000000000000000000000000000001",
			"runtimeCodeHash": "0x1111111111111111111111111111111111111111111111111111111111111111",
			"version": "v0.9.0"
		},
		"security": {
			"experimental": true,
			"notice": "unsafe",
			"securityGuarantee": false,
			"status": "unaudited",
			"totalLossRisk": true
		},
		"artifacts": {%s}
	}`, artifacts)
}
