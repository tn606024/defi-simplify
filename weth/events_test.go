package weth

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/shopspring/decimal"
	defi "github.com/tn606024/defi-simplify"
	txamount "github.com/tn606024/defi-simplify/amount"
	"github.com/tn606024/defi-simplify/assets/base"
	"github.com/tn606024/defi-simplify/config"
	"github.com/tn606024/defi-simplify/weth/bindings"
)

var _ = Describe("WETH event expectations", func() {
	It("reports nil and unknown expectations accurately", func() {
		var nilExpectation *eventExpectation
		Expect(nilExpectation.ExpectationName()).To(BeEmpty())
		_, err := nilExpectation.Match(nil, defi.MatchContext{})
		Expect(err).To(MatchError(ContainSubstring("expectation is nil")))

		unknown := &eventExpectation{kind: eventKind(255)}
		Expect(unknown.ExpectationName()).To(Equal("weth.<unknown>"))
	})

	It("decodes and validates ordered Deposit and Withdrawal events", func() {
		wethABI, err := bindings.WETHMetaData.GetAbi()
		Expect(err).NotTo(HaveOccurred())
		Expect(depositEventTopic).To(Equal(wethABI.Events["Deposit"].ID))
		Expect(withdrawalEventTopic).To(Equal(wethABI.Events["Withdrawal"].ID))

		account := common.HexToAddress("0x00000000000000000000000000000000000000aa")
		wrapAmount := decimal.RequireFromString("1.5")
		unwrapAmount := decimal.RequireFromString("0.25")
		wrapWei := decimal.RequireFromString("1.5e18").BigInt()
		unwrapWei := decimal.RequireFromString("0.25e18").BigInt()
		plan, err := defi.NewFlow(account, defi.WithChain(config.Base)).
			Add(Wrap(base.WETH, txamount.Exact(wrapAmount))).
			Add(Unwrap(base.WETH, txamount.Exact(unwrapAmount))).
			Build(context.Background(), nil)
		Expect(err).NotTo(HaveOccurred())

		receipt := &types.Receipt{
			Status:      types.ReceiptStatusSuccessful,
			BlockNumber: big.NewInt(42),
			Logs: []*types.Log{
				wethEventLog(wethABI.Events["Deposit"], base.WETH.Address(), account, big.NewInt(1), 1),
				wethEventLog(wethABI.Events["Deposit"], base.WETH.Address(), account, wrapWei, 2),
				wethEventLog(wethABI.Events["Withdrawal"], base.WETH.Address(), account, unwrapWei, 3),
			},
		}

		result, err := defi.ValidateExecution(plan, receipt)

		Expect(err).NotTo(HaveOccurred())
		Expect(result.Steps[0].Expectations[0].CandidateCount).To(Equal(2))
		Expect(result.Steps[0].Expectations[0].Mismatches).To(HaveLen(1))
		deposits := defi.EventsOf[*DepositEvent](result)
		withdrawals := defi.EventsOf[*WithdrawalEvent](result)
		Expect(deposits).To(HaveLen(1))
		Expect(deposits[0].Contract).To(Equal(base.WETH.Address()))
		Expect(deposits[0].Account).To(Equal(account))
		Expect(deposits[0].Amount).To(Equal(wrapWei))
		Expect(withdrawals).To(HaveLen(1))
		Expect(withdrawals[0].Contract).To(Equal(base.WETH.Address()))
		Expect(withdrawals[0].Account).To(Equal(account))
		Expect(withdrawals[0].Amount).To(Equal(unwrapWei))
	})

	It("accepts a positive runtime withdrawal amount", func() {
		wethABI, err := bindings.WETHMetaData.GetAbi()
		Expect(err).NotTo(HaveOccurred())
		account := common.HexToAddress("0x00000000000000000000000000000000000000aa")
		plan, err := defi.NewFlow(account, defi.WithChain(config.Base)).
			Add(Unwrap(base.WETH, txamount.CurrentBalance(base.WETH))).
			Build(context.Background(), nil)
		Expect(err).NotTo(HaveOccurred())

		receipt := &types.Receipt{
			Status:      types.ReceiptStatusSuccessful,
			BlockNumber: big.NewInt(42),
			Logs: []*types.Log{
				wethEventLog(
					wethABI.Events["Withdrawal"],
					base.WETH.Address(),
					account,
					big.NewInt(7),
					1,
				),
			},
		}

		result, err := defi.ValidateExecution(plan, receipt)

		Expect(err).NotTo(HaveOccurred())
		withdrawals := defi.EventsOf[*WithdrawalEvent](result)
		Expect(withdrawals).To(HaveLen(1))
		Expect(withdrawals[0].Amount).To(Equal(big.NewInt(7)))
	})
})

func wethEventLog(
	event abi.Event,
	contract common.Address,
	account common.Address,
	amount *big.Int,
	index uint,
) *types.Log {
	data, err := event.Inputs.NonIndexed().Pack(amount)
	Expect(err).NotTo(HaveOccurred())
	return &types.Log{
		Address: contract,
		Topics: []common.Hash{
			event.ID,
			common.BytesToHash(account.Bytes()),
		},
		Data:  data,
		Index: index,
	}
}
