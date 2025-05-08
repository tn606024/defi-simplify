package contract

import (
	"context"
	_ "embed"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/shopspring/decimal"
	"github.com/tn606024/defi-simplify/bind/multicall"
	"github.com/tn606024/defi-simplify/config"
	"github.com/tn606024/defi-simplify/helper"
)

//go:embed abi/multicall/IMulticall3.json
var multicallABI string

// BaseClient is the base client for all contract interactions
type BaseClient struct {
	conn   EthereumClient
	chain  config.Chain
	opts   *bind.TransactOpts
	signer *helper.MsgSigner
}

// BaseClientWithConverter is a client that can convert between wei and decimal amounts
type BaseClientWithConverter struct {
	*BaseClient
}

// NewBaseClient creates a new BaseClient
func NewBaseClient(conn EthereumClient, chain config.Chain, opts *bind.TransactOpts, signer *helper.MsgSigner) *BaseClient {
	return &BaseClient{
		conn:   conn,
		chain:  chain,
		opts:   opts,
		signer: signer,
	}
}

// ToWei converts a decimal amount to wei
func (c *BaseClientWithConverter) ToWei(amount decimal.Decimal, decimals uint8) *big.Int {
	return helper.ToWei(amount, decimals)
}

// FromWei converts wei to a decimal amount
func (c *BaseClientWithConverter) FromWei(amount *big.Int, decimals uint8) decimal.Decimal {
	return helper.FromWei(amount, decimals)
}

// Add new action struct
type MulticallAction struct {
	BaseAction
	target common.Address // Multicall contract address
	calls  []multicall.IMulticall3Call3
}

// Update client methods
func (c *BaseClient) ExecuteTxActions(ctx context.Context, actions []ExecuteAction) (*types.Receipt, error) {
	calls := make([]multicall.IMulticall3Call3, 0, len(actions))
	for _, action := range actions {
		call, err := action.ToIMulticall3Call3(ctx, c.conn, c.opts, action.AllowFailure())
		if err != nil {
			return nil, err
		}
		calls = append(calls, *call)
	}
	multicallAction := BuildMulticallAction(c.chain.MulticallAddress(), calls)
	return executeAction(ctx, c.conn, c.opts, multicallAction)
}

func (a *MulticallAction) ToTransaction(ctx context.Context, conn EthereumClient, opt *bind.TransactOpts) (*types.Transaction, error) {
	multicallInterface, err := multicall.NewIMulticall3(a.target, conn)
	if err != nil {
		return nil, err
	}
	return multicallInterface.Aggregate3(opt, a.calls)
}

func (a *MulticallAction) ToData(ctx context.Context, conn EthereumClient, opt *bind.TransactOpts) (common.Address, []byte, error) {
	parsed, err := abi.JSON(strings.NewReader(multicallABI))
	if err != nil {
		return common.Address{}, nil, err
	}
	data, err := parsed.Pack("aggregate3", a.calls)
	if err != nil {
		return common.Address{}, nil, err
	}
	return a.target, data, nil
}
