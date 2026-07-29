# DeFi Simplify

`defi-simplify` is a Go SDK for composing DeFi flows and executing them from
the user's own EOA.

The SDK turns protocol-level steps such as ERC20 approval, WETH wrap/unwrap,
Aave supply, and Aave borrow into an ordered execution plan. On Base, static
plans can execute through `Simple7702Account`, while runtime balance plans
execute through `DefiSimplify7702Account.executeBatchDynamic`. Both paths
preserve the user's EOA as the downstream caller and position owner.

The project is intended for Go services, bots, and infrastructure that manage
their own keys and transaction submission.

## Supported Scope

| Area | Current support |
| --- | --- |
| Network | Base |
| Protocols | Aave V3, ERC20, and WETH |
| Composition | Ordered `Flow` values with exact or runtime token-balance amount sources |
| EOA-native execution | Static EIP-7702 batches through `Simple7702Account`; runtime balance batches through `DefiSimplify7702Account` |
| Results | Typed ERC20, WETH, Aave Pool, gateway, and credit-delegation events |
| Strategies | Aave supply/borrow, single-reserve close, and WETH-composed native-ETH flows |
| Dynamic execution | `CurrentBalance` and `CheckpointDelta` calldata patching through `ExecutionDynamicEOA` |

## Installation

```bash
go get github.com/tn606024/defi-simplify
```

The module currently requires Go 1.23.5 or later.

## Quick Start

The primary SDK path is:

```text
FlowStep -> ExecutionPlan -> Runner -> Executor -> Receipt -> Validator
```

The example below assumes `client` is a connected `ethclient.Client`, `opts`
is a `bind.TransactOpts` for `user`, and the EOA has already delegated to the
configured `Simple7702Account` implementation.

```go
supplyAmount := decimal.NewFromInt(100)
borrowAmount := decimal.RequireFromString("0.01")

market, err := aave.BaseV3Market()
if err != nil {
	return err
}
registry, err := aave.NewRegistry(client, market)
if err != nil {
	return err
}
snapshot, err := registry.Load(ctx)
if err != nil {
	return err
}
usdc, err := snapshot.Reserve(base.USDC)
if err != nil {
	return err
}
weth, err := snapshot.Reserve(base.WETH)
if err != nil {
	return err
}

flow := defi.NewFlow(user, defi.WithChain(market.Chain())).
	Add(aave.ApproveSupply(usdc, amount.Exact(supplyAmount))).
	Add(aave.Supply(usdc, amount.Exact(supplyAmount))).
	Add(aave.Borrow(weth, amount.Exact(borrowAmount)))

result, err := defi.NewRunner(client, opts, config.Base).
	ExecuteWithResult(ctx, flow, defi.ExecutionAtomicEOA)
if result != nil {
	log.Printf("transaction: %s", result.Receipt.TxHash.Hex())
}
if err != nil {
	return err
}

supplies := defi.EventsOf[*aave.SupplyEvent](result)
borrows := defi.EventsOf[*aave.BorrowEvent](result)
```

`ExecutionAtomicEOA` executes the three calls as one transaction. If a call
fails, protocol and asset changes made by the batch revert atomically. Gas and
nonces are still consumed.

## EIP-7702 Delegation

Atomic EOA execution requires the EOA to delegate to the implementation used
by the selected execution mode. The SDK provides a lifecycle manager for
installing, inspecting, changing, and clearing that delegation.

For a same-signer setup, create the manager with the EOA key and submit the
delegation transaction once:

```go
chainID, err := config.Base.ChainID()
if err != nil {
	return err
}

manager, err := eip7702.NewManager(
	client,
	opts,
	privateKey,
	big.NewInt(int64(chainID)),
)
if err != nil {
	return err
}

tx, err := manager.DelegateToSimple7702(ctx, config.Base)
if err != nil {
	return err
}

receipt, err := bind.WaitMined(ctx, client, tx)
if err != nil {
	return err
}
if receipt.Status != types.ReceiptStatusSuccessful {
	return fmt.Errorf("delegation transaction reverted")
}
```

EIP-7702 delegation is persistent. It remains installed until the EOA changes
or clears it, and a later execution revert does not roll it back. Clear it
explicitly when it is no longer required:

```go
tx, err := manager.Clear(ctx)
```

Use `State`, `AssertClean`, and `AssertDelegatedTo` to verify lifecycle state
before submitting account-sensitive transactions.

