// Package fluid provides the core functionality for the fluidfs replica
// daemon and the fluid client as well as secondary functionality including
// the web interface, global configuration service and other data services.
package fluid

import (
	"fmt"
	"time"
)

//===========================================================================
// FluidFS Server
//===========================================================================

// Server represents the primary application object. All application
// interactions must pass through an instance of the FluidServer.  On init
// the FluidServer loads the configuration, instantiates logging and database
// connections, then can be run in the background with various method calls
// from external service requests or other environmental detection.
type Server struct {
	PID    *PID     // Process ID and C&C information
	Config *Config  // The application configuration
	Logger *Logger  // Application logging and reporting
	DB     Database // A connection tot he database
}

// Init prepares the server for running by loading the configuration and
// setting up the logging handlers and other utilities. Note that this method
// does not write a PID file or open connections to databases, these items are
// handled when the Server is run, allowing non-destructive post-config tasks.
//
// Can optionally pass the path of a YAML configuration file on disk. Any
// configurations in that file will superceede those in the etc directory or
// in the user's home directory.
func (s *Server) Init(conf string) error {
	var err error

	// Load the configuration from YAML files on disk.
	s.Config, err = LoadConfig(conf)
	if err != nil {
		return err
	}

	// Load the logger from the logging configuration.
	s.Logger, err = InitLogger(s.Config.Logging)
	if err != nil {
		return err
	}

	// Log the initialization from the loaded configurations.
	for _, path := range s.Config.Loaded {
		s.Logger.Info("loaded configuration from %s", path)
	}

	return nil
}

// Run the server by creating a PID file, listening for command and control,
// opening connections to databases, mounting the FUSE directories, and
// listening for remote connections.
func (s *Server) Run() error {
	var err error

	// TODO: How to return any shutdown errors?
	defer s.Shutdown()

	// Handle any OS Signals
	go signalHandler(s)

	// Create a PID file
	s.PID = new(PID)
	if err = s.PID.Save(); err != nil {
		return fmt.Errorf("could not write PID file: %s", err.Error())
	}

	// Log the creation of the PID file
	s.Logger.Info("pid file created at %s", s.PID.Path())

	// Open a connection to the database
	s.DB, err = InitDatabase(s.Config.Database)
	if err != nil {
		return fmt.Errorf("could not connect to database: %s", err.Error())
	}

	// Log the connection to the database
	s.Logger.Info("connected to %s", s.Config.Database.String())

	// Just wait until told to stop (for now)
	counter := 0
	for {
		time.Sleep(100 * time.Millisecond)
		counter++
		if counter >= 600 {
			break
		}
	}

	return nil
}

// Shutdown the server gracefully by unmounting FUSE directories, closing
// database connections, closing listeners for command and control, and
// deleting the PID file, basically the reverse order of startup.
func (s *Server) Shutdown() error {
	s.Logger.Warn("starting shutdown process")

	// Close the Database
	if err := s.DB.Close(); err != nil {
		return err
	}

	// Free the PID file
	if err := s.PID.Free(); err != nil {
		return err
	}

	s.Logger.Info("fluidfs successfully shutdown")
	return nil
}
