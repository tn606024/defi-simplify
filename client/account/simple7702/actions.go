package simple7702

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/tn606024/defi-simplify/client/contract"
)

type ExecuteAction struct {
	contract.BaseAction
	account common.Address
	target  common.Address
	value   *big.Int
	data    []byte
}

type ExecuteBatchAction struct {
	contract.BaseAction
	account common.Address
	calls   []contract.Call
}

type accountCall struct {
	Target common.Address `abi:"target"`
	Value  *big.Int       `abi:"value"`
	Data   []byte         `abi:"data"`
}
