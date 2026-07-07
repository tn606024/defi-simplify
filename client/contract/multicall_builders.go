package contract

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/tn606024/defi-simplify/bind/multicall"
)

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
