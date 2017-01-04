// Defines the interface for database interaction for our key/value store.

package fluid

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

// Names of available drivers for validation
var databaseDriverNames = []string{BoltDBDriver, LevelDBDriver}

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
	Scan(prefix []byte, bucket string) (*Cursor, error)        // Scan a group of keys with a particular prefix
	Keys(bucket string) (*Cursor, error)                       // Returns all the keys for a bucket
}

// Cursor is an interator interface that enables iteration/search over
// multiple key/value pairs with a single query.
type Cursor interface {
	Next() ([]byte, []byte) // Returns the next key/value pair
	Error() error           // Returns any errors on the cursor
}

//===========================================================================
// Global Database API methods
//===========================================================================

// InitDatabase uses a database configuration object to select an appropriate
// driver that implements the Database interface and initializes it.
func InitDatabase(config *DatabaseConfig) (Database, error) {

	var db Database

	switch config.Driver {
	case BoltDBDriver:
		db = new(BoltDB)
	case LevelDBDriver:
		db = new(LevelDB)
	default:
		return nil, fmt.Errorf("unknown database driver: '%s'", config.Driver)
	}

	// Initialize the Database
	err := db.Init(config.Path)
	return db, err
}
