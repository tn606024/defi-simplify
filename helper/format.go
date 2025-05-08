package helper

import (
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/shopspring/decimal"
)

// Helper to convert decimal amount to big.Int with proper decimals
func ToWei(amount decimal.Decimal, decimals uint8) *big.Int {
	mul := decimal.NewFromInt(10).Pow(decimal.NewFromInt(int64(decimals)))
	result := amount.Mul(mul)
	wei := new(big.Int)
	wei.SetString(result.String(), 10)
	return wei
}

// Helper to convert big.Int to decimal with proper decimals
func FromWei(amount *big.Int, decimals uint8) decimal.Decimal {
	mul := decimal.NewFromInt(10).Pow(decimal.NewFromInt(int64(decimals)))
	result := decimal.NewFromBigInt(amount, 0)
	return result.Div(mul)
}

func ToCallMsg(tx *types.Transaction, from common.Address) ethereum.CallMsg {
	return ethereum.CallMsg{
		From:       from,
		To:         tx.To(),
		Gas:        tx.Gas(),
		GasPrice:   tx.GasPrice(),
		GasFeeCap:  tx.GasFeeCap(),
		GasTipCap:  tx.GasTipCap(),
		Value:      tx.Value(),
		Data:       tx.Data(),
		AccessList: tx.AccessList(),
	}
}
