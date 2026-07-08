# Phase 1 MVP Spec and Glossary

Status: Draft for review
Date: 2026-07-07
Related release: `v0.2.0` Pre-Phase 1 Architecture Baseline
Audience: project maintainers, SDK users, and engineers implementing Phase 1

## 1. Purpose

Phase 1 is intended to prove one thing:

> The SDK can compose existing Aave Actions into an EOA-native execution flow and execute it through EIP-7702 + `Simple7702Account`, so Aave observes the user's EOA as the execution context instead of an external Multicall contract.

This phase is not a full wallet product and not a generalized DeFi automation platform.

The goal is to revisit the original Multicall-based `defi-simplify` design and prove that the same Aave workflow can be expressed through an EIP-7702 delegated EOA path without depending on an external Multicall contract as the protocol-visible caller.

## 2. Background

The original SDK wraps Aave and ERC20 operations as Go Actions. Each Action builds the contract calldata for a specific on-chain operation. Multiple Actions can then be batched through Multicall.

That model is useful for generic batch calls, but it has an important limitation for DeFi protocols:

```text
EOA -> Multicall -> Aave

Aave sees msg.sender = Multicall
```

In this execution path, Aave does not see the user EOA as the caller. It sees the Multicall contract.

That changes the shape of many DeFi flows:

- Position ownership and caller semantics require extra care.
- Token receivers may need additional transfers back to the user.
- Approval and credit delegation flows become workaround-heavy.
- Protocol callbacks or caller-sensitive flows may not map cleanly.
- SDK users need to understand asset and authority movement between the user, Multicall, and Aave.

Pre-Phase 1 refactored the SDK toward a neutral execution model:

```text
Action -> Call -> Executor
```

In that model, Multicall is one executor backend instead of being part of the Action abstraction. Phase 1 builds on this by adding a `Simple7702Account` / EIP-7702 execution path.

## 3. Phase 1 MVP

Phase 1 has a narrow chain and protocol scope:

```text
Chain: Base
Protocol: Aave V3
Account model: EOA delegated to Simple7702Account through EIP-7702
Primary flow: approve USDC -> supply USDC -> borrow asset
```

The MVP flow is:

```text
1. USDC.approve(Aave Pool, supplyAmount)
2. AavePool.supply(USDC, supplyAmount, userEOA, 0)
3. AavePool.borrow(borrowAsset, borrowAmount, variableRateMode, 0, userEOA)
```

Expected execution path:

```text
EOA delegates to Simple7702Account
EOA calls itself through the delegated account execution entrypoint
Simple7702Account executes the batch calls
Calls reach USDC and Aave Pool
Aave observes the user EOA as the caller / account context
```

This MVP uses an exact-calldata batch. The approve, supply, and borrow parameters can all be calculated off-chain by the Go SDK before submission. Phase 1 does not require on-chain return-value piping from one call into the next call.

## 4. Out of Scope

Phase 1 does not include:

- A formal `Recipe` abstraction.
- A module wallet.
- ERC-4337 bundler or paymaster integration.
- `swap -> read amountOut -> supply` dynamic return-value piping.
- Multi-protocol routing.
- A custom Solidity policy module.
- Flash loans.
- Complex eMode or isolation-mode routing.
- A full wallet product.

These items may be useful later, but they are not required to prove the Phase 1 claim.

## 5. Glossary

### Action

`Action` is an existing core SDK concept.

An Action represents a protocol-specific operation and is responsible for turning intent into contract calldata.

Examples:

```text
ApproveAction
  -> USDC.approve(AavePool, amount)

SupplyAction
  -> AavePool.supply(USDC, amount, onBehalfOf, referralCode)

BorrowAction
  -> AavePool.borrow(asset, amount, rateMode, referralCode, onBehalfOf)
```

In Phase 1, an Action should not know whether it will be executed through Multicall, `Simple7702Account`, or a direct transaction.

### Call

`Call` is a neutral contract-call representation.

It does not belong to Aave, Multicall, or EIP-7702.

```go
type Call struct {
    Target common.Address
    Value  *big.Int
    Data   []byte
}
```

Conceptually:

> A Call is the low-level contract call produced after an Action has encoded its target, native value, and calldata.

