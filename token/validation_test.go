package token

import (
	"errors"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/tn606024/defi-simplify/config"
)

func TestValidateRef(t *testing.T) {
	validAddress := common.HexToAddress("0x1111111111111111111111111111111111111111")
	tests := []struct {
		name    string
		chain   config.Chain
		address common.Address
		wantErr error
	}{
		{name: "valid", chain: config.Base, address: validAddress},
		{name: "unsupported chain", chain: config.Chain(999), address: validAddress, wantErr: ErrInvalidRef},
		{name: "zero address", chain: config.Base, wantErr: ErrInvalidRef},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRef(tt.chain, tt.address)
			if tt.wantErr == nil && err != nil {
				t.Fatalf("validateRef() error = %v", err)
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Fatalf("validateRef() error = %v, want errors.Is(%v)", err, tt.wantErr)
			}
		})
	}
}
