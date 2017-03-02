// Mechanisms for interacting with configuration YAML files on disk.

package fluid

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"time"

	kvdb "github.com/bbengfort/fluidfs/fluid/db"

	"gopkg.in/yaml.v2"
)

// Configuration directories and fixtures
const (
	ConfigDirectory       = "fluidfs"
	HiddenConfigDirectory = ".fluidfs"
)

//===========================================================================
// Config Structs and Interfaces
//===========================================================================

// Configuration is an interface for all Config objects and provides a
// mechanism to create nested configurations read from a YAML file.
type Configuration interface {
	Defaults() error // Updates the configuration with default values.
	Validate() error // Validates the input data, returns an error if invalid.
	Environ() error  // Updates the configuration with environment variables.
	String() string  // Print out the pretty representation of the config.
}

// Config provides the base structure for reading configuration values
// from YAML configuration files and supplies the primary inputs to the
// FluidFS server as well as connection interfaces to clients.
type Config struct {
	Seed       int64           `yaml:"seed,omitempty"`        // Control random number generation
	Name       string          `yaml:"name,omitempty"`        // The name of the replica
	Hosts      string          `yaml:"hosts,omitempty"`       // The path to the hosts file on disk
	FStab      string          `yaml:"fstab,omitempty"`       // The path to the fstab file on disk
	FlushDelay int64           `yaml:"flush_delay,omitempty"` // Milliseconds to sleep betweeen flushes
	Logging    *LoggingConfig  `yaml:"logging"`               // Configuration for logging
	Database   *DatabaseConfig `yaml:"database"`              // Database configuration
	Storage    *StorageConfig  `yaml:"storage"`               // Storage/Chunking configuration
	Loaded     []string        `yaml:"-"`                     // Reference to the loaded configuration paths

}

//===========================================================================
// Primary Config Methods
//===========================================================================

// LoadConfig is a several step function that should be the primary method of
// creating a Config object. It first sets reasonable defaults, then goes
// through all the yaml config paths and loads their configuration, then
// sets final variables from the environment before performing validation.
//
// An optional parameter, the path to another configuration can be passed in.
// This configuration will be loaded after the configuration in Paths().
func LoadConfig(confPath string) (*Config, error) {
	// Initialize the config
	conf := new(Config)

	// Load the meaningful defaults.
	if err := conf.Defaults(); err != nil {
		return nil, err
	}

	// Load the configuration from paths on disk.
	// Note errors are supressed, if no file or bad read, just keep on going.
	for _, path := range conf.Paths() {
		conf.Read(path)
	}

	// Load a configuration that is passed in, ignoring errors
	if confPath != "" {
		conf.Read(confPath)
	}

	// Load environment variables as needed
	if err := conf.Environ(); err != nil {
		return nil, err
	}

	// Finally validate the configuration
	if err := conf.Validate(); err != nil {
		return nil, err
	}

	// Return the config, all is well!
	return conf, nil
}

// Paths returns the search paths to find configuration YAML files.
func (conf *Config) Paths() []string {

	// Initialize the list of paths.
	paths := make([]string, 0, 8)

	// Add the etc configuration
	paths = append(paths, filepath.Join("/", "etc", ConfigDirectory, "config.yml"))
	paths = append(paths, filepath.Join("/", "etc", ConfigDirectory, "config.yaml"))

	// Add the user configuration
	usr, err := user.Current()
	if err == nil {
		yaml := filepath.Join(usr.HomeDir, HiddenConfigDirectory, "config.yml")
		paths = append(paths, yaml)

		yaml = filepath.Join(usr.HomeDir, HiddenConfigDirectory, "config.yaml")
		paths = append(paths, yaml)
	}

	// Add the local and development configuration
	cwd, err := os.Getwd()
	if err == nil {
		paths = append(paths, filepath.Join(cwd, "fluidfs.yml"))
		paths = append(paths, filepath.Join(cwd, "fluidfs.yaml"))
		paths = append(paths, filepath.Join(cwd, "conf", "fluidfs.yml"))
		paths = append(paths, filepath.Join(cwd, "conf", "fluidfs.yaml"))
	}

	return paths
}

// Read a YAML configuration file from a path.
func (conf *Config) Read(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	// Unmarshal the YAML data
	if err := yaml.Unmarshal(data, conf); err != nil {
		return err
	}

	// Add the path to the loaded paths if successful
	conf.Loaded = append(conf.Loaded, path)

	return nil
}

//===========================================================================
// Config Interface (Defaults and Validation)
//===========================================================================

