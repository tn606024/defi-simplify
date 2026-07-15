package aave

import (
	"context"
	"errors"

	"github.com/ethereum/go-ethereum/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tn606024/defi-simplify/config"
)

var _ = Describe("Aave registry", func() {
	var market Market

	BeforeEach(func() {
		var err error
		market, err = NewMarket(
			"aave-v3-base",
			config.Base,
			common.HexToAddress("0x1000000000000000000000000000000000000001"),
			common.HexToAddress("0x1000000000000000000000000000000000000002"),
			common.HexToAddress("0x1000000000000000000000000000000000000003"),
			common.HexToAddress("0x1000000000000000000000000000000000000004"),
		)
		Expect(err).NotTo(HaveOccurred())
	})

	It("performs no reads until the first explicit load and then reuses the cache", func() {
		first := emptyRegistrySnapshot(market, 100, "0x100")
		loader := &recordingSnapshotLoader{snapshots: []*MarketSnapshot{first}}
		registry, err := newRegistry(market, loader)
		Expect(err).NotTo(HaveOccurred())
		Expect(loader.calls).To(Equal(0))
		Expect(registry.Market()).To(Equal(market))

		loaded, err := registry.Load(context.Background())
		Expect(err).NotTo(HaveOccurred())
		Expect(loaded).To(BeIdenticalTo(first))
		cached, err := registry.Load(context.Background())
		Expect(err).NotTo(HaveOccurred())
		Expect(cached).To(BeIdenticalTo(first))
		Expect(loader.calls).To(Equal(1))
	})

	It("refreshes only when requested and replaces the cache after success", func() {
		first := emptyRegistrySnapshot(market, 100, "0x100")
		second := emptyRegistrySnapshot(market, 200, "0x200")
		loader := &recordingSnapshotLoader{snapshots: []*MarketSnapshot{first, second}}
		registry, err := newRegistry(market, loader)
		Expect(err).NotTo(HaveOccurred())

		loaded, err := registry.Load(context.Background())
		Expect(err).NotTo(HaveOccurred())
		Expect(loaded).To(BeIdenticalTo(first))
		refreshed, err := registry.Refresh(context.Background())
		Expect(err).NotTo(HaveOccurred())
		Expect(refreshed).To(BeIdenticalTo(second))
		cached, err := registry.Load(context.Background())
		Expect(err).NotTo(HaveOccurred())
		Expect(cached).To(BeIdenticalTo(second))
		Expect(loader.calls).To(Equal(2))
	})

	It("keeps the previous cache when an explicit refresh fails", func() {
		first := emptyRegistrySnapshot(market, 100, "0x100")
		loadErr := errors.New("RPC unavailable")
		loader := &recordingSnapshotLoader{
			snapshots: []*MarketSnapshot{first, nil},
			errs:      []error{nil, loadErr},
		}
		registry, err := newRegistry(market, loader)
		Expect(err).NotTo(HaveOccurred())
		_, err = registry.Load(context.Background())
		Expect(err).NotTo(HaveOccurred())

		refreshed, err := registry.Refresh(context.Background())
		Expect(refreshed).To(BeNil())
		Expect(errors.Is(err, loadErr)).To(BeTrue())
		cached, err := registry.Load(context.Background())
		Expect(err).NotTo(HaveOccurred())
		Expect(cached).To(BeIdenticalTo(first))
	})

	It("rejects invalid public registry configuration before any RPC read", func() {
		registry, err := NewRegistry(nil, market)
		Expect(registry).To(BeNil())
		Expect(errors.Is(err, ErrInvalidRegistry)).To(BeTrue())

		registry, err = newRegistry(Market{}, &recordingSnapshotLoader{})
		Expect(registry).To(BeNil())
		Expect(errors.Is(err, ErrInvalidRegistry)).To(BeTrue())
	})
})

type recordingSnapshotLoader struct {
	snapshots []*MarketSnapshot
	errs      []error
	calls     int
}

func (l *recordingSnapshotLoader) Load(context.Context, Market) (*MarketSnapshot, error) {
	index := l.calls
	l.calls++
	var snapshot *MarketSnapshot
	if index < len(l.snapshots) {
		snapshot = l.snapshots[index]
	}
	var err error
	if index < len(l.errs) {
		err = l.errs[index]
	}
	return snapshot, err
}

func emptyRegistrySnapshot(market Market, blockNumber uint64, hash string) *MarketSnapshot {
	snapshot, err := NewMarketSnapshot(market, blockNumber, common.HexToHash(hash), nil)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	return snapshot
}
