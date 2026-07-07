package contract

import (
	"context"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

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
