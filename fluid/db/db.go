// Package db defines a standard interface for interaction with the FluidFS
// cache, implemented as an embedded key/value store. This package also
// provides driver implementations for BoltDB and LevelDB.
package db

import "fmt"

// Bucket names or prefixes used in the FluidFS application
const (
	NamesBucket    = "names"
	VersionsBucket = "versions"
	PrefixesBucket = "prefixes"
)

// Driver names for quick lookups and references
const (
	BoltDBDriver  = "boltdb"
	LevelDBDriver = "leveldb"
)

// DriverNames of available drivers for validation
var DriverNames = []string{BoltDBDriver, LevelDBDriver}

//===========================================================================
// Database Interfaces
//===========================================================================

// Database defines the methods required by a key/value store to be used by
// the fluid API. We use a general key/value framework with buckets.
type Database interface {
	Init(path string) error                                    // Open a connection to the database and configure
	Close() error                                              // Close the connection to the database
	Get(key []byte, bucket string) ([]byte, error)             // Get a value for a key from a bucket
	Put(key []byte, value []byte, bucket string) error         // Put a key/value pair into the bucket
	Delete(key []byte, bucket string) error                    // Delete a key from a bucket
	Batch(keys [][]byte, values [][]byte, bucket string) error // Batch insert key/value pairs into a bucket
	Scan(prefix []byte, bucket string) Cursor                  // Scan a group of keys with a particular prefix
	Count(bucket string) (uint64, error)                       // Returns the number of keys in the bucket
}

// Config defines a methods that a struct should provide to be considered a
// database configuration, and to pass options to initialization.
type Config interface {
	GetDriver() string // Return a string representing a driver
	GetPath() string   // Return the path to the database on disk
}

// Cursor is an interator interface that enables iteration/search over
// multiple key/value pairs with a single query.
type Cursor interface {
	Next() bool    // True if there is another k/v pair
	Pair() *KVPair // Returns the next k/v pair
	Error() error  // Returns any errors on the cursor
}

// KVPair is a struct for holding key/value pairs
type KVPair struct {
	key []byte
	val []byte
}

//===========================================================================
// Global Database API methods
//===========================================================================

// InitDatabase uses a database configuration object to select an appropriate
// driver that implements the Database interface and initializes it.
func InitDatabase(config Config) (Database, error) {

	var db Database

	switch config.GetDriver() {
	case BoltDBDriver:
		db = new(BoltDB)
	case LevelDBDriver:
		db = new(LevelDB)
	default:
		return nil, fmt.Errorf("unknown database driver: '%s'", config.GetDriver())
	}

	// Initialize the Database
	err := db.Init(config.GetPath())
	return db, err
}
