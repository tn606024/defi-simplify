//go:build integration

package integration

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	sdkcontract "github.com/tn606024/defi-simplify/client/contract"
	"github.com/tn606024/defi-simplify/config"
	"github.com/tn606024/defi-simplify/helper"
)

func newForkDefiClient(t testHelper, ctx context.Context, rpcClient *rpc.Client, ethClient *ethclient.Client) (*sdkcontract.DefiClient, common.Address) {
	t.Helper()

	opts, signer, user := newForkTransactor(t, ctx, rpcClient)
	return sdkcontract.NewDefiClient(opts, ethClient, signer, config.Base), user
}

func newForkTransactor(t testHelper, ctx context.Context, rpcClient *rpc.Client) (*bind.TransactOpts, *helper.MsgSigner, common.Address) {
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

	return opts, helper.NewMsgSigner(privateKey), user
}
