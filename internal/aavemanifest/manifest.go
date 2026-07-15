package aavemanifest

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/tn606024/defi-simplify/internal/aaveaddressbook"
)

const (
	SchemaVersion      = 1
	BaseMarketID       = "aave-v3-base"
	BaseExportName     = aaveaddressbook.BaseExportName
	BaseChainID        = aaveaddressbook.BaseChainID
	OfficialRepository = aaveaddressbook.OfficialRepository
	OfficialPackage    = aaveaddressbook.OfficialPackage
)

var (
	ErrInvalidExport   = aaveaddressbook.ErrInvalidExport
	ErrInvalidManifest = errors.New("invalid Aave deployment manifest")
)

type Contracts = aaveaddressbook.Contracts
type Source = aaveaddressbook.Source
type Export = aaveaddressbook.Export

// Manifest is the canonical machine-readable deployment manifest shape.
type Manifest struct {
	SchemaVersion int       `json:"schemaVersion"`
	MarketID      string    `json:"marketId"`
	ChainID       int       `json:"chainId"`
	Contracts     Contracts `json:"contracts"`
	Source        Source    `json:"source"`
}

// Generate validates one normalized Address Book export and renders the
// canonical checked-in manifest. Repeated generation from the same export is
// byte-for-byte deterministic.
func Generate(data []byte) ([]byte, error) {
	exported, err := aaveaddressbook.ParseExport(data)
	if err != nil {
		return nil, err
	}

	manifest := Manifest{
		SchemaVersion: SchemaVersion,
		MarketID:      BaseMarketID,
		ChainID:       exported.ChainID,
		Contracts:     aaveaddressbook.NormalizeContracts(exported.Contracts),
		Source:        aaveaddressbook.SourceFromExport(exported, BaseExportName),
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
	if err := aaveaddressbook.DecodeStrict(data, &manifest); err != nil {
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
	if err := aaveaddressbook.ValidateSource(manifest.Source, BaseExportName, ErrInvalidManifest); err != nil {
		return err
	}
	if err := aaveaddressbook.ValidateContracts(manifest.Contracts, ErrInvalidManifest); err != nil {
		return err
	}
	return nil
}
