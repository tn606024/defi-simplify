package contract

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type SupplyAction struct {
	BaseAction
	poolAddress  common.Address
	asset        common.Address
	amount       *big.Int
	onBehalfOf   common.Address
	referralCode uint16
}

type SupplyWithPermitAction struct {
	BaseAction
	poolAddress  common.Address
	asset        common.Address
	amount       *big.Int
	onBehalfOf   common.Address
	referralCode uint16
	deadline     *big.Int
	permitV      uint8
	permitR      [32]byte
	permitS      [32]byte
}

type WithdrawAction struct {
	BaseAction
	poolAddress common.Address
	asset       common.Address
	amount      *big.Int
	to          common.Address
}

type BorrowAction struct {
	BaseAction
	poolAddress      common.Address
	asset            common.Address
	amount           *big.Int
	interestRateMode *big.Int
	referralCode     uint16
	onBehalfOf       common.Address
}

type RepayAction struct {
	BaseAction
	poolAddress      common.Address
	asset            common.Address
	amount           *big.Int
	interestRateMode *big.Int
	onBehalfOf       common.Address
}

type RepayWithPermitAction struct {
	BaseAction
	poolAddress      common.Address
	asset            common.Address
	amount           *big.Int
	interestRateMode *big.Int
	onBehalfOf       common.Address
	deadline         *big.Int
	permitV          uint8
	permitR          [32]byte
	permitS          [32]byte
}

type GetReserveDataAction struct {
	BaseAction
	poolAddress common.Address
	asset       common.Address
}

type GetUserAccountDataAction struct {
	BaseAction
	poolAddress common.Address
	user        common.Address
}

type GetAllReservesTokensAction struct {
	BaseAction
	protocolDataProviderAddress common.Address
}

type GetUserReserveDataAction struct {
	BaseAction
	protocolDataProviderAddress common.Address
	asset                       common.Address
	user                        common.Address
}
