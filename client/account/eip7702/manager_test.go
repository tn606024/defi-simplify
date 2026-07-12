package eip7702_test

import (
	"context"
	"math"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/tn606024/defi-simplify/client/account/eip7702"
	"github.com/tn606024/defi-simplify/config"
)

func TestManagerDelegatesAndSubmitsSetCodeTransaction(t *testing.T) {
	ctx := context.Background()
	chainID := big.NewInt(8453)
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	auth, err := bind.NewKeyedTransactorWithChainID(key, chainID)
	if err != nil {
		t.Fatalf("auth: %v", err)
	}
	implementation := common.HexToAddress("0x2000000000000000000000000000000000000000")
	client := newFakeSetCodeClient(auth.From)
	client.nonces[auth.From] = 10

	manager, err := eip7702.NewManager(client, auth, key, chainID)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	tx, err := manager.Delegate(ctx, implementation)
	if err != nil {
		t.Fatalf("delegate: %v", err)
	}

	assertSetCodeTxSubmitted(t, client, tx)
	if tx.Nonce() != 10 {
		t.Fatalf("unexpected tx nonce: got %d want 10", tx.Nonce())
	}
	if tx.Type() != types.SetCodeTxType {
		t.Fatalf("unexpected tx type: got %d want %d", tx.Type(), types.SetCodeTxType)
	}
	if *tx.To() != auth.From {
		t.Fatalf("expected pure delegation tx to self-call EOA: got %s want %s", tx.To().Hex(), auth.From.Hex())
	}
	wantGas := client.estimatedGas + client.estimatedGas/5
	if tx.Gas() != wantGas {
		t.Fatalf("unexpected gas limit: got %d want %d", tx.Gas(), wantGas)
	}
	if len(client.estimatedCall.AuthorizationList) != 1 {
		t.Fatalf("gas estimation should include authorization list: got %d authorizations want 1", len(client.estimatedCall.AuthorizationList))
	}
	if client.estimatedCall.AuthorizationList[0].Address != implementation {
		t.Fatalf("gas estimation used unexpected implementation: got %s want %s", client.estimatedCall.AuthorizationList[0].Address.Hex(), implementation.Hex())
	}

	authList := tx.SetCodeAuthorizations()
	if len(authList) != 1 {
		t.Fatalf("unexpected auth list length: got %d want 1", len(authList))
	}
	if authList[0].Address != implementation {
		t.Fatalf("unexpected implementation: got %s want %s", authList[0].Address.Hex(), implementation.Hex())
	}
	if authList[0].Nonce != 11 {
		t.Fatalf("same-EOA authorization nonce should be tx nonce + 1: got %d want 11", authList[0].Nonce)
	}
	authority, err := authList[0].Authority()
	if err != nil {
		t.Fatalf("recover authority: %v", err)
	}
	if authority != auth.From {
		t.Fatalf("unexpected authority: got %s want %s", authority.Hex(), auth.From.Hex())
	}
	sender, err := types.Sender(types.NewPragueSigner(chainID), tx)
	if err != nil {
		t.Fatalf("recover tx sender: %v", err)
	}
	if sender != auth.From {
		t.Fatalf("unexpected tx sender: got %s want %s", sender.Hex(), auth.From.Hex())
	}
}

func TestManagerRejectsGasBufferOverflow(t *testing.T) {
	ctx := context.Background()
	chainID := big.NewInt(8453)
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	auth, err := bind.NewKeyedTransactorWithChainID(key, chainID)
	if err != nil {
		t.Fatalf("auth: %v", err)
	}
	client := newFakeSetCodeClient(auth.From)
	client.estimatedGas = math.MaxUint64

	manager, err := eip7702.NewManager(client, auth, key, chainID)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	_, err = manager.BuildDelegationTransaction(ctx, common.HexToAddress("0x2000000000000000000000000000000000000000"))
	if err == nil || !strings.Contains(err.Error(), "gas limit buffer overflow") {
		t.Fatalf("expected gas limit buffer overflow, got %v", err)
	}
}

func TestManagerClearSubmitsZeroAddressAuthorization(t *testing.T) {
	ctx := context.Background()
	chainID := big.NewInt(8453)
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	auth, err := bind.NewKeyedTransactorWithChainID(key, chainID)
	if err != nil {
		t.Fatalf("auth: %v", err)
	}
	client := newFakeSetCodeClient(auth.From)

	manager, err := eip7702.NewManager(client, auth, key, chainID)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	tx, err := manager.Clear(ctx)
	if err != nil {
		t.Fatalf("clear: %v", err)
	}

	assertSetCodeTxSubmitted(t, client, tx)
	authList := tx.SetCodeAuthorizations()
	if len(authList) != 1 {
		t.Fatalf("unexpected auth list length: got %d want 1", len(authList))
	}
	if authList[0].Address != (common.Address{}) {
		t.Fatalf("clear authorization should use zero address, got %s", authList[0].Address.Hex())
	}
}

