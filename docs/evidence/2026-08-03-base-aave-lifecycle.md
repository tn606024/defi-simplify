# Base Mainnet Aave Lifecycle Execution Evidence

On August 3, 2026, the public DeFi Simplify SDK executed a complete EIP-7702
Aave V3 lifecycle on Base mainnet. A dedicated EOA installed the reviewed
account implementation, opened an Aave position through one static Flow,
closed it through one dynamic Flow, and cleared its delegation.

This record links the fixed source revision to public transactions and
independently re-read chain state. It is execution evidence, not a security
audit, production recommendation, or guarantee of future behavior.

## Source And Configuration

| Item | Value |
| --- | --- |
| Execution date | August 3, 2026 |
| Network | Base mainnet |
| Chain ID | `8453` |
| SDK release baseline | [`v0.4.0`](https://github.com/tn606024/defi-simplify/releases/tag/v0.4.0) |
| Executed SDK revision | [`b15ea9acd023498182e7226992b5743a934f7a3b`](https://github.com/tn606024/defi-simplify/commit/b15ea9acd023498182e7226992b5743a934f7a3b) (`v0.4.0-21-gb15ea9a`) |
| Example source | [`examples/base-aave-lifecycle`](https://github.com/tn606024/defi-simplify/tree/b15ea9acd023498182e7226992b5743a934f7a3b/examples/base-aave-lifecycle) |
| Dedicated EOA | [`0xb5FD9f60Fc6ca7662Ff22D09ec7832CD221fbcdD`](https://base.blockscout.com/address/0xb5FD9f60Fc6ca7662Ff22D09ec7832CD221fbcdD) |
| Account contract release | [`defi-simplify-contracts v1.1.0`](https://github.com/tn606024/defi-simplify-contracts/tree/v1.1.0) |
| Delegated implementation | [`0x9B1854c65Ce4656349d04e612260dFCEaf5B1d69`](https://base.blockscout.com/address/0x9B1854c65Ce4656349d04e612260dFCEaf5B1d69) |
| Aave V3 Pool | [`0xA238Dd80C259a72e81d7e4664a9801593F98d1c5`](https://base.blockscout.com/address/0xA238Dd80C259a72e81d7e4664a9801593F98d1c5) |
| WETH | [`0x4200000000000000000000000000000000000006`](https://base.blockscout.com/token/0x4200000000000000000000000000000000000006) |
| USDC | [`0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913`](https://base.blockscout.com/token/0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913) |

The command loaded the reviewed Base deployment and Aave market definitions,
verified chain ID `8453`, and verified the implementation runtime code before
allowing a transaction to be signed or submitted. No private key, RPC
credential, seed phrase, or signed raw transaction is included in this record.

## Transaction Summary

| Phase | Transaction | Block | Status | Gas used |
| --- | --- | ---: | --- | ---: |
| Install delegation | [`0x8bdac720...d506fc9`](https://base.blockscout.com/tx/0x8bdac72068a505e1b6ea1995bf3abc73aafd2a925984cea8488c5f780d506fc9) | `49472034` | Success | `36,837` |
| Open Aave position | [`0x7d33273f...8bfbde5`](https://base.blockscout.com/tx/0x7d33273fa3c5936a976b378a6ab2c63d48c86e267253fbe055b393ad78bfbde5) | `49472273` | Success | `408,261` |
| Close Aave position | [`0x43b37dff...900ed7`](https://base.blockscout.com/tx/0x43b37dff405e3b599aad274bc77d8ced1807269b6ee22cee5fefa21597900ed7) | `49472688` | Success | `320,948` |
| Clear delegation | [`0xe9bee162...acff14`](https://base.blockscout.com/tx/0xe9bee162d8e66011adbaf5a0830f49698cbe0b76072471ad32d4c7dbfcacff14) | `49472905` | Success | `36,800` |

The install transaction authorized the reviewed v1.1.0 implementation. The
clear transaction authorized the zero address. After the clear transaction,
`eth_getCode` for the EOA at block `49472905` returned `0x`.

## Open Position

The example built the following public SDK Flow and submitted it through
`Runner.Execute(ctx, flow)` as a static plan using the account's inherited
`executeBatch` entrypoint:

| Order | Flow step | Operation |
| ---: | --- | --- |
| 1 | `weth.Wrap` | Wrap `0.001 ETH` into `0.001 WETH` |
| 2 | `erc20.Approve` | Approve the Aave Pool for `0.001 WETH` |
| 3 | `aave.Supply` | Supply `0.001 WETH` for the EOA |
| 4 | `aave.Borrow` | Borrow `0.5 USDC` for the EOA using variable-rate debt |

All four calls were included in transaction
[`0x7d33273f...8bfbde5`](https://base.blockscout.com/tx/0x7d33273fa3c5936a976b378a6ab2c63d48c86e267253fbe055b393ad78bfbde5).
The SDK decoded and validated these typed events from the mined receipt:

| Event | Core validated fields | Log index |
| --- | --- | ---: |
| WETH `Deposit` | `account=0xb5FD...bcdD`, `amount=1000000000000000` | `1758` |
| ERC20 `Approval` | `owner=0xb5FD...bcdD`, `spender=0xA238...d1c5`, `amount=1000000000000000` | `1759` |
| Aave `Supply` | `asset=WETH`, `user=0xb5FD...bcdD`, `onBehalfOf=0xb5FD...bcdD`, `amount=1000000000000000` | `1765` |
| Aave `Borrow` | `asset=USDC`, `user=0xb5FD...bcdD`, `onBehalfOf=0xb5FD...bcdD`, `amount=500000`, `interestRateMode=2` | `1770` |

At block `49472273`, direct token reads reported:

| Address | aWETH balance | Variable-debt USDC balance |
| --- | ---: | ---: |
| Dedicated EOA | `999999999999999` | `500002` |
| Delegated implementation | `0` | `0` |

The Aave events name the EOA as both the user and `onBehalfOf` account, and the
position tokens were held by the EOA rather than the implementation. This is
the protocol-visible caller and ownership behavior that an external Multicall
contract cannot preserve.

## Close Position

Before submission, the close Flow used a reviewed repayment bound of `0.51
USDC`. The SDK compiled the Flow as a dynamic plan because the unwrap amount
depended on the WETH balance gained by the earlier withdrawal in the same
transaction.

| Order | Flow step | Operation |
| ---: | --- | --- |
| 1 | `erc20.Approve` | Approve the Aave Pool for at most `0.51 USDC` |
| 2 | `aave.RepayAll` | Repay all EOA variable-rate USDC debt |
| 3 | `erc20.Approve` | Clear the remaining USDC Pool allowance to zero |
| 4 | `aave.WithdrawAll` | Withdraw all EOA WETH collateral and record the WETH balance delta |
| 5 | `weth.Unwrap` | Patch calldata with that delta and unwrap only the withdrawn WETH |

Transaction
[`0x43b37dff...900ed7`](https://base.blockscout.com/tx/0x43b37dff405e3b599aad274bc77d8ced1807269b6ee22cee5fefa21597900ed7)
executed the plan atomically through `executeBatchDynamic`. The SDK decoded
and validated:

| Event | Core validated fields | Log index |
| --- | --- | ---: |
| ERC20 `Approval` | `owner=0xb5FD...bcdD`, `spender=0xA238...d1c5`, `amount=510000` | `143` |
| Aave `Repay` | `asset=USDC`, `user=0xb5FD...bcdD`, `repayer=0xb5FD...bcdD`, `amount=500002` | `148` |
| ERC20 `Approval` | `owner=0xb5FD...bcdD`, `spender=0xA238...d1c5`, `amount=0` | `149` |
| Aave `Withdraw` | `asset=WETH`, `user=0xb5FD...bcdD`, `to=0xb5FD...bcdD`, `amount=1000000389254931` | `155` |
| WETH `Withdrawal` | `account=0xb5FD...bcdD`, `amount=1000000389254931` | `156` |

The Aave `Withdraw` amount and WETH `Withdrawal` amount are identical. This
confirms that the checkpoint delta was resolved during execution and patched
into the later unwrap call.

## Final State

After cleanup and the delegation-clear transaction, the example re-read all
targeted state at block `49472905`:

| State | Final value |
| --- | ---: |
| Aave WETH collateral | `0` |
| Aave WETH stable debt | `0` |
| Aave WETH variable debt | `0` |
| Aave USDC collateral | `0` |
| Aave USDC stable debt | `0` |
| Aave USDC variable debt | `0` |
| WETH allowance to Aave Pool | `0` |
| USDC allowance to Aave Pool | `0` |
| EOA WETH balance | `10000000000000` (`0.00001 WETH`) |
| EOA USDC balance | `999998` (`0.999998 USDC`) |
| EOA native balance | `24454527432571627` wei |
| EOA code | `0x` |
| EIP-7702 delegation | Cleared |

The final balances are reported only to make the cleanup state reproducible;
they are not performance or profitability results.

## Reproduction Boundary

The exact example source and configuration can be checked out at commit
`b15ea9acd023498182e7226992b5743a934f7a3b`. Follow the
[Base Aave lifecycle instructions](../../examples/base-aave-lifecycle/README.md)
to inspect and dry-run each phase, and execute it first against an
EIP-7702-capable Anvil Base fork. Reproduction requires a separately funded,
dedicated account and must use amounts compatible with the forked Aave market
state.

This successful run proves that the fixed SDK revision produced and validated
the documented transactions on Base at the listed blocks. It does not prove
the absence of defects, replace independent review, or establish that the same
flows are safe for other accounts, assets, market conditions, or future
contract and SDK revisions.
