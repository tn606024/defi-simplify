package contract

import (
	"context"
	_ "embed"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/shopspring/decimal"
	"github.com/tn606024/defi-simplify/bind/erc20"
	"github.com/tn606024/defi-simplify/config"
)

//go:embed abi/erc20/ERC20.json
var erc20ABI string

//go:embed abi/erc20/IERC20Permit.json
var erc20PermitABI string

// ERC20Interface defines the interface for ERC20 token operations
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

// Action structs
type TransferAction struct {
	BaseAction
	token  common.Address // Token contract address
	to     common.Address // Recipient address
	amount *big.Int       // Amount to transfer
}

type ApproveAction struct {
	BaseAction
	token   common.Address // Token contract address
	spender common.Address // Address to approve
	amount  *big.Int       // Amount to approve
}

type TransferFromAction struct {
	BaseAction
	token  common.Address
	from   common.Address
	to     common.Address
	amount *big.Int
}

type BalanceOfAction struct {
	BaseAction
	token common.Address
	user  common.Address
}

type PermitAction struct {
	BaseAction
	token    common.Address
	owner    common.Address
	spender  common.Address
	amount   *big.Int
	deadline *big.Int
	v        uint8
	r        [32]byte
	s        [32]byte
}

type NoncesAction struct {
	BaseAction
	token common.Address
	owner common.Address
}

// Client struct and constructors
type ERC20Client struct {
	*BaseClientWithConverter
}

// NewERC20Client creates a new ERC20Client
func NewERC20Client(base *BaseClient) ERC20Interface {
	return &ERC20Client{
		BaseClientWithConverter: &BaseClientWithConverter{
			BaseClient: base,
		},
	}
}

// Client methods
func (c *ERC20Client) BalanceOf(chain config.Chain, coin config.Coin) (decimal.Decimal, error) {
	action := BuildBalanceOfAction(coin.Address(chain), c.opts.From)
	balance, err := balanceOf(c.conn, action)
	if err != nil {
		return decimal.Zero, err
	}
	return c.FromWei(balance, coin.Decimals()), nil
}

func (c *ERC20Client) Transfer(ctx context.Context, coin config.Coin, to common.Address, amount decimal.Decimal) (*types.Receipt, error) {
	action := BuildTransferAction(
		coin.Address(c.chain),
		to,
		c.ToWei(amount, coin.Decimals()),
	)
	return executeAction(ctx, c.conn, c.opts, action)
}

func (c *ERC20Client) Approve(ctx context.Context, coin config.Coin, spender common.Address, amount decimal.Decimal) (*types.Receipt, error) {
	action := BuildApproveAction(
		coin.Address(c.chain),
		spender,
		c.ToWei(amount, coin.Decimals()),
	)
	return executeAction(ctx, c.conn, c.opts, action)
}

func (c *ERC20Client) TransferFrom(ctx context.Context, coin config.Coin, from common.Address, to common.Address, amount decimal.Decimal) (*types.Receipt, error) {
	action := BuildTransferFromAction(
		coin.Address(c.chain),
		from,
		to,
		c.ToWei(amount, coin.Decimals()),
	)
	return executeAction(ctx, c.conn, c.opts, action)
}

func (c *ERC20Client) Permit(ctx context.Context, coin config.Coin, spender common.Address, amount decimal.Decimal, deadline *big.Int) (*types.Receipt, error) {
	action, err := SignAndBuildPermitAction(
		ctx,
		c.conn,
		c.chain,
		coin,
		c.opts.From,
		spender,
		c.ToWei(amount, coin.Decimals()),
		deadline,
		c.signer,
	)
	if err != nil {
		return nil, err
	}
	return executeAction(ctx, c.conn, c.opts, action)
}

func (c *ERC20Client) Nonces(ctx context.Context, coin config.Coin, owner common.Address) (*big.Int, error) {
	action := BuildNoncesAction(coin.Address(c.chain), owner)
	return nonces(c.conn, action)
}

func (c *ERC20Client) Allowance(ctx context.Context, coin config.Coin, spender common.Address) (decimal.Decimal, error) {
	erc20Instance, err := erc20.NewErc20(coin.Address(c.chain), c.conn)
	if err != nil {
		return decimal.Zero, err
	}
	allowance, err := erc20Instance.Allowance(nil, c.opts.From, spender)
	if err != nil {
		return decimal.Zero, err
	}
	return c.FromWei(allowance, coin.Decimals()), nil
}

