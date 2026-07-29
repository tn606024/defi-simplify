package defisimplify7702

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// BalanceSource is the account ABI value used to resolve a runtime token
// balance.
type BalanceSource uint8

const (
	BalanceSourceCurrentBalance BalanceSource = iota
	BalanceSourceCheckpointDelta
)

// BalanceCheckpoint is one account-level token balance checkpoint.
type BalanceCheckpoint struct {
	Token common.Address
	ID    [32]byte
}

// BalancePatch describes one account-level runtime calldata replacement.
type BalancePatch struct {
	Token        common.Address
	CheckpointID [32]byte
	Offset       uint32
	BPS          uint16
	Source       BalanceSource
}

// DynamicCall is the account wire value accepted by executeBatchDynamic.
//
// The root defi package owns executor-neutral plan values. Runner translates
// those values into this account-owned type before submission so this package
// remains independent from the root execution package.
type DynamicCall struct {
	Target            common.Address
	Value             *big.Int
	Data              []byte
	CheckpointsBefore []BalanceCheckpoint
	Patches           []BalancePatch
	ExpectsCallback   bool
}
