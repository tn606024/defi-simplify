package contract

import (
	"context"
	_ "embed"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/tn606024/defi-simplify/bind/aave"
)

//go:embed abi/aave/AaveProtocolDataProvider.json
var aaveProtocolDataProviderABI string

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

type DataTypesUserReserveData struct {
	CurrentATokenBalance     *big.Int
	CurrentStableDebt        *big.Int
	CurrentVariableDebt      *big.Int
	PrincipalStableDebt      *big.Int
	ScaledVariableDebt       *big.Int
	StableBorrowRate         *big.Int
	LiquidityRate            *big.Int
	StableRateLastUpdated    *big.Int
	UsageAsCollateralEnabled bool
}

type TokenReserveData struct {
	TokenAddress    common.Address
	UserReserveData DataTypesUserReserveData
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

func getAllReservesTokens(conn EthereumClient, action *GetAllReservesTokensAction) ([]aave.IPoolDataProviderTokenData, error) {
	protocolDataProvider, err := aave.NewAaveProtocolDataProvider(action.protocolDataProviderAddress, conn)
	if err != nil {
		return nil, err
	}
	allReservesToken, err := protocolDataProvider.GetAllReservesTokens(nil)
	if err != nil {
		return nil, err
	}
	return allReservesToken, nil
}

func getUserReserveData(conn EthereumClient, action *GetUserReserveDataAction) (*DataTypesUserReserveData, error) {
	protocolDataProvider, err := aave.NewAaveProtocolDataProvider(action.protocolDataProviderAddress, conn)
	if err != nil {
		return nil, err
	}
	userReserveData, err := protocolDataProvider.GetUserReserveData(nil, action.user, action.asset)
	if err != nil {
		return nil, err
	}
	return &DataTypesUserReserveData{
		CurrentATokenBalance:     userReserveData.CurrentATokenBalance,
		CurrentStableDebt:        userReserveData.CurrentStableDebt,
		CurrentVariableDebt:      userReserveData.CurrentVariableDebt,
		PrincipalStableDebt:      userReserveData.PrincipalStableDebt,
		ScaledVariableDebt:       userReserveData.ScaledVariableDebt,
		StableBorrowRate:         userReserveData.StableBorrowRate,
		LiquidityRate:            userReserveData.LiquidityRate,
		StableRateLastUpdated:    userReserveData.StableRateLastUpdated,
		UsageAsCollateralEnabled: userReserveData.UsageAsCollateralEnabled,
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
