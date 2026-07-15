package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/tn606024/defi-simplify/internal/aavemanifest"
	"github.com/tn606024/defi-simplify/internal/assetmanifest"
)

const (
	defaultExtractor        = "tools/aave-address-book/export-base.mjs"
	defaultDeploymentOutput = "aave/manifests/aave-v3-base.json"
	defaultAssetOutput      = "assets/base/manifest.json"
)

func main() {
	extractor := flag.String("extractor", defaultExtractor, "Address Book Node extractor")
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
	flag.Parse()

	if err := run(context.Background(), *extractor, *deploymentOutput, *assetOutput); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, extractor string, deploymentOutput string, assetOutput string) error {
	command := exec.CommandContext(ctx, "node", extractor)
	exported, err := command.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("extract pinned Aave Address Book package: %w: %s", err, bytes.TrimSpace(exitErr.Stderr))
		}
		return fmt.Errorf("extract pinned Aave Address Book package: %w", err)
	}

	deploymentManifest, err := aavemanifest.Generate(exported)
	if err != nil {
		return fmt.Errorf("generate Aave deployment manifest: %w", err)
	}
	assetManifest, err := assetmanifest.Generate(exported)
	if err != nil {
		return fmt.Errorf("generate Base asset manifest: %w", err)
	}
	if err := validateAssetEvolution(assetOutput, assetManifest); err != nil {
		return err
	}
	if err := writeIfChanged(deploymentOutput, deploymentManifest); err != nil {
		return fmt.Errorf("write Aave deployment manifest: %w", err)
	}
	if err := writeIfChanged(assetOutput, assetManifest); err != nil {
		return fmt.Errorf("write Base asset manifest: %w", err)
	}
	return nil
}

func validateAssetEvolution(output string, nextData []byte) error {
	currentData, err := os.ReadFile(output)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read existing Base asset manifest: %w", err)
	}
	current, err := assetmanifest.Parse(currentData)
	if err != nil {
		return fmt.Errorf("parse existing Base asset manifest: %w", err)
	}
	next, err := assetmanifest.Parse(nextData)
	if err != nil {
		return fmt.Errorf("parse generated Base asset manifest: %w", err)
	}
	if err := assetmanifest.ValidateEvolution(current, next); err != nil {
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
		return fmt.Errorf("read existing Aave deployment manifest: %w", err)
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
