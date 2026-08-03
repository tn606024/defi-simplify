# Base Aave Lifecycle

This example runs one reviewable phase of an EOA-native Aave lifecycle per
invocation. It uses the public SDK to inspect a dedicated EOA, install or clear
EIP-7702 delegation, build atomic open and close Flows, submit through
`Runner.Execute`, and print typed mined events.

The command never runs the complete lifecycle through one flag. Each
state-changing phase is a dry run unless `--broadcast` is also present.

## Environment

Create a local environment file from the committed template:

```bash
cp examples/base-aave-lifecycle/.env.example .env
chmod 600 .env
```

Set `BASE_RPC_URL` and `PRIVATE_KEY` in `.env`. The command loads that file from
the current directory before parsing its configuration. Existing process
environment values take precedence, which lets CI and one-off invocations
override the file without changing it. The committed `.env.example` contains
no secrets, while `.env` is ignored by Git.

The command derives the EOA from `PRIVATE_KEY`; neither value is printed. The
RPC must report Base chain ID `8453`. The EOA needs native ETH for gas and the
WETH supply amount. Before close, it must hold enough USDC to cover the
configured repayment bound, including accrued interest. A local `.env` keeps
the key out of shell history, but the key still exists in process memory while
the command runs. Use only a dedicated, minimally funded EOA.

## Inspect

Omit all phase flags to inspect the pending delegation, native balance, WETH
and USDC balances and Pool allowances, and Aave reserve positions:

```bash
go run ./examples/base-aave-lifecycle
```

Initialization verifies the reviewed account implementation runtime code and
loads the reviewed Base Aave market snapshot before reading account state.

## Dry Runs

Dry-run delegation or clear prints the intended lifecycle action without
building or signing a set-code transaction:

```bash
go run ./examples/base-aave-lifecycle --delegate
go run ./examples/base-aave-lifecycle --clear
```

Dry-run open and close build the real public SDK Flow and print its account,
plan kind, ordered steps, call targets, native values, selectors, checkpoints,
and calldata patches. They do not sign, estimate gas, simulate, or submit:

```bash
go run ./examples/base-aave-lifecycle \
  --open \
  --weth-supply 0.1 \
  --usdc-borrow 10

go run ./examples/base-aave-lifecycle \
  --close \
  --usdc-repay-limit 10.5
```

Amounts are explicit human-unit decimal strings. WETH accepts at most 18
decimal places and USDC accepts at most 6. No phase has a non-zero default.

## Base Fork

Start an EIP-7702-capable Anvil fork in one terminal:

```bash
BASE_RPC_URL=https://mainnet.base.org make anvil-base
```

In another terminal, set `BASE_RPC_URL=http://127.0.0.1:8545` and the dedicated
funded fork EOA key in `.env`, then run one broadcast phase at a time:

```bash
go run ./examples/base-aave-lifecycle --delegate --broadcast

go run ./examples/base-aave-lifecycle \
  --open --broadcast \
  --weth-supply 0.1 \
  --usdc-borrow 10

go run ./examples/base-aave-lifecycle \
  --close --broadcast \
  --usdc-repay-limit 10.5

go run ./examples/base-aave-lifecycle --clear --broadcast
```

Inspect the printed state and plan between commands. The close bound must be at
least the current variable debt, and the EOA's USDC balance must cover the full
bound. The close Flow atomically approves that bound, repays all USDC variable
debt, clears the allowance, withdraws all WETH collateral, and unwraps only the
WETH gained by that withdrawal.

## Base Execution

The same commands can target Base by setting `BASE_RPC_URL` in `.env` to a Base
endpoint and using a dedicated funded EOA. Run inspect and the matching dry run
first, review every target and amount, then add `--broadcast` only to the one
phase being submitted.

After a mined transaction, the command prints its hash, block, status, actual
gas used, and typed protocol event fields. A reverted Flow can leave EIP-7702
delegation installed, so inspect delegation explicitly after every failure.

Clear refuses to run while either reviewed reserve has collateral or debt, or
while a WETH or USDC Aave Pool allowance remains.
