package contract

import (
	"context"
	_ "embed"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/tn606024/defi-simplify/bind/erc20"
)

//go:embed abi/erc20/ERC20.json
var erc20ABI string

//go:embed abi/erc20/IERC20Permit.json
var erc20PermitABI string

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
