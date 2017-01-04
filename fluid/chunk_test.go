package fluid_test

import (
	"crypto/rand"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	. "github.com/bbengfort/fluidfs/fluid"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Chunk", func() {

	Describe("hashing algorithm selection", func() {

		It("should select a hashing function based on name", func() {
			names := []string{"md5", "sha1", "sha224", "sha256", "murmur"}
			for _, name := range names {
				hasher, err := CreateHasher(name)
				Ω(err).Should(BeNil())

				hash := hasher()
				hash.Write([]byte("This is a test string"))
				Ω(hash.Sum(nil)).ShouldNot(BeZero())
			}
		})

		It("should return an error when attempting to create hasher with unknown name", func() {
			_, err := CreateHasher("bibbidy")
			Ω(err).ShouldNot(BeNil())
		})

	})

	Describe("signed chunkers", func() {

		short := []byte("The eagle flies at midnight")
		text1k := []byte(strings.Repeat("fizzbuzz", 128))
		text4k := []byte(strings.Repeat("buzzfizz", 512))
		text12k := []byte(strings.Repeat("foo bar ", 1536))

		It("should be able to create md5 hashes", func() {

			// Create the signed chunker
			chunker := new(SignedChunker)
			hasher, err := CreateHasher(MD5)
			Ω(err).Should(BeNil())
			chunker.SetHasher(hasher)

			var sigTests = []struct {
				value  []byte
				signed string
			}{
				{short, "UPo8xAOMJzMMfi6FRJTGGQ"},
				{text1k, "NHp-r7-uqpqTvZ01mfNNtw"},
				{text4k, "JXz16f8Il9XMBoAuJm4LLw"},
				{text12k, "NL12EoerAzd1vr1XeP3Erg"},
			}

			for _, st := range sigTests {
				Ω(chunker.Signature(st.value)).Should(Equal(st.signed))
			}

		})

		It("should be able to create sha1 hashes", func() {

			// Create the signed chunker
			chunker := new(SignedChunker)
			hasher, err := CreateHasher(SHA1)
			Ω(err).Should(BeNil())
			chunker.SetHasher(hasher)

			var sigTests = []struct {
				value  []byte
				signed string
			}{
				{short, "ddd1hFDMqHPp1_QWxKsggbEeuIE"},
				{text1k, "kqNvrccP2wWHd67FNgtPOR-KkHk"},
				{text4k, "7G5yFdWboOylYC5fsrcV8EGCNL8"},
				{text12k, "hnxjQAyD2YQCOlw96M2nV1vGFJQ"},
			}

			for _, st := range sigTests {
				Ω(chunker.Signature(st.value)).Should(Equal(st.signed))
			}

		})

		It("should be able to create sha224 hashes", func() {

			// Create the signed chunker
			chunker := new(SignedChunker)
			hasher, err := CreateHasher(SHA224)
			Ω(err).Should(BeNil())
			chunker.SetHasher(hasher)

			var sigTests = []struct {
				value  []byte
				signed string
			}{
				{short, "kP4LwHw2LqagkR0xKxx8wYMUcvjY698k3L5iaA"},
				{text1k, "pyEWhLW0_h8PQkEAlzBhXYcI_hOFnWBGhNgJkw"},
				{text4k, "rLJji-vpv7erGwMBJwnLChxobBwWbtO7CHFztg"},
				{text12k, "YAnA0SbyDV6Ler4NObMo8NWJbA6m19J4P3fgVw"},
			}

			for _, st := range sigTests {
				Ω(chunker.Signature(st.value)).Should(Equal(st.signed))
			}

		})

		It("should be able to create sha256 hashes", func() {

			// Create the signed chunker
			chunker := new(SignedChunker)
			hasher, err := CreateHasher(SHA256)
			Ω(err).Should(BeNil())
			chunker.SetHasher(hasher)

			var sigTests = []struct {
				value  []byte
				signed string
			}{
				{short, "BT5kuWbJ_z-3eDVnXfj2ou0eTHBdPcniq3wATmAStRw"},
				{text1k, "H_K52yLi9bHCZR4B6i9Tg_QCTD8yHt25DFWtvtGkUCY"},
				{text4k, "d0wHLy3EhSpQa4yaXlieEa4c5DaSv6q9QqElX2mHIw4"},
				{text12k, "q1q5lGtxC_FBg179MO3gJPAMjXe4aP7KhddW3EiEswY"},
			}

			for _, st := range sigTests {
				Ω(chunker.Signature(st.value)).Should(Equal(st.signed))
			}

		})

		It("should be able to create murmur hashes", func() {

			// Create the signed chunker
			chunker := new(SignedChunker)
			hasher, err := CreateHasher(Murmur)
			Ω(err).Should(BeNil())
			chunker.SetHasher(hasher)

			var sigTests = []struct {
				value  []byte
				signed string
			}{
				{short, "pOalFoebLnN03XVP31S9gw"},
				{text1k, "UjjAhpjBZFykusXZEZ-2hw"},
				{text4k, "mhX6H2FP3bbWAubrGgZsSw"},
				{text12k, "WDAF3cGMHlygHhuNN8xEVg"},
			}

			for _, st := range sigTests {
				Ω(chunker.Signature(st.value)).Should(Equal(st.signed))
			}

		})

		It("should return same hash no matter the hash ordering", func() {

			// create random data of 7168 bytes each
			cases := [][]byte{
				make([]byte, 7168), make([]byte, 7168), make([]byte, 7168),
			}

			for _, data := range cases {
				length, err := rand.Read(data)
				Ω(err).Should(BeNil())
				Ω(length).Should(Equal(7168))
			}

			// Evaluate all hashing algorithms
			names := []string{MD5, SHA1, SHA224, SHA256, Murmur}
			for _, name := range names {

				signer := new(SignedChunker)

				// Create the first hasher
				hasher, err := CreateHasher(name)
				Ω(err).Should(BeNil())
				signer.SetHasher(hasher)

				// Create the first list of hash strings
				alpha := make([]string, len(cases))
				for _, data := range cases {
					alpha = append(alpha, signer.Signature(data))
				}

				// Reverse the data list and create second hasher
				reverse(cases)
				hasher, err = CreateHasher(name)
				Ω(err).Should(BeNil())
				signer.SetHasher(hasher)

				bravo := make([]string, len(cases))
				for _, data := range cases {
					bravo = append(bravo, signer.Signature(data))
				}

				// Reverse the bravo list
				sreverse(bravo)

				// Compare the hashes made in reverse order
				for i := 0; i < len(cases); i++ {
					Ω(alpha[i]).ShouldNot(Equal(bravo[i]))
				}
			}

		})

	})

	Describe("blob structs", func() {

		var err error
		var tmpDir string

		BeforeEach(func() {
			tmpDir, err = ioutil.TempDir("", TempDirPrefix)
			Ω(err).Should(BeNil())
		})

		AfterEach(func() {
			err = os.RemoveAll(tmpDir)
			Ω(err).Should(BeNil())
		})

		It("should compute a path based on the hash", func() {
			blob, err := MakeBlob([]byte("I shot the elephant in my pajamas"), SHA256)
			Ω(err).Should(BeNil())
			Ω(blob.Hash()).Should(Equal("7rYjqdSaixrocwtlp86HAEYTMfPS71tObgYGVtR-SUI"))
			Ω(blob.Path()).Should(Equal("7rYjqdSa/ixrocwtl/p86HAEYT/MfPS71tO/bgYGVtR-/7rYjqdSaixrocwtlp86HAEYTMfPS71tObgYGVtR-SUI.blob"))
		})

		It("should compute the size of the data", func() {
			blob, err := MakeBlob([]byte("I shot the elephant in my pajamas"), SHA256)
			Ω(err).Should(BeNil())
			Ω(blob.Size()).Should(Equal(33))
		})

		It("should return data unmodified", func() {
			data := []byte("I shot the elephant in my pajamas\nThey were a tight fit!")
			blob, err := MakeBlob(data, SHA256)
			Ω(err).Should(BeNil())
			Ω(blob.Data()).Should(Equal(data))
		})

		It("should be able to save blobs to disk", func() {
			// Set up the fixtures
			data := []byte("I shot the elephant in my pajamas\nThey were a tight fit!")
			path := "6waaG_UO/nw_7OyJ2/2SO_dNvj/qO1L3Aml/iiQjwYBT/6waaG_UOnw_7OyJ22SO_dNvjqO1L3AmliiQjwYBTSEc.blob"
			path = filepath.Join(tmpDir, path)

			// Make the blob
			blob, err := MakeBlob(data, SHA256)
			Ω(err).Should(BeNil())

			// Save the blob
			err = blob.Save(tmpDir)
			Ω(err).Should(BeNil())

			// Ensure the file exists
			info, err := os.Stat(path)
			Ω(os.IsNotExist(err)).Should(BeFalse())
			Ω(info.Mode().IsRegular()).Should(BeTrue())

			// Read the file and make sure it contains exactly the data
			fdata, err := ioutil.ReadFile(path)
			Ω(err).Should(BeNil())
			Ω(fdata).Should(Equal(data))

		})

		It("should be able to load a blob from disk", func() {
			// Set up the fixtures
			data := []byte("I shot the elephant in my pajamas\nThey were a tight fit!")
			path := "6waaG_UO/nw_7OyJ2/2SO_dNvj/qO1L3Aml/iiQjwYBT/6waaG_UOnw_7OyJ22SO_dNvjqO1L3AmliiQjwYBTSEc.blob"
			path = filepath.Join(tmpDir, path)

			// Write the data to disk.
			os.MkdirAll(filepath.Dir(path), ModeStorageDir)
			err := ioutil.WriteFile(path, data, ModeBlob)
			Ω(err).Should(BeNil())

			// Load the blob
			blob := new(Blob)
			err = blob.Load(path)
			Ω(err).Should(BeNil())
			Ω(blob).ShouldNot(BeZero())

			// Check the data, size, path, and hash
			Ω(blob.Data()).Should(Equal(data))
			Ω(blob.Size()).Should(Equal(len(data)))
			Ω(blob.Path()).Should(Equal(path))
			Ω(blob.Hash()).Should(Equal("6waaG_UOnw_7OyJ22SO_dNvjqO1L3AmliiQjwYBTSEc"))
		})

		It("should be able to load arbitrary data", func() {
			// Set up the fixtures
			data := []byte("I shot the elephant in my pajamas\nThey were a tight fit!")
			path := filepath.Join(tmpDir, "note.txt")

			// Write the data to disk.
			os.MkdirAll(filepath.Dir(path), ModeStorageDir)
			err := ioutil.WriteFile(path, data, ModeBlob)
			Ω(err).Should(BeNil())

			// Load the blob
			blob := new(Blob)
			err = blob.Load(path)
			Ω(err).Should(BeNil())
			Ω(blob).ShouldNot(BeZero())

			// Check the data, size, path, and hash
			Ω(blob.Data()).Should(Equal(data))
			Ω(blob.Size()).Should(Equal(len(data)))
			Ω(blob.Path()).Should(Equal(path))
			Ω(blob.Hash()).Should(Equal(""))
		})

		It("should be able to save a blob then load it", func() {
			data := []byte(randString(4096))
			blob, err := MakeBlob(data, SHA256)
			Ω(err).Should(BeNil())

			// Save the blob
			err = blob.Save(tmpDir)
			Ω(err).Should(BeNil())

			// Load the new blob
			// NOTE: on save the blob path is stored with the temp directory
			// TODO: is this a problem for serialization?
			newBlob := new(Blob)
			err = newBlob.Load(blob.Path())
			Ω(err).Should(BeNil())

			Ω(blob).Should(Equal(newBlob))
		})

	})

	Describe("fixed length chunking", func() {

	})

	Describe("rabin-karp chunking", func() {

	})

})
