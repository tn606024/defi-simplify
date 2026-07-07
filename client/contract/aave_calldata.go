package contract

import (
	"context"
	"strings"

	"github.com/ethereum/go-ethereum"
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

func (a *DepositETHAction) ToCall(ctx context.Context, conn EthereumClient, opt *bind.TransactOpts) (*Call, error) {
	parsed, err := abi.JSON(strings.NewReader(wrappedTokenGatewayV3ABI))
	if err != nil {
		return nil, err
	}
	data, err := parsed.Pack("depositETH", a.pool, a.onBehalfOf, a.referral)
	if err != nil {
		return nil, err
	}
	return &Call{
		Target: a.wrappedTokenGatewayAddress,
		Value:  a.amount,
		Data:   data,
	}, nil
}

func (a *DepositETHAction) ToCallMsg(ctx context.Context, conn EthereumClient, opt *bind.TransactOpts) (*ethereum.CallMsg, error) {
	call, err := a.ToCall(ctx, conn, opt)
	if err != nil {
		return nil, err
	}
	return callToCallMsg(call), nil
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
