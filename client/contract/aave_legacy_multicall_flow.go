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
