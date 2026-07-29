package contract

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
)

var (
	// ErrTransactionReverted is returned with the mined receipt when EVM execution fails.
	ErrTransactionReverted = errors.New("transaction reverted")
)

type transactionSender struct {
	conn EthereumClient
	opts *bind.TransactOpts
}

// SendCall signs, submits, and waits for one low-level transaction call.
//
// This utility supports account self-calls and delegation lifecycle tooling.
// Normal Flow execution is owned by defi.Runner.
func SendCall(
	ctx context.Context,
	conn EthereumClient,
	opts *bind.TransactOpts,
	call Call,
) (*types.Receipt, error) {
	sender := &transactionSender{
		conn: conn,
		opts: opts,
	}
	tx, err := sender.callToTransaction(ctx, call)
	if err != nil {
		return nil, err
	}
	receipt, err := bind.WaitMined(ctx, conn, tx)
	if err != nil {
		return nil, err
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return receipt, fmt.Errorf("%w: tx %s", ErrTransactionReverted, tx.Hash().Hex())
	}
	return receipt, nil
}

func (e *transactionSender) callToTransaction(ctx context.Context, call Call) (*types.Transaction, error) {
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

func (e *transactionSender) pendingNonce(ctx context.Context) (uint64, error) {
	if e.opts.Nonce != nil {
		return e.opts.Nonce.Uint64(), nil
	}
	nonce, err := e.conn.PendingNonceAt(ctx, e.opts.From)
	if err != nil {
		return 0, fmt.Errorf("read pending nonce: %w", err)
	}
	return nonce, nil
}

func (e *transactionSender) gasPrice(ctx context.Context) (*big.Int, error) {
	if e.opts.GasPrice != nil {
		return e.opts.GasPrice, nil
	}
	gasPrice, err := e.conn.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("suggest gas price: %w", err)
	}
	return gasPrice, nil
}

func (e *transactionSender) gasLimit(ctx context.Context, call Call, value *big.Int, gasPrice *big.Int) (uint64, error) {
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
	return addTransactionGasLimitBuffer(gasLimit)
}

func addTransactionGasLimitBuffer(estimated uint64) (uint64, error) {
	buffer := estimated / 4
	if math.MaxUint64-estimated < buffer {
		return 0, errors.New("transaction gas limit buffer overflow")
	}
	return estimated + buffer, nil
}
