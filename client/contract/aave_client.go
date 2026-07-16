package contract

import (
	"context"
	"fmt"
	"math/big"
	"time"

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
	// SupplyWithPermit supplies tokens to Aave with permit
	SupplyWithPermit(ctx context.Context, coin config.Coin, amount decimal.Decimal) (*types.Receipt, error)
	// Withdraw withdraws tokens from Aave
	Withdraw(ctx context.Context, coin config.Coin, amount decimal.Decimal) (*types.Receipt, error)
	// Borrow borrows tokens from Aave
	Borrow(ctx context.Context, coin config.Coin, amount decimal.Decimal) (*types.Receipt, error)
	// BorrowETH borrows ETH from Aave
	BorrowETH(ctx context.Context, amount decimal.Decimal) (*types.Receipt, error)
	// Repay repays borrowed tokens
	Repay(ctx context.Context, coin config.Coin, amount decimal.Decimal) (*types.Receipt, error)
	// RepayWithPermit repays borrowed tokens with permit
	RepayWithPermit(ctx context.Context, coin config.Coin, amount decimal.Decimal) (*types.Receipt, error)
	// DepositETH deposits ETH to Aave
	DepositETH(ctx context.Context, amount decimal.Decimal) (*types.Receipt, error)
	// WithdrawETH withdraws ETH from Aave
	WithdrawETH(ctx context.Context, amount decimal.Decimal) (*types.Receipt, error)
	// WithdrawETHWithPermit withdraws ETH from Aave with permit
	WithdrawETHWithPermit(ctx context.Context, amount decimal.Decimal) (*types.Receipt, error)
	// ApproveDelegation approves delegation of tokens
	ApproveDelegation(ctx context.Context, coin config.Coin, delegatee common.Address, amount decimal.Decimal) (*types.Receipt, error)
	// DelegationWithSig delegates tokens with signature
	DelegationWithSig(ctx context.Context, coin config.Coin, delegatee common.Address, value decimal.Decimal) (*types.Receipt, error)
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

func (c *AaveV3Client) SupplyWithPermit(ctx context.Context, coin config.Coin, amount decimal.Decimal) (*types.Receipt, error) {
	permitSupported, err := coin.PermitSupported(c.chain)
	if err != nil {
		return nil, err
	}
	if !permitSupported {
		return nil, fmt.Errorf("coin %v does not support permit on chain %v", coin, c.chain)
	}
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
	deadline := big.NewInt(time.Now().Add(time.Minute * 10).Unix())
	amountWei := c.ToWei(amount, decimals)

	permitAction, err := SignAndBuildPermitAction(
		ctx,
		c.conn,
		c.chain,
		coin,
		c.opts.From,
		poolAddress,
		amountWei,
		deadline,
		c.signer,
	)
	if err != nil {
		return nil, err
	}

	action := BuildSupplyWithPermitAction(
		poolAddress,
		coinAddress,
		amountWei,
		c.opts.From,
		0,
		deadline,
		permitAction.v,
		permitAction.r,
		permitAction.s,
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

func (c *AaveV3Client) BorrowETH(ctx context.Context, amount decimal.Decimal) (*types.Receipt, error) {
	wrappedTokenGatewayAddress, err := c.chain.WrappedTokenGatewayV3Address()
	if err != nil {
		return nil, err
	}
	poolAddress, err := c.chain.AaveV3PoolAddress()
	if err != nil {
		return nil, err
	}
	gasTokenDecimals, err := c.chain.GasTokenDecimals()
	if err != nil {
		return nil, err
	}
	action := BuildBorrowETHAction(
		wrappedTokenGatewayAddress,
		poolAddress,
		c.ToWei(amount, gasTokenDecimals),
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

func (c *AaveV3Client) RepayWithPermit(ctx context.Context, coin config.Coin, amount decimal.Decimal) (*types.Receipt, error) {
	permitSupported, err := coin.PermitSupported(c.chain)
	if err != nil {
		return nil, err
	}
	if !permitSupported {
		return nil, fmt.Errorf("coin %v does not support permit on chain %v", coin, c.chain)
	}
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
	deadline := big.NewInt(time.Now().Add(time.Minute * 10).Unix())
	amountWei := c.ToWei(amount, decimals)

	permitAction, err := SignAndBuildPermitAction(
		ctx,
		c.conn,
		c.chain,
		coin,
		c.opts.From,
		poolAddress,
		amountWei,
		deadline,
		c.signer,
	)
	if err != nil {
		return nil, err
	}
	action := BuildRepayWithPermitAction(
		poolAddress,
		coinAddress,
		amountWei,
		c.opts.From,
		deadline,
		permitAction.v,
		permitAction.r,
		permitAction.s,
	)
	return executeAction(ctx, c.conn, c.opts, action)
}

func (c *AaveV3Client) DepositETH(ctx context.Context, amount decimal.Decimal) (*types.Receipt, error) {
	wrappedTokenGatewayAddress, err := c.chain.WrappedTokenGatewayV3Address()
	if err != nil {
		return nil, err
	}
	poolAddress, err := c.chain.AaveV3PoolAddress()
	if err != nil {
		return nil, err
	}
	gasTokenDecimals, err := c.chain.GasTokenDecimals()
	if err != nil {
		return nil, err
	}
	action := BuildDepositETHAction(
		wrappedTokenGatewayAddress,
		poolAddress,
		c.opts.From,
		0,
		c.ToWei(amount, gasTokenDecimals),
	)
	return executeAction(ctx, c.conn, c.opts, action)
}

func (c *AaveV3Client) WithdrawETH(ctx context.Context, amount decimal.Decimal) (*types.Receipt, error) {
	wrappedTokenGatewayAddress, err := c.chain.WrappedTokenGatewayV3Address()
	if err != nil {
		return nil, err
	}
	poolAddress, err := c.chain.AaveV3PoolAddress()
	if err != nil {
		return nil, err
	}
	gasTokenDecimals, err := c.chain.GasTokenDecimals()
	if err != nil {
		return nil, err
	}
	action := BuildWithdrawETHAction(
		wrappedTokenGatewayAddress,
		poolAddress,
		c.ToWei(amount, gasTokenDecimals),
		c.opts.From,
	)
	return executeAction(ctx, c.conn, c.opts, action)
}

func (c *AaveV3Client) WithdrawETHWithPermit(ctx context.Context, amount decimal.Decimal) (*types.Receipt, error) {
	wrappedTokenGatewayAddress, err := c.chain.WrappedTokenGatewayV3Address()
	if err != nil {
		return nil, err
	}
	poolAddress, err := c.chain.AaveV3PoolAddress()
	if err != nil {
		return nil, err
	}
	gasTokenDecimals, err := c.chain.GasTokenDecimals()
	if err != nil {
		return nil, err
	}
	aWETHDecimals, err := config.AWETH.Decimals()
	if err != nil {
		return nil, err
	}
	deadline := big.NewInt(time.Now().Add(time.Minute * 10).Unix())
	amountWei := c.ToWei(amount, aWETHDecimals)

	permitAction, err := SignAndBuildPermitAction(
		ctx,
		c.conn,
		c.chain,
		config.AWETH,
		c.opts.From,
		wrappedTokenGatewayAddress,
		amountWei,
		deadline,
		c.signer,
	)
	if err != nil {
		return nil, err
	}
	action := BuildWithdrawETHWithPermitAction(
		wrappedTokenGatewayAddress,
		poolAddress,
		c.ToWei(amount, gasTokenDecimals),
		c.opts.From,
		permitAction.deadline,
		permitAction.v,
		permitAction.r,
		permitAction.s,
	)
	return executeAction(ctx, c.conn, c.opts, action)
}

func (c *AaveV3Client) ApproveDelegation(ctx context.Context, asset config.Coin, delegatee common.Address, amount decimal.Decimal) (*types.Receipt, error) {
	debtToken, err := asset.DebtToken()
	if err != nil {
		return nil, err
	}
	debtTokenAddress, err := debtToken.Address(c.chain)
	if err != nil {
		return nil, err
	}
	debtTokenDecimals, err := debtToken.Decimals()
	if err != nil {
		return nil, err
	}
	action := BuildApproveDelegationAction(
		debtTokenAddress,
		delegatee,
		c.ToWei(amount, debtTokenDecimals),
	)
	return executeAction(ctx, c.conn, c.opts, action)
}

func (c *AaveV3Client) DelegationWithSig(ctx context.Context, asset config.Coin, delegatee common.Address, value decimal.Decimal) (*types.Receipt, error) {
	debtToken, err := asset.DebtToken()
	if err != nil {
		return nil, err
	}
	debtTokenDecimals, err := debtToken.Decimals()
	if err != nil {
		return nil, err
	}
	deadline := big.NewInt(time.Now().Add(time.Minute * 10).Unix())
	amountWei := c.ToWei(value, debtTokenDecimals)

	DelegationWithSig, err := SignAndBuildDelegationWithSigAction(
		ctx,
		c.conn,
		c.chain,
		debtToken,
		c.opts.From,
		delegatee,
		amountWei,
		deadline,
		c.signer,
	)
	if err != nil {
		return nil, err
	}

	return executeAction(ctx, c.conn, c.opts, DelegationWithSig)
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
