package contract

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

func (a *Simple7702AccountExecuteAction) ToData(ctx context.Context, conn EthereumClient, opt *bind.TransactOpts) (common.Address, []byte, error) {
	data, err := EncodeSimple7702AccountExecute(a.target, a.value, a.data)
	if err != nil {
		return common.Address{}, nil, err
	}
	return a.account, data, nil
}

func (a *Simple7702AccountExecuteBatchAction) ToData(ctx context.Context, conn EthereumClient, opt *bind.TransactOpts) (common.Address, []byte, error) {
	data, err := EncodeSimple7702AccountExecuteBatch(a.calls)
	if err != nil {
		return common.Address{}, nil, err
	}
	return a.account, data, nil
}

func EncodeSimple7702AccountExecute(target common.Address, value *big.Int, data []byte) ([]byte, error) {
	parsed, err := parseSimple7702AccountABI()
	if err != nil {
		return nil, err
	}
	return parsed.Pack("execute", target, zeroIfNil(value), data)
}

func EncodeSimple7702AccountExecuteBatch(calls []Call) ([]byte, error) {
	parsed, err := parseSimple7702AccountABI()
	if err != nil {
		return nil, err
	}

	accountCalls := make([]simple7702AccountCall, 0, len(calls))
	for _, call := range calls {
		accountCalls = append(accountCalls, simple7702AccountCall{
			Target: call.Target,
			Value:  zeroIfNil(call.Value),
			Data:   call.Data,
		})
	}

	return parsed.Pack("executeBatch", accountCalls)
}

func parseSimple7702AccountABI() (abi.ABI, error) {
	parsed, err := abi.JSON(strings.NewReader(simple7702AccountABI))
	if err != nil {
		return abi.ABI{}, fmt.Errorf("failed to parse Simple7702Account ABI: %w", err)
	}
	return parsed, nil
}

func zeroIfNil(value *big.Int) *big.Int {
	if value == nil {
		return big.NewInt(0)
	}
	return value
}
