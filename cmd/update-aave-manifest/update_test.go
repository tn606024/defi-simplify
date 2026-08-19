package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestUpdateFromExportWritesAllReviewedOutputs(t *testing.T) {
	exported, err := os.ReadFile("../../internal/aaveaddressbook/testdata/aave-v3-base-export.json")
	if err != nil {
		t.Fatalf("read export fixture: %v", err)
	}

	directory := t.TempDir()
	exportOutput := filepath.Join(directory, "export.json")
	deploymentOutput := filepath.Join(directory, "deployment.json")
	assetOutput := filepath.Join(directory, "assets.json")
	assetGoOutput := filepath.Join(directory, "catalog_gen.go")

	err = updateFromExport(
		exported,
		exportOutput,
		deploymentOutput,
		assetOutput,
		assetGoOutput,
		"base",
	)
	if err != nil {
		t.Fatalf("updateFromExport() error = %v", err)
	}

	writtenExport, err := os.ReadFile(exportOutput)
	if err != nil {
		t.Fatalf("read written export: %v", err)
	}
	if !bytes.Equal(writtenExport, exported) {
		t.Fatalf("written export differs from normalized input\n%s", writtenExport)
	}
	for _, path := range []string{deploymentOutput, assetOutput, assetGoOutput} {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read reviewed output %s: %v", path, err)
		}
		if len(data) == 0 {
			t.Fatalf("reviewed output %s is empty", path)
		}
	}
}
