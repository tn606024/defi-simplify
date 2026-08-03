package main

import (
	"encoding/hex"
	"fmt"
	"io"
	"math/big"

	defi "github.com/tn606024/defi-simplify"
	"github.com/tn606024/defi-simplify/aave"
	"github.com/tn606024/defi-simplify/amount"
	"github.com/tn606024/defi-simplify/erc20"
	"github.com/tn606024/defi-simplify/weth"
)

func (app *application) printHeader() {
	fmt.Fprintf(app.out, "network=base chain_id=8453\n")
	fmt.Fprintf(app.out, "account=%s\n", app.account.Hex())
	fmt.Fprintf(app.out, "implementation=%s version=%s\n", app.implementation.Address.Hex(), app.implementation.Version)
	fmt.Fprintf(app.out, "aave_pool=%s snapshot_block=%d snapshot_hash=%s\n", app.market.Pool().Hex(), app.snapshotBlock, app.snapshotHash.Hex())
	fmt.Fprintf(app.out, "phase=%s broadcast=%t\n", app.config.phase, app.config.broadcast)
	switch app.config.phase {
	case phaseOpen:
		fmt.Fprintf(app.out, "amount.weth_supply=%s amount.usdc_borrow=%s\n", app.config.wethSupplyAmount, app.config.usdcBorrowAmount)
	case phaseClose:
		fmt.Fprintf(app.out, "amount.usdc_repay_limit=%s\n", app.config.usdcRepayLimit)
	}
}

func (app *application) printState(label string, state *accountSnapshot) {
	fmt.Fprintf(
		app.out,
		"state.%s.delegation status=%s implementation=%s\n",
		label,
		state.delegation.Status,
		state.delegation.Implementation.Hex(),
	)
	fmt.Fprintf(app.out, "state.%s.native_balance=%s\n", label, amountString(state.nativeBalance))
	fmt.Fprintf(
		app.out,
		"state.%s.weth_token balance=%s allowance=%s block=%d hash=%s\n",
		label,
		amountString(state.wethToken.balance),
		amountString(state.wethToken.allowance),
		state.wethToken.blockNumber,
		state.wethToken.blockHash.Hex(),
	)
	fmt.Fprintf(
		app.out,
		"state.%s.weth_aave collateral=%s stable_debt=%s variable_debt=%s collateral_enabled=%t block=%d hash=%s\n",
		label,
		amountString(state.wethPosition.aTokenBalance),
		amountString(state.wethPosition.stableDebt),
		amountString(state.wethPosition.variableDebt),
		state.wethPosition.collateralEnabled,
		state.wethPosition.blockNumber,
		state.wethPosition.blockHash.Hex(),
	)
	fmt.Fprintf(
		app.out,
		"state.%s.usdc_token balance=%s allowance=%s block=%d hash=%s\n",
		label,
		amountString(state.usdcToken.balance),
		amountString(state.usdcToken.allowance),
		state.usdcToken.blockNumber,
		state.usdcToken.blockHash.Hex(),
	)
	fmt.Fprintf(
		app.out,
		"state.%s.usdc_aave collateral=%s stable_debt=%s variable_debt=%s collateral_enabled=%t block=%d hash=%s\n",
		label,
		amountString(state.usdcPosition.aTokenBalance),
		amountString(state.usdcPosition.stableDebt),
		amountString(state.usdcPosition.variableDebt),
		state.usdcPosition.collateralEnabled,
		state.usdcPosition.blockNumber,
		state.usdcPosition.blockHash.Hex(),
	)
}

func (app *application) printPlan(plan *defi.ExecutionPlan) {
	fmt.Fprintf(app.out, "plan account=%s kind=%s\n", plan.Account().Hex(), planKindName(plan.Kind()))
	for _, step := range plan.Steps() {
		fmt.Fprintf(app.out, "plan.step id=%s name=%s\n", step.ID, step.Name)
		for callIndex, planned := range step.Calls {
			fmt.Fprintf(
				app.out,
				"plan.call step=%s index=%d target=%s value=%s selector=%s\n",
				step.ID,
				callIndex,
				planned.Call.Target.Hex(),
				amountString(planned.Call.Value),
				selectorString(planned.Call.Data),
			)
			for _, checkpoint := range planned.CheckpointsBefore {
				fmt.Fprintf(
					app.out,
					"plan.checkpoint step=%s call=%d label=%s token=%s\n",
					step.ID,
					callIndex,
					checkpoint.Ref.Label(),
					checkpoint.Ref.Token().Address().Hex(),
				)
			}
			for _, patch := range planned.Patches {
				fmt.Fprintf(
					app.out,
					"plan.patch step=%s call=%d offset=%d source=%s token=%s bps=%d\n",
					step.ID,
					callIndex,
					patch.Offset,
					amountKindName(patch.Source.Kind()),
					amountSourceToken(patch.Source),
					patch.Source.BPS(),
				)
			}
		}
	}
}

