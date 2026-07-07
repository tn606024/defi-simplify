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

type BorrowETHAction struct {
	BaseAction
	wrappedTokenGatewayAddress common.Address
	pool                       common.Address
	amount                     *big.Int
	referralCode               uint16
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

type DepositETHAction struct {
	BaseAction
	wrappedTokenGatewayAddress common.Address
	pool                       common.Address
	onBehalfOf                 common.Address
	referral                   uint16
	amount                     *big.Int
}

type WithdrawETHAction struct {
	BaseAction
	wrappedTokenGatewayAddress common.Address
	pool                       common.Address
	amount                     *big.Int
	to                         common.Address
}

type WithdrawETHWithPermitAction struct {
	BaseAction
	wrappedTokenGatewayAddress common.Address
	pool                       common.Address
	amount                     *big.Int
	to                         common.Address
	deadline                   *big.Int
	permitV                    uint8
	permitR                    [32]byte
	permitS                    [32]byte
}

type ApproveDelegationAction struct {
	BaseAction
	asset     common.Address
	delegatee common.Address
	amount    *big.Int
}

type DelegationWithSigAction struct {
	BaseAction
	asset     common.Address
	delegator common.Address
	delegatee common.Address
	value     *big.Int
	deadline  *big.Int
	v         uint8
	r         [32]byte
	s         [32]byte
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
