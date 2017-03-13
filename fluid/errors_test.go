package fluid_test

import (
	"errors"

	. "github.com/bbengfort/fluidfs/fluid"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Errors", func() {

	It("should create an error with formatting", func() {
		err := Errorf("this is test %d with %s", 0, "", 8, "aplomb")
		Ω(err.Error()).Should(Equal("this is test 8 with aplomb"))
	})

	It("should wrap another error with formatting", func() {
		erra := errors.New("something bad happened")
		errb := WrapError("trouble with %s", 0, "", erra, "our code")
		Ω(errb.Error()).Should(Equal("trouble with our code: something bad happened"))
	})

	It("should accept a prefix and a code", func() {
		err := Errorf("it's probably influenza", 21, "coming down with something: ")
		Ω(err.Error()).Should(Equal("coming down with something: it's probably influenza"))
	})

	It("should create an improperly configured error", func() {
		err := ImproperlyConfigured("could not find path: %s", "test/foo")
		Ω(err.(*Error).Code).Should(Equal(ErrImproperlyConfigured))
		Ω(err.Error()).Should(Equal("Improperly configured: could not find path: test/foo"))
	})

})
