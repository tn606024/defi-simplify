package contract

import (
	"github.com/ethereum/go-ethereum/common"
	. "github.com/onsi/gomega"
)

func mustAddress(address common.Address, err error) common.Address {
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	return address
}
