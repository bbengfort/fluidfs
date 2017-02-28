package db_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/bbengfort/fluidfs/fluid"
	. "github.com/bbengfort/fluidfs/fluid/db"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Database", func() {

	var err error
	var tmpDir string
	var config *fluid.DatabaseConfig

	BeforeEach(func() {
		tmpDir, err = ioutil.TempDir("", TempDirPrefix)
		Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))

		config = new(fluid.DatabaseConfig)
		config.Defaults()
		config.Path = filepath.Join(tmpDir, "test.db")
		Ω(config.Validate()).Should(BeNil())
	})

	AfterEach(func() {
		err = os.RemoveAll(tmpDir)
		Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))
	})

	It("should create a BoltDB on demand", func() {
		config.Driver = BoltDBDriver
		db, err := InitDatabase(config)
		Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))

		defer func() {
			err := db.Close()
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))
		}()

		_, ok := db.(*BoltDB)
		Ω(ok).Should(BeTrue())
	})

	It("should create a LevelDB on demand", func() {
		config.Driver = LevelDBDriver
		db, err := InitDatabase(config)
		Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))

		defer func() {
			err := db.Close()
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))
		}()

		_, ok := db.(*LevelDB)
		Ω(ok).Should(BeTrue())
	})

})
