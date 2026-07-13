package aave

import (
	"math/big"
	"strings"
	"testing"
)

func TestEIP712SignatureValidate(t *testing.T) {
	validR := [32]byte{1}
	validS := [32]byte{2}

	tests := []struct {
		name    string
		sig     eip712Signature
		wantErr string
	}{
		{
			name: "valid v27 signature",
			sig:  newEIP712Signature(big.NewInt(1), 27, validR, validS),
		},
		{
			name: "valid v28 signature",
			sig:  newEIP712Signature(big.NewInt(1), 28, validR, validS),
		},
		{
			name:    "nil deadline",
			sig:     newEIP712Signature(nil, 27, validR, validS),
			wantErr: "signature deadline is nil",
		},
		{
			name:    "non-positive deadline",
			sig:     newEIP712Signature(big.NewInt(0), 27, validR, validS),
			wantErr: "signature deadline must be positive",
		},
		{
			name:    "invalid v",
			sig:     newEIP712Signature(big.NewInt(1), 1, validR, validS),
			wantErr: "signature v must be 27 or 28",
		},
		{
			name:    "zero r",
			sig:     newEIP712Signature(big.NewInt(1), 27, [32]byte{}, validS),
			wantErr: "signature r is zero",
		},
		{
			name:    "zero s",
			sig:     newEIP712Signature(big.NewInt(1), 27, validR, [32]byte{}),
			wantErr: "signature s is zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.sig.validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("validate() error = %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("validate() error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}
