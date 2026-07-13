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

const (
	// VariableInterestRateMode is Aave V3's variable-rate borrowing mode.
	VariableInterestRateMode uint8 = 2
)

var (
	supplyEventTopic = crypto.Keccak256Hash([]byte("Supply(address,address,address,uint256,uint16)"))
	borrowEventTopic = crypto.Keccak256Hash([]byte("Borrow(address,address,address,uint256,uint8,uint256,uint16)"))
)

type protocolEventKind uint8

const (
	supplyProtocolEvent protocolEventKind = iota
	borrowProtocolEvent
)

// SupplyEvent is the stable SDK representation of an Aave V3 Supply event.
type SupplyEvent struct {
	Metadata     defi.EventMetadata
	Asset        common.Address
	User         common.Address
	OnBehalfOf   common.Address
	Amount       *big.Int
	ReferralCode uint16
}

// EventMetadata returns protocol-neutral event identity.
func (e *SupplyEvent) EventMetadata() defi.EventMetadata {
	return e.Metadata
}

// BorrowEvent is the stable SDK representation of an Aave V3 Borrow event.
type BorrowEvent struct {
	Metadata         defi.EventMetadata
	Asset            common.Address
	User             common.Address
	OnBehalfOf       common.Address
	Amount           *big.Int
	InterestRateMode uint8
	BorrowRate       *big.Int
	ReferralCode     uint16
}

// EventMetadata returns protocol-neutral event identity.
func (e *BorrowEvent) EventMetadata() defi.EventMetadata {
	return e.Metadata
}

type protocolEventExpectation struct {
	kind             protocolEventKind
	pool             common.Address
	asset            common.Address
	user             common.Address
	onBehalfOf       common.Address
	interestRateMode uint8
	constraints      []defi.AmountConstraint
}

// ExpectSupply validates an Aave Supply event emitted by pool.
func ExpectSupply(pool, asset, user, onBehalfOf common.Address, constraints ...defi.AmountConstraint) defi.EventExpectation {
	return &protocolEventExpectation{
		kind:        supplyProtocolEvent,
		pool:        pool,
		asset:       asset,
		user:        user,
		onBehalfOf:  onBehalfOf,
		constraints: append([]defi.AmountConstraint(nil), constraints...),
	}
}

// ExpectBorrow validates an Aave Borrow event emitted by pool.
func ExpectBorrow(
	pool,
	asset,
	user,
	onBehalfOf common.Address,
	interestRateMode uint8,
	constraints ...defi.AmountConstraint,
) defi.EventExpectation {
	return &protocolEventExpectation{
		kind:             borrowProtocolEvent,
		pool:             pool,
		asset:            asset,
		user:             user,
		onBehalfOf:       onBehalfOf,
		interestRateMode: interestRateMode,
		constraints:      append([]defi.AmountConstraint(nil), constraints...),
	}
}

func (e *protocolEventExpectation) ExpectationName() string {
	if e == nil {
		return ""
	}
	switch e.kind {
	case supplyProtocolEvent:
		return "aave.Supply"
	case borrowProtocolEvent:
		return "aave.Borrow"
	default:
		return "aave.<unknown>"
	}
}

func (e *protocolEventExpectation) IsCandidate(log *types.Log) bool {
	if e == nil || log == nil || log.Address != e.pool || len(log.Topics) == 0 {
		return false
	}
	switch e.kind {
	case supplyProtocolEvent:
		return log.Topics[0] == supplyEventTopic
	case borrowProtocolEvent:
		return log.Topics[0] == borrowEventTopic
	default:
		return false
	}
}

func (e *protocolEventExpectation) Decode(log *types.Log) (defi.DecodedEvent, error) {
	if e == nil || log == nil {
		return nil, fmt.Errorf("Aave event expectation or log is nil")
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
	case supplyProtocolEvent:
		parsed, err := filterer.ParseSupply(*log)
		if err != nil {
			return nil, err
		}
		metadata.Name = "Supply"
		return &SupplyEvent{
			Metadata:     metadata,
			Asset:        parsed.Reserve,
			User:         parsed.User,
			OnBehalfOf:   parsed.OnBehalfOf,
			Amount:       cloneEventBigInt(parsed.Amount),
			ReferralCode: parsed.ReferralCode,
		}, nil
	case borrowProtocolEvent:
		parsed, err := filterer.ParseBorrow(*log)
		if err != nil {
			return nil, err
		}
		metadata.Name = "Borrow"
		return &BorrowEvent{
			Metadata:         metadata,
			Asset:            parsed.Reserve,
			User:             parsed.User,
			OnBehalfOf:       parsed.OnBehalfOf,
			Amount:           cloneEventBigInt(parsed.Amount),
			InterestRateMode: parsed.InterestRateMode,
			BorrowRate:       cloneEventBigInt(parsed.BorrowRate),
			ReferralCode:     parsed.ReferralCode,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported Aave event kind %d", e.kind)
	}
}

func (e *protocolEventExpectation) Match(event defi.DecodedEvent, ctx defi.MatchContext) (defi.MatchResult, error) {
	if e == nil {
		return defi.MatchResult{}, fmt.Errorf("Aave event expectation is nil")
	}
	mismatches := make([]defi.FieldMismatch, 0, 5)
	var amount *big.Int
	switch e.kind {
	case supplyProtocolEvent:
		supply, ok := event.(*SupplyEvent)
		if !ok {
			return defi.MatchResult{}, fmt.Errorf("expected *aave.SupplyEvent, got %T", event)
		}
		mismatches = appendAaveAddressMismatch(mismatches, "asset", e.asset, supply.Asset)
		mismatches = appendAaveAddressMismatch(mismatches, "user", e.user, supply.User)
		mismatches = appendAaveAddressMismatch(mismatches, "onBehalfOf", e.onBehalfOf, supply.OnBehalfOf)
		amount = supply.Amount
	case borrowProtocolEvent:
		borrow, ok := event.(*BorrowEvent)
		if !ok {
			return defi.MatchResult{}, fmt.Errorf("expected *aave.BorrowEvent, got %T", event)
		}
		mismatches = appendAaveAddressMismatch(mismatches, "asset", e.asset, borrow.Asset)
		mismatches = appendAaveAddressMismatch(mismatches, "user", e.user, borrow.User)
		mismatches = appendAaveAddressMismatch(mismatches, "onBehalfOf", e.onBehalfOf, borrow.OnBehalfOf)
		if e.interestRateMode != borrow.InterestRateMode {
			mismatches = append(mismatches, defi.FieldMismatch{
				Field:    "interestRateMode",
				Expected: fmt.Sprint(e.interestRateMode),
				Actual:   fmt.Sprint(borrow.InterestRateMode),
			})
		}
		amount = borrow.Amount
	default:
		return defi.MatchResult{}, fmt.Errorf("unsupported Aave event kind %d", e.kind)
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

func appendAaveAddressMismatch(mismatches []defi.FieldMismatch, field string, expected, actual common.Address) []defi.FieldMismatch {
	if expected == actual {
		return mismatches
	}
	return append(mismatches, defi.FieldMismatch{
		Field: field, Expected: expected.Hex(), Actual: actual.Hex(),
	})
}

func cloneEventBigInt(value *big.Int) *big.Int {
	if value == nil {
		return nil
	}
	return new(big.Int).Set(value)
}
