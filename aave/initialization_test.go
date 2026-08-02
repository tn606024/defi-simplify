package aave

import (
	"context"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Base Aave initialization", func() {
	It("rejects a missing registry backend", func() {
		_, err := LoadBaseV3Snapshot(context.Background(), nil)

		Expect(errors.Is(err, ErrInvalidRegistry)).To(BeTrue())
		Expect(err).To(MatchError(ContainSubstring("create Base Aave V3 registry")))
	})

	It("preserves snapshot discovery failures", func() {
		readErr := errors.New("RPC unavailable")
		_, err := LoadBaseV3Snapshot(
			context.Background(),
			&failingBaseSnapshotBackend{err: readErr},
		)

		Expect(errors.Is(err, ErrRegistryDiscovery)).To(BeTrue())
		Expect(errors.Is(err, readErr)).To(BeTrue())
		Expect(err).To(MatchError(ContainSubstring("load Base Aave V3 snapshot")))
	})
})

type failingBaseSnapshotBackend struct {
	err error
}

func (b *failingBaseSnapshotBackend) HeaderByNumber(
	context.Context,
	*big.Int,
) (*types.Header, error) {
	return nil, b.err
}

func (b *failingBaseSnapshotBackend) CodeAt(
	context.Context,
	common.Address,
	*big.Int,
) ([]byte, error) {
	return nil, b.err
}

func (b *failingBaseSnapshotBackend) CallContract(
	context.Context,
	ethereum.CallMsg,
	*big.Int,
) ([]byte, error) {
	return nil, b.err
}

func (b *failingBaseSnapshotBackend) CodeAtHash(
	context.Context,
	common.Address,
	common.Hash,
) ([]byte, error) {
	return nil, b.err
}

func (b *failingBaseSnapshotBackend) CallContractAtHash(
	context.Context,
	ethereum.CallMsg,
	common.Hash,
) ([]byte, error) {
	return nil, b.err
}
