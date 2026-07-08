package contract

import (
	"context"

	"github.com/ethereum/go-ethereum/core/types"
)

// ActionExecutor executes write actions through a concrete backend.
type ActionExecutor interface {
	ExecuteActions(ctx context.Context, actions []ExecuteAction) (*types.Receipt, error)
}

// CallExecutor executes neutral calls through a concrete backend.
type CallExecutor interface {
	ExecuteCalls(ctx context.Context, calls []Call) (*types.Receipt, error)
}
