package defisimplify7702_test

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tn606024/defi-simplify/client/account/defisimplify7702"
	"github.com/tn606024/defi-simplify/config"
)

var _ = Describe("Defi Simplify contract artifacts", func() {
	It("loads the configured Base deployment", func() {
		Expect(defisimplify7702.VerifyArtifacts()).To(Succeed())

		deployment, err := defisimplify7702.DeploymentForChain(config.Base)
		Expect(err).NotTo(HaveOccurred())
		Expect(deployment.Chain).To(Equal(config.Base))
		Expect(deployment.ChainID).To(Equal(8453))
		Expect(deployment.NetworkName).To(Equal("Base"))
		Expect(deployment.EntryPoint.Address).NotTo(Equal(common.Address{}))
		Expect(deployment.EntryPoint.RuntimeCodeHash).NotTo(Equal(common.Hash{}))
		Expect(deployment.EntryPoint.Version).NotTo(BeEmpty())
		Expect(deployment.Contracts()).To(ConsistOf(
			defisimplify7702.AccountContract,
			defisimplify7702.FlowAssertionsContract,
			defisimplify7702.StaticCallUint256AssertionsContract,
		))

		for _, contract := range deployment.Contracts() {
			artifact, err := deployment.Contract(contract)
			Expect(err).NotTo(HaveOccurred())
			Expect(artifact.Address).NotTo(Equal(common.Address{}))
			Expect(artifact.RuntimeCodeHash).NotTo(Equal(common.Hash{}))
			Expect(artifact.ABIPath).NotTo(BeEmpty())

			abiJSON, err := defisimplify7702.ABI(contract)
			Expect(err).NotTo(HaveOccurred())
			_, err = abi.JSON(bytes.NewReader(abiJSON))
			Expect(err).NotTo(HaveOccurred())
		}

		Expect(deployment.Security.Experimental).To(BeTrue())
		Expect(deployment.Security.Status).To(Equal("unaudited"))
		Expect(deployment.Security.SecurityGuarantee).To(BeFalse())
		Expect(deployment.Security.TotalLossRisk).To(BeTrue())
		Expect(deployment.Security.Notice).NotTo(BeEmpty())
	})

	It("returns copies of embedded artifacts and contract lists", func() {
		firstABI, err := defisimplify7702.ABI(defisimplify7702.AccountContract)
		Expect(err).NotTo(HaveOccurred())
		firstABI[0] = '!'

		secondABI, err := defisimplify7702.ABI(defisimplify7702.AccountContract)
		Expect(err).NotTo(HaveOccurred())
		Expect(secondABI[0]).NotTo(Equal(byte('!')))

		deployment, err := defisimplify7702.DeploymentForChain(config.Base)
		Expect(err).NotTo(HaveOccurred())
		contracts := deployment.Contracts()
		contracts[0] = defisimplify7702.Contract("changed")
		Expect(deployment.Contracts()).NotTo(ContainElement(defisimplify7702.Contract("changed")))
	})

	It("makes all parity vectors available as JSON", func() {
		vectors := []defisimplify7702.GoldenVector{
			defisimplify7702.BaseAaveV3DynamicStrategyVector,
			defisimplify7702.BaseAaveV3FlashLifecycleVector,
			defisimplify7702.CallbackExecutionVector,
			defisimplify7702.DynamicCalldataPatchingVector,
			defisimplify7702.DynamicExecutionVector,
		}
		for _, vector := range vectors {
			data, err := defisimplify7702.Golden(vector)
			Expect(err).NotTo(HaveOccurred())
			Expect(json.Valid(data)).To(BeTrue(), string(vector))
		}
	})

	It("rejects unsupported chains and unknown artifacts", func() {
		_, err := defisimplify7702.DeploymentForChain(config.Chain(999))
		Expect(errors.Is(err, defisimplify7702.ErrUnsupportedChain)).To(BeTrue())

		_, err = defisimplify7702.ABI(defisimplify7702.Contract("Unknown"))
		Expect(errors.Is(err, defisimplify7702.ErrUnknownArtifact)).To(BeTrue())
	})

	It("checks runtime bytecode against the configured identity", func() {
		code := []byte{0x60, 0x00, 0x60, 0x00}
		identity := defisimplify7702.RuntimeIdentity{
			Name:            "TestContract",
			Address:         common.HexToAddress("0x0000000000000000000000000000000000000001"),
			RuntimeCodeHash: crypto.Keccak256Hash(code),
		}
		Expect(identity.VerifyRuntimeCode(code)).To(Succeed())
		err := identity.VerifyRuntimeCode([]byte{0x00})
		Expect(errors.Is(err, defisimplify7702.ErrRuntimeCodeMismatch)).To(BeTrue())
	})
})