The Phase 1 `Simple7702Account` executor consumes `[]Call` and turns those calls into delegated-account batch execution calldata.

### Executor

An `Executor` decides how one or more Calls are submitted on-chain.

The same `[]Call` can be executed through different backends:

```text
DirectExecutor
  Sends a single Call as a normal transaction.

MulticallExecutor
  Wraps Calls into Multicall3 aggregate3.

Simple7702Executor
  Wraps Calls into Simple7702Account batch execution and submits them through a delegated EOA path.
```

The executor's responsibility is not to build Aave calldata. Its responsibility is to select and drive the execution backend.

Executors are not semantically interchangeable for caller-sensitive protocols.

The important difference is the protocol-visible caller:

```text
DirectExecutor
  caller = user EOA
  atomic multi-call = no

MulticallExecutor
  caller = Multicall contract
  atomic multi-call = yes

Simple7702Executor
  caller = user EOA through delegated code
  atomic multi-call = yes
```

`Flow` should stay executor-agnostic, but user-facing execution APIs should not force users to understand every executor backend. The public SDK should move toward execution modes that describe account semantics first.

### Execution Mode

An `ExecutionMode` is the user-facing way to choose execution semantics.

Unlike an executor backend, an execution mode should answer the questions SDK users actually care about:

```text
Who is the protocol-visible caller?
Can the flow contain multiple calls?
Is the flow atomic?
Is this mode suitable for EOA-native DeFi positions?
```

Planned modes:

```text
ExecutionEOA
  Single EOA transaction.
  Preserves msg.sender = user EOA.
  Rejects multi-call flows.

ExecutionAtomicEOA
  Atomic EOA-native batch.
  Preserves msg.sender = user EOA.
  Backed by EIP-7702 / Simple7702Account.

ExecutionLegacyMulticall
  Atomic Multicall batch.
  msg.sender = Multicall contract.
  Intended for legacy flows, read batching, or caller-insensitive operations.
```

The low-level executor API can remain available for tests and advanced users, but primary documentation should lead with execution modes.

### Flow

`Flow` is the planned Phase 1 composition abstraction.

It means an ordered sequence of protocol steps that completes a concrete DeFi operation and builds into neutral Calls.

Example:

```text
approve USDC
-> supply USDC
-> borrow WETH
```

This can be described as an Aave supply/borrow flow or a composed Action flow.

IAN-34 defines a static Flow builder API for Phase 1. The Flow builder is intentionally thin: it improves SDK ergonomics by letting users compose protocol steps, but it still builds on the existing Action -> Call -> Executor architecture.

The intended shape is:

```text
FlowStep -> Action -> Call -> Executor
```

Phase 1 Flow is static. It only supports exact call parameters that can be computed before transaction submission. It does not support dynamic output piping, guards, strategy lifecycle, or a formal Recipe system.

### Recipe

`Recipe` is a possible future high-level abstraction. Phase 1 does not introduce it.

A Recipe may become useful once the SDK supports many workflows such as:

```text
Aave supply + borrow
Aave repay + withdraw
swap -> supply
rebalance position
multi-protocol leverage loop
```

Those workflows may eventually share a common lifecycle:

```text
input validation
action composition
simulation
execution
receipt parsing
slippage / amount bounds
policy checks
```

At that point, a formal `Recipe` abstraction could be justified.

Phase 1 has only one Aave MVP flow, so the public terminology remains `composed Action flow`. This keeps the current implementation honest and avoids implying that the SDK already contains a generalized workflow system.

### Simple7702Account

`Simple7702Account` is the delegated account implementation selected for Phase 1.

It comes from the eth-infinitism account-abstraction repository. Phase 1 uses it because the SDK should first prove the end-to-end execution path with an existing account implementation instead of implementing a custom account core.

In Phase 1, `Simple7702Account` is treated as:

```text
a reference delegated account implementation that provides batch execution capability
```

It is not a module wallet and does not permanently define the SDK's long-term account architecture.

### EIP-7702 Delegation

EIP-7702 allows an EOA to set delegated code.

For this project, the relevant model is:

```text
user EOA delegates to Simple7702Account implementation
user EOA gains delegated account execution behavior
SDK submits transactions through that delegated EOA path
```

