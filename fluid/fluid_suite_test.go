package fluid_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestFluid(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Fluid Suite")
}
