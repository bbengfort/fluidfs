// Defines the CLIClient and all command and control from the command line.

package fluid

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

//===========================================================================
// Command Line Interface
//===========================================================================

// CLIClient is the primary CLI app object. It holds references and attributes
// to allow it to connect to the server if it's running and exposes a command
// API to create interfaces that can make calls to the server.
type CLIClient struct {
	PID    *PID         // A reference to the PID file to connect to the server.
	client *http.Client // Internal HTTP Client to make requests to the server.
}

// Init the CLIClient by loading the PID file and creating connections with
// the FluidFS Server. This function may return an error.
func (c *CLIClient) Init() error {

	// Load the PID file to detect the location to query the web service.
	c.PID = new(PID)
	if err := c.PID.Load(); err != nil {
		return errors.New("Could not connect to the FluidFS server: no PID file detected.")
	}

	// Create an HTTP client with a 30 second timeout.
	c.client = &http.Client{
		Timeout: 30 * time.Second,
	}

	return nil
}

//===========================================================================
// CLI Commands
//===========================================================================

// Status reports on the running FluidFS Server. Note that if the server isn't
// running, or if there is no PID file detected, then the Status message will
// not run because the error is caught/returned in Init().
func (c *CLIClient) Status() error {
	res, err := c.Get(StatusEndpoint)
	if err != nil {
		return err
	}

	status := res["status"].(string)
	timestamp := res["timestamp"].(string)
	mounts := res["mounts"].(string)
	fmt.Printf("FluidFS Status: %s at %s\n%s\n", status, timestamp, mounts)
	return nil
}

// Mount adds a new mount point to the FluidFS Server.
// Right now this doesn't allow very much flexibility, you can only create a
// mount with the specified path and prefix, the UID and GID is set from the
// user that calls the command, a UUID is generated, and all other options are
// set to reasonable defaults.
// TODO: add mount and umount commands, see #39
// TODO: unhack this!
func (c *CLIClient) Mount(path string, prefix string) error {
	data := make(JSON)
	data["path"] = path
	data["prefix"] = prefix
	data["uid"] = os.Geteuid()
	data["gid"] = os.Getegid()

	res, err := c.Post(MountEndpoint, data)
	if err != nil {
		return fmt.Errorf("could not post request to fluidfs: %s", err.Error())
	}

	mp := res["mount"].(string)
	fmt.Printf("created mount point for fluid://%s at %s:\n%s\n", prefix, path, mp)
	return nil
}

// Web returns the address to the web interface. It also uses an operating
// system specific helper program to open the URL on demand. If the command
// is unable to open the browser, it will simply ignore the exec error.
func (c *CLIClient) Web() error {
	var err error
	addr := c.Endpoint(RootEndpoint).String()

	// Notify the user of the web browser.
	fmt.Printf("Access the FluidFS web interface at %s\n", addr)

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", addr).Start()
	case "windows", "darwin":
		err = exec.Command("open", addr).Start()
	default:
		err = fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	if err != nil {
		fmt.Printf("Could not open web browser: %s\n", err.Error())
	}

	return nil
}

//===========================================================================
// CLIClient Helper Functions
//===========================================================================

// Endpoint constructs an absolute URL to the specified C2SAPI resource in a
// similar fashion to filepath.Join. This method also ensures that the
// endpoint is well-formed and valid, return a url.URL that can be modfied to
// add a query string or any other helper functions down the line.
func (c *CLIClient) Endpoint(resource string, detail ...string) *url.URL {
	var path string

	if len(detail) > 0 {
		parts := append([]string{resource}, detail...)
		path = filepath.Join(parts...)
	} else {
		path = resource
	}

	return &url.URL{
		Scheme: "http",
		Host:   c.PID.Addr(),
		Path:   path,
	}
}

// Do executes a request with the internal client, ensuring that all necessary
// headers are set and that any required authentication is added.
// TODO: Ensure that the server verifies the version information.
func (c *CLIClient) Do(request *http.Request) (*http.Response, error) {
	// Add the application version header and content type
	request.Header.Set(HeaderAcceptKey, HeaderContentTypeVal)
	request.Header.Set(HeaderVersionKey, fmt.Sprintf(HeaderVersionVal, Version()))

	// Execute the request
	return c.client.Do(request)
}

// Get makes an http GET request to the FLuidFS C2S API resource or command
// along with any specified details to the endpoint. The Get function returns
// arbitrary JSON data. It is up to the caller to parse and handle responses.
// TODO: allow the specificiation to submit a query string.
// TODO: return an error if the server returns an error.
func (c *CLIClient) Get(resource string, detail ...string) (JSON, error) {
	// Construct the URL and the HTTP request
	url := c.Endpoint(resource, detail...)
	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	// Execute the HTTP request
	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	// Parse the JSON response and return
	data := make(JSON)
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, err
	}

	// Check if an error has occurred
	if res.StatusCode != http.StatusOK {
		msg, ok := data["error"].(string)
		if ok {
			return data, errors.New(msg)
		}

		return data, errors.New(res.Status)
	}

	return data, nil
}

// Post an http POST request along with JSON data to the FluidFS C2S API
// resource or command. Teh Post function returns JSON data from the server,
// it's up to the caller to handle responses.
func (c *CLIClient) Post(resource string, data JSON, detail ...string) (JSON, error) {
	// Get the URL with the associated endpoint
	url := c.Endpoint(resource, detail...)

	// Marshall the POST data into a byte buffer
	body := new(bytes.Buffer)
	if err := json.NewEncoder(body).Encode(data); err != nil {
		return nil, err
	}

	// Create the POST request
	req, err := http.NewRequest(http.MethodPost, url.String(), body)
	if err != nil {
		return nil, err
	}

	// Set any necessary headers
	req.Header.Set(HeaderContentTypeKey, HeaderContentTypeVal)

	// Execute the http request
	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	// Parse the JSON response
	var resData JSON
	if err := json.NewDecoder(res.Body).Decode(&resData); err != nil {
		return nil, err
	}

	// Check if an error has occurred
	if res.StatusCode != http.StatusOK {
		msg, ok := resData["error"].(string)
		if ok {
			return resData, errors.New(msg)
		}

		return resData, errors.New(res.Status)
	}

	return resData, nil
}
