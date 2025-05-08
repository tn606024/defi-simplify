package config

import "github.com/ethereum/go-ethereum/common"

var AaveV3PoolAddress = map[Chain]common.Address{
	Base: common.HexToAddress("0xA238Dd80C259a72e81d7e4664a9801593F98d1c5"),
}

var WrappedTokenGatewayV3Address = map[Chain]common.Address{
	Base: common.HexToAddress("0xa0d9C1E9E48Ca30c8d8C3B5D69FF5dc1f6DFfC24"),
}

func (c Chain) AaveV3PoolAddress() common.Address {
	return AaveV3PoolAddress[c]
}

func (c Chain) WrappedTokenGatewayV3Address() common.Address {
	return WrappedTokenGatewayV3Address[c]
}

var MulticallAddress = map[Chain]common.Address{
	Base: common.HexToAddress("0xcA11bde05977b3631167028862bE2a173976CA11"),
}

func (c Chain) MulticallAddress() common.Address {
	return MulticallAddress[c]
}
