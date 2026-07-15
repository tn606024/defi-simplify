package aave

import (
	_ "embed"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/tn606024/defi-simplify/config"
	"github.com/tn606024/defi-simplify/internal/aavemanifest"
)

const DeploymentManifestSchemaVersion = aavemanifest.SchemaVersion

var ErrInvalidDeploymentManifest = errors.New("invalid Aave deployment manifest")

//go:embed manifests/aave-v3-base.json
var baseV3DeploymentManifest []byte

// DeploymentSource identifies the pinned official Address Book artifact used
// to generate a checked-in deployment manifest.
type DeploymentSource struct {
	repository     string
	packageName    string
	packageVersion string
	release        string
	commit         string
	export         string
}

// Repository returns the upstream source repository.
func (s DeploymentSource) Repository() string { return s.repository }

// Package returns the upstream npm package name.
func (s DeploymentSource) Package() string { return s.packageName }

// PackageVersion returns the exact upstream package version.
func (s DeploymentSource) PackageVersion() string { return s.packageVersion }

// Release returns the exact upstream release tag.
func (s DeploymentSource) Release() string { return s.release }

// Commit returns the upstream release commit.
func (s DeploymentSource) Commit() string { return s.commit }

// Export returns the Address Book market export used by the generator.
func (s DeploymentSource) Export() string { return s.export }

// DeploymentManifest is an immutable SDK view of one reviewed deployment
// manifest. It contains trust anchors only; reserve membership is discovered
// separately from the selected market at a pinned block.
type DeploymentManifest struct {
	schemaVersion int
	market        Market
	source        DeploymentSource
}

// ParseDeploymentManifest strictly parses one supported deployment manifest.
// Unknown markets, chains, sources, fields, and malformed addresses fail
// closed.
func ParseDeploymentManifest(data []byte) (DeploymentManifest, error) {
	parsed, err := aavemanifest.Parse(data)
	if err != nil {
		return DeploymentManifest{}, fmt.Errorf("%w: %w", ErrInvalidDeploymentManifest, err)
	}

	chain, err := config.ChainIDToChain(parsed.ChainID)
	if err != nil {
		return DeploymentManifest{}, fmt.Errorf(
			"%w: unsupported chain ID %d: %v",
			ErrInvalidDeploymentManifest,
			parsed.ChainID,
			err,
		)
	}
	market, err := NewMarket(
		parsed.MarketID,
		chain,
		common.HexToAddress(parsed.Contracts.Pool),
		common.HexToAddress(parsed.Contracts.PoolAddressesProvider),
		common.HexToAddress(parsed.Contracts.AaveProtocolDataProvider),
		common.HexToAddress(parsed.Contracts.WrappedTokenGateway),
	)
	if err != nil {
		return DeploymentManifest{}, fmt.Errorf("%w: %w", ErrInvalidDeploymentManifest, err)
	}

	return DeploymentManifest{
		schemaVersion: parsed.SchemaVersion,
		market:        market,
		source: DeploymentSource{
			repository:     parsed.Source.Repository,
			packageName:    parsed.Source.Package,
			packageVersion: parsed.Source.PackageVersion,
			release:        parsed.Source.Release,
			commit:         parsed.Source.Commit,
			export:         parsed.Source.Export,
		},
	}, nil
}

// BaseV3Deployment returns the checked-in, reviewed Base Aave V3 deployment.
// It performs no network access.
func BaseV3Deployment() (DeploymentManifest, error) {
	return ParseDeploymentManifest(baseV3DeploymentManifest)
}

// BaseV3Market returns the checked-in Base Aave V3 market trust anchors.
func BaseV3Market() (Market, error) {
	manifest, err := BaseV3Deployment()
	if err != nil {
		return Market{}, err
	}
	return manifest.Market(), nil
}

// SchemaVersion returns the manifest schema version.
func (m DeploymentManifest) SchemaVersion() int { return m.schemaVersion }

// Market returns the resolved immutable Aave market.
func (m DeploymentManifest) Market() Market { return m.market }

// Source returns the immutable upstream source identity.
func (m DeploymentManifest) Source() DeploymentSource { return m.source }
