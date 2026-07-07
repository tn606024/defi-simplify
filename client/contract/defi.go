package contract

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/shopspring/decimal"
	"github.com/tn606024/defi-simplify/config"
	"github.com/tn606024/defi-simplify/helper"
)

// DefiClient composes all DeFi related clients
type DefiClient struct {
	*BaseClientWithConverter
	ERC20 ERC20Interface
	Aave  AaveV3Interface
}

// NewDefiClient creates a new DefiClient with all sub-clients
func NewDefiClient(opts *bind.TransactOpts, conn EthereumClient, signer *helper.MsgSigner, chain config.Chain) *DefiClient {
	base := &BaseClient{
		opts:   opts,
		conn:   conn,
		signer: signer,
		chain:  chain,
	}

	return &DefiClient{
		BaseClientWithConverter: &BaseClientWithConverter{
			BaseClient: base,
		},
		ERC20: NewERC20Client(base),
		Aave:  NewAaveV3Client(base),
	}
}

func (c *DefiClient) SupplyAndBorrowAaveV3Coin(ctx context.Context, coin config.Coin, supplyAmount decimal.Decimal, borrowAmount decimal.Decimal) (*types.Receipt, error) {
	debtToken, err := coin.DebtToken()
	if err != nil {
		return nil, err
	}
	coinAddress, err := coin.Address(c.chain)
	if err != nil {
		return nil, err
	}
	coinDecimals, err := coin.Decimals()
	if err != nil {
		return nil, err
	}
	multicallAddress, err := c.chain.MulticallAddress()
	if err != nil {
		return nil, err
	}
	aaveV3PoolAddress, err := c.chain.AaveV3PoolAddress()
	if err != nil {
		return nil, err
	}
	supplyAmountWei := c.ToWei(supplyAmount, coinDecimals)
	borrowAmountWei := c.ToWei(borrowAmount, coinDecimals)
	deadline := big.NewInt(time.Now().Add(time.Minute * 10).Unix())
	permitAction, err := SignAndBuildPermitAction(
		ctx,
		c.conn,
		c.chain,
		coin,
		c.opts.From,
		multicallAddress,
		supplyAmountWei,
		deadline,
		c.signer,
	)
	if err != nil {
		return nil, err
	}

	// Build transaction actions
	transferFromAction := BuildTransferFromAction(
		coinAddress,
		c.opts.From,
		multicallAddress,
		supplyAmountWei,
	)

	// Approve Aave V3 pool to spend coin
	approveAction := BuildApproveAction(
		coinAddress,
		aaveV3PoolAddress,
		supplyAmountWei,
	)

	// Supply coin to Aave V3 pool
	supplyAction := BuildSupplyAction(
		aaveV3PoolAddress,
		coinAddress,
		supplyAmountWei,
		c.opts.From,
	)

	delegationWithSigAction, err := SignAndBuildDelegationWithSigAction(
		ctx,
		c.conn,
		c.chain,
		debtToken,
		c.opts.From,
		multicallAddress,
		borrowAmountWei,
		deadline,
		c.signer,
	)
	if err != nil {
		return nil, err
	}
	borrowAction := BuildBorrowAction(
		aaveV3PoolAddress,
		coinAddress,
		borrowAmountWei,
		c.opts.From,
	)
	transferAction := BuildTransferAction(
		coinAddress,
		c.opts.From,
		borrowAmountWei,
	)
	actions := []ExecuteAction{
		NewExecuteAction(permitAction, false),
		NewExecuteAction(transferFromAction, false),
		NewExecuteAction(approveAction, false),
		NewExecuteAction(supplyAction, false),
		NewExecuteAction(delegationWithSigAction, false),
		NewExecuteAction(borrowAction, false),
		NewExecuteAction(transferAction, false),
	}
	return c.ExecuteTxActions(ctx, actions)
}

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
