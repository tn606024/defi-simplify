package main

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	"github.com/tn606024/defi-simplify/aave"
	"github.com/tn606024/defi-simplify/client/account/eip7702"
	"github.com/tn606024/defi-simplify/erc20"
)

type accountSnapshot struct {
	delegation    eip7702.DelegationState
	nativeBalance *big.Int
	wethDecimals  uint8
	usdcDecimals  uint8
	wethToken     tokenStateSnapshot
	usdcToken     tokenStateSnapshot
	wethPosition  reserveStateSnapshot
	usdcPosition  reserveStateSnapshot
}

type tokenStateSnapshot struct {
	balance     *big.Int
	allowance   *big.Int
	blockNumber uint64
	blockHash   common.Hash
}

type reserveStateSnapshot struct {
	aTokenBalance     *big.Int
	stableDebt        *big.Int
	variableDebt      *big.Int
	collateralEnabled bool
	blockNumber       uint64
	blockHash         common.Hash
}

func (app *application) inspectAccount(ctx context.Context) (*accountSnapshot, error) {
	delegation, err := eip7702.ReadPendingDelegationState(ctx, app.client, app.account)
	if err != nil {
		return nil, fmt.Errorf("read pending delegation: %w", err)
	}
	nativeBalance, err := app.client.BalanceAt(ctx, app.account, nil)
	if err != nil {
		return nil, fmt.Errorf("read native balance: %w", err)
	}
	wethToken, err := erc20.LoadAccountState(
		ctx,
		app.client,
		app.wethReserve.Underlying(),
		app.account,
		app.market.Pool(),
	)
	if err != nil {
		return nil, fmt.Errorf("read WETH account state: %w", err)
	}
	usdcToken, err := erc20.LoadAccountState(
		ctx,
		app.client,
		app.usdcReserve.Underlying(),
		app.account,
		app.market.Pool(),
	)
	if err != nil {
		return nil, fmt.Errorf("read USDC account state: %w", err)
	}
	wethPosition, err := aave.LoadUserReserveState(ctx, app.client, app.wethReserve, app.account)
	if err != nil {
		return nil, fmt.Errorf("read WETH Aave position: %w", err)
	}
	usdcPosition, err := aave.LoadUserReserveState(ctx, app.client, app.usdcReserve, app.account)
	if err != nil {
		return nil, fmt.Errorf("read USDC Aave position: %w", err)
	}
	return &accountSnapshot{
		delegation:    delegation,
		nativeBalance: new(big.Int).Set(nativeBalance),
		wethDecimals:  app.wethReserve.Underlying().Decimals(),
		usdcDecimals:  app.usdcReserve.Underlying().Decimals(),
		wethToken:     snapshotTokenState(wethToken),
		usdcToken:     snapshotTokenState(usdcToken),
		wethPosition:  snapshotReserveState(wethPosition),
		usdcPosition:  snapshotReserveState(usdcPosition),
	}, nil
}

func snapshotTokenState(state *erc20.AccountState) tokenStateSnapshot {
	return tokenStateSnapshot{
		balance:     state.Balance(),
		allowance:   state.Allowance(),
		blockNumber: state.BlockNumber(),
		blockHash:   state.BlockHash(),
	}
}

func snapshotReserveState(state *aave.UserReserveState) reserveStateSnapshot {
	return reserveStateSnapshot{
		aTokenBalance:     state.CurrentATokenBalance(),
		stableDebt:        state.CurrentStableDebt(),
		variableDebt:      state.CurrentVariableDebt(),
		collateralEnabled: state.UsageAsCollateralEnabled(),
		blockNumber:       state.BlockNumber(),
		blockHash:         state.BlockHash(),
	}
}

func validatePhasePreconditions(
	config commandConfig,
	state *accountSnapshot,
	implementation common.Address,
) error {
	if state == nil {
		return errors.New("account state is nil")
	}
	switch config.phase {
	case phaseInspect:
		return nil
	case phaseDelegate:
		if !state.delegation.Clean() {
			return fmt.Errorf(
				"delegation requires a clean EOA, got %s delegated to %s",
				state.delegation.Status,
				state.delegation.Implementation.Hex(),
			)
		}
		return nil
	case phaseOpen:
		if err := requireExpectedDelegation(state.delegation, implementation); err != nil {
			return err
		}
		if reserveStateHasPosition(state.wethPosition) ||
			reserveStateHasPosition(state.usdcPosition) {
			return errors.New("open requires zero WETH and USDC Aave positions")
		}
		needed := decimalToBaseUnits(config.wethSupplyAmount, state.wethDecimals)
		if state.nativeBalance.Cmp(needed) < 0 {
			return fmt.Errorf("native balance %s is below WETH supply amount %s", state.nativeBalance, needed)
		}
		return nil
	case phaseClose:
		if err := requireExpectedDelegation(state.delegation, implementation); err != nil {
			return err
		}
		if state.wethPosition.aTokenBalance.Sign() == 0 ||
			state.usdcPosition.variableDebt.Sign() == 0 {
			return errors.New("close requires WETH collateral and USDC variable debt")
		}
		if state.wethPosition.stableDebt.Sign() != 0 ||
			state.wethPosition.variableDebt.Sign() != 0 ||
			state.usdcPosition.aTokenBalance.Sign() != 0 ||
			state.usdcPosition.stableDebt.Sign() != 0 {
			return errors.New("close requires an isolated WETH-collateral and USDC-variable-debt position")
		}
		limit := decimalToBaseUnits(config.usdcRepayLimit, state.usdcDecimals)
		if limit.Cmp(state.usdcPosition.variableDebt) < 0 {
			return fmt.Errorf(
				"USDC repay limit %s is below current variable debt %s",
				limit,
				state.usdcPosition.variableDebt,
			)
		}
		if state.usdcToken.balance.Cmp(limit) < 0 {
			return fmt.Errorf(
				"USDC balance %s cannot cover configured repay limit %s",
				state.usdcToken.balance,
				limit,
			)
		}
		return nil
	case phaseClear:
		if err := requireExpectedDelegation(state.delegation, implementation); err != nil {
			return err
		}
		if reserveStateHasPosition(state.wethPosition) ||
			reserveStateHasPosition(state.usdcPosition) {
			return errors.New("clear requires zero targeted collateral and debt")
		}
		if state.wethToken.allowance.Sign() != 0 || state.usdcToken.allowance.Sign() != 0 {
			return errors.New("clear requires zero Aave Pool allowance")
		}
		return nil
	default:
		return fmt.Errorf("unsupported lifecycle phase %q", config.phase)
	}
}

func requireExpectedDelegation(
	state eip7702.DelegationState,
	implementation common.Address,
) error {
	if !state.Delegated() || state.Implementation != implementation {
		return fmt.Errorf(
			"EOA must delegate to reviewed implementation %s, got %s delegated to %s",
			implementation.Hex(),
			state.Status,
			state.Implementation.Hex(),
		)
	}
	return nil
}

func decimalToBaseUnits(value decimal.Decimal, decimals uint8) *big.Int {
	return value.Shift(int32(decimals)).BigInt()
}

func reserveStateHasPosition(state reserveStateSnapshot) bool {
	return state.aTokenBalance.Sign() != 0 ||
		state.stableDebt.Sign() != 0 ||
		state.variableDebt.Sign() != 0
}
