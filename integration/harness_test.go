//go:build integration

package integration

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

const baseRPCURLEnv = "BASE_RPC_URL"

func baseForkClient(t *testing.T) *ethclient.Client {
	t.Helper()

	rpcURL := strings.TrimSpace(os.Getenv(baseRPCURLEnv))
	if rpcURL == "" {
		t.Skipf("set %s to a Base mainnet or local Anvil fork RPC URL to run integration tests", baseRPCURLEnv)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		t.Fatalf("dial %s: %v", baseRPCURLEnv, err)
	}
	t.Cleanup(client.Close)

	return client
}

func assertContractCode(t *testing.T, ctx context.Context, client *ethclient.Client, address common.Address, label string) {
	t.Helper()

	code, err := client.CodeAt(ctx, address, nil)
	if err != nil {
		t.Fatalf("read %s code at %s: %v", label, address.Hex(), err)
	}
	if len(code) == 0 {
		t.Fatalf("expected %s contract code at %s", label, address.Hex())
	}
}
