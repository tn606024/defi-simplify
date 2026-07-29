package weth

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestWETH(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "WETH Suite")
}
