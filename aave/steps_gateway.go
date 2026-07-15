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
	reserve   Reserve
	permit    erc20.PermitCapability
	amount    decimal.Decimal
	signature eip712Signature
}

// GatewaySpender resolves the configured chain's Aave WrappedTokenGateway as an ERC20 spender.
func GatewaySpender(market Market) erc20.Spender {
	return erc20.SpenderFunc(func(chain config.Chain) (common.Address, error) {
		if err := market.Validate(); err != nil {
			return common.Address{}, err
		}
		if market.Chain() != chain {
			return common.Address{}, fmt.Errorf(
				"Aave market chain %d does not match flow chain %d",
				market.Chain(),
				chain,
			)
		}
		gateway, ok := market.WrappedTokenGateway()
		if !ok {
			return common.Address{}, fmt.Errorf("Aave market does not define a WrappedTokenGateway")
		}
		return gateway, nil
	})
}

// DepositETH wraps and supplies native ETH through Aave's WrappedTokenGateway for the flow account.
// reserve must be the market's wrapped-native reserve used by that gateway.
func DepositETH(reserve Reserve, amount decimal.Decimal) defi.FlowStep {
	return gatewayStep{name: "aave.DepositETH", kind: depositETHStep, reserve: reserve, amount: amount}
}

// BorrowETH borrows WETH debt through Aave's WrappedTokenGateway and sends native ETH to the flow account.
// The flow account must first delegate WETH borrowing power to the gateway.
// reserve must be the market's wrapped-native reserve used by that gateway.
func BorrowETH(reserve Reserve, amount decimal.Decimal) defi.FlowStep {
	return gatewayStep{name: "aave.BorrowETH", kind: borrowETHStep, reserve: reserve, amount: amount}
}

// WithdrawETH withdraws WETH through Aave's WrappedTokenGateway and sends native ETH to the flow account.
// The flow account must first approve its aWETH to the gateway.
// reserve must be the market's wrapped-native reserve used by that gateway.
func WithdrawETH(reserve Reserve, amount decimal.Decimal) defi.FlowStep {
	return gatewayStep{name: "aave.WithdrawETH", kind: withdrawETHStep, reserve: reserve, amount: amount}
}

// WithdrawETHWithPermit withdraws through Aave's WrappedTokenGateway using an aWETH permit from the flow account.
// reserve must be the market's wrapped-native reserve used by that gateway.
func WithdrawETHWithPermit(
	reserve Reserve,
	permit erc20.PermitCapability,
	amount decimal.Decimal,
	deadline *big.Int,
	v uint8,
	r, s [32]byte,
) defi.FlowStep {
	return gatewayStep{
		name:      "aave.WithdrawETHWithPermit",
		kind:      withdrawETHWithPermitStep,
		reserve:   reserve,
		permit:    permit,
		amount:    amount,
		signature: newEIP712Signature(deadline, v, r, s),
	}
}

func (s gatewayStep) Build(ctx context.Context, env defi.BuildEnv) (defi.BuiltStep, error) {
	built := defi.BuiltStep{Name: s.name}
	if !s.amount.IsPositive() {
		return built, fmt.Errorf("amount must be positive")
	}
	resolved, err := resolveStepReserve(s.reserve, env.Chain)
	if err != nil {
		return built, err
	}
	if s.kind == withdrawETHWithPermitStep {
		if err := s.signature.validate(); err != nil {
			return built, err
		}
		if err := validatePermitCapability(s.permit, resolved.aToken); err != nil {
			return built, fmt.Errorf("resolve aToken permit capability: %w", err)
		}
	}

	gatewayAddress, ok := resolved.market.WrappedTokenGateway()
	if !ok {
		return built, fmt.Errorf("Aave market does not define a WrappedTokenGateway")
	}
	poolAddress := resolved.market.Pool()
	wethAddress := resolved.underlying.Address()
	amountWei := helper.ToWei(s.amount, resolved.underlying.Decimals())

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
