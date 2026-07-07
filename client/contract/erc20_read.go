package contract

import (
	"math/big"

	"github.com/tn606024/defi-simplify/bind/erc20"
)

func balanceOf(conn EthereumClient, action *BalanceOfAction) (*big.Int, error) {
	erc20Instance, err := erc20.NewErc20(action.token, conn)
	if err != nil {
		return nil, err
	}
	balance, err := erc20Instance.BalanceOf(nil, action.user)
	if err != nil {
		return nil, err
	}
	return balance, nil
}

func nonces(conn EthereumClient, action *NoncesAction) (*big.Int, error) {
	erc20Instance, err := erc20.NewIErc20WithPermit(action.token, conn)
	if err != nil {
		return nil, err
	}
	return erc20Instance.Nonces(nil, action.owner)
}
