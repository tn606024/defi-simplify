package aave

import (
	"fmt"
	"math/big"
)

type eip712Signature struct {
	deadline *big.Int
	v        uint8
	r        [32]byte
	s        [32]byte
}

func newEIP712Signature(deadline *big.Int, v uint8, r, s [32]byte) eip712Signature {
	return eip712Signature{
		deadline: cloneStepBigInt(deadline),
		v:        v,
		r:        r,
		s:        s,
	}
}

func (s eip712Signature) validate() error {
	if s.deadline == nil {
		return fmt.Errorf("signature deadline is nil")
	}
	if s.deadline.Sign() <= 0 {
		return fmt.Errorf("signature deadline must be positive")
	}
	if s.v != 27 && s.v != 28 {
		return fmt.Errorf("signature v must be 27 or 28")
	}
	if s.r == ([32]byte{}) {
		return fmt.Errorf("signature r is zero")
	}
	if s.s == ([32]byte{}) {
		return fmt.Errorf("signature s is zero")
	}
	return nil
}

func cloneStepBigInt(value *big.Int) *big.Int {
	if value == nil {
		return nil
	}
	return new(big.Int).Set(value)
}
