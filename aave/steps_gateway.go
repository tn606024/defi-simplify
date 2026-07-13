package aave

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	defi "github.com/tn606024/defi-simplify"
	"github.com/tn606024/defi-simplify/client/contract"
	"github.com/tn606024/defi-simplify/config"
	"github.com/tn606024/defi-simplify/erc20"
	"github.com/tn606024/defi-simplify/helper"
)

type gatewayStepKind uint8

const (
	depositETHStep gatewayStepKind = iota
	borrowETHStep
	withdrawETHStep
	withdrawETHWithPermitStep
)

type gatewayStep struct {
	name      string
	kind      gatewayStepKind
	amount    decimal.Decimal
	signature eip712Signature
}

// GatewaySpender resolves the configured chain's Aave WrappedTokenGateway as an ERC20 spender.
func GatewaySpender() erc20.Spender {
	return erc20.SpenderFunc(func(chain config.Chain) (common.Address, error) {
		return chain.WrappedTokenGatewayV3Address()
	})
}

// DepositETH wraps and supplies native ETH through Aave's WrappedTokenGateway for the flow account.
func DepositETH(amount decimal.Decimal) defi.FlowStep {
	return gatewayStep{name: "aave.DepositETH", kind: depositETHStep, amount: amount}
}

// BorrowETH borrows WETH debt through Aave's WrappedTokenGateway and sends native ETH to the flow account.
// The flow account must first delegate WETH borrowing power to the gateway.
func BorrowETH(amount decimal.Decimal) defi.FlowStep {
	return gatewayStep{name: "aave.BorrowETH", kind: borrowETHStep, amount: amount}
}

// WithdrawETH withdraws WETH through Aave's WrappedTokenGateway and sends native ETH to the flow account.
// The flow account must first approve its aWETH to the gateway.
func WithdrawETH(amount decimal.Decimal) defi.FlowStep {
	return gatewayStep{name: "aave.WithdrawETH", kind: withdrawETHStep, amount: amount}
}

// WithdrawETHWithPermit withdraws through Aave's WrappedTokenGateway using an aWETH permit from the flow account.
func WithdrawETHWithPermit(
	amount decimal.Decimal,
	deadline *big.Int,
	v uint8,
	r, s [32]byte,
) defi.FlowStep {
	return gatewayStep{
		name:      "aave.WithdrawETHWithPermit",
		kind:      withdrawETHWithPermitStep,
		amount:    amount,
		signature: newEIP712Signature(deadline, v, r, s),
	}
}

func (s gatewayStep) Build(ctx context.Context, env defi.BuildEnv) (defi.BuiltStep, error) {
	built := defi.BuiltStep{Name: s.name}
	if !s.amount.IsPositive() {
		return built, fmt.Errorf("amount must be positive")
	}
	if s.kind == withdrawETHWithPermitStep {
		if err := s.signature.validate(); err != nil {
			return built, err
		}
		permitSupported, err := config.AWETH.PermitSupported(env.Chain)
		if err != nil {
			return built, fmt.Errorf("resolve aWETH permit support: %w", err)
		}
		if !permitSupported {
			return built, fmt.Errorf("aWETH does not support permit on chain %d", env.Chain)
		}
	}

	gatewayAddress, err := env.Chain.WrappedTokenGatewayV3Address()
	if err != nil {
		return built, fmt.Errorf("resolve Aave WrappedTokenGateway: %w", err)
	}
	poolAddress, err := env.Chain.AaveV3PoolAddress()
	if err != nil {
		return built, fmt.Errorf("resolve Aave pool: %w", err)
	}
	wethAddress, err := config.WETH.Address(env.Chain)
	if err != nil {
		return built, fmt.Errorf("resolve wrapped gas token: %w", err)
	}
	decimals, err := env.Chain.GasTokenDecimals()
	if err != nil {
		return built, fmt.Errorf("resolve gas token decimals: %w", err)
	}
	amountWei := helper.ToWei(s.amount, decimals)

	var (
		action      defi.Action
		expectation defi.EventExpectation
	)
	switch s.kind {
	case depositETHStep:
		action = contract.BuildDepositETHAction(gatewayAddress, poolAddress, env.Account, 0, amountWei)
		expectation = ExpectSupply(
			poolAddress,
			wethAddress,
			gatewayAddress,
			env.Account,
			defi.Exact(amountWei),
		)
	case borrowETHStep:
		action = contract.BuildBorrowETHAction(gatewayAddress, poolAddress, amountWei)
		expectation = ExpectBorrow(
			poolAddress,
			wethAddress,
			gatewayAddress,
			env.Account,
			VariableInterestRateMode,
			defi.Exact(amountWei),
		)
	case withdrawETHStep:
		action = contract.BuildWithdrawETHAction(gatewayAddress, poolAddress, amountWei, env.Account)
		expectation = ExpectWithdraw(
			poolAddress,
			wethAddress,
			gatewayAddress,
			gatewayAddress,
			defi.Exact(amountWei),
		)
	case withdrawETHWithPermitStep:
		action = contract.BuildWithdrawETHWithPermitAction(
			gatewayAddress,
			poolAddress,
			amountWei,
			env.Account,
			s.signature.deadline,
			s.signature.v,
			s.signature.r,
			s.signature.s,
		)
		expectation = ExpectWithdraw(
			poolAddress,
			wethAddress,
			gatewayAddress,
			gatewayAddress,
			defi.Exact(amountWei),
		)
	default:
		return built, fmt.Errorf("unsupported Aave gateway step kind %d", s.kind)
	}

	call, err := action.ToCall(ctx, env.Conn, nil)
	if err != nil {
		return built, err
	}
	if call == nil {
		return built, fmt.Errorf("action returned nil call")
	}
	built.Calls = []defi.Call{*call}
	built.Expectations = []defi.EventExpectation{expectation}
	return built, nil
}