Dynamic execution delegates to the reviewed deployment selected from the
checked-in manifest:

```go
deployment, err := defisimplify7702.DeploymentForChain(config.Base)
if err != nil {
	return err
}
account, err := deployment.Contract(defisimplify7702.AccountContract)
if err != nil {
	return err
}
tx, err := manager.Delegate(ctx, account.Address)
```

An EOA has one active delegation target. Change or clear the delegation
explicitly when switching between the static Simple7702 implementation and the
dynamic Defi Simplify implementation.

## Execution Modes

### `ExecutionEOA`

Executes exactly one call as a normal EOA transaction. The protocol observes
the EOA as `msg.sender`, but multiple calls cannot be combined atomically.

### `ExecutionAtomicEOA`

Executes an ordered static batch through the EOA's delegated
`Simple7702Account` code. Protocols still observe the EOA as the downstream
caller. The runner verifies the expected delegation before submission.

### `ExecutionDynamicEOA`

Executes a `PlanDynamic` atomically through the EOA's delegated
`DefiSimplify7702Account.executeBatchDynamic`. Before signing or submission,
the executor validates the account wire shape, selects the reviewed immutable
implementation address from the checked-in deployment manifest, and checks the
EOA's pending delegation target before every transaction. Mined reverts
preserve the receipt and expose decoded account custom errors, including call,
checkpoint, and patch indices, through `errors.As`.

For all modes, the Flow account must match the transaction signer when the
step derives owner, sender, recipient, or `onBehalfOf` fields from the account.

Dynamic plans are intentionally rejected by these static execution modes before
signing or submission. `ExecutionDynamicEOA` likewise rejects static plans;
execution capability is explicit rather than inferred by Runner.

## Amount Sources

Amount-bearing FlowSteps accept `amount.Source`:

- `amount.Exact(value)` converts a human-unit decimal while building the plan.
- `amount.CurrentBalance(tokenRef)` reads the delegated EOA's entire token
  balance immediately before the call.
- `amount.CheckpointDelta(checkpoint)` reads the balance increase since one
  explicitly declared earlier checkpoint.
- `amount.Scale(source, bps)` applies 1 to 10,000 basis points to a runtime
  source.

An exact-only flow produces `defi.PlanStatic`. A checkpoint or runtime patch
produces `defi.PlanDynamic`:

```go
output := amount.Checkpoint("swap-output", base.WETH)

flow := defi.NewFlow(user, defi.WithChain(config.Base)).
	Add(defi.CheckpointBefore(swapStep, output)).
	Add(aave.Supply(weth, amount.CheckpointDelta(output)))

result, err := defi.NewRunner(client, opts, config.Base).
	ExecuteWithResult(ctx, flow, defi.ExecutionDynamicEOA)
```

The producer step above is illustrative until the SDK adds its Uniswap
FlowSteps. A delta may only consume a checkpoint declared by an earlier call;
same-call consumption, duplicate declarations, invalid ABI offsets, and
non-zero placeholders are rejected during `Flow.Build`.

Callers that build before execution can inspect `plan.Kind()`.
`plan.StaticCalls()` and `plan.DynamicCalls()` enforce plan capability.
Deprecated `plan.Calls()` returns calls only for static plans, so placeholder
calldata cannot leak into an existing static executor.

## Flow Steps

Protocol packages expose public `FlowStep` builders. They resolve calldata and
typed event expectations from the same account, chain, asset, and amount data.

The ERC20 package includes:

- `Approve`
- `Transfer`
- `TransferFrom`
- `Permit`

The WETH package includes:

- `Wrap` with an exact native ETH value
- `Unwrap` with an exact, current-balance, or checkpoint-delta amount

`Wrap` places its amount in `Call.Value`. Runtime native-value patching is not
supported, so runtime amount sources are accepted only by `Unwrap`.

The Aave package includes:

- supply: `ApproveSupply`, `Supply`, `SupplyWithPermit`
- position management: `Borrow`, `Repay`, `RepayAll`, `Withdraw`, `WithdrawAll`
- credit delegation: `ApproveDelegation`, `DelegationWithSig`
- native ETH gateway: `DepositETH`, `BorrowETH`, `WithdrawETH`, `WithdrawETHWithPermit`

