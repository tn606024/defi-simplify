package defi

import (
	"github.com/tn606024/defi-simplify/client/account/defisimplify7702"
)

func toDynamicAccountCalls(calls []DynamicCall) []defisimplify7702.DynamicCall {
	translated := make([]defisimplify7702.DynamicCall, len(calls))
	for i, call := range calls {
		translated[i] = defisimplify7702.DynamicCall{
			Target:            call.Target,
			Value:             cloneBigInt(call.Value),
			Data:              append([]byte(nil), call.Data...),
			ExpectsCallback:   call.ExpectsCallback,
			CheckpointsBefore: make([]defisimplify7702.BalanceCheckpoint, len(call.CheckpointsBefore)),
			Patches:           make([]defisimplify7702.BalancePatch, len(call.Patches)),
		}
		for j, checkpoint := range call.CheckpointsBefore {
			translated[i].CheckpointsBefore[j] = defisimplify7702.BalanceCheckpoint{
				Token: checkpoint.Token,
				ID:    checkpoint.ID,
			}
		}
		for j, patch := range call.Patches {
			translated[i].Patches[j] = defisimplify7702.BalancePatch{
				Token:        patch.Token,
				CheckpointID: patch.CheckpointID,
				Offset:       patch.Offset,
				BPS:          patch.BPS,
				Source:       defisimplify7702.BalanceSource(patch.Source),
			}
		}
	}
	return translated
}
