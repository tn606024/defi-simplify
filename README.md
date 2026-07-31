# DeFi Simplify

DeFi Simplify is a Go SDK for composing and executing EOA-native DeFi flows
through EIP-7702.

The SDK builds protocol-aware `FlowStep` values, compiles them into an
`ExecutionPlan`, executes the plan atomically through a delegated
`DefiSimplify7702Account`, and validates the mined receipt against typed event
expectations.

> [!WARNING]
> The SDK and deployed account contracts are experimental and unaudited. They
> can submit transactions that move funds or create debt. Use isolated accounts
> and review every generated flow before using real assets.

## What It Provides

| Area | Current support |
| --- | --- |
| Chain | Base |
| Account execution | EIP-7702 delegation to the configured `DefiSimplify7702Account` |
| Static plans | Atomic inherited `executeBatch` execution |
| Dynamic plans | Atomic `executeBatchDynamic` execution with balance checkpoints and calldata patches |
| Aave V3 | Supply, borrow, repay, repay all, withdraw, withdraw all |
| ERC20 | Approve, transfer, transfer from |
| WETH | Wrap and unwrap |
| Strategies | Aave supply/borrow, close position, and native ETH compositions |
| Results | Mined receipt, per-step validation status, typed protocol events, and wrapped execution errors |

The SDK does not expose external Multicall or direct-EOA Flow execution as
alternative modes. Every normal Flow executes through the configured delegated
account so downstream protocols observe the Flow account as the caller.

## Account Contract

