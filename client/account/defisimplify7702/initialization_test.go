package defisimplify7702_test

import (
	"context"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tn606024/defi-simplify/client/account/defisimplify7702"
	"github.com/tn606024/defi-simplify/config"
)

var _ = Describe("Account deployment initialization", func() {
	It("reads the reviewed account implementation at latest state", func() {
		deployment, err := defisimplify7702.DeploymentForChain(config.Base)
		Expect(err).NotTo(HaveOccurred())
		implementation, err := deployment.Contract(defisimplify7702.AccountContract)
		Expect(err).NotTo(HaveOccurred())

		reader := &recordingRuntimeCodeReader{code: []byte{0x01}}
		_, err = defisimplify7702.ResolveAndVerifyAccountDeployment(
			context.Background(),
			reader,
			config.Base,
		)

		Expect(errors.Is(err, defisimplify7702.ErrRuntimeCodeMismatch)).To(BeTrue())
		Expect(reader.addresses).To(Equal([]common.Address{implementation.Address}))
		Expect(reader.blockNumbers).To(Equal([]*big.Int{nil}))
	})

	It("rejects a deployment address without runtime code", func() {
		_, err := defisimplify7702.ResolveAndVerifyAccountDeployment(
			context.Background(),
			&recordingRuntimeCodeReader{},
			config.Base,
		)

		Expect(errors.Is(err, defisimplify7702.ErrMissingRuntimeCode)).To(BeTrue())
	})

	It("preserves runtime code read failures", func() {
		readErr := errors.New("RPC unavailable")
		_, err := defisimplify7702.ResolveAndVerifyAccountDeployment(
			context.Background(),
			&recordingRuntimeCodeReader{err: readErr},
			config.Base,
		)

		Expect(errors.Is(err, readErr)).To(BeTrue())
		Expect(err).To(MatchError(ContainSubstring("read account implementation runtime code")))
	})

	It("rejects missing readers and unsupported chains before chain access", func() {
		_, err := defisimplify7702.ResolveAndVerifyAccountDeployment(
			context.Background(),
			nil,
			config.Base,
		)
		Expect(err).To(MatchError("runtime code reader is nil"))

		reader := &recordingRuntimeCodeReader{}
		_, err = defisimplify7702.ResolveAndVerifyAccountDeployment(
			context.Background(),
			reader,
			config.Chain(999),
		)
		Expect(errors.Is(err, defisimplify7702.ErrUnsupportedChain)).To(BeTrue())
		Expect(reader.addresses).To(BeEmpty())
	})
})

type recordingRuntimeCodeReader struct {
	code         []byte
	err          error
	addresses    []common.Address
	blockNumbers []*big.Int
}

func (r *recordingRuntimeCodeReader) CodeAt(
	_ context.Context,
	address common.Address,
	blockNumber *big.Int,
) ([]byte, error) {
	r.addresses = append(r.addresses, address)
	r.blockNumbers = append(r.blockNumbers, cloneBlockNumber(blockNumber))
	return append([]byte(nil), r.code...), r.err
}

func cloneBlockNumber(value *big.Int) *big.Int {
	if value == nil {
		return nil
	}
	return new(big.Int).Set(value)
}
