package config

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
)

var AaveV3PoolAddress = map[Chain]common.Address{
	Base: common.HexToAddress("0xA238Dd80C259a72e81d7e4664a9801593F98d1c5"),
}

var WrappedTokenGatewayV3Address = map[Chain]common.Address{
	Base: common.HexToAddress("0xa0d9C1E9E48Ca30c8d8C3B5D69FF5dc1f6DFfC24"),
}

var AaveProtocolDataProviderAddress = map[Chain]common.Address{
	Base: common.HexToAddress("0xC4Fcf9893072d61Cc2899C0054877Cb752587981"),
}

func addressForChain(addresses map[Chain]common.Address, chain Chain, label string) (common.Address, error) {
	address, ok := addresses[chain]
	if !ok || address == (common.Address{}) {
		return common.Address{}, fmt.Errorf("unsupported %s for chain %d", label, chain)
	}
	return address, nil
}

func (c Chain) AaveV3PoolAddress() (common.Address, error) {
	return addressForChain(AaveV3PoolAddress, c, "Aave V3 pool address")
}

func (c Chain) WrappedTokenGatewayV3Address() (common.Address, error) {
	return addressForChain(WrappedTokenGatewayV3Address, c, "wrapped token gateway address")
}

func (c Chain) AaveProtocolDataProviderAddress() (common.Address, error) {
	return addressForChain(AaveProtocolDataProviderAddress, c, "Aave protocol data provider address")
}

var MulticallAddress = map[Chain]common.Address{
	Base: common.HexToAddress("0xcA11bde05977b3631167028862bE2a173976CA11"),
}

func (c Chain) MulticallAddress() (common.Address, error) {
	return addressForChain(MulticallAddress, c, "multicall address")
}

// Simple7702AccountImplementationAddress points to the audited account implementation
// used as EIP-7702 delegated code.
//
// Source: eth-infinitism/account-abstraction v0.9.0,
// contracts/accounts/Simple7702Account.sol.
var Simple7702AccountImplementationAddress = map[Chain]common.Address{
	Base: common.HexToAddress("0xa625961dcb8a01c75DBeA172F58181FC5C711dA4"),
}

func (c Chain) Simple7702AccountImplementationAddress() (common.Address, error) {
	return addressForChain(Simple7702AccountImplementationAddress, c, "Simple7702Account implementation address")
}
