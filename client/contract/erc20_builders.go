package contract

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
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
