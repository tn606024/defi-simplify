package contract

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/shopspring/decimal"
	"github.com/tn606024/defi-simplify/bind/aave"
	"github.com/tn606024/defi-simplify/config"
)

// AaveV3Interface defines the legacy config.Coin-based Aave V3 client surface.
//
// Deprecated: load aave.Reserve values through aave.Registry and compose Aave FlowSteps instead.
type AaveV3Interface interface {
	// Supply supplies tokens to Aave
	Supply(ctx context.Context, coin config.Coin, amount decimal.Decimal) (*types.Receipt, error)
	// Withdraw withdraws tokens from Aave
	Withdraw(ctx context.Context, coin config.Coin, amount decimal.Decimal) (*types.Receipt, error)
	// Borrow borrows tokens from Aave
	Borrow(ctx context.Context, coin config.Coin, amount decimal.Decimal) (*types.Receipt, error)
	// Repay repays borrowed tokens
	Repay(ctx context.Context, coin config.Coin, amount decimal.Decimal) (*types.Receipt, error)
	// GetReserveData gets the reserve data for a given coin
	GetReserveData(ctx context.Context, coin config.Coin) (*aave.DataTypesReserveData, error)
	// GetUserAccountData gets the user account data for a given coin
	GetUserAccountData(ctx context.Context) (*DataTypesUserAccountData, error)
	GetAllReservesTokens(ctx context.Context) ([]aave.IPoolDataProviderTokenData, error)
	GetUserReserveData(ctx context.Context, asset common.Address) (*DataTypesUserReserveData, error)
}

// AaveV3Client executes legacy config.Coin-based Aave V3 actions with the shared base client.
//
// Deprecated: load aave.Reserve values through aave.Registry and compose Aave FlowSteps instead.
type AaveV3Client struct {
	*BaseClientWithConverter
}

// NewAaveV3Client creates a legacy AaveV3Client.
//
// Deprecated: load aave.Reserve values through aave.Registry and compose Aave FlowSteps instead.
func NewAaveV3Client(base *BaseClient) AaveV3Interface {
	return &AaveV3Client{
		BaseClientWithConverter: &BaseClientWithConverter{
			BaseClient: base,
		},
	}
}

func (c *AaveV3Client) Supply(ctx context.Context, coin config.Coin, amount decimal.Decimal) (*types.Receipt, error) {
	poolAddress, err := c.chain.AaveV3PoolAddress()
	if err != nil {
		return nil, err
	}
	coinAddress, err := coin.Address(c.chain)
	if err != nil {
		return nil, err
	}
	decimals, err := coin.Decimals()
	if err != nil {
		return nil, err
	}
	action := BuildSupplyAction(
		poolAddress,
		coinAddress,
		c.ToWei(amount, decimals),
		c.opts.From,
	)

	return executeAction(ctx, c.conn, c.opts, action)
}

func (c *AaveV3Client) Withdraw(ctx context.Context, coin config.Coin, amount decimal.Decimal) (*types.Receipt, error) {
	poolAddress, err := c.chain.AaveV3PoolAddress()
	if err != nil {
		return nil, err
	}
	coinAddress, err := coin.Address(c.chain)
	if err != nil {
		return nil, err
	}
	decimals, err := coin.Decimals()
	if err != nil {
		return nil, err
	}
	action := BuildWithdrawAction(
		poolAddress,
		coinAddress,
		c.ToWei(amount, decimals),
		c.opts.From,
	)
	return executeAction(ctx, c.conn, c.opts, action)
}

func (c *AaveV3Client) Borrow(ctx context.Context, coin config.Coin, amount decimal.Decimal) (*types.Receipt, error) {
	poolAddress, err := c.chain.AaveV3PoolAddress()
	if err != nil {
		return nil, err
	}
	coinAddress, err := coin.Address(c.chain)
	if err != nil {
		return nil, err
	}
	decimals, err := coin.Decimals()
	if err != nil {
		return nil, err
	}
	action := BuildBorrowAction(
		poolAddress,
		coinAddress,
		c.ToWei(amount, decimals),
		c.opts.From,
	)

	return executeAction(ctx, c.conn, c.opts, action)
}

func (c *AaveV3Client) Repay(ctx context.Context, coin config.Coin, amount decimal.Decimal) (*types.Receipt, error) {
	poolAddress, err := c.chain.AaveV3PoolAddress()
	if err != nil {
		return nil, err
	}
	coinAddress, err := coin.Address(c.chain)
	if err != nil {
		return nil, err
	}
	decimals, err := coin.Decimals()
	if err != nil {
		return nil, err
	}
	action := BuildRepayAction(
		poolAddress,
		coinAddress,
		c.ToWei(amount, decimals),
		c.opts.From,
	)
	return executeAction(ctx, c.conn, c.opts, action)
}

func (c *AaveV3Client) GetReserveData(ctx context.Context, coin config.Coin) (*aave.DataTypesReserveData, error) {
	poolAddress, err := c.chain.AaveV3PoolAddress()
	if err != nil {
		return nil, err
	}
	coinAddress, err := coin.Address(c.chain)
	if err != nil {
		return nil, err
	}
	action := BuildGetReserveDataAction(
		poolAddress,
		coinAddress,
	)
	return getReserveData(c.conn, action)
}

func (c *AaveV3Client) GetUserAccountData(ctx context.Context) (*DataTypesUserAccountData, error) {
	poolAddress, err := c.chain.AaveV3PoolAddress()
	if err != nil {
		return nil, err
	}
	action := BuildGetUserAccountDataAction(
		poolAddress,
		c.opts.From,
	)
	return getUserAccountData(c.conn, action)
}

func (c *AaveV3Client) GetAllReservesTokens(ctx context.Context) ([]aave.IPoolDataProviderTokenData, error) {
	protocolDataProviderAddress, err := c.chain.AaveProtocolDataProviderAddress()
	if err != nil {
		return nil, err
	}
	action := BuildGetAllReservesTokensAction(
		protocolDataProviderAddress,
	)
	return getAllReservesTokens(c.conn, action)
}

func (c *AaveV3Client) GetUserReserveData(ctx context.Context, asset common.Address) (*DataTypesUserReserveData, error) {
	protocolDataProviderAddress, err := c.chain.AaveProtocolDataProviderAddress()
	if err != nil {
		return nil, err
	}
	action := BuildGetUserReserveDataAction(
		protocolDataProviderAddress,
		asset,
		c.opts.From,
	)
	return getUserReserveData(c.conn, action)
}