func TestManagerDelegateToSimple7702UsesBaseConfig(t *testing.T) {
	ctx := context.Background()
	chainID := big.NewInt(8453)
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	auth, err := bind.NewKeyedTransactorWithChainID(key, chainID)
	if err != nil {
		t.Fatalf("auth: %v", err)
	}
	client := newFakeSetCodeClient(auth.From)
	want, err := config.Base.Simple7702AccountImplementationAddress()
	if err != nil {
		t.Fatalf("simple7702 config: %v", err)
	}

	manager, err := eip7702.NewManager(client, auth, key, chainID)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	tx, err := manager.DelegateToSimple7702(ctx, config.Base)
	if err != nil {
		t.Fatalf("delegate to simple7702: %v", err)
	}

	authList := tx.SetCodeAuthorizations()
	if got := authList[0].Address; got != want {
		t.Fatalf("unexpected Simple7702 implementation: got %s want %s", got.Hex(), want.Hex())
	}
}

func TestManagerCanSwitchDelegationImplementationWithSameEOA(t *testing.T) {
	ctx := context.Background()
	chainID := big.NewInt(8453)
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	auth, err := bind.NewKeyedTransactorWithChainID(key, chainID)
	if err != nil {
		t.Fatalf("auth: %v", err)
	}
	client := newFakeSetCodeClient(auth.From)
	firstImplementation := common.HexToAddress("0x2000000000000000000000000000000000000000")
	secondImplementation := common.HexToAddress("0x3000000000000000000000000000000000000000")

	manager, err := eip7702.NewManager(client, auth, key, chainID)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	firstTx, err := manager.Delegate(ctx, firstImplementation)
	if err != nil {
		t.Fatalf("delegate first implementation: %v", err)
	}
	secondTx, err := manager.Delegate(ctx, secondImplementation)
	if err != nil {
		t.Fatalf("delegate second implementation: %v", err)
	}

	if len(client.sent) != 2 {
		t.Fatalf("unexpected sent tx count: got %d want 2", len(client.sent))
	}
	firstAuth := firstTx.SetCodeAuthorizations()[0]
	secondAuth := secondTx.SetCodeAuthorizations()[0]
	if firstAuth.Address != firstImplementation {
		t.Fatalf("unexpected first implementation: got %s want %s", firstAuth.Address.Hex(), firstImplementation.Hex())
	}
	if secondAuth.Address != secondImplementation {
		t.Fatalf("unexpected second implementation: got %s want %s", secondAuth.Address.Hex(), secondImplementation.Hex())
	}
	if firstAuth.Nonce != 1 {
		t.Fatalf("unexpected first authorization nonce: got %d want 1", firstAuth.Nonce)
	}
	if secondAuth.Nonce != 2 {
		t.Fatalf("unexpected second authorization nonce: got %d want 2", secondAuth.Nonce)
	}
}

func assertSetCodeTxSubmitted(t *testing.T, client *fakeSetCodeClient, tx *types.Transaction) {
	t.Helper()

	if len(client.sent) != 1 {
		t.Fatalf("unexpected sent tx count: got %d want 1", len(client.sent))
	}
	if client.sent[0] != tx {
		t.Fatal("returned tx was not the submitted tx")
	}
}

type fakeSetCodeClient struct {
	sender        common.Address
	nonces        map[common.Address]uint64
	code          map[common.Address][]byte
	tipCap        *big.Int
	baseFee       *big.Int
	estimatedGas  uint64
	estimatedCall ethereum.CallMsg
	sent          []*types.Transaction
}

func newFakeSetCodeClient(account common.Address) *fakeSetCodeClient {
	return &fakeSetCodeClient{
		sender: account,
		nonces: map[common.Address]uint64{
			account: 0,
		},
		code:         make(map[common.Address][]byte),
		tipCap:       big.NewInt(2_000_000_000),
		baseFee:      big.NewInt(10_000_000_000),
		estimatedGas: params.TxGas,
	}
}

func (c *fakeSetCodeClient) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	return c.nonces[account], nil
}

func (c *fakeSetCodeClient) SuggestGasTipCap(ctx context.Context) (*big.Int, error) {
	return c.tipCap, nil
}

func (c *fakeSetCodeClient) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	return &types.Header{BaseFee: c.baseFee}, nil
}

func (c *fakeSetCodeClient) EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error) {
	c.estimatedCall = msg
	return c.estimatedGas, nil
}

func (c *fakeSetCodeClient) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	c.sent = append(c.sent, tx)
	c.nonces[c.sender] = tx.Nonce() + 1
	return nil
}

func (c *fakeSetCodeClient) CodeAt(ctx context.Context, account common.Address, blockNumber *big.Int) ([]byte, error) {
	return c.code[account], nil
}