The EIP-7702 account implementation is developed and released from the
[Defi Simplify Contracts](https://github.com/tn606024/defi-simplify-contracts)
repository. The SDK currently selects the following reviewed deployment
identity:

| Network | Contracts release | Delegation implementation |
| --- | --- | --- |
| Base (chain ID `8453`) | [`v1.1.0`](https://github.com/tn606024/defi-simplify-contracts/tree/v1.1.0) | [`DefiSimplify7702Account` at `0x9B1854c65Ce4656349d04e612260dFCEaf5B1d69`](https://basescan.org/address/0x9B1854c65Ce4656349d04e612260dFCEaf5B1d69#code) |

The complete ABI, runtime code hash, deployment transaction, and reproducible
deployment metadata are recorded in the
[`v1.1.0` Base deployment manifest](https://github.com/tn606024/defi-simplify-contracts/blob/v1.1.0/deployments/base-v1.1.json).
This address is the immutable implementation that an EOA delegates to. It is
not the user's account or a custody address, and users must not send funds to
it.

## Execution Model

```text
FlowStep[]
    |
    v
ExecutionPlan
    |
    +-- PlanStatic  --> DefiSimplify7702Account.executeBatch
    |
    +-- PlanDynamic --> DefiSimplify7702Account.executeBatchDynamic
    |
    v
mined receipt
    |
    v
typed event validation
```

The caller does not choose an execution mode. `Runner.Execute` builds the Flow
and dispatches from `ExecutionPlan.Kind()`.

```go
result, err := defi.NewRunner(client, opts, config.Base).Execute(ctx, flow)
```

Exact-only steps produce a static plan. Runtime amount sources, checkpoints,
or calldata patches produce a dynamic plan.

## Delegation Setup

EIP-7702 delegation is persistent. It is not limited to one Flow, and a reverted
Flow does not clear it. Applications must treat delegation setup, switching,
and clearing as explicit account lifecycle operations.

Load the checked-in Base deployment and verify its runtime code once during
application initialization. The Base manifest currently selects the official
Defi Simplify Contracts `v1.1.0` release:

```go
deployment, err := defisimplify7702.DeploymentForChain(config.Base)
if err != nil {
	return err
}
accountDeployment, err := deployment.Contract(defisimplify7702.AccountContract)
if err != nil {
	return err
}

code, err := client.CodeAt(ctx, accountDeployment.Address, nil)
if err != nil {
	return err
}
if err := accountDeployment.VerifyRuntimeCode(code); err != nil {
	return err
}
```

Delegate the EOA to that implementation:

```go
chainID, err := config.Base.ChainID()
if err != nil {
	return err
}
manager, err := eip7702.NewManager(
	client,
	opts,
	authorizationKey,
	big.NewInt(int64(chainID)),
)
if err != nil {
	return err
}

tx, err := manager.Delegate(ctx, accountDeployment.Address)
if err != nil {
	return err
}
receipt, err := bind.WaitMined(ctx, client, tx)
if err != nil {
	return err
}
if receipt.Status != types.ReceiptStatusSuccessful {
	return errors.New("delegation transaction reverted")
}
```

The Runner checks the pending delegation target before each submission. This
protects against submitting through an unexpected implementation, but it cannot
remove the lifecycle race between preflight and transaction inclusion.
Applications must serialize delegation changes and Flow execution for each EOA.

Applications upgrading from the earlier Base deployment must switch each EOA's
persistent delegation explicitly by submitting a new
`manager.Delegate(ctx, accountDeployment.Address)` transaction and waiting for
its successful receipt. Existing balances and DeFi positions remain owned by
the EOA; switching delegated code does not move them to the implementation
contract. A reverted transaction does not restore the previous delegation.

Clear the delegation explicitly when it is no longer needed:

```go
tx, err := manager.Clear(ctx)
if err != nil {
	return err
}
receipt, err := bind.WaitMined(ctx, client, tx)
```

## Build An Aave Flow

Resolve the current Base Aave market and reserve metadata:

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
if err != nil {
	return err
}
wethReserve, err := snapshot.Reserve(base.WETH)
if err != nil {
	return err
}
```

Compose protocol steps and execute them atomically:

```go
supplyAmount := decimal.NewFromInt(100)
borrowAmount := decimal.RequireFromString("0.01")

flow := defi.NewFlow(opts.From, defi.WithChain(config.Base)).
	Add(erc20.Approve(
		usdc.Underlying(),
		aave.PoolSpender(market),
		amount.Exact(supplyAmount),
	)).
	Add(aave.Supply(usdc, amount.Exact(supplyAmount))).
	Add(aave.Borrow(wethReserve, amount.Exact(borrowAmount)))

result, err := defi.NewRunner(client, opts, config.Base).Execute(ctx, flow)
if err != nil {
	return err
}
fmt.Println(result.Receipt.TxHash)
```

Allowance operations use the protocol-neutral `erc20.Approve` step.
`aave.PoolSpender` supplies the reviewed market-specific Pool address without
wrapping or renaming the ERC20 operation.

For code written against the earlier v0 API, replace:

```go
aave.ApproveSupply(reserve, amount.Exact(value))
```

with:

```go
erc20.Approve(
	reserve.Underlying(),
	aave.PoolSpender(market),
	amount.Exact(value),
)
```

All three amounts are known before submission, so this Flow compiles to
`PlanStatic` and executes through inherited `executeBatch`.

## Runtime Amounts

Dynamic plans let a later call consume a token balance observed during account
execution.

`amount.CurrentBalance` reads the delegated EOA's current token balance:

```go
flow := defi.NewFlow(opts.From, defi.WithChain(config.Base)).
	Add(erc20.Approve(
		usdc.Underlying(),
		aave.PoolSpender(market),
		amount.CurrentBalance(usdc.Underlying().Ref()),
	))
```

`amount.CheckpointDelta` uses only the balance gained after an explicit earlier
checkpoint:

```go
checkpoint := amount.Checkpoint("before-wrap", base.WETH)

flow := defi.NewFlow(opts.From, defi.WithChain(config.Base)).
	Add(defi.CheckpointBefore(
		weth.Wrap(base.WETH, amount.Exact(decimal.RequireFromString("0.1"))),
		checkpoint,
	)).
	Add(weth.Unwrap(base.WETH, amount.CheckpointDelta(checkpoint)))
```

This protects pre-existing inventory from being swept by a later step.
`amount.Scale(source, bps)` can apply a basis-point ratio to a runtime source.

Runtime dependencies must be explicit. The SDK does not infer dependencies from
call order, pipe arbitrary return data between calls, or patch native `value`.

## Results And Errors

`Runner.Execute` always performs semantic receipt validation after a successful
transaction:

```go
supplies := defi.EventsOf[*aave.SupplyEvent](result)
borrows := defi.EventsOf[*aave.BorrowEvent](result)
```

Failures before submission return a nil result. Once a transaction is mined,
transaction reverts and semantic validation failures return both a result and
an error so the transaction hash and partial validation state remain available.

```go
result, err := runner.Execute(ctx, flow)
if err != nil && result != nil {
	fmt.Println("mined transaction:", result.Receipt.TxHash)
}
```

Use `errors.Is` and `errors.As` with exported sentinel and typed errors. Dynamic
account reverts can include the failed call index through
`defisimplify7702.ContractError`.

## Strategy Builders

The `strategy` package contains thin compositions of public FlowSteps:

- `AaveSupplyBorrow`
- `AaveClosePosition`
- `AaveSupplyNativeETH`
- `AaveBorrowNativeETH`
- `AaveWithdrawNativeETH`
- `AaveWithdrawAllNativeETH`

Strategy builders return a normal `*defi.Flow`. They do not read chain state,
sign, submit, or select execution behavior.

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

Refresh imported contract artifacts from a local
`defi-simplify-contracts` checkout:

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
extension rules. Breaking v0 API changes are documented under
[`docs/migrations`](docs/migrations).
