package weth

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/shopspring/decimal"
	defi "github.com/tn606024/defi-simplify"
	txamount "github.com/tn606024/defi-simplify/amount"
	"github.com/tn606024/defi-simplify/assets/base"
	"github.com/tn606024/defi-simplify/config"
	"github.com/tn606024/defi-simplify/token"
	"github.com/tn606024/defi-simplify/weth/bindings"
)

var _ = Describe("WETH Flow steps", func() {
	var (
		ctx     context.Context
		account common.Address
	)

	BeforeEach(func() {
		ctx = context.Background()
		account = common.HexToAddress("0x00000000000000000000000000000000000000aa")
	})

	It("builds an exact wrap as a payable static call", func() {
		value := decimal.RequireFromString("1.25")
		expected := decimal.RequireFromString("1.25e18").BigInt()
		parsed, err := bindings.WETHMetaData.GetAbi()
		Expect(err).NotTo(HaveOccurred())
		expectedData, err := parsed.Pack("deposit")
		Expect(err).NotTo(HaveOccurred())

		plan, err := defi.NewFlow(account, defi.WithChain(config.Base)).
			Add(Wrap(base.WETH, txamount.Exact(value))).
			Build(ctx, nil)

		Expect(err).NotTo(HaveOccurred())
		Expect(plan.Kind()).To(Equal(defi.PlanStatic))
		calls, err := plan.StaticCalls()
		Expect(err).NotTo(HaveOccurred())
		Expect(calls).To(Equal([]defi.Call{{
			Target: base.WETH.Address(),
			Value:  expected,
			Data:   expectedData,
		}}))
		Expect(plan.Steps()[0].Expectations[0].ExpectationName()).To(Equal("weth.Deposit"))
	})

	It("builds an exact unwrap as a zero-value static call", func() {
		value := decimal.RequireFromString("0.75")
		expected := decimal.RequireFromString("0.75e18").BigInt()
		parsed, err := bindings.WETHMetaData.GetAbi()
		Expect(err).NotTo(HaveOccurred())
		expectedData, err := parsed.Pack("withdraw", expected)
		Expect(err).NotTo(HaveOccurred())

		plan, err := defi.NewFlow(account, defi.WithChain(config.Base)).
			Add(Unwrap(base.WETH, txamount.Exact(value))).
			Build(ctx, nil)

		Expect(err).NotTo(HaveOccurred())
		Expect(plan.Kind()).To(Equal(defi.PlanStatic))
		calls, err := plan.StaticCalls()
		Expect(err).NotTo(HaveOccurred())
		Expect(calls).To(Equal([]defi.Call{{
			Target: base.WETH.Address(),
			Value:  new(big.Int),
			Data:   expectedData,
		}}))
		Expect(plan.Steps()[0].Expectations[0].ExpectationName()).To(Equal("weth.Withdrawal"))
	})

	It("compiles current-balance unwrap at the first ABI argument", func() {
		plan, err := defi.NewFlow(account, defi.WithChain(config.Base)).
			Add(Unwrap(base.WETH, txamount.CurrentBalance(base.WETH))).
			Build(ctx, nil)

		Expect(err).NotTo(HaveOccurred())
		Expect(plan.Kind()).To(Equal(defi.PlanDynamic))
		calls, err := plan.DynamicCalls()
		Expect(err).NotTo(HaveOccurred())
		Expect(calls).To(HaveLen(1))
		Expect(calls[0].Target).To(Equal(base.WETH.Address()))
		Expect(calls[0].Value).To(Equal(new(big.Int)))
		Expect(calls[0].Patches).To(Equal([]defi.BalancePatch{{
			Token:  base.WETH.Address(),
			Offset: withdrawAmountOffset,
			BPS:    txamount.FullBPS,
			Source: defi.BalanceSourceCurrentBalance,
		}}))
		Expect(calls[0].Data[withdrawAmountOffset : withdrawAmountOffset+32]).
			To(Equal(make([]byte, 32)))
	})

	It("compiles checkpoint-delta unwrap without consuming the earlier balance", func() {
		checkpoint := txamount.Checkpoint("before-wrap", base.WETH)
		wrapAmount := decimal.RequireFromString("0.2")

		plan, err := defi.NewFlow(account, defi.WithChain(config.Base)).
			Add(defi.CheckpointBefore(
				Wrap(base.WETH, txamount.Exact(wrapAmount)),
				checkpoint,
			)).
			Add(Unwrap(base.WETH, txamount.CheckpointDelta(checkpoint))).
			Build(ctx, nil)

		Expect(err).NotTo(HaveOccurred())
		Expect(plan.Kind()).To(Equal(defi.PlanDynamic))
		calls, err := plan.DynamicCalls()
		Expect(err).NotTo(HaveOccurred())
		Expect(calls).To(HaveLen(2))
		Expect(calls[0].Value).To(Equal(decimal.RequireFromString("0.2e18").BigInt()))
		Expect(calls[0].CheckpointsBefore).To(HaveLen(1))
		Expect(calls[0].CheckpointsBefore[0].Token).To(Equal(base.WETH.Address()))
		Expect(calls[1].Patches).To(Equal([]defi.BalancePatch{{
			Token:        base.WETH.Address(),
			CheckpointID: calls[0].CheckpointsBefore[0].ID,
			Offset:       withdrawAmountOffset,
			BPS:          txamount.FullBPS,
			Source:       defi.BalanceSourceCheckpointDelta,
		}}))
	})

	It("rejects runtime native value for wrap", func() {
		plan, err := defi.NewFlow(account, defi.WithChain(config.Base)).
			Add(Wrap(base.WETH, txamount.CurrentBalance(base.WETH))).
			Build(ctx, nil)

		Expect(plan).To(BeNil())
		Expect(err).To(MatchError(ContainSubstring("runtime amount source is not supported")))
	})

	It("rejects an unwrap source for a different token", func() {
		other, err := token.NewRef(
			config.Base,
			common.HexToAddress("0x0000000000000000000000000000000000000101"),
		)
		Expect(err).NotTo(HaveOccurred())

		plan, err := defi.NewFlow(account, defi.WithChain(config.Base)).
			Add(Unwrap(base.WETH, txamount.CurrentBalance(other))).
			Build(ctx, nil)

		Expect(plan).To(BeNil())
		Expect(err).To(MatchError(ContainSubstring("does not match WETH contract")))
	})

	DescribeTable(
		"rejects non-positive exact amounts",
		func(step defi.FlowStep) {
			plan, err := defi.NewFlow(account, defi.WithChain(config.Base)).
				Add(step).
				Build(ctx, nil)

			Expect(plan).To(BeNil())
			Expect(err).To(MatchError(ContainSubstring("amount must be positive")))
		},
		Entry("zero wrap", Wrap(base.WETH, txamount.Exact(decimal.Zero))),
		Entry("zero unwrap", Unwrap(base.WETH, txamount.Exact(decimal.Zero))),
	)

	DescribeTable(
		"rejects exact amounts below one native-token unit",
		func(step defi.FlowStep) {
			plan, err := defi.NewFlow(account, defi.WithChain(config.Base)).
				Add(step).
				Build(ctx, nil)

			Expect(plan).To(BeNil())
			Expect(err).To(MatchError(ContainSubstring("rounds to zero at 18 decimals")))
		},
		Entry(
			"wrap",
			Wrap(base.WETH, txamount.Exact(decimal.RequireFromString("0.0000000000000000001"))),
		),
		Entry(
			"unwrap",
			Unwrap(base.WETH, txamount.Exact(decimal.RequireFromString("0.0000000000000000001"))),
		),
	)
})
