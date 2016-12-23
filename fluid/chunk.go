// Mechanisms for chunking data into blobs. Currently there are two mechanims
// for chunking: Rabin-Karp and fixed length chunking.

package fluid

import (
	"hash"
	"io"
)

// Specifies the names of available chunking mechanisms
const (
	VariableLengthChunking = "variable"
	FixedLengthChunking    = "fixed"
)

// Names of available chunking mechanisms for validation
var chunkingMethodNames = []string{VariableLengthChunking, FixedLengthChunking}

//===========================================================================
// Chunking Structs and Interfaces
//===========================================================================

// Chunk defines the methods of a single, discrete blob of data and are
// returned by a chunking mechanism to describe a larger collection of bytes.
type Chunk interface {
	Size() int              // Returns the number of bytes in the chunk
	Data() []byte           // Returns the data of the chunk
	Hash() []byte           // Returns the hashed signature of the chunk
	Load(path string) error // Load the chunk from a path on disk or by key.
	Save(path string) error // Save the chunk to disk by path or by key.
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
	io.Writer                 // Must implement the Write() method to add data to the chunker
	Next() bool               // Advances the chunker to the next chunk
	Chunk() Chunk             // Returns the current chunk on the chunker
	Reset() error             // Reset the chunker to its original state
	BlockSize() int           // Returns the underlying size (or maximum size) of chunks
	SetHasher(hash hash.Hash) // Set the hashing method on the chunker
}

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
	hash []byte // Cached value of the signature of the blob.
	path string // A reference to the location on disk
}

//===========================================================================
// Fixed Length Chunking
//===========================================================================

// FixedLengthChunker creates chunks of approximately the same length. The
// exception is the last block which will be the length of the remaining data
// unless the length of the remaining data is smaller than the minimum block
// size, in which case the last block is longer than the block size.
type FixedLengthChunker struct {
	data         []byte    // The internal data to chunk on
	blockIndex   int       // The current index of the chunker
	blockSize    int       // The target size of the blobs
	minBlockSize int       // The minimium size for a block
	hasher       hash.Hash // The hashing algorithm to sign blobs
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
func (c *FixedLengthChunker) Chunk() *Blob {
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

// SetHasher allows users to specify a different hashing algorithm other than
// the default hashing algorithm. If this is set in the middle of chunking
// then some blobs will have a different hash than others, which is not
// recommended. The hashing algorithm can also be specified in the config.
func (c *FixedLengthChunker) SetHasher(hash hash.Hash) {
	c.hasher = hash
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
	data   []byte    // The internal data to chunk on
	hasher hash.Hash // The hashing algorithm to sign blobs
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
func (c *RabinKarpChunker) Chunk() *Blob {
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

// SetHasher allows users to specify a different hashing algorithm other than
// the default hashing algorithm. If this is set in the middle of chunking
// then some blobs will have a different hash than others, which is not
// recommended. The hashing algorithm can also be specified in the config.
func (c *RabinKarpChunker) SetHasher(hash hash.Hash) {
	c.hasher = hash
}