func printTypedEvents(out io.Writer, result *defi.ExecutionResult) {
	for _, event := range defi.EventsOf[*weth.DepositEvent](result) {
		fmt.Fprintf(out, "event.weth.deposit contract=%s account=%s amount=%s log_index=%d\n", event.Contract.Hex(), event.Account.Hex(), amountString(event.Amount), event.Metadata.LogIndex)
	}
	for _, event := range defi.EventsOf[*erc20.ApprovalEvent](result) {
		fmt.Fprintf(out, "event.erc20.approval token=%s owner=%s spender=%s amount=%s log_index=%d\n", event.Token.Hex(), event.Owner.Hex(), event.Spender.Hex(), amountString(event.Amount), event.Metadata.LogIndex)
	}
	for _, event := range defi.EventsOf[*aave.SupplyEvent](result) {
		fmt.Fprintf(out, "event.aave.supply asset=%s user=%s on_behalf_of=%s amount=%s log_index=%d\n", event.Asset.Hex(), event.User.Hex(), event.OnBehalfOf.Hex(), amountString(event.Amount), event.Metadata.LogIndex)
	}
	for _, event := range defi.EventsOf[*aave.BorrowEvent](result) {
		fmt.Fprintf(out, "event.aave.borrow asset=%s user=%s on_behalf_of=%s amount=%s rate_mode=%d borrow_rate=%s log_index=%d\n", event.Asset.Hex(), event.User.Hex(), event.OnBehalfOf.Hex(), amountString(event.Amount), event.InterestRateMode, amountString(event.BorrowRate), event.Metadata.LogIndex)
	}
	for _, event := range defi.EventsOf[*aave.RepayEvent](result) {
		fmt.Fprintf(out, "event.aave.repay asset=%s user=%s repayer=%s amount=%s use_atokens=%t log_index=%d\n", event.Asset.Hex(), event.User.Hex(), event.Repayer.Hex(), amountString(event.Amount), event.UseATokens, event.Metadata.LogIndex)
	}
	for _, event := range defi.EventsOf[*aave.WithdrawEvent](result) {
		fmt.Fprintf(out, "event.aave.withdraw asset=%s user=%s to=%s amount=%s log_index=%d\n", event.Asset.Hex(), event.User.Hex(), event.To.Hex(), amountString(event.Amount), event.Metadata.LogIndex)
	}
	for _, event := range defi.EventsOf[*weth.WithdrawalEvent](result) {
		fmt.Fprintf(out, "event.weth.withdrawal contract=%s account=%s amount=%s log_index=%d\n", event.Contract.Hex(), event.Account.Hex(), amountString(event.Amount), event.Metadata.LogIndex)
	}
}

func planKindName(kind defi.PlanKind) string {
	switch kind {
	case defi.PlanStatic:
		return "static"
	case defi.PlanDynamic:
		return "dynamic"
	default:
		return fmt.Sprintf("unknown(%d)", kind)
	}
}

func amountKindName(kind amount.Kind) string {
	switch kind {
	case amount.KindExact:
		return "exact"
	case amount.KindCurrentBalance:
		return "current_balance"
	case amount.KindCheckpointDelta:
		return "checkpoint_delta"
	default:
		return fmt.Sprintf("unknown(%d)", kind)
	}
}

func amountSourceToken(source amount.Source) string {
	asset, ok := source.Token()
	if !ok {
		return "none"
	}
	return asset.Address().Hex()
}

func selectorString(data []byte) string {
	if len(data) < 4 {
		return "none"
	}
	return "0x" + hex.EncodeToString(data[:4])
}

func amountString(value *big.Int) string {
	if value == nil {
		return "<nil>"
	}
	return value.String()
}
