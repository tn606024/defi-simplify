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
	name     string
	calls    []Call
	err      error
	seenEnvs []BuildEnv
}

func (s *fakeFlowStep) BuildCalls(ctx context.Context, env BuildEnv) ([]Call, error) {
	s.seenEnvs = append(s.seenEnvs, env)
	if s.err != nil {
		return nil, s.err
	}
	return s.calls, nil
}

func (s *fakeFlowStep) FlowStepName() string {
	return s.name
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

		calls, err := flow.Build(ctx, nil)

		Expect(err).To(MatchError(ContainSubstring("empty flow")))
		Expect(calls).To(BeNil())
	})

	It("requires a non-zero account", func() {
		flow := NewFlow(common.Address{}, WithChain(config.Base)).
			Add(&fakeFlowStep{calls: []Call{{Target: common.HexToAddress("0x1")}}})

		calls, err := flow.Build(ctx, nil)

		Expect(err).To(MatchError(ContainSubstring("account")))
		Expect(calls).To(BeNil())
	})

	It("requires explicit chain context", func() {
		flow := NewFlow(user).
			Add(&fakeFlowStep{calls: []Call{{Target: common.HexToAddress("0x1")}}})

		calls, err := flow.Build(ctx, nil)

		Expect(err).To(MatchError(ContainSubstring("chain")))
		Expect(calls).To(BeNil())
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

		calls, err := NewFlow(user, WithChain(config.Base)).
			Add(first).
			Add(second).
			Build(ctx, nil)

		Expect(err).NotTo(HaveOccurred())
		Expect(calls).To(Equal([]Call{
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

		calls, err := NewFlow(user, WithChain(config.Base)).
			Add(step).
			Build(ctx, nil)

		Expect(calls).To(BeNil())
		Expect(err).To(MatchError(ContainSubstring("build flow step 1 failing-step")))
		Expect(errors.Is(err, boom)).To(BeTrue())
	})

	It("converts existing actions into calls through ActionStep", func() {
		token := common.HexToAddress("0x0000000000000000000000000000000000000010")
		recipient := common.HexToAddress("0x0000000000000000000000000000000000000020")
		action := contract.BuildTransferAction(token, recipient, big.NewInt(100))

		calls, err := NewFlow(user, WithChain(config.Base)).
			Add(ActionStep("erc20.Transfer", action)).
			Build(ctx, nil)

		Expect(err).NotTo(HaveOccurred())
		Expect(calls).To(HaveLen(1))
		Expect(calls[0].Target).To(Equal(token))
		Expect(calls[0].Value.Sign()).To(Equal(0))
		Expect(calls[0].Data).NotTo(BeEmpty())
	})
})