// Defaults sets the reasonable defaults on the Config object.
func (conf *Config) Defaults() error {

	// Set the random seed to a unix timestamp.
	conf.Seed = time.Now().UnixNano()

	// Get the Hostname
	name, err := os.Hostname()
	if err == nil {
		conf.Name = name
	}

	// The default hosts path is in the user's hidden config directory: ~/.fluid/hosts
	// The default fstab path is in the user's hidden config directory: ~/.fluid/fstab
	usr, err := user.Current()
	if err == nil {
		conf.Hosts = filepath.Join(usr.HomeDir, HiddenConfigDirectory, "hosts")
		conf.FStab = filepath.Join(usr.HomeDir, HiddenConfigDirectory, "fstab")
	}

	// Set the default flush delay
	conf.FlushDelay = 750

	// Create the logging configuration and call its defaults.
	conf.Logging = new(LoggingConfig)
	conf.Logging.Defaults()

	// Create the database configuration and call its defaults.
	conf.Database = new(DatabaseConfig)
	conf.Database.Defaults()

	// Create the storage configuration and call its defaults.
	conf.Storage = new(StorageConfig)
	conf.Storage.Defaults()

	return nil
}

// Validate ensures that required settings are correctly set.
func (conf *Config) Validate() error {

	// Return an error if there is no seed
	if conf.Seed == 0 {
		return errors.New("Improperly configured: no random seed set.")
	}

	// Return an error if there is no replica name
	if conf.Name == "" {
		return errors.New("Improperly configured: a name is required.")
	}

	// Return an error if there is no fstab path
	if conf.FStab == "" {
		return errors.New("Improperly configured: an fstab path is required.")
	}

	// Return an error if flush delay is zero
	if conf.FlushDelay < 1 {
		return errors.New("Improperly configured: specify a flush delay.")
	}

	// Validate the LoggingConfig
	if err := conf.Logging.Validate(); err != nil {
		return err
	}

	// Validate the DatabaseConfig
	if err := conf.Database.Validate(); err != nil {
		return err
	}

	// Validate the StorageConfig
	if err := conf.Storage.Validate(); err != nil {
		return err
	}

	return nil
}

// Environ sets configurations from the environment.
func (conf *Config) Environ() error {
	// Make sure the logging configuration can get environment variables.
	if err := conf.Logging.Environ(); err != nil {
		return err
	}

	// Make sure the database configuration can get environment variables.
	if err := conf.Database.Environ(); err != nil {
		return err
	}

	// Make sure the storage configuration can get environment variables.
	if err := conf.Storage.Environ(); err != nil {
		return err
	}

	return nil
}

// String returns a pretty representation of the Configuration.
func (conf *Config) String() string {
	output := fmt.Sprintf("%s configuration", conf.Name)
	output += "\n" + conf.Database.String()
	output += "\n" + conf.Storage.String()
	output += "\n" + conf.Logging.String()
	return output
}

//===========================================================================
// Logging Configuration
//===========================================================================

// LoggingConfig is passed to the InitLogger function to create meaningful,
// leveled logging to a file or to stdout depending on the configuration.
type LoggingConfig struct {
	Level string `yaml:"level,omitempty"` // specifies the minimum log level
	Path  string `yaml:"path,omitempty"`  // optional path to location on disk to write file
}

// Defaults sets the reasonable defaults on the LoggingConfig object.
func (conf *LoggingConfig) Defaults() error {
	// LogLevel is INFO by default.
	conf.Level = "INFO"
	return nil
}

// Validate ensures that required logging settings are correct
func (conf *LoggingConfig) Validate() error {

	// Return an error if the log level is incorrect.
	if !ListContains(conf.Level, levelNames) {
		msg := "Improperly Configured: '%s' is not a valid log level."
		return fmt.Errorf(msg, conf.Level)
	}

	return nil
}

// Environ sets the logging conifguration from the environment.
func (conf *LoggingConfig) Environ() error {
	return nil
}

// String returns a pretty representation of the logging configuration.
func (conf *LoggingConfig) String() string {

	path := conf.Path
	if conf.Path == "" {
		path = "stdout"
	}

	output := fmt.Sprintf("%s logging to %s", conf.Level, path)
	return output
}

//===========================================================================
// Database Configuration
//===========================================================================

// DatabaseConfig is passed to the InitDatabase function to correctly open the
// right type of database and Database interface.
type DatabaseConfig struct {
	Driver string `yaml:"driver,omitempty"` // specifies the database interface to use
	Path   string `yaml:"path,omitempty"`   // optional path to location on disk to write file
}

// Defaults sets the reasonable defaults on the DatabaseConfig object.
func (conf *DatabaseConfig) Defaults() error {

	// The default driver is the boltdb driver
	conf.Driver = kvdb.BoltDBDriver

	// The default path to the database is in a hidden directory in the home
	// directory of the user, namely ~/.fluidfs/cache.db
	usr, err := user.Current()
	if err == nil {
		conf.Path = filepath.Join(usr.HomeDir, HiddenConfigDirectory, "cache.bdb")
	}

	return nil
}

// Validate ensures that required database settings are correct
func (conf *DatabaseConfig) Validate() error {

	// Ensure that the driver is all lowercase with whitespace trimmed.
	conf.Driver = Regularize(conf.Driver)

	// Ensure that the driver is in the list of drivers.
	if !ListContains(conf.Driver, kvdb.DriverNames) {
		return fmt.Errorf("Improperly configured: '%s' is not a valid database driver", conf.Driver)
	}

	// Ensure that a path has been passed in.
	if conf.Path == "" {
		return errors.New("Improperly configured: must specify a path to the database")
	}

	return nil
}

