package contract

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/tn606024/defi-simplify/bind/multicall"
	"github.com/tn606024/defi-simplify/config"
	"github.com/tn606024/defi-simplify/helper"
)

// BuildTransferAction creates a new TransferAction
func BuildTransferAction(token common.Address, to common.Address, amount *big.Int) *TransferAction {
	action := &TransferAction{
		token:  token,
		to:     to,
		amount: amount,
	}
	action.BaseAction = BaseAction{
		ToDataFunc: action.ToData,
	}
	return action
}

// BuildApproveAction creates a new ApproveAction
func BuildApproveAction(token common.Address, spender common.Address, amount *big.Int) *ApproveAction {
	action := &ApproveAction{
		token:   token,
		spender: spender,
		amount:  amount,
	}
	action.BaseAction = BaseAction{
		ToDataFunc: action.ToData,
	}
	return action
}

// BuildTransferFromAction creates a new TransferFromAction
func BuildTransferFromAction(token common.Address, from common.Address, to common.Address, amount *big.Int) *TransferFromAction {
	action := &TransferFromAction{
		token:  token,
		from:   from,
		to:     to,
		amount: amount,
	}
	action.BaseAction = BaseAction{
		ToDataFunc: action.ToData,
	}
	return action
}

// BuildBalanceOfAction creates a new BalanceOfAction
func BuildBalanceOfAction(token common.Address, user common.Address) *BalanceOfAction {
	action := &BalanceOfAction{
		token: token,
		user:  user,
	}
	action.BaseAction = BaseAction{
		ToDataFunc: action.ToData,
	}
	return action
}

// BuildPermitAction creates a new PermitAction
func BuildPermitAction(token common.Address, owner common.Address, spender common.Address, amount *big.Int, deadline *big.Int, v uint8, r [32]byte, s [32]byte) *PermitAction {
	action := &PermitAction{
		token:    token,
		owner:    owner,
		spender:  spender,
		amount:   amount,
		deadline: deadline,
		v:        v,
		r:        r,
		s:        s,
	}
	action.BaseAction = BaseAction{
		ToDataFunc: action.ToData,
	}
	return action
}

// BuildNoncesAction creates a new NoncesAction
func BuildNoncesAction(token common.Address, owner common.Address) *NoncesAction {
	action := &NoncesAction{
		token: token,
		owner: owner,
	}
	action.BaseAction = BaseAction{
		ToDataFunc: action.ToData,
	}
	return action
}

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

// BuildMulticallAction creates a new MulticallAction
func BuildMulticallAction(target common.Address, calls []multicall.IMulticall3Call3) *MulticallAction {
	action := &MulticallAction{
		target: target,
		calls:  calls,
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
	nonceAction := BuildNoncesAction(coin.Address(chain), delegator)
	nonce, err := nonces(conn, nonceAction)
	if err != nil {
		return nil, err
	}
	domain := coin.PermitDomain(chain)
	DelegationWithSigTypedData := helper.NewDelegationWithSig(delegatee, amount, nonce, deadline)
	DelegationWithSigMsg := helper.NewDelegationWithSigEIP712Msg(domain, DelegationWithSigTypedData)
	v, r, s, err := helper.SignEIP712MsgAndGetVRS(signer, DelegationWithSigMsg)
	if err != nil {
		return nil, err
	}
	return BuildDelegationWithSigAction(
		coin.Address(chain),
		delegator,
		delegatee,
		amount,
		deadline,
		v,
		r,
		s,
	), nil
}

// SignAndBuildPermitAction signs a permit message and builds a PermitAction
func SignAndBuildPermitAction(
	ctx context.Context,
	conn EthereumClient,
	chain config.Chain,
	coin config.Coin,
	owner common.Address,
	spender common.Address,
	amount *big.Int,
	deadline *big.Int,
	signer *helper.MsgSigner,
) (*PermitAction, error) {
	nonceAction := BuildNoncesAction(coin.Address(chain), owner)
	nonce, err := nonces(conn, nonceAction)
	if err != nil {
		return nil, err
	}
	domain := coin.PermitDomain(chain)
	permitTypedData := helper.NewPermit(owner, spender, amount, nonce, deadline)
	permitMsg := helper.NewPermitEIP712Msg(domain, permitTypedData)
	v, r, s, err := helper.SignEIP712MsgAndGetVRS(signer, permitMsg)
	if err != nil {
		return nil, err
	}
	return BuildPermitAction(
		coin.Address(chain),
		owner,
		spender,
		amount,
		deadline,
		v,
		r,
		s,
	), nil
}
