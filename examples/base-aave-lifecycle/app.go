package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	defi "github.com/tn606024/defi-simplify"
	"github.com/tn606024/defi-simplify/aave"
	"github.com/tn606024/defi-simplify/assets/base"
	"github.com/tn606024/defi-simplify/client/account/defisimplify7702"
	"github.com/tn606024/defi-simplify/client/account/eip7702"
	"github.com/tn606024/defi-simplify/config"
)

type application struct {
	config         commandConfig
	client         *ethclient.Client
	txOpts         *bind.TransactOpts
	manager        *eip7702.Manager
	account        common.Address
	implementation defisimplify7702.ContractDeployment
	market         aave.Market
	snapshotBlock  uint64
	snapshotHash   common.Hash
	wethReserve    aave.Reserve
	usdcReserve    aave.Reserve
	out            io.Writer
}

func runCLI(
	ctx context.Context,
	args []string,
	getenv func(string) string,
	out io.Writer,
) error {
	config, err := parseCommandConfig(args, getenv)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			printUsage(out)
			return nil
		}
		return err
	}
	app, err := initializeApplication(ctx, config, out)
	if err != nil {
		return err
	}
	defer app.client.Close()
	return app.run(ctx)
}

func initializeApplication(
	ctx context.Context,
	command commandConfig,
	out io.Writer,
) (*application, error) {
	if out == nil {
		return nil, errors.New("output writer is nil")
	}
	client, err := ethclient.DialContext(ctx, command.rpcURL)
	if err != nil {
		return nil, fmt.Errorf("connect Base RPC: %w", err)
	}
	keepClient := false
	defer func() {
		if !keepClient {
			client.Close()
		}
	}()

	expectedChainID, err := config.Base.ChainID()
	if err != nil {
		return nil, err
	}
	chainID, err := client.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("read chain ID: %w", err)
	}
	if !chainID.IsInt64() || chainID.Int64() != int64(expectedChainID) {
		return nil, fmt.Errorf("unsupported chain ID %s, expected %d", chainID, expectedChainID)
	}

	keyHex := strings.TrimPrefix(strings.TrimPrefix(command.privateKey, "0x"), "0X")
	key, err := crypto.HexToECDSA(keyHex)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", privateKeyEnv, err)
	}
	txOpts, err := bind.NewKeyedTransactorWithChainID(key, chainID)
	if err != nil {
		return nil, fmt.Errorf("create Base signer: %w", err)
	}
	txOpts.Context = ctx

	implementation, err := defisimplify7702.ResolveAndVerifyAccountDeployment(
		ctx,
		client,
		config.Base,
	)
	if err != nil {
		return nil, fmt.Errorf("initialize reviewed account deployment: %w", err)
	}
	snapshot, err := aave.LoadBaseV3Snapshot(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("initialize Base Aave market: %w", err)
	}
	wethReserve, err := snapshot.Reserve(base.WETH)
	if err != nil {
		return nil, fmt.Errorf("resolve WETH reserve: %w", err)
	}
	usdcReserve, err := snapshot.Reserve(base.USDC)
	if err != nil {
		return nil, fmt.Errorf("resolve USDC reserve: %w", err)
	}
	manager, err := eip7702.NewManager(client, txOpts, key, chainID)
	if err != nil {
		return nil, fmt.Errorf("initialize delegation manager: %w", err)
	}

	keepClient = true
	return &application{
		config:         command,
		client:         client,
		txOpts:         txOpts,
		manager:        manager,
		account:        txOpts.From,
		implementation: implementation,
		market:         snapshot.Market(),
		snapshotBlock:  snapshot.BlockNumber(),
		snapshotHash:   snapshot.BlockHash(),
		wethReserve:    wethReserve,
		usdcReserve:    usdcReserve,
		out:            out,
	}, nil
}

