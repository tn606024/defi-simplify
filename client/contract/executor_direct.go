package contract

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
)

var (
	// ErrDirectExecutorCallCount is returned when direct EOA execution receives zero or multiple calls.
	ErrDirectExecutorCallCount = errors.New("direct executor requires exactly one call")
	// ErrTransactionReverted is returned with the mined receipt when EVM execution fails.
	ErrTransactionReverted = errors.New("transaction reverted")
)

// DirectExecutor executes a single neutral call as a normal EOA transaction.
type DirectExecutor struct {
	conn EthereumClient
	opts *bind.TransactOpts
}

var _ CallExecutor = (*DirectExecutor)(nil)

// NewDirectExecutor creates an executor for single-call EOA transactions.
func NewDirectExecutor(conn EthereumClient, opts *bind.TransactOpts) *DirectExecutor {
	return &DirectExecutor{
		conn: conn,
		opts: opts,
	}
}

// ExecuteCalls executes exactly one call directly from the configured transaction signer.
func (e *DirectExecutor) ExecuteCalls(ctx context.Context, calls []Call) (*types.Receipt, error) {
	if len(calls) != 1 {
		return nil, fmt.Errorf("%w, got %d", ErrDirectExecutorCallCount, len(calls))
	}
	tx, err := e.callToTransaction(ctx, calls[0])
	if err != nil {
		return nil, err
	}
	receipt, err := bind.WaitMined(ctx, e.conn, tx)
	if err != nil {
		return nil, err
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return receipt, fmt.Errorf("%w: tx %s", ErrTransactionReverted, tx.Hash().Hex())
	}
	return receipt, nil
}

func (e *DirectExecutor) callToTransaction(ctx context.Context, call Call) (*types.Transaction, error) {
	if e.conn == nil {
		return nil, errors.New("ethereum client is nil")
	}
	if e.opts == nil {
		return nil, errors.New("transaction options are nil")
	}
	if e.opts.Signer == nil {
		return nil, errors.New("transaction signer is nil")
	}

	value := call.Value
	if value == nil {
		value = big.NewInt(0)
	}

	nonce, err := e.pendingNonce(ctx)
	if err != nil {
		return nil, err
	}
	gasPrice, err := e.gasPrice(ctx)
	if err != nil {
		return nil, err
	}
	gasLimit, err := e.gasLimit(ctx, call, value, gasPrice)
	if err != nil {
		return nil, err
	}

	tx := types.NewTransaction(nonce, call.Target, value, gasLimit, gasPrice, call.Data)
	signedTx, err := e.opts.Signer(e.opts.From, tx)
	if err != nil {
		return nil, fmt.Errorf("sign direct call transaction: %w", err)
	}
	if err := e.conn.SendTransaction(ctx, signedTx); err != nil {
		return nil, fmt.Errorf("send direct call transaction: %w", err)
	}
	return signedTx, nil
}

func (e *DirectExecutor) pendingNonce(ctx context.Context) (uint64, error) {
	if e.opts.Nonce != nil {
		return e.opts.Nonce.Uint64(), nil
	}
	nonce, err := e.conn.PendingNonceAt(ctx, e.opts.From)
	if err != nil {
		return 0, fmt.Errorf("read pending nonce: %w", err)
	}
	return nonce, nil
}

func (e *DirectExecutor) gasPrice(ctx context.Context) (*big.Int, error) {
	if e.opts.GasPrice != nil {
		return e.opts.GasPrice, nil
	}
	gasPrice, err := e.conn.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("suggest gas price: %w", err)
	}
	return gasPrice, nil
}

func (e *DirectExecutor) gasLimit(ctx context.Context, call Call, value *big.Int, gasPrice *big.Int) (uint64, error) {
	if e.opts.GasLimit != 0 {
		return e.opts.GasLimit, nil
	}
	gasLimit, err := e.conn.EstimateGas(ctx, ethereum.CallMsg{
		From:     e.opts.From,
		To:       &call.Target,
		GasPrice: gasPrice,
		Value:    value,
		Data:     call.Data,
	})
	if err != nil {
		return 0, fmt.Errorf("estimate gas: %w", err)
	}
	return gasLimit, nil
}
