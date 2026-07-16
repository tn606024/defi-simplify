package contract

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/shopspring/decimal"
	"github.com/tn606024/defi-simplify/bind/erc20"
	"github.com/tn606024/defi-simplify/config"
)

// ERC20Interface defines the legacy config.Coin-based ERC20 client surface.
//
// Deprecated: resolve token.Token values and compose erc20 FlowSteps instead.
type ERC20Interface interface {
	// BalanceOf returns the token balance of the given address
	BalanceOf(chain config.Chain, coin config.Coin) (decimal.Decimal, error)
	// Transfer sends tokens to the given address
	Transfer(ctx context.Context, coin config.Coin, to common.Address, amount decimal.Decimal) (*types.Receipt, error)
	// Approve allows the given address to spend tokens
	Approve(ctx context.Context, coin config.Coin, spender common.Address, amount decimal.Decimal) (*types.Receipt, error)
	// TransferFrom transfers tokens from one address to another
	TransferFrom(ctx context.Context, coin config.Coin, from common.Address, to common.Address, amount decimal.Decimal) (*types.Receipt, error)
	// Permit signs a permit message
	Permit(ctx context.Context, coin config.Coin, spender common.Address, amount decimal.Decimal, deadline *big.Int) (*types.Receipt, error)
	Nonces(ctx context.Context, coin config.Coin, owner common.Address) (*big.Int, error)
	// Allowance returns the amount of tokens the spender is allowed to spend
	Allowance(ctx context.Context, coin config.Coin, spender common.Address) (decimal.Decimal, error)
}

// ERC20Client executes legacy config.Coin-based ERC20 actions with the shared base client.
//
// Deprecated: resolve token.Token values and compose erc20 FlowSteps instead.
type ERC20Client struct {
	*BaseClientWithConverter
}

// NewERC20Client creates a legacy ERC20Client.
//
// Deprecated: resolve token.Token values and compose erc20 FlowSteps instead.
func NewERC20Client(base *BaseClient) ERC20Interface {
	return &ERC20Client{
		BaseClientWithConverter: &BaseClientWithConverter{
			BaseClient: base,
		},
	}
}

func (c *ERC20Client) BalanceOf(chain config.Chain, coin config.Coin) (decimal.Decimal, error) {
	coinAddress, err := coin.Address(chain)
	if err != nil {
		return decimal.Zero, err
	}
	decimals, err := coin.Decimals()
	if err != nil {
		return decimal.Zero, err
	}
	action := BuildBalanceOfAction(coinAddress, c.opts.From)
	balance, err := balanceOf(c.conn, action)
	if err != nil {
		return decimal.Zero, err
	}
	return c.FromWei(balance, decimals), nil
}

func (c *ERC20Client) Transfer(ctx context.Context, coin config.Coin, to common.Address, amount decimal.Decimal) (*types.Receipt, error) {
	coinAddress, err := coin.Address(c.chain)
	if err != nil {
		return nil, err
	}
	decimals, err := coin.Decimals()
	if err != nil {
		return nil, err
	}
	action := BuildTransferAction(
		coinAddress,
		to,
		c.ToWei(amount, decimals),
	)
	return executeAction(ctx, c.conn, c.opts, action)
}

func (c *ERC20Client) Approve(ctx context.Context, coin config.Coin, spender common.Address, amount decimal.Decimal) (*types.Receipt, error) {
	coinAddress, err := coin.Address(c.chain)
	if err != nil {
		return nil, err
	}
	decimals, err := coin.Decimals()
	if err != nil {
		return nil, err
	}
	action := BuildApproveAction(
		coinAddress,
		spender,
		c.ToWei(amount, decimals),
	)
	return executeAction(ctx, c.conn, c.opts, action)
}

func (c *ERC20Client) TransferFrom(ctx context.Context, coin config.Coin, from common.Address, to common.Address, amount decimal.Decimal) (*types.Receipt, error) {
	coinAddress, err := coin.Address(c.chain)
	if err != nil {
		return nil, err
	}
	decimals, err := coin.Decimals()
	if err != nil {
		return nil, err
	}
	action := BuildTransferFromAction(
		coinAddress,
		from,
		to,
		c.ToWei(amount, decimals),
	)
	return executeAction(ctx, c.conn, c.opts, action)
}

func (c *ERC20Client) Permit(ctx context.Context, coin config.Coin, spender common.Address, amount decimal.Decimal, deadline *big.Int) (*types.Receipt, error) {
	decimals, err := coin.Decimals()
	if err != nil {
		return nil, err
	}
	action, err := SignAndBuildPermitAction(
		ctx,
		c.conn,
		c.chain,
		coin,
		c.opts.From,
		spender,
		c.ToWei(amount, decimals),
		deadline,
		c.signer,
	)
	if err != nil {
		return nil, err
	}
	return executeAction(ctx, c.conn, c.opts, action)
}

func (c *ERC20Client) Nonces(ctx context.Context, coin config.Coin, owner common.Address) (*big.Int, error) {
	coinAddress, err := coin.Address(c.chain)
	if err != nil {
		return nil, err
	}
	action := BuildNoncesAction(coinAddress, owner)
	return nonces(c.conn, action)
}

func (c *ERC20Client) Allowance(ctx context.Context, coin config.Coin, spender common.Address) (decimal.Decimal, error) {
	coinAddress, err := coin.Address(c.chain)
	if err != nil {
		return decimal.Zero, err
	}
	decimals, err := coin.Decimals()
	if err != nil {
		return decimal.Zero, err
	}
	erc20Instance, err := erc20.NewErc20(coinAddress, c.conn)
	if err != nil {
		return decimal.Zero, err
	}
	allowance, err := erc20Instance.Allowance(nil, c.opts.From, spender)
	if err != nil {
		return decimal.Zero, err
	}
	return c.FromWei(allowance, decimals), nil
}