func (app *application) run(ctx context.Context) error {
	state, err := app.inspectAccount(ctx)
	if err != nil {
		return err
	}
	app.printHeader()
	app.printState("before", state)
	if err := validatePhasePreconditions(app.config, state, app.implementation.Address); err != nil {
		return fmt.Errorf("%s precondition: %w", app.config.phase, err)
	}

	switch app.config.phase {
	case phaseInspect:
		return nil
	case phaseDelegate:
		err = app.runDelegationPhase(ctx, false)
	case phaseOpen:
		err = app.runFlowPhase(ctx, buildOpenFlow(
			app.account,
			app.market,
			app.wethReserve,
			app.usdcReserve,
			app.config.wethSupplyAmount,
			app.config.usdcBorrowAmount,
		))
	case phaseClose:
		err = app.runFlowPhase(ctx, buildCloseFlow(
			app.account,
			app.market,
			app.wethReserve,
			app.usdcReserve,
			app.config.usdcRepayLimit,
		))
	case phaseClear:
		err = app.runDelegationPhase(ctx, true)
	default:
		err = fmt.Errorf("unsupported lifecycle phase %q", app.config.phase)
	}
	if err != nil {
		return err
	}
	if !app.config.broadcast {
		return nil
	}
	after, err := app.inspectAccount(ctx)
	if err != nil {
		return fmt.Errorf("inspect account after %s: %w", app.config.phase, err)
	}
	app.printState("after", after)
	return nil
}

func (app *application) runDelegationPhase(ctx context.Context, clear bool) error {
	action := "delegate"
	implementation := app.implementation.Address
	if clear {
		action = "clear"
		implementation = common.Address{}
	}
	fmt.Fprintf(app.out, "action=%s implementation=%s broadcast=%t\n", action, implementation.Hex(), app.config.broadcast)
	if !app.config.broadcast {
		return nil
	}

	var (
		tx  *types.Transaction
		err error
	)
	if clear {
		tx, err = app.manager.Clear(ctx)
	} else {
		tx, err = app.manager.Delegate(ctx, implementation)
	}
	if err != nil {
		return fmt.Errorf("submit delegation %s: %w", action, err)
	}
	fmt.Fprintf(app.out, "transaction.submitted hash=%s\n", tx.Hash().Hex())
	receipt, err := bind.WaitMined(ctx, app.client, tx)
	if err != nil {
		return fmt.Errorf("wait for delegation %s %s: %w", action, tx.Hash().Hex(), err)
	}
	app.printReceipt(receipt)
	if receipt.Status != types.ReceiptStatusSuccessful {
		return fmt.Errorf("delegation %s reverted; delegation state must be inspected explicitly", action)
	}
	return nil
}

func (app *application) runFlowPhase(ctx context.Context, flow *defi.Flow) error {
	plan, err := flow.Build(ctx, app.client)
	if err != nil {
		return fmt.Errorf("build %s flow: %w", app.config.phase, err)
	}
	app.printPlan(plan)
	if !app.config.broadcast {
		return nil
	}

	result, executionErr := defi.NewRunner(app.client, app.txOpts, config.Base).Execute(ctx, flow)
	if result != nil {
		app.printExecution(result)
	}
	if executionErr != nil {
		return fmt.Errorf(
			"execute %s flow: %w; EIP-7702 delegation can remain installed after a revert",
			app.config.phase,
			executionErr,
		)
	}
	return nil
}

func (app *application) printReceipt(receipt *types.Receipt) {
	if receipt == nil {
		return
	}
	fmt.Fprintf(
		app.out,
		"transaction.mined hash=%s block=%s status=%d gas_used=%d\n",
		receipt.TxHash.Hex(),
		receipt.BlockNumber,
		receipt.Status,
		receipt.GasUsed,
	)
}

func (app *application) printExecution(result *defi.ExecutionResult) {
	if result == nil {
		return
	}
	app.printReceipt(result.Receipt)
	printTypedEvents(app.out, result)
}
