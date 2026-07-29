package defi

import (
	"context"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tn606024/defi-simplify/client/contract"
	"github.com/tn606024/defi-simplify/config"
)

type fakeFlowStep struct {
	name         string
	calls        []Call
	plannedCalls []PlannedCall
	expectations []EventExpectation
	err          error
	seenEnvs     []BuildEnv
}

func (s *fakeFlowStep) Build(ctx context.Context, env BuildEnv) (BuiltStep, error) {
	s.seenEnvs = append(s.seenEnvs, env)
	planned := s.plannedCalls
	if planned == nil {
		planned = make([]PlannedCall, len(s.calls))
		for i, call := range s.calls {
			planned[i] = PlannedCall{Call: call}
		}
	}
	built := BuiltStep{Name: s.name, Calls: planned, Expectations: s.expectations}
	if s.err != nil {
		return built, s.err
	}
	return built, nil
}

var _ = Describe("Flow", func() {
	var (
		ctx  context.Context
		user common.Address
	)

	BeforeEach(func() {
		ctx = context.Background()
		user = common.HexToAddress("0x00000000000000000000000000000000000000aa")
	})

	It("returns an error for an empty flow", func() {
		flow := NewFlow(user, WithChain(config.Base))

		plan, err := flow.Build(ctx, nil)

		Expect(err).To(MatchError(ContainSubstring("empty flow")))
		Expect(plan).To(BeNil())
	})

	It("requires a non-zero account", func() {
		flow := NewFlow(common.Address{}, WithChain(config.Base)).
			Add(&fakeFlowStep{calls: []Call{{Target: common.HexToAddress("0x1")}}})

		plan, err := flow.Build(ctx, nil)

		Expect(err).To(MatchError(ContainSubstring("account")))
		Expect(plan).To(BeNil())
	})

	It("requires explicit chain context", func() {
		flow := NewFlow(user).
			Add(&fakeFlowStep{calls: []Call{{Target: common.HexToAddress("0x1")}}})

		plan, err := flow.Build(ctx, nil)

		Expect(err).To(MatchError(ContainSubstring("chain")))
		Expect(plan).To(BeNil())
	})

	It("builds calls from steps in insertion order", func() {
		firstTarget := common.HexToAddress("0x0000000000000000000000000000000000000001")
		secondTarget := common.HexToAddress("0x0000000000000000000000000000000000000002")
		first := &fakeFlowStep{
			name: "first",
			calls: []Call{{
				Target: firstTarget,
				Value:  big.NewInt(0),
				Data:   []byte{0x01},
			}},
		}
		second := &fakeFlowStep{
			name: "second",
			calls: []Call{{
				Target: secondTarget,
				Value:  big.NewInt(2),
				Data:   []byte{0x02},
			}},
		}

		plan, err := NewFlow(user, WithChain(config.Base)).
			Add(first).
			Add(second).
			Build(ctx, nil)

		Expect(err).NotTo(HaveOccurred())
		Expect(plan.Account()).To(Equal(user))
		steps := plan.Steps()
		Expect(steps).To(HaveLen(2))
		Expect(steps[0].ID).To(Equal(StepID("first#1")))
		Expect(steps[1].ID).To(Equal(StepID("second#1")))
		Expect(plan.Calls()).To(Equal([]Call{
			{Target: firstTarget, Value: big.NewInt(0), Data: []byte{0x01}},
			{Target: secondTarget, Value: big.NewInt(2), Data: []byte{0x02}},
		}))
		Expect(first.seenEnvs).To(HaveLen(1))
		Expect(first.seenEnvs[0].Account).To(Equal(user))
		Expect(first.seenEnvs[0].Chain).To(Equal(config.Base))
		Expect(second.seenEnvs).To(HaveLen(1))
	})

	It("wraps step build errors with step context", func() {
		boom := errors.New("boom")
		step := &fakeFlowStep{name: "failing-step", err: boom}

		plan, err := NewFlow(user, WithChain(config.Base)).
			Add(step).
			Build(ctx, nil)

		Expect(plan).To(BeNil())
		Expect(err).To(MatchError(ContainSubstring("build flow step 1 failing-step")))
		Expect(errors.Is(err, boom)).To(BeTrue())
	})

	It("requires every built step to have a name", func() {
		plan, err := NewFlow(user, WithChain(config.Base)).
			Add(&fakeFlowStep{calls: []Call{{Target: common.HexToAddress("0x1")}}}).
			Build(ctx, nil)

		Expect(plan).To(BeNil())
		Expect(err).To(MatchError(ContainSubstring("step name is required")))
	})

	It("requires every built step to contain at least one call", func() {
		plan, err := NewFlow(user, WithChain(config.Base)).
			Add(&fakeFlowStep{name: "empty.Step"}).
			Build(ctx, nil)

		Expect(plan).To(BeNil())
		Expect(err).To(MatchError(ContainSubstring("step returned no calls")))
	})

	It("converts existing actions into calls through ActionStep", func() {
		token := common.HexToAddress("0x0000000000000000000000000000000000000010")
		recipient := common.HexToAddress("0x0000000000000000000000000000000000000020")
		action := contract.BuildTransferAction(token, recipient, big.NewInt(100))

		plan, err := NewFlow(user, WithChain(config.Base)).
			Add(ActionStep("erc20.Transfer", action)).
			Build(ctx, nil)

		Expect(err).NotTo(HaveOccurred())
		calls := plan.Calls()
		Expect(calls).To(HaveLen(1))
		Expect(calls[0].Target).To(Equal(token))
		Expect(calls[0].Value.Sign()).To(Equal(0))
		Expect(calls[0].Data).NotTo(BeEmpty())
		Expect(plan.Steps()[0].Expectations).To(BeEmpty())
	})

	It("assigns occurrence-based IDs to repeated step names", func() {
		plan, err := NewFlow(user, WithChain(config.Base)).
			Add(&fakeFlowStep{name: "aave.Supply", calls: []Call{{Target: common.HexToAddress("0x1")}}}).
			Add(&fakeFlowStep{name: "aave.Borrow", calls: []Call{{Target: common.HexToAddress("0x2")}}}).
			Add(&fakeFlowStep{name: "aave.Supply", calls: []Call{{Target: common.HexToAddress("0x3")}}}).
			Build(ctx, nil)

		Expect(err).NotTo(HaveOccurred())
		steps := plan.Steps()
		Expect(steps).To(HaveLen(3))
		Expect(steps[0].ID).To(Equal(StepID("aave.Supply#1")))
		Expect(steps[1].ID).To(Equal(StepID("aave.Borrow#1")))
		Expect(steps[2].ID).To(Equal(StepID("aave.Supply#2")))
	})

})
