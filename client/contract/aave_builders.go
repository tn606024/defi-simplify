package contract

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/tn606024/defi-simplify/config"
	"github.com/tn606024/defi-simplify/helper"
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

// BuildSupplyWithPermitAction creates a new SupplyWithPermitAction
func BuildSupplyWithPermitAction(poolAddress common.Address, asset common.Address, amount *big.Int, onBehalfOf common.Address, referralCode uint16, deadline *big.Int, permitV uint8, permitR [32]byte, permitS [32]byte) *SupplyWithPermitAction {
	action := &SupplyWithPermitAction{
		poolAddress:  poolAddress,
		asset:        asset,
		amount:       amount,
		onBehalfOf:   onBehalfOf,
		referralCode: referralCode,
		deadline:     deadline,
		permitV:      permitV,
		permitR:      permitR,
		permitS:      permitS,
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

// BuildBorrowETHAction creates a new BorrowETHAction
func BuildBorrowETHAction(wrappedTokenGatewayAddress common.Address, pool common.Address, amount *big.Int) *BorrowETHAction {
	action := &BorrowETHAction{
		wrappedTokenGatewayAddress: wrappedTokenGatewayAddress,
		pool:                       pool,
		amount:                     amount,
		referralCode:               0,
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

// BuildRepayWithPermitAction creates a new RepayWithPermitAction
func BuildRepayWithPermitAction(poolAddress common.Address, asset common.Address, amount *big.Int, onBehalfOf common.Address, deadline *big.Int, permitV uint8, permitR [32]byte, permitS [32]byte) *RepayWithPermitAction {
	action := &RepayWithPermitAction{
		poolAddress:      poolAddress,
		asset:            asset,
		amount:           amount,
		interestRateMode: big.NewInt(2),
		onBehalfOf:       onBehalfOf,
		deadline:         deadline,
		permitV:          permitV,
		permitR:          permitR,
		permitS:          permitS,
	}
	action.BaseAction = BaseAction{
		ToDataFunc: action.ToData,
	}
	return action
}

// BuildDepositETHAction creates a new DepositETHAction
func BuildDepositETHAction(wrappedTokenGatewayAddress common.Address, pool common.Address, onBehalfOf common.Address, referral uint16, amount *big.Int) *DepositETHAction {
	action := &DepositETHAction{
		wrappedTokenGatewayAddress: wrappedTokenGatewayAddress,
		pool:                       pool,
		onBehalfOf:                 onBehalfOf,
		referral:                   referral,
		amount:                     amount,
	}
	action.BaseAction = BaseAction{
		ToDataFunc: action.ToData,
	}
	return action
}

// BuildWithdrawETHAction creates a new WithdrawETHAction
func BuildWithdrawETHAction(wrappedTokenGatewayAddress common.Address, pool common.Address, amount *big.Int, to common.Address) *WithdrawETHAction {
	action := &WithdrawETHAction{
		wrappedTokenGatewayAddress: wrappedTokenGatewayAddress,
		pool:                       pool,
		amount:                     amount,
		to:                         to,
	}
	action.BaseAction = BaseAction{
		ToDataFunc: action.ToData,
	}
	return action
}

// BuildWithdrawETHWithPermitAction creates a new WithdrawETHWithPermitAction
func BuildWithdrawETHWithPermitAction(wrappedTokenGatewayAddress common.Address, pool common.Address, amount *big.Int, to common.Address, deadline *big.Int, permitV uint8, permitR [32]byte, permitS [32]byte) *WithdrawETHWithPermitAction {
	action := &WithdrawETHWithPermitAction{
		wrappedTokenGatewayAddress: wrappedTokenGatewayAddress,
		pool:                       pool,
		amount:                     amount,
		to:                         to,
		deadline:                   deadline,
		permitV:                    permitV,
		permitR:                    permitR,
		permitS:                    permitS,
	}
	action.BaseAction = BaseAction{
		ToDataFunc: action.ToData,
	}
	return action
}

// BuildApproveDelegationAction creates a new ApproveDelegationAction
func BuildApproveDelegationAction(asset common.Address, delegatee common.Address, amount *big.Int) *ApproveDelegationAction {
	action := &ApproveDelegationAction{
		asset:     asset,
		delegatee: delegatee,
		amount:    amount,
	}
	action.BaseAction = BaseAction{
		ToDataFunc: action.ToData,
	}
	return action
}

// BuildDelegationWithSigAction creates a new DelegationWithSigAction
func BuildDelegationWithSigAction(asset common.Address, delegator common.Address, delegatee common.Address, value *big.Int, deadline *big.Int, v uint8, r [32]byte, s [32]byte) *DelegationWithSigAction {
	action := &DelegationWithSigAction{
		asset:     asset,
		delegator: delegator,
		delegatee: delegatee,
		value:     value,
		deadline:  deadline,
		v:         v,
		r:         r,
		s:         s,
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

func SignAndBuildDelegationWithSigAction(
	ctx context.Context,
	conn EthereumClient,
	chain config.Chain,
	coin config.Coin,
	delegator common.Address,
	delegatee common.Address,
	amount *big.Int,
	deadline *big.Int,
	signer *helper.MsgSigner,
) (*DelegationWithSigAction, error) {
	coinAddress, err := coin.Address(chain)
	if err != nil {
		return nil, err
	}
	nonceAction := BuildNoncesAction(coinAddress, delegator)
	nonce, err := nonces(conn, nonceAction)
	if err != nil {
		return nil, err
	}
	domain, err := coin.PermitDomain(chain)
	if err != nil {
		return nil, err
	}
	DelegationWithSigTypedData := helper.NewDelegationWithSig(delegatee, amount, nonce, deadline)
	DelegationWithSigMsg := helper.NewDelegationWithSigEIP712Msg(domain, DelegationWithSigTypedData)
	v, r, s, err := helper.SignEIP712MsgAndGetVRS(signer, DelegationWithSigMsg)
	if err != nil {
		return nil, err
	}
	return BuildDelegationWithSigAction(
		coinAddress,
		delegator,
		delegatee,
		amount,
		deadline,
		v,
		r,
		s,
	), nil
}
