package aave

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	bindaave "github.com/tn606024/defi-simplify/bind/aave"
)

var (
	// ErrUserReserveStateRead is returned when an Aave user reserve-state
	// request is invalid or cannot be completed at one block.
	ErrUserReserveStateRead = errors.New("read Aave user reserve state")
)

// UserReserveState is an immutable view of one account's Aave position for one
// resolved reserve at one block.
type UserReserveState struct {
	reserve                  Reserve
	account                  common.Address
	blockNumber              uint64
	blockHash                common.Hash
	currentATokenBalance     *big.Int
	currentStableDebt        *big.Int
	currentVariableDebt      *big.Int
	principalStableDebt      *big.Int
	scaledVariableDebt       *big.Int
	stableBorrowRate         *big.Int
	liquidityRate            *big.Int
	stableRateLastUpdated    *big.Int
	usageAsCollateralEnabled bool
}

// LoadUserReserveState reads one account's position from the resolved
// reserve's AaveProtocolDataProvider at one latest block hash.
func LoadUserReserveState(
	ctx context.Context,
	backend RegistryBackend,
	reserve Reserve,
	account common.Address,
) (*UserReserveState, error) {
	if backend == nil {
		return nil, fmt.Errorf("%w: backend is nil", ErrUserReserveStateRead)
	}
	if err := reserve.Validate(); err != nil {
		return nil, fmt.Errorf("%w: reserve: %w", ErrUserReserveStateRead, err)
	}
	if account == (common.Address{}) {
		return nil, fmt.Errorf("%w: account is zero", ErrUserReserveStateRead)
	}

	header, err := backend.HeaderByNumber(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: resolve block: %w", ErrUserReserveStateRead, err)
	}
	if header == nil || header.Number == nil || !header.Number.IsUint64() {
		return nil, fmt.Errorf("%w: header has no uint64 block number", ErrUserReserveStateRead)
	}
	block := registryBlock{Number: new(big.Int).Set(header.Number), Hash: header.Hash()}
	if block.Hash == (common.Hash{}) {
		return nil, fmt.Errorf("%w: block hash is zero", ErrUserReserveStateRead)
	}

	dataProvider, err := bindaave.NewAaveProtocolDataProviderCaller(
		reserve.Market().ProtocolDataProvider(),
		backend,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: bind protocol data provider: %w", ErrUserReserveStateRead, err)
	}
	data, err := dataProvider.GetUserReserveData(
		registryCallOpts(ctx, block),
		reserve.Underlying().Address(),
		account,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: read reserve data: %w", ErrUserReserveStateRead, err)
	}
	if err := validateUserReserveAmounts(
		data.CurrentATokenBalance,
		data.CurrentStableDebt,
		data.CurrentVariableDebt,
		data.PrincipalStableDebt,
		data.ScaledVariableDebt,
		data.StableBorrowRate,
		data.LiquidityRate,
		data.StableRateLastUpdated,
	); err != nil {
		return nil, err
	}

	return &UserReserveState{
		reserve:                  reserve,
		account:                  account,
		blockNumber:              block.Number.Uint64(),
		blockHash:                block.Hash,
		currentATokenBalance:     cloneUserReserveAmount(data.CurrentATokenBalance),
		currentStableDebt:        cloneUserReserveAmount(data.CurrentStableDebt),
		currentVariableDebt:      cloneUserReserveAmount(data.CurrentVariableDebt),
		principalStableDebt:      cloneUserReserveAmount(data.PrincipalStableDebt),
		scaledVariableDebt:       cloneUserReserveAmount(data.ScaledVariableDebt),
		stableBorrowRate:         cloneUserReserveAmount(data.StableBorrowRate),
		liquidityRate:            cloneUserReserveAmount(data.LiquidityRate),
		stableRateLastUpdated:    cloneUserReserveAmount(data.StableRateLastUpdated),
		usageAsCollateralEnabled: data.UsageAsCollateralEnabled,
	}, nil
}

// Reserve returns the resolved reserve observed by this state read.
func (s *UserReserveState) Reserve() Reserve {
	return s.reserve
}

// Account returns the Aave position owner whose state was read.
func (s *UserReserveState) Account() common.Address {
	return s.account
}

// BlockNumber returns the block number shared by every value in this state.
func (s *UserReserveState) BlockNumber() uint64 {
	return s.blockNumber
}

// BlockHash returns the block hash shared by every value in this state.
func (s *UserReserveState) BlockHash() common.Hash {
	return s.blockHash
}

// CurrentATokenBalance returns the current aToken balance in base units.
func (s *UserReserveState) CurrentATokenBalance() *big.Int {
	return cloneUserReserveAmount(s.currentATokenBalance)
}

// CurrentStableDebt returns the current stable-rate debt in base units.
func (s *UserReserveState) CurrentStableDebt() *big.Int {
	return cloneUserReserveAmount(s.currentStableDebt)
}

// CurrentVariableDebt returns the current variable-rate debt in base units.
func (s *UserReserveState) CurrentVariableDebt() *big.Int {
	return cloneUserReserveAmount(s.currentVariableDebt)
}

// PrincipalStableDebt returns the stable-rate principal debt in base units.
func (s *UserReserveState) PrincipalStableDebt() *big.Int {
	return cloneUserReserveAmount(s.principalStableDebt)
}

// ScaledVariableDebt returns the scaled variable-rate debt in base units.
func (s *UserReserveState) ScaledVariableDebt() *big.Int {
	return cloneUserReserveAmount(s.scaledVariableDebt)
}

// StableBorrowRate returns the user's stable borrow rate.
func (s *UserReserveState) StableBorrowRate() *big.Int {
	return cloneUserReserveAmount(s.stableBorrowRate)
}

// LiquidityRate returns the reserve liquidity rate observed for this user.
func (s *UserReserveState) LiquidityRate() *big.Int {
	return cloneUserReserveAmount(s.liquidityRate)
}

// StableRateLastUpdated returns the user's stable-rate update timestamp.
func (s *UserReserveState) StableRateLastUpdated() *big.Int {
	return cloneUserReserveAmount(s.stableRateLastUpdated)
}

// UsageAsCollateralEnabled reports whether the reserve is enabled as
// collateral for this account.
func (s *UserReserveState) UsageAsCollateralEnabled() bool {
	return s.usageAsCollateralEnabled
}

func validateUserReserveAmounts(values ...*big.Int) error {
	for _, value := range values {
		if value == nil {
			return fmt.Errorf("%w: protocol data provider returned nil amount", ErrUserReserveStateRead)
		}
	}
	return nil
}

func cloneUserReserveAmount(value *big.Int) *big.Int {
	if value == nil {
		return nil
	}
	return new(big.Int).Set(value)
}