Permit and delegation signatures are prepared before `Flow.Build`; building a
Flow is deterministic and does not own a signer. Permit-capable tokens and
credit-delegation debt tokens require explicit `erc20.PermitCapability` or
`aave.DelegationCapability` values with a reviewed EIP-712 domain version. The
SDK does not infer signature support from symbols or the presence of a
`nonces()` method.

`RepayAll` and `WithdrawAll` encode Aave's `uint256.max` sentinel while receipt
validation checks the actual positive amount emitted by Aave. `RepayAll`
requires enough token balance and allowance to cover the debt at execution
time, including accrued interest and protocol rounding.

## Built-in Strategies

Strategies are thin templates over public FlowSteps. They validate static
inputs and return `*defi.Flow`; they do not read chain state, sign, submit, or
execute transactions.

Open an exact Aave supply and borrow position:

```go
flow, err := strategy.AaveSupplyBorrow(strategy.AaveSupplyBorrowParams{
	Account:       user,
	SupplyReserve: usdc,
	SupplyAmount:  decimal.NewFromInt(100),
	BorrowReserve: weth,
	BorrowAmount:  decimal.RequireFromString("0.01"),
})
```

Close one variable-debt and collateral reserve pair:

```go
flow, err := strategy.AaveClosePosition(strategy.AaveClosePositionParams{
	Account:                 user,
	DebtReserve:             usdc,
	TemporaryRepayAllowance: decimal.NewFromInt(102),
	CollateralReserve:       weth,
})
```

The close strategy builds:

```text
Approve(temporary allowance) -> RepayAll -> Approve(0) -> WithdrawAll
```

The temporary allowance is an upper bound rather than the actual debt. The
final approval clears any unused allowance. The first version assumes standard
ERC20 allowance replacement semantics and closes only the selected reserve
pair. If another debt makes `WithdrawAll` unsafe, the atomic transaction
reverts.

Native ETH strategies use the Aave WETH reserve directly instead of
`WrappedTokenGateway`:

```go
supplyFlow, err := strategy.AaveSupplyNativeETH(strategy.AaveSupplyNativeETHParams{
	Account: user,
	Reserve: weth,
	Amount:  decimal.RequireFromString("0.1"),
})

borrowFlow, err := strategy.AaveBorrowNativeETH(strategy.AaveBorrowNativeETHParams{
	Account: user,
	Reserve: weth,
	Amount:  decimal.RequireFromString("0.01"),
})
```

The strategy builders represent these manual FlowStep compositions:

```text
AaveSupplyNativeETH
  WETH.Wrap -> ERC20.Approve(Aave Pool) -> Aave.Supply

AaveBorrowNativeETH
  checkpoint WETH -> Aave.Borrow(WETH) -> WETH.Unwrap(checkpoint delta)

AaveWithdrawNativeETH
  checkpoint WETH -> Aave.Withdraw(WETH, exact) -> WETH.Unwrap(checkpoint delta)

AaveWithdrawAllNativeETH
  checkpoint WETH -> Aave.WithdrawAll(WETH) -> WETH.Unwrap(checkpoint delta)
```

Native supply is an exact static plan and executes with `ExecutionAtomicEOA`.
Borrow and withdrawal plans use `ExecutionDynamicEOA`; their checkpoints
ensure that pre-existing WETH is not unwrapped.

## Execution Results

`Runner.ExecuteWithResult` validates the mined receipt against expectations
produced by the same FlowSteps that built the calls.

Result behavior is explicit:

- failures before submission or mining return `result == nil` and an error;
- a mined transaction revert returns both the receipt-bearing result and an error;
- semantic event validation failure returns a partial result and an error;
- successful execution and validation return the complete result and no error.

Always inspect a non-nil result before returning the error when the transaction
hash is operationally important. Wrapped execution and validation errors retain
their `errors.Is` and `errors.As` chains.

Typed events can be selected without parsing raw logs:

```go
approvals := defi.EventsOf[*erc20.ApprovalEvent](result)
repayments := defi.EventsOf[*aave.RepayEvent](result)
withdrawals := defi.EventsOf[*aave.WithdrawEvent](result)
```

## Aave Market Discovery

Applications can resolve Aave reserve membership and reserve-token roles from
a trusted market definition instead of maintaining a symbol-keyed token map:

```go
market, err := aave.BaseV3Market()
if err != nil {
	return err
}
registry, err := aave.NewRegistry(client, market)
if err != nil {
	return err
}

snapshot, err := registry.Load(ctx)
if err != nil {
	return err
}
usdc, err := snapshot.Reserve(base.USDC)
```

