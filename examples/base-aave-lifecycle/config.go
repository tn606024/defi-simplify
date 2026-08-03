package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/shopspring/decimal"
)

const (
	baseRPCURLEnv = "BASE_RPC_URL"
	privateKeyEnv = "PRIVATE_KEY"
)

type lifecyclePhase string

const (
	phaseInspect  lifecyclePhase = "inspect"
	phaseDelegate lifecyclePhase = "delegate"
	phaseOpen     lifecyclePhase = "open"
	phaseClose    lifecyclePhase = "close"
	phaseClear    lifecyclePhase = "clear"
)

type commandConfig struct {
	rpcURL           string
	privateKey       string
	phase            lifecyclePhase
	broadcast        bool
	wethSupplyAmount decimal.Decimal
	usdcBorrowAmount decimal.Decimal
	usdcRepayLimit   decimal.Decimal
}

func parseCommandConfig(
	args []string,
	getenv func(string) string,
) (commandConfig, error) {
	if getenv == nil {
		return commandConfig{}, errors.New("environment reader is nil")
	}

	flags := flag.NewFlagSet("base-aave-lifecycle", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	delegate := flags.Bool("delegate", false, "prepare or submit EIP-7702 delegation")
	open := flags.Bool("open", false, "prepare or submit the Aave open flow")
	closePosition := flags.Bool("close", false, "prepare or submit the Aave close flow")
	clearDelegation := flags.Bool("clear", false, "prepare or submit delegation clear")
	broadcast := flags.Bool("broadcast", false, "submit the selected phase")
	wethSupply := flags.String("weth-supply", "", "exact WETH collateral amount in human units")
	usdcBorrow := flags.String("usdc-borrow", "", "exact USDC borrow amount in human units")
	usdcRepayLimit := flags.String("usdc-repay-limit", "", "temporary USDC repayment bound in human units")
	if err := flags.Parse(args); err != nil {
		return commandConfig{}, fmt.Errorf("parse flags: %w", err)
	}
	if flags.NArg() != 0 {
		return commandConfig{}, fmt.Errorf("unexpected positional arguments: %s", strings.Join(flags.Args(), " "))
	}

	phase, err := selectPhase(*delegate, *open, *closePosition, *clearDelegation)
	if err != nil {
		return commandConfig{}, err
	}
	if phase == phaseInspect && *broadcast {
		return commandConfig{}, errors.New("--broadcast requires one lifecycle phase")
	}

	config := commandConfig{
		rpcURL:     strings.TrimSpace(getenv(baseRPCURLEnv)),
		privateKey: strings.TrimSpace(getenv(privateKeyEnv)),
		phase:      phase,
		broadcast:  *broadcast,
	}
	if config.rpcURL == "" {
		return commandConfig{}, fmt.Errorf("%s is required", baseRPCURLEnv)
	}
	if config.privateKey == "" {
		return commandConfig{}, fmt.Errorf("%s is required", privateKeyEnv)
	}

	switch phase {
	case phaseOpen:
		config.wethSupplyAmount, err = parsePositiveAmount(*wethSupply, 18, "--weth-supply")
		if err != nil {
			return commandConfig{}, err
		}
		config.usdcBorrowAmount, err = parsePositiveAmount(*usdcBorrow, 6, "--usdc-borrow")
		if err != nil {
			return commandConfig{}, err
		}
		if strings.TrimSpace(*usdcRepayLimit) != "" {
			return commandConfig{}, errors.New("--usdc-repay-limit is only valid with --close")
		}
	case phaseClose:
		config.usdcRepayLimit, err = parsePositiveAmount(*usdcRepayLimit, 6, "--usdc-repay-limit")
		if err != nil {
			return commandConfig{}, err
		}
		if strings.TrimSpace(*wethSupply) != "" || strings.TrimSpace(*usdcBorrow) != "" {
			return commandConfig{}, errors.New("--weth-supply and --usdc-borrow are only valid with --open")
		}
	default:
		if strings.TrimSpace(*wethSupply) != "" || strings.TrimSpace(*usdcBorrow) != "" || strings.TrimSpace(*usdcRepayLimit) != "" {
			return commandConfig{}, errors.New("amount flags require --open or --close")
		}
	}
	return config, nil
}

func selectPhase(delegate, open, closePosition, clear bool) (lifecyclePhase, error) {
	selected := make([]lifecyclePhase, 0, 1)
	if delegate {
		selected = append(selected, phaseDelegate)
	}
	if open {
		selected = append(selected, phaseOpen)
	}
	if closePosition {
		selected = append(selected, phaseClose)
	}
	if clear {
		selected = append(selected, phaseClear)
	}
	if len(selected) > 1 {
		return "", errors.New("select exactly one of --delegate, --open, --close, or --clear")
	}
	if len(selected) == 0 {
		return phaseInspect, nil
	}
	return selected[0], nil
}

func parsePositiveAmount(raw string, decimals uint8, flagName string) (decimal.Decimal, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return decimal.Decimal{}, fmt.Errorf("%s is required", flagName)
	}
	value, err := decimal.NewFromString(trimmed)
	if err != nil {
		return decimal.Decimal{}, fmt.Errorf("%s must be a decimal string: %w", flagName, err)
	}
	if !value.IsPositive() {
		return decimal.Decimal{}, fmt.Errorf("%s must be positive", flagName)
	}
	scaled := value.Shift(int32(decimals))
	if !scaled.Equal(decimal.NewFromBigInt(scaled.BigInt(), 0)) {
		return decimal.Decimal{}, fmt.Errorf("%s exceeds %d decimal places", flagName, decimals)
	}
	return value, nil
}

func printUsage(out io.Writer) {
	fmt.Fprintln(out, "Usage: base-aave-lifecycle [--delegate|--open|--close|--clear] [--broadcast] [amount flags]")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Environment: BASE_RPC_URL, PRIVATE_KEY")
	fmt.Fprintln(out, "Phases: --delegate, --open, --close, --clear (omit all phases to inspect)")
	fmt.Fprintln(out, "Open amounts: --weth-supply <decimal> --usdc-borrow <decimal>")
	fmt.Fprintln(out, "Close amount: --usdc-repay-limit <decimal>")
	fmt.Fprintln(out, "State changes require --broadcast; otherwise the selected phase is a dry run.")
}
