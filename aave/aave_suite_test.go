package aave

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAave(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Aave Flow Step Suite")
}
