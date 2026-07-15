package aaveassetmanifest

import (
	"bytes"
	"os"
	"testing"

	"github.com/tn606024/defi-simplify/internal/aaveaddressbook"
)

func TestGenerateMatchesCheckedInBaseManifest(t *testing.T) {
	exported, err := os.ReadFile("../aaveaddressbook/testdata/aave-v3-base-export.json")
	if err != nil {
		t.Fatalf("read export fixture: %v", err)
	}
	want, err := os.ReadFile("../../assets/base/manifest.json")
	if err != nil {
		t.Fatalf("read checked-in manifest: %v", err)
	}

	definition := aaveaddressbook.BaseV3ExportDefinition()
	first, err := Generate(exported, definition)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	second, err := Generate(exported, definition)
	if err != nil {
		t.Fatalf("Generate() second error = %v", err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("repeated generation produced different bytes")
	}
	if !bytes.Equal(first, want) {
		t.Fatalf("generated manifest differs from checked-in manifest\n%s", first)
	}
}

func TestDefinitionForPreservesChainAndScopesAssetExport(t *testing.T) {
	definition := DefinitionFor(aaveaddressbook.ExportDefinition{
		Name:    "AaveV3Ethereum",
		ChainID: 1,
	})
	if definition.ChainID != 1 {
		t.Fatalf("chain ID = %d, want 1", definition.ChainID)
	}
	if definition.Source.Export != "AaveV3Ethereum.ASSETS" {
		t.Fatalf("source export = %q, want AaveV3Ethereum.ASSETS", definition.Source.Export)
	}
}
