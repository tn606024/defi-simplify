package aave

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
	bindaave "github.com/tn606024/defi-simplify/bind/aave"
	"github.com/tn606024/defi-simplify/config"
)

var _ = Describe("Aave credit-delegation event expectations", func() {
	It("decodes and validates BorrowAllowanceDelegated", func() {
		debtABI, err := bindaave.DebtTokenBaseMetaData.GetAbi()
		Expect(err).NotTo(HaveOccurred())
		Expect(borrowAllowanceDelegatedEventTopic).To(Equal(debtABI.Events["BorrowAllowanceDelegated"].ID))

		account := common.HexToAddress("0x00000000000000000000000000000000000000aa")
		delegatee := common.HexToAddress("0x00000000000000000000000000000000000000cc")
		_, usdc, _ := stepTestReserves()
		asset := usdc.Underlying().Address()
		debtToken := usdc.VariableDebtToken().Address()

		plan, err := defi.NewFlow(account, defi.WithChain(config.Base)).
			Add(ApproveDelegation(usdc, delegatee, decimal.NewFromInt(2))).
			Build(context.Background(), nil)
		Expect(err).NotTo(HaveOccurred())

		delegationAmount := big.NewInt(2_000_000)
		receipt := &types.Receipt{
			Status:      types.ReceiptStatusSuccessful,
			TxHash:      common.HexToHash("0x1234"),
			BlockNumber: big.NewInt(42),
			Logs: []*types.Log{
				delegationLog(
					debtABI.Events["BorrowAllowanceDelegated"],
					debtToken,
					account,
					delegatee,
					asset,
					delegationAmount,
					1,
				),
			},
		}

		result, err := defi.ValidateExecution(plan, receipt)

		Expect(err).NotTo(HaveOccurred())
		delegations := defi.EventsOf[*BorrowAllowanceDelegatedEvent](result)
		Expect(delegations).To(HaveLen(1))
		Expect(delegations[0].DebtToken).To(Equal(debtToken))
		Expect(delegations[0].Asset).To(Equal(asset))
		Expect(delegations[0].FromUser).To(Equal(account))
		Expect(delegations[0].ToUser).To(Equal(delegatee))
		Expect(delegations[0].Amount).To(Equal(delegationAmount))
	})
})

func delegationLog(
	event abi.Event,
	debtToken,
	fromUser,
	toUser,
	asset common.Address,
	amount *big.Int,
	index uint,
) *types.Log {
	return &types.Log{
		Address: debtToken,
		Topics: []common.Hash{
			event.ID,
			addressTopic(fromUser),
			addressTopic(toUser),
			addressTopic(asset),
		},
		Data:  mustPackEventData(event, amount),
		Index: index,
	}
}
