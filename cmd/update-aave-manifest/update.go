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
)

const (
	defaultExtractor = "tools/aave-address-book/export-base.mjs"
	defaultOutput    = "aave/manifests/aave-v3-base.json"
)

func main() {
	extractor := flag.String("extractor", defaultExtractor, "Address Book Node extractor")
	output := flag.String("output", defaultOutput, "checked-in manifest output")
	flag.Parse()

	if err := run(context.Background(), *extractor, *output); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, extractor string, output string) error {
	command := exec.CommandContext(ctx, "node", extractor)
	exported, err := command.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("extract pinned Aave Address Book package: %w: %s", err, bytes.TrimSpace(exitErr.Stderr))
		}
		return fmt.Errorf("extract pinned Aave Address Book package: %w", err)
	}

	manifest, err := aavemanifest.Generate(exported)
	if err != nil {
		return fmt.Errorf("generate Aave deployment manifest: %w", err)
	}
	current, err := os.ReadFile(output)
	if err == nil && bytes.Equal(current, manifest) {
		return nil
	}
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read existing Aave deployment manifest: %w", err)
	}
	if err := writeFileAtomically(output, manifest); err != nil {
		return fmt.Errorf("write Aave deployment manifest: %w", err)
	}
	return nil
}

func writeFileAtomically(path string, data []byte) error {
	directory := filepath.Dir(path)
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
