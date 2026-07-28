package defisimplify7702_test

import (
	"encoding/json"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	defi "github.com/tn606024/defi-simplify"
	"github.com/tn606024/defi-simplify/client/account/defisimplify7702"
)

type dynamicExecutionGolden struct {
	Token                       string            `json:"token"`
	Target                      string            `json:"target"`
	CheckpointID                string            `json:"checkpointId"`
	PatchOffset                 uint32            `json:"patchOffset"`
	BPS                         uint16            `json:"bps"`
	OriginalCalldata            string            `json:"originalCalldata"`
	ExecuteBatchDynamicCalldata string            `json:"executeBatchDynamicCalldata"`
	Errors                      map[string]string `json:"errors"`
}

var _ = Describe("Dynamic account encoding", func() {
	It("matches the committed Solidity executeBatchDynamic golden vector", func() {
		vector := loadDynamicExecutionGolden()
		original, err := hexutil.Decode(vector.OriginalCalldata)
		Expect(err).NotTo(HaveOccurred())
		expected, err := hexutil.Decode(vector.ExecuteBatchDynamicCalldata)
		Expect(err).NotTo(HaveOccurred())
		checkpointID := common.HexToHash(vector.CheckpointID)

		actual, err := defisimplify7702.EncodeExecuteBatchDynamic([]defi.DynamicCall{{
			Target: common.HexToAddress(vector.Target),
			Value:  big.NewInt(12_345),
			Data:   original,
			CheckpointsBefore: []defi.BalanceCheckpoint{{
				Token: common.HexToAddress(vector.Token),
				ID:    checkpointID,
			}},
			Patches: []defi.BalancePatch{{
				Token:        common.HexToAddress(vector.Token),
				CheckpointID: checkpointID,
				Offset:       vector.PatchOffset,
				BPS:          vector.BPS,
				Source:       defi.BalanceSourceCheckpointDelta,
			}},
			ExpectsCallback: false,
		}})

		Expect(err).NotTo(HaveOccurred())
		Expect(actual).To(Equal(expected))
	})

	It("decodes custom errors with call and patch attribution", func() {
		vector := loadDynamicExecutionGolden()
		data, err := hexutil.Decode(vector.Errors["InvalidPatchOffset"])
		Expect(err).NotTo(HaveOccurred())

		decoded, err := defisimplify7702.DecodeContractError(data)

		Expect(err).NotTo(HaveOccurred())
		Expect(decoded.Name).To(Equal("InvalidPatchOffset"))
		Expect(decoded.Signature).To(Equal("InvalidPatchOffset(uint256,uint256,uint256,uint256)"))
		Expect(decoded.CallIndex).To(Equal(big.NewInt(1)))
		Expect(decoded.PatchIndex).To(Equal(big.NewInt(2)))
		Expect(decoded.Error()).To(ContainSubstring("call=1"))
		Expect(decoded.Error()).To(ContainSubstring("patch=2"))
	})

	It("rejects malformed and unknown custom errors", func() {
		decoded, err := defisimplify7702.DecodeContractError([]byte{0x01})
		Expect(decoded).To(BeNil())
		Expect(errors.Is(err, defisimplify7702.ErrMalformedContractError)).To(BeTrue())

		decoded, err = defisimplify7702.DecodeContractError([]byte{0xde, 0xad, 0xbe, 0xef})
		Expect(decoded).To(BeNil())
		Expect(errors.Is(err, defisimplify7702.ErrUnknownContractError)).To(BeTrue())
	})

	DescribeTable(
		"rejects incompatible dynamic wire shapes",
		func(calls []defi.DynamicCall, message string) {
			data, err := defisimplify7702.EncodeExecuteBatchDynamic(calls)
			Expect(data).To(BeNil())
			Expect(errors.Is(err, defisimplify7702.ErrIncompatibleDynamicCall)).To(BeTrue())
			Expect(err).To(MatchError(ContainSubstring(message)))
		},
		Entry("zero target", []defi.DynamicCall{{}}, "target is zero"),
		Entry(
			"unaligned patch",
			[]defi.DynamicCall{{
				Target: common.HexToAddress("0x1"),
				Data:   make([]byte, 68),
				Patches: []defi.BalancePatch{{
					Token:  common.HexToAddress("0x2"),
					Offset: 5,
					BPS:    10_000,
					Source: defi.BalanceSourceCurrentBalance,
				}},
			}},
			"offset 5",
		),
		Entry(
			"current balance with checkpoint ID",
			[]defi.DynamicCall{{
				Target: common.HexToAddress("0x1"),
				Data:   make([]byte, 36),
				Patches: []defi.BalancePatch{{
					Token:        common.HexToAddress("0x2"),
					CheckpointID: common.HexToHash("0x1234"),
					Offset:       4,
					BPS:          10_000,
					Source:       defi.BalanceSourceCurrentBalance,
				}},
			}},
			"checkpoint ID is not zero",
		),
	)

	It("rejects empty dynamic batches separately", func() {
		data, err := defisimplify7702.EncodeExecuteBatchDynamic(nil)
		Expect(data).To(BeNil())
		Expect(errors.Is(err, defisimplify7702.ErrEmptyDynamicBatch)).To(BeTrue())
	})
})

func loadDynamicExecutionGolden() dynamicExecutionGolden {
	GinkgoHelper()
	data, err := defisimplify7702.Golden(defisimplify7702.DynamicExecutionVector)
	Expect(err).NotTo(HaveOccurred())
	var vector dynamicExecutionGolden
	Expect(json.Unmarshal(data, &vector)).To(Succeed())
	return vector
}
