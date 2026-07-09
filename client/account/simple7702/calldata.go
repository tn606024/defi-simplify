package simple7702

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/tn606024/defi-simplify/client/contract"
)

func (a *ExecuteAction) ToData(ctx context.Context, conn contract.EthereumClient, opt *bind.TransactOpts) (common.Address, []byte, error) {
	data, err := EncodeExecute(a.target, a.value, a.data)
	if err != nil {
		return common.Address{}, nil, err
	}
	return a.account, data, nil
}

func (a *ExecuteBatchAction) ToData(ctx context.Context, conn contract.EthereumClient, opt *bind.TransactOpts) (common.Address, []byte, error) {
	data, err := EncodeExecuteBatch(a.calls)
	if err != nil {
		return common.Address{}, nil, err
	}
	return a.account, data, nil
}

func EncodeExecute(target common.Address, value *big.Int, data []byte) ([]byte, error) {
	parsed, err := parseABI()
	if err != nil {
		return nil, err
	}
	return parsed.Pack("execute", target, zeroIfNil(value), data)
}

func EncodeExecuteBatch(calls []contract.Call) ([]byte, error) {
	parsed, err := parseABI()
	if err != nil {
		return nil, err
	}

	accountCalls := make([]accountCall, 0, len(calls))
	for _, call := range calls {
		accountCalls = append(accountCalls, accountCall{
			Target: call.Target,
			Value:  zeroIfNil(call.Value),
			Data:   call.Data,
		})
	}

	return parsed.Pack("executeBatch", accountCalls)
}

func parseABI() (abi.ABI, error) {
	parsed, err := abi.JSON(strings.NewReader(ABIJSON))
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