Important lifecycle detail:

> EIP-7702 delegation is not naturally temporary.

After delegation is set, the EOA's delegation indicator persists until it is explicitly changed or cleared. A transaction revert does not automatically roll back delegation.

Phase 1 must keep two concepts separate:

```text
delegation lifecycle
  Account-level capability state. It persists and requires explicit set / inspect / clear handling.

flow execution
  One Aave flow execution. It can be constrained by exact calldata, nonce/deadline-like authorization, and test setup.
```

Phase 1 must include clear support for delegation revoke / clear in tests, because the same test EOA may need to be reused with different delegated implementations.

### Multicall Execution

Multicall execution is the legacy path:

```text
EOA -> Multicall -> Aave
```

Benefits:

- Easy to batch multiple calls into one transaction.
- Does not require EIP-7702 support.
- Useful for read batching or caller-insensitive flows.

Limitations:

- Downstream protocols see `msg.sender = Multicall`.
- The user EOA is not the direct caller.
- Aave supply / borrow requires extra workarounds.

Phase 1 does not remove Multicall. It remains a legacy executor and comparison baseline.

Multicall can technically execute Flow-built calls, but it is not the semantic target for EOA-native Aave composition. For example:

```go
flow := defi.NewFlow(user, defi.WithChain(config.Base)).
    Add(erc20.Approve(config.USDC, spender, amount))

receipt, err := flow.Execute(ctx, conn, multicallExecutor)
```

This submits `USDC.approve(spender, amount)` from the Multicall contract. The resulting allowance is:

```text
allowance[Multicall][spender] = amount
```

not:

```text
allowance[userEOA][spender] = amount
```

This is why Multicall is useful as a backend and comparison point, but not as the main executor for user-owned Aave positions.

### EOA-Native Execution

EOA-native execution is the Phase 1 path:

```text
EOA delegated to Simple7702Account -> Aave
```

The point is not only batching. The point is that batch execution happens in the user's account context.

That lets the Aave flow more closely match the user's intended operation:

```text
approve from user EOA
supply onBehalfOf user EOA
borrow onBehalfOf user EOA
```

The flow should not need to park tokens in Multicall or transfer borrowed assets back from Multicall to the user.

## 6. Execution Model

### Legacy Multicall Aave Flow

The existing legacy Multicall flow roughly works like this:

```text
1. User signs permit for Multicall.
2. Multicall transferFrom user -> Multicall.
3. Multicall approves Aave Pool.
4. Multicall supplies to Aave onBehalfOf user.
5. User signs credit delegation to Multicall.
6. Multicall borrows onBehalfOf user.
7. Multicall transfers borrowed asset back to user.
```

This flow remains useful because it proves that the original SDK can perform atomic execution and documents why `msg.sender` is a limitation for DeFi composition.

### Phase 1 EIP-7702 Aave Flow

The Phase 1 target flow is:

```text
1. User EOA delegates to Simple7702Account.
2. SDK builds approve / supply / borrow Actions.
3. Actions are encoded into neutral Calls.
4. Simple7702Executor builds batch execution calldata.
5. Transaction is submitted through the EIP-7702 delegated EOA path.
6. Aave position belongs to the user EOA.
7. SDK parses receipt and validates events / balances / Aave account data.
```

Compared with the legacy Multicall flow, the expected EOA-native flow removes workaround asset movement:

```text
No user -> Multicall token parking
No Multicall credit delegation workaround
No borrowed-token transfer back from Multicall
```

## 7. Delegation Lifecycle Requirements

Phase 1 must handle the EIP-7702 delegation lifecycle explicitly.

Minimum requirements:

```text
Set delegation
  Delegate the test EOA to the Simple7702Account implementation.

Inspect delegation
  Observe whether the EOA is currently delegated and identify the implementation where possible.

Execute through delegation
  Execute a simple batch through the delegated account path, then execute the Aave flow.

Clear / revoke delegation
  Explicitly clear or replace delegation after tests or before switching implementations.
```

Project documentation should avoid describing EIP-7702 as "temporarily turning an EOA into a smart account." A more accurate statement is:

> EIP-7702 delegation persists until explicitly changed or cleared. Phase 1 can make each flow authorization short-lived, but delegation itself is account state and must be managed explicitly.

