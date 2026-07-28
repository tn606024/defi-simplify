package defi

import (
	"context"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/shopspring/decimal"
	"github.com/tn606024/defi-simplify/amount"
	"github.com/tn606024/defi-simplify/config"
	"github.com/tn606024/defi-simplify/token"
)

var _ = Describe("Dynamic execution plans", func() {
	var (
		account common.Address
		asset   token.Ref
	)

	BeforeEach(func() {
		account = common.HexToAddress("0x00000000000000000000000000000000000000aa")
		var err error
		asset, err = token.NewRef(
			config.Base,
			common.HexToAddress("0x0000000000000000000000000000000000000101"),
		)
		Expect(err).NotTo(HaveOccurred())
	})

	It("keeps exact-only flows static", func() {
		expected := Call{
			Target: common.HexToAddress("0x0000000000000000000000000000000000000001"),
			Value:  big.NewInt(7),
			Data:   []byte{0x01, 0x02},
		}

		plan, err := NewFlow(account, WithChain(config.Base)).
			Add(&fakeFlowStep{name: "exact.Step", calls: []Call{expected}}).
			Build(context.Background(), nil)

		Expect(err).NotTo(HaveOccurred())
		Expect(plan.Kind()).To(Equal(PlanStatic))
		Expect(plan.StaticCalls()).To(Equal([]Call{expected}))
		_, err = plan.DynamicCalls()
		Expect(errors.Is(err, ErrPlanKindMismatch)).To(BeTrue())
	})

	It("materializes current-balance patches without exposing placeholder static calls", func() {
		data := calldataWithWords(1)
		step := &fakeFlowStep{
			name: "dynamic.Approve",
			plannedCalls: []PlannedCall{{
				Call: Call{
					Target: common.HexToAddress("0x0000000000000000000000000000000000000001"),
					Data:   data,
				},
				Patches: []CalldataPatch{{
					Source: amount.Scale(amount.CurrentBalance(asset), 3_750),
					Offset: 4,
				}},
			}},
		}

		plan, err := NewFlow(account, WithChain(config.Base)).
			Add(step).
			Build(context.Background(), nil)

		Expect(err).NotTo(HaveOccurred())
		Expect(plan.Kind()).To(Equal(PlanDynamic))
		Expect(plan.Calls()).To(BeNil())
		_, err = plan.StaticCalls()
		Expect(errors.Is(err, ErrPlanKindMismatch)).To(BeTrue())
		calls, err := plan.DynamicCalls()
		Expect(err).NotTo(HaveOccurred())
		Expect(calls).To(HaveLen(1))
		Expect(calls[0].Patches).To(Equal([]BalancePatch{{
			Token:  asset.Address(),
			Offset: 4,
			BPS:    3_750,
			Source: BalanceSourceCurrentBalance,
		}}))
	})

	It("resolves an earlier named checkpoint into a deterministic delta patch", func() {
		checkpoint := amount.Checkpoint("swap-output", asset)
		producer := CheckpointBefore(
			&fakeFlowStep{
				name: "producer.Swap",
				calls: []Call{{
					Target: common.HexToAddress("0x0000000000000000000000000000000000000001"),
					Data:   []byte{0x01},
				}},
			},
			checkpoint,
		)
		consumer := &fakeFlowStep{
			name: "consumer.Supply",
			plannedCalls: []PlannedCall{{
				Call: Call{
					Target: common.HexToAddress("0x0000000000000000000000000000000000000002"),
					Data:   calldataWithWords(1),
				},
				Patches: []CalldataPatch{{
					Source: amount.CheckpointDelta(checkpoint),
					Offset: 4,
				}},
			}},
		}

		plan, err := NewFlow(account, WithChain(config.Base)).
			Add(producer).
			Add(consumer).
			Build(context.Background(), nil)

		Expect(err).NotTo(HaveOccurred())
		calls, err := plan.DynamicCalls()
		Expect(err).NotTo(HaveOccurred())
		Expect(calls).To(HaveLen(2))
		Expect(calls[0].CheckpointsBefore).To(HaveLen(1))
		Expect(calls[0].CheckpointsBefore[0].ID).NotTo(Equal([32]byte{}))
		Expect(calls[1].Patches[0].CheckpointID).To(Equal(calls[0].CheckpointsBefore[0].ID))
		Expect(calls[1].Patches[0].Source).To(Equal(BalanceSourceCheckpointDelta))
	})

	It("rejects a delta consumed before or on its declaring call", func() {
		checkpoint := amount.Checkpoint("later", asset)
		consumer := &fakeFlowStep{
			name: "consumer",
			plannedCalls: []PlannedCall{{
				Call: Call{
					Target: common.HexToAddress("0x0000000000000000000000000000000000000001"),
					Data:   calldataWithWords(1),
				},
				CheckpointsBefore: []CheckpointDeclaration{{Ref: checkpoint}},
				Patches: []CalldataPatch{{
					Source: amount.CheckpointDelta(checkpoint),
					Offset: 4,
				}},
			}},
		}

		plan, err := NewFlow(account, WithChain(config.Base)).
			Add(consumer).
			Build(context.Background(), nil)

		Expect(plan).To(BeNil())
		Expect(errors.Is(err, ErrInvalidCheckpointGraph)).To(BeTrue())
	})

	It("rejects duplicate checkpoint declarations", func() {
		checkpoint := amount.Checkpoint("same", asset)
		first := CheckpointBefore(
			&fakeFlowStep{name: "first", calls: []Call{{Target: common.HexToAddress("0x1")}}},
			checkpoint,
		)
		second := CheckpointBefore(
			&fakeFlowStep{name: "second", calls: []Call{{Target: common.HexToAddress("0x2")}}},
			checkpoint,
		)

		plan, err := NewFlow(account, WithChain(config.Base)).
			Add(first).
			Add(second).
			Build(context.Background(), nil)

		Expect(plan).To(BeNil())
		Expect(errors.Is(err, ErrInvalidCheckpointGraph)).To(BeTrue())
		Expect(err).To(MatchError(ContainSubstring("declared more than once")))
	})

	It("returns defensive copies of static, dynamic, and step call data", func() {
		source := amount.CurrentBalance(asset)
		plan, err := NewFlow(account, WithChain(config.Base)).
			Add(&fakeFlowStep{
				name: "dynamic",
				plannedCalls: []PlannedCall{{
					Call: Call{
						Target: common.HexToAddress("0x1"),
						Value:  big.NewInt(3),
						Data:   calldataWithWords(1),
					},
					Patches: []CalldataPatch{{Source: source, Offset: 4}},
				}},
			}).
			Build(context.Background(), nil)
		Expect(err).NotTo(HaveOccurred())

		first, err := plan.DynamicCalls()
		Expect(err).NotTo(HaveOccurred())
		first[0].Value.SetInt64(99)
		first[0].Data[0] = 0xff
		first[0].Patches[0].BPS = 1
		second, err := plan.DynamicCalls()
		Expect(err).NotTo(HaveOccurred())
		Expect(second[0].Value).To(Equal(big.NewInt(3)))
		Expect(second[0].Data[0]).NotTo(Equal(byte(0xff)))
		Expect(second[0].Patches[0].BPS).To(Equal(uint16(10_000)))

		steps := plan.Steps()
		steps[0].Calls[0].Call.Data[0] = 0xee
		Expect(plan.Steps()[0].Calls[0].Call.Data[0]).NotTo(Equal(byte(0xee)))
	})

	It("rejects non-zero dynamic placeholders", func() {
		data := calldataWithWords(1)
		data[35] = 1
		plan, err := NewFlow(account, WithChain(config.Base)).
			Add(&fakeFlowStep{
				name: "dynamic",
				plannedCalls: []PlannedCall{{
					Call:    Call{Target: common.HexToAddress("0x1"), Data: data},
					Patches: []CalldataPatch{{Source: amount.CurrentBalance(asset), Offset: 4}},
				}},
			}).
			Build(context.Background(), nil)

		Expect(plan).To(BeNil())
		Expect(errors.Is(err, ErrDynamicPlaceholderNotZero)).To(BeTrue())
	})

	It("rejects invalid amount source kinds", func() {
		plan, err := NewFlow(account, WithChain(config.Base)).
			Add(&fakeFlowStep{
				name: "dynamic",
				plannedCalls: []PlannedCall{{
					Call:    Call{Target: common.HexToAddress("0x1"), Data: calldataWithWords(1)},
					Patches: []CalldataPatch{{Source: amount.Source{}, Offset: 4}},
				}},
			}).
			Build(context.Background(), nil)

		Expect(plan).To(BeNil())
		Expect(errors.Is(err, amount.ErrInvalidSource)).To(BeTrue())
	})

	It("does not make amount.Exact dynamic", func() {
		source := amount.Exact(decimal.NewFromInt(1))
		Expect(source.Kind()).To(Equal(amount.KindExact))
	})
})

func calldataWithWords(wordCount int) []byte {
	data := make([]byte, 4+wordCount*32)
	copy(data[:4], []byte{0x12, 0x34, 0x56, 0x78})
	return data
}
