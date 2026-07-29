package weth

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	defi "github.com/tn606024/defi-simplify"
	"github.com/tn606024/defi-simplify/weth/bindings"
)

var (
	depositEventTopic    = crypto.Keccak256Hash([]byte("Deposit(address,uint256)"))
	withdrawalEventTopic = crypto.Keccak256Hash([]byte("Withdrawal(address,uint256)"))
)

type eventKind uint8

const (
	depositEvent eventKind = iota
	withdrawalEvent
)

// DepositEvent is the stable SDK representation of a WETH Deposit event.
type DepositEvent struct {
	Metadata defi.EventMetadata
	Contract common.Address
	Account  common.Address
	Amount   *big.Int
}

// EventMetadata returns protocol-neutral event identity.
func (e *DepositEvent) EventMetadata() defi.EventMetadata {
	return e.Metadata
}

// WithdrawalEvent is the stable SDK representation of a WETH Withdrawal event.
type WithdrawalEvent struct {
	Metadata defi.EventMetadata
	Contract common.Address
	Account  common.Address
	Amount   *big.Int
}

// EventMetadata returns protocol-neutral event identity.
func (e *WithdrawalEvent) EventMetadata() defi.EventMetadata {
	return e.Metadata
}

type eventExpectation struct {
	kind        eventKind
	contract    common.Address
	account     common.Address
	constraints []defi.AmountConstraint
}

// ExpectDeposit validates a Deposit emitted by contract for account.
func ExpectDeposit(
	contract common.Address,
	account common.Address,
	constraints ...defi.AmountConstraint,
) defi.EventExpectation {
	return &eventExpectation{
		kind:        depositEvent,
		contract:    contract,
		account:     account,
		constraints: append([]defi.AmountConstraint(nil), constraints...),
	}
}

// ExpectWithdrawal validates a Withdrawal emitted by contract for account.
func ExpectWithdrawal(
	contract common.Address,
	account common.Address,
	constraints ...defi.AmountConstraint,
) defi.EventExpectation {
	return &eventExpectation{
		kind:        withdrawalEvent,
		contract:    contract,
		account:     account,
		constraints: append([]defi.AmountConstraint(nil), constraints...),
	}
}

func (e *eventExpectation) ExpectationName() string {
	if e == nil {
		return ""
	}
	switch e.kind {
	case depositEvent:
		return "weth.Deposit"
	case withdrawalEvent:
		return "weth.Withdrawal"
	default:
		return "weth.<unknown>"
	}
}

func (e *eventExpectation) IsCandidate(log *types.Log) bool {
	if e == nil || log == nil || log.Address != e.contract || len(log.Topics) == 0 {
		return false
	}
	switch e.kind {
	case depositEvent:
		return log.Topics[0] == depositEventTopic
	case withdrawalEvent:
		return log.Topics[0] == withdrawalEventTopic
	default:
		return false
	}
}

func (e *eventExpectation) Decode(log *types.Log) (defi.DecodedEvent, error) {
	if e == nil || log == nil {
		return nil, fmt.Errorf("WETH event expectation or log is nil")
	}
	filterer, err := bindings.NewWETHFilterer(e.contract, nil)
	if err != nil {
		return nil, fmt.Errorf("create WETH event parser: %w", err)
	}
	metadata := defi.EventMetadata{
		Protocol: "weth",
		Emitter:  log.Address,
		LogIndex: log.Index,
	}
	switch e.kind {
	case depositEvent:
		parsed, err := filterer.ParseDeposit(*log)
		if err != nil {
			return nil, err
		}
		metadata.Name = "Deposit"
		return &DepositEvent{
			Metadata: metadata,
			Contract: log.Address,
			Account:  parsed.Dst,
			Amount:   copyBigInt(parsed.Wad),
		}, nil
	case withdrawalEvent:
		parsed, err := filterer.ParseWithdrawal(*log)
		if err != nil {
			return nil, err
		}
		metadata.Name = "Withdrawal"
		return &WithdrawalEvent{
			Metadata: metadata,
			Contract: log.Address,
			Account:  parsed.Src,
			Amount:   copyBigInt(parsed.Wad),
		}, nil
	default:
		return nil, fmt.Errorf("unsupported WETH event kind %d", e.kind)
	}
}

func (e *eventExpectation) Match(
	event defi.DecodedEvent,
	ctx defi.MatchContext,
) (defi.MatchResult, error) {
	if e == nil {
		return defi.MatchResult{}, fmt.Errorf("WETH event expectation is nil")
	}
	mismatches := make([]defi.FieldMismatch, 0, 2)
	var amount *big.Int
	switch e.kind {
	case depositEvent:
		deposit, ok := event.(*DepositEvent)
		if !ok {
			return defi.MatchResult{}, fmt.Errorf("expected *weth.DepositEvent, got %T", event)
		}
		mismatches = appendAddressMismatch(mismatches, "account", e.account, deposit.Account)
		amount = deposit.Amount
	case withdrawalEvent:
		withdrawal, ok := event.(*WithdrawalEvent)
		if !ok {
			return defi.MatchResult{}, fmt.Errorf("expected *weth.WithdrawalEvent, got %T", event)
		}
		mismatches = appendAddressMismatch(mismatches, "account", e.account, withdrawal.Account)
		amount = withdrawal.Amount
	default:
		return defi.MatchResult{}, fmt.Errorf("unsupported WETH event kind %d", e.kind)
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

func appendAddressMismatch(
	mismatches []defi.FieldMismatch,
	field string,
	expected common.Address,
	actual common.Address,
) []defi.FieldMismatch {
	if expected == actual {
		return mismatches
	}
	return append(mismatches, defi.FieldMismatch{
		Field:    field,
		Expected: expected.Hex(),
		Actual:   actual.Hex(),
	})
}

func copyBigInt(value *big.Int) *big.Int {
	if value == nil {
		return nil
	}
	return new(big.Int).Set(value)
}
