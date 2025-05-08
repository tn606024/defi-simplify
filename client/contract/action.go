package contract

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/tn606024/defi-simplify/bind/multicall"
)

type ExecuteAction interface {
	TxAction
	AllowFailure() bool
}

type TxAction interface {
	ToTransaction(ctx context.Context, conn EthereumClient, opt *bind.TransactOpts) (*types.Transaction, error)
	Action
}

// Action defines the interface for all blockchain actions
type Action interface {
	// ToData converts the action to transaction data
	ToData(ctx context.Context, conn EthereumClient, opt *bind.TransactOpts) (common.Address, []byte, error)
	// ToIMulticall3Call3 converts the action to multicall data
	ToIMulticall3Call3(ctx context.Context, conn EthereumClient, opt *bind.TransactOpts, allowFailure bool) (*multicall.IMulticall3Call3, error)
	ToCallMsg(ctx context.Context, conn EthereumClient, opt *bind.TransactOpts) (*ethereum.CallMsg, error)
}

type BaseAction struct {
	ToDataFunc func(ctx context.Context, conn EthereumClient, opt *bind.TransactOpts) (common.Address, []byte, error)
}

func (b *BaseAction) ToData(ctx context.Context, conn EthereumClient, opt *bind.TransactOpts) (common.Address, []byte, error) {
	return b.ToDataFunc(ctx, conn, opt)
}

// DefaultToIMulticall3Call3 provides a default implementation of ToIMulticall3Call3
func (b *BaseAction) ToIMulticall3Call3(ctx context.Context, conn EthereumClient, opt *bind.TransactOpts, allowFailure bool) (*multicall.IMulticall3Call3, error) {
	target, data, err := b.ToData(ctx, conn, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to get action data: %w", err)
	}
	return &multicall.IMulticall3Call3{
		Target:       target,
		CallData:     data,
		AllowFailure: allowFailure,
	}, nil
}

// DefaultToCallMsg provides a default implementation of ToCallMsg
func (b *BaseAction) ToCallMsg(ctx context.Context, conn EthereumClient, opt *bind.TransactOpts) (*ethereum.CallMsg, error) {
	target, data, err := b.ToData(ctx, conn, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to get action data: %w", err)
	}

	return &ethereum.CallMsg{
		To:   &target,
		Data: data,
	}, nil
}

// Add wrapper struct
type ExecuteActionWrapper struct {
	TxAction
	allowFailure bool
}

// Implement ExecuteAction interface
func (a *ExecuteActionWrapper) AllowFailure() bool {
	return a.allowFailure
}

// Helper function to create wrapper
func NewExecuteAction(action TxAction, allowFailure bool) ExecuteAction {
	return &ExecuteActionWrapper{
		TxAction:     action,
		allowFailure: allowFailure,
	}
}

func SetAllExecuteAction(actions []TxAction, allowFailure bool) []ExecuteAction {
	executeActions := make([]ExecuteAction, len(actions))
	for i, action := range actions {
		executeActions[i] = NewExecuteAction(action, allowFailure)
	}
	return executeActions
}

// executeAction is a generic function to execute any action that implements TxAction
func executeAction(ctx context.Context, conn EthereumClient, opts *bind.TransactOpts, action TxAction) (*types.Receipt, error) {
	tx, err := action.ToTransaction(ctx, conn, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}
	return bind.WaitMined(ctx, conn, tx)
}
