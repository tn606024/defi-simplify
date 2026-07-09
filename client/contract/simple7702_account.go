package contract

import (
	_ "embed"
)

const Simple7702AccountSource = "eth-infinitism/account-abstraction@v0.9.0 contracts/accounts/Simple7702Account.sol"

//go:embed abi/simple7702/Simple7702Account.json
var simple7702AccountABI string
