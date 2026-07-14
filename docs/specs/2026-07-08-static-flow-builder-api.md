# Static Flow Builder API

Status: Implemented Phase 1 reference
Originally drafted: 2026-07-08
Updated: 2026-07-14
Related spec: [Phase 1 MVP Spec and Glossary](2026-07-07-phase-1-mvp-spec-and-glossary.md)
Audience: SDK users and contributors

## 1. Purpose

The static Flow API lets callers describe ordered DeFi operations in protocol
language while keeping transaction submission and receipt validation separate.

```go
flow := defi.NewFlow(user, defi.WithChain(config.Base)).
    Add(aave.ApproveSupply(config.USDC, supplyAmount)).
    Add(aave.Supply(config.USDC, supplyAmount)).
    Add(aave.Borrow(config.WETH, borrowAmount))

result, err := defi.NewRunner(client, opts, config.Base).
    ExecuteWithResult(ctx, flow, defi.ExecutionAtomicEOA)
```

The public API begins with FlowSteps. Callers do not need to construct
low-level Actions for common ERC20 or Aave workflows.

## 2. Architecture

The implemented static pipeline is:

```text
Flow
  -> FlowStep.Build
  -> BuiltStep { Calls, EventExpectations }
  -> ExecutionPlan
  -> Executor
  -> Receipt

ExecutionPlan + Receipt
  -> Validator
  -> ExecutionResult
```

Responsibilities are deliberately separated:

- `Flow` preserves user-declared step order and shared account/chain context.
- Protocol `FlowStep` implementations own calldata and event expectations.
- `ExecutionPlan` is the immutable executor-neutral build result.
- `Runner` maps a user-facing execution mode to an executor.
- Executors submit calls without importing protocol packages.
- The validator matches the receipt against the plan's expectations.

Protocol-specific calldata, decoded events, and matching rules remain in the
owning protocol package. Neutral planning and validation contracts remain in
the root `defi` package.

## 3. Public Packages

The primary composition imports are:

```go
import (
    defi "github.com/tn606024/defi-simplify"
    "github.com/tn606024/defi-simplify/aave"
    "github.com/tn606024/defi-simplify/config"
    "github.com/tn606024/defi-simplify/erc20"
    "github.com/tn606024/defi-simplify/strategy"
)
```

The root package exposes Flow, plan, runner, validation, result, and amount
constraint types. Protocol packages expose FlowStep builders and typed events.
The strategy package returns ordinary Flows composed from public protocol
steps.

## 4. Core Types

### Flow and Build Context

`Flow` contains an account, explicit chain context, and an ordered list of
steps. The account is the semantic protocol-visible caller for steps that
derive owner, sender, recipient, or `onBehalfOf` values from shared context.

```go
type BuildEnv struct {
    Account common.Address
    Chain   config.Chain
    Conn    EthereumClient
}
```

`Flow.Build` rejects a nil or empty Flow, a zero account, a missing or
unsupported chain, nil steps, unnamed built steps, and built steps without
calls.

### FlowStep and BuiltStep

```go
type FlowStep interface {
    Build(ctx context.Context, env BuildEnv) (BuiltStep, error)
}

type BuiltStep struct {
    ID           StepID
    Name         string
    Calls        []Call
    Expectations []EventExpectation
}
```

A protocol step must build its Calls and expectations from the same resolved
account, target, asset, and amount data. This prevents the submitted calldata
and semantic receipt contract from drifting apart.

Step implementations set `Name` and leave `ID` empty. `Flow.Build` assigns
occurrence-based IDs such as `aave.Supply#1` and `aave.Supply#2`.

### Call

```go
type Call struct {
    Target common.Address
    Value  *big.Int
    Data   []byte
}
```

Calls are neutral executor inputs. Flow and ExecutionPlan clone mutable call
data before exposing it.

### ExecutionPlan

```go
type ExecutionPlan struct {
    Account common.Address
    Steps   []BuiltStep
}
```

`Flow.Build` returns `*ExecutionPlan`. `plan.Calls()` returns a cloned,
flattened call slice in step order for executor submission.

## 5. Build Semantics

```go
plan, err := flow.Build(ctx, client)
if err != nil {
    return err
}
calls := plan.Calls()
```

Build performs these operations in order:

