package contract

import (
	"context"
	_ "embed"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/shopspring/decimal"
	"github.com/tn606024/defi-simplify/bind/aave"
	"github.com/tn606024/defi-simplify/config"
)

//go:embed abi/aave/Pool.json
var aavePoolABI string

//go:embed abi/aave/WrappedTokenGatewayV3.json
var wrappedTokenGatewayV3ABI string

//go:embed abi/aave/DebtTokenBase.json
var debtTokenBaseABI string

type DataTypesUserAccountData struct {
	TotalCollateralBase         *big.Int
	TotalDebtBase               *big.Int
	AvailableBorrowsBase        *big.Int
	CurrentLiquidationThreshold *big.Int
	Ltv                         *big.Int
	HealthFactor                *big.Int
}

// AaveV3Interface defines the interface for Aave V3 operations
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
	GetUserAccountData(ctx context.Context, coin config.Coin) (*DataTypesUserAccountData, error)
}

// Action structs
type SupplyAction struct {
	BaseAction
	poolAddress  common.Address
	asset        common.Address
	amount       *big.Int
	onBehalfOf   common.Address
	referralCode uint16
}

type SupplyWithPermitAction struct {
	BaseAction
	poolAddress  common.Address
	asset        common.Address
	amount       *big.Int
	onBehalfOf   common.Address
	referralCode uint16
	deadline     *big.Int
	permitV      uint8
	permitR      [32]byte
	permitS      [32]byte
}

type WithdrawAction struct {
	BaseAction
	poolAddress common.Address
	asset       common.Address
	amount      *big.Int
	to          common.Address
}

type BorrowAction struct {
	BaseAction
	poolAddress      common.Address
	asset            common.Address
	amount           *big.Int
	interestRateMode *big.Int
	referralCode     uint16
	onBehalfOf       common.Address
}

type BorrowETHAction struct {
	BaseAction
	wrappedTokenGatewayAddress common.Address
	pool                       common.Address
	amount                     *big.Int
	referralCode               uint16
}

type RepayAction struct {
	BaseAction
	poolAddress      common.Address
	asset            common.Address
	amount           *big.Int
	interestRateMode *big.Int
	onBehalfOf       common.Address
}

type RepayWithPermitAction struct {
	BaseAction
	poolAddress      common.Address
	asset            common.Address
	amount           *big.Int
	interestRateMode *big.Int
	onBehalfOf       common.Address
	deadline         *big.Int
	permitV          uint8
	permitR          [32]byte
	permitS          [32]byte
}

type DepositETHAction struct {
	BaseAction
	wrappedTokenGatewayAddress common.Address
	pool                       common.Address
	onBehalfOf                 common.Address
	referral                   uint16
	amount                     *big.Int
}

type WithdrawETHAction struct {
	BaseAction
	wrappedTokenGatewayAddress common.Address
	pool                       common.Address
	amount                     *big.Int
	to                         common.Address
}

type WithdrawETHWithPermitAction struct {
	BaseAction
	wrappedTokenGatewayAddress common.Address
	pool                       common.Address
	amount                     *big.Int
	to                         common.Address
	deadline                   *big.Int
	permitV                    uint8
	permitR                    [32]byte
	permitS                    [32]byte
}

type ApproveDelegationAction struct {
	BaseAction
	asset     common.Address
	delegatee common.Address
	amount    *big.Int
}

type DelegationWithSigAction struct {
	BaseAction
	asset     common.Address
	delegator common.Address
	delegatee common.Address
	value     *big.Int
	deadline  *big.Int
	v         uint8
	r         [32]byte
	s         [32]byte
}

type GetReserveDataAction struct {
	BaseAction
	poolAddress common.Address
	asset       common.Address
}

type GetUserAccountDataAction struct {
	BaseAction
	poolAddress common.Address
	user        common.Address
}

// Client struct and constructors
type AaveV3Client struct {
	*BaseClientWithConverter
}

// NewAaveV3Client creates a new AaveV3Client
func NewAaveV3Client(base *BaseClient) AaveV3Interface {
	return &AaveV3Client{
		BaseClientWithConverter: &BaseClientWithConverter{
			BaseClient: base,
		},
	}
}

