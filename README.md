# DeFi Simplify

A Go SDK for building DeFi actions and composing them into executable flows.

The current codebase focuses on Aave V3 and ERC20 operations on EVM chains. It started as a Multicall-based helper for Aave workflows, and is being refactored toward an `Action -> Call -> Executor` architecture so the same action builders can later be executed by different backends, including an EIP-7702 account executor.

## Current Status

Supported today:

- Aave V3 actions:
  - supply
  - withdraw
  - borrow
  - repay
  - approve delegation
  - delegation with signature
- ERC20 actions:
  - transfer
  - transferFrom
  - approve
  - permit
  - balanceOf
  - nonces
- Multicall execution for batched calls.
- Legacy Aave supply/borrow composed flow through Multicall.

Planned next:

- EIP-7702 transaction building.
- Simple7702Account-based execution.
- EOA-native Aave flows that avoid the legacy Multicall `msg.sender` limitation.

## Installation

```bash
go get github.com/tn606024/defi-simplify
```

## Quick Start

```go
package main

import (
	"context"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shopspring/decimal"
	"github.com/tn606024/defi-simplify/client/contract"
	"github.com/tn606024/defi-simplify/config"
	"github.com/tn606024/defi-simplify/helper"
)

func main() {
	ctx := context.Background()

	client, err := ethclient.Dial("YOUR_RPC_URL")
	if err != nil {
		log.Fatalf("failed to connect to network: %v", err)
	}

	key, err := crypto.HexToECDSA("YOUR_PRIVATE_KEY")
	if err != nil {
		log.Fatalf("failed to load private key: %v", err)
	}

	opts, err := bind.NewKeyedTransactorWithChainID(key, big.NewInt(int64(config.ChainInfo[config.Base].ChainID)))
	if err != nil {
		log.Fatalf("failed to create transactor: %v", err)
	}

	defiClient := contract.NewDefiClient(opts, client, helper.NewMsgSigner(key), config.Base)

	amount := decimal.NewFromFloat(1.0)
	receipt, err := defiClient.Aave.SupplyWithPermit(ctx, config.USDC, amount)
	if err != nil {
		log.Fatalf("failed to supply USDC: %v", err)
	}

	log.Printf("transaction successful: %s", receipt.TxHash.Hex())
}
```

## Architecture

The SDK is moving toward a neutral action architecture:

```text
Action -> Call -> Executor
```

An `Action` describes one protocol operation, such as ERC20 approve or Aave supply. Each action can encode itself into a neutral `Call`:

```go
type Call struct {
	Target common.Address
	Value  *big.Int
	Data   []byte
}
```

An `Executor` decides how those calls are submitted. Today the main batch executor is `MulticallExecutor`, which converts actions into Multicall3 calls and submits one transaction.

This separation matters because future execution paths should not need to understand Aave-specific or ERC20-specific builders. A future EIP-7702 executor should be able to consume the same action-generated calls while preserving EOA-native execution semantics.

## Testing

Run unit tests:

```bash
make test
```

Run unit tests plus whitespace checks:

```bash
make check
```

Integration tests are behind the `integration` build tag and are intended to run against Base mainnet state or a local Anvil fork of Base.

Start a local Anvil fork of Base mainnet:

```bash
BASE_RPC_URL=<base-mainnet-rpc-url> make anvil-base
```

In another terminal, run the integration tests against that fork:

```bash
BASE_RPC_URL=http://127.0.0.1:8545 make test-integration
```

## Actions

Actions are the building blocks of DeFi operations. They expose calldata-oriented methods for composition and transaction-oriented methods for direct execution.

Core methods:

- `ToData()`: returns target address and encoded calldata.
- `ToCall()`: returns a neutral `Call{Target, Value, Data}`.
- `ToCallMsg()`: converts a call into an `ethereum.CallMsg` for simulation or read calls.
- `ToTransaction()`: creates a direct Ethereum transaction for actions that support direct execution.

### ERC20 Actions

- `TransferAction`
- `TransferFromAction`
- `ApproveAction`
- `PermitAction`
- `BalanceOfAction`
- `NoncesAction`

### Aave V3 Actions

- `SupplyAction`
- `SupplyWithPermitAction`
- `WithdrawAction`
- `BorrowAction`
- `RepayAction`
- `RepayWithPermitAction`
- `DepositETHAction`
- `WithdrawETHAction`
- `ApproveDelegationAction`
- `DelegationWithSigAction`
- reserve/user data read actions

## Legacy Multicall Aave Flow

The existing composed Aave supply/borrow helper is:

```go
receipt, err := defiClient.LegacyMulticallSupplyAndBorrowAaveV3Coin(
	ctx,
	config.USDC,
	supplyAmount,
	borrowAmount,
)
```

This is intentionally named `LegacyMulticall...` because the flow depends on Multicall as the transaction caller.

The flow combines:

1. ERC20 permit for Multicall.
2. ERC20 transferFrom from user to Multicall.
3. ERC20 approve from Multicall to Aave Pool.
4. Aave supply.
5. Aave delegation signature.
6. Aave borrow.
7. ERC20 transfer from Multicall back to user.

This proves atomic composition, but it has an architectural limitation: when Aave is called through Multicall, Aave sees `msg.sender` as the Multicall contract, not the user's EOA. That affects position ownership, receiver semantics, approvals, delegation, callbacks, and other DeFi flows that depend on the caller being the user.

The planned Phase 1 EIP-7702 work exists to revisit this same action composition problem with EOA-native execution.

## Common Use Cases

### Supply to Aave V3

```go
amount := decimal.NewFromFloat(1.0)
receipt, err := defiClient.Aave.SupplyWithPermit(ctx, config.USDC, amount)
```

### Borrow from Aave V3

```go
amount := decimal.NewFromFloat(0.5)
receipt, err := defiClient.Aave.Borrow(ctx, config.USDC, amount)
```

### Transfer Tokens

```go
amount := decimal.NewFromFloat(1.0)
receipt, err := defiClient.ERC20.Transfer(ctx, config.USDC, recipientAddress, amount)
```

### Build Actions for Batch Execution

```go
approveAction := contract.BuildApproveAction(tokenAddress, spender, amount)
supplyAction := contract.BuildSupplyAction(poolAddress, tokenAddress, amount, user)

actions := []contract.ExecuteAction{
	contract.NewExecuteAction(approveAction, false),
	contract.NewExecuteAction(supplyAction, false),
}

receipt, err := defiClient.ExecuteTxActions(ctx, actions)
```

`ExecuteTxActions` currently uses the default Multicall executor.

## Roadmap

Near-term work:

- Keep protocol action builders reusable.
- Add EIP-7702 transaction building.
- Add Simple7702Account execution.
- Demonstrate an EOA-native Aave supply/borrow flow on Base.

Longer-term possibilities:

- Additional protocol action builders.
- More executor backends.
- Higher-level DeFi flow builders once the action and executor boundaries are stable.

## Security

This library is experimental and provided as-is with no guarantees. Please use it at your own risk and test thoroughly before using it with real funds.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
