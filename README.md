# DeFi Simplify

A Go library that simplifies interactions with DeFi protocols, making it easier to build DeFi applications.

## Supported Protocols

- **Aave V3**: Simplified interface for Aave V3 operations including:
  - Supply assets
  - Withdraw assets
  - Borrow assets
  - Repay loans
  - Approve delegations

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
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/shopspring/decimal"
    "github.com/tn606024/defi-simplify/client/contract"
    "github.com/tn606024/defi-simplify/config"
    "github.com/tn606024/defi-simplify/helper"
)

func main() {
    ctx := context.Background()

    // 1. Connect to Ethereum network
    client, err := ethclient.Dial("YOUR_RPC_URL")
    if err != nil {
        log.Fatalf("Failed to connect to network: %v", err)
    }

    // 2. Load your private key
    key, err := crypto.HexToECDSA("YOUR_PRIVATE_KEY")
    if err != nil {
        log.Fatalf("Failed to load private key: %v", err)
    }

    // 3. Create transaction options
    opts, err := bind.NewKeyedTransactorWithChainID(key, big.NewInt(int64(config.ChainInfo[config.Base].ChainID)))
    if err != nil {
        log.Fatalf("Failed to create transactor: %v", err)
    }

    // 4. Create DeFi client
    defiClient := contract.NewDefiClient(opts, client, helper.NewMsgSigner(key), config.Base)

    // 5. Supply and borrow USDC
    supplyAmount := decimal.NewFromFloat(0.1)  // 0.1 USDC
    borrowAmount := decimal.NewFromFloat(0.001) // 0.001 USDC

    receipt, err := defiClient.SupplyAndBorrowAaveV3Coin(ctx, config.USDC, supplyAmount, borrowAmount)
    if err != nil {
        log.Fatalf("Failed to execute supply and borrow: %v", err)
    }
    log.Printf("Transaction successful: %s", receipt.TxHash.Hex())
}
```

## Actions and Multicall

Actions are the building blocks of DeFi operations. Each action represents a single blockchain operation (like transferring tokens or supplying to Aave) that can be combined with others and executed in a single transaction using the multicall contract.

### What is an Action?

An action is a struct that implements these methods:

- `ToData()`: Converts the action into encoded contract call data
- `ToTransaction()`: Creates an Ethereum transaction
- `ToCallMsg()`: Creates a read-only call message
- `ToIMulticall3Call3()`: Formats the action for multicall execution

### Available Action Types

#### Token Actions

- `TransferAction`: Send tokens to another address
- `TransferFromAction`: Transfer tokens on behalf of another address
- `ApproveAction`: Allow another address to spend your tokens
- `PermitAction`: Approve tokens using a signature (gasless approval)

#### Aave V3 Actions

- `SupplyAction`: Deposit tokens into Aave V3
- `WithdrawAction`: Remove tokens from Aave V3
- `BorrowAction`: Borrow tokens from Aave V3
- `RepayAction`: Pay back borrowed tokens
- `DelegationAction`: Set up borrowing permissions

### Example: Supply and Borrow Flow

Here's an example of how actions are combined in the `SupplyAndBorrowAaveV3Coin` function:

```go
// Create a sequence of actions for supply and borrow
actions := []ExecuteAction{
    // 1. Create permit signature for token approval
    NewExecuteAction(permitAction, false),
    // 2. Transfer tokens from user to multicall contract
    NewExecuteAction(transferFromAction, false),
    // 3. Approve Aave V3 pool to spend tokens
    NewExecuteAction(approveAction, false),
    // 4. Supply tokens to Aave V3
    NewExecuteAction(supplyAction, false),
    // 5. Create delegation signature for borrowing
    NewExecuteAction(delegationWithSigAction, false),
    // 6. Borrow tokens from Aave V3
    NewExecuteAction(borrowAction, false),
    // 7. Transfer borrowed tokens to user
    NewExecuteAction(transferAction, false),
}

// Execute all actions in a single transaction
receipt, err := defiClient.ExecuteTxActions(ctx, actions)
```

Each action in the sequence is executed atomically within a single transaction. The order of actions is important:

1. First, we approve token spending through permit
2. Then transfer tokens to the multicall contract
3. Approve Aave V3 pool to spend our tokens
4. Supply tokens to Aave V3
5. Set up delegation for borrowing
6. Borrow tokens from Aave V3
7. Finally, transfer borrowed tokens to the user

If any required action fails, the entire transaction is reverted, ensuring atomicity of the operation.

## Common Use Cases

### Supply to Aave V3

```go
// Supply 1 USDC to Aave V3
amount := decimal.NewFromFloat(1.0)
receipt, err := defiClient.Aave.SupplyWithPermit(ctx, config.USDC, amount)
```

### Borrow from Aave V3

```go
// Borrow 0.5 USDC from Aave V3
amount := decimal.NewFromFloat(0.5)
receipt, err := defiClient.Aave.Borrow(ctx, config.USDC, amount)
```

### Transfer Tokens

```go
// Transfer 1 USDC to another address
amount := decimal.NewFromFloat(1.0)
receipt, err := defiClient.ERC20.Transfer(ctx, config.USDC, recipientAddress, amount)
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Security

This library is provided as-is with no guarantees. Please use at your own risk and always test thoroughly before using in production.

## Support

For support, please open an issue in the GitHub repository.
