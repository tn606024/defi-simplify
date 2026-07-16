package contract

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/shopspring/decimal"
	"github.com/tn606024/defi-simplify/config"
)

func (c *DefiClient) GetAllReservesTokensAndGetUserReserveData(ctx context.Context) ([]TokenReserveData, error) {
	allReservesTokens, err := c.Aave.GetAllReservesTokens(ctx)
	if err != nil {
		return nil, err
	}

	from := c.opts.From
	protocolDataProviderAddress, err := c.chain.AaveProtocolDataProviderAddress()
	if err != nil {
		return nil, err
	}
	actions := make([]Action, 0, len(allReservesTokens))
	tokenReserveData := make([]TokenReserveData, len(allReservesTokens))
	for _, token := range allReservesTokens {
		action := BuildGetUserReserveDataAction(protocolDataProviderAddress, token.TokenAddress, from)
		actions = append(actions, action)
	}
	results, err := c.BaseClient.ExecuteMulticalls(ctx, actions)
	if err != nil {
		return nil, err
	}

	for i, result := range results {
		if !result.Success {
			return nil, fmt.Errorf("failed to get user reserve data for token %s", allReservesTokens[i].TokenAddress.Hex())
		}

		parsed, err := abi.JSON(strings.NewReader(aaveProtocolDataProviderABI))
		if err != nil {
			return nil, err
		}

		var userReserveData struct {
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

		err = parsed.UnpackIntoInterface(&userReserveData, "getUserReserveData", result.ReturnData)
		if err != nil {
			return nil, err
		}

		tokenReserveData[i] = TokenReserveData{
			TokenAddress: allReservesTokens[i].TokenAddress,
			UserReserveData: DataTypesUserReserveData{
				CurrentATokenBalance:     userReserveData.CurrentATokenBalance,
				CurrentStableDebt:        userReserveData.CurrentStableDebt,
				CurrentVariableDebt:      userReserveData.CurrentVariableDebt,
				PrincipalStableDebt:      userReserveData.PrincipalStableDebt,
				ScaledVariableDebt:       userReserveData.ScaledVariableDebt,
				StableBorrowRate:         userReserveData.StableBorrowRate,
				LiquidityRate:            userReserveData.LiquidityRate,
				StableRateLastUpdated:    userReserveData.StableRateLastUpdated,
				UsageAsCollateralEnabled: userReserveData.UsageAsCollateralEnabled,
			},
		}
	}

	return tokenReserveData, nil
}

// GetMultipleCoinBalances reads balances through the legacy config.Coin map.
//
// Deprecated: resolve token.Token values and build address-based balance calls instead.
func (c *DefiClient) GetMultipleCoinBalances(ctx context.Context, coins []config.Coin) ([]decimal.Decimal, error) {
	actions := make([]Action, 0, len(coins))
	for _, coin := range coins {
		coinAddress, err := coin.Address(c.chain)
		if err != nil {
			return nil, err
		}
		action := BuildBalanceOfAction(coinAddress, c.opts.From)
		actions = append(actions, action)
	}
	results, err := c.BaseClient.ExecuteMulticalls(ctx, actions)
	if err != nil {
		return nil, err
	}

	balances := make([]decimal.Decimal, 0, len(coins))
	for i, result := range results {
		if !result.Success {
			coinAddress, err := coins[i].Address(c.chain)
			if err != nil {
				return nil, err
			}
			return nil, fmt.Errorf("failed to get balance for coin %s", coinAddress.Hex())
		}
		abi, err := abi.JSON(strings.NewReader(erc20ABI))
		if err != nil {
			return nil, err
		}
		var balance *big.Int
		err = abi.UnpackIntoInterface(&balance, "balanceOf", result.ReturnData)
		if err != nil {
			return nil, err
		}
		decimals, err := coins[i].Decimals()
		if err != nil {
			return nil, err
		}
		balances = append(balances, c.FromWei(balance, decimals))
	}
	return balances, nil
}
