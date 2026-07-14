package strategy_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestStrategy(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Strategy Suite")
}
