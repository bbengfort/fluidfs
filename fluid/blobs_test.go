package fluid_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/bbengfort/fluidfs/fluid"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Anti-Entropy Blob Replication", func() {

	// Tests for the BlobTree component of Anti-Entropy
	Describe("Blob Tree", func() {

		It("should detect root nodes", func() {

			// Build a simple tree
			root := &BlobTree{Name: "root", Children: make(map[string]*BlobTree)}
			left := &BlobTree{Name: "l", Parent: root}
			rght := &BlobTree{Name: "r", Parent: root}

			root.Children["l"] = left
			root.Children["r"] = rght

			Ω(root.IsRoot()).Should(BeTrue())
			Ω(left.IsRoot()).Should(BeFalse())
			Ω(rght.IsRoot()).Should(BeFalse())

		})

		It("should be able to add files from root", func() {

			// Create a simple tree using AddFile
			root := &BlobTree{Name: "foo", Children: make(map[string]*BlobTree)}
			err := root.AddFile("foo/bar/baz/test.txt")
			Ω(err).ShouldNot(HaveOccurred())

			// Get the bar node, which should exist
			bar, ok := root.Children["bar"]
			Ω(ok).Should(BeTrue())

			// Get the baz node, which should exist
			baz, ok := bar.Children["baz"]
			Ω(ok).Should(BeTrue())

			// There should be no test.txt node
			_, ok = baz.Children["test.txt"]
			Ω(ok).Should(BeFalse())

			// Test the counts and path on each node
			Ω(root.Count).Should(Equal(uint64(1)))
			Ω(bar.Count).Should(Equal(uint64(1)))
			Ω(baz.Count).Should(Equal(uint64(1)))
			Ω(root.Path()).Should(Equal("foo"))
			Ω(bar.Path()).Should(Equal("foo/bar"))
			Ω(baz.Path()).Should(Equal("foo/bar/baz"))

			// Add a new file and assert counts
			err = root.AddFile("foo/bar/zab/test.html")
			Ω(err).ShouldNot(HaveOccurred())

			zab, ok := bar.Children["zab"]
			Ω(ok).Should(BeTrue())

			Ω(root.Count).Should(Equal(uint64(2)))
			Ω(bar.Count).Should(Equal(uint64(2)))
			Ω(baz.Count).Should(Equal(uint64(1)))
			Ω(zab.Count).Should(Equal(uint64(1)))
		})

		It("should return an error when not adding paths relative to root", func() {
			// Create a simple tree using AddFile
			root := &BlobTree{Name: "foo", Children: make(map[string]*BlobTree)}
			err := root.AddFile("/foo/bar/baz/test.txt")
			Ω(err).Should(HaveOccurred())
		})

		Context("building from disk", func() {

			var err error
			var tmpDir string
			var btree *BlobTree

			BeforeEach(func() {
				tmpDir, err = ioutil.TempDir("", TempDirPrefix)
				Ω(err).ShouldNot(HaveOccurred())

				// Construct data in the tmp directory
				err = os.MkdirAll(filepath.Join(tmpDir, "foo", "bar"), 0755)
				Ω(err).ShouldNot(HaveOccurred())

				Ω(ioutil.WriteFile(filepath.Join(tmpDir, "foo", "a.txt"), []byte("a"), 0644)).ShouldNot(HaveOccurred())
				Ω(ioutil.WriteFile(filepath.Join(tmpDir, "foo", "b.txt"), []byte("b"), 0644)).ShouldNot(HaveOccurred())
				Ω(ioutil.WriteFile(filepath.Join(tmpDir, "foo", "bar", "c.txt"), []byte("c"), 0644)).ShouldNot(HaveOccurred())
				Ω(ioutil.WriteFile(filepath.Join(tmpDir, "foo", "bar", "d.txt"), []byte("d"), 0644)).ShouldNot(HaveOccurred())

				err = os.MkdirAll(filepath.Join(tmpDir, "foo", "baz"), 0755)
				Ω(err).ShouldNot(HaveOccurred())

				Ω(ioutil.WriteFile(filepath.Join(tmpDir, "foo", "baz", "e.txt"), []byte("e"), 0644)).ShouldNot(HaveOccurred())

				err = os.MkdirAll(filepath.Join(tmpDir, "zab"), 0755)
				Ω(err).ShouldNot(HaveOccurred())

				Ω(ioutil.WriteFile(filepath.Join(tmpDir, "zab", "f.txt"), []byte("f"), 0644)).ShouldNot(HaveOccurred())
				Ω(ioutil.WriteFile(filepath.Join(tmpDir, "zab", "g.txt"), []byte("g"), 0644)).ShouldNot(HaveOccurred())

				Ω(ioutil.WriteFile(filepath.Join(tmpDir, "h.txt"), []byte("h"), 0644)).ShouldNot(HaveOccurred())
				Ω(ioutil.WriteFile(filepath.Join(tmpDir, "i.txt"), []byte("i"), 0644)).ShouldNot(HaveOccurred())

				btree = new(BlobTree)
			})

			AfterEach(func() {
				err = os.RemoveAll(tmpDir)
				Ω(err).ShouldNot(HaveOccurred())
			})

			It("should build a blob tree from disk", func() {
				err := btree.Init(tmpDir, true)
				Ω(err).ShouldNot(HaveOccurred())

				Ω(btree.Count).Should(Equal(uint64(9)))

				foo, ok := btree.Children["foo"]
				Ω(ok).Should(BeTrue())
				Ω(foo.Count).Should(Equal(uint64(5)))

				bar, ok := foo.Children["bar"]
				Ω(ok).Should(BeTrue())
				Ω(bar.Count).Should(Equal(uint64(2)))

				baz, ok := foo.Children["baz"]
				Ω(ok).Should(BeTrue())
				Ω(baz.Count).Should(Equal(uint64(1)))

				zab, ok := btree.Children["zab"]
				Ω(ok).Should(BeTrue())
				Ω(zab.Count).Should(Equal(uint64(2)))

			})

			It("should be able to update a btree built from disk", func() {
				err := btree.Init(tmpDir, true)
				Ω(err).ShouldNot(HaveOccurred())

				Ω(btree.Count).Should(Equal(uint64(9)))

				btree.AddFile(filepath.Join(tmpDir, "foo", "baz", "j.txt"))
				btree.AddFile(filepath.Join(tmpDir, "zab", "far", "k.txt"))

				Ω(btree.Count).Should(Equal(uint64(11)))

				foo, ok := btree.Children["foo"]
				Ω(ok).Should(BeTrue())
				Ω(foo.Count).Should(Equal(uint64(6)))

				bar, ok := foo.Children["bar"]
				Ω(ok).Should(BeTrue())
				Ω(bar.Count).Should(Equal(uint64(2)))

				baz, ok := foo.Children["baz"]
				Ω(ok).Should(BeTrue())
				Ω(baz.Count).Should(Equal(uint64(2)))

				zab, ok := btree.Children["zab"]
				Ω(ok).Should(BeTrue())
				Ω(zab.Count).Should(Equal(uint64(3)))

				far, ok := zab.Children["far"]
				Ω(ok).Should(BeTrue())
				Ω(far.Count).Should(Equal(uint64(1)))

			})

		})

	})

})
