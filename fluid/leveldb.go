// Implements the Database interface for LevelDB

package fluid

import (
	"errors"
	"fmt"

	"github.com/syndtr/goleveldb/leveldb"
)

//===========================================================================
// Wrapper for LevelDB and management methods
//===========================================================================

// LevelDB implements the Database interface, wrapping the LevelDB library.
type LevelDB struct {
	db *leveldb.DB
}

// Init opens a LevelDB file (creating the file if it doesn't already exist)
// and initializes the buckets if they haven't already been created.
func (ldb *LevelDB) Init(path string) error {
	var err error
	ldb.db, err = leveldb.OpenFile(path, nil)
	return err
}

// Close the connection to the LevelDB
func (ldb *LevelDB) Close() error {
	return ldb.db.Close()
}

// CreateBucket modifies a key using the bucket name as a prefix.
func (ldb *LevelDB) CreateBucket(bucket string, key []byte) []byte {
	prefixed := fmt.Sprintf("%s/%s", bucket, key)
	return []byte(prefixed)
}

//===========================================================================
// LevelDB interaction methods
//===========================================================================

// Get a value for a key from a bucket using the LevelDB API
// NOTE: To maintain compatibility with the BoltDB API this function does not
// return an error on NotFound but rather returns nil value and nil error.
func (ldb *LevelDB) Get(key []byte, bucket string) ([]byte, error) {
	pkey := ldb.CreateBucket(bucket, key)
	val, err := ldb.db.Get(pkey, nil)

	if err == leveldb.ErrNotFound {
		return nil, nil
	}

	return val, err
}

// Put a key/value pair into the bucket using the LevelDB API
func (ldb *LevelDB) Put(key []byte, value []byte, bucket string) error {
	pkey := ldb.CreateBucket(bucket, key)
	return ldb.db.Put(pkey, value, nil)
}

// Delete a key from a bucket using the LevelDB API
func (ldb *LevelDB) Delete(key []byte, bucket string) error {
	pkey := ldb.CreateBucket(bucket, key)
	return ldb.db.Delete(pkey, nil)
}

//===========================================================================
// LevelDB cursor interaction methods
//===========================================================================

// Batch insert key/value pairs into a bucket using the LevelDB Batch writes.
func (ldb *LevelDB) Batch(keys [][]byte, values [][]byte, bucket string) error {
	var pkey []byte

	if len(keys) != len(values) {
		return errors.New("specify the same number of keys and values for batch update")
	}

	// Create the batch transaction
	batch := new(leveldb.Batch)

	for i := 0; i < len(keys); i++ {
		pkey = ldb.CreateBucket(bucket, keys[i])
		batch.Put(pkey, values[i])
	}

	// Write the batch operation to disk
	return ldb.db.Write(batch, nil)
}

// Scan a group of keys with a particular prefix using LevelDB prefix seek.
func (ldb *LevelDB) Scan(prefix []byte, bucket string) (*Cursor, error) {

	return nil, nil
}

// Keys gets all the keys for a bucket using the LevelDB API
func (ldb *LevelDB) Keys(bucket string) (*Cursor, error) {
	return nil, nil
}
