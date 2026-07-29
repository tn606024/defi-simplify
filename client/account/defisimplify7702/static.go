package defisimplify7702

import (
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/tn606024/defi-simplify/client/account/defisimplify7702/bindings"
	"github.com/tn606024/defi-simplify/client/contract"
)

var (
	ErrEmptyStaticBatch       = errors.New("static batch is empty")
	ErrIncompatibleStaticCall = errors.New("static call is incompatible with account ABI")
)

// EncodeExecuteBatch validates and ABI-encodes exact calls for the inherited
// BaseAccount.executeBatch entrypoint.
func EncodeExecuteBatch(calls []contract.Call) ([]byte, error) {
	if len(calls) == 0 {
		return nil, ErrEmptyStaticBatch
	}
	bindingCalls := make([]bindings.BaseAccountCall, len(calls))
	for i, call := range calls {
		if call.Target == (common.Address{}) {
			return nil, fmt.Errorf("%w: call %d: target is zero", ErrIncompatibleStaticCall, i)
		}
		bindingCalls[i] = bindings.BaseAccountCall{
			Target: call.Target,
			Value:  cloneBigIntOrZero(call.Value),
			Data:   append([]byte(nil), call.Data...),
		}
	}

	parsed, err := bindings.DefiSimplify7702AccountMetaData.GetAbi()
	if err != nil {
		return nil, fmt.Errorf("load DefiSimplify7702Account ABI: %w", err)
	}
	data, err := parsed.Pack("executeBatch", bindingCalls)
	if err != nil {
		return nil, fmt.Errorf("encode executeBatch: %w", err)
	}
	return data, nil
}
