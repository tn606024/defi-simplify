package assetmanifest

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/tn606024/defi-simplify/internal/aaveaddressbook"
)

func TestGenerateMatchesCheckedInManifest(t *testing.T) {
	exported, err := os.ReadFile("../aaveaddressbook/testdata/aave-v3-base-export.json")
	if err != nil {
		t.Fatalf("read export fixture: %v", err)
	}
	want, err := os.ReadFile("../../assets/base/manifest.json")
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

func TestGenerateRejectsInvalidAssets(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*aaveaddressbook.Export)
	}{
		{
			name: "empty assets",
			mutate: func(exported *aaveaddressbook.Export) {
				exported.Assets = nil
			},
		},
		{
			name: "unsupported key character",
			mutate: func(exported *aaveaddressbook.Export) {
				exported.Assets[0].Key = "USD.C"
			},
		},
		{
			name: "zero address",
			mutate: func(exported *aaveaddressbook.Export) {
				exported.Assets[0].Address = "0x0000000000000000000000000000000000000000"
			},
		},
		{
			name: "duplicate canonical ID",
			mutate: func(exported *aaveaddressbook.Export) {
				exported.Assets[1].Key = "usdc"
			},
		},
		{
			name: "duplicate address",
			mutate: func(exported *aaveaddressbook.Export) {
				exported.Assets[1].Address = exported.Assets[0].Address
			},
		},
		{
			name: "non-HTTPS issuer source",
			mutate: func(exported *aaveaddressbook.Export) {
				exported.Assets[0].IssuerSource = "http://example.com/token"
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
			if !errors.Is(err, ErrInvalidManifest) {
				t.Fatalf("Generate() error = %v, want ErrInvalidManifest", err)
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
				`"chainId": 8453`,
				`"chainId": 8453, "unreviewed": true`,
				1,
			)),
		},
		{
			name: "wrong chain",
			data: []byte(strings.Replace(string(valid), `"chainId": 8453`, `"chainId": 1`, 1)),
		},
		{
			name: "wrong source export",
			data: []byte(strings.Replace(
				string(valid),
				BaseAssetsExportName,
				aaveaddressbook.BaseExportName,
				1,
			)),
		},
		{
			name: "trailing JSON",
			data: append(append([]byte(nil), valid...), []byte("{}")...),
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

func TestCatalogID(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		want    string
		wantErr bool
	}{
		{name: "already uppercase", key: "USDC", want: "USDC"},
		{name: "mixed case", key: "cbETH", want: "CBETH"},
		{name: "digits", key: "TOKEN2", want: "TOKEN2"},
		{name: "empty", key: "", wantErr: true},
		{name: "leading digit", key: "2TOKEN", wantErr: true},
		{name: "punctuation", key: "USDC.e", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := catalogID(test.key)
			if test.wantErr {
				if err == nil {
					t.Fatalf("catalogID(%q) returned no error", test.key)
				}
				return
			}
			if err != nil {
				t.Fatalf("catalogID(%q) error = %v", test.key, err)
			}
			if got != test.want {
				t.Fatalf("catalogID(%q) = %q, want %q", test.key, got, test.want)
			}
		})
	}
}

func TestValidateEvolution(t *testing.T) {
	current := mustManifest(t, validExport())

	tests := []struct {
		name    string
		mutate  func(*Manifest)
		wantErr bool
	}{
		{
			name:   "unchanged",
			mutate: func(*Manifest) {},
		},
		{
			name: "addition",
			mutate: func(manifest *Manifest) {
				manifest.Assets = append(manifest.Assets, Asset{
					ID:          "ZRX",
					UpstreamKey: "ZRX",
					Address:     "0x1111111111111111111111111111111111111111",
				})
			},
		},
		{
			name: "removal",
			mutate: func(manifest *Manifest) {
				manifest.Assets = manifest.Assets[1:]
			},
			wantErr: true,
		},
		{
			name: "retargeted address",
			mutate: func(manifest *Manifest) {
				manifest.Assets[0].Address = "0x2222222222222222222222222222222222222222"
			},
			wantErr: true,
		},
		{
			name: "changed upstream key",
			mutate: func(manifest *Manifest) {
				manifest.Assets[0].UpstreamKey = "Usdc"
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			next := current
			next.Assets = append([]Asset(nil), current.Assets...)
			test.mutate(&next)
			err := ValidateEvolution(current, next)
			if test.wantErr {
				if !errors.Is(err, ErrUnsafeEvolution) {
					t.Fatalf("ValidateEvolution() error = %v, want ErrUnsafeEvolution", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("ValidateEvolution() error = %v", err)
			}
		})
	}
}

func mustManifest(t *testing.T, exported aaveaddressbook.Export) Manifest {
	t.Helper()
	encoded, err := json.Marshal(exported)
	if err != nil {
		t.Fatalf("encode export: %v", err)
	}
	generated, err := Generate(encoded)
	if err != nil {
		t.Fatalf("generate manifest: %v", err)
	}
	manifest, err := Parse(generated)
	if err != nil {
		t.Fatalf("parse generated manifest: %v", err)
	}
	return manifest
}

func validExport() aaveaddressbook.Export {
	return aaveaddressbook.Export{
		PackageName:    aaveaddressbook.OfficialPackage,
		PackageVersion: "4.60.0",
		GitHead:        "7e444a1e73b538fd0b9e093e5156401d6fccca7d",
		Export:         aaveaddressbook.BaseExportName,
		ChainID:        aaveaddressbook.BaseChainID,
		Contracts: aaveaddressbook.Contracts{
			PoolAddressesProvider:    "0xe20fCBdBfFC4Dd138cE8b2E6FBb6CB49777ad64D",
			Pool:                     "0xA238Dd80C259a72e81d7e4664a9801593F98d1c5",
			AaveProtocolDataProvider: "0x0F43731EB8d45A581f4a36DD74F5f358bc90C73A",
			WrappedTokenGateway:      "0xa0d9C1E9E48Ca30c8d8C3B5D69FF5dc1f6DFfC24",
		},
		Assets: []aaveaddressbook.Asset{
			{
				Key:          "USDC",
				Address:      "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
				IssuerSource: "https://developers.circle.com/stablecoins/usdc-contract-addresses",
			},
			{
				Key:     "WETH",
				Address: "0x4200000000000000000000000000000000000006",
			},
		},
	}
}
