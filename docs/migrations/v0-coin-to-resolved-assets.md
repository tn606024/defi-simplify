# Migrating from `config.Coin` to Resolved Assets

IAN-63 changes the public ERC20 and Aave `FlowStep` APIs from SDK-maintained
`config.Coin` values to address-first runtime values. This is a breaking v0 API
change.

The new execution path is:

```text
reviewed market manifest
  -> Aave Registry snapshot
  -> token.Token / aave.Reserve
  -> FlowStep
  -> ExecutionPlan
```

This removes the manual coin map as the source of truth for transaction
targets, decimals, Aave reserve membership, and reserve-token roles.

## Resolve Assets Once

Load a block-pinned snapshot and resolve the reserves needed by the flow:

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
weth, err := snapshot.Reserve(base.WETH)
if err != nil {
	return err
}
```

`base.USDC` and `base.WETH` are reviewed chain-and-address references. The
registry resolves live token metadata and Aave roles at one block. For an
uncatalogued asset, construct a `token.Ref` with `token.NewRef` and resolve it
through the same snapshot.

## FlowStep Changes

ERC20 steps now accept `token.Token`:

```go
erc20.Approve(usdc.Underlying(), aave.PoolSpender(market), amount)
erc20.Transfer(usdc.Underlying(), recipient, amount)
```

Aave steps now accept `aave.Reserve`:

```go
aave.ApproveSupply(usdc, supplyAmount)
aave.Supply(usdc, supplyAmount)
aave.Borrow(weth, borrowAmount)
```

Spender helpers now require the resolved market:

```go
aave.PoolSpender(market)
aave.GatewaySpender(market)
```

This makes market selection explicit and prevents a step from silently taking
its Pool or gateway address from an unrelated global chain map.

## Permit and Delegation Changes

Signature-based steps require an explicit capability with a reviewed EIP-712
domain version:

```go
permit, err := erc20.NewPermitCapability(usdc.Underlying(), "2")
if err != nil {
	return err
}

delegation, err := aave.NewDelegationCapability(weth, "1")
if err != nil {
	return err
}
```

Pass the capability to `erc20.Permit`, `aave.SupplyWithPermit`,
`aave.RepayWithPermit`, `aave.WithdrawETHWithPermit`, or
`aave.DelegationWithSig`. The SDK deliberately does not infer signature support
or domain versions from token symbols, names, or `nonces()` calls.

## Strategy Changes

Strategies derive their chain and Pool from their reserves. Remove `Chain` and
replace coin fields with reserve fields:

```go
flow, err := strategy.AaveSupplyBorrow(strategy.AaveSupplyBorrowParams{
	Account:       user,
	SupplyReserve: usdc,
	SupplyAmount:  supplyAmount,
	BorrowReserve: weth,
	BorrowAmount:  borrowAmount,
})
```

All reserves passed to one strategy must belong to the same resolved Aave
market.

## Legacy Clients

The `config.Coin`-based convenience clients in `client/contract` remain
available during migration but are deprecated. Address-based low-level
`Action` builders remain supported implementation primitives. New application
code should use registry-resolved FlowSteps or strategies with `defi.Runner`.
