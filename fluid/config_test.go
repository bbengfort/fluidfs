package fluid_test

import (
	. "github.com/bbengfort/fluidfs/fluid"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {

	Describe("loading configuration from disk", func() {

		It("should return four search paths", func() {
			config := new(Config)
			paths := config.Paths()

			Ω(paths).Should(HaveLen(4))
			Ω(paths).Should(HaveCap(4))
		})

		It("should return the etc config path", func() {
			config := new(Config)
			paths := config.Paths()

			Ω(paths).Should(ContainElement("/etc/fluid/fluidfs.yml"))
		})

		It("should read a YAML file from a path", func() {
			Skip("test needs to be implemented")
		})

		It("should load the configuration calling interface methods", func() {
			Skip("test needs to be implemented")
		})

	})

	Describe("fluid configuration interface", func() {

		It("should load defaults when called", func() {
			config := new(Config)

			// Assert that the properties are all false
			Ω(config.Debug).Should(BeZero())
			Ω(config.PID).Should(BeZero())
			Ω(config.Hostname).Should(BeZero())
			Ω(config.Addr).Should(BeZero())
			Ω(config.Logging).Should(BeZero())

			// Run the defaults and assert that default values are set.
			err := config.Defaults()
			Ω(err).Should(BeNil())

			Ω(config.Debug).Should(BeFalse())
			Ω(config.Hostname).ShouldNot(BeZero())
			Ω(config.Logging).ShouldNot(BeZero())
		})

		Context("validation after defaults", func() {

			var config *Config

			BeforeEach(func() {
				config = new(Config)
				config.Defaults()
			})

			It("should not allow a zero pid value", func() {
				config.PID = 0
				err := config.Validate()
				Ω(err).ShouldNot(BeNil())
			})

			It("should not allow a null hostname", func() {
				config.PID = 1
				config.Hostname = ""
				err := config.Validate()
				Ω(err).ShouldNot(BeNil())
			})

			It("should validate the logging configuration", func() {
				config.PID = 1
				config.Hostname = "alaska"
				config.Logging.Level = "KLONDIKE"
				err := config.Validate()
				Ω(err).ShouldNot(BeNil())
			})

		})

	})

	Describe("logging configuration interface", func() {

		It("should load defaults when called", func() {
			config := new(LoggingConfig)

			// Assert that config has zero values
			Ω(config.Level).Should(BeZero())
			Ω(config.Path).Should(BeZero())

			// Call defaults and assert default values
			config.Defaults()
			Ω(config.Level).ShouldNot(BeZero())
		})

		Context("validation after defaults", func() {

			var config *LoggingConfig

			BeforeEach(func() {
				config = new(LoggingConfig)
				config.Defaults()
			})

			It("should not allow bad logging levels", func() {
				config.Level = "KODIAC"
				err := config.Validate()
				Ω(err).ShouldNot(BeNil())
			})

			It("should allow good logging levels", func() {
				var levelNames = []string{
					"DEBUG", "INFO", "WARN", "ERROR", "FATAL",
				}

				for _, level := range levelNames {
					config.Level = level
					err := config.Validate()
					Ω(err).Should(BeNil())
				}

			})

		})

	})

})
