package assetmanifest

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/tn606024/defi-simplify/internal/aaveaddressbook"
)

const (
	SchemaVersion        = 1
	BaseAssetsExportName = "AaveV3Base.ASSETS"
)

var (
	ErrInvalidManifest = errors.New("invalid Base asset manifest")
	ErrUnsafeEvolution = errors.New("unsafe Base asset catalog evolution")
)

// Asset is one reviewed catalog identity. ID is an SDK-owned catalog key, not
// an ERC20 symbol and not a protocol capability declaration.
type Asset struct {
	ID           string `json:"id"`
	UpstreamKey  string `json:"upstreamKey"`
	Address      string `json:"address"`
	IssuerSource string `json:"issuerSource,omitempty"`
}

// Manifest is the canonical machine-readable Base asset catalog.
type Manifest struct {
	SchemaVersion int                    `json:"schemaVersion"`
	ChainID       int                    `json:"chainId"`
	Source        aaveaddressbook.Source `json:"source"`
	Assets        []Asset                `json:"assets"`
}

// Generate renders a deterministic catalog from the pinned Address Book
// export. It copies only underlying identities and optional reviewed issuer
// provenance; live metadata and Aave token roles are intentionally excluded.
func Generate(data []byte) ([]byte, error) {
	exported, err := aaveaddressbook.ParseExport(data)
	if err != nil {
		return nil, err
	}
	assets, err := normalizeExportedAssets(exported.Assets)
	if err != nil {
		return nil, err
	}
	manifest := Manifest{
		SchemaVersion: SchemaVersion,
		ChainID:       exported.ChainID,
		Source:        aaveaddressbook.SourceFromExport(exported, BaseAssetsExportName),
		Assets:        assets,
	}
	if err := Validate(manifest); err != nil {
		return nil, err
	}
	encoded, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode Base asset manifest: %w", err)
	}
	return append(encoded, '\n'), nil
}

// Parse strictly decodes and validates one checked-in asset manifest.
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

// Validate verifies source provenance, canonical ordering, and unique token
// identities. The manifest is Base-specific by design.
func Validate(manifest Manifest) error {
	if manifest.SchemaVersion != SchemaVersion {
		return fmt.Errorf(
			"%w: unsupported schema version %d",
			ErrInvalidManifest,
			manifest.SchemaVersion,
		)
	}
	if manifest.ChainID != aaveaddressbook.BaseChainID {
		return fmt.Errorf("%w: unsupported chain ID %d", ErrInvalidManifest, manifest.ChainID)
	}
	if err := aaveaddressbook.ValidateSource(
		manifest.Source,
		BaseAssetsExportName,
		ErrInvalidManifest,
	); err != nil {
		return err
	}
	if len(manifest.Assets) == 0 {
		return fmt.Errorf("%w: asset list is empty", ErrInvalidManifest)
	}

	seenIDs := make(map[string]struct{}, len(manifest.Assets))
	seenKeys := make(map[string]struct{}, len(manifest.Assets))
	seenAddresses := make(map[common.Address]string, len(manifest.Assets))
	previousID := ""
	for i, asset := range manifest.Assets {
		if err := validateAsset(asset); err != nil {
			return fmt.Errorf("%w: asset %d: %v", ErrInvalidManifest, i+1, err)
		}
		if i > 0 && asset.ID <= previousID {
			return fmt.Errorf(
				"%w: assets are not strictly ordered by ID at %q",
				ErrInvalidManifest,
				asset.ID,
			)
		}
		previousID = asset.ID
		if _, ok := seenIDs[asset.ID]; ok {
			return fmt.Errorf("%w: duplicate asset ID %q", ErrInvalidManifest, asset.ID)
		}
		seenIDs[asset.ID] = struct{}{}
		if _, ok := seenKeys[asset.UpstreamKey]; ok {
			return fmt.Errorf("%w: duplicate upstream key %q", ErrInvalidManifest, asset.UpstreamKey)
		}
		seenKeys[asset.UpstreamKey] = struct{}{}
		address := common.HexToAddress(asset.Address)
		if previous, ok := seenAddresses[address]; ok {
			return fmt.Errorf(
				"%w: assets %q and %q use the same address %s",
				ErrInvalidManifest,
				previous,
				asset.ID,
				address.Hex(),
			)
		}
		seenAddresses[address] = asset.ID
	}
	return nil
}

