package contract

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type TransferAction struct {
	BaseAction
	token  common.Address // Token contract address
	to     common.Address // Recipient address
	amount *big.Int       // Amount to transfer
}

type ApproveAction struct {
	BaseAction
	token   common.Address // Token contract address
	spender common.Address // Address to approve
	amount  *big.Int       // Amount to approve
}

type TransferFromAction struct {
	BaseAction
	token  common.Address
	from   common.Address
	to     common.Address
	amount *big.Int
}

type BalanceOfAction struct {
	BaseAction
	token common.Address
	user  common.Address
}
