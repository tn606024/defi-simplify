package aavemanifest

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

const (
	SchemaVersion      = 1
	BaseMarketID       = "aave-v3-base"
	BaseExportName     = "AaveV3Base"
	BaseChainID        = 8453
	OfficialRepository = "https://github.com/aave-dao/aave-address-book"
	OfficialPackage    = "@aave-dao/aave-address-book"
)

var (
	ErrInvalidExport   = errors.New("invalid Aave Address Book export")
	ErrInvalidManifest = errors.New("invalid Aave deployment manifest")
)

// Contracts contains the deployment anchors owned by this SDK. Dynamic
// reserve and token data deliberately does not belong in this structure.
type Contracts struct {
	PoolAddressesProvider    string `json:"poolAddressesProvider"`
	Pool                     string `json:"pool"`
	AaveProtocolDataProvider string `json:"aaveProtocolDataProvider"`
	WrappedTokenGateway      string `json:"wrappedTokenGateway,omitempty"`
}

// Source identifies the immutable upstream artifact used to build a manifest.
type Source struct {
	Repository     string `json:"repository"`
	Package        string `json:"package"`
	PackageVersion string `json:"packageVersion"`
	Release        string `json:"release"`
	Commit         string `json:"commit"`
	Export         string `json:"export"`
}

// Manifest is the canonical machine-readable deployment manifest shape.
type Manifest struct {
	SchemaVersion int       `json:"schemaVersion"`
	MarketID      string    `json:"marketId"`
	ChainID       int       `json:"chainId"`
	Contracts     Contracts `json:"contracts"`
	Source        Source    `json:"source"`
}

// Export is the normalized data emitted by the update-only Address Book
// extractor. It keeps JavaScript package loading outside the SDK runtime.
type Export struct {
	PackageName    string    `json:"packageName"`
	PackageVersion string    `json:"packageVersion"`
	GitHead        string    `json:"gitHead"`
	Export         string    `json:"export"`
	ChainID        int       `json:"chainId"`
	Contracts      Contracts `json:"contracts"`
}

// Generate validates one normalized Address Book export and renders the
// canonical checked-in manifest. Repeated generation from the same export is
// byte-for-byte deterministic.
func Generate(data []byte) ([]byte, error) {
	var exported Export
	if err := decodeStrict(data, &exported); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidExport, err)
	}
	if err := validateExport(exported); err != nil {
		return nil, err
	}

	manifest := Manifest{
		SchemaVersion: SchemaVersion,
		MarketID:      BaseMarketID,
		ChainID:       exported.ChainID,
		Contracts:     normalizeContracts(exported.Contracts),
		Source: Source{
			Repository:     OfficialRepository,
			Package:        exported.PackageName,
			PackageVersion: exported.PackageVersion,
			Release:        "v" + exported.PackageVersion,
			Commit:         strings.ToLower(exported.GitHead),
			Export:         exported.Export,
		},
	}
	if err := Validate(manifest); err != nil {
		return nil, err
	}

	encoded, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode Aave deployment manifest: %w", err)
	}
	return append(encoded, '\n'), nil
}

// Parse decodes and validates one deployment manifest. Unknown fields and
// trailing JSON are rejected so reviewed manifests cannot contain ignored
// configuration.
func Parse(data []byte) (Manifest, error) {
	var manifest Manifest
	if err := decodeStrict(data, &manifest); err != nil {
		return Manifest{}, fmt.Errorf("%w: %v", ErrInvalidManifest, err)
	}
	if err := Validate(manifest); err != nil {
		return Manifest{}, err
	}
	return manifest, nil
}

