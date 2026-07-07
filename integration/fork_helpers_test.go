//go:build integration

package integration

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/tn606024/defi-simplify/bind/erc20"
	"github.com/tn606024/defi-simplify/config"
)

var forkTestUser = common.HexToAddress("0x1000000000000000000000000000000000000001")

func TestForkHelperSetsETHBalance(t *testing.T) {
	ctx := context.Background()
	ethClient := baseForkClient(t)
	rpcClient := baseForkRPCClient(t)
	requireAnvilFork(t, ctx, rpcClient)

	balance := new(big.Int).Mul(big.NewInt(2), big.NewInt(1_000_000_000_000_000_000))
	if err := setForkETHBalance(ctx, rpcClient, forkTestUser, balance); err != nil {
		t.Fatalf("set ETH balance: %v", err)
	}

	actual, err := ethClient.BalanceAt(ctx, forkTestUser, nil)
	if err != nil {
		t.Fatalf("read ETH balance: %v", err)
	}
	if actual.Cmp(balance) != 0 {
		t.Fatalf("expected ETH balance %s, got %s", balance.String(), actual.String())
	}
}

func TestForkHelperStartsAndStopsImpersonation(t *testing.T) {
	ctx := context.Background()
	rpcClient := baseForkRPCClient(t)
	requireAnvilFork(t, ctx, rpcClient)

	if err := impersonateForkAccount(ctx, rpcClient, baseUSDCFunder); err != nil {
		t.Fatalf("impersonate account: %v", err)
	}
	if err := stopImpersonatingForkAccount(ctx, rpcClient, baseUSDCFunder); err != nil {
		t.Fatalf("stop impersonating account: %v", err)
	}
}

func TestForkHelperFundsUserWithUSDC(t *testing.T) {
	ctx := context.Background()
	ethClient := baseForkClient(t)
	rpcClient := baseForkRPCClient(t)
	requireAnvilFork(t, ctx, rpcClient)

	usdc, err := config.USDC.Address(config.Base)
	if err != nil {
		t.Fatalf("load Base USDC address: %v", err)
	}
	token, err := erc20.NewErc20(usdc, ethClient)
	if err != nil {
		t.Fatalf("create USDC binding: %v", err)
	}

	before, err := token.BalanceOf(nil, forkTestUser)
	if err != nil {
		t.Fatalf("read USDC balance before funding: %v", err)
	}

	amount := big.NewInt(1_000_000)
	if err := fundBaseUSDCFromHolder(ctx, rpcClient, ethClient, forkTestUser, amount); err != nil {
		t.Fatalf("fund USDC: %v", err)
	}

	after, err := token.BalanceOf(nil, forkTestUser)
	if err != nil {
		t.Fatalf("read USDC balance after funding: %v", err)
	}
	delta := new(big.Int).Sub(after, before)
	if delta.Cmp(amount) != 0 {
		t.Fatalf("expected USDC balance delta %s, got %s", amount.String(), delta.String())
	}
}
