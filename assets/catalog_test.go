package assets_test

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tn606024/defi-simplify/assets"
	"github.com/tn606024/defi-simplify/config"
	"github.com/tn606024/defi-simplify/token"
)

var _ = Describe("Catalog", func() {
	newRef := func(address string) token.Ref {
		ref, err := token.NewRef(config.Base, common.HexToAddress(address))
		Expect(err).NotTo(HaveOccurred())
		return ref
	}
	newEntry := func(id string, ref token.Ref) assets.Entry {
		entry, err := assets.NewEntry(id, ref)
		Expect(err).NotTo(HaveOccurred())
		return entry
	}

	It("sorts and resolves an immutable single-chain catalog", func() {
		usdc := newRef("0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913")
		weth := newRef("0x4200000000000000000000000000000000000006")
		catalog, err := assets.NewCatalog([]assets.Entry{
			newEntry("WETH", weth),
			newEntry("USDC", usdc),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(catalog.Chain()).To(Equal(config.Base))
		Expect(catalog.Entries()).To(HaveExactElements(
			newEntry("USDC", usdc),
			newEntry("WETH", weth),
		))
		resolved, ok := catalog.Lookup("USDC")
		Expect(ok).To(BeTrue())
		Expect(resolved).To(Equal(usdc))
		_, ok = catalog.Lookup("usdc")
		Expect(ok).To(BeFalse())

		entries := catalog.Entries()
		entries[0] = assets.Entry{}
		Expect(catalog.Entries()[0].ID()).To(Equal("USDC"))
	})

	DescribeTable("rejects malformed catalogs",
		func(entries func() []assets.Entry) {
			_, err := assets.NewCatalog(entries())
			Expect(errors.Is(err, assets.ErrInvalidCatalog)).To(BeTrue())
		},
		Entry("empty", func() []assets.Entry { return nil }),
		Entry("zero entry", func() []assets.Entry { return []assets.Entry{{}} }),
		Entry("duplicate ID", func() []assets.Entry {
			return []assets.Entry{
				newEntry("USDC", newRef("0x1111111111111111111111111111111111111111")),
				newEntry("USDC", newRef("0x2222222222222222222222222222222222222222")),
			}
		}),
		Entry("duplicate token reference", func() []assets.Entry {
			ref := newRef("0x1111111111111111111111111111111111111111")
			return []assets.Entry{newEntry("ONE", ref), newEntry("TWO", ref)}
		}),
	)
})
