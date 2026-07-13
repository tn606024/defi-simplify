package aave

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/shopspring/decimal"
	bindaave "github.com/tn606024/defi-simplify/bind/aave"
	binderc20 "github.com/tn606024/defi-simplify/bind/erc20"
	"github.com/tn606024/defi-simplify/config"
	"github.com/tn606024/defi-simplify/helper"
)

var (
	// ErrInvalidExecutionExpectation is returned when expected Aave flow fields are incomplete or invalid.
	ErrInvalidExecutionExpectation = errors.New("invalid Aave execution expectation")
	// ErrInvalidExecutionReceipt is returned when a receipt is nil, incomplete, or unsuccessful.
	ErrInvalidExecutionReceipt = errors.New("invalid Aave execution receipt")
	// ErrExpectedEventNotFound is returned when a receipt lacks an event matching the expected flow.
	ErrExpectedEventNotFound = errors.New("expected Aave execution event not found")
	// ErrMalformedExecutionEvent is returned when an expected contract emits a log that cannot be decoded.
	ErrMalformedExecutionEvent = errors.New("malformed Aave execution event")
)

const variableInterestRateMode uint8 = VariableInterestRateMode

// ExecutionExpectation describes the exact Phase 1 approve, supply, and borrow result to validate.
type ExecutionExpectation struct {
	Account      common.Address
	Pool         common.Address
	SupplyAsset  common.Address
	SupplyAmount *big.Int
	BorrowAsset  common.Address
	BorrowAmount *big.Int
}

// ApprovalResult describes the ERC20 approval used by the Aave supply call.
type ApprovalResult struct {
	Token    common.Address
	Owner    common.Address
	Spender  common.Address
	Amount   *big.Int
	LogIndex uint
}

// SupplyResult describes the validated Aave Supply event.
type SupplyResult struct {
	Asset        common.Address
	User         common.Address
	OnBehalfOf   common.Address
	Amount       *big.Int
	ReferralCode uint16
	LogIndex     uint
}

// BorrowResult describes the validated Aave Borrow event.
type BorrowResult struct {
	Asset            common.Address
	User             common.Address
	OnBehalfOf       common.Address
	Amount           *big.Int
	InterestRateMode uint8
	BorrowRate       *big.Int
	ReferralCode     uint16
	LogIndex         uint
}

// ExecutionSummary is a stable SDK representation of a validated Aave Flow receipt.
type ExecutionSummary struct {
	TransactionHash common.Hash
	BlockHash       common.Hash
	BlockNumber     *big.Int
	Approval        ApprovalResult
	Supply          SupplyResult
	Borrow          BorrowResult
}

// NewExecutionExpectation resolves chain configuration and converts decimal amounts to raw token units.
func NewExecutionExpectation(
	chain config.Chain,
	account common.Address,
	supplyCoin config.Coin,
	supplyAmount decimal.Decimal,
	borrowCoin config.Coin,
	borrowAmount decimal.Decimal,
) (*ExecutionExpectation, error) {
	if !supplyAmount.IsPositive() {
		return nil, fmt.Errorf("%w: supply amount must be positive", ErrInvalidExecutionExpectation)
	}
	if !borrowAmount.IsPositive() {
		return nil, fmt.Errorf("%w: borrow amount must be positive", ErrInvalidExecutionExpectation)
	}

	pool, err := chain.AaveV3PoolAddress()
	if err != nil {
		return nil, fmt.Errorf("%w: resolve Aave Pool: %v", ErrInvalidExecutionExpectation, err)
	}
	supplyAsset, err := supplyCoin.Address(chain)
	if err != nil {
		return nil, fmt.Errorf("%w: resolve supply asset: %v", ErrInvalidExecutionExpectation, err)
	}
	borrowAsset, err := borrowCoin.Address(chain)
	if err != nil {
		return nil, fmt.Errorf("%w: resolve borrow asset: %v", ErrInvalidExecutionExpectation, err)
	}
	supplyDecimals, err := supplyCoin.Decimals()
	if err != nil {
		return nil, fmt.Errorf("%w: resolve supply asset decimals: %v", ErrInvalidExecutionExpectation, err)
	}
	borrowDecimals, err := borrowCoin.Decimals()
	if err != nil {
		return nil, fmt.Errorf("%w: resolve borrow asset decimals: %v", ErrInvalidExecutionExpectation, err)
	}

	expected := &ExecutionExpectation{
		Account:      account,
		Pool:         pool,
		SupplyAsset:  supplyAsset,
		SupplyAmount: helper.ToWei(supplyAmount, supplyDecimals),
		BorrowAsset:  borrowAsset,
		BorrowAmount: helper.ToWei(borrowAmount, borrowDecimals),
	}
	if err := expected.validate(); err != nil {
		return nil, err
	}
	return expected, nil
}

