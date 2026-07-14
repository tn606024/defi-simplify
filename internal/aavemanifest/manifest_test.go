package aavemanifest

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
)

func TestGenerateMatchesCheckedInManifest(t *testing.T) {
	exported, err := os.ReadFile("testdata/aave-v3-base-export.json")
	if err != nil {
		t.Fatalf("read export fixture: %v", err)
	}
	want, err := os.ReadFile("../../aave/manifests/aave-v3-base.json")
	if err != nil {
		t.Fatalf("read checked-in manifest: %v", err)
	}

	first, err := Generate(exported)
	if err != nil {
		t.Fatalf("generate manifest: %v", err)
	}
	second, err := Generate(exported)
	if err != nil {
		t.Fatalf("regenerate manifest: %v", err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("repeated generation produced different bytes")
	}
	if !bytes.Equal(first, want) {
		t.Fatalf("generated manifest differs from checked-in manifest\n%s", first)
	}
}

func TestGenerateRejectsInvalidExports(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Export)
	}{
		{
			name: "wrong package",
			mutate: func(exported *Export) {
				exported.PackageName = "untrusted/address-book"
			},
		},
		{
			name: "wrong chain",
			mutate: func(exported *Export) {
				exported.ChainID = 1
			},
		},
		{
			name: "zero address",
			mutate: func(exported *Export) {
				exported.Contracts.Pool = "0x0000000000000000000000000000000000000000"
			},
		},
		{
			name: "duplicate address",
			mutate: func(exported *Export) {
				exported.Contracts.AaveProtocolDataProvider = exported.Contracts.Pool
			},
		},
		{
			name: "malformed commit",
			mutate: func(exported *Export) {
				exported.GitHead = "not-a-commit"
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			exported := validExport()
			test.mutate(&exported)
			encoded, err := json.Marshal(exported)
			if err != nil {
				t.Fatalf("encode export: %v", err)
			}
			_, err = Generate(encoded)
			if !errors.Is(err, ErrInvalidExport) {
				t.Fatalf("Generate() error = %v, want ErrInvalidExport", err)
			}
		})
	}
}

func TestParseRejectsUnreviewedManifestShape(t *testing.T) {
	exported, err := json.Marshal(validExport())
	if err != nil {
		t.Fatalf("encode export: %v", err)
	}
	valid, err := Generate(exported)
	if err != nil {
		t.Fatalf("generate valid manifest: %v", err)
	}

	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "unknown field",
			data: []byte(strings.Replace(
				string(valid),
				"\"marketId\": \"aave-v3-base\"",
				"\"marketId\": \"aave-v3-base\", \"unreviewed\": true",
				1,
			)),
		},
		{
			name: "trailing JSON",
			data: append(append([]byte(nil), valid...), []byte("{}")...),
		},
		{
			name: "unknown market",
			data: []byte(strings.Replace(string(valid), BaseMarketID, "aave-v3-unknown", 1)),
		},
		{
			name: "release mismatch",
			data: []byte(strings.Replace(string(valid), "v4.60.0", "v4.59.0", 1)),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Parse(test.data)
			if !errors.Is(err, ErrInvalidManifest) {
				t.Fatalf("Parse() error = %v, want ErrInvalidManifest", err)
			}
		})
	}
}

func validExport() Export {
	return Export{
		PackageName:    OfficialPackage,
		PackageVersion: "4.60.0",
		GitHead:        "7e444a1e73b538fd0b9e093e5156401d6fccca7d",
		Export:         BaseExportName,
		ChainID:        BaseChainID,
		Contracts: Contracts{
			PoolAddressesProvider:    "0xe20fCBdBfFC4Dd138cE8b2E6FBb6CB49777ad64D",
			Pool:                     "0xA238Dd80C259a72e81d7e4664a9801593F98d1c5",
			AaveProtocolDataProvider: "0x0F43731EB8d45A581f4a36DD74F5f358bc90C73A",
			WrappedTokenGateway:      "0xa0d9C1E9E48Ca30c8d8C3B5D69FF5dc1f6DFfC24",
		},
	}
}
