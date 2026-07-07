//go:build integration

package integration

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

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
