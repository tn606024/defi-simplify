package contract

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
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
	coinAddress, err := coin.Address(chain)
	if err != nil {
		return nil, err
	}
	nonceAction := BuildNoncesAction(coinAddress, owner)
	nonce, err := nonces(conn, nonceAction)
	if err != nil {
		return nil, err
	}
	domain, err := coin.PermitDomain(chain)
	if err != nil {
		return nil, err
	}
	permitTypedData := helper.NewPermit(owner, spender, amount, nonce, deadline)
	permitMsg := helper.NewPermitEIP712Msg(domain, permitTypedData)
	v, r, s, err := helper.SignEIP712MsgAndGetVRS(signer, permitMsg)
	if err != nil {
		return nil, err
	}
	return BuildPermitAction(
		coinAddress,
		owner,
		spender,
		amount,
		deadline,
		v,
		r,
		s,
	), nil
}
