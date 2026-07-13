package aave

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	defi "github.com/tn606024/defi-simplify"
	bindaave "github.com/tn606024/defi-simplify/bind/aave"
)

var (
	repayEventTopic    = crypto.Keccak256Hash([]byte("Repay(address,address,address,uint256,bool)"))
	withdrawEventTopic = crypto.Keccak256Hash([]byte("Withdraw(address,address,address,uint256)"))
)

type poolExitEventKind uint8

const (
	repayPoolEvent poolExitEventKind = iota
	withdrawPoolEvent
)

// RepayEvent is the stable SDK representation of an Aave V3 Repay event.
type RepayEvent struct {
	Metadata   defi.EventMetadata
	Asset      common.Address
	User       common.Address
	Repayer    common.Address
	Amount     *big.Int
	UseATokens bool
}

// EventMetadata returns protocol-neutral event identity.
func (e *RepayEvent) EventMetadata() defi.EventMetadata {
	return e.Metadata
}

// WithdrawEvent is the stable SDK representation of an Aave V3 Withdraw event.
type WithdrawEvent struct {
	Metadata defi.EventMetadata
	Asset    common.Address
	User     common.Address
	To       common.Address
	Amount   *big.Int
}

// EventMetadata returns protocol-neutral event identity.
func (e *WithdrawEvent) EventMetadata() defi.EventMetadata {
	return e.Metadata
}

type poolExitEventExpectation struct {
	kind        poolExitEventKind
	pool        common.Address
	asset       common.Address
	user        common.Address
	repayer     common.Address
	to          common.Address
	useATokens  bool
	constraints []defi.AmountConstraint
}

// ExpectRepay validates an Aave Repay event emitted by pool.
func ExpectRepay(
	pool,
	asset,
	user,
	repayer common.Address,
	useATokens bool,
	constraints ...defi.AmountConstraint,
) defi.EventExpectation {
	return &poolExitEventExpectation{
		kind:        repayPoolEvent,
		pool:        pool,
		asset:       asset,
		user:        user,
		repayer:     repayer,
		useATokens:  useATokens,
		constraints: append([]defi.AmountConstraint(nil), constraints...),
	}
}

// ExpectWithdraw validates an Aave Withdraw event emitted by pool.
func ExpectWithdraw(
	pool,
	asset,
	user,
	to common.Address,
	constraints ...defi.AmountConstraint,
) defi.EventExpectation {
	return &poolExitEventExpectation{
		kind:        withdrawPoolEvent,
		pool:        pool,
		asset:       asset,
		user:        user,
		to:          to,
		constraints: append([]defi.AmountConstraint(nil), constraints...),
	}
}

func (e *poolExitEventExpectation) ExpectationName() string {
	if e == nil {
		return ""
	}
	switch e.kind {
	case repayPoolEvent:
		return "aave.Repay"
	case withdrawPoolEvent:
		return "aave.Withdraw"
	default:
		return "aave.<unknown>"
	}
}

func (e *poolExitEventExpectation) IsCandidate(log *types.Log) bool {
	if e == nil || log == nil || log.Address != e.pool || len(log.Topics) == 0 {
		return false
	}
	switch e.kind {
	case repayPoolEvent:
		return log.Topics[0] == repayEventTopic
	case withdrawPoolEvent:
		return log.Topics[0] == withdrawEventTopic
	default:
		return false
	}
}

func (e *poolExitEventExpectation) Decode(log *types.Log) (defi.DecodedEvent, error) {
	if e == nil || log == nil {
		return nil, fmt.Errorf("Aave Pool event expectation or log is nil")
	}
	filterer, err := bindaave.NewPoolFilterer(e.pool, nil)
	if err != nil {
		return nil, fmt.Errorf("create Aave Pool event parser: %w", err)
	}
	metadata := defi.EventMetadata{
		Protocol: "aave",
		Emitter:  log.Address,
		LogIndex: log.Index,
	}
	switch e.kind {
	case repayPoolEvent:
		parsed, err := filterer.ParseRepay(*log)
		if err != nil {
			return nil, err
		}
		metadata.Name = "Repay"
		return &RepayEvent{
			Metadata:   metadata,
			Asset:      parsed.Reserve,
			User:       parsed.User,
			Repayer:    parsed.Repayer,
			Amount:     cloneEventBigInt(parsed.Amount),
			UseATokens: parsed.UseATokens,
		}, nil
	case withdrawPoolEvent:
		parsed, err := filterer.ParseWithdraw(*log)
		if err != nil {
			return nil, err
		}
		metadata.Name = "Withdraw"
		return &WithdrawEvent{
			Metadata: metadata,
			Asset:    parsed.Reserve,
			User:     parsed.User,
			To:       parsed.To,
			Amount:   cloneEventBigInt(parsed.Amount),
		}, nil
	default:
		return nil, fmt.Errorf("unsupported Aave Pool event kind %d", e.kind)
	}
}

func (e *poolExitEventExpectation) Match(event defi.DecodedEvent, ctx defi.MatchContext) (defi.MatchResult, error) {
	if e == nil {
		return defi.MatchResult{}, fmt.Errorf("Aave Pool event expectation is nil")
	}
	mismatches := make([]defi.FieldMismatch, 0, 5)
	var amount *big.Int
	switch e.kind {
	case repayPoolEvent:
		repay, ok := event.(*RepayEvent)
		if !ok {
			return defi.MatchResult{}, fmt.Errorf("expected *aave.RepayEvent, got %T", event)
		}
		mismatches = appendAaveAddressMismatch(mismatches, "asset", e.asset, repay.Asset)
		mismatches = appendAaveAddressMismatch(mismatches, "user", e.user, repay.User)
		mismatches = appendAaveAddressMismatch(mismatches, "repayer", e.repayer, repay.Repayer)
		if e.useATokens != repay.UseATokens {
			mismatches = append(mismatches, defi.FieldMismatch{
				Field: "useATokens", Expected: fmt.Sprint(e.useATokens), Actual: fmt.Sprint(repay.UseATokens),
			})
		}
		amount = repay.Amount
	case withdrawPoolEvent:
		withdraw, ok := event.(*WithdrawEvent)
		if !ok {
			return defi.MatchResult{}, fmt.Errorf("expected *aave.WithdrawEvent, got %T", event)
		}
		mismatches = appendAaveAddressMismatch(mismatches, "asset", e.asset, withdraw.Asset)
		mismatches = appendAaveAddressMismatch(mismatches, "user", e.user, withdraw.User)
		mismatches = appendAaveAddressMismatch(mismatches, "to", e.to, withdraw.To)
		amount = withdraw.Amount
	default:
		return defi.MatchResult{}, fmt.Errorf("unsupported Aave Pool event kind %d", e.kind)
	}

	amountMatch, err := defi.MatchAmountConstraints("amount", amount, ctx, e.constraints...)
	if err != nil {
		return defi.MatchResult{}, err
	}
	mismatches = append(mismatches, amountMatch.Mismatches...)
	if len(mismatches) != 0 {
		return defi.MatchResult{Decision: defi.MatchSkip, Mismatches: mismatches}, nil
	}
	return defi.MatchResult{Decision: defi.MatchAccepted}, nil
}
