package erc20

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	defi "github.com/tn606024/defi-simplify"
	binderc20 "github.com/tn606024/defi-simplify/bind/erc20"
)

var (
	approvalEventTopic = crypto.Keccak256Hash([]byte("Approval(address,address,uint256)"))
	transferEventTopic = crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))
)

type eventKind uint8

const (
	approvalEvent eventKind = iota
	transferEvent
)

// ApprovalEvent is the stable SDK representation of an ERC20 Approval event.
type ApprovalEvent struct {
	Metadata defi.EventMetadata
	Token    common.Address
	Owner    common.Address
	Spender  common.Address
	Amount   *big.Int
}

// EventMetadata returns protocol-neutral event identity.
func (e *ApprovalEvent) EventMetadata() defi.EventMetadata {
	return e.Metadata
}

// TransferEvent is the stable SDK representation of an ERC20 Transfer event.
type TransferEvent struct {
	Metadata defi.EventMetadata
	Token    common.Address
	From     common.Address
	To       common.Address
	Amount   *big.Int
}

// EventMetadata returns protocol-neutral event identity.
func (e *TransferEvent) EventMetadata() defi.EventMetadata {
	return e.Metadata
}

type eventExpectation struct {
	kind        eventKind
	token       common.Address
	from        common.Address
	to          common.Address
	constraints []defi.AmountConstraint
}

// ExpectApproval validates an Approval emitted by token for owner and spender.
func ExpectApproval(token, owner, spender common.Address, constraints ...defi.AmountConstraint) defi.EventExpectation {
	return &eventExpectation{
		kind:        approvalEvent,
		token:       token,
		from:        owner,
		to:          spender,
		constraints: append([]defi.AmountConstraint(nil), constraints...),
	}
}

// ExpectTransfer validates a Transfer emitted by token from sender to recipient.
func ExpectTransfer(token, sender, recipient common.Address, constraints ...defi.AmountConstraint) defi.EventExpectation {
	return &eventExpectation{
		kind:        transferEvent,
		token:       token,
		from:        sender,
		to:          recipient,
		constraints: append([]defi.AmountConstraint(nil), constraints...),
	}
}

func (e *eventExpectation) ExpectationName() string {
	if e != nil && e.kind == transferEvent {
		return "erc20.Transfer"
	}
	return "erc20.Approval"
}

func (e *eventExpectation) IsCandidate(log *types.Log) bool {
	if e == nil || log == nil || log.Address != e.token || len(log.Topics) == 0 {
		return false
	}
	switch e.kind {
	case approvalEvent:
		return log.Topics[0] == approvalEventTopic
	case transferEvent:
		return log.Topics[0] == transferEventTopic
	default:
		return false
	}
}

func (e *eventExpectation) Decode(log *types.Log) (defi.DecodedEvent, error) {
	if e == nil || log == nil {
		return nil, fmt.Errorf("ERC20 event expectation or log is nil")
	}
	filterer, err := binderc20.NewErc20Filterer(e.token, nil)
	if err != nil {
		return nil, fmt.Errorf("create ERC20 event parser: %w", err)
	}
	metadata := defi.EventMetadata{
		Protocol: "erc20",
		Emitter:  log.Address,
		LogIndex: log.Index,
	}
	switch e.kind {
	case approvalEvent:
		parsed, err := filterer.ParseApproval(*log)
		if err != nil {
			return nil, err
		}
		metadata.Name = "Approval"
		return &ApprovalEvent{
			Metadata: metadata,
			Token:    log.Address,
			Owner:    parsed.Owner,
			Spender:  parsed.Spender,
			Amount:   copyBigInt(parsed.Value),
		}, nil
	case transferEvent:
		parsed, err := filterer.ParseTransfer(*log)
		if err != nil {
			return nil, err
		}
		metadata.Name = "Transfer"
		return &TransferEvent{
			Metadata: metadata,
			Token:    log.Address,
			From:     parsed.From,
			To:       parsed.To,
			Amount:   copyBigInt(parsed.Value),
		}, nil
	default:
		return nil, fmt.Errorf("unsupported ERC20 event kind %d", e.kind)
	}
}

func (e *eventExpectation) Match(event defi.DecodedEvent, ctx defi.MatchContext) (defi.MatchResult, error) {
	mismatches := make([]defi.FieldMismatch, 0, 3)
	var amount *big.Int
	switch e.kind {
	case approvalEvent:
		approval, ok := event.(*ApprovalEvent)
		if !ok {
			return defi.MatchResult{}, fmt.Errorf("expected *erc20.ApprovalEvent, got %T", event)
		}
		mismatches = appendAddressMismatch(mismatches, "owner", e.from, approval.Owner)
		mismatches = appendAddressMismatch(mismatches, "spender", e.to, approval.Spender)
		amount = approval.Amount
	case transferEvent:
		transfer, ok := event.(*TransferEvent)
		if !ok {
			return defi.MatchResult{}, fmt.Errorf("expected *erc20.TransferEvent, got %T", event)
		}
		mismatches = appendAddressMismatch(mismatches, "from", e.from, transfer.From)
		mismatches = appendAddressMismatch(mismatches, "to", e.to, transfer.To)
		amount = transfer.Amount
	default:
		return defi.MatchResult{}, fmt.Errorf("unsupported ERC20 event kind %d", e.kind)
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

func appendAddressMismatch(mismatches []defi.FieldMismatch, field string, expected, actual common.Address) []defi.FieldMismatch {
	if expected == actual {
		return mismatches
	}
	return append(mismatches, defi.FieldMismatch{
		Field: field, Expected: expected.Hex(), Actual: actual.Hex(),
	})
}

func copyBigInt(value *big.Int) *big.Int {
	if value == nil {
		return nil
	}
	return new(big.Int).Set(value)
}
