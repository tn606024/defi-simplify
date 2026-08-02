//go:build integration

package integration

import (
	"context"
	"crypto/ecdsa"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/tn606024/defi-simplify/aave"
	baseassets "github.com/tn606024/defi-simplify/assets/base"
	"github.com/tn606024/defi-simplify/client/account/defisimplify7702"
	"github.com/tn606024/defi-simplify/client/account/eip7702"
	sdkcontract "github.com/tn606024/defi-simplify/client/contract"
	"github.com/tn606024/defi-simplify/config"
)

func newForkDefiClient(t testHelper, ctx context.Context, rpcClient *rpc.Client, ethClient *ethclient.Client) (*sdkcontract.DefiClient, common.Address) {
	t.Helper()

	opts, user := newForkTransactor(t, ctx, rpcClient)
	return sdkcontract.NewDefiClient(opts, ethClient, config.Base), user
}

func newForkTransactor(t testHelper, ctx context.Context, rpcClient *rpc.Client) (*bind.TransactOpts, common.Address) {
	t.Helper()

	opts, _, user := newForkTransactorWithKey(t, ctx, rpcClient)
	return opts, user
}

func newForkTransactorWithKey(t testHelper, ctx context.Context, rpcClient *rpc.Client) (*bind.TransactOpts, *ecdsa.PrivateKey, common.Address) {
	t.Helper()

	privateKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("generate fork test private key: %v", err)
	}
	user := crypto.PubkeyToAddress(privateKey.PublicKey)

	ethBalance := big.NewInt(1_000_000_000_000_000_000)
	if err := setForkETHBalance(ctx, rpcClient, user, ethBalance); err != nil {
		t.Fatalf("fund fork test user ETH: %v", err)
	}

	chainID, err := config.Base.ChainID()
	if err != nil {
		t.Fatalf("load Base chain id: %v", err)
	}
	opts, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(int64(chainID)))
	if err != nil {
		t.Fatalf("create fork test transactor: %v", err)
	}

	return opts, privateKey, user
}

func loadDefiSimplifyAccountIdentity(
	t testHelper,
	ctx context.Context,
	client *ethclient.Client,
) defisimplify7702.RuntimeIdentity {
	t.Helper()

	implementation, err := defisimplify7702.ResolveAndVerifyAccountDeployment(
		ctx,
		client,
		config.Base,
	)
	if err != nil {
		t.Fatalf("resolve verified DefiSimplify7702Account deployment: %v", err)
	}
	return implementation.RuntimeIdentity
}

func delegateForkEOA(
	t testHelper,
	ctx context.Context,
	client *ethclient.Client,
	manager *eip7702.Manager,
	user common.Address,
	implementation common.Address,
) {
	t.Helper()

	tx, err := manager.Delegate(ctx, implementation)
	if err != nil {
		t.Fatalf("delegate fork EOA: %v", err)
	}
	receipt, err := bind.WaitMined(ctx, client, tx)
	if err != nil {
		t.Fatalf("wait for fork EOA delegation: %v", err)
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		t.Fatalf("fork EOA delegation reverted: tx %s", tx.Hash().Hex())
	}
	if err := manager.AssertDelegatedTo(ctx, user, implementation); err != nil {
		t.Fatalf("verify fork EOA delegation: %v", err)
	}
}

func loadBaseAaveReserves(
	t testHelper,
	ctx context.Context,
	client *ethclient.Client,
) (aave.Market, aave.Reserve, aave.Reserve) {
	t.Helper()

	snapshot, err := aave.LoadBaseV3Snapshot(ctx, client)
	if err != nil {
		t.Fatalf("load Base Aave reserve snapshot: %v", err)
	}
	usdc, err := snapshot.Reserve(baseassets.USDC)
	if err != nil {
		t.Fatalf("resolve Base Aave USDC reserve: %v", err)
	}
	weth, err := snapshot.Reserve(baseassets.WETH)
	if err != nil {
		t.Fatalf("resolve Base Aave WETH reserve: %v", err)
	}
	return snapshot.Market(), usdc, weth
}
