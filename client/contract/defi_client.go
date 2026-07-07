package contract

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/tn606024/defi-simplify/config"
	"github.com/tn606024/defi-simplify/helper"
)

// DefiClient composes all DeFi related clients.
type DefiClient struct {
	*BaseClientWithConverter
	ERC20 ERC20Interface
	Aave  AaveV3Interface
}

// NewDefiClient creates a new DefiClient with all sub-clients.
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