func balanceOf(conn EthereumClient, action *BalanceOfAction) (*big.Int, error) {
	erc20Instance, err := erc20.NewErc20(action.token, conn)
	if err != nil {
		return nil, err
	}
	balance, err := erc20Instance.BalanceOf(nil, action.user)
	if err != nil {
		return nil, err
	}
	return balance, nil
}

func nonces(conn EthereumClient, action *NoncesAction) (*big.Int, error) {
	erc20Instance, err := erc20.NewIErc20WithPermit(action.token, conn)
	if err != nil {
		return nil, err
	}
	return erc20Instance.Nonces(nil, action.owner)
}

// Action interface implementations
func (a *TransferAction) ToData(ctx context.Context, conn EthereumClient, opt *bind.TransactOpts) (common.Address, []byte, error) {
	parsed, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		return common.Address{}, nil, err
	}

	data, err := parsed.Pack("transfer", a.to, a.amount)
	if err != nil {
		return common.Address{}, nil, err
	}

	return a.token, data, nil
}

func (a *ApproveAction) ToData(ctx context.Context, conn EthereumClient, opt *bind.TransactOpts) (common.Address, []byte, error) {
	parsed, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		return common.Address{}, nil, err
	}

	data, err := parsed.Pack("approve", a.spender, a.amount)
	if err != nil {
		return common.Address{}, nil, err
	}

	return a.token, data, nil
}

func (a *TransferFromAction) ToData(ctx context.Context, conn EthereumClient, opt *bind.TransactOpts) (common.Address, []byte, error) {
	parsed, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		return common.Address{}, nil, err
	}

	data, err := parsed.Pack("transferFrom", a.from, a.to, a.amount)
	if err != nil {
		return common.Address{}, nil, err
	}

	return a.token, data, nil
}

func (a *BalanceOfAction) ToData(ctx context.Context, conn EthereumClient, opt *bind.TransactOpts) (common.Address, []byte, error) {
	parsed, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		return common.Address{}, nil, err
	}

	data, err := parsed.Pack("balanceOf", a.user)
	if err != nil {
		return common.Address{}, nil, err
	}

	return a.token, data, nil
}

func (a *PermitAction) ToData(ctx context.Context, conn EthereumClient, opt *bind.TransactOpts) (common.Address, []byte, error) {
	parsed, err := abi.JSON(strings.NewReader(erc20PermitABI))
	if err != nil {
		return common.Address{}, nil, err
	}
	data, err := parsed.Pack("permit", a.owner, a.spender, a.amount, a.deadline, a.v, a.r, a.s)
	if err != nil {
		return common.Address{}, nil, err
	}
	return a.token, data, nil
}

func (a *NoncesAction) ToData(ctx context.Context, conn EthereumClient, opt *bind.TransactOpts) (common.Address, []byte, error) {
	parsed, err := abi.JSON(strings.NewReader(erc20PermitABI))
	if err != nil {
		return common.Address{}, nil, err
	}

	data, err := parsed.Pack("nonces", a.owner)
	if err != nil {
		return common.Address{}, nil, err
	}

	return a.token, data, nil
}

// Add ToTx implementations for each action
func (a *TransferAction) ToTransaction(ctx context.Context, conn EthereumClient, opt *bind.TransactOpts) (*types.Transaction, error) {
	token, err := erc20.NewErc20(a.token, conn)
	if err != nil {
		return nil, err
	}
	return token.Transfer(opt, a.to, a.amount)
}

func (a *ApproveAction) ToTransaction(ctx context.Context, conn EthereumClient, opt *bind.TransactOpts) (*types.Transaction, error) {
	token, err := erc20.NewErc20(a.token, conn)
	if err != nil {
		return nil, err
	}
	return token.Approve(opt, a.spender, a.amount)
}

func (a *TransferFromAction) ToTransaction(ctx context.Context, conn EthereumClient, opt *bind.TransactOpts) (*types.Transaction, error) {
	token, err := erc20.NewErc20(a.token, conn)
	if err != nil {
		return nil, err
	}
	return token.TransferFrom(opt, a.from, a.to, a.amount)
}

func (a *PermitAction) ToTransaction(ctx context.Context, conn EthereumClient, opt *bind.TransactOpts) (*types.Transaction, error) {
	token, err := erc20.NewIErc20WithPermit(a.token, conn)
	if err != nil {
		return nil, err
	}
	return token.Permit(opt, a.owner, a.spender, a.amount, a.deadline, a.v, a.r, a.s)
}
