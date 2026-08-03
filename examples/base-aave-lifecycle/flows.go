package main

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	defi "github.com/tn606024/defi-simplify"
	"github.com/tn606024/defi-simplify/aave"
	txamount "github.com/tn606024/defi-simplify/amount"
	"github.com/tn606024/defi-simplify/assets/base"
	"github.com/tn606024/defi-simplify/config"
	"github.com/tn606024/defi-simplify/erc20"
	"github.com/tn606024/defi-simplify/weth"
)

func buildOpenFlow(
	accountAddress common.Address,
	market aave.Market,
	wethReserve aave.Reserve,
	usdcReserve aave.Reserve,
	wethSupply decimal.Decimal,
	usdcBorrow decimal.Decimal,
) *defi.Flow {
	wethAmount := txamount.Exact(wethSupply)
	return defi.NewFlow(accountAddress, defi.WithChain(config.Base)).
		Add(weth.Wrap(base.WETH, wethAmount)).
		Add(erc20.Approve(
			wethReserve.Underlying(),
			aave.PoolSpender(market),
			wethAmount,
		)).
		Add(aave.Supply(wethReserve, wethAmount)).
		Add(aave.Borrow(usdcReserve, txamount.Exact(usdcBorrow)))
}

func buildCloseFlow(
	accountAddress common.Address,
	market aave.Market,
	wethReserve aave.Reserve,
	usdcReserve aave.Reserve,
	usdcRepayLimit decimal.Decimal,
) *defi.Flow {
	checkpoint := txamount.Checkpoint("before-close-withdraw", base.WETH)
	return defi.NewFlow(accountAddress, defi.WithChain(config.Base)).
		Add(erc20.Approve(
			usdcReserve.Underlying(),
			aave.PoolSpender(market),
			txamount.Exact(usdcRepayLimit),
		)).
		Add(aave.RepayAll(usdcReserve)).
		Add(erc20.Approve(
			usdcReserve.Underlying(),
			aave.PoolSpender(market),
			txamount.Exact(decimal.Zero),
		)).
		Add(defi.CheckpointBefore(aave.WithdrawAll(wethReserve), checkpoint)).
		Add(weth.Unwrap(base.WETH, txamount.CheckpointDelta(checkpoint)))
}
