# DeFi Simplify

Many DeFi workflows require several calls in order. Supplying collateral and
borrowing from Aave, for example, first requires a token approval and then the
two Pool operations. These calls are most useful when they succeed or revert as
one atomic transaction.

An external Multicall contract can batch those calls, but it becomes
`msg.sender` at every downstream protocol. Aave and ERC20 contracts then observe
the Multicall contract instead of the user's EOA, which changes allowance,
position ownership, receiver, and callback semantics.

DeFi Simplify uses EIP-7702 to give the user's EOA a reusable batch executor.
Calls made through the delegated account still originate from the EOA, so the
user remains the protocol-visible caller and continues to own the resulting
assets and positions.

The on-chain executor is the
[`DefiSimplify7702Account`](https://github.com/tn606024/defi-simplify-contracts/blob/v1.1.0/src/DefiSimplify7702Account.sol)
implementation maintained in the companion
[`defi-simplify-contracts`](https://github.com/tn606024/defi-simplify-contracts)
repository. This repository provides the Go SDK that builds, submits, and
validates Flows executed by that account.

Developers compose each workflow in Go instead of deploying a new Solidity
strategy contract for every combination. The SDK resolves protocol data,
builds an execution plan from reusable steps, submits it as one atomic
transaction, and validates the mined protocol events. This avoids the
deployment and upgrade overhead of per-strategy contracts while keeping atomic
execution in a shared on-chain account implementation.

```text
Go Flow: Approve -> Supply -> Borrow
                    |
                    v
           EIP-7702 delegated EOA
                    |
                    v
        Aave sees caller == user EOA
```

EIP-7702 provides delegation, not batching by itself. The delegated account
provides `executeBatch` for calls known before submission and
`executeBatchDynamic` when later calls depend on balances produced earlier in
the same transaction.

## Current Support

| Area | Support |
| --- | --- |
| Chain | Base (chain ID `8453`) |
| Account execution | EIP-7702 delegation to the reviewed `DefiSimplify7702Account` deployment |
| Static plans | Atomic inherited `executeBatch` execution |
| Dynamic plans | Atomic `executeBatchDynamic` execution with balance checkpoints and calldata patches |
| Aave V3 | Supply, borrow, repay, repay all, withdraw, withdraw all |
| ERC20 | Approve, transfer, transfer from |
| WETH | Wrap and unwrap |
| Strategies | Aave supply/borrow, close position, and native ETH compositions |
| Results | Mined receipt, per-step validation status, typed protocol events, and wrapped execution errors |

Normal Flow execution always uses the configured delegated account. External
Multicall and direct-EOA execution are not alternative Flow modes.

## Installation

DeFi Simplify requires Go 1.23.5 or later.

```bash
go get github.com/tn606024/defi-simplify@v0.4.0
```

## Quick Start

For a phase-by-phase executable lifecycle with inspect, dry-run, Base-fork, and
broadcast commands, see the [Base Aave lifecycle example](examples/base-aave-lifecycle/README.md).

A completed execution of that lifecycle is recorded in the
[Base mainnet execution evidence](docs/evidence/2026-08-03-base-aave-lifecycle.md),
including transaction hashes, typed events, position ownership, cleanup, and
delegation clearing.

The main SDK operation is composing and executing a Flow. The required client,
delegation, and market-resolution setup is included below in expandable
sections so the Flow remains the focus.

Before running it, provide:

- a Base RPC endpoint;
- a dedicated EOA private key with ETH for gas;
- enough USDC in that EOA for the selected supply amount; and
- supply and borrow amounts that satisfy current Aave risk parameters.

Never hard-code or commit a private key. Validate this lifecycle against an
Anvil Base mainnet fork before broadcasting it to Base.

<details>
<summary>Set up the Base client, signer, and EIP-7702 delegation</summary>

Connect to Base and create a signer for the EOA that will own the assets and
DeFi positions:

```go
ethClient, err := ethclient.DialContext(ctx, os.Getenv("BASE_RPC_URL"))
if err != nil {
	return err
}

keyHex := strings.TrimPrefix(os.Getenv("PRIVATE_KEY"), "0x")
eoaKey, err := crypto.HexToECDSA(keyHex)
if err != nil {
	return err
}

chainID, err := config.Base.ChainID()
if err != nil {
	return err
}
txOpts, err := bind.NewKeyedTransactorWithChainID(
	eoaKey,
	big.NewInt(int64(chainID)),
)
if err != nil {
	return err
}
txOpts.Context = ctx
```

Load the checked-in Base deployment and verify the implementation's runtime
code before enabling signing or submission:

```go
implementationDeployment, err := defisimplify7702.ResolveAndVerifyAccountDeployment(
	ctx,
	ethClient,
	config.Base,
)
if err != nil {
	return err
}

delegationManager, err := eip7702.NewManager(
	ethClient,
	txOpts,
	eoaKey,
	big.NewInt(int64(chainID)),
)
if err != nil {
	return err
}
if err := delegationManager.AssertClean(ctx, txOpts.From); err != nil {
	return err
}

delegateTx, err := delegationManager.Delegate(ctx, implementationDeployment.Address)
if err != nil {
	return err
}
delegateReceipt, err := bind.WaitMined(ctx, ethClient, delegateTx)
if err != nil {
	return err
}
if delegateReceipt.Status != types.ReceiptStatusSuccessful {
	return errors.New("delegation transaction reverted")
}
if err := delegationManager.AssertDelegatedTo(
	ctx,
	txOpts.From,
	implementationDeployment.Address,
); err != nil {
	return err
}
```

Delegation is persistent. It remains installed until it is switched or cleared,
and a later transaction revert does not roll back an authorization that was
already processed.

</details>

<details>
<summary>Resolve the Base Aave market and reserves</summary>

The SDK selects the reviewed Base Aave market and resolves reserve relationships
from one immutable snapshot:

```go
aaveSnapshot, err := aave.LoadBaseV3Snapshot(ctx, ethClient)
if err != nil {
	return err
}
usdcReserve, err := aaveSnapshot.Reserve(base.USDC)
if err != nil {
	return err
}
wethReserve, err := aaveSnapshot.Reserve(base.WETH)
if err != nil {
	return err
}
```

`base.USDC` identifies the token by chain and address. `Reserve` returns the
resolved Aave reserve, which also contains its aToken and debt-token
relationships. `usdcReserve.Underlying()` selects the ERC20 token that the EOA
actually approves and supplies.

</details>

### Compose And Execute The Flow

`amount.Exact(value)` means the amount is fixed before the transaction is
submitted. The value uses human-readable token units, such as `100` USDC rather
than `100000000`; the protocol step converts it using the resolved token
decimals.

`aave.PoolSpender(market)` represents the selected market's reviewed Pool
address as an ERC20 spender. It does not execute a call. It supplies the typed
spender required by `erc20.Approve`, while the approval itself remains an ERC20
operation.

Define those values once, then compose the operations in execution order:

```go
supplyAmount := amount.Exact(decimal.NewFromInt(100))
borrowAmount := amount.Exact(decimal.RequireFromString("0.01"))
poolSpender := aave.PoolSpender(aaveSnapshot.Market())

flow := defi.NewFlow(txOpts.From, defi.WithChain(config.Base)).
	Add(erc20.Approve(
		usdcReserve.Underlying(),
		poolSpender,
		supplyAmount,
	)).
	Add(aave.Supply(usdcReserve, supplyAmount)).
	Add(aave.Borrow(wethReserve, borrowAmount))

executionResult, err := defi.NewRunner(ethClient, txOpts, config.Base).Execute(ctx, flow)
if err != nil {
	if executionResult != nil {
		fmt.Println("mined transaction:", executionResult.Receipt.TxHash)
	}
	return err
}

fmt.Println("transaction:", executionResult.Receipt.TxHash)
supplies := defi.EventsOf[*aave.SupplyEvent](executionResult)
borrows := defi.EventsOf[*aave.BorrowEvent](executionResult)
```

Both amount sources are exact, so the Flow compiles to a static plan and
executes atomically through `executeBatch`. Aave observes `txOpts.From` as the
caller and position owner.

<details>
<summary>Clear the EIP-7702 delegation</summary>

Clear the delegation when the EOA no longer needs the account implementation.
Close or otherwise manage any open DeFi position separately; clearing delegated
code does not move or close positions owned by the EOA.

```go
clearTx, err := delegationManager.Clear(ctx)
if err != nil {
	return err
}
clearReceipt, err := bind.WaitMined(ctx, ethClient, clearTx)
if err != nil {
	return err
}
if clearReceipt.Status != types.ReceiptStatusSuccessful {
	return errors.New("delegation clear transaction reverted")
}
if err := delegationManager.AssertClean(ctx, txOpts.From); err != nil {
	return err
}
```

</details>

## Execution Model

```text
FlowStep[]
    |
    v
ExecutionPlan
    |
    v
Runner
    |
    +-- PlanStatic  --> delegated EOA self-call --> executeBatch
    |
    +-- PlanDynamic --> delegated EOA self-call --> executeBatchDynamic
    |
    v
protocol calls from the EOA
    |
    v
mined receipt --> typed event validation --> ExecutionResult
```

The caller does not select an execution mode. `Flow.Build` derives the plan kind
from step metadata, and `Runner.Execute` dispatches to the required account
entrypoint.

| Plan kind | Selected when | Account entrypoint |
| --- | --- | --- |
| `PlanStatic` | Every call and amount is known before submission | `executeBatch` |
| `PlanDynamic` | A call declares a balance checkpoint, runtime amount patch, or callback | `executeBatchDynamic` |

`ExecutionPlan.Account()` is the protocol-visible caller identity. FlowSteps
derive owner, sender, receiver, or `onBehalfOf` values from that account when
the protocol operation requires it.

## Runtime Amounts

Dynamic plans let a later call consume an ERC20 balance observed during account
execution. Runtime amount intent remains explicit:

- `amount.CurrentBalance(token)` uses the account's current token balance;
- `amount.CheckpointDelta(checkpoint)` uses only the increase after a named
  earlier checkpoint; and
- `amount.Scale(source, bps)` applies a basis-point ratio to a runtime source.

This Flow wraps an exact amount of ETH, records the WETH balance immediately
before the wrap, and unwraps only the WETH gained by that call:

```go
checkpoint := amount.Checkpoint("before-wrap", base.WETH)

flow := defi.NewFlow(txOpts.From, defi.WithChain(config.Base)).
	Add(defi.CheckpointBefore(
		weth.Wrap(base.WETH, amount.Exact(decimal.RequireFromString("0.1"))),
		checkpoint,
	)).
	Add(weth.Unwrap(base.WETH, amount.CheckpointDelta(checkpoint)))
```

The checkpoint protects pre-existing WETH from being swept by the later step.
The Flow compiles to `PlanDynamic`; protocol packages own the ABI locations
patched by the account, so high-level callers do not provide raw offsets.

The SDK does not infer dependencies from call order, pipe arbitrary return data
between calls, or patch native `value` from a runtime source.

## Delegation Lifecycle

EIP-7702 delegation is account state, not a one-transaction permission:

- installation, switching, and clearing are explicit set-code transactions;
- successful and reverted Flow transactions leave delegation installed;
- switching the delegated implementation does not move EOA balances or DeFi
  positions; and
- applications must serialize delegation changes and Flow submission for each
  EOA.

Before each Flow submission, the Runner checks the EOA's pending delegation
target against the configured implementation. This detects an already switched
or cleared delegation, but it cannot remove the race between preflight and
transaction inclusion.

## Reviewed Account Deployment

The SDK currently selects this checked-in deployment identity:

| Network | Release | Delegation implementation |
| --- | --- | --- |
| Base | [`v1.1.0`](https://github.com/tn606024/defi-simplify-contracts/tree/v1.1.0) | [`0x9B1854c65Ce4656349d04e612260dFCEaf5B1d69`](https://basescan.org/address/0x9B1854c65Ce4656349d04e612260dFCEaf5B1d69#code) |

The ABI, runtime code hash, deployment transaction, and reproducible metadata
are recorded in the
[`v1.1.0` Base deployment manifest](https://github.com/tn606024/defi-simplify-contracts/blob/v1.1.0/deployments/base-v1.1.json).
The address is an immutable implementation to which an EOA delegates. It is not
the user's account or a custody address, and users must not send funds to it.

## Results And Errors

Every protocol FlowStep builds its call and typed event expectations from the
same resolved account, target, asset, and amount data. After a successful
transaction, `Runner.Execute` validates those expectations against the mined
receipt in step order.

```go
executionResult, err := runner.Execute(ctx, flow)
if err != nil && executionResult != nil {
	fmt.Println("mined transaction:", executionResult.Receipt.TxHash)
}
```

The result contract is:

| Outcome | Result | Error |
| --- | --- | --- |
| Build, signing, or pre-submission failure | `nil` | non-nil |
| Mined transaction revert | Contains receipt | non-nil |
| Successful receipt with semantic validation failure | Contains receipt and partial step results | non-nil |
| Successful receipt with successful validation | Contains receipt and step results | `nil` |

Use `errors.Is` and `errors.As` with exported sentinel and typed errors. Dynamic
account reverts can include the failed call index through
`defisimplify7702.ContractError`.

Receipt validation confirms what the mined transaction emitted. It is not
pre-submission simulation, a background event listener, or an indexing system.

## Security And Trust Boundaries

This project is experimental and has not been independently audited. It can
submit transactions that move assets or create debt. Use isolated accounts,
review every target and amount, and validate flows on a local fork before using
real assets.

- Verify the configured account implementation's runtime code once during
  application initialization, before enabling signing or submission.
- The Runner verifies the EOA's pending delegation target before every Flow
  submission.
- The signer remains responsible for the complete plan: targets, calldata,
  amounts, approvals, protocol-native slippage or price limits, and economic
  post-conditions.
- Private-key custody, transaction simulation, target admission policy,
  indexing, and monitoring are application responsibilities.
- A successful simulation cannot guarantee execution against later chain state.
  Use protocol-native bounds and atomic on-chain assertions when available.
- Low-level `ActionStep` usage may not provide typed event expectations. Prefer
  public protocol FlowSteps for supported operations.
- A mined receipt with failed semantic validation still represents an on-chain
  transaction. The SDK preserves its receipt and transaction hash for diagnosis.

## Strategy Builders

The `strategy` package contains thin compositions of public FlowSteps:

- `AaveSupplyBorrow`
- `AaveClosePosition`
- `AaveSupplyNativeETH`
- `AaveBorrowNativeETH`
- `AaveWithdrawNativeETH`
- `AaveWithdrawAllNativeETH`

Strategy builders return a normal `*defi.Flow`. They validate static inputs but
do not read chain state, sign, submit, simulate, or select execution behavior.

## Development

Run unit tests:

```bash
go test -count=1 ./...
```

Compile integration tests:

```bash
go test -count=1 -run '^$' -tags=integration ./integration/...
```

Run the Base mainnet fork suite:

```bash
BASE_RPC_URL=https://mainnet.base.org make anvil-base
```

In another terminal:

```bash
BASE_RPC_URL=http://127.0.0.1:8545 make test-integration
```

Refresh imported account artifacts from a local `defi-simplify-contracts`
checkout:

```bash
make update-contract-artifacts
```

Do not hand-edit generated contract bindings or asset catalog references.

## Package Layout

| Package | Responsibility |
| --- | --- |
| root `defi` package | Flow, ExecutionPlan, Runner, validator, and neutral execution types |
| `amount` | Exact and runtime amount intent |
| `token`, `assets`, `assets/base` | Chain-scoped token identity and reviewed asset catalog |
| `aave`, `erc20`, `weth` | Protocol-specific FlowSteps, calldata, and event semantics |
| `strategy` | Reusable compositions of public FlowSteps |
| `client/account/eip7702` | Delegation authorization, state inspection, switching, and clear |
| `client/account/defisimplify7702` | Deployment manifests, ABI translation, and account execution |
| `integration` | Ginkgo Base-fork behavior tests |

See [Adding an asset chain](docs/guides/adding-an-asset-chain.md) for catalog
extension rules.
