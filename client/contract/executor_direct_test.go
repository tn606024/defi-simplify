package contract

import (
	"math"
	"testing"
)

func TestAddDirectGasLimitBuffer(t *testing.T) {
	tests := []struct {
		name      string
		estimated uint64
		want      uint64
		wantErr   bool
	}{
		{name: "adds twenty five percent", estimated: 21_000, want: 26_250},
		{name: "zero remains zero", estimated: 0, want: 0},
		{name: "rejects overflow", estimated: math.MaxUint64, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := addDirectGasLimitBuffer(tt.estimated)
			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error state: got %v wantErr %t", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("unexpected buffered gas limit: got %d want %d", got, tt.want)
			}
		})
	}
}
