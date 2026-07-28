package erc20

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	defi "github.com/tn606024/defi-simplify"
	"github.com/tn606024/defi-simplify/amount"
	"github.com/tn606024/defi-simplify/client/contract"
	"github.com/tn606024/defi-simplify/config"
	"github.com/tn606024/defi-simplify/helper"
	"github.com/tn606024/defi-simplify/token"
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

const (
	approveAmountOffset      uint32 = 36
	transferAmountOffset     uint32 = 36
	transferFromAmountOffset uint32 = 68
	permitAmountOffset       uint32 = 68
)

type step struct {
	name     string
	kind     stepKind
	token    token.Token
	permit   PermitCapability
	amount   amount.Source
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
func Approve(asset token.Token, spender Spender, value amount.Source) defi.FlowStep {
	return step{
		name:    "erc20.Approve",
		kind:    approveStep,
		token:   asset,
		spender: spender,
		amount:  value,
	}
}

// Transfer builds an ERC20 transfer call.
func Transfer(asset token.Token, to common.Address, value amount.Source) defi.FlowStep {
	return step{
		name:   "erc20.Transfer",
		kind:   transferStep,
		token:  asset,
		to:     to,
		amount: value,
	}
}

// TransferFrom builds an ERC20 transferFrom call.
func TransferFrom(asset token.Token, from common.Address, to common.Address, value amount.Source) defi.FlowStep {
	return step{
		name:   "erc20.TransferFrom",
		kind:   transferFromStep,
		token:  asset,
		from:   from,
		to:     to,
		amount: value,
	}
}

// Permit builds an ERC20 permit call for tokens that support EIP-2612-style permits.
func Permit(capability PermitCapability, owner common.Address, spender Spender, value amount.Source, deadline *big.Int, v uint8, r [32]byte, s [32]byte) defi.FlowStep {
	return step{
		name:     "erc20.Permit",
		kind:     permitStep,
		permit:   capability,
		owner:    owner,
		spender:  spender,
		amount:   value,
		deadline: deadline,
		v:        v,
		r:        r,
		s:        s,
	}
}

func (s step) Build(ctx context.Context, env defi.BuildEnv) (defi.BuiltStep, error) {
	built := defi.BuiltStep{Name: s.name}
	asset := s.token
	if s.kind == permitStep {
		if err := s.permit.Validate(); err != nil {
			return built, err
		}
		asset = s.permit.Token()
	}
	if err := asset.Validate(); err != nil {
		return built, fmt.Errorf("resolve token: %w", err)
	}
	if asset.Chain() != env.Chain {
		return built, fmt.Errorf(
			"token chain %d does not match flow chain %d",
			asset.Chain(),
			env.Chain,
		)
	}
	tokenAddress := asset.Address()
	allowDynamic := s.kind == approveStep || s.kind == transferStep
	amountWei, patch, amountConstraints, err := resolveStepAmount(
		s.amount,
		asset,
		allowDynamic,
		s.amountOffset(),
	)
	if err != nil {
		return built, err
	}

	var (
		action      defi.Action
		expectation defi.EventExpectation
	)
	switch s.kind {
	case approveStep:
		spender, err := s.resolveSpender(env.Chain)
		if err != nil {
			return built, err
		}
		action = contract.BuildApproveAction(tokenAddress, spender, amountWei)
		expectation = ExpectApproval(tokenAddress, env.Account, spender, amountConstraints...)
	case transferStep:
		action = contract.BuildTransferAction(tokenAddress, s.to, amountWei)
		expectation = ExpectTransfer(tokenAddress, env.Account, s.to, amountConstraints...)
	case transferFromStep:
		action = contract.BuildTransferFromAction(tokenAddress, s.from, s.to, amountWei)
		expectation = ExpectTransfer(tokenAddress, s.from, s.to, amountConstraints...)
	case permitStep:
		spender, err := s.resolveSpender(env.Chain)
		if err != nil {
			return built, err
		}
		action = contract.BuildPermitAction(tokenAddress, s.owner, spender, amountWei, s.deadline, s.v, s.r, s.s)
		expectation = ExpectApproval(tokenAddress, s.owner, spender, amountConstraints...)
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
	planned := defi.PlannedCall{Call: *call}
	if patch != nil {
		planned.Patches = []defi.CalldataPatch{*patch}
	}
	built.Calls = []defi.PlannedCall{planned}
	built.Expectations = []defi.EventExpectation{expectation}
	return built, nil
}

func (s step) amountOffset() uint32 {
	switch s.kind {
	case approveStep:
		return approveAmountOffset
	case transferStep:
		return transferAmountOffset
	case transferFromStep:
		return transferFromAmountOffset
	case permitStep:
		return permitAmountOffset
	default:
		return 0
	}
}

func resolveStepAmount(
	source amount.Source,
	asset token.Token,
	allowDynamic bool,
	offset uint32,
) (*big.Int, *defi.CalldataPatch, []defi.AmountConstraint, error) {
	if err := source.Validate(); err != nil {
		return nil, nil, nil, fmt.Errorf("resolve amount: %w", err)
	}
	if exact, ok := source.ExactValue(); ok {
		value := helper.ToWei(exact, asset.Decimals())
		return value, nil, []defi.AmountConstraint{defi.Exact(value)}, nil
	}
	if !allowDynamic {
		return nil, nil, nil, fmt.Errorf("resolve amount: runtime amount source is not supported for this ERC20 step")
	}
	sourceToken, _ := source.Token()
	if !sourceToken.SameAsset(asset.Ref()) {
		return nil, nil, nil, fmt.Errorf(
			"resolve amount: source token %s does not match ERC20 token %s",
			sourceToken.Address().Hex(),
			asset.Address().Hex(),
		)
	}
	return new(big.Int), &defi.CalldataPatch{Source: source, Offset: offset}, nil, nil
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
