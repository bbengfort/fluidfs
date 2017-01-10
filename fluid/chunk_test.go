package fluid_test

import (
	"crypto/rand"
	"fmt"
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
				Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))

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
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))
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
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))
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
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))
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
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))
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
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))
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
				Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))
				Ω(length).Should(Equal(7168))
			}

			// Evaluate all hashing algorithms
			names := []string{MD5, SHA1, SHA224, SHA256, Murmur}
			for _, name := range names {

				signer := new(SignedChunker)

				// Create the first hasher
				hasher, err := CreateHasher(name)
				Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))
				signer.SetHasher(hasher)

				// Create the first list of hash strings
				alpha := make([]string, len(cases))
				for _, data := range cases {
					alpha = append(alpha, signer.Signature(data))
				}

				// Reverse the data list and create second hasher
				reverse(cases)
				hasher, err = CreateHasher(name)
				Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))
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
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))
		})

		AfterEach(func() {
			err = os.RemoveAll(tmpDir)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))
		})

		It("should compute a path based on the hash", func() {
			blob, err := MakeBlob([]byte("I shot the elephant in my pajamas"), SHA256)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))
			Ω(blob.Hash()).Should(Equal("7rYjqdSaixrocwtlp86HAEYTMfPS71tObgYGVtR-SUI"))
			Ω(blob.Path()).Should(Equal("7rYjqdSa/ixrocwtl/p86HAEYT/MfPS71tO/bgYGVtR-/7rYjqdSaixrocwtlp86HAEYTMfPS71tObgYGVtR-SUI.blob"))
		})

		It("should compute the size of the data", func() {
			blob, err := MakeBlob([]byte("I shot the elephant in my pajamas"), SHA256)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))
			Ω(blob.Size()).Should(Equal(33))
		})

		It("should return data unmodified", func() {
			data := []byte("I shot the elephant in my pajamas\nThey were a tight fit!")
			blob, err := MakeBlob(data, SHA256)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))
			Ω(blob.Data()).Should(Equal(data))
		})

		It("should be able to save blobs to disk", func() {
			// Set up the fixtures
			data := []byte("I shot the elephant in my pajamas\nThey were a tight fit!")
			path := "6waaG_UO/nw_7OyJ2/2SO_dNvj/qO1L3Aml/iiQjwYBT/6waaG_UOnw_7OyJ22SO_dNvjqO1L3AmliiQjwYBTSEc.blob"
			path = filepath.Join(tmpDir, path)

			// Make the blob
			blob, err := MakeBlob(data, SHA256)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))

			// Save the blob
			err = blob.Save(tmpDir)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))

			// Ensure the file exists
			info, err := os.Stat(path)
			Ω(os.IsNotExist(err)).Should(BeFalse())
			Ω(info.Mode().IsRegular()).Should(BeTrue())

			// Read the file and make sure it contains exactly the data
			fdata, err := ioutil.ReadFile(path)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))
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
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))

			// Load the blob
			blob := new(Blob)
			err = blob.Load(path)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))
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
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))

			// Load the blob
			blob := new(Blob)
			err = blob.Load(path)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))
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
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))

			// Save the blob
			err = blob.Save(tmpDir)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))

			// Load the new blob
			// NOTE: on save the blob path is stored with the temp directory
			// TODO: is this a problem for serialization?
			newBlob := new(Blob)
			err = newBlob.Load(blob.Path())
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))

			Ω(blob).Should(Equal(newBlob))
		})

	})

	Describe("fixed length chunking", func() {

		var config *StorageConfig
		var tmpDir string
		var err error

		BeforeEach(func() {

			tmpDir, err = ioutil.TempDir("", TempDirPrefix)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))

			config = &StorageConfig{
				Path:         tmpDir,
				Chunking:     FixedLengthChunking,
				BlockSize:    512,
				MinBlockSize: 128,
				MaxBlockSize: 640,
				Hashing:      SHA256,
			}

			Ω(config.Validate()).Should(BeNil())
		})

		AfterEach(func() {
			err = os.RemoveAll(tmpDir)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))
		})

		It("should create a FixedLengthChunker on demand", func() {

			data := []byte(randString(512))
			chunker, err := NewChunker(data, config)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))

			chunker, ok := chunker.(*FixedLengthChunker)
			Ω(ok).Should(BeTrue())

			Ω(chunker.BlockSize()).Should(Equal(config.BlockSize))

		})

		It("should create even length chunks", func() {
			data := []byte(randString(2048))
			chunker, err := NewChunker(data, config)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))

			chunks := make([]*Blob, 0, 4)

			for chunker.Next() {
				blob := chunker.Chunk()
				chunks = append(chunks, blob.(*Blob))
				Ω(blob.Data()).Should(HaveLen(config.BlockSize))
				Ω(blob.Hash()).ShouldNot(BeZero())
			}

			Ω(chunks).Should(HaveLen(4))
		})

		It("should create a small last chunk bigger than the minimum", func() {
			data := []byte(randString(2304))
			chunker, err := NewChunker(data, config)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))

			chunks := make([]*Blob, 0, 5)

			for chunker.Next() {
				blob := chunker.Chunk()
				chunks = append(chunks, blob.(*Blob))
				Ω(len(blob.Data())).Should(BeNumerically(">=", config.MinBlockSize))
				Ω(blob.Hash()).ShouldNot(BeZero())
			}

			Ω(chunks).Should(HaveLen(5))
		})

		It("should respect the minimum blob size", func() {
			data := []byte(randString(2144))
			chunker, err := NewChunker(data, config)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))

			chunks := make([]*Blob, 0, 4)

			for chunker.Next() {
				blob := chunker.Chunk()
				chunks = append(chunks, blob.(*Blob))
				Ω(len(blob.Data())).Should(BeNumerically(">=", config.MinBlockSize))
				Ω(blob.Hash()).ShouldNot(BeZero())
			}

			Ω(chunks).Should(HaveLen(4))
		})

		It("should respect the exact minimum blob size", func() {
			data := []byte(randString(2176))
			chunker, err := NewChunker(data, config)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))

			chunks := make([]*Blob, 0, 5)

			for chunker.Next() {
				blob := chunker.Chunk()
				chunks = append(chunks, blob.(*Blob))
				Ω(len(blob.Data())).Should(BeNumerically(">=", config.MinBlockSize))
				Ω(blob.Hash()).ShouldNot(BeZero())
			}

			Ω(chunks).Should(HaveLen(5))
		})

		It("should be able to chunk and recombine without errors", func() {
			data := []byte(randString(2144))
			chunker, err := NewChunker(data, config)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))

			chunks := make([]*Blob, 0, 4)

			for chunker.Next() {
				blob := chunker.Chunk()
				chunks = append(chunks, blob.(*Blob))
			}

			combined := make([]byte, 0, 2144)
			for _, blob := range chunks {
				combined = append(combined, blob.Data()...)
			}

			Ω(data).Should(Equal(combined))
		})

		It("should be able to chunk, save to disk, and read without errors", func() {
			data := []byte(randString(2304))
			chunker, err := NewChunker(data, config)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))

			paths := make([]string, 0, 5)

			for chunker.Next() {
				blob := chunker.Chunk().(*Blob)
				blob.Save(tmpDir)
				paths = append(paths, blob.Path())
			}

			fdata := make([]byte, 0, 2304)
			for _, path := range paths {
				rdata, err := ioutil.ReadFile(path)
				Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))
				fdata = append(fdata, rdata...)
			}

			Ω(data).Should(Equal(fdata))

		})

		It("should be able to chunk foo.txt with default block sizes", func() {
			fixture := filepath.Join("testdata", "foo.txt")
			data, err := ioutil.ReadFile(fixture)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))

			config := new(StorageConfig)
			config.Defaults()
			config.Path = tmpDir
			config.Chunking = FixedLengthChunking
			Ω(config.Validate()).Should(BeNil())

			chunker, err := NewChunker(data, config)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))

			sizes := make([]int, 0, 10)

			for chunker.Next() {
				blob := chunker.Chunk()
				sizes = append(sizes, blob.Size())
			}

			Ω(sizes).Should(HaveLen(10))
			for i, s := range sizes {
				if i < 9 {
					Ω(s).Should(Equal(4096))
				} else {
					Ω(s).Should(Equal(5199))
				}
			}

		})

		It("should be able to iterate through chunks multiple times (call reset)", func() {
			data := []byte(randString(3264))
			chunker, err := NewChunker(data, config)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))

			alpha := make([]*Blob, 0, 8)
			bravo := make([]*Blob, 0, 8)

			// First iteration
			for chunker.Next() {
				alpha = append(alpha, chunker.Chunk().(*Blob))
			}

			// Second iteration
			for chunker.Next() {
				bravo = append(bravo, chunker.Chunk().(*Blob))
			}

			Ω(alpha).Should(Equal(bravo))
		})

	})

	Describe("rabin-karp chunking", func() {

		var config *StorageConfig
		var tmpDir string
		var err error

		BeforeEach(func() {

			tmpDir, err = ioutil.TempDir("", TempDirPrefix)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))

			config = new(StorageConfig)
			config.Defaults()
			config.Path = tmpDir
			config.Chunking = VariableLengthChunking
			Ω(config.Validate()).Should(BeNil())
		})

		It("should create a RabinKarpChunker on demand", func() {
			data := []byte(randString(512))
			chunker, err := NewChunker(data, config)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))

			chunker, ok := chunker.(*RabinKarpChunker)
			Ω(ok).Should(BeTrue())

			Ω(chunker.BlockSize()).Should(Equal(8192))

		})

		It("should create variable length blobs", func() {
			data := []byte(randString(87542))
			chunker, err := NewChunker(data, config)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))

			sizes := make([]int, 0, 11)

			// Iterate over the chunks
			for chunker.Next() {
				blob := chunker.Chunk().(*Blob)

				Ω(blob.Size()).Should(BeNumerically(">=", config.MinBlockSize))
				Ω(blob.Size()).Should(BeNumerically("<=", config.MaxBlockSize))
				sizes = append(sizes, blob.Size())

			}

			Ω(len(sizes)).Should(BeNumerically(">", 10))

			sum := 0
			for _, size := range sizes {
				sum += size
			}
			Ω(sum).Should(Equal(87542))
		})

		It("should be able to chunk foo.txt with default block sizes", func() {

			/*
			   Target for foo.txt:
			   Chunk at offset       0, len 8118
			   Chunk at offset    8118, len 3638
			   Chunk at offset   11756, len 3286
			   Chunk at offset   15042, len 2479
			   Chunk at offset   17521, len 8192
			   Chunk at offset   25713, len 2841
			   Chunk at offset   28554, len 2705
			   Chunk at offset   31259, len 5685
			   Chunk at offset   36944, len 5119
			*/

			fixture := filepath.Join("testdata", "foo.txt")
			expected := []int{8118, 3638, 3286, 2479, 8192, 2841, 2705, 5685, 5119}

			data, err := ioutil.ReadFile(fixture)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))

			chunker, err := NewChunker(data, config)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))

			sizes := make([]int, 0, 9)
			for chunker.Next() {
				sizes = append(sizes, chunker.Chunk().Size())
			}

			Ω(sizes).Should(HaveLen(9))
			Ω(sizes).Should(Equal(expected))
		})

		It("should be able to iterate through chunks multiple times (call reset)", func() {
			data := []byte(randString(32640))
			chunker, err := NewChunker(data, config)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))

			alpha := make([]*Blob, 0, 5)
			bravo := make([]*Blob, 0, 5)

			// First iteration
			for chunker.Next() {
				alpha = append(alpha, chunker.Chunk().(*Blob))
			}

			// Second iteration
			for chunker.Next() {
				bravo = append(bravo, chunker.Chunk().(*Blob))
			}

			Ω(len(alpha)).Should(BeNumerically(">", 3))
			Ω(len(bravo)).Should(BeNumerically(">", 3))
			Ω(alpha).Should(Equal(bravo))
		})

		It("should be able to chunk and recombine without errors", func() {
			data := []byte(randString(96524))
			chunker, err := NewChunker(data, config)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))

			chunks := make([]*Blob, 0, 4)

			for chunker.Next() {
				blob := chunker.Chunk()
				chunks = append(chunks, blob.(*Blob))
			}

			combined := make([]byte, 0, 2144)
			for _, blob := range chunks {
				combined = append(combined, blob.Data()...)
			}

			Ω(data).Should(Equal(combined))
		})

		It("should be able to chunk, save to disk, and read without errors", func() {
			data := []byte(randString(1120304))
			chunker, err := NewChunker(data, config)
			Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))

			paths := make([]string, 0, 5)

			for chunker.Next() {
				blob := chunker.Chunk().(*Blob)
				blob.Save(tmpDir)
				paths = append(paths, blob.Path())
			}

			fdata := make([]byte, 0, 2304)
			for _, path := range paths {
				rdata, err := ioutil.ReadFile(path)
				Ω(err).Should(BeNil(), fmt.Sprintf("%s", err))
				fdata = append(fdata, rdata...)
			}

			Ω(data).Should(Equal(fdata))

		})
	})

})
