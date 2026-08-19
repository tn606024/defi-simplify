package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/tn606024/defi-simplify/internal/aaveaddressbook"
	"github.com/tn606024/defi-simplify/internal/aaveassetmanifest"
	"github.com/tn606024/defi-simplify/internal/aavemanifest"
	"github.com/tn606024/defi-simplify/internal/assetmanifest"
	"github.com/tn606024/defi-simplify/internal/catalogcodegen"
)

const (
	defaultExtractor        = "tools/aave-address-book/export-base.mjs"
	defaultExportOutput     = "internal/aaveaddressbook/testdata/aave-v3-base-export.json"
	defaultDeploymentOutput = "aave/manifests/aave-v3-base.json"
	defaultAssetOutput      = "assets/base/manifest.json"
	defaultAssetGoOutput    = "assets/base/catalog_gen.go"
	defaultAssetGoPackage   = "base"
)

func main() {
	extractor := flag.String("extractor", defaultExtractor, "Address Book Node extractor")
	exportOutput := flag.String(
		"export-output",
		defaultExportOutput,
		"checked-in normalized Address Book export output",
	)
	deploymentOutput := flag.String(
		"deployment-output",
		defaultDeploymentOutput,
		"checked-in Aave deployment manifest output",
	)
	assetOutput := flag.String(
		"asset-output",
		defaultAssetOutput,
		"checked-in Base asset manifest output",
	)
	assetGoOutput := flag.String(
		"asset-go-output",
		defaultAssetGoOutput,
		"generated Base asset Go declarations output",
	)
	assetGoPackage := flag.String(
		"asset-go-package",
		defaultAssetGoPackage,
		"Go package for generated asset declarations",
	)
	flag.Parse()

	if err := run(
		context.Background(),
		*extractor,
		*exportOutput,
		*deploymentOutput,
		*assetOutput,
		*assetGoOutput,
		*assetGoPackage,
	); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(
	ctx context.Context,
	extractor string,
	exportOutput string,
	deploymentOutput string,
	assetOutput string,
	assetGoOutput string,
	assetGoPackage string,
) error {
	command := exec.CommandContext(ctx, "node", extractor)
	exported, err := command.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("extract pinned Aave Address Book package: %w: %s", err, bytes.TrimSpace(exitErr.Stderr))
		}
		return fmt.Errorf("extract pinned Aave Address Book package: %w", err)
	}
	return updateFromExport(
		exported,
		exportOutput,
		deploymentOutput,
		assetOutput,
		assetGoOutput,
		assetGoPackage,
	)
}

func updateFromExport(
	exported []byte,
	exportOutput string,
	deploymentOutput string,
	assetOutput string,
	assetGoOutput string,
	assetGoPackage string,
) error {
	reviewedExport, err := normalizeReviewedExport(exported)
	if err != nil {
		return err
	}
	deploymentManifest, err := aavemanifest.Generate(exported)
	if err != nil {
		return fmt.Errorf("generate Aave deployment manifest: %w", err)
	}
	assetDefinition := aaveassetmanifest.DefinitionFor(aaveaddressbook.BaseV3ExportDefinition())
	assetManifest, err := aaveassetmanifest.Generate(exported, aaveaddressbook.BaseV3ExportDefinition())
	if err != nil {
		return fmt.Errorf("generate Base asset manifest: %w", err)
	}
	if err := validateAssetEvolution(assetOutput, assetManifest, assetDefinition); err != nil {
		return err
	}
	parsedAssetManifest, err := assetmanifest.Parse(assetManifest, assetDefinition)
	if err != nil {
		return fmt.Errorf("parse generated Base asset manifest for Go generation: %w", err)
	}
	assetGoSource, err := catalogcodegen.Generate(assetGoPackage, parsedAssetManifest.Assets)
	if err != nil {
		return fmt.Errorf("generate Base asset Go declarations: %w", err)
	}
	if err := writeIfChanged(deploymentOutput, deploymentManifest); err != nil {
		return fmt.Errorf("write Aave deployment manifest: %w", err)
	}
	if err := writeIfChanged(assetOutput, assetManifest); err != nil {
		return fmt.Errorf("write Base asset manifest: %w", err)
	}
	if err := writeIfChanged(assetGoOutput, assetGoSource); err != nil {
		return fmt.Errorf("write Base asset Go declarations: %w", err)
	}
	if err := writeIfChanged(exportOutput, reviewedExport); err != nil {
		return fmt.Errorf("write normalized Address Book export: %w", err)
	}
	return nil
}

func normalizeReviewedExport(data []byte) ([]byte, error) {
	exported, err := aaveaddressbook.ParseExportFor(data, aaveaddressbook.BaseV3ExportDefinition())
	if err != nil {
		return nil, fmt.Errorf("validate normalized Address Book export: %w", err)
	}
	encoded, err := json.MarshalIndent(exported, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode normalized Address Book export: %w", err)
	}
	return append(encoded, '\n'), nil
}

func validateAssetEvolution(
	output string,
	nextData []byte,
	definition assetmanifest.Definition,
) error {
	currentData, err := os.ReadFile(output)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read existing Base asset manifest: %w", err)
	}
	current, err := assetmanifest.Parse(currentData, definition)
	if err != nil {
		return fmt.Errorf("parse existing Base asset manifest: %w", err)
	}
	next, err := assetmanifest.Parse(nextData, definition)
	if err != nil {
		return fmt.Errorf("parse generated Base asset manifest: %w", err)
	}
	if err := assetmanifest.ValidateEvolution(current, next, definition); err != nil {
		return fmt.Errorf("validate Base asset catalog evolution: %w", err)
	}
	return nil
}

func writeIfChanged(output string, manifest []byte) error {
	current, err := os.ReadFile(output)
	if err == nil && bytes.Equal(current, manifest) {
		return nil
	}
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read existing generated output: %w", err)
	}
	return writeFileAtomically(output, manifest)
}

func writeFileAtomically(path string, data []byte) error {
	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return err
	}
	temporary, err := os.CreateTemp(directory, ".aave-manifest-*")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)

	if _, err := temporary.Write(data); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Chmod(0o644); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return os.Rename(temporaryPath, path)
}
