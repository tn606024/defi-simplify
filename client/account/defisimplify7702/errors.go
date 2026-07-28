package defisimplify7702

import (
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/tn606024/defi-simplify/client/account/defisimplify7702/bindings"
)

var (
	ErrMalformedContractError = errors.New("malformed Defi Simplify contract error")
	ErrUnknownContractError   = errors.New("unknown Defi Simplify contract error")
)

// ContractErrorArgument is one ABI-decoded custom-error argument.
type ContractErrorArgument struct {
	Name  string
	Value interface{}
}

// ContractError is an ABI-decoded DefiSimplify7702Account custom error.
// Index fields are populated when the ABI exposes the corresponding
// attribution argument.
type ContractError struct {
	Name              string
	Signature         string
	Arguments         []ContractErrorArgument
	CallIndex         *big.Int
	OuterCallIndex    *big.Int
	CallbackCallIndex *big.Int
	CheckpointIndex   *big.Int
	PatchIndex        *big.Int
}

func (e *ContractError) Error() string {
	if e == nil {
		return "<nil>"
	}
	locations := make([]string, 0, 5)
	appendIndex := func(name string, value *big.Int) {
		if value != nil {
			locations = append(locations, name+"="+value.String())
		}
	}
	appendIndex("call", e.CallIndex)
	appendIndex("outerCall", e.OuterCallIndex)
	appendIndex("callbackCall", e.CallbackCallIndex)
	appendIndex("checkpoint", e.CheckpointIndex)
	appendIndex("patch", e.PatchIndex)
	if len(locations) == 0 {
		return e.Name
	}
	return e.Name + " (" + strings.Join(locations, ", ") + ")"
}

// DecodeContractError decodes account custom-error revert data and preserves
// call/checkpoint/patch indices for diagnostics.
func DecodeContractError(data []byte) (*ContractError, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("%w: have %d bytes, need at least 4", ErrMalformedContractError, len(data))
	}
	parsed, err := bindings.DefiSimplify7702AccountMetaData.GetAbi()
	if err != nil {
		return nil, fmt.Errorf("%w: load ABI: %v", ErrMalformedContractError, err)
	}
	var selector [4]byte
	copy(selector[:], data[:4])
	definition, err := parsed.ErrorByID(selector)
	if err != nil {
		return nil, fmt.Errorf("%w: selector 0x%x", ErrUnknownContractError, selector)
	}
	values, err := definition.Inputs.Unpack(data[4:])
	if err != nil {
		return nil, fmt.Errorf("%w: decode %s: %v", ErrMalformedContractError, definition.Name, err)
	}
	if len(values) != len(definition.Inputs) {
		return nil, fmt.Errorf(
			"%w: %s returned %d arguments, expected %d",
			ErrMalformedContractError,
			definition.Name,
			len(values),
			len(definition.Inputs),
		)
	}

	decoded := &ContractError{
		Name:      definition.Name,
		Signature: definition.Sig,
		Arguments: make([]ContractErrorArgument, len(values)),
	}
	for i, value := range values {
		name := definition.Inputs[i].Name
		cloned := cloneContractErrorValue(value)
		decoded.Arguments[i] = ContractErrorArgument{Name: name, Value: cloned}
		switch name {
		case "callIndex", "index":
			decoded.CallIndex = contractErrorBigInt(cloned)
		case "outerCallIndex":
			decoded.OuterCallIndex = contractErrorBigInt(cloned)
		case "callbackCallIndex":
			decoded.CallbackCallIndex = contractErrorBigInt(cloned)
		case "checkpointIndex":
			decoded.CheckpointIndex = contractErrorBigInt(cloned)
		case "patchIndex":
			decoded.PatchIndex = contractErrorBigInt(cloned)
		}
	}
	return decoded, nil
}

func cloneContractErrorValue(value interface{}) interface{} {
	switch typed := value.(type) {
	case *big.Int:
		if typed == nil {
			return (*big.Int)(nil)
		}
		return new(big.Int).Set(typed)
	case []byte:
		return append([]byte(nil), typed...)
	case common.Address:
		return typed
	case [32]byte:
		return typed
	default:
		return typed
	}
}

func contractErrorBigInt(value interface{}) *big.Int {
	typed, ok := value.(*big.Int)
	if !ok || typed == nil {
		return nil
	}
	return new(big.Int).Set(typed)
}
