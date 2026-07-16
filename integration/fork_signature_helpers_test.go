//go:build integration

package integration

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/tn606024/defi-simplify/aave"
	bindaave "github.com/tn606024/defi-simplify/bind/aave"
	binderc20 "github.com/tn606024/defi-simplify/bind/erc20"
	sdkerc20 "github.com/tn606024/defi-simplify/erc20"
	"github.com/tn606024/defi-simplify/helper"
)

func signPermit(
	ctx context.Context,
	client *ethclient.Client,
	capability sdkerc20.PermitCapability,
	owner,
	spender common.Address,
	amount,
	deadline *big.Int,
	signer *helper.MsgSigner,
) (uint8, [32]byte, [32]byte, error) {
	if err := capability.Validate(); err != nil {
		return 0, [32]byte{}, [32]byte{}, err
	}
	token, err := binderc20.NewIErc20WithPermit(capability.Token().Address(), client)
	if err != nil {
		return 0, [32]byte{}, [32]byte{}, err
	}
	nonce, err := token.Nonces(&bind.CallOpts{Context: ctx}, owner)
	if err != nil {
		return 0, [32]byte{}, [32]byte{}, fmt.Errorf("read permit nonce: %w", err)
	}
	domain, err := capability.Domain()
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
	capability aave.DelegationCapability,
	delegator,
	delegatee common.Address,
	amount,
	deadline *big.Int,
	signer *helper.MsgSigner,
) (uint8, [32]byte, [32]byte, error) {
	if err := capability.Validate(); err != nil {
		return 0, [32]byte{}, [32]byte{}, err
	}
	debtToken := capability.Reserve().VariableDebtToken()
	debtTokenAddress := debtToken.Address()
	token, err := bindaave.NewDebtTokenBase(debtTokenAddress, client)
	if err != nil {
		return 0, [32]byte{}, [32]byte{}, err
	}
	nonce, err := token.Nonces(&bind.CallOpts{Context: ctx}, delegator)
	if err != nil {
		return 0, [32]byte{}, [32]byte{}, fmt.Errorf("read delegation nonce: %w", err)
	}
	domain, err := capability.Domain()
	if err != nil {
		return 0, [32]byte{}, [32]byte{}, err
	}
	delegation := helper.NewDelegationWithSig(delegatee, amount, nonce, deadline)
	message := helper.NewDelegationWithSigEIP712Msg(domain, delegation)
	return helper.SignEIP712MsgAndGetVRS(signer, message)
}
