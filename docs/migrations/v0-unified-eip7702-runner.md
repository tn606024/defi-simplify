# Migrate To The Unified EIP-7702 Runner

DS-63 removes legacy execution backends and makes EIP-7702 account execution
the only normal public Flow path.

This is a breaking v0 change.

## Runner API

Before:

```go
receipt, err := runner.Execute(ctx, flow, defi.ExecutionAtomicEOA)
result, err := runner.ExecuteWithResult(ctx, flow, defi.ExecutionDynamicEOA)
```

After:

```go
result, err := runner.Execute(ctx, flow)
receipt := result.Receipt
```

`Runner.Execute` now always returns `*defi.ExecutionResult`. It builds the Flow,
executes the plan, and validates the mined receipt.

The following public execution modes were removed:

- `ExecutionEOA`
- `ExecutionAtomicEOA`
- `ExecutionDynamicEOA`
- `ExecutionMode`

`Runner.ExecuteWithResult` was removed because result validation is no longer
optional or a separate execution path.

## Automatic Plan Dispatch

Callers no longer select an executor from the outside:

```text
PlanStatic  -> DefiSimplify7702Account.executeBatch
PlanDynamic -> DefiSimplify7702Account.executeBatchDynamic
```

Exact-only Flows still compile to static plans. They now execute through the
configured delegated account rather than through a separate
`Simple7702Account` selection.

Runtime balance sources, checkpoints, or calldata patches compile to dynamic
plans and execute through `executeBatchDynamic`.

## Removed Flow And Executor APIs

The following normal execution paths were removed:

- `Flow.Execute(ctx, conn, executor)`
- root `CallExecutor`
- `DirectExecutor`
- `MulticallExecutor`
- `ActionExecutor`
- `BaseClient.SetCallExecutor`
- `BaseClient.SetActionExecutor`
- `BaseClient.ExecuteCalls`
- `BaseClient.ExecuteTxActions`

Use `Runner.Execute(ctx, flow)` for public Flow execution.

`client/contract.SendCall` remains a low-level single-transaction utility for
account execution and lifecycle tooling. It is not a Flow backend and does not
provide batching.

## Removed Account Selection

The standalone `client/account/simple7702` package and these convenience
helpers were removed:

- `config.Chain.Simple7702AccountImplementationAddress`
- `eip7702.Manager.DelegateToSimple7702`
- `eip7702.SignSimple7702Authorization`

Resolve the checked-in Defi Simplify deployment instead:

```go
deployment, err := defisimplify7702.DeploymentForChain(config.Base)
if err != nil {
	return err
}
accountDeployment, err := deployment.Contract(defisimplify7702.AccountContract)
if err != nil {
	return err
}
```

Verify the configured implementation code during application initialization:

```go
code, err := client.CodeAt(ctx, accountDeployment.Address, nil)
if err != nil {
	return err
}
if err := accountDeployment.VerifyRuntimeCode(code); err != nil {
	return err
}
```

Delegate through the generic lifecycle manager:

```go
tx, err := manager.Delegate(ctx, accountDeployment.Address)
```

The Runner verifies the EOA's pending delegation target before submission.

## Delegation Lifecycle

Delegation remains installed after a successful or reverted Flow. It is not a
transaction-scoped permission.

Applications should:

1. Verify the configured implementation identity at startup.
2. Delegate the EOA explicitly.
3. Serialize Flow execution and delegation changes for that EOA.
4. Switch delegation explicitly when another implementation is required.
5. Clear delegation explicitly with `manager.Clear(ctx)`.

The pending-state preflight cannot eliminate the race between checking a
delegation and transaction inclusion.

## Multicall Removal

External Multicall execution was removed because it changes the downstream
protocol caller to the Multicall contract. That violates
`ExecutionPlan.Account` semantics for approvals, Aave position ownership,
receivers, and callback-sensitive operations.

Deprecated aggregate read helpers still return the same logical data, but now
perform sequential RPC calls instead of routing reads through Multicall.

## Error And Receipt Semantics

Failures before a transaction receipt exists return:

```go
result == nil
err != nil
```

Once a transaction is mined, a transaction revert or semantic validation
failure returns:

```go
result != nil
result.Receipt != nil
err != nil
```

Continue using `errors.Is` and `errors.As` for wrapped sentinel and typed
errors. Do not discard a non-nil result when handling an error.
