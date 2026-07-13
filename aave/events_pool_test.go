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

var _ = Describe("Aave repay and withdraw event expectations", func() {
	It("decodes events while skipping mismatched candidates", func() {
		poolABI, err := bindaave.PoolMetaData.GetAbi()
		Expect(err).NotTo(HaveOccurred())
		Expect(repayEventTopic).To(Equal(poolABI.Events["Repay"].ID))
		Expect(withdrawEventTopic).To(Equal(poolABI.Events["Withdraw"].ID))

		account := common.HexToAddress("0x00000000000000000000000000000000000000aa")
		wrongRepayer := common.HexToAddress("0x00000000000000000000000000000000000000bb")
		pool, err := config.Base.AaveV3PoolAddress()
		Expect(err).NotTo(HaveOccurred())
		asset, err := config.USDC.Address(config.Base)
		Expect(err).NotTo(HaveOccurred())

		plan, err := defi.NewFlow(account, defi.WithChain(config.Base)).
			Add(Repay(config.USDC, decimal.NewFromInt(10))).
			Add(Withdraw(config.USDC, decimal.NewFromInt(5))).
			Build(context.Background(), nil)
		Expect(err).NotTo(HaveOccurred())

		repayAmount := big.NewInt(8_000_000)
		withdrawAmount := big.NewInt(5_000_000)
		receipt := &types.Receipt{
			Status:      types.ReceiptStatusSuccessful,
			TxHash:      common.HexToHash("0x1234"),
			BlockNumber: big.NewInt(42),
			Logs: []*types.Log{
				repayLog(poolABI.Events["Repay"], pool, asset, account, wrongRepayer, repayAmount, false, 1),
				repayLog(poolABI.Events["Repay"], pool, asset, account, account, repayAmount, false, 2),
				withdrawLog(poolABI.Events["Withdraw"], pool, asset, account, account, withdrawAmount, 3),
			},
		}

		result, err := defi.ValidateExecution(plan, receipt)

		Expect(err).NotTo(HaveOccurred())
		Expect(result.Steps[0].Expectations[0].CandidateCount).To(Equal(2))
		Expect(result.Steps[0].Expectations[0].Mismatches).NotTo(BeEmpty())
		repayments := defi.EventsOf[*RepayEvent](result)
		withdrawals := defi.EventsOf[*WithdrawEvent](result)
		Expect(repayments).To(HaveLen(1))
		Expect(repayments[0].Amount).To(Equal(repayAmount))
		Expect(repayments[0].Repayer).To(Equal(account))
		Expect(withdrawals).To(HaveLen(1))
		Expect(withdrawals[0].To).To(Equal(account))
	})
})

func repayLog(
	event abi.Event,
	pool,
	asset,
	user,
	repayer common.Address,
	amount *big.Int,
	useATokens bool,
	index uint,
) *types.Log {
	return &types.Log{
		Address: pool,
		Topics: []common.Hash{
			event.ID,
			addressTopic(asset),
			addressTopic(user),
			addressTopic(repayer),
		},
		Data:  mustPackEventData(event, amount, useATokens),
		Index: index,
	}
}

func withdrawLog(
	event abi.Event,
	pool,
	asset,
	user,
	to common.Address,
	amount *big.Int,
	index uint,
) *types.Log {
	return &types.Log{
		Address: pool,
		Topics: []common.Hash{
			event.ID,
			addressTopic(asset),
			addressTopic(user),
			addressTopic(to),
		},
		Data:  mustPackEventData(event, amount),
		Index: index,
	}
}
