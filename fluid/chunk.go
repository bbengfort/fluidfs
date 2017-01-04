// Mechanisms for chunking data into blobs. Currently there are two mechanims
// for chunking: Rabin-Karp and fixed length chunking.

package fluid

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/spaolacci/murmur3"
)

// Specifies the names of available chunking mechanisms
const (
	VariableLengthChunking = "variable"
	FixedLengthChunking    = "fixed"
)

// Specifies the names of available hashing algorithms
const (
	MD5      = "md5"
	SHA1     = "sha1"
	SHA224   = "sha224"
	SHA256   = "sha256"
	CityHash = "cityhash"
	Murmur   = "murmur"
	SipHash  = "siphash"
)

// Specifies the storage permission modes
const (
	ModeStorageDir = 0755
	ModeBlob       = 0644
)

// BlobExt specifies the extension of blob files on disk
const BlobExt = ".blob"

// Names of available chunking mechanisms for validation
var chunkingMethodNames = []string{VariableLengthChunking, FixedLengthChunking}

// Names of hashing algorithms for validation
var hashingAlgorithmNames = []string{MD5, SHA1, SHA224, SHA256, Murmur}

//===========================================================================
// Chunking Structs and Interfaces
//===========================================================================

// Chunk defines the methods of a single, discrete blob of data and are
// returned by a chunking mechanism to describe a larger collection of bytes.
type Chunk interface {
	Size() int              // Returns the number of bytes in the chunk
	Data() []byte           // Returns the data of the chunk
	Hash() string           // Returns the hashed signature of the chunk
	Load(path string) error // Load the chunk from a path on disk or by key.
	Save(path string) error // Save the chunk to disk by path or by key.
}

// Hasher is an interface that defines the ability for a struct to accept
// arbitrary data and return a string signature from it. Chunkers use this
// interface to uniquely identify blobs by the hash of the blob contents.
// All hashers can accept arbitrary hashing algorithms via the SetHasher
// method. Specific implementations may add prefixes or other utilities.
// Note the string encoding of the signature can vary.
type Hasher interface {
	Signature(data []byte) string // Returns a string representation of the hash sum of the data
	SetHasher(func() hash.Hash)   // Set the hashing algorithm of the chunker.
}

// Chunker is similar to a hash.Hash but provides methods for dividing up a
// a slice of bytes into smaller chunks. The primary interface for a Chunker
// is an iteration that yields Chunk objects. Basic usage is as follows:
//
//      chunker := NewChunker()
//      chunker.Write(data)
//      for chunker.Next() {
//          chunk := chunker.Chunk()
//      }
//
// Note that any chunker needs to understand how to hash each chunk and have
// knowledge about the block size. For variable length chunks we choose the
// maximal possible block size so that callers can allocate space correctly.
type Chunker interface {
	io.Writer       // Must implement the Write() method to add data to the chunker
	Hasher          // Must implement the hasher interface
	Next() bool     // Advances the chunker to the next chunk
	Chunk() Chunk   // Returns the current chunk on the chunker
	Reset() error   // Reset the chunker to its original state
	BlockSize() int // Returns the underlying size (or maximum size) of chunks
}

// NewChunker uses a storage configuration to initialize the appropriate
// chunking mechanism on the specified data. NewChunker is not usually called
// directly but wrapped by a closure that passes in the default configuration.
func NewChunker(data []byte, config *StorageConfig) (Chunker, error) {

	var chunker Chunker

	// Create the chunker based on the configuration
	switch config.Chunking {
	case FixedLengthChunking:
		chunker = &FixedLengthChunker{
			data:         data,
			blockIndex:   0,
			blockSize:    config.BlockSize,
			minBlockSize: config.MinBlockSize,
		}
	case VariableLengthChunking:
		chunker = &RabinKarpChunker{
			data: data,
		}
	default:
		return nil, fmt.Errorf("unknown chunking method: '%s'", config.Chunking)
	}

	// Initialize a new hasher to uniquely identify chunks (blobs).
	hasher, err := CreateHasher(config.Hashing)
	if err != nil {
		return nil, err
	}

	chunker.SetHasher(hasher)

	return chunker, nil
}

// CreateHasher evalautes the string passed in and initializes the appropriate
// hashing algorithm for use with the SetHasher function of a Chunker.
// TODO: Make hashingAlgorithmNames a map of names to functions instead of switch.
// NOTE: This function optimizes murmur for x64 architecutres and will result in different values for x86 vs. x64 systems!
func CreateHasher(name string) (func() hash.Hash, error) {
	switch name {
	case MD5:
		return md5.New, nil
	case SHA1:
		return sha1.New, nil
	case SHA224:
		return sha256.New224, nil
	case SHA256:
		return sha256.New, nil
	case Murmur:
		return func() hash.Hash {
			// NOTE: this function optimizes murmur3 for x64 architectures
			return murmur3.New128()
		}, nil
	default:
		return nil, fmt.Errorf("unknown hashing algorithm: '%s'", name)
	}
}

