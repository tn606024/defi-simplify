package assetmanifest

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestGenerateSupportsChainOwnedDefinitions(t *testing.T) {
	tests := []struct {
		name       string
		definition Definition
	}{
		{name: "Base", definition: testDefinition(8453, "AaveV3Base.ASSETS")},
		{name: "Ethereum", definition: testDefinition(1, "AaveV3Ethereum.ASSETS")},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			source := testSource(test.definition)
			first, err := Generate(test.definition, source, validCandidates())
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}
			second, err := Generate(test.definition, source, validCandidates())
			if err != nil {
				t.Fatalf("Generate() second error = %v", err)
			}
			if string(first) != string(second) {
				t.Fatal("repeated generation produced different bytes")
			}
			manifest, err := Parse(first, test.definition)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			if manifest.ChainID != test.definition.ChainID {
				t.Fatalf("manifest chain ID = %d, want %d", manifest.ChainID, test.definition.ChainID)
			}
			if manifest.Source.Export != test.definition.Source.Export {
				t.Fatalf("manifest source export = %q, want %q", manifest.Source.Export, test.definition.Source.Export)
			}
		})
	}
}

func TestGenerateRejectsInvalidCandidates(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*[]Candidate)
	}{
		{name: "empty", mutate: func(candidates *[]Candidate) { *candidates = nil }},
		{name: "unsupported key character", mutate: func(candidates *[]Candidate) {
			(*candidates)[0].Key = "USD.C"
		}},
		{name: "zero address", mutate: func(candidates *[]Candidate) {
			(*candidates)[0].Address = "0x0000000000000000000000000000000000000000"
		}},
		{name: "malformed address", mutate: func(candidates *[]Candidate) {
			(*candidates)[0].Address = "not-an-address"
		}},
		{name: "duplicate canonical ID", mutate: func(candidates *[]Candidate) {
			(*candidates)[1].Key = "usdc"
		}},
		{name: "duplicate address", mutate: func(candidates *[]Candidate) {
			(*candidates)[1].Address = (*candidates)[0].Address
		}},
		{name: "non-HTTPS issuer source", mutate: func(candidates *[]Candidate) {
			(*candidates)[0].IssuerSource = "http://example.com/token"
		}},
	}
	definition := testDefinition(1, "ExampleEthereumAssets")
	source := testSource(definition)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			candidates := validCandidates()
			test.mutate(&candidates)
			_, err := Generate(definition, source, candidates)
			if !errors.Is(err, ErrInvalidManifest) {
				t.Fatalf("Generate() error = %v, want ErrInvalidManifest", err)
			}
		})
	}
}

func TestParseRejectsDefinitionAndShapeMismatch(t *testing.T) {
	definition := testDefinition(1, "ExampleEthereumAssets")
	valid, err := Generate(definition, testSource(definition), validCandidates())
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	tests := []struct {
		name       string
		data       []byte
		definition Definition
	}{
		{
			name: "unknown field",
			data: []byte(strings.Replace(
				string(valid),
				`"chainId": 1`,
				`"chainId": 1, "unreviewed": true`,
				1,
			)),
			definition: definition,
		},
		{
			name:       "wrong chain",
			data:       valid,
			definition: testDefinition(8453, definition.Source.Export),
		},
		{
			name: "wrong source export",
			data: []byte(strings.Replace(
				string(valid),
				definition.Source.Export,
				"AnotherExport",
				1,
			)),
			definition: definition,
		},
		{
			name:       "trailing JSON",
			data:       append(append([]byte(nil), valid...), []byte("{}")...),
			definition: definition,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Parse(test.data, test.definition)
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
	definition := testDefinition(1, "ExampleEthereumAssets")
	current := mustManifest(t, definition)

	tests := []struct {
		name    string
		mutate  func(*Manifest)
		wantErr bool
	}{
		{name: "unchanged", mutate: func(*Manifest) {}},
		{name: "addition", mutate: func(manifest *Manifest) {
			manifest.Assets = append(manifest.Assets, Asset{
				ID: "ZRX", UpstreamKey: "ZRX", Address: "0x1111111111111111111111111111111111111111",
			})
		}},
		{name: "removal", mutate: func(manifest *Manifest) {
			manifest.Assets = manifest.Assets[1:]
		}, wantErr: true},
		{name: "retargeted address", mutate: func(manifest *Manifest) {
			manifest.Assets[0].Address = "0x2222222222222222222222222222222222222222"
		}, wantErr: true},
		{name: "changed upstream key", mutate: func(manifest *Manifest) {
			manifest.Assets[0].UpstreamKey = "Usdc"
		}, wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			next := current
			next.Assets = append([]Asset(nil), current.Assets...)
			test.mutate(&next)
			err := ValidateEvolution(current, next, definition)
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

func mustManifest(t *testing.T, definition Definition) Manifest {
	t.Helper()
	generated, err := Generate(definition, testSource(definition), validCandidates())
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	var manifest Manifest
	if err := json.Unmarshal(generated, &manifest); err != nil {
		t.Fatalf("decode manifest: %v", err)
	}
	return manifest
}

func testDefinition(chainID int, export string) Definition {
	return Definition{
		ChainID: chainID,
		Source: SourceDefinition{
			Repository: "https://github.com/example/assets",
			Package:    "@example/assets",
			Export:     export,
		},
	}
}

func testSource(definition Definition) Source {
	return Source{
		Repository:     definition.Source.Repository,
		Package:        definition.Source.Package,
		PackageVersion: "1.2.3",
		Release:        "v1.2.3",
		Commit:         "7e444a1e73b538fd0b9e093e5156401d6fccca7d",
		Export:         definition.Source.Export,
	}
}

func validCandidates() []Candidate {
	return []Candidate{
		{
			Key:          "USDC",
			Address:      "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
			IssuerSource: "https://developers.circle.com/stablecoins/usdc-contract-addresses",
		},
		{Key: "WETH", Address: "0x4200000000000000000000000000000000000006"},
	}
}
