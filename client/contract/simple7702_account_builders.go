package contract

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// BuildSimple7702AccountExecuteAction builds a call to execute one target from a delegated EOA.
//
// account is the EOA address whose code delegates to Simple7702Account. It is not
// the implementation contract address.
func BuildSimple7702AccountExecuteAction(account common.Address, target common.Address, value *big.Int, data []byte) *Simple7702AccountExecuteAction {
	action := &Simple7702AccountExecuteAction{
		account: account,
		target:  target,
		value:   value,
		data:    data,
	}
	action.BaseAction = BaseAction{
		ToDataFunc: action.ToData,
	}
	return action
}

// BuildSimple7702AccountExecuteBatchAction builds a batch call to a delegated EOA.
//
// account is the EOA address whose code delegates to Simple7702Account. It is not
// the implementation contract address.
func BuildSimple7702AccountExecuteBatchAction(account common.Address, calls []Call) *Simple7702AccountExecuteBatchAction {
	action := &Simple7702AccountExecuteBatchAction{
		account: account,
		calls:   calls,
	}
	action.BaseAction = BaseAction{
		ToDataFunc: action.ToData,
	}
	return action
}