1. Validate Flow account, chain, and step presence.
2. Pass the same BuildEnv to every step.
3. Build each step in insertion order.
4. Assign a stable occurrence-based StepID.
5. Clone calls and expectations into the plan.

Build does not sign or submit transactions and does not own transaction
options. Protocol steps convert `decimal.Decimal` amounts into token units by
using configured asset decimals.

Step failures are wrapped with their index and stable name:

```text
build flow step 2 aave.Supply: ...
```

## 6. Static Amount Contract

Static Flow amounts are known before Build:

```go
supplyAmount := decimal.RequireFromString("100")
borrowAmount := decimal.RequireFromString("0.01")
```

The current Flow model does not support:

- reading one call's return value into a later call;
- using a runtime token balance as a later amount;
- swap-output-to-supply data flow;
- dynamic leverage loops;
- runtime calldata patching.

`Exact`, `Positive`, `AtLeast`, and `AtMost` are receipt-validation
constraints. They validate emitted event values; they do not provide runtime
cross-step amount resolution.

## 7. Execution

The user-facing runner exposes two execution modes:

```go
receipt, err := runner.Execute(ctx, flow, defi.ExecutionAtomicEOA)
result, err := runner.ExecuteWithResult(ctx, flow, defi.ExecutionAtomicEOA)
```

### ExecutionEOA

Submits exactly one call as a normal EOA transaction. It preserves the EOA as
the protocol-visible caller and rejects multi-call plans.

### ExecutionAtomicEOA

Submits an ordered static batch through EIP-7702 delegated
`Simple7702Account` code. It preserves the EOA as the downstream caller and
requires the account to be delegated to the configured implementation.

The Flow account must match the runner's transaction signer. This prevents a
plan built for one account from being executed through another account's
context.

## 8. Semantic Validation

Each `EventExpectation` identifies candidate logs, decodes the protocol event,
and matches its semantic fields.

Receipt matching distinguishes these outcomes:

- unrelated log: ignore it;
- candidate decode failure: hard validation error;
- decoded field mismatch: record the mismatch and continue scanning;
- accepted event: consume the log and move the cursor forward.

Expectations are processed in step and declaration order. A consumed or earlier
log cannot be reused. Expectations within one BuiltStep must therefore be
declared in the same order their calls emit the events on-chain.

Receipt logs do not expose call boundaries. A step without expectations is
unvalidated and cannot safely consume logs. A semantic plan may contain a
validated prefix followed by an unvalidated suffix, but an expectation-bearing
step after an unvalidated step is rejected before submission.

## 9. Execution Results

`Runner.ExecuteWithResult` preserves mined transaction information across
execution and validation failures:

```text
pre-submission failure
  result = nil, error != nil

mined revert
  result contains receipt, error != nil

semantic validation failure
  result contains receipt and partial step results, error != nil

successful execution and validation
  result contains receipt and complete step results, error = nil
```

`ExecutionError.Unwrap` preserves executor and validator errors for
`errors.Is` and `errors.As`. Protocol events are available through typed
selection:

```go
supplies := defi.EventsOf[*aave.SupplyEvent](result)
```

## 10. Custom and Strategy Steps

`ActionStep` adapts a low-level Action into a FlowStep, but it has no semantic
event expectations. It is an escape hatch rather than the normal protocol API.
Custom unvalidated steps must not precede later validated steps.

Built-in strategies sit above protocol FlowSteps:

```text
strategy builder -> Flow -> protocol FlowSteps -> ExecutionPlan
```

Strategies may validate static input and choose a documented step sequence.
They do not duplicate protocol calldata or event rules, read protocol state,
construct executors, sign, submit, or execute transactions.

## 11. Contributor Contract

Changes to the static Flow API must preserve these invariants:

- Flow preserves insertion order.
- Calls and expectations come from the same resolved step data.
- ExecutionPlan.Account is the semantic caller contract.
- Generic execution code does not import protocol packages.
- Every transactional protocol Action with stable receipt semantics has a
  corresponding public FlowStep.
- Every new FlowStep has deterministic behavior tests and Base-fork integration
  coverage through the public Flow API.
- Every built-in strategy proves plan equivalence with its manual FlowStep
  composition and verifies final protocol state on a Base fork.

The repository's `AGENTS.md` contains the complete architecture, testing, and
Git workflow requirements.
