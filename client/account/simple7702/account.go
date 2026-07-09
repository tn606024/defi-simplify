package simple7702

import (
	_ "embed"
)

const Source = "eth-infinitism/account-abstraction@v0.9.0 contracts/accounts/Simple7702Account.sol"

//go:embed abi/Simple7702Account.json
var ABIJSON string
