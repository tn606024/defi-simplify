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

var borrowAllowanceDelegatedEventTopic = crypto.Keccak256Hash(
	[]byte("BorrowAllowanceDelegated(address,address,address,uint256)"),
)

// BorrowAllowanceDelegatedEvent is the stable SDK representation of an Aave credit-delegation event.
type BorrowAllowanceDelegatedEvent struct {
	Metadata  defi.EventMetadata
	DebtToken common.Address
	FromUser  common.Address
	ToUser    common.Address
	Asset     common.Address
	Amount    *big.Int
}

// EventMetadata returns protocol-neutral event identity.
func (e *BorrowAllowanceDelegatedEvent) EventMetadata() defi.EventMetadata {
	return e.Metadata
}

type delegationEventExpectation struct {
	debtToken   common.Address
	asset       common.Address
	fromUser    common.Address
	toUser      common.Address
	constraints []defi.AmountConstraint
}

// ExpectBorrowAllowanceDelegated validates an event emitted by an Aave variable debt token.
func ExpectBorrowAllowanceDelegated(
	debtToken,
	asset,
	fromUser,
	toUser common.Address,
	constraints ...defi.AmountConstraint,
) defi.EventExpectation {
	return &delegationEventExpectation{
		debtToken:   debtToken,
		asset:       asset,
		fromUser:    fromUser,
		toUser:      toUser,
		constraints: append([]defi.AmountConstraint(nil), constraints...),
	}
}

func (e *delegationEventExpectation) ExpectationName() string {
	if e == nil {
		return ""
	}
	return "aave.BorrowAllowanceDelegated"
}

func (e *delegationEventExpectation) IsCandidate(log *types.Log) bool {
	return e != nil &&
		log != nil &&
		log.Address == e.debtToken &&
		len(log.Topics) != 0 &&
		log.Topics[0] == borrowAllowanceDelegatedEventTopic
}

func (e *delegationEventExpectation) Decode(log *types.Log) (defi.DecodedEvent, error) {
	if e == nil || log == nil {
		return nil, fmt.Errorf("Aave delegation event expectation or log is nil")
	}
	filterer, err := bindaave.NewDebtTokenBaseFilterer(e.debtToken, nil)
	if err != nil {
		return nil, fmt.Errorf("create Aave debt-token event parser: %w", err)
	}
	parsed, err := filterer.ParseBorrowAllowanceDelegated(*log)
	if err != nil {
		return nil, err
	}
	return &BorrowAllowanceDelegatedEvent{
		Metadata: defi.EventMetadata{
			Protocol: "aave",
			Name:     "BorrowAllowanceDelegated",
			Emitter:  log.Address,
			LogIndex: log.Index,
		},
		DebtToken: log.Address,
		FromUser:  parsed.FromUser,
		ToUser:    parsed.ToUser,
		Asset:     parsed.Asset,
		Amount:    cloneEventBigInt(parsed.Amount),
	}, nil
}

func (e *delegationEventExpectation) Match(event defi.DecodedEvent, ctx defi.MatchContext) (defi.MatchResult, error) {
	if e == nil {
		return defi.MatchResult{}, fmt.Errorf("Aave delegation event expectation is nil")
	}
	delegation, ok := event.(*BorrowAllowanceDelegatedEvent)
	if !ok {
		return defi.MatchResult{}, fmt.Errorf("expected *aave.BorrowAllowanceDelegatedEvent, got %T", event)
	}
	mismatches := make([]defi.FieldMismatch, 0, 5)
	mismatches = appendAaveAddressMismatch(mismatches, "debtToken", e.debtToken, delegation.DebtToken)
	mismatches = appendAaveAddressMismatch(mismatches, "asset", e.asset, delegation.Asset)
	mismatches = appendAaveAddressMismatch(mismatches, "fromUser", e.fromUser, delegation.FromUser)
	mismatches = appendAaveAddressMismatch(mismatches, "toUser", e.toUser, delegation.ToUser)

	amountMatch, err := defi.MatchAmountConstraints("amount", delegation.Amount, ctx, e.constraints...)
	if err != nil {
		return defi.MatchResult{}, err
	}
	mismatches = append(mismatches, amountMatch.Mismatches...)
	if len(mismatches) != 0 {
		return defi.MatchResult{Decision: defi.MatchSkip, Mismatches: mismatches}, nil
	}
	return defi.MatchResult{Decision: defi.MatchAccepted}, nil
}
