package contract

import (
	"context"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

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

func (a *GetAllReservesTokensAction) ToData(ctx context.Context, conn EthereumClient, opt *bind.TransactOpts) (common.Address, []byte, error) {
	parsed, err := abi.JSON(strings.NewReader(aaveProtocolDataProviderABI))
	if err != nil {
		return common.Address{}, nil, err
	}
	data, err := parsed.Pack("getAllReservesTokens")
	if err != nil {
		return common.Address{}, nil, err
	}
	return a.protocolDataProviderAddress, data, nil
}

func (a *GetUserReserveDataAction) ToData(ctx context.Context, conn EthereumClient, opt *bind.TransactOpts) (common.Address, []byte, error) {
	parsed, err := abi.JSON(strings.NewReader(aaveProtocolDataProviderABI))
	if err != nil {
		return common.Address{}, nil, err
	}
	data, err := parsed.Pack("getUserReserveData", a.asset, a.user)
	if err != nil {
		return common.Address{}, nil, err
	}
	return a.protocolDataProviderAddress, data, nil
}
