package defisimplify7702

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/tn606024/defi-simplify/config"
)

// RuntimeCodeReader is the read-only Ethereum capability required to verify a
// reviewed deployment during application initialization.
type RuntimeCodeReader interface {
	CodeAt(ctx context.Context, account common.Address, blockNumber *big.Int) ([]byte, error)
}

// ResolveAndVerifyAccountDeployment resolves the reviewed account deployment
// for chain and verifies its latest runtime code. Applications should call it
// once during initialization before enabling signing or submission.
//
// This check does not inspect or change an EOA's EIP-7702 delegation. The
// executor separately verifies the pending delegation target before every
// transaction submission because delegation can be switched or cleared.
func ResolveAndVerifyAccountDeployment(
	ctx context.Context,
	reader RuntimeCodeReader,
	chain config.Chain,
) (ContractDeployment, error) {
	if reader == nil {
		return ContractDeployment{}, fmt.Errorf("runtime code reader is nil")
	}

	deployment, err := DeploymentForChain(chain)
	if err != nil {
		return ContractDeployment{}, fmt.Errorf("resolve reviewed deployment: %w", err)
	}
	implementation, err := deployment.Contract(AccountContract)
	if err != nil {
		return ContractDeployment{}, fmt.Errorf("resolve account implementation: %w", err)
	}

	code, err := reader.CodeAt(ctx, implementation.Address, nil)
	if err != nil {
		return ContractDeployment{}, fmt.Errorf(
			"read account implementation runtime code at %s: %w",
			implementation.Address.Hex(),
			err,
		)
	}
	if len(code) == 0 {
		return ContractDeployment{}, fmt.Errorf(
			"%w: %s at %s",
			ErrMissingRuntimeCode,
			implementation.Name,
			implementation.Address.Hex(),
		)
	}
	if err := implementation.VerifyRuntimeCode(code); err != nil {
		return ContractDeployment{}, fmt.Errorf("verify account implementation runtime code: %w", err)
	}
	return implementation, nil
}
