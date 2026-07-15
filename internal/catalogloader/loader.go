// Package catalogloader turns checked-in neutral manifests into public asset
// catalogs for thin chain-specific packages.
package catalogloader

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/tn606024/defi-simplify/assets"
	"github.com/tn606024/defi-simplify/config"
	"github.com/tn606024/defi-simplify/internal/assetmanifest"
	"github.com/tn606024/defi-simplify/token"
)

// Load validates an embedded manifest and builds an immutable catalog for the
// expected SDK chain.
func Load(data []byte, definition assetmanifest.Definition, chain config.Chain) (assets.Catalog, error) {
	chainID, err := chain.ChainID()
	if err != nil {
		return assets.Catalog{}, fmt.Errorf("load asset catalog chain: %w", err)
	}
	if chainID != definition.ChainID {
		return assets.Catalog{}, fmt.Errorf(
			"catalog definition chain ID %d does not match SDK chain %d",
			definition.ChainID,
			chainID,
		)
	}
	manifest, err := assetmanifest.Parse(data, definition)
	if err != nil {
		return assets.Catalog{}, err
	}
	entries := make([]assets.Entry, 0, len(manifest.Assets))
	for _, asset := range manifest.Assets {
		ref, err := token.NewRef(chain, common.HexToAddress(asset.Address))
		if err != nil {
			return assets.Catalog{}, fmt.Errorf("load asset %s: %w", asset.ID, err)
		}
		entry, err := assets.NewEntry(asset.ID, ref)
		if err != nil {
			return assets.Catalog{}, fmt.Errorf("load asset %s: %w", asset.ID, err)
		}
		entries = append(entries, entry)
	}
	catalog, err := assets.NewCatalog(entries)
	if err != nil {
		return assets.Catalog{}, fmt.Errorf("build asset catalog: %w", err)
	}
	return catalog, nil
}
