package contract

import (
	_ "embed"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
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
