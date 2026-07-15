// Package base provides reviewed token references for common assets on Base.
// Catalog IDs are SDK-owned, case-sensitive keys; they are not ERC20 symbols
// and catalog membership is not an execution allowlist.
package base

import (
	_ "embed"
	"fmt"

	assetcatalog "github.com/tn606024/defi-simplify/assets"
	"github.com/tn606024/defi-simplify/config"
	"github.com/tn606024/defi-simplify/internal/aaveaddressbook"
	"github.com/tn606024/defi-simplify/internal/aaveassetmanifest"
	"github.com/tn606024/defi-simplify/internal/catalogloader"
	"github.com/tn606024/defi-simplify/token"
)

//go:embed manifest.json
var manifestData []byte

// Entry is one immutable reviewed catalog identity.
type Entry = assetcatalog.Entry

var defaultCatalog = mustLoadCatalog()

// Entries returns a deterministic copy ordered by catalog ID.
func Entries() []Entry {
	return defaultCatalog.Entries()
}

// Lookup resolves an exact, case-sensitive reviewed catalog ID. The ID is an
// SDK key, not an ERC20 symbol. Unknown assets remain usable through
// token.NewRef with a caller-supplied address.
func Lookup(id string) (token.Ref, bool) {
	return defaultCatalog.Lookup(id)
}

func mustLoadCatalog() assetcatalog.Catalog {
	definition := aaveassetmanifest.DefinitionFor(aaveaddressbook.BaseV3ExportDefinition())
	catalog, err := catalogloader.Load(manifestData, definition, config.Base)
	if err != nil {
		panic(fmt.Sprintf("load checked-in Base asset manifest: %v", err))
	}
	return catalog
}

func mustLookup(id string) token.Ref {
	ref, ok := Lookup(id)
	if !ok {
		panic(fmt.Sprintf("checked-in Base asset %s is missing", id))
	}
	return ref
}
