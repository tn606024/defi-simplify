# Static Flow Builder API

Status: Draft for IAN-34
Date: 2026-07-08
Related spec: `docs/specs/2026-07-07-phase-1-mvp-spec-and-glossary.md`
Audience: SDK users and maintainers implementing Phase 1

## 1. Purpose

The static Flow builder is the user-facing composition layer for Phase 1.

It should let SDK users describe an ordered DeFi operation in protocol language:

```go
flow := defi.NewFlow(user, defi.WithChain(config.Base)).
    Add(erc20.Approve(config.USDC, aave.PoolSpender(), supplyAmount)).
    Add(aave.Supply(config.USDC, supplyAmount)).
    Add(aave.Borrow(config.WETH, borrowAmount))

calls, err := flow.Build(ctx, conn)
receipt, err := executor.ExecuteCalls(ctx, calls)
```

The public API should read like DeFi composition. Users should not need to manually build `ApproveAction`, `SupplyAction`, or `BorrowAction` values for common flows.

Internally, the implementation can still reuse the existing architecture:

```text
FlowStep -> Action -> Call -> Executor
```

This keeps the SDK ergonomic without throwing away the existing Aave and ERC20 action builders.

## 2. Design Goals

The Phase 1 Flow builder should:

- Provide a clear public API for ordered DeFi composition.
- Keep protocol-specific calldata construction inside protocol step packages.
- Produce neutral `[]Call` values without submitting a transaction.
- Keep execution backend independent from flow construction.
- Work with Multicall, direct transaction, and future Simple7702Account executors.
- Support exact, static amounts known before transaction submission.
- Avoid introducing a full Recipe system before the SDK needs one.

## 3. Non-Goals

The Phase 1 Flow builder does not support:

- Dynamic amount sources such as `FullBalanceOf`.
- Reading a return value from one call and using it to build a later call.
- On-chain guards.
- Slippage guards.
- Simulation policy.
- Uniswap routing.
- Leverage loops.
- Flash loans.
- A formal Recipe lifecycle.

Those features belong to future dynamic Recipe or strategy work.

## 4. Public API Shape

### Package Layout

The intended public package shape is:

```go
import (
    defi "github.com/tn606024/defi-simplify"
    "github.com/tn606024/defi-simplify/aave"
    "github.com/tn606024/defi-simplify/erc20"
    "github.com/tn606024/defi-simplify/config"
)
```

The root package exposes composition primitives:

```go
defi.NewFlow(...)
defi.WithChain(...)
defi.Call
defi.FlowStep
defi.EthereumClient
```

Protocol packages expose user-facing step builders:

```go
erc20.Approve(...)
aave.Supply(...)
aave.Borrow(...)
```

The exact package layout may be implemented incrementally, but the public API should move toward this shape instead of exposing low-level `client/contract` constructors as the primary SDK experience.

### Flow Construction

```go
flow := defi.NewFlow(user, defi.WithChain(config.Base)).
    Add(erc20.Approve(config.USDC, aave.PoolSpender(), supplyAmount)).
    Add(aave.Supply(config.USDC, supplyAmount)).
    Add(aave.Borrow(config.WETH, borrowAmount))
```

`user` is the default account context for the Flow.

For Aave steps, the default user account should be used for:

- `onBehalfOf` in `supply`
- `onBehalfOf` in `borrow`
- default owner context when a step requires ownership semantics

The Flow should preserve step order exactly. It should not sort, merge, or optimize calls in Phase 1.

### Aave-Specific Convenience

The generic ERC20 approval form is useful:

```go
erc20.Approve(config.USDC, spender, amount)
```

For Aave flows, the SDK may also expose a clearer Aave-specific helper:

```go
flow := defi.NewFlow(user, defi.WithChain(config.Base)).
    Add(aave.ApproveSupply(config.USDC, supplyAmount)).
    Add(aave.Supply(config.USDC, supplyAmount)).
    Add(aave.Borrow(config.WETH, borrowAmount))
```

`aave.ApproveSupply` should resolve the correct Aave Pool spender from the Flow chain context.

This avoids forcing users to manually pass an Aave Pool address for the common approve-then-supply path.

## 5. Core Types

### Flow

`Flow` is an ordered list of `FlowStep` values plus shared build context.

Conceptually:

```go
type Flow struct {
    account common.Address
    chain   config.Chain
    steps   []FlowStep
}
```

The exact fields do not need to be exported.

### FlowStep

`FlowStep` is a public interface implemented by protocol step builders.

Conceptually:

```go
type FlowStep interface {
    BuildCalls(ctx context.Context, env BuildEnv) ([]Call, error)
}
```

`BuildEnv` supplies shared context:

```go
type BuildEnv struct {
    Account common.Address
    Chain   config.Chain
    Conn    EthereumClient
}
```

Each step may build one or more calls. Phase 1 steps should usually build exactly one call.

### Call

The public `defi.Call` type should stay aligned with the existing neutral call model:

```go
type Call struct {
    Target common.Address
    Value  *big.Int
    Data   []byte
}
```

Implementation may initially use a type alias to the existing `client/contract.Call` to avoid duplicating call models.

### Amounts

Phase 1 should use exact static amounts known before `Build`.

