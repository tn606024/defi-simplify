package contract

import (
	_ "embed"
)

//go:embed abi/erc20/ERC20.json
var erc20ABI string

//go:embed abi/erc20/IERC20Permit.json
var erc20PermitABI string
