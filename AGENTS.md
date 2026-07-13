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
- Preserve the receipt and transaction hash for every mined transaction,
  including reverted transactions and successful transactions whose semantic
  validation fails.
- Treat a candidate event decode failure as a hard error. Treat a decoded event
  field mismatch as a skippable candidate and continue scanning.
- Match errors are hard errors and must not represent ordinary field
  mismatches. A zero-value match result is a skip by design.
- Steps without expectations are unvalidated, not failed. Steps not reached
  after a hard validation error are skipped, not unvalidated.
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
- Do not claim integration validation unless the fork test was actually run.

## Git Workflow

- Use the Linear issue identifier in the branch name and pull request title.
- Do not add a `codex/` branch prefix.
- Keep commits focused and give each commit a message that explains its exact
  review scope.
- Do not stage, modify, or revert unrelated local changes.
- Open pull requests as drafts unless explicitly requested otherwise.
- Treat GitHub Actions as required remote verification, but still report the
  relevant local validation performed for the change.
