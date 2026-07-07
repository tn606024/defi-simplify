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
