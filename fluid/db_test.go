package fluid_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/bbengfort/fluidfs/fluid"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Database", func() {

	var err error
	var tmpDir string
	var config *DatabaseConfig

	BeforeEach(func() {
		tmpDir, err = ioutil.TempDir("", TempDirPrefix)
		Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))

		config = new(DatabaseConfig)
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

	Describe("BoltDB Driver", func() {

		var db Database

		BeforeEach(func() {
			config.Driver = BoltDBDriver
			db, err = InitDatabase(config)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))
		})

		AfterEach(func() {
			err := db.Close()
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))
		})

		It("should be able to put a key/value into the names bucket and get it back", func() {
			// Fixtures
			key := []byte("foo")
			val := []byte("bar")

			// Do the Put
			err := db.Put(key, val, NamesBucket)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))

			// Do the Get
			dval, err := db.Get(key, NamesBucket)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))
			Ω(val).Should(Equal(dval))
		})

		It("should be able to put a key/value into the names bucket and delete it", func() {
			// Fixtures
			key := []byte("color")
			val := []byte("purple")

			// Do the Put
			err := db.Put(key, val, NamesBucket)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))

			// Do a Get
			dval, err := db.Get(key, NamesBucket)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))
			Ω(val).Should(Equal(dval))

			// Do the Delete
			err = db.Delete(key, NamesBucket)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))

			// Do the second Get
			dval, err = db.Get(key, NamesBucket)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))
			Ω(dval).Should(BeNil())
		})

		It("should be able to batch insert key/values", func() {
			// Fixtures
			keys := [][]byte{
				[]byte("foo"), []byte("bar"), []byte("baz"),
			}
			vals := [][]byte{
				[]byte("purple"), []byte("green"), []byte("orange"),
			}

			// Perform the Batch insert
			err := db.Batch(keys, vals, NamesBucket)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))

			// Perform Gets to verify
			for idx, key := range keys {
				want := vals[idx]
				got, err := db.Get(key, NamesBucket)
				Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))
				Ω(got).Should(Equal(want))
			}
		})

		It("should error on batch insert with mismatched key/val lengths", func() {
			// Fixtures
			keys := [][]byte{
				[]byte("foo"), []byte("bar"), []byte("baz"),
			}
			vals := [][]byte{
				[]byte("purple"), []byte("green"),
			}

			err := db.Batch(keys, vals, NamesBucket)
			Ω(err).ShouldNot(BeNil())
		})

	})

	Describe("LevelDB Driver", func() {

		var db Database

		BeforeEach(func() {
			config.Driver = LevelDBDriver
			db, err = InitDatabase(config)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))
		})

		AfterEach(func() {
			err := db.Close()
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))
		})

		It("should be able to put a key/value into the names bucket and get it back", func() {
			// Fixtures
			key := []byte("foo")
			val := []byte("bar")

			// Do the Put
			err := db.Put(key, val, NamesBucket)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))

			// Do the Get
			dval, err := db.Get(key, NamesBucket)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))
			Ω(val).Should(Equal(dval))
		})

		It("should be able to put a key/value into the names bucket and delete it", func() {
			// Fixtures
			key := []byte("color")
			val := []byte("purple")

			// Do the Put
			err := db.Put(key, val, NamesBucket)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))

			// Do a Get
			dval, err := db.Get(key, NamesBucket)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))
			Ω(val).Should(Equal(dval))

			// Do the Delete
			err = db.Delete(key, NamesBucket)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))

			// Do the second Get
			dval, err = db.Get(key, NamesBucket)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))
			Ω(dval).Should(BeNil())
		})

		It("should be able to batch insert key/values", func() {
			// Fixtures
			keys := [][]byte{
				[]byte("foo"), []byte("bar"), []byte("baz"),
			}
			vals := [][]byte{
				[]byte("purple"), []byte("green"), []byte("orange"),
			}

			// Perform the Batch insert
			err := db.Batch(keys, vals, NamesBucket)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))

			// Perform Gets to verify
			for idx, key := range keys {
				want := vals[idx]
				got, err := db.Get(key, NamesBucket)
				Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))
				Ω(got).Should(Equal(want))
			}
		})

		It("should error on batch insert with mismatched key/val lengths", func() {
			// Fixtures
			keys := [][]byte{
				[]byte("foo"), []byte("bar"), []byte("baz"),
			}
			vals := [][]byte{
				[]byte("purple"), []byte("green"),
			}

			err := db.Batch(keys, vals, NamesBucket)
			Ω(err).ShouldNot(BeNil())
		})

	})

})
