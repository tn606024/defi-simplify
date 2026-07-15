package assetmanifest

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"sort"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

const SchemaVersion = 1

var (
	ErrInvalidManifest = errors.New("invalid asset manifest")
	ErrUnsafeEvolution = errors.New("unsafe asset catalog evolution")
)

// SourceDefinition pins the expected upstream identity for one catalog. It is
// supplied by a source adapter and is not serialized into SDK configuration.
type SourceDefinition struct {
	Repository string
	Package    string
	Export     string
}

// Definition identifies one chain-scoped catalog and its reviewed upstream.
// The neutral manifest package does not own a list of supported chains.
type Definition struct {
	ChainID int
	Source  SourceDefinition
}

// Source records the immutable upstream artifact used to build a manifest.
type Source struct {
	Repository     string `json:"repository"`
	Package        string `json:"package"`
	PackageVersion string `json:"packageVersion"`
	Release        string `json:"release"`
	Commit         string `json:"commit"`
	Export         string `json:"export"`
}

// Candidate is one normalized underlying identity supplied by a source
// adapter. Provider-specific export parsing stays outside this package.
type Candidate struct {
	Key          string
	Address      string
	IssuerSource string
}

// Asset is one reviewed catalog identity. ID is an SDK-owned catalog key, not
// an ERC20 symbol and not a protocol capability declaration.
type Asset struct {
	ID           string `json:"id"`
	UpstreamKey  string `json:"upstreamKey"`
	Address      string `json:"address"`
	IssuerSource string `json:"issuerSource,omitempty"`
}

// Manifest is the canonical machine-readable chain-scoped asset catalog.
type Manifest struct {
	SchemaVersion int     `json:"schemaVersion"`
	ChainID       int     `json:"chainId"`
	Source        Source  `json:"source"`
	Assets        []Asset `json:"assets"`
}

// Generate renders a deterministic catalog from normalized candidates. It
// copies only underlying identities and optional reviewed issuer provenance;
// live metadata and protocol token roles are intentionally excluded.
func Generate(definition Definition, source Source, candidates []Candidate) ([]byte, error) {
	if err := validateDefinition(definition); err != nil {
		return nil, err
	}
	assets, err := normalizeCandidates(candidates)
	if err != nil {
		return nil, err
	}
	manifest := Manifest{
		SchemaVersion: SchemaVersion,
		ChainID:       definition.ChainID,
		Source:        source,
		Assets:        assets,
	}
	if err := Validate(manifest, definition); err != nil {
		return nil, err
	}
	encoded, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode asset manifest: %w", err)
	}
	return append(encoded, '\n'), nil
}

// Parse strictly decodes and validates one checked-in asset manifest against
// the chain and upstream identity owned by its catalog package.
func Parse(data []byte, definition Definition) (Manifest, error) {
	var manifest Manifest
	if err := decodeStrict(data, &manifest); err != nil {
		return Manifest{}, fmt.Errorf("%w: %v", ErrInvalidManifest, err)
	}
	if err := Validate(manifest, definition); err != nil {
		return Manifest{}, err
	}
	return manifest, nil
}

// Validate verifies source provenance, chain identity, canonical ordering, and
// unique token identities without assuming a specific chain or provider.
func Validate(manifest Manifest, definition Definition) error {
	if err := validateDefinition(definition); err != nil {
		return err
	}
	if manifest.SchemaVersion != SchemaVersion {
		return fmt.Errorf(
			"%w: unsupported schema version %d",
			ErrInvalidManifest,
			manifest.SchemaVersion,
		)
	}
	if manifest.ChainID != definition.ChainID {
		return fmt.Errorf(
			"%w: expected chain ID %d, got %d",
			ErrInvalidManifest,
			definition.ChainID,
			manifest.ChainID,
		)
	}
	if err := validateSource(manifest.Source, definition.Source); err != nil {
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
func ValidateEvolution(current Manifest, next Manifest, definition Definition) error {
	if err := Validate(current, definition); err != nil {
		return err
	}
	if err := Validate(next, definition); err != nil {
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

func normalizeCandidates(candidates []Candidate) ([]Asset, error) {
	if len(candidates) == 0 {
		return nil, fmt.Errorf("%w: source adapter supplied no assets", ErrInvalidManifest)
	}
	assets := make([]Asset, 0, len(candidates))
	for i, candidate := range candidates {
		id, err := catalogID(candidate.Key)
		if err != nil {
			return nil, fmt.Errorf(
				"%w: candidate %d key %q: %v",
				ErrInvalidManifest,
				i+1,
				candidate.Key,
				err,
			)
		}
		if !common.IsHexAddress(candidate.Address) {
			return nil, fmt.Errorf(
				"%w: candidate %q address %q is invalid",
				ErrInvalidManifest,
				candidate.Key,
				candidate.Address,
			)
		}
		asset := Asset{
			ID:           id,
			UpstreamKey:  candidate.Key,
			Address:      common.HexToAddress(candidate.Address).Hex(),
			IssuerSource: candidate.IssuerSource,
		}
		if err := validateAsset(asset); err != nil {
			return nil, fmt.Errorf("%w: candidate %q: %v", ErrInvalidManifest, candidate.Key, err)
		}
		assets = append(assets, asset)
	}
	sort.Slice(assets, func(i, j int) bool { return assets[i].ID < assets[j].ID })
	return assets, nil
}

func validateDefinition(definition Definition) error {
	if definition.ChainID <= 0 {
		return fmt.Errorf("%w: expected chain ID %d is invalid", ErrInvalidManifest, definition.ChainID)
	}
	if strings.TrimSpace(definition.Source.Repository) == "" {
		return fmt.Errorf("%w: expected source repository is empty", ErrInvalidManifest)
	}
	if strings.TrimSpace(definition.Source.Package) == "" {
		return fmt.Errorf("%w: expected source package is empty", ErrInvalidManifest)
	}
	if strings.TrimSpace(definition.Source.Export) == "" {
		return fmt.Errorf("%w: expected source export is empty", ErrInvalidManifest)
	}
	return nil
}

func validateSource(source Source, expected SourceDefinition) error {
	if source.Repository != expected.Repository {
		return fmt.Errorf("%w: unsupported source repository %q", ErrInvalidManifest, source.Repository)
	}
	if source.Package != expected.Package {
		return fmt.Errorf("%w: unsupported source package %q", ErrInvalidManifest, source.Package)
	}
	if strings.TrimSpace(source.PackageVersion) == "" {
		return fmt.Errorf("%w: source package version is empty", ErrInvalidManifest)
	}
	if source.Release != "v"+source.PackageVersion {
		return fmt.Errorf(
			"%w: release %q does not match package version %q",
			ErrInvalidManifest,
			source.Release,
			source.PackageVersion,
		)
	}
	if !validCommit(source.Commit) {
		return fmt.Errorf(
			"%w: source commit %q is not a 40-character Git commit",
			ErrInvalidManifest,
			source.Commit,
		)
	}
	if source.Export != expected.Export {
		return fmt.Errorf("%w: unsupported source export %q", ErrInvalidManifest, source.Export)
	}
	return nil
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