## 8. Why Not Recipe Yet

The repository originally explored action / recipe-style DeFi composition. The stable code abstraction today is `Action`, plus the newer neutral executor boundary.

Phase 1 does not introduce a formal `Recipe` abstraction because:

1. The MVP has one Aave flow.
2. Approve / supply / borrow can be built as exact calldata off-chain.
3. No conditional branching is required.
4. No previous return value needs to be piped into a later call.
5. No generalized policy engine is required.
6. Naming the flow `Recipe` too early would imply a complete workflow system before the SDK has proven the account execution model.

The Phase 1 term is:

```text
composed Action flow
```

A formal Recipe system can be introduced later if the SDK needs:

- Shared composition lifecycle across multiple protocols.
- Per-flow input schemas, validation, simulation, execution, and parsing.
- Dynamic return-value piping, such as `swap -> amountOut -> supply`.
- User signatures over high-level intent instead of exact calldata.
- Protocol-specific policy modules or slippage / amount bounds.

## 9. Relationship to Existing Architecture

This document is not a private planning note and does not redefine the entire project.

It is the GitHub-facing Phase 1 reference for the public repository. It fixes the MVP scope and shared terminology needed to understand the next implementation issues.

This document builds on the `v0.2.0` Pre-Phase 1 architecture baseline:

```text
Action -> Call -> Executor
```

Phase 1 should not redesign that SDK boundary. Instead, it should explain what the `Simple7702Account` / EIP-7702 execution path must prove on top of the existing model, and which concepts are intentionally left out of scope.

Future Phase 1 implementation issues and README updates should use this document as the public terminology and scope reference. More exploratory research notes should not be required to understand the public repository.

## 10. Phase 1 Issue Map

The existing Phase 1 implementation issues can be understood as:

```text
IAN-7
  Add Simple7702Account integration surface.
  Teach the Go SDK about the Simple7702Account ABI / calldata surface.

IAN-8
  Implement EIP-7702 delegation lifecycle manager.
  Handle set / inspect / clear delegation.

IAN-9
  Implement Simple7702Account executor.
  Convert []Call into the Simple7702Account batch execution path.

IAN-10
  Build Base Aave EOA-native supply/borrow flow.
  Compose approve -> supply -> borrow as a composed Action flow.

IAN-11
  Parse and validate Aave execution results.
  Implement reusable SDK receipt / event parsing and result validation.

IAN-12
  Base fork integration tests umbrella.
  Track the overall Phase 1 fork-test work.

IAN-13
  Update README and project narrative for Phase 1.
  Document the completed engineering story and usage path in README.
```

Integration test issues:

```text
IAN-27
  Simple7702Account batch integration smoke test.

IAN-28
  7702 delegation revoke integration test.

IAN-29
  7702 Aave approve -> supply integration test.

IAN-30
  7702 Aave approve -> supply -> borrow integration test.

IAN-31
  Integration receipt and event assertions for Aave flow.
```

Boundary between `IAN-11` and `IAN-31`:

```text
IAN-11
  Implement reusable SDK receipt / event parsers.

IAN-31
  Use those parsers in integration tests to validate the actual Phase 1 Aave flow.
```

## 11. Success Criteria

Phase 1 is complete when the project can demonstrate:

- The SDK composes existing Aave Actions into approve -> supply -> borrow.
- Actions are converted into neutral `Call` values before execution.
- `Simple7702Executor` executes those Calls through a delegated EOA path.
- The test EOA can delegate to `Simple7702Account`.
- The SDK can execute batch calls through the delegated EOA.
- The Aave supply / borrow account context is the user EOA, not Multicall.
- Tests can explicitly set, inspect, and clear EIP-7702 delegation.
- Base fork integration tests validate ERC20 allowance, aToken balance, Aave user reserve data, debt-token state, and relevant events.
- README can explain the original Multicall limitation and how EIP-7702 improves caller semantics.

One-sentence definition of success:

> Phase 1 succeeds when `defi-simplify` can run the Aave MVP flow through an EIP-7702 delegated EOA path and prove, with Base fork tests and parsed execution results, that the flow no longer depends on an external Multicall contract as the protocol-visible caller.
