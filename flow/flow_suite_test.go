package flow_test

import (
    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"

    "testing"
)

func TestCrate(t *testing.T) {
    RegisterFailHandler(Fail)
    RunSpecs(t, "Flow Suite")
}
