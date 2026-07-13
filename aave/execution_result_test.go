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
	binderc20 "github.com/tn606024/defi-simplify/bind/erc20"
	"github.com/tn606024/defi-simplify/config"
	sdkerc20 "github.com/tn606024/defi-simplify/erc20"
)

var _ = Describe("Aave execution receipt parsing", func() {
	var (
		poolABI     *abi.ABI
		tokenABI    *abi.ABI
		expected    *ExecutionExpectation
		receipt     *types.Receipt
		account     common.Address
		pool        common.Address
		supplyAsset common.Address
		borrowAsset common.Address
	)

	BeforeEach(func() {
		var err error
		poolABI, err = bindaave.PoolMetaData.GetAbi()
		Expect(err).NotTo(HaveOccurred())
		tokenABI, err = binderc20.Erc20MetaData.GetAbi()
		Expect(err).NotTo(HaveOccurred())

		account = common.HexToAddress("0x00000000000000000000000000000000000000aa")
		pool = common.HexToAddress("0x00000000000000000000000000000000000000bb")
		supplyAsset = common.HexToAddress("0x00000000000000000000000000000000000000cc")
		borrowAsset = common.HexToAddress("0x00000000000000000000000000000000000000dd")
		expected = &ExecutionExpectation{
			Account:      account,
			Pool:         pool,
			SupplyAsset:  supplyAsset,
			SupplyAmount: big.NewInt(10_000_000),
			BorrowAsset:  borrowAsset,
			BorrowAmount: big.NewInt(1_000_000_000_000),
		}
		receipt = &types.Receipt{
			Status:      types.ReceiptStatusSuccessful,
			TxHash:      common.HexToHash("0x1234"),
			BlockHash:   common.HexToHash("0x5678"),
			BlockNumber: big.NewInt(42),
		}
	})

	It("resolves an expectation from the same chain, coin, and decimal inputs as the Flow", func() {
		resolved, err := NewExecutionExpectation(
			config.Base,
			account,
			config.USDC,
			decimal.NewFromInt(10),
			config.WETH,
			decimal.NewFromInt(1).Shift(-6),
		)

		Expect(err).NotTo(HaveOccurred())
		Expect(resolved.Account).To(Equal(account))
		Expect(resolved.Pool).To(Equal(mustAavePoolAddress()))
		Expect(resolved.SupplyAsset).To(Equal(mustCoinAddress(config.USDC)))
		Expect(resolved.SupplyAmount).To(Equal(big.NewInt(10_000_000)))
		Expect(resolved.BorrowAsset).To(Equal(mustCoinAddress(config.WETH)))
		Expect(resolved.BorrowAmount).To(Equal(big.NewInt(1_000_000_000_000)))
	})

	It("decodes and validates the expected approval, supply, and borrow events", func() {
		receipt.Logs = []*types.Log{
			{
				Address: common.HexToAddress("0x00000000000000000000000000000000000000ee"),
				Topics:  []common.Hash{poolABI.Events["Supply"].ID},
				Data:    []byte{0x01},
			},
			approvalLog(tokenABI.Events["Approval"], supplyAsset, account, pool, expected.SupplyAmount, 2),
			supplyLog(poolABI.Events["Supply"], pool, supplyAsset, account, account, expected.SupplyAmount, 3),
			approvalLog(tokenABI.Events["Approval"], supplyAsset, account, pool, big.NewInt(0), 4),
			borrowLog(poolABI.Events["Borrow"], pool, borrowAsset, account, account, expected.BorrowAmount, variableInterestRateMode, big.NewInt(123), 5),
		}

		summary, err := ParseExecutionReceipt(receipt, expected)

		Expect(err).NotTo(HaveOccurred())
		Expect(summary.TransactionHash).To(Equal(receipt.TxHash))
		Expect(summary.BlockHash).To(Equal(receipt.BlockHash))
		Expect(summary.BlockNumber).To(Equal(big.NewInt(42)))
		Expect(summary.Approval).To(Equal(ApprovalResult{
			Token: supplyAsset, Owner: account, Spender: pool, Amount: expected.SupplyAmount, LogIndex: 2,
		}))
		Expect(summary.Supply).To(Equal(SupplyResult{
			Asset: supplyAsset, User: account, OnBehalfOf: account, Amount: expected.SupplyAmount, LogIndex: 3,
		}))
		Expect(summary.Borrow).To(Equal(BorrowResult{
			Asset: borrowAsset, User: account, OnBehalfOf: account, Amount: expected.BorrowAmount,
			InterestRateMode: variableInterestRateMode, BorrowRate: big.NewInt(123), LogIndex: 5,
		}))
	})

	It("returns a clear error when an expected event is missing", func() {
		receipt.Logs = []*types.Log{
			approvalLog(tokenABI.Events["Approval"], supplyAsset, account, pool, expected.SupplyAmount, 1),
			supplyLog(poolABI.Events["Supply"], pool, supplyAsset, account, account, expected.SupplyAmount, 2),
		}

		summary, err := ParseExecutionReceipt(receipt, expected)

		Expect(summary).To(BeNil())
		Expect(errors.Is(err, ErrExpectedEventNotFound)).To(BeTrue())
		Expect(err).To(MatchError(ContainSubstring("Borrow matched 0 of 0 candidate logs")))
	})

	It("reports candidate events whose fields do not match the expectation", func() {
		receipt.Logs = []*types.Log{
			approvalLog(tokenABI.Events["Approval"], supplyAsset, account, pool, expected.SupplyAmount, 1),
			supplyLog(poolABI.Events["Supply"], pool, supplyAsset, account, account, big.NewInt(9_000_000), 2),
			borrowLog(poolABI.Events["Borrow"], pool, borrowAsset, account, account, expected.BorrowAmount, variableInterestRateMode, big.NewInt(123), 3),
		}

		summary, err := ParseExecutionReceipt(receipt, expected)

		Expect(summary).To(BeNil())
		Expect(errors.Is(err, ErrExpectedEventNotFound)).To(BeTrue())
		Expect(err).To(MatchError(ContainSubstring("Supply matched 0 of 1 candidate logs")))
		Expect(err).To(MatchError(ContainSubstring(expected.SupplyAmount.String())))
	})

	It("rejects malformed logs from an expected contract", func() {
		receipt.Logs = []*types.Log{
			approvalLog(tokenABI.Events["Approval"], supplyAsset, account, pool, expected.SupplyAmount, 1),
			{
				Address: pool,
				Topics: []common.Hash{
					poolABI.Events["Supply"].ID,
					addressTopic(supplyAsset),
					addressTopic(account),
					uintTopic(0),
				},
				Data:  []byte{0x01},
				Index: 2,
			},
		}

		summary, err := ParseExecutionReceipt(receipt, expected)

		Expect(summary).To(BeNil())
		Expect(errors.Is(err, ErrMalformedExecutionEvent)).To(BeTrue())
		Expect(err).To(MatchError(ContainSubstring("decode Supply log 2")))
	})

	It("rejects unsuccessful receipts before parsing logs", func() {
		receipt.Status = types.ReceiptStatusFailed

		summary, err := ParseExecutionReceipt(receipt, expected)

		Expect(summary).To(BeNil())
		Expect(errors.Is(err, ErrInvalidExecutionReceipt)).To(BeTrue())
		Expect(err).To(MatchError(ContainSubstring(receipt.TxHash.Hex())))
	})

	It("rejects incomplete expectations", func() {
		expected.Account = common.Address{}

		summary, err := ParseExecutionReceipt(receipt, expected)

		Expect(summary).To(BeNil())
		Expect(errors.Is(err, ErrInvalidExecutionExpectation)).To(BeTrue())
		Expect(err).To(MatchError(ContainSubstring("account is zero")))
	})
})

