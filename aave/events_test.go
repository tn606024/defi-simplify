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

	It("reports nil and unknown expectation implementations accurately", func() {
		var nilExpectation *protocolEventExpectation
		Expect(nilExpectation.ExpectationName()).To(BeEmpty())
		_, err := nilExpectation.Match(nil, defi.MatchContext{})
		Expect(err).To(MatchError(ContainSubstring("expectation is nil")))

		unknown := &protocolEventExpectation{kind: protocolEventKind(255)}
		Expect(unknown.ExpectationName()).To(Equal("aave.<unknown>"))
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
