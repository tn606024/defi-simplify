package contract

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/shopspring/decimal"
	"github.com/tn606024/defi-simplify/bind/multicall"
	"github.com/tn606024/defi-simplify/config"
	"github.com/tn606024/defi-simplify/helper"
)

// BaseClient is the base client for all contract interactions
type BaseClient struct {
	conn           EthereumClient
	chain          config.Chain
	opts           *bind.TransactOpts
	signer         *helper.MsgSigner
	actionExecutor ActionExecutor
	callExecutor   CallExecutor
}

// BaseClientWithConverter is a client that can convert between wei and decimal amounts
type BaseClientWithConverter struct {
	*BaseClient
}

// NewBaseClient creates a new BaseClient
func NewBaseClient(conn EthereumClient, chain config.Chain, opts *bind.TransactOpts, signer *helper.MsgSigner) *BaseClient {
	return &BaseClient{
		conn:   conn,
		chain:  chain,
		opts:   opts,
		signer: signer,
	}
}

// ToWei converts a decimal amount to wei
func (c *BaseClient) ToWei(amount decimal.Decimal, decimals uint8) *big.Int {
	return helper.ToWei(amount, decimals)
}

// FromWei converts wei to a decimal amount
func (c *BaseClient) FromWei(amount *big.Int, decimals uint8) decimal.Decimal {
	return helper.FromWei(amount, decimals)
}

// SetActionExecutor configures the write executor used by ExecuteTxActions.
func (c *BaseClient) SetActionExecutor(executor ActionExecutor) {
	c.actionExecutor = executor
}

// SetCallExecutor configures the call executor used by ExecuteCalls.
func (c *BaseClient) SetCallExecutor(executor CallExecutor) {
	c.callExecutor = executor
}

// ExecuteTxActions executes write actions through the configured executor.
// If no executor is configured, it uses the default Multicall executor.
func (c *BaseClient) ExecuteTxActions(ctx context.Context, actions []ExecuteAction) (*types.Receipt, error) {
	executor := c.actionExecutor
	if executor == nil {
		executor = NewMulticallExecutor(c.conn, c.chain, c.opts)
	}
	return executor.ExecuteActions(ctx, actions)
}

// ExecuteCalls executes neutral calls through the configured executor.
// If no executor is configured, it uses the default Multicall executor.
func (c *BaseClient) ExecuteCalls(ctx context.Context, calls []Call) (*types.Receipt, error) {
	executor := c.callExecutor
	if executor == nil {
		executor = NewMulticallExecutor(c.conn, c.chain, c.opts)
	}
	return executor.ExecuteCalls(ctx, calls)
}

// ExecuteMulticalls executes read-only actions through the default Multicall executor.
func (c *BaseClient) ExecuteMulticalls(ctx context.Context, actions []Action) ([]multicall.IMulticall3Result, error) {
	return NewMulticallExecutor(c.conn, c.chain, c.opts).ExecuteReadActions(ctx, actions)
}
