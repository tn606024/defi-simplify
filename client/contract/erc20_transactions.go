package contract

import (
	"context"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/tn606024/defi-simplify/bind/erc20"
)

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
