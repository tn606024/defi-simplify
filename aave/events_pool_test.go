package aave

import (
	"context"
	"errors"
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

	It("matches actual positive amounts emitted for full-position steps", func() {
		poolABI, err := bindaave.PoolMetaData.GetAbi()
		Expect(err).NotTo(HaveOccurred())
		account := common.HexToAddress("0x00000000000000000000000000000000000000aa")
		pool, err := config.Base.AaveV3PoolAddress()
		Expect(err).NotTo(HaveOccurred())
		repayAsset, err := config.WETH.Address(config.Base)
		Expect(err).NotTo(HaveOccurred())
		withdrawAsset, err := config.USDC.Address(config.Base)
		Expect(err).NotTo(HaveOccurred())

		plan, err := defi.NewFlow(account, defi.WithChain(config.Base)).
			Add(RepayAll(config.WETH)).
			Add(WithdrawAll(config.USDC)).
			Build(context.Background(), nil)
		Expect(err).NotTo(HaveOccurred())

		repayAmount := big.NewInt(1_000_000_000_000)
		withdrawAmount := big.NewInt(10_000_000)
		receipt := &types.Receipt{
			Status:      types.ReceiptStatusSuccessful,
			TxHash:      common.HexToHash("0x5678"),
			BlockNumber: big.NewInt(43),
			Logs: []*types.Log{
				repayLog(poolABI.Events["Repay"], pool, repayAsset, account, account, repayAmount, false, 1),
				withdrawLog(poolABI.Events["Withdraw"], pool, withdrawAsset, account, account, withdrawAmount, 2),
			},
		}

		result, err := defi.ValidateExecution(plan, receipt)

		Expect(err).NotTo(HaveOccurred())
		repayments := defi.EventsOf[*RepayEvent](result)
		withdrawals := defi.EventsOf[*WithdrawEvent](result)
		Expect(repayments).To(HaveLen(1))
		Expect(repayments[0].Amount).To(Equal(repayAmount))
		Expect(repayments[0].Amount).NotTo(Equal(newUint256Max()))
		Expect(withdrawals).To(HaveLen(1))
		Expect(withdrawals[0].Amount).To(Equal(withdrawAmount))
		Expect(withdrawals[0].Amount).NotTo(Equal(newUint256Max()))
	})

	It("rejects a zero actual repayment amount for RepayAll", func() {
		poolABI, err := bindaave.PoolMetaData.GetAbi()
		Expect(err).NotTo(HaveOccurred())
		account := common.HexToAddress("0x00000000000000000000000000000000000000aa")
		pool, err := config.Base.AaveV3PoolAddress()
		Expect(err).NotTo(HaveOccurred())
		asset, err := config.WETH.Address(config.Base)
		Expect(err).NotTo(HaveOccurred())
		plan, err := defi.NewFlow(account, defi.WithChain(config.Base)).
			Add(RepayAll(config.WETH)).
			Build(context.Background(), nil)
		Expect(err).NotTo(HaveOccurred())
		receipt := &types.Receipt{
			Status:      types.ReceiptStatusSuccessful,
			TxHash:      common.HexToHash("0x9abc"),
			BlockNumber: big.NewInt(44),
			Logs: []*types.Log{
				repayLog(poolABI.Events["Repay"], pool, asset, account, account, big.NewInt(0), false, 1),
			},
		}

		result, err := defi.ValidateExecution(plan, receipt)

		Expect(errors.Is(err, defi.ErrExpectedEventNotFound)).To(BeTrue())
		Expect(result).NotTo(BeNil())
		Expect(result.Steps[0].Status).To(Equal(defi.ValidationFailed))
		mismatches := result.Steps[0].Expectations[0].Mismatches
		Expect(mismatches).NotTo(BeEmpty())
		Expect(mismatches[0].Field).To(Equal("amount"))
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
