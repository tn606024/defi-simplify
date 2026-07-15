// Package aaveassetmanifest adapts official Aave Address Book market exports
// into the provider-neutral asset manifest model.
package aaveassetmanifest

import (
	"github.com/tn606024/defi-simplify/internal/aaveaddressbook"
	"github.com/tn606024/defi-simplify/internal/assetmanifest"
)

// DefinitionFor returns the neutral catalog definition owned by one Aave
// Address Book market export.
func DefinitionFor(export aaveaddressbook.ExportDefinition) assetmanifest.Definition {
	return assetmanifest.Definition{
		ChainID: export.ChainID,
		Source: assetmanifest.SourceDefinition{
			Repository: aaveaddressbook.OfficialRepository,
			Package:    aaveaddressbook.OfficialPackage,
			Export:     export.Name + ".ASSETS",
		},
	}
}

// Generate converts one pinned market export into a deterministic reviewed
// asset manifest. Address Book parsing stays here instead of leaking provider
// assumptions into the neutral manifest package.
func Generate(data []byte, definition aaveaddressbook.ExportDefinition) ([]byte, error) {
	exported, err := aaveaddressbook.ParseExportFor(data, definition)
	if err != nil {
		return nil, err
	}
	candidates := make([]assetmanifest.Candidate, 0, len(exported.Assets))
	for _, asset := range exported.Assets {
		candidates = append(candidates, assetmanifest.Candidate{
			Key:          asset.Key,
			Address:      asset.Address,
			IssuerSource: asset.IssuerSource,
		})
	}
	source := aaveaddressbook.SourceFromExport(exported, definition.Name+".ASSETS")
	return assetmanifest.Generate(
		DefinitionFor(definition),
		assetmanifest.Source{
			Repository:     source.Repository,
			Package:        source.Package,
			PackageVersion: source.PackageVersion,
			Release:        source.Release,
			Commit:         source.Commit,
			Export:         source.Export,
		},
		candidates,
	)
}
