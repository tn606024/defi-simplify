package base_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestBaseAssets(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Base Asset Catalog Suite")
}