// Validate enforces the supported Base V3 trust anchor and source identity.
func Validate(manifest Manifest) error {
	if manifest.SchemaVersion != SchemaVersion {
		return fmt.Errorf(
			"%w: unsupported schema version %d",
			ErrInvalidManifest,
			manifest.SchemaVersion,
		)
	}
	if manifest.MarketID != BaseMarketID {
		return fmt.Errorf("%w: unsupported market %q", ErrInvalidManifest, manifest.MarketID)
	}
	if manifest.ChainID != BaseChainID {
		return fmt.Errorf("%w: market %q has chain ID %d", ErrInvalidManifest, manifest.MarketID, manifest.ChainID)
	}
	if manifest.Source.Repository != OfficialRepository {
		return fmt.Errorf("%w: unsupported source repository %q", ErrInvalidManifest, manifest.Source.Repository)
	}
	if manifest.Source.Package != OfficialPackage {
		return fmt.Errorf("%w: unsupported source package %q", ErrInvalidManifest, manifest.Source.Package)
	}
	if strings.TrimSpace(manifest.Source.PackageVersion) == "" {
		return fmt.Errorf("%w: source package version is empty", ErrInvalidManifest)
	}
	if manifest.Source.Release != "v"+manifest.Source.PackageVersion {
		return fmt.Errorf(
			"%w: release %q does not match package version %q",
			ErrInvalidManifest,
			manifest.Source.Release,
			manifest.Source.PackageVersion,
		)
	}
	if !validCommit(manifest.Source.Commit) {
		return fmt.Errorf("%w: source commit %q is not a 40-character Git commit", ErrInvalidManifest, manifest.Source.Commit)
	}
	if manifest.Source.Export != BaseExportName {
		return fmt.Errorf("%w: unsupported source export %q", ErrInvalidManifest, manifest.Source.Export)
	}
	if err := validateContracts(manifest.Contracts, ErrInvalidManifest); err != nil {
		return err
	}
	return nil
}

func validateExport(exported Export) error {
	if exported.PackageName != OfficialPackage {
		return fmt.Errorf("%w: unsupported package %q", ErrInvalidExport, exported.PackageName)
	}
	if strings.TrimSpace(exported.PackageVersion) == "" {
		return fmt.Errorf("%w: package version is empty", ErrInvalidExport)
	}
	if !validCommit(exported.GitHead) {
		return fmt.Errorf("%w: gitHead %q is not a 40-character Git commit", ErrInvalidExport, exported.GitHead)
	}
	if exported.Export != BaseExportName {
		return fmt.Errorf("%w: unsupported export %q", ErrInvalidExport, exported.Export)
	}
	if exported.ChainID != BaseChainID {
		return fmt.Errorf("%w: export has chain ID %d", ErrInvalidExport, exported.ChainID)
	}
	return validateContracts(exported.Contracts, ErrInvalidExport)
}

func validateContracts(contracts Contracts, sentinel error) error {
	addresses := []struct {
		name     string
		value    string
		optional bool
	}{
		{name: "pool addresses provider", value: contracts.PoolAddressesProvider},
		{name: "pool", value: contracts.Pool},
		{name: "Aave protocol data provider", value: contracts.AaveProtocolDataProvider},
		{name: "wrapped token gateway", value: contracts.WrappedTokenGateway, optional: true},
	}
	seen := make(map[common.Address]string, len(addresses))
	for _, field := range addresses {
		if field.optional && field.value == "" {
			continue
		}
		if !common.IsHexAddress(field.value) {
			return fmt.Errorf("%w: %s address %q is invalid", sentinel, field.name, field.value)
		}
		address := common.HexToAddress(field.value)
		if address == (common.Address{}) {
			return fmt.Errorf("%w: %s address is zero", sentinel, field.name)
		}
		if previous, ok := seen[address]; ok {
			return fmt.Errorf(
				"%w: %s and %s use the same address %s",
				sentinel,
				previous,
				field.name,
				address.Hex(),
			)
		}
		seen[address] = field.name
	}
	return nil
}

func normalizeContracts(contracts Contracts) Contracts {
	return Contracts{
		PoolAddressesProvider:    common.HexToAddress(contracts.PoolAddressesProvider).Hex(),
		Pool:                     common.HexToAddress(contracts.Pool).Hex(),
		AaveProtocolDataProvider: common.HexToAddress(contracts.AaveProtocolDataProvider).Hex(),
		WrappedTokenGateway:      normalizeOptionalAddress(contracts.WrappedTokenGateway),
	}
}

func normalizeOptionalAddress(value string) string {
	if value == "" {
		return ""
	}
	return common.HexToAddress(value).Hex()
}

func validCommit(value string) bool {
	if len(value) != 40 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func decodeStrict(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("unexpected trailing JSON value")
		}
		return fmt.Errorf("decode trailing JSON: %w", err)
	}
	return nil
}
