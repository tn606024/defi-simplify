// Package assets provides protocol-neutral, chain-scoped asset catalogs.
package assets

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/tn606024/defi-simplify/config"
	"github.com/tn606024/defi-simplify/token"
)

// ErrInvalidCatalog identifies malformed catalog entries or mixed-chain
// catalogs.
var ErrInvalidCatalog = errors.New("invalid asset catalog")

// Entry binds one exact, case-sensitive catalog ID to a token reference.
type Entry struct {
	id  string
	ref token.Ref
}

// NewEntry creates a validated catalog entry.
func NewEntry(id string, ref token.Ref) (Entry, error) {
	if id == "" || strings.TrimSpace(id) != id {
		return Entry{}, fmt.Errorf("%w: catalog ID %q is empty or has surrounding whitespace", ErrInvalidCatalog, id)
	}
	if err := ref.Validate(); err != nil {
		return Entry{}, fmt.Errorf("%w: catalog ID %q: %v", ErrInvalidCatalog, id, err)
	}
	return Entry{id: id, ref: ref}, nil
}

// ID returns the exact SDK catalog ID. It is not an ERC20 symbol.
func (e Entry) ID() string {
	return e.id
}

// Ref returns the chain-scoped token identity.
func (e Entry) Ref() token.Ref {
	return e.ref
}

// Catalog is an immutable, deterministic collection for exactly one chain.
type Catalog struct {
	chain   config.Chain
	entries []Entry
	byID    map[string]token.Ref
}

// NewCatalog validates and sorts entries for one chain. Duplicate IDs and
// duplicate token identities are rejected.
func NewCatalog(entries []Entry) (Catalog, error) {
	if len(entries) == 0 {
		return Catalog{}, fmt.Errorf("%w: catalog is empty", ErrInvalidCatalog)
	}
	cloned := append([]Entry(nil), entries...)
	sort.Slice(cloned, func(i, j int) bool { return cloned[i].ID() < cloned[j].ID() })
	chain := cloned[0].Ref().Chain()
	byID := make(map[string]token.Ref, len(cloned))
	byRef := make(map[token.Ref]string, len(cloned))
	for _, entry := range cloned {
		validated, err := NewEntry(entry.ID(), entry.Ref())
		if err != nil {
			return Catalog{}, err
		}
		if validated.Ref().Chain() != chain {
			return Catalog{}, fmt.Errorf(
				"%w: catalog ID %q belongs to chain %d instead of %d",
				ErrInvalidCatalog,
				validated.ID(),
				validated.Ref().Chain(),
				chain,
			)
		}
		if _, ok := byID[validated.ID()]; ok {
			return Catalog{}, fmt.Errorf("%w: duplicate catalog ID %q", ErrInvalidCatalog, validated.ID())
		}
		if previous, ok := byRef[validated.Ref()]; ok {
			return Catalog{}, fmt.Errorf(
				"%w: catalog IDs %q and %q use the same token reference",
				ErrInvalidCatalog,
				previous,
				validated.ID(),
			)
		}
		byID[validated.ID()] = validated.Ref()
		byRef[validated.Ref()] = validated.ID()
	}
	return Catalog{chain: chain, entries: cloned, byID: byID}, nil
}

// Chain returns the chain shared by every catalog entry.
func (c Catalog) Chain() config.Chain {
	return c.chain
}

// Entries returns a deterministic copy ordered by catalog ID.
func (c Catalog) Entries() []Entry {
	return append([]Entry(nil), c.entries...)
}

// Lookup resolves an exact, case-sensitive catalog ID.
func (c Catalog) Lookup(id string) (token.Ref, bool) {
	ref, ok := c.byID[id]
	return ref, ok
}
