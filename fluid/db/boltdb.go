// Implements the Database interface for BoltDB

package db

import (
	"bytes"
	"errors"
	"fmt"
	"time"

	"github.com/boltdb/bolt"
)

//===========================================================================
// Wrapper for BoltDB and management methods
//===========================================================================

// BoltDB implements the Database interface, wrapping the BoltDB library.
type BoltDB struct {
	db *bolt.DB
}

// Init opens a BoltDB file (creating the file if it doesn't already exist)
// and initializes the buckets if they haven't already been created.
func (bdb *BoltDB) Init(path string) error {
	var err error

	// Open the bolt database
	bdb.db, err = bolt.Open(path, 0644, &bolt.Options{Timeout: 15 * time.Second})
	if err != nil {
		return err
	}

	// Create the buckets if they don't already exist
	err = bdb.db.Update(func(tx *bolt.Tx) error {
		buckets := []string{NamesBucket, VersionsBucket, PrefixesBucket}

		for _, name := range buckets {
			_, err := tx.CreateBucketIfNotExists([]byte(name))
			if err != nil {
				return fmt.Errorf("could not create %s bucket: %s", name, err)
			}
		}

		return nil
	})

	return err
}

// Close the connection to the BoltDB
func (bdb *BoltDB) Close() error {
	return bdb.db.Close()
}

// Count the number of keys in the given bucket.
func (bdb *BoltDB) Count(bucket string) (uint64, error) {
	var count uint64

	err := bdb.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		c := b.Cursor()

		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			count++
		}

		return nil
	})

	return count, err
}

//===========================================================================
// BoltDB interaction methods
//===========================================================================

// Get a value for a key from a bucket using BoltDB transactions
func (bdb *BoltDB) Get(key []byte, bucket string) ([]byte, error) {

	// Store a reference to the value
	var val []byte

	// Create the transaction
	err := bdb.db.View(func(tx *bolt.Tx) error {
		// Get a reference to the bucket
		bkt := tx.Bucket([]byte(bucket))
		val = bkt.Get(key)
		return nil
	})

	// Return the error from the transaction
	if err != nil {
		return nil, err
	}

	// Return the value
	return val, nil
}

// Put a key/value pair into the bucket using BoltDB transactions
func (bdb *BoltDB) Put(key []byte, value []byte, bucket string) error {
	// Create the transaction
	return bdb.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(bucket))
		return bkt.Put(key, value)
	})
}

// Delete a key from a bucket using BoltDB transaction
func (bdb *BoltDB) Delete(key []byte, bucket string) error {
	// Create the transaction
	return bdb.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(bucket))
		return bkt.Delete(key)
	})
}

//===========================================================================
// BoltDB cursor interaction methods
//===========================================================================

// Batch insert key/value pairs into a bucket using BoltDB batch transactions.
// Wanted the input to be a mapping, but you can't use a slice as a map key.
func (bdb *BoltDB) Batch(keys [][]byte, values [][]byte, bucket string) error {
	if len(keys) != len(values) {
		return errors.New("specify the same number of keys and values for batch update")
	}

	return bdb.db.Batch(func(tx *bolt.Tx) error {

		bkt := tx.Bucket([]byte(bucket))

		for i := 0; i < len(keys); i++ {
			if err := bkt.Put(keys[i], values[i]); err != nil {
				return err
			}
		}

		return nil
	})
}

// Scan a group of keys with a particular prefix using BoltDB prefix seek.
// NOTE: this is dangerous thanks to the manual management of the transaction.
// Avoid if possible until a better, non-racy solution can be found.
func (bdb *BoltDB) Scan(prefix []byte, bucket string) Cursor {
	tx, err := bdb.db.Begin(false)
	if err != nil {
		return &BoltCursor{err: err}
	}

	b := tx.Bucket([]byte(bucket))
	c := b.Cursor()

	return &BoltCursor{
		tx, b, c, true, nil, prefix, nil,
	}

}

//===========================================================================
// BoltCursor type and methods
//===========================================================================

// BoltCursor implements the cursor interface, wrapping a view transaction
// with a channel so that iteration can happen safely. This might not be the
// fastest way to implement it, but it is the safest.
type BoltCursor struct {
	tx     *bolt.Tx     // Hook to the transaction object to close it.
	b      *bolt.Bucket // The bucket accessed by the transaction
	c      *bolt.Cursor // The cursor on the bucket
	seek   bool         // Whether or not we need to seek
	pair   *KVPair      // The current k/v pair on the cursor
	prefix []byte       // The prefix on the cursor
	err    error        // Any errors encountered during processing.
}

// Next returns true if there is another key/value pair available.
func (c *BoltCursor) Next() bool {
	var k, v []byte
	if c.seek {
		k, v = c.c.Seek(c.prefix)
		c.seek = false
	} else {
		k, v = c.c.Next()
	}

	if k != nil && bytes.HasPrefix(k, c.prefix) {
		c.pair = &KVPair{k, v}
		return true
	}

	c.err = c.tx.Rollback()
	c.pair = nil
	return false
}

// Pair returns the current key/value pair on the cursor.
func (c *BoltCursor) Pair() *KVPair {
	return c.pair
}

// Error returns any errors from the database transaction.
func (c *BoltCursor) Error() error {
	return c.err
}