//===========================================================================
// Blob Implementation of Chunk
//===========================================================================

// Blob implements the Chunk interface and is the basic data structure for
// chunks of a file on disk. Blobs can be variable or fixed length, depending
// on the chunking mechanism, but every blob is a unique, immutable data
// structure. The primary extension of blobs are encrypted blobs, to be
// implemented in the future.
//
// Note the use of unexported fields (which you probably can't see if you're
// reading this in the documentation). This makes serialization using the
// marshalling interface difficult, but is intended to force blob interaction
// through the Chunk interface. Moreover, our current plan is to store blobs
// on disk, therefore the data structure will always be computed at load time.
// If we move to storing the blobs in a key/value store then we should modify
// the API for JSON or other binary representation of the structure.
type Blob struct {
	data []byte // Internal data store, returned by the Data() method.
	size int    // Internal reference to the size of the data on disk.
	hash string // Cached value of the signature of the blob.
	path string // A reference to the location on disk
}

// MakeBlob creates a blob directly from data and a hashing function. It is
// used as a diagnostic utility only; blobs should be created from Chunkers.
func MakeBlob(data []byte, hash string) (*Blob, error) {
	hasher, err := CreateHasher(hash)
	if err != nil {
		return nil, err
	}

	signer := new(SignedChunker)
	signer.SetHasher(hasher)

	return &Blob{
		data: data,
		hash: signer.Signature(data),
	}, nil
}

// Size computes the size of the blob in bytes if it is not already cached,
// memoizes the data in memory and returns the cached value.
func (b *Blob) Size() int {
	if b.size == 0 {
		b.size = len(b.data)
	}
	return b.size
}

// Data returns the complete data representing the blob.
func (b *Blob) Data() []byte {
	return b.data
}

// Hash returns the unique hash signature of the blob. Hash is interesting
// because the hashing methodology is stored on the chunker, not the Blob.
// In order for the blob to compute the hash it needs to know the hash type
// (e.g. SHA256 or SHA1) or store the hash on disk along with the data. To
// handle this, the Hash function doesn't do any computation, returning saved
// information about the blob from its construction or loading from disk.
func (b *Blob) Hash() string {
	return b.hash
}

// Path returns the file path to the blob on disk. If the blob is loaded then
// the path should be stored on load; otherwise the path is computed based on
// the hash. See the Blob.Save() method for more information on paths.
func (b *Blob) Path() string {
	if b.path == "" {
		parts := StrideFixed(b.hash, 8)
		dirname := filepath.Join(parts...)
		b.path = filepath.Join(dirname, b.hash+BlobExt)
	}

	return b.path
}

// Load a blob from a path on disk, the path on disk should include a
// computable representation of the hash assigned the the blob.
//
// Currently the Load method expects the hash to be the filename followed by
// the .blob extension as defined by the Blob.Save method.
func (b *Blob) Load(path string) error {

	// Read the data from the file.
	var err error
	if b.data, err = ioutil.ReadFile(path); err != nil {
		return err
	}

	// Store the path on the blob.
	b.path = path

	// Compute the hash from the filename if it has the .blob extension
	if filepath.Ext(path) == BlobExt {
		_, filename := filepath.Split(path)
		b.hash = strings.TrimSuffix(filename, BlobExt)
	}

	return nil
}

// Save a blob to a directory on disk. The blob will be stored in a file name
// based on its hash to prevent duplicates and collisions and to allow for
// easy lookups on disk.
//
// Currently we expect the blob to be stored in a directory whose parts are
// made up of the prefixes of the blob signature, while the blob filename is
// the complete signature. For example for blob hash UPo8xAOMJzMMfi6FRJTGGQ:
// dataDir/UPo8xAOM/JzMMfi6F/UPo8xAOMJzMMfi6FRJTGGQ.blob
//
// This method will therefore create the appropriate subdirectories and join
// it to the root dataDir passed into the function and write the file to that
// location so Blob.Load can use the filename to retrieve the hash.
func (b *Blob) Save(dataDir string) error {

	// Compute the path with the data directory
	// NOTE: this stores the data directory with the blob; is this a problem for serialization?
	path := b.Path()
	if dataDir != "" && !strings.HasPrefix(path, dataDir) {
		path = filepath.Join(dataDir, path)
		b.path = path
	}

	// Ensure the parent directory exists
	dir := filepath.Dir(path)
	_ = os.MkdirAll(dir, ModeStorageDir)

	// Write the file
	if err := ioutil.WriteFile(path, b.data, ModeBlob); err != nil {
		return err
	}

	return nil
}

//===========================================================================
// Base struct so that chunkers can create blob signatures.
//===========================================================================