The first implementation can accept `decimal.Decimal`, matching current client methods:

```go
supplyAmount := decimal.RequireFromString("100")
borrowAmount := decimal.RequireFromString("0.01")
```

Protocol steps should convert decimal amounts to token units during `Build` by using the configured asset decimals.

Future typed amount helpers can be added later, but Phase 1 does not require a separate amount package.

## 6. Build Semantics

`Flow.Build` should encode calls only. It should not submit transactions.

Target shape:

```go
calls, err := flow.Build(ctx, conn)
```

Build steps:

1. Validate shared Flow context.
2. Validate the Flow is not empty.
3. Build each step in insertion order.
4. Convert each step into one or more neutral `Call` values.
5. Return the complete ordered call list.

`Build` should not mutate the Flow or the caller's transaction options.

## 7. Error Handling

`Build` should return clear errors for:

- missing chain context
- empty Flow
- invalid or zero account address
- unsupported asset on the configured chain
- unsupported protocol deployment on the configured chain
- invalid amount
- step build failure

Step errors should include the step index and a stable step name where practical:

```text
build flow step 1 aave.Supply: unsupported asset USDC on chain ...
```

This makes composed flows debuggable without forcing users to inspect generated calldata.

## 8. Execution Boundary

The Flow builder stops at `[]Call`.

Executors are responsible for transaction submission:

```go
type CallExecutor interface {
    ExecuteCalls(ctx context.Context, calls []defi.Call) (*types.Receipt, error)
}
```

This boundary is important:

- Flow knows how to compose protocol operations.
- Actions know how to encode calldata.
- Executors know how to submit calls through a backend.

The same Flow output should be executable through different backends:

```text
[]Call -> MulticallExecutor
[]Call -> DirectExecutor
[]Call -> Simple7702Executor
```

Phase 1 uses this separation so the 7702 work can focus on EOA-native execution instead of also solving SDK composition ergonomics.

## 9. Relationship to Existing Actions

The user-facing API should not require direct action construction for common flows.

This is not the desired primary user experience:

```go
approveAction := contract.BuildApproveAction(token, spender, amount)
supplyAction := contract.BuildSupplyAction(pool, token, amount, user)

flow := contract.NewFlow(user).
    Add(contract.Step(approveAction)).
    Add(contract.Step(supplyAction))
```

That form may exist internally or for advanced users, but it should not be the main SDK story.

The preferred API is:

```go
flow := defi.NewFlow(user, defi.WithChain(config.Base)).
    Add(erc20.Approve(config.USDC, aave.PoolSpender(), supplyAmount)).
    Add(aave.Supply(config.USDC, supplyAmount)).
    Add(aave.Borrow(config.WETH, borrowAmount))
```

Internally, those step builders can create existing actions:

```text
erc20.Approve(...) -> BuildApproveAction(...) -> Call
aave.Supply(...)   -> BuildSupplyAction(...)  -> Call
aave.Borrow(...)   -> BuildBorrowAction(...)  -> Call
```

This keeps the public API readable while preserving the existing tested action layer.

## 10. Relationship to Recipe

`Flow` is not `Recipe`.

In Phase 1:

- Flow is static.
- Flow has no lifecycle beyond build.
- Flow does not perform simulation or policy validation.
- Flow does not support runtime values from previous calls.
- Flow does not model strategy-level risk.

`Recipe` may become useful later when the SDK supports dynamic composition:

```text
swap -> use output amount -> supply
borrow -> swap -> supply -> repeat
health factor guard
slippage bounds
position rebalance
```

Until then, `Flow` should remain a thin, readable composition layer over Actions and Calls.

## 11. Phase 1 Acceptance Criteria

The IAN-35 implementation should satisfy this design by:

- Adding the core static Flow builder.
- Keeping `Build` side-effect free.
- Returning ordered `[]Call` values.
- Preserving the existing neutral `Call` model.
- Supporting fake or simple test steps before protocol-specific steps are added.

The IAN-36 implementation should add protocol step builders by:

- Adding ERC20 approve composition.
- Adding Aave supply composition.
- Adding Aave borrow composition.
- Ensuring Aave default account context maps to `onBehalfOf`.
- Resolving Aave Pool addresses from chain context instead of forcing users to pass them for common flows.

## 12. Example MVP Flow

```go
package main

import (
    "context"

    "github.com/ethereum/go-ethereum/common"
    "github.com/shopspring/decimal"
    defi "github.com/tn606024/defi-simplify"
    "github.com/tn606024/defi-simplify/aave"
    "github.com/tn606024/defi-simplify/config"
)

func buildAaveFlow(ctx context.Context, user common.Address, conn defi.EthereumClient) ([]defi.Call, error) {
    supplyAmount := decimal.RequireFromString("100")
    borrowAmount := decimal.RequireFromString("0.01")

    flow := defi.NewFlow(user, defi.WithChain(config.Base)).
        Add(aave.ApproveSupply(config.USDC, supplyAmount)).
        Add(aave.Supply(config.USDC, supplyAmount)).
        Add(aave.Borrow(config.WETH, borrowAmount))

    return flow.Build(ctx, conn)
}
```

This example is the intended Phase 1 developer experience. It hides low-level action wiring while still producing neutral calls for whichever executor the caller chooses.
