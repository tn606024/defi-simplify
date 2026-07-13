package aave

import (
	"math/big"
	"testing"
)

func TestNewUint256MaxReturnsIndependentValues(t *testing.T) {
	want := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))
	first := newUint256Max()
	second := newUint256Max()

	if first.Cmp(want) != 0 || second.Cmp(want) != 0 {
		t.Fatalf("newUint256Max() values = %s and %s, want %s", first, second, want)
	}
	first.SetInt64(0)
	if second.Cmp(want) != 0 {
		t.Fatalf("mutating one max value changed another: got %s, want %s", second, want)
	}
}