// SignedChunker implements the methods used to
type SignedChunker struct {
	hasher func() hash.Hash // The hashing algorithm to sign blobs
}

// Signature returns the string encoded representation of the hash sum of the
// data passed in. The hash is determined by the hashing function set on the
// SignedChunker. String encoding is fixed to hexadecimal encoding for now,
// though we could use path safe base64 encoding in the future.
func (c *SignedChunker) Signature(data []byte) string {
	hash := c.hasher()
	hash.Write(data)
	return base64.RawURLEncoding.EncodeToString(hash.Sum(nil))
}

// SetHasher allows users to specify a different hashing algorithm other than
// the default hashing algorithm. If this is set in the middle of chunking
// then some blobs will have a different hash than others, which is not
// recommended. The hashing algorithm can also be specified in the config.
func (c *SignedChunker) SetHasher(hash func() hash.Hash) {
	c.hasher = hash
}

//===========================================================================
// Fixed Length Chunking
//===========================================================================

// FixedLengthChunker creates chunks of approximately the same length. The
// exception is the last block which will be the length of the remaining data
// unless the length of the remaining data is smaller than the minimum block
// size, in which case the last block is longer than the block size.
type FixedLengthChunker struct {
	SignedChunker
	data         []byte // The internal data to chunk on
	blockIndex   int    // The current index of the chunker
	blockSize    int    // The target size of the blobs
	minBlockSize int    // The minimium size for a block
}

// Write data into a FixedLengthChunker. This method does not reset the
// chunker, which could lead to differently sized chunks based on the position
// of the blockIndex when the data is written. This can be avoided by ensuring
// that Reset() is called after a Write().
func (c *FixedLengthChunker) Write(p []byte) (n int, err error) {

	return len(p), nil
}

// Next advances the chunker by computing the next blob and modifing its
// internal indices so that it can be returned via the Chunk() method.
func (c *FixedLengthChunker) Next() bool {
	return true
}

// Chunk fetches the current blob in the iteration. Chunk() can be called
// multiple times and will return a new pointer to a new struct when called,
// therefore it is advisable to only call the method once.
func (c *FixedLengthChunker) Chunk() Chunk {
	return &Blob{}
}

// Reset the chunker back to the first index so that the data structure can
// be chunked again. Useful to run after more data has been written into the
// chunker but the chunking has already been started.
//
// Note that Reset() will be called after Next() returns false, so without
// resetting the chunker, every iteration will leave the chunker in a state
// that it can be iterated over again.
func (c *FixedLengthChunker) Reset() error {
	return nil
}

// BlockSize simply returns the fixed length in bytes of the blobs being
// chunked. Note that the last block in the chunker may have a different size
// between minBlockSize and blockSize + minBlockSize.
func (c *FixedLengthChunker) BlockSize() int {
	return c.blockSize
}

//===========================================================================
// Rabin-Karp Chunking
//===========================================================================

// RabinKarpChunker implements variable length chunking with the Rabin-Karp
// rolling hash algorithm. Variable hashing guards from data displacement and
// shifting due to prepended data by guarding boundaries by a pattern. For
// text we typically choose newline. This allows us to create a rolling hash
// over windows of data creating variable length blobs.
type RabinKarpChunker struct {
	SignedChunker
	data []byte // The internal data to chunk on
}

// Write data into a RabinKarpChunker. This method does not reset the
// chunker, which could lead to differently sized chunks based on the position
// of the blockIndex when the data is written. This can be avoided by ensuring
// that Reset() is called after a Write().
func (c *RabinKarpChunker) Write(p []byte) (n int, err error) {

	return len(p), nil
}

// Next advances the chunker by computing the next blob and modifing its
// internal indices so that it can be returned via the Chunk() method. Next()
// advances the chunker in a variable length mechanism as per the Rabin-Karp
// algorithm.
func (c *RabinKarpChunker) Next() bool {
	return true
}

// Chunk returns the current variable length Rabin-Karp Blob in the iteration,
// computed through the rolling hash mechanism. Multiple calls to Chunk() will
// return pointers to new structs, so it may not be adviseable.
func (c *RabinKarpChunker) Chunk() Chunk {
	return &Blob{}
}

// Reset the chunker back to the first index so that the data structure can
// be chunked again. Useful to run after more data has been written into the
// chunker but the chunking has already been started.
//
// Note that Reset() will be called after Next() returns false, so without
// resetting the chunker, every iteration will leave the chunker in a state
// that it can be iterated over again.
func (c *RabinKarpChunker) Reset() error {
	return nil
}

// BlockSize returns the largest possible blob size in bytes. Since Rabin-Karp
// chunking implements variable length chunks, the size of each Blob must be
// checked on the Blob data structure. However this method will return the
// maximal size so that memory can be allocated correctly by callers.
func (c *RabinKarpChunker) BlockSize() int {
	return 12
}
