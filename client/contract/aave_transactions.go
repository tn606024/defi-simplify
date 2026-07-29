package contract

import (
	"context"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/tn606024/defi-simplify/bind/aave"
)

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