The reviewed `assets/base` catalog provides convenient `token.Ref` values such
as `base.USDC` and `base.WETH`. Each value contains only the Base chain and
underlying contract address. Catalog membership is not an execution allowlist
or a promise that an asset is currently active in Aave. Applications can use an
uncatalogued asset by constructing `token.NewRef(config.Base, address)` and
resolving it through the snapshot.

`base.Lookup` accepts exact, case-sensitive SDK catalog IDs. These IDs are not
ERC20 symbols and are never matched against on-chain display metadata.

One discovery load resolves a block number and hash, pins every Pool,
DataProvider, ERC20 metadata, and contract-code read to that block, validates
the market's PoolAddressesProvider relationships, and returns an immutable
`MarketSnapshot`. Address is the execution identity; symbols and names are
display metadata only.

`Load` caches the first successful snapshot for that registry. It never
refreshes implicitly. Call `Refresh` when the application intentionally wants
a newer block; a failed refresh leaves the previous snapshot cached. Flow
building does not own or trigger registry refreshes.

## Architecture

The public composition and execution boundary is:

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

Protocol packages own calldata, decoded event types, and semantic
expectations. The root package owns neutral Flow, plan, validation, and result
types. Executors know how to submit calls but do not import protocol packages.

Low-level Actions and Calls remain reusable implementation primitives under
the public FlowStep API. Application code should normally begin with FlowSteps
or a built-in strategy. The `config.Coin`-based protocol clients under
`client/contract` remain available for migration but are deprecated; they are
not the source of truth for executable Flow assets.

## Repository Layout

| Path | Responsibility |
| --- | --- |
| `/` | Flow, execution plan, runner, validator, constraints, and results |
| `assets/` | Chain-neutral catalog runtime and chain-specific reviewed asset packages |
| `aave/` | Aave FlowSteps, typed events, and event expectations |
| `erc20/` | ERC20 FlowSteps, typed events, and event expectations |
| `weth/` | WETH ABI, FlowSteps, typed events, and event expectations |
| `strategy/` | Opinionated Flow compositions over protocol FlowSteps |
| `client/account/eip7702/` | Delegation authorization, transactions, state, and lifecycle manager |
| `client/account/simple7702/` | `Simple7702Account` ABI, calldata, and executor |
| `client/account/defisimplify7702/` | Pinned custom-account release, Base deployment identity, ABIs, bindings, and parity vectors |
| `client/contract/` | Low-level Actions, Calls, protocol clients, and executor primitives |
| `config/` | Supported chains and legacy static SDK configuration |
| `integration/` | Ginkgo tests against an Anvil Base mainnet fork |

## Deployment and Asset Trust

