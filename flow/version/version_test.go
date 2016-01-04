package version_test

import (
	. "github.com/bbengfort/flow/flow/version"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Version", func() {

	const (
		ExpectedVersion = "0.1"
	)

	It("should have a version that matches the test version", func() {
		Expect(Version()).To(Equal(ExpectedVersion))
	})

})
