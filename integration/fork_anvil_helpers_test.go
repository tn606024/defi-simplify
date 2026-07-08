//go:build integration

package integration

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/tn606024/defi-simplify/bind/erc20"
	sdkcontract "github.com/tn606024/defi-simplify/client/contract"
	"github.com/tn606024/defi-simplify/config"
)

// Base Aave V3 Pool holds USDC on mainnet and is impersonated only on local forks.
var baseUSDCFunder = common.HexToAddress("0xA238Dd80C259a72e81d7e4664a9801593F98d1c5")

func setForkETHBalance(ctx context.Context, client *rpc.Client, account common.Address, balance *big.Int) error {
	if balance == nil {
		return fmt.Errorf("balance is nil")
	}
	return client.CallContext(ctx, nil, "anvil_setBalance", account, (*hexutil.Big)(balance))
}

func impersonateForkAccount(ctx context.Context, client *rpc.Client, account common.Address) error {
	return client.CallContext(ctx, nil, "anvil_impersonateAccount", account)
}

func stopImpersonatingForkAccount(ctx context.Context, client *rpc.Client, account common.Address) error {
	err := client.CallContext(ctx, nil, "anvil_stopImpersonatingAccount", account)
	if err != nil && strings.Contains(strings.ToLower(err.Error()), "method not found") {
		return nil
	}
	return err
}

func fundBaseUSDCFromHolder(ctx context.Context, rpcClient *rpc.Client, ethClient *ethclient.Client, recipient common.Address, amount *big.Int) (err error) {
	if amount == nil {
		return fmt.Errorf("amount is nil")
	}

	usdc, err := config.USDC.Address(config.Base)
	if err != nil {
		return err
	}
	token, err := erc20.NewErc20(usdc, ethClient)
	if err != nil {
		return err
	}
	holderBalance, err := token.BalanceOf(&bind.CallOpts{Context: ctx}, baseUSDCFunder)
	if err != nil {
		return fmt.Errorf("read Base USDC funder balance: %w", err)
	}
	if holderBalance.Cmp(amount) < 0 {
		return fmt.Errorf("Base USDC funder balance %s is less than requested amount %s", holderBalance.String(), amount.String())
	}

	holderETH := big.NewInt(1_000_000_000_000_000_000)
	if err := setForkETHBalance(ctx, rpcClient, baseUSDCFunder, holderETH); err != nil {
		return fmt.Errorf("fund Base USDC holder ETH: %w", err)
	}
	if err := impersonateForkAccount(ctx, rpcClient, baseUSDCFunder); err != nil {
		return fmt.Errorf("impersonate Base USDC holder: %w", err)
	}
	defer func() {
		stopErr := stopImpersonatingForkAccount(ctx, rpcClient, baseUSDCFunder)
		if err == nil && stopErr != nil {
			err = fmt.Errorf("stop impersonating Base USDC holder: %w", stopErr)
		}
	}()

	action := sdkcontract.BuildTransferAction(usdc, recipient, amount)
	target, data, err := action.ToData(ctx, ethClient, nil)
	if err != nil {
		return fmt.Errorf("encode USDC transfer: %w", err)
	}

	txHash, err := sendForkTransaction(ctx, rpcClient, baseUSDCFunder, target, data)
	if err != nil {
		return fmt.Errorf("send USDC transfer: %w", err)
	}
	receipt, err := bind.WaitMinedHash(ctx, ethClient, txHash)
	if err != nil {
		return fmt.Errorf("wait for USDC transfer %s: %w", txHash.Hex(), err)
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return fmt.Errorf("USDC transfer %s reverted with status %d", txHash.Hex(), receipt.Status)
	}

	return nil
}

func sendForkTransaction(ctx context.Context, client *rpc.Client, from common.Address, to common.Address, data []byte) (common.Hash, error) {
	args := map[string]interface{}{
		"from": from,
		"to":   to,
		"gas":  hexutil.EncodeUint64(100_000),
		"data": hexutil.Encode(data),
	}

	var txHash common.Hash
	if err := client.CallContext(ctx, &txHash, "eth_sendTransaction", args); err != nil {
		return common.Hash{}, err
	}
	return txHash, nil
}
