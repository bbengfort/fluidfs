// Package fluid provides the core functionality for the fluidfs replica
// daemon and the fluid client as well as secondary functionality including
// the web interface, global configuration service and other data services.
package fluid

import (
	"fmt"

	"github.com/bbengfort/fluidfs/fluid/db"
)

//===========================================================================
// FluidFS Replica
//===========================================================================

// Replica represents the primary application object. All application
// interactions must pass through an instance of the Fluid Replica.  On init
// the Fluid Replica loads the configuration, instantiates logging and database
// connections, then can be run in the background with various method calls
// from external service requests or other environmental detection.
type Replica struct {
	PID    *PID         // Process ID and C&C information
	Config *Config      // The application configuration
	FS     *FuseFSTable // Mount Points and FS handling
	Logger *Logger      // Application logging and reporting
	DB     db.Database  // A connection to the database
	Web    *C2SAPI      // The listener for command and control.
}

// Init prepares the replica for running by loading the configuration and
// setting up the logging handlers and other utilities. Note that this method
// does not write a PID file or open connections to databases, these items are
// handled when the Replica is run, allowing non-destructive post-config tasks.
//
// Can optionally pass the path of a YAML configuration file on disk. Any
// configurations in that file will superceede those in the etc directory or
// in the user's home directory.
func (r *Replica) Init(conf string) error {
	var err error

	// Load the configuration from YAML files on disk.
	r.Config, err = LoadConfig(conf)
	if err != nil {
		return err
	}

	// Load the logger from the logging configuration.
	r.Logger, err = InitLogger(r.Config.Logging)
	if err != nil {
		return err
	}

	// Log the initialization from the loaded configurations.
	for _, path := range r.Config.Loaded {
		r.Logger.Info("loaded configuration from %s", path)
	}

	// Initialize the FSTable from the fstab path
	r.FS = new(FuseFSTable)
	if err = r.FS.Load(r.Config.FStab); err != nil {
		return err
	}

	// Initialize the C2S API
	r.Web = new(C2SAPI)
	if err = r.Web.Init(r); err != nil {
		return err
	}

	return nil
}

// Run the replica by creating a PID file, listening for command and control,
// opening connections to databases, mounting the FUSE directories, and
// listening for remote connections.
func (r *Replica) Run() error {
	var err error

	// Handle any OS Signals
	go signalHandler(r)

	// Create X-Thread and X-Process Resources

	// Create a PID file
	r.PID = new(PID)
	if err = r.PID.Save(); err != nil {
		return fmt.Errorf("could not write PID file: %s", err.Error())
	}

	// Log the creation of the PID file
	r.Logger.Info("pid file created at %s", r.PID.Path())

	// Open a connection to the database
	r.DB, err = db.InitDatabase(r.Config.Database)
	if err != nil {
		return fmt.Errorf("could not connect to database: %s", err.Error())
	}

	// Log the connection to the database
	r.Logger.Info("connected to %s", r.Config.Database.String())

	// Create the error channel for go routines
	echan := make(chan error)

	// Run services in independent go routines.

	// Run the FUSE File Systems
	if err = r.FS.Run(r, echan); err != nil {
		return fmt.Errorf("could not run the FUSE file system: %s", err.Error())
	}

	// Run the C2S API and web interface
	go r.Web.Run(r.PID.Addr(), echan)

	// Listen for an error from any of the go routines (also blocks)
	// Log the error and shutdown gracefully (returning only shutdown errors).
	err = <-echan
	r.Logger.Error(err.Error())
	return r.Shutdown()
}

// Shutdown the replica gracefully by unmounting FUSE directories, closing
// database connections, closing listeners for command and control, and
// deleting the PID file, basically the reverse order of startup.
func (r *Replica) Shutdown() error {
	r.Logger.Warn("starting shutdown process")

	// Close all the FUSE FS Connections
	if err := r.FS.Shutdown(); err != nil {
		r.Logger.Warn("could not shutdown FS: %s", err.Error())
	}

	// Close the Database
	if err := r.DB.Close(); err != nil {
		r.Logger.Warn("could not shutdown database: %s", err.Error())
	}

	// Free the PID file
	if err := r.PID.Free(); err != nil {
		// If we can't free the PID file, then we have a problem.
		return err
	}

	r.Logger.Info("fluidfs successfully shutdown")
	return nil
}
