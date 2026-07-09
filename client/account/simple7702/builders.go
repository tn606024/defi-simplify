package simple7702

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/tn606024/defi-simplify/client/contract"
)

// BuildExecuteAction builds a call to execute one target from a delegated EOA.
//
// account is the EOA address whose code delegates to Simple7702Account. It is not
// the implementation contract address.
func BuildExecuteAction(account common.Address, target common.Address, value *big.Int, data []byte) *ExecuteAction {
	action := &ExecuteAction{
		account: account,
		target:  target,
		value:   value,
		data:    data,
	}
	action.BaseAction = contract.BaseAction{
		ToDataFunc: action.ToData,
	}
	return action
}

// BuildExecuteBatchAction builds a batch call to a delegated EOA.
//
// account is the EOA address whose code delegates to Simple7702Account. It is not
// the implementation contract address.
func BuildExecuteBatchAction(account common.Address, calls []contract.Call) *ExecuteBatchAction {
	action := &ExecuteBatchAction{
		account: account,
		calls:   calls,
	}
	action.BaseAction = contract.BaseAction{
		ToDataFunc: action.ToData,
	}
	return action
}
