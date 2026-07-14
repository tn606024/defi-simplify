package aave

import (
	"errors"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tn606024/defi-simplify/config"
)

var _ = Describe("Aave deployment manifests", func() {
	It("loads reviewed Base V3 trust anchors without network access", func() {
		manifest, err := BaseV3Deployment()
		Expect(err).NotTo(HaveOccurred())
		Expect(manifest.SchemaVersion()).To(Equal(DeploymentManifestSchemaVersion))

		market := manifest.Market()
		Expect(market.Validate()).To(Succeed())
		Expect(market.ID()).To(Equal("aave-v3-base"))
		Expect(market.Chain()).To(Equal(config.Base))
		Expect(market.Pool()).To(Equal(common.HexToAddress("0xA238Dd80C259a72e81d7e4664a9801593F98d1c5")))
		Expect(market.AddressesProvider()).To(Equal(common.HexToAddress("0xe20fCBdBfFC4Dd138cE8b2E6FBb6CB49777ad64D")))
		Expect(market.ProtocolDataProvider()).To(Equal(common.HexToAddress("0x0F43731EB8d45A581f4a36DD74F5f358bc90C73A")))
		gateway, ok := market.WrappedTokenGateway()
		Expect(ok).To(BeTrue())
		Expect(gateway).To(Equal(common.HexToAddress("0xa0d9C1E9E48Ca30c8d8C3B5D69FF5dc1f6DFfC24")))

		source := manifest.Source()
		Expect(source.Repository()).To(Equal("https://github.com/aave-dao/aave-address-book"))
		Expect(source.Package()).To(Equal("@aave-dao/aave-address-book"))
		Expect(source.PackageVersion()).To(Equal("4.60.0"))
		Expect(source.Release()).To(Equal("v4.60.0"))
		Expect(source.Commit()).To(Equal("7e444a1e73b538fd0b9e093e5156401d6fccca7d"))
		Expect(source.Export()).To(Equal("AaveV3Base"))
	})

	It("returns the manifest market through the convenience API", func() {
		manifest, err := BaseV3Deployment()
		Expect(err).NotTo(HaveOccurred())
		market, err := BaseV3Market()
		Expect(err).NotTo(HaveOccurred())
		Expect(market).To(Equal(manifest.Market()))
	})

	It("fails closed for unsupported or extended manifests", func() {
		unsupported := strings.Replace(
			string(baseV3DeploymentManifest),
			`"marketId": "aave-v3-base"`,
			`"marketId": "aave-v3-unknown"`,
			1,
		)
		_, err := ParseDeploymentManifest([]byte(unsupported))
		Expect(errors.Is(err, ErrInvalidDeploymentManifest)).To(BeTrue())

		extended := strings.Replace(
			string(baseV3DeploymentManifest),
			`"schemaVersion": 1`,
			`"schemaVersion": 1, "unreviewed": true`,
			1,
		)
		_, err = ParseDeploymentManifest([]byte(extended))
		Expect(errors.Is(err, ErrInvalidDeploymentManifest)).To(BeTrue())
	})
})
