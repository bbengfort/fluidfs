package fluid_test

import (
	. "github.com/bbengfort/fluidfs/fluid"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Core", func() {

	Describe("Package Meta", func() {

		const (
			ExpectedVersion = "0.1"
		)

		It("should have a version that matches the test version", func() {
			Expect(PackageVersion()).To(Equal(ExpectedVersion))
		})

	})

})
