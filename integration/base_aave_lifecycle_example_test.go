//go:build integration

package integration

import (
	"context"
	"encoding/hex"
	"math/big"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tn606024/defi-simplify/aave"
	"github.com/tn606024/defi-simplify/client/account/eip7702"
	sdkerc20 "github.com/tn606024/defi-simplify/erc20"
)

var _ = Describe("Base Aave lifecycle example", func() {
	It("completes delegate, open, close, and clear through the public command", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 240*time.Second)
		defer cancel()

		ethClient := baseForkClient(GinkgoT())
		rpcClient := baseForkRPCClient(GinkgoT())
		requireAnvilFork(GinkgoT(), ctx, rpcClient)

		key, err := crypto.GenerateKey()
		Expect(err).NotTo(HaveOccurred())
		account := crypto.PubkeyToAddress(key.PublicKey)
		Expect(setForkETHBalance(
			ctx,
			rpcClient,
			account,
			big.NewInt(2_000_000_000_000_000_000),
		)).To(Succeed())
		privateKey := hex.EncodeToString(crypto.FromECDSA(key))

		delegateOutput := runLifecycleExample(
			ctx,
			privateKey,
			"--delegate",
			"--broadcast",
		)
		Expect(delegateOutput).To(ContainSubstring("phase=delegate broadcast=true"))
		Expect(delegateOutput).To(ContainSubstring("transaction.mined"))
		Expect(delegateOutput).To(ContainSubstring("state.after.delegation status=delegated"))

		openOutput := runLifecycleExample(
			ctx,
			privateKey,
			"--open",
			"--broadcast",
			"--weth-supply",
			"0.1",
			"--usdc-borrow",
			"10",
		)
		Expect(openOutput).To(ContainSubstring("plan account=" + account.Hex() + " kind=static"))
		Expect(openOutput).To(ContainSubstring("event.weth.deposit"))
		Expect(openOutput).To(ContainSubstring("event.erc20.approval"))
		Expect(openOutput).To(ContainSubstring("event.aave.supply"))
		Expect(openOutput).To(ContainSubstring("event.aave.borrow"))
		Expect(openOutput).To(ContainSubstring("user=" + account.Hex()))
		Expect(openOutput).To(ContainSubstring("on_behalf_of=" + account.Hex()))

		market, usdcReserve, wethReserve := loadBaseAaveReserves(GinkgoT(), ctx, ethClient)
		wethPosition, err := aave.LoadUserReserveState(ctx, ethClient, wethReserve, account)
		Expect(err).NotTo(HaveOccurred())
		Expect(wethPosition.CurrentATokenBalance().Sign()).To(Equal(1))
		usdcPosition, err := aave.LoadUserReserveState(ctx, ethClient, usdcReserve, account)
		Expect(err).NotTo(HaveOccurred())
		Expect(usdcPosition.CurrentVariableDebt().Sign()).To(Equal(1))

		implementation := loadDefiSimplifyAccountIdentity(GinkgoT(), ctx, ethClient).Address
		implementationWETH, err := aave.LoadUserReserveState(ctx, ethClient, wethReserve, implementation)
		Expect(err).NotTo(HaveOccurred())
		Expect(implementationWETH.CurrentATokenBalance().Sign()).To(Equal(0))
		implementationUSDC, err := aave.LoadUserReserveState(ctx, ethClient, usdcReserve, implementation)
		Expect(err).NotTo(HaveOccurred())
		Expect(implementationUSDC.CurrentVariableDebt().Sign()).To(Equal(0))

		Expect(fundBaseUSDCFromHolder(
			ctx,
			rpcClient,
			ethClient,
			account,
			big.NewInt(1_000_000),
		)).To(Succeed())

		closeOutput := runLifecycleExample(
			ctx,
			privateKey,
			"--close",
			"--broadcast",
			"--usdc-repay-limit",
			"10.5",
		)
		Expect(closeOutput).To(ContainSubstring("plan account=" + account.Hex() + " kind=dynamic"))
		Expect(closeOutput).To(ContainSubstring("label=before-close-withdraw"))
		Expect(closeOutput).To(ContainSubstring("source=checkpoint_delta"))
		Expect(closeOutput).To(ContainSubstring("event.aave.repay"))
		Expect(closeOutput).To(ContainSubstring("event.aave.withdraw"))
		Expect(closeOutput).To(ContainSubstring("event.weth.withdrawal"))

		wethPosition, err = aave.LoadUserReserveState(ctx, ethClient, wethReserve, account)
		Expect(err).NotTo(HaveOccurred())
		Expect(wethPosition.CurrentATokenBalance().Sign()).To(Equal(0))
		usdcPosition, err = aave.LoadUserReserveState(ctx, ethClient, usdcReserve, account)
		Expect(err).NotTo(HaveOccurred())
		Expect(usdcPosition.CurrentStableDebt().Sign()).To(Equal(0))
		Expect(usdcPosition.CurrentVariableDebt().Sign()).To(Equal(0))
		wethToken, err := sdkerc20.LoadAccountState(
			ctx,
			ethClient,
			wethReserve.Underlying(),
			account,
			market.Pool(),
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(wethToken.Allowance().Sign()).To(Equal(0))
		usdcToken, err := sdkerc20.LoadAccountState(
			ctx,
			ethClient,
			usdcReserve.Underlying(),
			account,
			market.Pool(),
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(usdcToken.Allowance().Sign()).To(Equal(0))

		clearOutput := runLifecycleExample(
			ctx,
			privateKey,
			"--clear",
			"--broadcast",
		)
		Expect(clearOutput).To(ContainSubstring("phase=clear broadcast=true"))
		Expect(clearOutput).To(ContainSubstring("state.after.delegation status=clean"))
		delegation, err := eip7702.ReadPendingDelegationState(ctx, ethClient, account)
		Expect(err).NotTo(HaveOccurred())
		Expect(delegation.Clean()).To(BeTrue())
	})
})

func runLifecycleExample(ctx context.Context, privateKey string, args ...string) string {
	GinkgoHelper()
	commandArgs := append([]string{"run", "./examples/base-aave-lifecycle"}, args...)
	command := exec.CommandContext(ctx, "go", commandArgs...)
	command.Dir = ".."
	command.Env = lifecycleExampleEnvironment(privateKey)
	output, err := command.CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), "command output:\n%s", string(output))
	return string(output)
}

func lifecycleExampleEnvironment(privateKey string) []string {
	environment := make([]string, 0, len(os.Environ())+2)
	for _, value := range os.Environ() {
		if strings.HasPrefix(value, "BASE_RPC_URL=") ||
			strings.HasPrefix(value, "PRIVATE_KEY=") {
			continue
		}
		environment = append(environment, value)
	}
	return append(
		environment,
		"BASE_RPC_URL=http://127.0.0.1:8545",
		"PRIVATE_KEY="+privateKey,
	)
}
