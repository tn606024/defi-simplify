package contract

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// BuildSupplyAction creates a new SupplyAction
func BuildSupplyAction(poolAddress common.Address, asset common.Address, amount *big.Int, onBehalfOf common.Address) *SupplyAction {
	action := &SupplyAction{
		poolAddress:  poolAddress,
		asset:        asset,
		amount:       amount,
		onBehalfOf:   onBehalfOf,
		referralCode: 0,
	}
	action.BaseAction = BaseAction{
		ToDataFunc: action.ToData,
	}
	return action
}

// BuildWithdrawAction creates a new WithdrawAction
func BuildWithdrawAction(poolAddress common.Address, asset common.Address, amount *big.Int, to common.Address) *WithdrawAction {
	action := &WithdrawAction{
		poolAddress: poolAddress,
		asset:       asset,
		amount:      amount,
		to:          to,
	}
	action.BaseAction = BaseAction{
		ToDataFunc: action.ToData,
	}
	return action
}

// BuildBorrowAction creates a new BorrowAction
func BuildBorrowAction(poolAddress common.Address, asset common.Address, amount *big.Int, onBehalfOf common.Address) *BorrowAction {
	action := &BorrowAction{
		poolAddress:      poolAddress,
		asset:            asset,
		amount:           amount,
		interestRateMode: big.NewInt(2),
		referralCode:     0,
		onBehalfOf:       onBehalfOf,
	}
	action.BaseAction = BaseAction{
		ToDataFunc: action.ToData,
	}
	return action
}

// BuildRepayAction creates a new RepayAction
func BuildRepayAction(poolAddress common.Address, asset common.Address, amount *big.Int, onBehalfOf common.Address) *RepayAction {
	action := &RepayAction{
		poolAddress:      poolAddress,
		asset:            asset,
		amount:           amount,
		interestRateMode: big.NewInt(2),
		onBehalfOf:       onBehalfOf,
	}
	action.BaseAction = BaseAction{
		ToDataFunc: action.ToData,
	}
	return action
}

// BuildGetReserveDataAction creates a new GetReserveDataAction
func BuildGetReserveDataAction(poolAddress common.Address, asset common.Address) *GetReserveDataAction {
	action := &GetReserveDataAction{
		poolAddress: poolAddress,
		asset:       asset,
	}
	action.BaseAction = BaseAction{
		ToDataFunc: action.ToData,
	}
	return action
}

// BuildGetUserAccountDataAction creates a new GetUserAccountDataAction
func BuildGetUserAccountDataAction(poolAddress common.Address, user common.Address) *GetUserAccountDataAction {
	action := &GetUserAccountDataAction{
		poolAddress: poolAddress,
		user:        user,
	}
	action.BaseAction = BaseAction{
		ToDataFunc: action.ToData,
	}
	return action
}

// BuildGetAllReservesTokensAction creates a new GetAllReservesTokensAction
func BuildGetAllReservesTokensAction(protocolDataProviderAddress common.Address) *GetAllReservesTokensAction {
	action := &GetAllReservesTokensAction{
		protocolDataProviderAddress: protocolDataProviderAddress,
	}
	action.BaseAction = BaseAction{
		ToDataFunc: action.ToData,
	}
	return action
}

// BuildGetUserReserveDataAction creates a new GetUserReserveDataAction
func BuildGetUserReserveDataAction(protocolDataProviderAddress common.Address, asset common.Address, user common.Address) *GetUserReserveDataAction {
	action := &GetUserReserveDataAction{
		protocolDataProviderAddress: protocolDataProviderAddress,
		asset:                       asset,
		user:                        user,
	}
	action.BaseAction = BaseAction{
		ToDataFunc: action.ToData,
	}
	return action
}