var _ = Describe("Aave event expectations", func() {
	var (
		poolABI     *abi.ABI
		tokenABI    *abi.ABI
		account     common.Address
		pool        common.Address
		supplyAsset common.Address
		borrowAsset common.Address
		plan        *defi.ExecutionPlan
	)

	BeforeEach(func() {
		var err error
		poolABI, err = bindaave.PoolMetaData.GetAbi()
		Expect(err).NotTo(HaveOccurred())
		tokenABI, err = binderc20.Erc20MetaData.GetAbi()
		Expect(err).NotTo(HaveOccurred())
		Expect(supplyEventTopic).To(Equal(poolABI.Events["Supply"].ID))
		Expect(borrowEventTopic).To(Equal(poolABI.Events["Borrow"].ID))

		account = common.HexToAddress("0x00000000000000000000000000000000000000aa")
		pool, err = config.Base.AaveV3PoolAddress()
		Expect(err).NotTo(HaveOccurred())
		supplyAsset, err = config.USDC.Address(config.Base)
		Expect(err).NotTo(HaveOccurred())
		borrowAsset, err = config.WETH.Address(config.Base)
		Expect(err).NotTo(HaveOccurred())
		plan, err = defi.NewFlow(account, defi.WithChain(config.Base)).
			Add(sdkerc20.Approve(config.USDC, PoolSpender(), decimal.NewFromInt(10))).
			Add(Supply(config.USDC, decimal.NewFromInt(10))).
			Add(Borrow(config.WETH, decimal.NewFromInt(1).Shift(-6))).
			Build(context.Background(), nil)
		Expect(err).NotTo(HaveOccurred())
	})

	It("validates protocol events while skipping decoded field mismatches", func() {
		supplyAmount := big.NewInt(10_000_000)
		borrowAmount := big.NewInt(1_000_000_000_000)
		wrongUser := common.HexToAddress("0x00000000000000000000000000000000000000ff")
		receipt := &types.Receipt{
			Status:      types.ReceiptStatusSuccessful,
			TxHash:      common.HexToHash("0x1234"),
			BlockNumber: big.NewInt(42),
			Logs: []*types.Log{
				approvalLog(tokenABI.Events["Approval"], supplyAsset, account, pool, big.NewInt(0), 1),
				approvalLog(tokenABI.Events["Approval"], supplyAsset, account, pool, supplyAmount, 2),
				supplyLog(poolABI.Events["Supply"], pool, supplyAsset, wrongUser, account, big.NewInt(9_000_000), 3),
				supplyLog(poolABI.Events["Supply"], pool, supplyAsset, account, account, supplyAmount, 4),
				borrowLog(poolABI.Events["Borrow"], pool, borrowAsset, account, account, borrowAmount, VariableInterestRateMode, big.NewInt(123), 5),
			},
		}

		result, err := defi.ValidateExecution(plan, receipt)

		Expect(err).NotTo(HaveOccurred())
		Expect(result.Steps[0].Expectations[0].CandidateCount).To(Equal(2))
		Expect(result.Steps[1].Expectations[0].CandidateCount).To(Equal(2))
		Expect(result.Steps[1].Expectations[0].Mismatches).To(HaveLen(2))
		approvals := defi.EventsOf[*sdkerc20.ApprovalEvent](result)
		supplies := defi.EventsOf[*SupplyEvent](result)
		borrows := defi.EventsOf[*BorrowEvent](result)
		Expect(approvals).To(HaveLen(1))
		Expect(supplies).To(HaveLen(1))
		Expect(supplies[0].Asset).To(Equal(supplyAsset))
		Expect(supplies[0].User).To(Equal(account))
		Expect(supplies[0].OnBehalfOf).To(Equal(account))
		Expect(supplies[0].Amount).To(Equal(supplyAmount))
		Expect(borrows).To(HaveLen(1))
		Expect(borrows[0].Asset).To(Equal(borrowAsset))
		Expect(borrows[0].User).To(Equal(account))
		Expect(borrows[0].Amount).To(Equal(borrowAmount))
		Expect(borrows[0].InterestRateMode).To(Equal(VariableInterestRateMode))
	})

	It("hard fails when the expected pool emits a malformed event", func() {
		receipt := &types.Receipt{
			Status:      types.ReceiptStatusSuccessful,
			BlockNumber: big.NewInt(42),
			Logs: []*types.Log{
				approvalLog(tokenABI.Events["Approval"], supplyAsset, account, pool, big.NewInt(10_000_000), 1),
				{
					Address: pool,
					Topics: []common.Hash{
						poolABI.Events["Supply"].ID,
						addressTopic(supplyAsset),
						addressTopic(account),
						uintTopic(0),
					},
					Data:  []byte{0x01},
					Index: 2,
				},
			},
		}

		result, err := defi.ValidateExecution(plan, receipt)

		Expect(result).NotTo(BeNil())
		Expect(errors.Is(err, defi.ErrMalformedExecutionEvent)).To(BeTrue())
		Expect(result.Steps[0].Status).To(Equal(defi.ValidationValidated))
		Expect(result.Steps[1].Status).To(Equal(defi.ValidationFailed))
		Expect(result.Steps[2].Status).To(Equal(defi.ValidationSkipped))
	})
})

