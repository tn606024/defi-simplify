# v0 migration: decimal amounts to amount sources

Amount-bearing ERC20 and Aave FlowSteps now accept `amount.Source` instead of
`decimal.Decimal`. This makes exact and runtime-resolved amounts explicit in
the public type system.

## Exact amounts

Wrap existing human-unit decimal values with `amount.Exact`:

```go
// Before
flow.Add(aave.Supply(usdc, decimal.NewFromInt(100)))

// After
flow.Add(aave.Supply(usdc, amount.Exact(decimal.NewFromInt(100))))
```

Exact-only flows still produce `defi.PlanStatic`. The unified Runner executes
them through the delegated account's inherited `executeBatch` entrypoint.

## Runtime amounts

Use an explicit token reference for runtime balance sources:

```go
flow.Add(erc20.Approve(
	usdc.Underlying(),
	aave.PoolSpender(market),
	amount.CurrentBalance(usdc.Underlying().Ref()),
))
```

Use a named earlier checkpoint when only the balance increase produced within
the flow should be consumed:

```go
output := amount.Checkpoint("swap-output", weth.Underlying().Ref())

flow.
	Add(defi.CheckpointBefore(swapStep, output)).
	Add(aave.Supply(weth, amount.CheckpointDelta(output)))
```

`amount.Scale(source, bps)` applies a runtime ratio. Basis points must be
between 1 and 10,000.

Runtime sources must identify the same chain-scoped token as the protocol
amount being patched. Checkpoint deltas must reference a unique declaration on
an earlier call.

## Plan accessors

`ExecutionPlan` now exposes defensive-copy accessors:

```go
account := plan.Account()
steps := plan.Steps()

staticCalls, err := plan.StaticCalls()
dynamicCalls, err := plan.DynamicCalls()
```

`plan.Kind()` returns `defi.PlanStatic` or `defi.PlanDynamic`. Calling the
incompatible accessor returns `defi.ErrPlanKindMismatch`.

`plan.Calls()` remains as a deprecated static-only convenience. It returns
`nil` for dynamic plans so placeholder calldata cannot be submitted through a
legacy static executor.

The unified Runner now submits dynamic plans through
`DefiSimplify7702Account.executeBatchDynamic`. See
[Migrate to the unified EIP-7702 Runner](v0-unified-eip7702-runner.md).

## Custom FlowSteps

`BuiltStep.Calls` now contains `[]defi.PlannedCall` instead of `[]defi.Call`.
Wrap exact calls in a `PlannedCall`:

```go
// Before
return defi.BuiltStep{
	Name:  "protocol.Operation",
	Calls: []defi.Call{call},
}, nil

// After
return defi.BuiltStep{
	Name:  "protocol.Operation",
	Calls: []defi.PlannedCall{{Call: call}},
}, nil
```

Protocol packages that support runtime amounts add a `defi.CalldataPatch` to
the same `PlannedCall`. The protocol owns the ABI word offset; callers should
select an `amount.Source` and must not provide raw offsets.
