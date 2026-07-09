package eip7702

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/holiman/uint256"
)

type SetCodeTransactionRequest struct {
	From       common.Address
	Signer     bind.SignerFn
	ChainID    *big.Int
	Nonce      uint64
	To         common.Address
	Value      *big.Int
	Data       []byte
	Gas        uint64
	GasFeeCap  *big.Int
	GasTipCap  *big.Int
	AccessList types.AccessList
	AuthList   []types.SetCodeAuthorization
}

func BuildSetCodeTransaction(req SetCodeTransactionRequest) (*types.Transaction, error) {
	if req.Signer == nil {
		return nil, errors.New("transaction signer is nil")
	}
	if len(req.AuthList) == 0 {
		return nil, errors.New("set-code transaction requires at least one authorization")
	}
	if req.Gas == 0 {
		return nil, errors.New("set-code transaction gas limit is zero")
	}

	chainID, err := uint256FromBig("chain ID", req.ChainID)
	if err != nil {
		return nil, err
	}
	value, err := uint256FromOptionalBig("value", req.Value)
	if err != nil {
		return nil, err
	}
	gasFeeCap, err := uint256FromBig("gas fee cap", req.GasFeeCap)
	if err != nil {
		return nil, err
	}
	gasTipCap, err := uint256FromBig("gas tip cap", req.GasTipCap)
	if err != nil {
		return nil, err
	}

	tx := types.NewTx(&types.SetCodeTx{
		ChainID:    chainID,
		Nonce:      req.Nonce,
		GasTipCap:  gasTipCap,
		GasFeeCap:  gasFeeCap,
		Gas:        req.Gas,
		To:         req.To,
		Value:      value,
		Data:       common.CopyBytes(req.Data),
		AccessList: req.AccessList,
		AuthList:   req.AuthList,
	})

	signedTx, err := req.Signer(req.From, tx)
	if err != nil {
		return nil, fmt.Errorf("sign set-code transaction: %w", err)
	}
	return signedTx, nil
}

func uint256FromOptionalBig(label string, value *big.Int) (*uint256.Int, error) {
	if value == nil {
		return new(uint256.Int), nil
	}
	return uint256FromBig(label, value)
}
