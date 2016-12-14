package fluid_test

import (
    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"

    "testing"
)

func TestFlow(t *testing.T) {
    RegisterFailHandler(Fail)
    RunSpecs(t, "Fluid Suite")
}
