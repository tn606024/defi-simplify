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

	deployment, err := defisimplify7702.DeploymentForChain(config.Base)
	if err != nil {
		t.Fatalf("load DefiSimplify7702Account deployment: %v", err)
	}
	accountDeployment, err := deployment.Contract(defisimplify7702.AccountContract)
	if err != nil {
		t.Fatalf("load DefiSimplify7702Account contract: %v", err)
	}
	identity := accountDeployment.RuntimeIdentity
	code, err := client.CodeAt(ctx, identity.Address, nil)
	if err != nil {
		t.Fatalf("load DefiSimplify7702Account runtime code: %v", err)
	}
	if err := identity.VerifyRuntimeCode(code); err != nil {
		t.Fatalf("verify DefiSimplify7702Account runtime code: %v", err)
	}
	return identity
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

	market, err := aave.BaseV3Market()
	if err != nil {
		t.Fatalf("load Base Aave V3 market: %v", err)
	}
	registry, err := aave.NewRegistry(client, market)
	if err != nil {
		t.Fatalf("create Base Aave registry: %v", err)
	}
	snapshot, err := registry.Load(ctx)
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
	return market, usdc, weth
}
