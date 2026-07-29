# Phase 1 MVP Spec And Glossary

Status: Superseded historical design note

Originally drafted: 2026-07-07

Superseded by: unified EIP-7702 Runner architecture

## Original Goal

Phase 1 set out to prove that an SDK-composed Aave flow could execute
atomically from an EIP-7702 delegated EOA while preserving that EOA as Aave's
protocol-visible caller.

The target flow was:

```text
ERC20 approve -> Aave supply -> Aave borrow
```

The original external batch design changed `msg.sender` to the batching
contract. That made approvals, position ownership, receivers, and
caller-sensitive protocol behavior harder to compose correctly.

## Outcome

The repository now implements a broader version of the original goal:

```text
Flow
  -> ExecutionPlan
  -> DefiSimplify7702Account
  -> mined receipt
  -> typed semantic validation
```

Normal public Flow execution has one API:

```go
result, err := runner.Execute(ctx, flow)
```

The plan determines the account entrypoint:

```text
PlanStatic  -> inherited executeBatch
PlanDynamic -> executeBatchDynamic
```

The previous direct-EOA, external Multicall, and separate static-account modes
are no longer supported public Flow paths.

## Current Glossary

### FlowStep

A protocol-owned operation that builds neutral planned calls and typed event
expectations from one shared `BuildEnv`.

Examples include `erc20.Approve`, `aave.Supply`, `aave.Borrow`, and
`weth.Wrap`.

### Flow

An ordered sequence of `FlowStep` values for one chain and one semantic account.
The Flow account is the expected downstream caller for account-derived
protocol fields.

### PlannedCall

A neutral contract call plus optional checkpoint or calldata-patch metadata.
Protocol packages own ABI offsets and semantic assets; generic execution code
does not infer them.

### ExecutionPlan

The immutable result of building a Flow. Exact-only calls produce
`PlanStatic`. Explicit runtime dependencies produce `PlanDynamic`.

### Runner

The public execution coordinator. It builds a Flow, validates plan invariants,
selects the configured Defi Simplify account entrypoint from the plan kind,
submits one EOA self-call, and validates the mined receipt.

### EIP-7702 Delegation

Persistent account code delegation installed on an EOA. Delegation survives
successful and reverted Flow transactions until it is explicitly switched or
cleared.

### ExecutionResult

The mined receipt plus per-step validation state and typed decoded events.
Mined failures preserve the result and return an error.

## Current References

- [README](../../README.md)
- [Static Flow Builder API](2026-07-08-static-flow-builder-api.md)
- [Unified Runner migration](../migrations/v0-unified-eip7702-runner.md)
