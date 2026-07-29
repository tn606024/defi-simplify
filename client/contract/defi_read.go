package contract

import (
	"context"

	"github.com/shopspring/decimal"
	"github.com/tn606024/defi-simplify/config"
)

func (c *DefiClient) GetAllReservesTokensAndGetUserReserveData(ctx context.Context) ([]TokenReserveData, error) {
	allReservesTokens, err := c.Aave.GetAllReservesTokens(ctx)
	if err != nil {
		return nil, err
	}

	tokenReserveData := make([]TokenReserveData, len(allReservesTokens))
	for i, token := range allReservesTokens {
		userReserveData, err := c.Aave.GetUserReserveData(ctx, token.TokenAddress)
		if err != nil {
			return nil, err
		}

		tokenReserveData[i] = TokenReserveData{
			TokenAddress:    token.TokenAddress,
			UserReserveData: *userReserveData,
		}
	}

	return tokenReserveData, nil
}

// GetMultipleCoinBalances reads balances through the legacy config.Coin map.
//
// Deprecated: resolve token.Token values and build address-based balance calls instead.
func (c *DefiClient) GetMultipleCoinBalances(ctx context.Context, coins []config.Coin) ([]decimal.Decimal, error) {
	balances := make([]decimal.Decimal, 0, len(coins))
	for _, coin := range coins {
		balance, err := c.ERC20.BalanceOf(c.chain, coin)
		if err != nil {
			return nil, err
		}
		balances = append(balances, balance)
	}
	return balances, nil
}
