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

var _ = Describe("LevelDB Driver", func() {

	var err error
	var db Database
	var tmpDir string
	var config *fluid.DatabaseConfig

	BeforeEach(func() {
		tmpDir, err = ioutil.TempDir("", TempDirPrefix)
		Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))

		config = new(fluid.DatabaseConfig)
		config.Defaults()
		config.Path = filepath.Join(tmpDir, "test.db")
		config.Driver = LevelDBDriver
		Ω(config.Validate()).Should(BeNil())

		db, err = InitDatabase(config)
		Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))
	})

	AfterEach(func() {
		err := db.Close()
		Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))

		err = os.RemoveAll(tmpDir)
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

	Context("initialized database", func() {

		BeforeEach(func() {
			for _, name := range []string{NamesBucket, VersionsBucket, PrefixesBucket} {
				path := filepath.Join("testdata", fmt.Sprintf("%s.json", name))
				err := loadDBFixture(db, name, path)
				Ω(err).ShouldNot(HaveOccurred())
			}
		})

		It("should be able to count a bucket", func() {
			Ω(db.Count(NamesBucket)).Should(Equal(uint64(43)))
			Ω(db.Count(PrefixesBucket)).Should(Equal(uint64(10)))
			Ω(db.Count(VersionsBucket)).Should(Equal(uint64(33)))
		})

		It("should be able to scan through key/value pairs", func() {
			cursor := db.Scan(nil, NamesBucket)
			count := 0

			for cursor.Next() {
				_ = cursor.Pair()
				count++
			}

			Ω(cursor.Error()).ShouldNot(HaveOccurred())
			Ω(count).Should(Equal(43))
		})

		It("should be able to scan a specific prefix", func() {
			cursor := db.Scan([]byte("/~bbengfort/sequence"), NamesBucket)
			count := 0

			for cursor.Next() {
				_ = cursor.Pair()
				count++
			}

			Ω(cursor.Error()).ShouldNot(HaveOccurred())
			Ω(count).Should(Equal(7))
		})

	})

})
