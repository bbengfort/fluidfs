// Package fluid provides the core functionality for the fluidfs replica
// daemon and the fluid client as well as secondary functionality including
// the web interface, global configuration service and other data services.
package fluid

import (
	"fmt"
	"math/rand"

	kvdb "github.com/bbengfort/fluidfs/fluid/db"
)

const (
	programName  = "fluidfs"
	majorVersion = 0
	minorVersion = 1
	microVersion = 0
	releaseLevel = "final"
)

var (
	pid    *PID          // Process ID and C&C information
	config *Config       // The application configuration
	fstab  *FuseFSTable  // Mount Points and FS handling
	hosts  *Hosts        // Describes members of the network
	local  *Replica      // Describes the locally running Replica
	logger *Logger       // Application logging and reporting
	db     kvdb.Database // A connection to the database
	web    *C2SAPI       // The listener for command and control.
)

//===========================================================================
// Package Meta
//===========================================================================

// PackageVersion composes version information from the constants in this package
// and returns a string that defines current information about the package.
func PackageVersion() string {
	vstr := fmt.Sprintf("%d.%d", majorVersion, minorVersion)

	if microVersion > 0 {
		vstr += fmt.Sprintf(".%d", microVersion)
	}

	switch releaseLevel {
	case "final":
		return vstr
	case "alpha":
		return vstr + "a"
	case "beta":
		return vstr + "b"
	default:
		return vstr
	}

}

//===========================================================================
// FluidFS Server Functions
//===========================================================================

// Init prepares the replica for running by loading the configuration and
// setting up the logging handlers and other utilities. Note that this method
// does not write a PID file or open connections to databases, these items are
// handled when the Replica is run, allowing non-destructive post-config tasks.
//
// Can optionally pass the path of a YAML configuration file on disk. Any
// configurations in that file will superceede those in the etc directory or
// in the user's home directory.
func Init(conf string) error {
	var err error

	// Load the configuration from YAML files on disk.
	config, err = LoadConfig(conf)
	if err != nil {
		return err
	}

	// Set the random seed for things that require randomness.
	rand.Seed(config.Seed)

	// Load the logger from the logging configuration.
	logger, err = InitLogger(config.Logging)
	if err != nil {
		return err
	}

	// Log the initialization from the loaded configurations.
	for _, path := range config.Loaded {
		logger.Info("loaded configuration from %s", path)
	}

	// Initialize the FSTable from the fstab path
	fstab = new(FuseFSTable)
	if err = fstab.Load(config.FStab); err != nil {
		return fmt.Errorf("could not load fstab: %s", err)
	}

	// Initialize the Hosts from the hosts path
	hosts = new(Hosts)
	if err = hosts.Load(config.Hosts); err != nil {
		return fmt.Errorf("could not load hosts: %s", err)
	}

	// Initialize the local replica
	local, err = hosts.Local()
	if err != nil {
		return fmt.Errorf("could not initialize local replica: %s", err)
	}
	logger.Info("local replica: %s with precedence %d", local, local.Precedence)

	// Initialize the C2S API
	web = new(C2SAPI)
	if err = web.Init(); err != nil {
		return fmt.Errorf("could not initialize web api: %s", err)
	}

	return nil
}

// Run the replica by creating a PID file, listening for command and control,
// opening connections to databases, mounting the FUSE directories, and
// listening for remote connections.
func Run() error {
	var err error

	// Handle any OS Signals
	go signalHandler()

	// Create X-Thread and X-Process Resources

	// Create a PID file
	pid = new(PID)
	if err = pid.Save(); err != nil {
		return fmt.Errorf("could not write PID file: %s", err.Error())
	}

	// Log the creation of the PID file
	logger.Info("pid file created at %s", pid.Path())

	// Open a connection to the database
	db, err = kvdb.InitDatabase(config.Database)
	if err != nil {
		return fmt.Errorf("could not connect to database: %s", err.Error())
	}

	// Log the connection to the database
	logger.Info("connected to %s", config.Database.String())

	// Create the error channel for go routines
	echan := make(chan error)

	// Run services in independent go routines.

	// Run the FUSE File Systems
	if err = fstab.Run(echan); err != nil {
		return fmt.Errorf("could not run the FUSE file system: %s", err.Error())
	}

	// Run the C2S API and web interface
	go web.Run(pid.Addr(), echan)

	// Run the Flusher
	go Flusher(config.FlushDelay, echan)

	// Run anti-entropy
	go RunAntiEntropy(echan)

	// Listen for an error from any of the go routines (also blocks)
	// Log the error and shutdown gracefully (returning only shutdown errors).
	err = <-echan
	logger.Error(err.Error())
	return Shutdown()
}

// Shutdown the replica gracefully by unmounting FUSE directories, closing
// database connections, closing listeners for command and control, and
// deleting the PID file, basically the reverse order of startup.
func Shutdown() error {
	logger.Warn("starting shutdown process")

	// Close all the FUSE FS Connections
	if err := fstab.Shutdown(); err != nil {
		logger.Warn("could not shutdown FS: %s", err.Error())
	}

	// Close the Database
	if err := db.Close(); err != nil {
		logger.Warn("could not shutdown database: %s", err.Error())
	}

	// Free the PID file
	if err := pid.Free(); err != nil {
		// If we can't free the PID file, then we have a problem.
		return err
	}

	logger.Info("fluidfs successfully shutdown")
	return nil
}

// ShowConfig returns the string representation of the current configuration
// of the FluidFS server. Useful for debugging and locating configurations.
func ShowConfig() string {
	return config.String()
}