// ParseExecutionReceipt decodes and validates the Approval, Supply, and Borrow events for a Phase 1 flow.
// Only logs emitted by the expected supply token and Aave Pool are considered.
func ParseExecutionReceipt(receipt *types.Receipt, expected *ExecutionExpectation) (*ExecutionSummary, error) {
	if receipt == nil {
		return nil, fmt.Errorf("%w: receipt is nil", ErrInvalidExecutionReceipt)
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return nil, fmt.Errorf("%w: transaction %s has status %d", ErrInvalidExecutionReceipt, receipt.TxHash.Hex(), receipt.Status)
	}
	if receipt.BlockNumber == nil {
		return nil, fmt.Errorf("%w: transaction %s has no block number", ErrInvalidExecutionReceipt, receipt.TxHash.Hex())
	}
	if expected == nil {
		return nil, fmt.Errorf("%w: expectation is nil", ErrInvalidExecutionExpectation)
	}
	if err := expected.validate(); err != nil {
		return nil, err
	}

	poolABI, err := bindaave.PoolMetaData.GetAbi()
	if err != nil {
		return nil, fmt.Errorf("load Aave Pool ABI: %w", err)
	}
	tokenABI, err := binderc20.Erc20MetaData.GetAbi()
	if err != nil {
		return nil, fmt.Errorf("load ERC20 ABI: %w", err)
	}
	poolFilterer, err := bindaave.NewPoolFilterer(expected.Pool, nil)
	if err != nil {
		return nil, fmt.Errorf("create Aave Pool event parser: %w", err)
	}
	tokenFilterer, err := binderc20.NewErc20Filterer(expected.SupplyAsset, nil)
	if err != nil {
		return nil, fmt.Errorf("create ERC20 event parser: %w", err)
	}

	approvalTopic := tokenABI.Events["Approval"].ID
	supplyTopic := poolABI.Events["Supply"].ID
	borrowTopic := poolABI.Events["Borrow"].ID

	var (
		approval           *ApprovalResult
		supply             *SupplyResult
		borrow             *BorrowResult
		approvalCandidates int
		supplyCandidates   int
		borrowCandidates   int
	)
	for _, receiptLog := range receipt.Logs {
		if receiptLog == nil || len(receiptLog.Topics) == 0 {
			continue
		}

		switch {
		case receiptLog.Address == expected.SupplyAsset && receiptLog.Topics[0] == approvalTopic:
			approvalCandidates++
			event, err := tokenFilterer.ParseApproval(*receiptLog)
			if err != nil {
				return nil, fmt.Errorf("%w: decode Approval log %d: %v", ErrMalformedExecutionEvent, receiptLog.Index, err)
			}
			if approval == nil && event.Owner == expected.Account && event.Spender == expected.Pool && equalAmount(event.Value, expected.SupplyAmount) {
				approval = &ApprovalResult{
					Token:    receiptLog.Address,
					Owner:    event.Owner,
					Spender:  event.Spender,
					Amount:   cloneBigInt(event.Value),
					LogIndex: receiptLog.Index,
				}
			}

		case receiptLog.Address == expected.Pool && receiptLog.Topics[0] == supplyTopic:
			supplyCandidates++
			event, err := poolFilterer.ParseSupply(*receiptLog)
			if err != nil {
				return nil, fmt.Errorf("%w: decode Supply log %d: %v", ErrMalformedExecutionEvent, receiptLog.Index, err)
			}
			if supply == nil && event.Reserve == expected.SupplyAsset && event.User == expected.Account && event.OnBehalfOf == expected.Account && equalAmount(event.Amount, expected.SupplyAmount) {
				supply = &SupplyResult{
					Asset:        event.Reserve,
					User:         event.User,
					OnBehalfOf:   event.OnBehalfOf,
					Amount:       cloneBigInt(event.Amount),
					ReferralCode: event.ReferralCode,
					LogIndex:     receiptLog.Index,
				}
			}

		case receiptLog.Address == expected.Pool && receiptLog.Topics[0] == borrowTopic:
			borrowCandidates++
			event, err := poolFilterer.ParseBorrow(*receiptLog)
			if err != nil {
				return nil, fmt.Errorf("%w: decode Borrow log %d: %v", ErrMalformedExecutionEvent, receiptLog.Index, err)
			}
			if borrow == nil && event.Reserve == expected.BorrowAsset && event.User == expected.Account && event.OnBehalfOf == expected.Account && equalAmount(event.Amount, expected.BorrowAmount) && event.InterestRateMode == variableInterestRateMode {
				borrow = &BorrowResult{
					Asset:            event.Reserve,
					User:             event.User,
					OnBehalfOf:       event.OnBehalfOf,
					Amount:           cloneBigInt(event.Amount),
					InterestRateMode: event.InterestRateMode,
					BorrowRate:       cloneBigInt(event.BorrowRate),
					ReferralCode:     event.ReferralCode,
					LogIndex:         receiptLog.Index,
				}
			}
		}
	}

	if approval == nil {
		return nil, expectedEventError("Approval", approvalCandidates, expected)
	}
	if supply == nil {
		return nil, expectedEventError("Supply", supplyCandidates, expected)
	}
	if borrow == nil {
		return nil, expectedEventError("Borrow", borrowCandidates, expected)
	}

	return &ExecutionSummary{
		TransactionHash: receipt.TxHash,
		BlockHash:       receipt.BlockHash,
		BlockNumber:     cloneBigInt(receipt.BlockNumber),
		Approval:        *approval,
		Supply:          *supply,
		Borrow:          *borrow,
	}, nil
}

