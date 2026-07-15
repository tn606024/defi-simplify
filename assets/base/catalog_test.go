package base_test

import (
	"bytes"
	"os"
	"sort"

	"github.com/ethereum/go-ethereum/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tn606024/defi-simplify/assets/base"
	"github.com/tn606024/defi-simplify/config"
	"github.com/tn606024/defi-simplify/internal/aaveaddressbook"
	"github.com/tn606024/defi-simplify/internal/aaveassetmanifest"
	"github.com/tn606024/defi-simplify/internal/assetmanifest"
	"github.com/tn606024/defi-simplify/internal/catalogcodegen"
)

var _ = Describe("Base asset catalog", func() {
	DescribeTable("exposes reviewed chain-scoped references",
		func(id string, refAddress common.Address) {
			ref, ok := base.Lookup(id)
			Expect(ok).To(BeTrue())
			Expect(ref.Validate()).To(Succeed())
			Expect(ref.Chain()).To(Equal(config.Base))
			Expect(ref.Address()).To(Equal(refAddress))
		},
		Entry("AAVE", "AAVE", base.AAVE.Address()),
		Entry("cbBTC", "CBBTC", base.CBBTC.Address()),
		Entry("cbETH", "CBETH", base.CBETH.Address()),
		Entry("EURC", "EURC", base.EURC.Address()),
		Entry("ezETH", "EZETH", base.EZETH.Address()),
		Entry("GHO", "GHO", base.GHO.Address()),
		Entry("LBTC", "LBTC", base.LBTC.Address()),
		Entry("syrupUSDC", "SYRUPUSDC", base.SYRUPUSDC.Address()),
		Entry("tBTC", "TBTC", base.TBTC.Address()),
		Entry("USDbC", "USDBC", base.USDBC.Address()),
		Entry(
			"native USDC",
			"USDC",
			common.HexToAddress("0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913"),
		),
		Entry("weETH", "WEETH", base.WEETH.Address()),
		Entry(
			"WETH",
			"WETH",
			common.HexToAddress("0x4200000000000000000000000000000000000006"),
		),
		Entry("wrsETH", "WRSETH", base.WRSETH.Address()),
		Entry("wstETH", "WSTETH", base.WSTETH.Address()),
	)

	It("uses exact catalog IDs rather than symbol-style matching", func() {
		_, ok := base.Lookup("usdc")
		Expect(ok).To(BeFalse())
		_, ok = base.Lookup("UNKNOWN")
		Expect(ok).To(BeFalse())
	})

	It("returns a deterministic immutable catalog view", func() {
		first := base.Entries()
		second := base.Entries()
		Expect(first).To(HaveLen(15))
		Expect(second).To(Equal(first))

		ids := make([]string, len(first))
		for i, entry := range first {
			ids[i] = entry.ID()
			Expect(entry.Ref().Validate()).To(Succeed())
		}
		Expect(sort.StringsAreSorted(ids)).To(BeTrue())

		first[0] = base.Entry{}
		Expect(base.Entries()[0].ID()).NotTo(BeEmpty())
	})

	It("keeps generated named references in sync with the reviewed manifest", func() {
		manifestData, err := os.ReadFile("manifest.json")
		Expect(err).NotTo(HaveOccurred())
		definition := aaveassetmanifest.DefinitionFor(aaveaddressbook.BaseV3ExportDefinition())
		manifest, err := assetmanifest.Parse(manifestData, definition)
		Expect(err).NotTo(HaveOccurred())
		want, err := catalogcodegen.Generate("base", manifest.Assets)
		Expect(err).NotTo(HaveOccurred())
		got, err := os.ReadFile("catalog_gen.go")
		Expect(err).NotTo(HaveOccurred())
		Expect(bytes.Equal(got, want)).To(BeTrue(), "catalog_gen.go is stale; run make update-aave-manifests")
	})
})
