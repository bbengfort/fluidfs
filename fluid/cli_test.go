package fluid_test

import (
	. "github.com/bbengfort/fluidfs/fluid"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CLI", func() {

	var cli *CLIClient

	BeforeEach(func() {
		cli = new(CLIClient)
	})

	It("should return an error if there is no PID file", func() {
		SkipIfPIDExists()

		err := cli.Init()
		Ω(err).ShouldNot(BeNil())
	})

	It("should return an endpoint constructed from the PID file", func() {
		SkipIfPIDExists()

		cli.Init()
		cli.PID.Port = 3264

		Ω(cli.Endpoint("").String()).Should(Equal("http://localhost:3264"))
		Ω(cli.Endpoint("/").String()).Should(Equal("http://localhost:3264/"))
		Ω(cli.Endpoint("status").String()).Should(Equal("http://localhost:3264/status"))
		Ω(cli.Endpoint("/status").String()).Should(Equal("http://localhost:3264/status"))
		Ω(cli.Endpoint("path", "to", "file.txt").String()).Should(Equal("http://localhost:3264/path/to/file.txt"))
	})

})
