package contract

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type Simple7702AccountExecuteAction struct {
	BaseAction
	account common.Address
	target  common.Address
	value   *big.Int
	data    []byte
}

type Simple7702AccountExecuteBatchAction struct {
	BaseAction
	account common.Address
	calls   []Call
}

type simple7702AccountCall struct {
	Target common.Address `abi:"target"`
	Value  *big.Int       `abi:"value"`
	Data   []byte         `abi:"data"`
}
