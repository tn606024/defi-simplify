package main

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestLifecycleExample(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Base Aave Lifecycle Example Suite")
}
