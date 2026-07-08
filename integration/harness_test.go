//go:build integration

package integration

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

const baseRPCURLEnv = "BASE_RPC_URL"

type testHelper interface {
	Helper()
	Fatalf(format string, args ...any)
	Skipf(format string, args ...any)
	Cleanup(func())
}

func baseForkClient(t testHelper) *ethclient.Client {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := ethclient.DialContext(ctx, baseForkRPCURL(t))
	if err != nil {
		t.Fatalf("dial %s: %v", baseRPCURLEnv, err)
	}
	t.Cleanup(client.Close)

	return client
}

func baseForkRPCClient(t testHelper) *rpc.Client {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := rpc.DialContext(ctx, baseForkRPCURL(t))
	if err != nil {
		t.Fatalf("dial %s: %v", baseRPCURLEnv, err)
	}
	t.Cleanup(client.Close)

	return client
}

func baseForkRPCURL(t testHelper) string {
	t.Helper()

	rpcURL := strings.TrimSpace(os.Getenv(baseRPCURLEnv))
	if rpcURL == "" {
		t.Skipf("set %s to a Base mainnet or local Anvil fork RPC URL to run integration tests", baseRPCURLEnv)
	}
	return rpcURL
}

func requireAnvilFork(t testHelper, ctx context.Context, client *rpc.Client) {
	t.Helper()

	var version string
	if err := client.CallContext(ctx, &version, "web3_clientVersion"); err != nil {
		t.Fatalf("read RPC client version: %v", err)
	}
	if !strings.Contains(strings.ToLower(version), "anvil") {
		t.Skipf("requires a local Anvil fork RPC, got %q", version)
	}
}

func assertContractCode(t testHelper, ctx context.Context, client *ethclient.Client, address common.Address, label string) {
	t.Helper()

	code, err := client.CodeAt(ctx, address, nil)
	if err != nil {
		t.Fatalf("read %s code at %s: %v", label, address.Hex(), err)
	}
	if len(code) == 0 {
		t.Fatalf("expected %s contract code at %s", label, address.Hex())
	}
}