func approvalLog(event abi.Event, token, owner, spender common.Address, amount *big.Int, index uint) *types.Log {
	return &types.Log{
		Address: token,
		Topics:  []common.Hash{event.ID, addressTopic(owner), addressTopic(spender)},
		Data:    mustPackEventData(event, amount),
		Index:   index,
	}
}

func supplyLog(event abi.Event, pool, asset, user, onBehalfOf common.Address, amount *big.Int, index uint) *types.Log {
	return &types.Log{
		Address: pool,
		Topics: []common.Hash{
			event.ID,
			addressTopic(asset),
			addressTopic(onBehalfOf),
			uintTopic(0),
		},
		Data:  mustPackEventData(event, user, amount),
		Index: index,
	}
}

func borrowLog(
	event abi.Event,
	pool common.Address,
	asset common.Address,
	user common.Address,
	onBehalfOf common.Address,
	amount *big.Int,
	interestRateMode uint8,
	borrowRate *big.Int,
	index uint,
) *types.Log {
	return &types.Log{
		Address: pool,
		Topics: []common.Hash{
			event.ID,
			addressTopic(asset),
			addressTopic(onBehalfOf),
			uintTopic(0),
		},
		Data:  mustPackEventData(event, user, amount, interestRateMode, borrowRate),
		Index: index,
	}
}

func mustPackEventData(event abi.Event, values ...interface{}) []byte {
	data, err := event.Inputs.NonIndexed().Pack(values...)
	Expect(err).NotTo(HaveOccurred())
	return data
}

func addressTopic(address common.Address) common.Hash {
	return common.BytesToHash(address.Bytes())
}

func uintTopic(value uint64) common.Hash {
	return common.BigToHash(new(big.Int).SetUint64(value))
}

func mustAavePoolAddress() common.Address {
	address, err := config.Base.AaveV3PoolAddress()
	Expect(err).NotTo(HaveOccurred())
	return address
}

func mustCoinAddress(coin config.Coin) common.Address {
	address, err := coin.Address(config.Base)
	Expect(err).NotTo(HaveOccurred())
	return address
}
