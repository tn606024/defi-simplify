//go:build integration

package integration

import (
	"context"
	"math/big"
	"testing"
)

func TestBaseForkSmoke(t *testing.T) {
	ctx := context.Background()
	client := baseForkClient(t)

	chainID, err := client.ChainID(ctx)
	if err != nil {
		t.Fatalf("read chain id: %v", err)
	}
	if chainID.Cmp(big.NewInt(8453)) != 0 {
		t.Fatalf("expected Base chain id 8453, got %s", chainID.String())
	}
}
