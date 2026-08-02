package aave

import (
	"context"
	"fmt"
)

// LoadBaseV3Snapshot resolves the reviewed Base Aave V3 market and loads one
// validated, block-pinned snapshot from backend.
//
// Each call performs an explicit snapshot load and does not retain a package
// cache. Callers that need refresh or cache control should construct a Registry
// directly with BaseV3Market and NewRegistry.
func LoadBaseV3Snapshot(
	ctx context.Context,
	backend RegistryBackend,
) (*MarketSnapshot, error) {
	market, err := BaseV3Market()
	if err != nil {
		return nil, fmt.Errorf("resolve reviewed Base Aave V3 market: %w", err)
	}
	registry, err := NewRegistry(backend, market)
	if err != nil {
		return nil, fmt.Errorf("create Base Aave V3 registry: %w", err)
	}
	snapshot, err := registry.Load(ctx)
	if err != nil {
		return nil, fmt.Errorf("load Base Aave V3 snapshot: %w", err)
	}
	return snapshot, nil
}
