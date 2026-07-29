# Migrating from Aave Gateway and Credit Delegation

This v0 change removes Aave WrappedTokenGateway and debt-token credit-delegation
APIs from the SDK. Aave operations now use direct Pool calls whose downstream
caller is the EIP-7702 EOA.

## Removed APIs

The following public FlowSteps were removed:

- `aave.DepositETH`
- `aave.BorrowETH`
- `aave.WithdrawETH`
- `aave.WithdrawETHWithPermit`
- `aave.ApproveDelegation`
- `aave.DelegationWithSig`

`aave.GatewaySpender`, `aave.DelegationCapability`, the corresponding typed
events, and their low-level `client/contract` actions and methods were also
removed. The Aave market model no longer exposes `WrappedTokenGateway`, and the
deployment manifest schema is now version 2.

## Native ETH

Use the WETH-composed strategy builders for the common native ETH flows:

| Removed flow | Replacement |
| --- | --- |
| `aave.DepositETH` | `strategy.AaveSupplyNativeETH` |
| `aave.BorrowETH` | `strategy.AaveBorrowNativeETH` |
| `aave.WithdrawETH` | `strategy.AaveWithdrawNativeETH` |
| Full native withdrawal | `strategy.AaveWithdrawAllNativeETH` |

The equivalent manual compositions are:

```text
supply native ETH
  WETH.Wrap -> ERC20.Approve(Aave Pool) -> Aave.Supply

borrow native ETH
  checkpoint WETH -> Aave.Borrow(WETH) -> WETH.Unwrap(checkpoint delta)

withdraw native ETH
  checkpoint WETH -> Aave.Withdraw(WETH) -> WETH.Unwrap(checkpoint delta)
```

Native supply is a static plan and uses `ExecutionAtomicEOA`. Borrow and
withdraw flows depend on runtime WETH deltas and use `ExecutionDynamicEOA`.
The checkpoint prevents pre-existing WETH from being unwrapped.

There is no direct replacement for `WithdrawETHWithPermit`: the EIP-7702 EOA
owns the Aave position and calls the Pool directly, so the gateway-specific
permit path is unnecessary.

## Credit Delegation

The removed delegation APIs authorized another address to borrow against the
EOA's Aave position. They were required by the old external-Multicall design,
where the Multicall contract became `msg.sender`.

The supported EIP-7702 executors preserve the EOA as the downstream Pool caller.
Normal self-borrow flows therefore call `aave.Borrow` directly and require no
debt-token credit delegation.

Applications that intentionally need third-party borrowing must integrate
Aave's debt-token delegation outside the core SDK.

## EIP-7702 Delegation Is Unchanged

This removal does not affect `client/account/eip7702`. EIP-7702 delegation
installs account code on the EOA and remains active until it is changed or
cleared. Continue to use the lifecycle manager to inspect, install, change, or
revoke that delegation.

## Deployment Manifest

Deployment manifest schema version 1 included the WrappedTokenGateway address.
Schema version 2 contains only the Pool, PoolAddressesProvider, and
AaveProtocolDataProvider trust anchors. Regenerate custom manifests with the
current tooling; schema version 1 is rejected rather than upgraded implicitly.
