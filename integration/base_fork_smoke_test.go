//go:build integration

package integration

import (
	"context"
	"math/big"
	"testing"

	"github.com/tn606024/defi-simplify/config"
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

	pool, err := config.Base.AaveV3PoolAddress()
	if err != nil {
		t.Fatalf("load Base Aave V3 pool address: %v", err)
	}
	assertContractCode(t, ctx, client, pool, "Aave V3 Pool")

	usdc, err := config.USDC.Address(config.Base)
	if err != nil {
		t.Fatalf("load Base USDC address: %v", err)
	}
	assertContractCode(t, ctx, client, usdc, "USDC")
}