func (e *ExecutionExpectation) validate() error {
	if e == nil {
		return fmt.Errorf("%w: expectation is nil", ErrInvalidExecutionExpectation)
	}
	if e.Account == (common.Address{}) {
		return fmt.Errorf("%w: account is zero", ErrInvalidExecutionExpectation)
	}
	if e.Pool == (common.Address{}) {
		return fmt.Errorf("%w: Aave Pool is zero", ErrInvalidExecutionExpectation)
	}
	if e.SupplyAsset == (common.Address{}) {
		return fmt.Errorf("%w: supply asset is zero", ErrInvalidExecutionExpectation)
	}
	if e.SupplyAmount == nil || e.SupplyAmount.Sign() <= 0 {
		return fmt.Errorf("%w: supply amount must be positive", ErrInvalidExecutionExpectation)
	}
	if e.BorrowAsset == (common.Address{}) {
		return fmt.Errorf("%w: borrow asset is zero", ErrInvalidExecutionExpectation)
	}
	if e.BorrowAmount == nil || e.BorrowAmount.Sign() <= 0 {
		return fmt.Errorf("%w: borrow amount must be positive", ErrInvalidExecutionExpectation)
	}
	return nil
}

func expectedEventError(name string, candidates int, expected *ExecutionExpectation) error {
	return fmt.Errorf(
		"%w: %s matched 0 of %d candidate logs for account %s, supply %s of %s, borrow %s of %s",
		ErrExpectedEventNotFound,
		name,
		candidates,
		expected.Account.Hex(),
		expected.SupplyAmount.String(),
		expected.SupplyAsset.Hex(),
		expected.BorrowAmount.String(),
		expected.BorrowAsset.Hex(),
	)
}

func equalAmount(actual, expected *big.Int) bool {
	return actual != nil && expected != nil && actual.Cmp(expected) == 0
}

func cloneBigInt(value *big.Int) *big.Int {
	if value == nil {
		return nil
	}
	return new(big.Int).Set(value)
}
