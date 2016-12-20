// Mechanisms for interacting with configuration YAML files on disk.

package fluid

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"

	"gopkg.in/yaml.v2"
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
	Debug    bool           `yaml:"debug"`    // Whether or not we're in debug mode
	PID      int            `yaml:"pid"`      // Used to determine replica presidence
	Hostname string         `yaml:"hostname"` // The name of the local device
	Addr     string         `yaml:"addr"`     // The listen address of the local device
	Logging  *LoggingConfig `yaml:"logging"`  // Configuration for logging
}

//===========================================================================
// Primary Config Methods
//===========================================================================

// LoadConfig is a several step function that should be the primary method of
// creating a Config object. It first sets reasonable defaults, then goes
// through all the yaml config paths and loads their configuration, then
// sets final variables from the environment before performing validation.
func LoadConfig() (*Config, error) {
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
	paths := make([]string, 0, 4)

	// Add the etc configuration
	paths = append(paths, "/etc/fluid/fluidfs.yml")

	// Add the user configuration
	usr, err := user.Current()
	if err == nil {
		yaml := filepath.Join(usr.HomeDir, ".fluid", "fluidfs.yml")
		paths = append(paths, yaml)
	}

	// Add the local and development configuration
	cwd, err := os.Getwd()
	if err == nil {
		paths = append(paths, filepath.Join(cwd, "fluidfs.yml"))
		paths = append(paths, filepath.Join(cwd, "conf", "fluidfs.yml"))
	}

	return paths
}

// Read a YAML configuration file from a path.
func (conf *Config) Read(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	// Unmarshall the YAML data
	if err := yaml.Unmarshal(data, conf); err != nil {
		return err
	}

	return nil
}

//===========================================================================
// Config Interface (Defaults and Validation)
//===========================================================================

// Defaults sets the reasonable defaults on the Config object.
func (conf *Config) Defaults() error {
	// Debug is false by default
	conf.Debug = false

	// Get the Hostname
	name, err := os.Hostname()
	if err == nil {
		conf.Hostname = name
	}

	// Create the logging configuration and call its defaults.
	conf.Logging = new(LoggingConfig)
	conf.Logging.Defaults()

	return nil
}

// Validate ensures that required settings are correctly set.
func (conf *Config) Validate() error {

	// Return an error if there is no PID
	if conf.PID == 0 {
		return errors.New("Improperly configured: no precedence ID (pid) set.")
	}

	// Return an error if there is no Hostname
	if conf.Hostname == "" {
		return errors.New("Improperly configured: a hostname is required.")
	}

	// Validate the LoggingConfig
	if err := conf.Logging.Validate(); err != nil {
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

	return nil
}

// String returns a pretty representation of the Configuration.
func (conf *Config) String() string {
	output := fmt.Sprintf("%s configuration (debug = %t)", conf.Hostname, conf.Debug)
	output += conf.Logging.String()
	return output
}

//===========================================================================
// Logging Configuration
//===========================================================================

// LoggingConfig is passed to the InitLogger function to create meaningful,
// leveled logging to a file or to stdout depending on the configuration.
type LoggingConfig struct {
	Level string `yaml:"level"` // specifies the minimum log level
	Path  string `yaml:"path"`  // optional path to location on disk to write file
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
