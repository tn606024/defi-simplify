package contract

import (
	"fmt"

	"github.com/tn606024/defi-simplify/bind/aave"
)

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
	userReserveData, err := protocolDataProvider.GetUserReserveData(nil, action.asset, action.user)
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
