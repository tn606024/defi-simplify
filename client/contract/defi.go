package contract

import (
	"context"
	"math/big"
	"time"

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
	debtToken := coin.DebtToken()
	coinAddress := coin.Address(c.chain)
	multicallAddress := c.chain.MulticallAddress()
	aaveV3PoolAddress := c.chain.AaveV3PoolAddress()
	supplyAmountWei := c.ToWei(supplyAmount, coin.Decimals())
	borrowAmountWei := c.ToWei(borrowAmount, coin.Decimals())
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
