# Repository Instructions

## Required Context

Before changing Flow, FlowStep, Runner, Executor, receipt validation, protocol
steps, or public APIs:

1. When `ARCHITECTURE.md` is present, read it, especially sections 6 through 8.
2. Preserve the documented Flow -> ExecutionPlan -> Executor -> Validator
   boundaries.
3. Do not introduce flow-specific or protocol-specific assumptions into generic
   execution code.
4. When the repository-local `ARCHITECTURE.md` is present, update it if a
   change alters a documented public contract, execution invariant, or
   ownership boundary.

## Architecture Rules

- Keep protocol-specific calldata, event decoding, and event validation inside
  the owning protocol package.
- Generic execution code must not import Aave, ERC20, or future protocol
  packages such as Uniswap.
- Keep neutral execution types, including `EventExpectation`, `BuiltStep`,
  `MatchResult`, and `FieldMismatch`, in the root `defi` package. Protocol
  packages implement these contracts by importing `defi`; the root package
  must never import protocol packages.
- Build calls and their event expectations from the same resolved step data so
  account, address, asset, and amount values cannot drift apart.
- When adding a transactional protocol `Action`, add the corresponding public
  `FlowStep` in the owning protocol package in the same change. If the Action
  cannot safely expose stable semantic expectations, document the concrete
  reason and keep it explicitly low-level; do not silently leave public
  protocol operations accessible only through an unvalidated `ActionStep`.
- Read/query Actions and executor-internal Actions are not FlowSteps. Keep this
  classification explicit when auditing Action-to-FlowStep coverage.
- Treat `ExecutionPlan.Account` as the protocol-visible caller contract for
  account-derived steps. Semantic execution must use an executor that preserves
  that account as the downstream call origin; external Multicall execution does
  not satisfy this contract.
- Preserve the receipt and transaction hash for every mined transaction,
  including reverted transactions and successful transactions whose semantic
  validation fails.
- Treat a candidate event decode failure as a hard error. Treat a decoded event
  field mismatch as a skippable candidate and continue scanning.
- Match errors are hard errors and must not represent ordinary field
  mismatches. A zero-value match result is a skip by design.
- Steps without expectations are unvalidated, not failed. Steps not reached
  after a hard validation error are skipped, not unvalidated.
- Receipt logs do not expose call boundaries, so unvalidated steps cannot
  consume their emitted logs. Semantic plans may contain a validated prefix and
  an unvalidated suffix, but must reject any expectation-bearing step after an
  unvalidated step before transaction submission.
- Match expected events in step order, scan forward from the last consumed log,
  and never reuse a consumed or earlier log. Within a `BuiltStep`, expectation
  authors must declare expectations in the same order that the step's calls emit
  their corresponding events on-chain.
- Guard all `*big.Int` inputs against nil and clone values stored in public
  results.
- Keep amount validation extensible through constraints. Implement only the
  constraints required by the current scope; do not add cross-step runtime data
  flow before the architecture supports it.
- Evaluate every amount constraint for a decoded value and aggregate all
  `FieldMismatch` values before returning a mismatch. A hard constraint error
  still aborts immediately.
- At the event level, `MatchSkip` tells the validator to continue scanning logs.
  At the constraint level, `MatchSkip` only means that the value does not satisfy
  the constraint; the enclosing event expectation decides whether to skip the
  candidate log.

## Testing Strategy

- Use simple Go table-driven tests for low-level, deterministic logic such as
  constraints, match decisions, ID generation, cursor behavior, error wrapping,
  and nil handling.
- Use Ginkgo and Gomega for behavior-oriented tests covering Flow composition,
  Runner behavior, execution results, partial failures, and public SDK
  workflows.
- Use Ginkgo and Gomega for integration tests against an Anvil Base mainnet
  fork.
- Every new or behaviorally changed transactional protocol `FlowStep` must have
  corresponding Base-fork integration coverage in the same change. The test
  must execute the public Flow API and assert the mined receipt plus the core
  typed protocol event fields, including caller/account fields when they are
  part of the step contract.
- For gateway, permit, delegation, or other adapter-backed steps, integration
  tests must verify the real on-chain event order and adapter-visible caller
  semantics; calldata-only tests are not sufficient.
- Keep unit tests synthetic and deterministic. Use ABI-encoded logs when testing
  event decoding; do not require RPC access for unit tests.
- Test both the returned result and error chain for mined failures. Use
  `errors.Is` and `errors.As` where wrapped sentinel or typed errors are part of
  the contract.
- Do not force BDD structure onto small pure functions, and do not express
  multi-step SDK behavior as large table tests.

## Validation Commands

- Run `go test -count=1 ./...` after Go changes.
- Compile integration tests with
  `go test -count=1 -run '^$' -tags=integration ./integration/...`.
- For execution, delegation, receipt, event, or Base protocol changes, run the
  full Base fork suite with
  `BASE_RPC_URL=http://127.0.0.1:8545 make test-integration` against a local
  Anvil fork.
- Start EIP-7702 fork tests through `make anvil-base`, which must select a
  hardfork that supports set-code transactions. Do not run 7702 integration
  tests against Anvil's unspecified default hardfork.
- Do not claim integration validation unless the fork test was actually run.

## Git Workflow

- Use the Linear issue identifier in the branch name and pull request title.
- Do not add a `codex/` branch prefix.
- Keep commits focused and give each commit a message that explains its exact
  review scope.
- Document migration steps whenever a v0 public API change breaks existing
  callers; do not rely on compiler errors as release notes.
- Do not stage, modify, or revert unrelated local changes.
- Open pull requests as drafts unless explicitly requested otherwise.
- Treat GitHub Actions as required remote verification, but still report the
  relevant local validation performed for the change.
