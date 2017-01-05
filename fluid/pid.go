// Handling process id information and enables cross-process communication
// between the server and the command line client.

package fluid

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"syscall"
)

//===========================================================================
// OS Signal Handlers
//===========================================================================

func signalHandler(s *Server) {
	// Make signal channel and register notifiers for Interupt and Terminate
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt)
	signal.Notify(sigchan, syscall.SIGTERM)

	// Block until we receive a signal on the channel
	<-sigchan

	// Defer the clean exit until the end of the function
	defer os.Exit(0)

	// Shutdown now that we've received the signal
	err := s.Shutdown()
	if err != nil {
		s.Logger.Fatal("could not shutdown %s", err.Error())
		os.Exit(1)
	}
}

//===========================================================================
// PID File Management
//===========================================================================

// PID describes the server process and is accessed by both the server and the
// command line client in order to facilitate cross-process communication.
type PID struct {
	PID  int `json:"pid"`  // The process id assigned by the OS
	PPID int `json:"ppid"` // The parent process id
	Port int `json:"port"` // The command port for client-server communication
}

// Path returns the best possible PID file for the current system, by first
// attempting to get the user directory then resorting to /var/run on Linux
// systems and elsewhere on other systems.
//
// Note that this method should always return a single PID path for a running
// instance of the FluidFS file system in order to prevent confusion.
func (pid *PID) Path() string {
	filename := "fluid.pid"

	usr, err := user.Current()
	if err == nil {
		return filepath.Join(usr.HomeDir, HiddenConfigDirectory, filename)
	}

	return filepath.Join("/", "var", "run", filename)
}

// Save the PID file to disk after first determining the process id and the
// command port -- used by the server on startup to allow clients to connect.
// NOTE: This method will fail if the PID file already exists.
func (pid *PID) Save() error {
	var err error

	// Get the currently running Process ID and Parent ID
	pid.PID = os.Getpid()
	pid.PPID = os.Getppid()

	// Get any available Port for communication
	pid.Port, err = pid.FreePort()
	if err != nil {
		return err
	}

	// Marshall the JSON representation
	data, err := json.Marshal(pid)
	if err != nil {
		return err
	}

	path := pid.Path()
	// Ensure that a PID file does not exist (race possible)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Write the JSON representation of the PID file to disk
		return ioutil.WriteFile(path, data, ModeBlob)
	}

	return fmt.Errorf("PID file exists already at '%s'", path)
}

// Load the PID file -- used by the command line client to populate the PID.
func (pid *PID) Load() error {
	data, err := ioutil.ReadFile(pid.Path())
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &pid)
}

// Free the PID file (delete it) -- used by the server on shutdown to cleanup
// and ensure that stray process information isn't just lying about.
// Does not return an error if the PID file does not exist.
func (pid *PID) Free() error {
	// If the PID file doesn't exist, just ignore and return.
	if _, err := os.Stat(pid.Path()); os.IsNotExist(err) {
		return nil
	}

	// Delete the PID file
	return os.Remove(pid.Path())
}

// FreePort asks the kernel for a free, open port that is ready to use.
// https://github.com/phayes/freeport
func (pid *PID) FreePort() (int, error) {

	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	listen, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}

	defer listen.Close()
	return listen.Addr().(*net.TCPAddr).Port, nil
}
