//go:build integration

package integration

import (
	"context"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tn606024/defi-simplify/aave"
	"github.com/tn606024/defi-simplify/config"
)

var _ = Describe("Aave registry discovery", func() {
	It("discovers Base reserves through one pinned block", func() {
		ctx := context.Background()
		client := baseForkClient(GinkgoT())
		backend := &recordingRegistryBackend{Client: client}

		market, err := aave.NewMarket(
			"aave-v3-base",
			config.Base,
			common.HexToAddress("0xA238Dd80C259a72e81d7e4664a9801593F98d1c5"),
			common.HexToAddress("0xe20fCBdBfFC4Dd138cE8b2E6FBb6CB49777ad64D"),
			common.HexToAddress("0x0F43731EB8d45A581f4a36DD74F5f358bc90C73A"),
			common.HexToAddress("0xa0d9C1E9E48Ca30c8d8C3B5D69FF5dc1f6DFfC24"),
		)
		Expect(err).NotTo(HaveOccurred())
		registry, err := aave.NewRegistry(backend, market)
		Expect(err).NotTo(HaveOccurred())

		snapshot, err := registry.Load(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(snapshot.Validate()).To(Succeed())
		Expect(snapshot.BlockNumber()).NotTo(BeZero())
		Expect(snapshot.BlockHash()).NotTo(Equal(common.Hash{}))
		Expect(snapshot.Len()).To(BeNumerically(">=", 2))

		usdc, err := snapshot.ReserveByAddress(
			common.HexToAddress("0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913"),
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(usdc.Underlying().Symbol()).To(Equal("USDC"))
		Expect(usdc.Underlying().Decimals()).To(Equal(uint8(6)))
		Expect(usdc.AToken().Address()).To(Equal(
			common.HexToAddress("0x4e65fE4DbA92790696d040ac24Aa414708F5c0AB"),
		))
		Expect(usdc.VariableDebtToken().Address()).To(Equal(
			common.HexToAddress("0x59dca05b6c26dbd64b5381374aAaC5CD05644C28"),
		))

		weth, err := snapshot.ReserveByAddress(
			common.HexToAddress("0x4200000000000000000000000000000000000006"),
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(weth.Underlying().Symbol()).To(Equal("WETH"))
		Expect(weth.Underlying().Decimals()).To(Equal(uint8(18)))
		Expect(weth.AToken().Address()).To(Equal(
			common.HexToAddress("0xD4a0e0b9149BCee3C920d2E00b5dE09138fd8bb7"),
		))
		Expect(weth.VariableDebtToken().Address()).To(Equal(
			common.HexToAddress("0x24e6e0795b3c7c71D965fCc4f371803d1c1DcA1E"),
		))

		headerRequests, readHashes := backend.recordedReads()
		Expect(headerRequests).To(HaveLen(1))
		Expect(headerRequests[0]).To(BeNil())
		Expect(readHashes).NotTo(BeEmpty())
		for _, hash := range readHashes {
			Expect(hash).To(Equal(snapshot.BlockHash()))
		}

		readsBeforeCachedLoad := len(readHashes)
		cached, err := registry.Load(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(cached).To(BeIdenticalTo(snapshot))
		_, readsAfterCachedLoad := backend.recordedReads()
		Expect(readsAfterCachedLoad).To(HaveLen(readsBeforeCachedLoad))
	})
})

type recordingRegistryBackend struct {
	*ethclient.Client

	mu             sync.Mutex
	headerRequests []*big.Int
	readHashes     []common.Hash
}

func (b *recordingRegistryBackend) HeaderByNumber(
	ctx context.Context,
	number *big.Int,
) (*types.Header, error) {
	b.mu.Lock()
	b.headerRequests = append(b.headerRequests, cloneRecordedBlock(number))
	b.mu.Unlock()
	return b.Client.HeaderByNumber(ctx, number)
}

func (b *recordingRegistryBackend) CodeAtHash(
	ctx context.Context,
	contract common.Address,
	blockHash common.Hash,
) ([]byte, error) {
	b.recordRead(blockHash)
	return b.Client.CodeAtHash(ctx, contract, blockHash)
}

func (b *recordingRegistryBackend) CallContractAtHash(
	ctx context.Context,
	call ethereum.CallMsg,
	blockHash common.Hash,
) ([]byte, error) {
	b.recordRead(blockHash)
	return b.Client.CallContractAtHash(ctx, call, blockHash)
}

func (b *recordingRegistryBackend) recordRead(blockHash common.Hash) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.readHashes = append(b.readHashes, blockHash)
}

func (b *recordingRegistryBackend) recordedReads() ([]*big.Int, []common.Hash) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return cloneRecordedBlocks(b.headerRequests), append([]common.Hash(nil), b.readHashes...)
}

func cloneRecordedBlocks(values []*big.Int) []*big.Int {
	cloned := make([]*big.Int, len(values))
	for i, value := range values {
		cloned[i] = cloneRecordedBlock(value)
	}
	return cloned
}

func cloneRecordedBlock(value *big.Int) *big.Int {
	if value == nil {
		return nil
	}
	return new(big.Int).Set(value)
}