The `client/account/defisimplify7702` package embeds the imported
[`defi-simplify-contracts`](https://github.com/tn606024/defi-simplify-contracts)
Base deployment manifest, generated Go bindings, ABI files, and golden parity
vectors. Runtime code hashes remain part of the deployment identity and are
verified against the Base fork by integration tests rather than adding an RPC
read to each SDK execution. The contracts remain experimental and unaudited;
checked-in artifacts and source verification do not imply an audit.

The Aave registry starts from a checked-in Base V3 deployment manifest under
`aave/manifests/`. It contains only reviewed deployment anchors such as the
Pool, PoolAddressesProvider, AaveProtocolDataProvider, and wrapped-token
gateway. Dynamic reserve membership and reserve-token addresses are not copied
into the manifest.

The root `assets` package provides the chain-neutral immutable catalog runtime.
Each chain has its own thin package, such as `assets/base`, so token names remain
unambiguous: future catalogs can expose `ethereum.USDC` or `arbitrum.USDC`
without introducing a global symbol registry.

Each chain package keeps its loader in `catalog.go`. Named references such as
`base.USDC` are generated from the reviewed manifest into `catalog_gen.go`; the
generated file must not be edited by hand. New public references therefore
appear as ordinary, reviewable Go diffs alongside the manifest update without
duplicating the asset list in hand-written code.


The `assets/base/manifest.json` file is the first reviewed convenience catalog.
It copies only chain-scoped underlying identities from the pinned
`AaveV3Base.ASSETS` export. It does not copy decimals, symbols, aToken or debt
token addresses, or protocol capabilities. Those values remain runtime-owned
metadata and Aave registry relationships. Provider-specific extraction is kept
in an internal adapter; neutral manifest validation and catalog lookup do not
assume Aave, Base, or chain ID 8453.

Both manifests record the exact official
`@aave-dao/aave-address-book` release and commit that produced it. SDK runtime
code reads the embedded manifest only; it never fetches a mutable remote branch
or npm package. A scheduled workflow can propose an upstream update as a draft
pull request, but deployment and catalog changes always require human review.
Routine updates may add reviewed catalog candidates, but they fail closed if an
existing public catalog ID disappears, changes upstream key, or points at a new
address. Such a change requires an explicit migration and deprecation decision.

Maintainers can reproduce the checked-in manifests with:

```bash
make update-aave-manifests
```

This update-only command installs the exact npm package pinned under
`tools/aave-address-book/`, extracts the Base deployment anchors and underlying
asset identities, and passes the normalized export through strict Go validators
and canonical JSON generators. It also generates the chain package's named Go
references from the validated asset manifest. The singular
`make update-aave-manifest` target remains as an alias. Ordinary SDK builds and
tests do not require Node.js or network access.

Maintainers can refresh the contracts package from a local checkout:

```bash
DEFI_SIMPLIFY_CONTRACTS_DIR=/path/to/defi-simplify-contracts \
  make update-contract-artifacts
```

The updater reads the checkout's current ABI, Base deployment manifest, and
golden vectors, then regenerates the bindings with go-ethereum's `abigen`.
Go tests validate the checked-in JSON. Normal CI uses
`make verify-contract-artifacts` to regenerate the bindings directly with
`abigen` and fail if they no longer match the checked-in ABI.

Adding another chain requires registering the SDK chain, adding a thin
`assets/<chain>` package with its reviewed manifest and named references, and
connecting an internal source adapter to the shared manifest generator. Lookup,
ordering, immutability, strict parsing, and fail-closed evolution rules are
shared and must not be copied into the new chain package. See
[Adding an Asset Chain](docs/guides/adding-an-asset-chain.md) for the complete
source, generation, validation, integration, and automation workflow.

## Development

Run unit tests and whitespace checks:

```bash
make check
```

Compile integration tests without connecting to an RPC endpoint:

```bash
go test -count=1 -run '^$' -tags=integration ./integration/...
```

Run the Base fork suite with Foundry's `anvil` installed:

```bash
BASE_RPC_URL=<base-mainnet-rpc-url> make anvil-base
```

In a second terminal:

```bash
BASE_RPC_URL=http://127.0.0.1:8545 make test-integration
```

The Anvil target selects a hardfork that supports EIP-7702 set-code
transactions. Integration tests execute public Flow and strategy APIs against
real Base Aave, ERC20, and WETH state.

Contributor references:

- [Adding an Asset Chain](docs/guides/adding-an-asset-chain.md)
- [Migrating from `config.Coin` to Resolved Assets](docs/migrations/v0-coin-to-resolved-assets.md)
- [Phase 1 MVP Spec and Glossary](docs/specs/2026-07-07-phase-1-mvp-spec-and-glossary.md)
- [Static Flow Builder API](docs/specs/2026-07-08-static-flow-builder-api.md)

Changes should preserve the
`Flow -> ExecutionPlan -> Executor -> Validator` boundaries. Protocol-specific
calldata and event semantics belong in their protocol package; generic
execution code must remain protocol-neutral.

## Current Limitations

- Base is the only configured network.
- Aave V3, ERC20, and WETH are the only public protocol step packages.
- Dynamic amount sources are ERC20 balance based; arbitrary call return values
  cannot feed later calls.
- There is no built-in swap routing, slippage guard, health-factor guard, or
  leverage-loop strategy.
- Strategy builders do not discover positions or simulate protocol state.
- Runner does not yet perform a separate pre-submission simulation pass.

## Security

This project is experimental and has not been independently audited as a
complete SDK and execution system. EIP-7702 delegation grants code execution in
the EOA's context, so use a dedicated operation account with limited funds,
verify the configured implementation, test against a fork, and keep a tested
clear or redelegation path.

Protocol and asset changes made by a reverted batch are atomic, but gas and
nonces are consumed and a newly processed delegation may remain installed.
Start with small values and independently verify every transaction before using
real funds.

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE).
