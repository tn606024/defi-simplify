package defi

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/tn606024/defi-simplify/client/account/defisimplify7702"
)

func TestToDynamicAccountCallsCopiesPlanValues(t *testing.T) {
	source := []DynamicCall{{
		Target: common.HexToAddress("0x1000000000000000000000000000000000000000"),
		Value:  big.NewInt(7),
		Data:   []byte{0x01, 0x02},
		CheckpointsBefore: []BalanceCheckpoint{{
			Token: common.HexToAddress("0x2000000000000000000000000000000000000000"),
			ID:    common.HexToHash("0x1234"),
		}},
		Patches: []BalancePatch{{
			Token:        common.HexToAddress("0x2000000000000000000000000000000000000000"),
			CheckpointID: common.HexToHash("0x1234"),
			Offset:       36,
			BPS:          5_000,
			Source:       BalanceSourceCheckpointDelta,
		}},
		ExpectsCallback: true,
	}}

	translated := toDynamicAccountCalls(source)

	if len(translated) != 1 {
		t.Fatalf("unexpected translated call count: %d", len(translated))
	}
	if translated[0].Target != source[0].Target ||
		translated[0].Value.Cmp(source[0].Value) != 0 ||
		string(translated[0].Data) != string(source[0].Data) ||
		!translated[0].ExpectsCallback {
		t.Fatalf("unexpected translated call: %+v", translated[0])
	}
	if translated[0].Patches[0].Source != defisimplify7702.BalanceSourceCheckpointDelta {
		t.Fatalf("unexpected translated balance source: %d", translated[0].Patches[0].Source)
	}

	translated[0].Value.SetInt64(99)
	translated[0].Data[0] = 0xff
	translated[0].CheckpointsBefore[0].ID = common.Hash{}
	translated[0].Patches[0].BPS = 1

	if source[0].Value.Cmp(big.NewInt(7)) != 0 ||
		source[0].Data[0] != 0x01 ||
		source[0].CheckpointsBefore[0].ID == (common.Hash{}) ||
		source[0].Patches[0].BPS != 5_000 {
		t.Fatal("account translation mutated plan-owned values")
	}
}
