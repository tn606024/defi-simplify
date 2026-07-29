package contract

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/shopspring/decimal"
	"github.com/tn606024/defi-simplify/config"
	"github.com/tn606024/defi-simplify/helper"
)

// BaseClient is the base client for all contract interactions
type BaseClient struct {
	conn  EthereumClient
	chain config.Chain
	opts  *bind.TransactOpts
}

// BaseClientWithConverter is a client that can convert between wei and decimal amounts
type BaseClientWithConverter struct {
	*BaseClient
}

// NewBaseClient creates a new BaseClient
func NewBaseClient(conn EthereumClient, chain config.Chain, opts *bind.TransactOpts) *BaseClient {
	return &BaseClient{
		conn:  conn,
		chain: chain,
		opts:  opts,
	}
}

// ToWei converts a decimal amount to wei
func (c *BaseClient) ToWei(amount decimal.Decimal, decimals uint8) *big.Int {
	return helper.ToWei(amount, decimals)
}

// FromWei converts wei to a decimal amount
func (c *BaseClient) FromWei(amount *big.Int, decimals uint8) decimal.Decimal {
	return helper.FromWei(amount, decimals)
}
