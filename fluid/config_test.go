package fluid_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

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
			Ω(config.PID).Should(BeZero())
			Ω(config.Name).Should(BeZero())
			Ω(config.Host).Should(BeZero())
			Ω(config.Port).Should(BeZero())
			Ω(config.Logging).Should(BeZero())
			Ω(config.Database).Should(BeZero())
			Ω(config.Storage).Should(BeZero())

			// Run the defaults and assert that default values are set.
			err := config.Defaults()
			Ω(err).Should(BeNil())

			Ω(config.Name).ShouldNot(BeZero())
			Ω(config.Port).ShouldNot(BeZero())
			Ω(config.Logging).ShouldNot(BeZero())
			Ω(config.Database).ShouldNot(BeZero())
			Ω(config.Storage).ShouldNot(BeZero())
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
				config.Name = ""
				err := config.Validate()
				Ω(err).ShouldNot(BeNil())
			})

			It("should validate the logging configuration", func() {
				config.PID = 1
				config.Name = "alaska"
				config.Logging.Level = "KLONDIKE"
				err := config.Validate()
				Ω(err).ShouldNot(BeNil())
			})

			It("should validate the database configuration", func() {
				config.PID = 1
				config.Name = "alaska"
				config.Database.Driver = "JunoDB"
				err := config.Validate()
				Ω(err).ShouldNot(BeNil())
			})

			It("should validate the chunking configuration", func() {
				config.PID = 1
				config.Name = "alaska"
				config.Storage.Chunking = "cloudy"
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

	Describe("database configuration interface", func() {

		It("should load defaults when called", func() {
			config := new(DatabaseConfig)

			// Assert that config has zero values
			Ω(config.Driver).Should(BeZero())
			Ω(config.Path).Should(BeZero())

			// Call defaults and assert default values
			config.Defaults()
			Ω(config.Driver).ShouldNot(BeZero())
			Ω(config.Path).ShouldNot(BeZero())
		})

		Context("validation after defaults", func() {

			var config *DatabaseConfig

			BeforeEach(func() {
				config = new(DatabaseConfig)
				config.Defaults()
			})

			It("should not allow bad database drivers", func() {
				config.Driver = "KODIAC"
				err := config.Validate()
				Ω(err).ShouldNot(BeNil())
			})

			It("should allow good database drivers", func() {
				var driverNames = []string{
					"boltdb", "leveldb",
				}

				for _, driver := range driverNames {
					config.Driver = driver
					err := config.Validate()
					Ω(err).Should(BeNil())
				}

			})

			It("shouldn't allow driver case to matter", func() {
				var driverNames = []string{
					"BoltDB", "levelDB", "BOLTDB", "LEVELdb",
				}

				for _, driver := range driverNames {
					config.Driver = driver
					err := config.Validate()
					Ω(err).Should(BeNil())
				}
			})

			It("should allow driver white space to matter", func() {
				var driverNames = []string{
					"boltdb ", " leveldb   ", " boltdb",
				}

				for _, driver := range driverNames {
					config.Driver = driver
					err := config.Validate()
					Ω(err).Should(BeNil())
				}
			})

			It("should not allow zero database paths", func() {
				config.Path = ""
				Ω(config.Validate()).ShouldNot(BeNil())
			})

		})

	})

	Describe("storage configuration interface", func() {

		It("should load defaults when called", func() {
			config := new(StorageConfig)

			// Assert that config has zero values
			Ω(config.Chunking).Should(BeZero())
			Ω(config.BlockSize).Should(BeZero())
			Ω(config.MinBlockSize).Should(BeZero())
			Ω(config.MaxBlockSize).Should(BeZero())
			Ω(config.Hashing).Should(BeZero())

			// Call defaults and assert default values
			config.Defaults()
			Ω(config.Chunking).ShouldNot(BeZero())
			Ω(config.BlockSize).ShouldNot(BeZero())
			Ω(config.MinBlockSize).ShouldNot(BeZero())
			Ω(config.MaxBlockSize).ShouldNot(BeZero())
			Ω(config.Hashing).ShouldNot(BeZero())
		})

		Context("validation after defaults", func() {

			var config *StorageConfig
			var tempDir string
			var err error

			BeforeEach(func() {
				config = new(StorageConfig)
				config.Defaults()

				tempDir, err = ioutil.TempDir("", TempDirPrefix)
				Ω(err).Should(BeNil())
				config.Path = tempDir
			})

			AfterEach(func() {
				Ω(os.RemoveAll(tempDir)).Should(BeNil())
			})

			It("should not allow zero storage paths", func() {
				config.Path = ""
				Ω(config.Validate()).ShouldNot(BeNil())
			})

			It("should create the storage directory immediately", func() {
				path := filepath.Join(tempDir, "path", "to", "storage")
				_, err := os.Stat(path)
				Ω(os.IsNotExist(err)).Should(BeTrue(), "path existed before testing")

				config.Path = path

				Ω(config.Validate()).Should(BeNil(), "error occurred during validation and should not have")

				info, err := os.Stat(path)
				Ω(os.IsNotExist(err)).Should(BeFalse(), "path does not exist after validation!")
				Ω(info.Mode().IsDir()).Should(BeTrue(), "path is not a directory!")
			})

			It("should not allow bad chunking mechanisms", func() {
				config.Chunking = "cloudy"
				err := config.Validate()
				Ω(err).ShouldNot(BeNil())
			})

			It("should allow good chunking mechanisms", func() {
				var chunkNames = []string{
					"variable", "fixed",
				}

				for _, chunks := range chunkNames {
					config.Chunking = chunks
					err := config.Validate()
					Ω(err).Should(BeNil())
				}

			})

			It("shouldn't allow chunks case to matter", func() {
				var methodNames = []string{
					"Variable", "FIXED", "vaRIAble", "variable",
				}

				for _, method := range methodNames {
					config.Chunking = method
					err := config.Validate()
					Ω(err).Should(BeNil())
				}
			})

			It("should allow chunks white space to matter", func() {
				var methodNames = []string{
					"variable ", "fixed ", " variable ", "    fixed   ",
				}

				for _, method := range methodNames {
					config.Chunking = method
					err := config.Validate()
					Ω(err).Should(BeNil())
				}
			})

			It("should not allow zero block sizes", func() {
				config.MinBlockSize = 0
				config.BlockSize = 0
				err := config.Validate()
				Ω(err).ShouldNot(BeNil())

				config.MinBlockSize = 10
				config.BlockSize = 10
				err = config.Validate()
				Ω(err).Should(BeNil())

				config.MinBlockSize = -1
				config.BlockSize = -1
				err = config.Validate()
				Ω(err).ShouldNot(BeNil())
			})

			It("should not allow maximum block sizes less than the target", func() {
				config.MaxBlockSize = 10
				config.BlockSize = 1000
				err := config.Validate()
				Ω(err).ShouldNot(BeNil())

			})

			It("should not allow maximum block sizes less than the target", func() {
				config.MaxBlockSize = 10
				config.MinBlockSize = 1000
				config.BlockSize = 2000
				err := config.Validate()
				Ω(err).ShouldNot(BeNil())

			})

			It("should not allow minimum block sizes greater than the target", func() {
				config.MinBlockSize = 1000
				config.BlockSize = 100
				err := config.Validate()
				Ω(err).ShouldNot(BeNil())

			})

			It("should not allow bad hashing alogrithms", func() {
				config.Hashing = "protobob"
				err := config.Validate()
				Ω(err).ShouldNot(BeNil())
			})

			It("should allow good hashing alogrithms", func() {
				var chunkNames = []string{
					"md5", "sha1", "sha224", "sha256", "murmur",
				}

				for _, chunks := range chunkNames {
					config.Hashing = chunks
					err := config.Validate()
					Ω(err).Should(BeNil())
				}

			})

			It("shouldn't allow hashing case to matter", func() {
				var driverNames = []string{
					"MD5", "Sha1", "SHA224", "sHA256", "Murmur",
				}

				for _, driver := range driverNames {
					config.Hashing = driver
					err := config.Validate()
					Ω(err).Should(BeNil())
				}
			})

			It("should allow hashing white space to matter", func() {
				var driverNames = []string{
					"md5 ", "sha256 ", " sha224 ", "   murmur   ",
				}

				for _, driver := range driverNames {
					config.Hashing = driver
					err := config.Validate()
					Ω(err).Should(BeNil())
				}
			})

		})

	})
})
