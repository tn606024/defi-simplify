package contract

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/shopspring/decimal"
	"github.com/tn606024/defi-simplify/config"
)

// This file contains the legacy Multicall-based Aave composed flow.
// It is kept separate from protocol action implementations because Multicall
// becomes msg.sender, which is the limitation Phase 1 EIP-7702 work aims to address.

func (c *DefiClient) LegacyMulticallSupplyAndBorrowAaveV3Coin(ctx context.Context, coin config.Coin, supplyAmount decimal.Decimal, borrowAmount decimal.Decimal) (*types.Receipt, error) {
	return c.LegacyMulticallSupplyAndBorrowAaveV3Coins(ctx, coin, coin, supplyAmount, borrowAmount)
}

func (c *DefiClient) LegacyMulticallSupplyAndBorrowAaveV3Coins(ctx context.Context, supplyCoin config.Coin, borrowCoin config.Coin, supplyAmount decimal.Decimal, borrowAmount decimal.Decimal) (*types.Receipt, error) {
	debtToken, err := borrowCoin.DebtToken()
	if err != nil {
		return nil, err
	}
	supplyCoinAddress, err := supplyCoin.Address(c.chain)
	if err != nil {
		return nil, err
	}
	borrowCoinAddress, err := borrowCoin.Address(c.chain)
	if err != nil {
		return nil, err
	}
	supplyCoinDecimals, err := supplyCoin.Decimals()
	if err != nil {
		return nil, err
	}
	borrowCoinDecimals, err := borrowCoin.Decimals()
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
	supplyAmountWei := c.ToWei(supplyAmount, supplyCoinDecimals)
	borrowAmountWei := c.ToWei(borrowAmount, borrowCoinDecimals)
	deadline := big.NewInt(time.Now().Add(time.Minute * 10).Unix())
	permitAction, err := SignAndBuildPermitAction(
		ctx,
		c.conn,
		c.chain,
		supplyCoin,
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
		supplyCoinAddress,
		c.opts.From,
		multicallAddress,
		supplyAmountWei,
	)

	// Approve Aave V3 pool to spend supplied collateral.
	approveAction := BuildApproveAction(
		supplyCoinAddress,
		aaveV3PoolAddress,
		supplyAmountWei,
	)

	// Supply collateral to Aave V3 pool.
	supplyAction := BuildSupplyAction(
		aaveV3PoolAddress,
		supplyCoinAddress,
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
		borrowCoinAddress,
		borrowAmountWei,
		c.opts.From,
	)
	transferAction := BuildTransferAction(
		borrowCoinAddress,
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