// Environ sets the database conifguration from the environment.
func (conf *DatabaseConfig) Environ() error {
	return nil
}

// String returns a pretty representation of the database configuration.
func (conf *DatabaseConfig) String() string {
	return fmt.Sprintf("%s at %s", conf.Driver, conf.Path)
}

// GetDriver implements db.Config
func (conf *DatabaseConfig) GetDriver() string {
	return conf.Driver
}

// GetPath implements db.Config
func (conf *DatabaseConfig) GetPath() string {
	return conf.Path
}

//===========================================================================
// Storage Configuration
//===========================================================================

// StorageConfig is passed to the NewChunker function to correctly initialize
// the storage and chunking mechanism for creating blobs from files.
type StorageConfig struct {
	Path         string `yaml:"path,omitempty"`       // Path to a directory to store blobs on disk
	Chunking     string `yaml:"chunking,omitempty"`   // Either "variable" (default) or "fixed"
	BlockSize    int    `yaml:"block_size,omitempty"` // The target block size for Blobs
	MinBlockSize int    `yaml:"min_block_size"`       // Used in both variable and fixed
	MaxBlockSize int    `yaml:"max_block_size"`       // Used only in variable length chunking
	Hashing      string `yaml:"hashing,omitempty"`    // Identifies the hashing algorithm used
}

// Defaults sets the reasonable defaults on the StorageConfig object.
func (conf *StorageConfig) Defaults() error {

	// The default path to the storage path is to a hidden directory in the
	// home directory of the user, namely ~/.fluidfs/data/
	usr, err := user.Current()
	if err == nil {
		conf.Path = filepath.Join(usr.HomeDir, HiddenConfigDirectory, "data")
	}

	// Default chunker is variable length Rabin-Karp chunking
	conf.Chunking = VariableLengthChunking

	// Default target block size is 4096 bytes, though where this comes from
	// I'm not sure and I'd advocate a larger target block size.
	conf.BlockSize = 4096

	// Minimum block size is half the target block size
	conf.MinBlockSize = 2048

	// Maximum block size is double the target block size
	conf.MaxBlockSize = 8192

	// Default hashing algorithm is SHA256 to prevent collisions
	conf.Hashing = SHA256

	return nil
}

// Validate ensures that required chunking settings are correct
func (conf *StorageConfig) Validate() error {

	// Return an error if there is no storage path
	if conf.Path == "" {
		return errors.New("Improperly configured: a path to the storage directory is required.")
	}

	// NOTE: The following happens in validate, e.g. ASAP so that errors happen at startup.
	// NOTE: Still need to handle errors at write time, just in case things change.
	// Create the storage path if it does not exist and validate that the user
	// has permission to read and write to the directory.
	if _, err := os.Stat(conf.Path); os.IsNotExist(err) {
		if err := os.MkdirAll(conf.Path, ModeStorageDir); err != nil {
			return fmt.Errorf("Improperly configured: could not create storage directory at '%s'", conf.Path)
		}
	}

	// Ensure that the storage path is a directory just in case.
	info, _ := os.Stat(conf.Path)
	if !info.Mode().IsDir() {
		return errors.New("Improperly configured: storage directory cannot be accessed.")
	}

	// Ensure that the chunks is all lowercase and whitespace trimmed.
	conf.Chunking = Regularize(conf.Chunking)

	// Ensure that the chunking method is in the list of available methods.
	if !ListContains(conf.Chunking, chunkingMethodNames) {
		return fmt.Errorf("Improperly configured: '%s' is not a valid chunking mechanism", conf.Chunking)
	}

	// Ensure that the block size is greater than zero.
	if conf.BlockSize < 1 {
		return errors.New("Improperly configured: must specify a block size greater than 0 bytes.")
	}

	// Ensure that MaxBlockSize is greater than target and minimum.
	if conf.MaxBlockSize < conf.BlockSize || conf.MaxBlockSize < conf.MinBlockSize {
		return errors.New("Improperly configured: maximum block size must be greater than or equal target and minimum block sizes.")
	}

	// Ensure that MinBlockSize is less than the target and maximum.
	if conf.MinBlockSize > conf.BlockSize || conf.MinBlockSize > conf.MaxBlockSize {
		return errors.New("Improperly configured: minimum block size must be less than or equal to the target block size.")
	}

	// Ensure that hashing is all lowercase and whitespace trimmed.
	conf.Hashing = Regularize(conf.Hashing)

	// Ensure that the hashing algorithm is in the list of available algorithms.
	if !ListContains(conf.Hashing, hashingAlgorithmNames) {
		return fmt.Errorf("Improperly configured: '%s' is not a valid hashing algorithm", conf.Hashing)
	}

	return nil
}

// Environ sets the storage conifguration from the environment.
func (conf *StorageConfig) Environ() error {
	return nil
}

// String returns a pretty representation of the storage configuration.
func (conf *StorageConfig) String() string {
	return fmt.Sprintf("%s length %d byte blobs stored at %s", conf.Chunking, conf.BlockSize, conf.Path)
}
