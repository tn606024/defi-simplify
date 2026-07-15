package aaveaddressbook

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
	BaseExportName     = "AaveV3Base"
	BaseChainID        = 8453
	OfficialRepository = "https://github.com/aave-dao/aave-address-book"
	OfficialPackage    = "@aave-dao/aave-address-book"
)

var ErrInvalidExport = errors.New("invalid Aave Address Book export")

// Contracts contains the deployment anchors exported for the Base market.
type Contracts struct {
	PoolAddressesProvider    string `json:"poolAddressesProvider"`
	Pool                     string `json:"pool"`
	AaveProtocolDataProvider string `json:"aaveProtocolDataProvider"`
	WrappedTokenGateway      string `json:"wrappedTokenGateway,omitempty"`
}

// Asset contains one normalized underlying identity from AaveV3Base.ASSETS.
// Display metadata and protocol token roles deliberately stay out of this
// update-only export.
type Asset struct {
	Key          string `json:"key"`
	Address      string `json:"address"`
	IssuerSource string `json:"issuerSource,omitempty"`
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

// Export is the normalized data emitted by the update-only Node extractor.
// JavaScript package loading never occurs in the SDK runtime.
type Export struct {
	PackageName    string    `json:"packageName"`
	PackageVersion string    `json:"packageVersion"`
	GitHead        string    `json:"gitHead"`
	Export         string    `json:"export"`
	ChainID        int       `json:"chainId"`
	Contracts      Contracts `json:"contracts"`
	Assets         []Asset   `json:"assets,omitempty"`
}

// ExportDefinition identifies one chain-specific export from the official
// Address Book package. Callers own the supported market list; the parser owns
// source and shape validation.
type ExportDefinition struct {
	Name    string
	ChainID int
}

// BaseV3ExportDefinition returns the supported Base V3 export identity.
func BaseV3ExportDefinition() ExportDefinition {
	return ExportDefinition{Name: BaseExportName, ChainID: BaseChainID}
}

// ParseExport strictly decodes and validates the pinned Address Book artifact.
// It preserves the existing Base V3 behavior for deployment-manifest callers.
func ParseExport(data []byte) (Export, error) {
	return ParseExportFor(data, BaseV3ExportDefinition())
}

// ParseExportFor strictly decodes and validates a pinned Address Book artifact
// against the caller-owned chain/export definition. Asset-specific manifest
// rules are enforced by the owning source adapter.
func ParseExportFor(data []byte, definition ExportDefinition) (Export, error) {
	var exported Export
	if err := DecodeStrict(data, &exported); err != nil {
		return Export{}, fmt.Errorf("%w: %v", ErrInvalidExport, err)
	}
	if err := validateExportDefinition(definition); err != nil {
		return Export{}, err
	}
	if err := validateExport(exported, definition); err != nil {
		return Export{}, err
	}
	return exported, nil
}

// SourceFromExport creates provenance for one reviewed sub-export.
func SourceFromExport(exported Export, exportName string) Source {
	return Source{
		Repository:     OfficialRepository,
		Package:        exported.PackageName,
		PackageVersion: exported.PackageVersion,
		Release:        "v" + exported.PackageVersion,
		Commit:         strings.ToLower(exported.GitHead),
		Export:         exportName,
	}
}

// ValidateSource verifies a checked-in manifest's pinned source identity.
func ValidateSource(source Source, expectedExport string, sentinel error) error {
	if source.Repository != OfficialRepository {
		return fmt.Errorf("%w: unsupported source repository %q", sentinel, source.Repository)
	}
	if source.Package != OfficialPackage {
		return fmt.Errorf("%w: unsupported source package %q", sentinel, source.Package)
	}
	if strings.TrimSpace(source.PackageVersion) == "" {
		return fmt.Errorf("%w: source package version is empty", sentinel)
	}
	if source.Release != "v"+source.PackageVersion {
		return fmt.Errorf(
			"%w: release %q does not match package version %q",
			sentinel,
			source.Release,
			source.PackageVersion,
		)
	}
	if !validCommit(source.Commit) {
		return fmt.Errorf(
			"%w: source commit %q is not a 40-character Git commit",
			sentinel,
			source.Commit,
		)
	}
	if source.Export != expectedExport {
		return fmt.Errorf("%w: unsupported source export %q", sentinel, source.Export)
	}
	return nil
}

// ValidateContracts verifies required addresses and rejects duplicate roles.
func ValidateContracts(contracts Contracts, sentinel error) error {
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

// NormalizeContracts returns checksummed deployment addresses.
func NormalizeContracts(contracts Contracts) Contracts {
	return Contracts{
		PoolAddressesProvider:    common.HexToAddress(contracts.PoolAddressesProvider).Hex(),
		Pool:                     common.HexToAddress(contracts.Pool).Hex(),
		AaveProtocolDataProvider: common.HexToAddress(contracts.AaveProtocolDataProvider).Hex(),
		WrappedTokenGateway:      normalizeOptionalAddress(contracts.WrappedTokenGateway),
	}
}

// DecodeStrict rejects unknown fields and trailing JSON values.
func DecodeStrict(data []byte, target any) error {
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

func validateExport(exported Export, definition ExportDefinition) error {
	if exported.PackageName != OfficialPackage {
		return fmt.Errorf("%w: unsupported package %q", ErrInvalidExport, exported.PackageName)
	}
	if strings.TrimSpace(exported.PackageVersion) == "" {
		return fmt.Errorf("%w: package version is empty", ErrInvalidExport)
	}
	if !validCommit(exported.GitHead) {
		return fmt.Errorf(
			"%w: gitHead %q is not a 40-character Git commit",
			ErrInvalidExport,
			exported.GitHead,
		)
	}
	if exported.Export != definition.Name {
		return fmt.Errorf("%w: unsupported export %q", ErrInvalidExport, exported.Export)
	}
	if exported.ChainID != definition.ChainID {
		return fmt.Errorf("%w: export has chain ID %d", ErrInvalidExport, exported.ChainID)
	}
	return ValidateContracts(exported.Contracts, ErrInvalidExport)
}

func validateExportDefinition(definition ExportDefinition) error {
	if strings.TrimSpace(definition.Name) == "" {
		return fmt.Errorf("%w: expected export name is empty", ErrInvalidExport)
	}
	if definition.ChainID <= 0 {
		return fmt.Errorf("%w: expected chain ID %d is invalid", ErrInvalidExport, definition.ChainID)
	}
	return nil
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
