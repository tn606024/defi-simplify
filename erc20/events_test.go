package erc20

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
	binderc20 "github.com/tn606024/defi-simplify/bind/erc20"
	"github.com/tn606024/defi-simplify/config"
)

var _ = Describe("ERC20 event expectations", func() {
	It("reports nil and unknown expectation implementations accurately", func() {
		var nilExpectation *eventExpectation
		Expect(nilExpectation.ExpectationName()).To(BeEmpty())
		_, err := nilExpectation.Match(nil, defi.MatchContext{})
		Expect(err).To(MatchError(ContainSubstring("expectation is nil")))

		unknown := &eventExpectation{kind: eventKind(255)}
		Expect(unknown.ExpectationName()).To(Equal("erc20.<unknown>"))
	})

	It("decodes and validates Approval and Transfer events from a built Flow", func() {
		tokenABI, err := binderc20.Erc20MetaData.GetAbi()
		Expect(err).NotTo(HaveOccurred())
		Expect(approvalEventTopic).To(Equal(tokenABI.Events["Approval"].ID))
		Expect(transferEventTopic).To(Equal(tokenABI.Events["Transfer"].ID))

		account := common.HexToAddress("0x00000000000000000000000000000000000000aa")
		spender := common.HexToAddress("0x00000000000000000000000000000000000000bb")
		recipient := common.HexToAddress("0x00000000000000000000000000000000000000cc")
		token, err := config.USDC.Address(config.Base)
		Expect(err).NotTo(HaveOccurred())
		plan, err := defi.NewFlow(account, defi.WithChain(config.Base)).
			Add(Approve(config.USDC, AddressSpender(spender), decimal.NewFromInt(10))).
			Add(Transfer(config.USDC, recipient, decimal.NewFromInt(2))).
			Build(context.Background(), nil)
		Expect(err).NotTo(HaveOccurred())

		receipt := &types.Receipt{
			Status:      types.ReceiptStatusSuccessful,
			BlockNumber: big.NewInt(42),
			Logs: []*types.Log{
				erc20ApprovalLog(tokenABI.Events["Approval"], token, account, spender, big.NewInt(0), 1),
				erc20ApprovalLog(tokenABI.Events["Approval"], token, account, spender, big.NewInt(10_000_000), 2),
				erc20TransferLog(tokenABI.Events["Transfer"], token, account, recipient, big.NewInt(2_000_000), 3),
			},
		}

		result, err := defi.ValidateExecution(plan, receipt)

		Expect(err).NotTo(HaveOccurred())
		Expect(result.Steps[0].Expectations[0].CandidateCount).To(Equal(2))
		Expect(result.Steps[0].Expectations[0].Mismatches).To(HaveLen(1))
		approvals := defi.EventsOf[*ApprovalEvent](result)
		transfers := defi.EventsOf[*TransferEvent](result)
		Expect(approvals).To(HaveLen(1))
		Expect(approvals[0].Owner).To(Equal(account))
		Expect(approvals[0].Spender).To(Equal(spender))
		Expect(approvals[0].Amount).To(Equal(big.NewInt(10_000_000)))
		Expect(transfers).To(HaveLen(1))
		Expect(transfers[0].From).To(Equal(account))
		Expect(transfers[0].To).To(Equal(recipient))
		Expect(transfers[0].Amount).To(Equal(big.NewInt(2_000_000)))
	})
})

func erc20ApprovalLog(event abi.Event, token, owner, spender common.Address, amount *big.Int, index uint) *types.Log {
	return &types.Log{
		Address: token,
		Topics:  []common.Hash{event.ID, erc20AddressTopic(owner), erc20AddressTopic(spender)},
		Data:    erc20PackEventData(event, amount),
		Index:   index,
	}
}

func erc20TransferLog(event abi.Event, token, from, to common.Address, amount *big.Int, index uint) *types.Log {
	return &types.Log{
		Address: token,
		Topics:  []common.Hash{event.ID, erc20AddressTopic(from), erc20AddressTopic(to)},
		Data:    erc20PackEventData(event, amount),
		Index:   index,
	}
}

func erc20PackEventData(event abi.Event, values ...interface{}) []byte {
	data, err := event.Inputs.NonIndexed().Pack(values...)
	Expect(err).NotTo(HaveOccurred())
	return data
}

func erc20AddressTopic(address common.Address) common.Hash {
	return common.BytesToHash(address.Bytes())
}
