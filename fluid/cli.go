// Defines the CLIClient and all command and control from the command line.

package fluid

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

//===========================================================================
// Command Line Interface
//===========================================================================

// CLIClient is the primary CLI app object. It holds references and attributes
// to allow it to connect to the server if it's running and exposes a command
// API to create interfaces that can make calls to the server.
type CLIClient struct {
	PID *PID // A reference to the PID file to connect to the server.
}

// Init the CLIClient by loading the PID file and creating connections with
// the FluidFS Server. This function may return an error.
func (c *CLIClient) Init() error {

	c.PID = new(PID)
	if err := c.PID.Load(); err != nil {
		return errors.New("Could not connect to the FluidFS server: no PID file detected.")
	}

	return nil
}

// Status reports on the running FluidFS Server. Note that if the server isn't
// running, or if there is no PID file detected, then the Status message will
// not run because the error is caught/returned in Init().
func (c *CLIClient) Status() error {
	res, err := c.Get("/status/")
	if err != nil {
		return err
	}

	status := res["status"].(string)
	timestamp := res["timestamp"].(string)
	fmt.Printf("FluidFS Status: %s at %s\n", status, timestamp)
	return nil
}

// Get an http request to the FLuidFS C2S API
func (c *CLIClient) Get(endpoint string) (map[string]interface{}, error) {
	url := fmt.Sprintf("http://%s%s", c.PID.Addr(), endpoint)
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data, nil
}