// Client methods
func (c *AaveV3Client) Supply(ctx context.Context, coin config.Coin, amount decimal.Decimal) (*types.Receipt, error) {
	action := BuildSupplyAction(
		c.chain.AaveV3PoolAddress(),
		coin.Address(c.chain),
		c.ToWei(amount, coin.Decimals()),
		c.opts.From,
	)

	return executeAction(ctx, c.conn, c.opts, action)
}

func (c *AaveV3Client) SupplyWithPermit(ctx context.Context, coin config.Coin, amount decimal.Decimal) (*types.Receipt, error) {
	if !coin.PermitSupported(c.chain) {
		return nil, fmt.Errorf("coin %v does not supported permit", coin)
	}
	deadline := big.NewInt(time.Now().Add(time.Minute * 10).Unix())
	amountWei := c.ToWei(amount, config.AWETH.Decimals())

	permitAction, err := SignAndBuildPermitAction(
		ctx,
		c.conn,
		c.chain,
		coin,
		c.opts.From,
		config.AaveV3PoolAddress[c.chain],
		amountWei,
		deadline,
		c.signer,
	)
	if err != nil {
		return nil, err
	}

	action := BuildSupplyWithPermitAction(
		c.chain.AaveV3PoolAddress(),
		coin.Address(c.chain),
		c.ToWei(amount, coin.Decimals()),
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
	action := BuildWithdrawAction(
		c.chain.AaveV3PoolAddress(),
		coin.Address(c.chain),
		c.ToWei(amount, coin.Decimals()),
		c.opts.From,
	)
	return executeAction(ctx, c.conn, c.opts, action)
}

func (c *AaveV3Client) Borrow(ctx context.Context, coin config.Coin, amount decimal.Decimal) (*types.Receipt, error) {
	action := BuildBorrowAction(
		c.chain.AaveV3PoolAddress(),
		coin.Address(c.chain),
		c.ToWei(amount, coin.Decimals()),
		c.opts.From,
	)

	return executeAction(ctx, c.conn, c.opts, action)
}

func (c *AaveV3Client) BorrowETH(ctx context.Context, amount decimal.Decimal) (*types.Receipt, error) {
	action := BuildBorrowETHAction(
		c.chain.WrappedTokenGatewayV3Address(),
		c.chain.AaveV3PoolAddress(),
		c.ToWei(amount, c.chain.GasTokenDecimals()),
	)
	return executeAction(ctx, c.conn, c.opts, action)
}

func (c *AaveV3Client) Repay(ctx context.Context, coin config.Coin, amount decimal.Decimal) (*types.Receipt, error) {
	action := BuildRepayAction(
		c.chain.AaveV3PoolAddress(),
		coin.Address(c.chain),
		c.ToWei(amount, coin.Decimals()),
		c.opts.From,
	)
	return executeAction(ctx, c.conn, c.opts, action)
}

func (c *AaveV3Client) RepayWithPermit(ctx context.Context, coin config.Coin, amount decimal.Decimal) (*types.Receipt, error) {
	if !coin.PermitSupported(c.chain) {
		return nil, fmt.Errorf("coin %v does not supported permit", coin)
	}
	deadline := big.NewInt(time.Now().Add(time.Minute * 10).Unix())
	amountWei := c.ToWei(amount, config.AWETH.Decimals())

	permitAction, err := SignAndBuildPermitAction(
		ctx,
		c.conn,
		c.chain,
		coin,
		c.opts.From,
		config.AaveV3PoolAddress[c.chain],
		amountWei,
		deadline,
		c.signer,
	)
	if err != nil {
		return nil, err
	}
	action := BuildRepayWithPermitAction(
		c.chain.AaveV3PoolAddress(),
		coin.Address(c.chain),
		c.ToWei(amount, coin.Decimals()),
		c.opts.From,
		deadline,
		permitAction.v,
		permitAction.r,
		permitAction.s,
	)
	return executeAction(ctx, c.conn, c.opts, action)
}

func (c *AaveV3Client) DepositETH(ctx context.Context, amount decimal.Decimal) (*types.Receipt, error) {
	action := BuildDepositETHAction(
		c.chain.WrappedTokenGatewayV3Address(),
		c.chain.AaveV3PoolAddress(),
		c.opts.From,
		0,
		c.ToWei(amount, c.chain.GasTokenDecimals()),
	)
	return executeAction(ctx, c.conn, c.opts, action)
}

func (c *AaveV3Client) WithdrawETH(ctx context.Context, amount decimal.Decimal) (*types.Receipt, error) {
	action := BuildWithdrawETHAction(
		c.chain.WrappedTokenGatewayV3Address(),
		c.chain.AaveV3PoolAddress(),
		c.ToWei(amount, c.chain.GasTokenDecimals()),
		c.opts.From,
	)
	return executeAction(ctx, c.conn, c.opts, action)
}

func (c *AaveV3Client) WithdrawETHWithPermit(ctx context.Context, amount decimal.Decimal) (*types.Receipt, error) {
	deadline := big.NewInt(time.Now().Add(time.Minute * 10).Unix())
	amountWei := c.ToWei(amount, config.AWETH.Decimals())

	permitAction, err := SignAndBuildPermitAction(
		ctx,
		c.conn,
		c.chain,
		config.AWETH,
		c.opts.From,
		config.WrappedTokenGatewayV3Address[c.chain],
		amountWei,
		deadline,
		c.signer,
	)
	if err != nil {
		return nil, err
	}
	action := BuildWithdrawETHWithPermitAction(
		c.chain.WrappedTokenGatewayV3Address(),
		c.chain.AaveV3PoolAddress(),
		c.ToWei(amount, c.chain.GasTokenDecimals()),
		c.opts.From,
		permitAction.deadline,
		permitAction.v,
		permitAction.r,
		permitAction.s,
	)
	return executeAction(ctx, c.conn, c.opts, action)
}

func (c *AaveV3Client) ApproveDelegation(ctx context.Context, asset config.Coin, delegatee common.Address, amount decimal.Decimal) (*types.Receipt, error) {
	debtToken := asset.DebtToken()
	action := BuildApproveDelegationAction(
		debtToken.Address(c.chain),
		delegatee,
		c.ToWei(amount, debtToken.Decimals()),
	)
	return executeAction(ctx, c.conn, c.opts, action)
}

func (c *AaveV3Client) DelegationWithSig(ctx context.Context, asset config.Coin, delegatee common.Address, value decimal.Decimal) (*types.Receipt, error) {
	debtToken := asset.DebtToken()
	deadline := big.NewInt(time.Now().Add(time.Minute * 10).Unix())
	amountWei := c.ToWei(value, debtToken.Decimals())

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
	action := BuildGetReserveDataAction(
		c.chain.AaveV3PoolAddress(),
		coin.Address(c.chain),
	)
	return getReserveData(c.conn, action)
}

func (c *AaveV3Client) GetUserAccountData(ctx context.Context, coin config.Coin) (*DataTypesUserAccountData, error) {
	action := BuildGetUserAccountDataAction(
		c.chain.AaveV3PoolAddress(),
		c.opts.From,
	)
	return getUserAccountData(c.conn, action)
}

func getReserveData(conn EthereumClient, action *GetReserveDataAction) (*aave.DataTypesReserveData, error) {
	pool, err := aave.NewPool(action.poolAddress, conn)
	if err != nil {
		return nil, err
	}
	reserveData, err := pool.GetReserveData(nil, action.asset)
	if err != nil {
		fmt.Println("Error getting reserve data:", err)
		return nil, err
	}
	return &reserveData, nil
}

func getUserAccountData(conn EthereumClient, action *GetUserAccountDataAction) (*DataTypesUserAccountData, error) {
	pool, err := aave.NewPool(action.poolAddress, conn)
	if err != nil {
		return nil, err
	}
	userAccountData, err := pool.GetUserAccountData(nil, action.user)
	if err != nil {
		return nil, err
	}
	return &DataTypesUserAccountData{
		TotalCollateralBase:         userAccountData.TotalCollateralBase,
		TotalDebtBase:               userAccountData.TotalDebtBase,
		AvailableBorrowsBase:        userAccountData.AvailableBorrowsBase,
		CurrentLiquidationThreshold: userAccountData.CurrentLiquidationThreshold,
		Ltv:                         userAccountData.Ltv,
		HealthFactor:                userAccountData.HealthFactor,
	}, nil
}

// 6. Transaction creation functions
func (a *SupplyAction) ToTransaction(ctx context.Context, conn EthereumClient, opt *bind.TransactOpts) (*types.Transaction, error) {
	pool, err := aave.NewPool(a.poolAddress, conn)
	if err != nil {
		return nil, err
	}
	return pool.Supply(opt, a.asset, a.amount, a.onBehalfOf, a.referralCode)
}

func (a *SupplyWithPermitAction) ToTransaction(ctx context.Context, conn EthereumClient, opt *bind.TransactOpts) (*types.Transaction, error) {
	pool, err := aave.NewPool(a.poolAddress, conn)
	if err != nil {
		return nil, err
	}
	return pool.SupplyWithPermit(opt, a.asset, a.amount, a.onBehalfOf, a.referralCode, a.deadline, a.permitV, a.permitR, a.permitS)
}

func (a *WithdrawAction) ToTransaction(ctx context.Context, conn EthereumClient, opt *bind.TransactOpts) (*types.Transaction, error) {
	pool, err := aave.NewPool(a.poolAddress, conn)
	if err != nil {
		return nil, err
	}
	return pool.Withdraw(opt, a.asset, a.amount, a.to)
}

func (a *BorrowAction) ToTransaction(ctx context.Context, conn EthereumClient, opt *bind.TransactOpts) (*types.Transaction, error) {
	pool, err := aave.NewPool(a.poolAddress, conn)
	if err != nil {
		return nil, err
	}
	return pool.Borrow(opt, a.asset, a.amount, a.interestRateMode, a.referralCode, a.onBehalfOf)
}

func (a *BorrowETHAction) ToTransaction(ctx context.Context, conn EthereumClient, opt *bind.TransactOpts) (*types.Transaction, error) {
	wrappedTokenGatewayV3, err := aave.NewWrappedTokenGatewayV3(a.wrappedTokenGatewayAddress, conn)
	if err != nil {
		return nil, err
	}
	return wrappedTokenGatewayV3.BorrowETH(opt, a.pool, a.amount, a.referralCode)
}

func (a *RepayAction) ToTransaction(ctx context.Context, conn EthereumClient, opt *bind.TransactOpts) (*types.Transaction, error) {
	pool, err := aave.NewPool(a.poolAddress, conn)
	if err != nil {
		return nil, err
	}
	return pool.Repay(opt, a.asset, a.amount, a.interestRateMode, a.onBehalfOf)
}

func (a *RepayWithPermitAction) ToTransaction(ctx context.Context, conn EthereumClient, opt *bind.TransactOpts) (*types.Transaction, error) {
	pool, err := aave.NewPool(a.poolAddress, conn)
	if err != nil {
		return nil, err
	}
	return pool.RepayWithPermit(opt, a.asset, a.amount, a.interestRateMode, a.onBehalfOf, a.deadline, a.permitV, a.permitR, a.permitS)
}

func (a *DepositETHAction) ToTransaction(ctx context.Context, conn EthereumClient, opt *bind.TransactOpts) (*types.Transaction, error) {
	wrappedTokenGatewayV3, err := aave.NewWrappedTokenGatewayV3(a.wrappedTokenGatewayAddress, conn)
	if err != nil {
		return nil, err
	}
	// Set the ETH value in the transaction options
	opt.Value = a.amount
	tx, err := wrappedTokenGatewayV3.DepositETH(opt, a.pool, a.onBehalfOf, a.referral)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (a *WithdrawETHAction) ToTransaction(ctx context.Context, conn EthereumClient, opt *bind.TransactOpts) (*types.Transaction, error) {
	wrappedTokenGatewayV3, err := aave.NewWrappedTokenGatewayV3(a.wrappedTokenGatewayAddress, conn)
	if err != nil {
		return nil, err
	}
	return wrappedTokenGatewayV3.WithdrawETH(opt, a.pool, a.amount, a.to)
}

func (a *WithdrawETHWithPermitAction) ToTransaction(ctx context.Context, conn EthereumClient, opt *bind.TransactOpts) (*types.Transaction, error) {
	wrappedTokenGatewayV3, err := aave.NewWrappedTokenGatewayV3(a.wrappedTokenGatewayAddress, conn)
	if err != nil {
		return nil, err
	}
	return wrappedTokenGatewayV3.WithdrawETHWithPermit(opt, a.pool, a.amount, a.to, a.deadline, a.permitV, a.permitR, a.permitS)
}

func (a *ApproveDelegationAction) ToTransaction(ctx context.Context, conn EthereumClient, opt *bind.TransactOpts) (*types.Transaction, error) {
	debtTokenBase, err := aave.NewDebtTokenBase(a.asset, conn)
	if err != nil {
		return nil, err
	}
	return debtTokenBase.ApproveDelegation(opt, a.delegatee, a.amount)
}

func (a *DelegationWithSigAction) ToTransaction(ctx context.Context, conn EthereumClient, opt *bind.TransactOpts) (*types.Transaction, error) {
	debtTokenBase, err := aave.NewDebtTokenBase(a.asset, conn)
	if err != nil {
		return nil, err
	}
	return debtTokenBase.DelegationWithSig(opt, a.delegator, a.delegatee, a.value, a.deadline, a.v, a.r, a.s)
}

// 7. Multicall implementations
func (a *SupplyAction) ToData(ctx context.Context, conn EthereumClient, opt *bind.TransactOpts) (common.Address, []byte, error) {
	parsed, err := abi.JSON(strings.NewReader(aavePoolABI))
	if err != nil {
		return common.Address{}, nil, err
	}
	data, err := parsed.Pack("supply", a.asset, a.amount, a.onBehalfOf, a.referralCode)
	if err != nil {
		return common.Address{}, nil, err
	}
	return a.poolAddress, data, nil
}

func (a *SupplyWithPermitAction) ToData(ctx context.Context, conn EthereumClient, opt *bind.TransactOpts) (common.Address, []byte, error) {
	parsed, err := abi.JSON(strings.NewReader(aavePoolABI))
	if err != nil {
		return common.Address{}, nil, err
	}
	data, err := parsed.Pack("supplyWithPermit", a.asset, a.amount, a.onBehalfOf, a.referralCode, a.deadline, a.permitV, a.permitR, a.permitS)
	if err != nil {
		return common.Address{}, nil, err
	}
	return a.poolAddress, data, nil
}

func (a *WithdrawAction) ToData(ctx context.Context, conn EthereumClient, opt *bind.TransactOpts) (common.Address, []byte, error) {
	parsed, err := abi.JSON(strings.NewReader(aavePoolABI))
	if err != nil {
		return common.Address{}, nil, err
	}
	data, err := parsed.Pack("withdraw", a.asset, a.amount, a.to)
	if err != nil {
		return common.Address{}, nil, err
	}
	return a.poolAddress, data, nil
}

func (a *BorrowAction) ToData(ctx context.Context, conn EthereumClient, opt *bind.TransactOpts) (common.Address, []byte, error) {
	parsed, err := abi.JSON(strings.NewReader(aavePoolABI))
	if err != nil {
		return common.Address{}, nil, err
	}
	data, err := parsed.Pack("borrow", a.asset, a.amount, a.interestRateMode, a.referralCode, a.onBehalfOf)
	if err != nil {
		return common.Address{}, nil, err
	}
	return a.poolAddress, data, nil
}

func (a *BorrowETHAction) ToData(ctx context.Context, conn EthereumClient, opt *bind.TransactOpts) (common.Address, []byte, error) {
	parsed, err := abi.JSON(strings.NewReader(wrappedTokenGatewayV3ABI))
	if err != nil {
		return common.Address{}, nil, err
	}
	data, err := parsed.Pack("borrowETH", a.pool, a.amount, a.referralCode)
	if err != nil {
		return common.Address{}, nil, err
	}
	return a.wrappedTokenGatewayAddress, data, nil
}

func (a *RepayAction) ToData(ctx context.Context, conn EthereumClient, opt *bind.TransactOpts) (common.Address, []byte, error) {
	parsed, err := abi.JSON(strings.NewReader(aavePoolABI))
	if err != nil {
		return common.Address{}, nil, err
	}
	data, err := parsed.Pack("repay", a.asset, a.amount, a.interestRateMode, a.onBehalfOf)
	if err != nil {
		return common.Address{}, nil, err
	}
	return a.poolAddress, data, nil
}

func (a *RepayWithPermitAction) ToData(ctx context.Context, conn EthereumClient, opt *bind.TransactOpts) (common.Address, []byte, error) {
	parsed, err := abi.JSON(strings.NewReader(aavePoolABI))
	if err != nil {
		return common.Address{}, nil, err
	}
	data, err := parsed.Pack("repayWithPermit", a.asset, a.amount, a.interestRateMode, a.onBehalfOf, a.deadline, a.permitV, a.permitR, a.permitS)
	if err != nil {
		return common.Address{}, nil, err
	}
	return a.poolAddress, data, nil
}

func (a *DepositETHAction) ToData(ctx context.Context, conn EthereumClient, opt *bind.TransactOpts) (common.Address, []byte, error) {
	parsed, err := abi.JSON(strings.NewReader(wrappedTokenGatewayV3ABI))
	if err != nil {
		return common.Address{}, nil, err
	}
	// Set the ETH value in the transaction options
	opt.Value = a.amount
	data, err := parsed.Pack("depositETH", a.pool, a.onBehalfOf, a.referral)
	if err != nil {
		return common.Address{}, nil, err
	}
	return a.wrappedTokenGatewayAddress, data, nil
}

func (a *WithdrawETHAction) ToData(ctx context.Context, conn EthereumClient, opt *bind.TransactOpts) (common.Address, []byte, error) {
	parsed, err := abi.JSON(strings.NewReader(wrappedTokenGatewayV3ABI))
	if err != nil {
		return common.Address{}, nil, err
	}
	data, err := parsed.Pack("withdrawETH", a.pool, a.amount, a.to)
	if err != nil {
		return common.Address{}, nil, err
	}
	return a.wrappedTokenGatewayAddress, data, nil
}

func (a *WithdrawETHWithPermitAction) ToData(ctx context.Context, conn EthereumClient, opt *bind.TransactOpts) (common.Address, []byte, error) {
	parsed, err := abi.JSON(strings.NewReader(wrappedTokenGatewayV3ABI))
	if err != nil {
		return common.Address{}, nil, err
	}
	data, err := parsed.Pack("withdrawETHWithPermit", a.pool, a.amount, a.to, a.deadline, a.permitV, a.permitR, a.permitS)
	if err != nil {
		return common.Address{}, nil, err
	}
	return a.wrappedTokenGatewayAddress, data, nil
}

func (a *ApproveDelegationAction) ToData(ctx context.Context, conn EthereumClient, opt *bind.TransactOpts) (common.Address, []byte, error) {
	parsed, err := abi.JSON(strings.NewReader(debtTokenBaseABI))
	if err != nil {
		return common.Address{}, nil, err
	}
	data, err := parsed.Pack("approveDelegation", a.delegatee, a.amount)
	if err != nil {
		return common.Address{}, nil, err
	}
	return a.asset, data, nil
}

func (a *DelegationWithSigAction) ToData(ctx context.Context, conn EthereumClient, opt *bind.TransactOpts) (common.Address, []byte, error) {
	parsed, err := abi.JSON(strings.NewReader(debtTokenBaseABI))
	if err != nil {
		return common.Address{}, nil, err
	}
	data, err := parsed.Pack("delegationWithSig", a.delegator, a.delegatee, a.value, a.deadline, a.v, a.r, a.s)
	if err != nil {
		return common.Address{}, nil, err
	}
	return a.asset, data, nil
}

func (a *GetReserveDataAction) ToData(ctx context.Context, conn EthereumClient, opt *bind.TransactOpts) (common.Address, []byte, error) {
	parsed, err := abi.JSON(strings.NewReader(aavePoolABI))
	if err != nil {
		return common.Address{}, nil, err
	}
	data, err := parsed.Pack("getReserveData", a.asset)
	if err != nil {
		return common.Address{}, nil, err
	}
	return a.poolAddress, data, nil
}

func (a *GetUserAccountDataAction) ToData(ctx context.Context, conn EthereumClient, opt *bind.TransactOpts) (common.Address, []byte, error) {
	parsed, err := abi.JSON(strings.NewReader(aavePoolABI))
	if err != nil {
		return common.Address{}, nil, err
	}
	data, err := parsed.Pack("getUserAccountData", a.user)
	if err != nil {
		return common.Address{}, nil, err
	}
	return a.poolAddress, data, nil
}
