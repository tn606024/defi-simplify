package erc20

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	defi "github.com/tn606024/defi-simplify"
	"github.com/tn606024/defi-simplify/client/contract"
	"github.com/tn606024/defi-simplify/config"
	"github.com/tn606024/defi-simplify/helper"
)

// Spender resolves the allowance spender address for the flow's chain.
type Spender interface {
	Address(chain config.Chain) (common.Address, error)
}

// SpenderFunc adapts a function into a Spender.
type SpenderFunc func(chain config.Chain) (common.Address, error)

func (f SpenderFunc) Address(chain config.Chain) (common.Address, error) {
	return f(chain)
}

// AddressSpender returns a Spender for a fixed address.
func AddressSpender(address common.Address) Spender {
	return SpenderFunc(func(config.Chain) (common.Address, error) {
		return address, nil
	})
}

type stepKind int

const (
	approveStep stepKind = iota
	transferStep
	transferFromStep
	permitStep
)

type step struct {
	name     string
	kind     stepKind
	token    config.Coin
	amount   decimal.Decimal
	spender  Spender
	from     common.Address
	to       common.Address
	owner    common.Address
	deadline *big.Int
	v        uint8
	r        [32]byte
	s        [32]byte
}

// Approve builds an ERC20 approve call.
func Approve(token config.Coin, spender Spender, amount decimal.Decimal) defi.FlowStep {
	return step{
		name:    "erc20.Approve",
		kind:    approveStep,
		token:   token,
		spender: spender,
		amount:  amount,
	}
}

// Transfer builds an ERC20 transfer call.
func Transfer(token config.Coin, to common.Address, amount decimal.Decimal) defi.FlowStep {
	return step{
		name:   "erc20.Transfer",
		kind:   transferStep,
		token:  token,
		to:     to,
		amount: amount,
	}
}

// TransferFrom builds an ERC20 transferFrom call.
func TransferFrom(token config.Coin, from common.Address, to common.Address, amount decimal.Decimal) defi.FlowStep {
	return step{
		name:   "erc20.TransferFrom",
		kind:   transferFromStep,
		token:  token,
		from:   from,
		to:     to,
		amount: amount,
	}
}

// Permit builds an ERC20 permit call for tokens that support EIP-2612-style permits.
func Permit(token config.Coin, owner common.Address, spender Spender, amount decimal.Decimal, deadline *big.Int, v uint8, r [32]byte, s [32]byte) defi.FlowStep {
	return step{
		name:     "erc20.Permit",
		kind:     permitStep,
		token:    token,
		owner:    owner,
		spender:  spender,
		amount:   amount,
		deadline: deadline,
		v:        v,
		r:        r,
		s:        s,
	}
}

func (s step) Build(ctx context.Context, env defi.BuildEnv) (defi.BuiltStep, error) {
	built := defi.BuiltStep{Name: s.name}
	if s.amount.IsNegative() {
		return built, fmt.Errorf("amount must not be negative")
	}

	tokenAddress, err := s.token.Address(env.Chain)
	if err != nil {
		return built, fmt.Errorf("resolve token: %w", err)
	}
	decimals, err := s.token.Decimals()
	if err != nil {
		return built, fmt.Errorf("resolve token decimals: %w", err)
	}
	amountWei := helper.ToWei(s.amount, decimals)

	var action defi.Action
	switch s.kind {
	case approveStep:
		spender, err := s.resolveSpender(env.Chain)
		if err != nil {
			return built, err
		}
		action = contract.BuildApproveAction(tokenAddress, spender, amountWei)
	case transferStep:
		action = contract.BuildTransferAction(tokenAddress, s.to, amountWei)
	case transferFromStep:
		action = contract.BuildTransferFromAction(tokenAddress, s.from, s.to, amountWei)
	case permitStep:
		spender, err := s.resolveSpender(env.Chain)
		if err != nil {
			return built, err
		}
		action = contract.BuildPermitAction(tokenAddress, s.owner, spender, amountWei, s.deadline, s.v, s.r, s.s)
	default:
		return built, fmt.Errorf("unsupported ERC20 step kind %d", s.kind)
	}

	call, err := action.ToCall(ctx, env.Conn, nil)
	if err != nil {
		return built, err
	}
	if call == nil {
		return built, fmt.Errorf("action returned nil call")
	}
	built.Calls = []defi.Call{*call}
	return built, nil
}

func (s step) resolveSpender(chain config.Chain) (common.Address, error) {
	if s.spender == nil {
		return common.Address{}, fmt.Errorf("spender is nil")
	}
	spender, err := s.spender.Address(chain)
	if err != nil {
		return common.Address{}, fmt.Errorf("resolve spender: %w", err)
	}
	return spender, nil
}
