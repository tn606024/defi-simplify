//go:build integration

package integration

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	bindaave "github.com/tn606024/defi-simplify/bind/aave"
	binderc20 "github.com/tn606024/defi-simplify/bind/erc20"
	"github.com/tn606024/defi-simplify/config"
	"github.com/tn606024/defi-simplify/helper"
)

func signPermit(
	ctx context.Context,
	client *ethclient.Client,
	coin config.Coin,
	owner,
	spender common.Address,
	amount,
	deadline *big.Int,
	signer *helper.MsgSigner,
) (uint8, [32]byte, [32]byte, error) {
	tokenAddress, err := coin.Address(config.Base)
	if err != nil {
		return 0, [32]byte{}, [32]byte{}, err
	}
	token, err := binderc20.NewIErc20WithPermit(tokenAddress, client)
	if err != nil {
		return 0, [32]byte{}, [32]byte{}, err
	}
	nonce, err := token.Nonces(&bind.CallOpts{Context: ctx}, owner)
	if err != nil {
		return 0, [32]byte{}, [32]byte{}, fmt.Errorf("read permit nonce: %w", err)
	}
	domain, err := coin.PermitDomain(config.Base)
	if err != nil {
		return 0, [32]byte{}, [32]byte{}, err
	}
	permit := helper.NewPermit(owner, spender, amount, nonce, deadline)
	message := helper.NewPermitEIP712Msg(domain, permit)
	return helper.SignEIP712MsgAndGetVRS(signer, message)
}

func signDelegation(
	ctx context.Context,
	client *ethclient.Client,
	asset config.Coin,
	delegator,
	delegatee common.Address,
	amount,
	deadline *big.Int,
	signer *helper.MsgSigner,
) (uint8, [32]byte, [32]byte, error) {
	debtToken, err := asset.DebtToken()
	if err != nil {
		return 0, [32]byte{}, [32]byte{}, err
	}
	debtTokenAddress, err := debtToken.Address(config.Base)
	if err != nil {
		return 0, [32]byte{}, [32]byte{}, err
	}
	token, err := bindaave.NewDebtTokenBase(debtTokenAddress, client)
	if err != nil {
		return 0, [32]byte{}, [32]byte{}, err
	}
	nonce, err := token.Nonces(&bind.CallOpts{Context: ctx}, delegator)
	if err != nil {
		return 0, [32]byte{}, [32]byte{}, fmt.Errorf("read delegation nonce: %w", err)
	}
	domain, err := debtToken.PermitDomain(config.Base)
	if err != nil {
		return 0, [32]byte{}, [32]byte{}, err
	}
	delegation := helper.NewDelegationWithSig(delegatee, amount, nonce, deadline)
	message := helper.NewDelegationWithSigEIP712Msg(domain, delegation)
	return helper.SignEIP712MsgAndGetVRS(signer, message)
}
