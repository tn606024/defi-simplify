# Migrating from Permit FlowSteps

DS-62 removes token-specific permit signing and permit-only FlowSteps from the
core SDK. This is a breaking v0 API change.

The SDK is centered on EIP-7702 EOA-native execution. An EOA can atomically
approve a protocol and perform the consuming operation in the same batch, so
the core execution path no longer needs a second token-specific signature
system.

## Removed APIs

The following public surfaces were removed:

- `erc20.Permit`
- `erc20.PermitCapability` and `erc20.NewPermitCapability`
- `aave.SupplyWithPermit`
- `aave.RepayWithPermit`
- low-level permit and nonce Actions, builders, calldata helpers, and client
  methods
- token permit-support and EIP-712 domain metadata in `config.Coin`
- the permit-only ERC20 binding and generic permit message signer

The complete generated Aave Pool binding still contains the protocol's
`supplyWithPermit` and `repayWithPermit` ABI methods because the same generated
binding owns all supported Pool operations. They are not exposed as core SDK
Actions or FlowSteps.

## Supply Migration

Replace `aave.SupplyWithPermit` with an approval and supply in the same Flow:

```go
flow := defi.NewFlow(account, defi.WithChain(market.Chain())).
	Add(aave.ApproveSupply(reserve, amount.Exact(value))).
	Add(aave.Supply(reserve, amount.Exact(value)))
```

`ApproveSupply` is the readable Aave-oriented builder for approving the
resolved Pool. The equivalent explicit ERC20 step is:

```go
erc20.Approve(
	reserve.Underlying(),
	aave.PoolSpender(market),
	amount.Exact(value),
)
```

Execute the static plan with `ExecutionAtomicEOA` so approval and supply either
both succeed or both revert.

## Repay Migration

Replace `aave.RepayWithPermit` with an approval and repay:

```go
flow := defi.NewFlow(account, defi.WithChain(market.Chain())).
	Add(erc20.Approve(
		reserve.Underlying(),
		aave.PoolSpender(market),
		amount.Exact(value),
	)).
	Add(aave.Repay(reserve, amount.Exact(value)))
```

For an upper-bound allowance, add a final zero approval in the same batch to
clear any remainder:

```text
Approve(temporary allowance) -> Repay or RepayAll -> Approve(0)
```

The built-in Aave close-position strategy already follows this pattern.

## Constructor Migration

The legacy protocol clients no longer own a permit signer:

```go
// Before
contract.NewBaseClient(conn, chain, opts, signer)
contract.NewDefiClient(opts, conn, signer, chain)

// After
contract.NewBaseClient(conn, chain, opts)
contract.NewDefiClient(opts, conn, chain)
```

Applications that used `helper.MsgSigner` only for token permits can remove it.

## Signatures That Remain

This change does not remove EIP-7702 authorization or account signatures:

- `client/account/eip7702` still owns delegation authorization, state changes,
  and explicit clear/change lifecycle operations.
- EIP-7702 delegation remains persistent until it is changed or cleared.
- account-level EIP-1271 and ERC-4337 signature behavior remains part of the
  selected delegated account implementation.

Token permits and account authorization solve different problems. Relayer-
oriented token permit support can be added later as a separate extension
without making protocol FlowSteps or the core execution pipeline own token
domain versions and signing policy.
