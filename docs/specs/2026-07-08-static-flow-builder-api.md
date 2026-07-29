# Flow Builder And Runner API

Status: Implemented reference

Originally drafted: 2026-07-08

Updated: 2026-07-29

Audience: SDK users and contributors

## Purpose

The Flow API describes ordered DeFi operations in protocol language:

```go
flow := defi.NewFlow(user, defi.WithChain(config.Base)).
	Add(erc20.Approve(
		usdc.Underlying(),
		aave.PoolSpender(market),
		amount.Exact(supplyAmount),
	)).
	Add(aave.Supply(usdc, amount.Exact(supplyAmount))).
	Add(aave.Borrow(weth, amount.Exact(borrowAmount)))

result, err := defi.NewRunner(client, opts, config.Base).Execute(ctx, flow)
```

Callers do not select an executor or execution mode. The Flow's compiled plan
determines which entrypoint is used on the configured
`DefiSimplify7702Account`.

## Architecture

```text
Flow
  -> FlowStep.Build
  -> BuiltStep { PlannedCalls, EventExpectations }
  -> ExecutionPlan
  -> Runner
  -> DefiSimplify7702Account
  -> Receipt
  -> Validator
  -> ExecutionResult
```

Responsibilities:

- `Flow` owns account, chain, and declared step order.
- Protocol `FlowStep` implementations own calldata and event expectations.
- `ExecutionPlan` is immutable and protocol-neutral.
- `Runner` selects the account entrypoint from `ExecutionPlan.Kind()`.
- The account package translates neutral plan metadata into contract ABI types.
- The validator matches the mined receipt against plan expectations.

Generic execution code must not import protocol packages.

## Core Types

### Flow And BuildEnv

`Flow` contains one semantic account, one chain, and ordered steps.

```go
type BuildEnv struct {
	Account common.Address
	Chain   config.Chain
	Conn    EthereumClient
}
```

The account is the protocol-visible caller contract for account-derived owner,
sender, recipient, or `onBehalfOf` fields.

### FlowStep And BuiltStep

```go
type FlowStep interface {
	Build(ctx context.Context, env BuildEnv) (BuiltStep, error)
}

type BuiltStep struct {
	ID           StepID
	Name         string
	Calls        []PlannedCall
	Expectations []EventExpectation
}
```

A protocol step must derive its calls and expectations from the same resolved
account, target, asset, and amount values.

Step implementations set `Name`; `Flow.Build` assigns occurrence-based IDs such
as `aave.Supply#1` and `aave.Supply#2`.

### Call And PlannedCall

```go
type Call struct {
	Target common.Address
	Value  *big.Int
	Data   []byte
}

type PlannedCall struct {
	Call              Call
	CheckpointsBefore []CheckpointDeclaration
	Patches           []CalldataPatch
	ExpectsCallback   bool
}
```

`Call` is the protocol-neutral EVM call. A `PlannedCall` adds explicit runtime
metadata without exposing account-contract ABI types to protocol packages.

### ExecutionPlan

`Flow.Build` returns an immutable `*ExecutionPlan`.

```go
plan, err := flow.Build(ctx, client)
kind := plan.Kind()
steps := plan.Steps()
```

`PlanStatic` contains exact calls only. `PlanDynamic` contains at least one
explicit checkpoint or calldata patch.

Use `StaticCalls` only for static plans and `DynamicCalls` only for dynamic
plans. Incompatible access returns `ErrPlanKindMismatch`; placeholder dynamic
calldata must never be exposed as a static call list.

## Amount Sources

Exact values are resolved during Build:

```go
amount.Exact(decimal.RequireFromString("100"))
```

Runtime values are resolved inside account execution:

```go
amount.CurrentBalance(asset)
amount.CheckpointDelta(checkpoint)
amount.Scale(source, bps)
```

Every checkpoint delta names one explicit earlier checkpoint. Protocol packages
own patchable ABI offsets and must validate that the source token matches the
operation's semantic asset.

The SDK does not infer dependencies, pipe arbitrary return data, patch native
`value`, or expose raw ABI offsets through high-level APIs.

## Execution

Normal public execution has one method:

```go
result, err := runner.Execute(ctx, flow)
```

The Runner:

1. Confirms the Flow account matches the transaction signer.
2. Builds the immutable execution plan.
3. Rejects invalid semantic-validation ordering before submission.
4. Resolves the configured Defi Simplify account deployment for the chain.
5. Dispatches static plans to inherited `executeBatch`.
6. Dispatches dynamic plans to `executeBatchDynamic`.
7. Checks the EOA's pending delegation target.
8. Sends one self-call from the delegated EOA.
9. Preserves the mined receipt on transaction failure.
10. Validates protocol events and returns an `ExecutionResult`.

The application must verify the configured account implementation's runtime
code hash during initialization and explicitly install the EIP-7702 delegation
before execution.

Delegation persists until it is switched or cleared. A reverted Flow does not
roll it back.

## Semantic Validation

Each `EventExpectation`:

1. Identifies candidate logs.
2. Decodes a candidate or returns a hard error.
3. Matches semantic fields.
4. Accepts the event or skips that candidate.

The validator scans logs forward in declared step and expectation order. A log
can be consumed only once.

Important invariants:

- Candidate decode failure is a hard error.
- Field mismatch skips that candidate and keeps scanning.
- `Match` errors are hard errors, not ordinary mismatches.
- Expectations within one step must be declared in emission order.
- Steps without expectations are unvalidated.
- An expectation-bearing step cannot follow an unvalidated step because receipt
  logs do not expose call boundaries.

## Result And Failure Contract

Before a receipt exists:

```text
result = nil
error  = non-nil
```

After a transaction is mined, transaction or semantic failure preserves the
result:

```text
result         = non-nil
result.Receipt = non-nil
error          = non-nil
```

Steps not executed after a reverted atomic batch use `ValidationSkipped` with
`SkipExecutionFailed`.

`ExecutionError.Unwrap` preserves `errors.Is` and `errors.As` behavior for
sentinel and typed errors.

## Extension Rules

When adding a transactional protocol operation:

1. Put calldata and event semantics in the owning protocol package.
2. Add the public `FlowStep` with its low-level Action.
3. Build calls and expectations from the same resolved values.
4. Add Ginkgo public behavior tests.
5. Add Base-fork coverage through `Runner.Execute`.

Strategy packages may compose public FlowSteps but must not duplicate calldata,
decode events, read chain state, construct a Runner, sign, or submit.

## Related Documents

- [README](../../README.md)
- [Unified Runner migration](../migrations/v0-unified-eip7702-runner.md)
- [Phase 1 historical outcome](2026-07-07-phase-1-mvp-spec-and-glossary.md)