// ValidateEvolution permits reviewed additions while preventing routine
// automation from removing or retargeting an existing public catalog ID.
// Those changes require an explicit migration and deprecation decision.
func ValidateEvolution(current Manifest, next Manifest) error {
	if err := Validate(current); err != nil {
		return err
	}
	if err := Validate(next); err != nil {
		return err
	}
	nextByID := make(map[string]Asset, len(next.Assets))
	for _, asset := range next.Assets {
		nextByID[asset.ID] = asset
	}
	for _, asset := range current.Assets {
		updated, ok := nextByID[asset.ID]
		if !ok {
			return fmt.Errorf("%w: existing asset ID %q was removed", ErrUnsafeEvolution, asset.ID)
		}
		if updated.Address != asset.Address {
			return fmt.Errorf(
				"%w: existing asset ID %q changed address from %s to %s",
				ErrUnsafeEvolution,
				asset.ID,
				asset.Address,
				updated.Address,
			)
		}
		if updated.UpstreamKey != asset.UpstreamKey {
			return fmt.Errorf(
				"%w: existing asset ID %q changed upstream key from %q to %q",
				ErrUnsafeEvolution,
				asset.ID,
				asset.UpstreamKey,
				updated.UpstreamKey,
			)
		}
	}
	return nil
}

func normalizeExportedAssets(exported []aaveaddressbook.Asset) ([]Asset, error) {
	if len(exported) == 0 {
		return nil, fmt.Errorf("%w: Address Book export contains no assets", ErrInvalidManifest)
	}
	assets := make([]Asset, 0, len(exported))
	for i, upstream := range exported {
		id, err := catalogID(upstream.Key)
		if err != nil {
			return nil, fmt.Errorf(
				"%w: exported asset %d key %q: %v",
				ErrInvalidManifest,
				i+1,
				upstream.Key,
				err,
			)
		}
		asset := Asset{
			ID:           id,
			UpstreamKey:  upstream.Key,
			Address:      common.HexToAddress(upstream.Address).Hex(),
			IssuerSource: upstream.IssuerSource,
		}
		if err := validateAsset(asset); err != nil {
			return nil, fmt.Errorf("%w: exported asset %q: %v", ErrInvalidManifest, upstream.Key, err)
		}
		assets = append(assets, asset)
	}
	sort.Slice(assets, func(i, j int) bool { return assets[i].ID < assets[j].ID })
	return assets, nil
}

func validateAsset(asset Asset) error {
	wantID, err := catalogID(asset.UpstreamKey)
	if err != nil {
		return fmt.Errorf("invalid upstream key %q: %v", asset.UpstreamKey, err)
	}
	if asset.ID != wantID {
		return fmt.Errorf("ID %q does not match canonical ID %q", asset.ID, wantID)
	}
	if !common.IsHexAddress(asset.Address) {
		return fmt.Errorf("address %q is invalid", asset.Address)
	}
	if common.HexToAddress(asset.Address) == (common.Address{}) {
		return errors.New("address is zero")
	}
	if asset.IssuerSource != "" {
		parsed, err := url.ParseRequestURI(asset.IssuerSource)
		if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
			return fmt.Errorf("issuer source %q is not an absolute HTTPS URL", asset.IssuerSource)
		}
	}
	return nil
}

func catalogID(key string) (string, error) {
	if key == "" {
		return "", errors.New("key is empty")
	}
	var id strings.Builder
	id.Grow(len(key))
	for _, character := range key {
		switch {
		case character >= 'a' && character <= 'z':
			id.WriteRune(character - ('a' - 'A'))
		case character >= 'A' && character <= 'Z':
			id.WriteRune(character)
		case character >= '0' && character <= '9':
			id.WriteRune(character)
		default:
			return "", fmt.Errorf("key contains unsupported character %q", character)
		}
	}
	value := id.String()
	if value[0] >= '0' && value[0] <= '9' {
		return "", errors.New("key starts with a digit")
	}
	return value, nil
}
