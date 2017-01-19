// Implements the JSON Command, Config, and Status API as well as the HTTP
// handlers for the Web interface. Note that the CLI client also uses the web
// API for command, config, and status (C2S) interfaction.

package fluid

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

//===========================================================================
// Web Types and Constants
//===========================================================================

// JSON is a helper type for the most generic json object decoded by the json
// package, namely map[string]interface{}.
type JSON map[string]interface{}

// APIHandler is a function type that defines how C2SAPI handlers are to be
// implemented. Functions of t his type are passed to the AddHandler() method
// of the API so that they can be routed on.
type APIHandler func(r *http.Request) (int, JSON, error)

// Request Header Keys and Values
const (
	HeaderAcceptKey      = "Accept"
	HeaderContentTypeKey = "Content-Type"
	HeaderContentTypeVal = "application/json;charset=UTF-8"
	HeaderVersionKey     = "X-FluidFS-Application"
	HeaderVersionVal     = "FluidFS/v%s"
)

// Define endpoint locations and names.
const (
	RootEndpoint   = "/"
	StatusEndpoint = "/status"
	MountEndpoint  = "/mounts"
)

//===========================================================================
// Stand-Alone C2S API and Web Server
//===========================================================================

// C2SAPI implements the web server for the command, config, and status JSON
// API that serves both a web interface and the command line client.
type C2SAPI struct {
	Router *mux.Router
	Fluid  *Replica
}

// Init the C2SAPI with a hook to the server that the API wraps.
func (api *C2SAPI) Init(fluid *Replica) error {
	// Initialize the API
	api.Fluid = fluid
	api.Router = mux.NewRouter().StrictSlash(true)

	// Add handlers and routes
	api.AddHandler(StatusEndpoint, api.StatusHandler)
	api.AddHandler(MountEndpoint, api.MountHandler)

	// Add the static files service from the binary assets
	api.Router.Handle(RootEndpoint, WebLogger(api.Fluid.Logger, http.FileServer(assetFS())))

	// No errors occurred
	return nil
}

// Run the API at the specified address.
func (api *C2SAPI) Run(addr string, echan chan error) {

	// Create the HTTP server
	srv := &http.Server{
		Handler:      api.Router,
		Addr:         addr,
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}

	// Report the server status
	api.Fluid.Logger.Info("starting C2S API and web interface at http://%s/", addr)

	// Listen and Serve
	if err := srv.ListenAndServe(); err != nil {
		echan <- fmt.Errorf("C2S API error: %s", err.Error())
	}
}

// AddHandler adds the specified handler to the API.
// TODO: Simply this function and decouple various optional methods.
func (api *C2SAPI) AddHandler(path string, inner APIHandler) {
	outer := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		code, data, err := inner(r)

		// Handle errors
		if err != nil {
			if code == 0 {
				code = http.StatusInternalServerError
			}

			// Make the data a JSON error representation.
			data = make(JSON)
			data["code"] = strconv.Itoa(code)
			data["error"] = err.Error()
		}

		// Otherwise respond to the request
		w.Header().Set(HeaderContentTypeKey, HeaderContentTypeVal)
		w.WriteHeader(code)

		// Write the responde data as JSON to the stream
		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	handler := WebLogger(api.Fluid.Logger, outer)
	api.Router.Handle(path, handler)
}

//===========================================================================
// API Handlers
//===========================================================================

// StatusHandler returns the status command information.
func (api *C2SAPI) StatusHandler(r *http.Request) (int, JSON, error) {
	data := make(JSON)
	data["status"] = "ok"
	data["timestamp"] = time.Now().Format(JSONDateTime)
	data["mounts"] = api.Fluid.FS.Status()
	return http.StatusOK, data, nil
}

// MountHandler accepts POST data for the mount command and creates a new
// MountPoint, then saves the fstable to disk, returning success or error.
// TODO: Make this a fully formed RESTful API. See #39
// TODO: Unhack this!
func (api *C2SAPI) MountHandler(r *http.Request) (int, JSON, error) {
	// Parse the JSON data from the request
	req, err := readRequestJSON(r)
	if err != nil {
		return http.StatusInternalServerError, nil, err
	}

	var (
		path   string
		prefix string
		uid    float64
		gid    float64
		ok     bool
	)
	fmt.Println(req)

	// Validate the passed in arguments
	if path, ok = req["path"].(string); !ok {
		return http.StatusBadRequest, nil, errors.New("missing required path argument")
	}

	if prefix, ok = req["prefix"].(string); !ok {
		return http.StatusBadRequest, nil, errors.New("missing required prefix argument")
	}

	if uid, ok = req["uid"].(float64); !ok {
		return http.StatusBadRequest, nil, errors.New("missing required uid argument")
	}

	if gid, ok = req["gid"].(float64); !ok {
		return http.StatusBadRequest, nil, errors.New("missing required gid argument")
	}

	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return http.StatusBadRequest, nil, fmt.Errorf("mount path '%s' does not exist", path)
	}

	if !info.IsDir() {
		return http.StatusBadRequest, nil, fmt.Errorf("mount path '%s' is not a directory", path)
	}

	if strings.HasPrefix(prefix, "/") {
		return http.StatusBadRequest, nil, errors.New("prefix cannot start with '/'")
	}

	// Create the mount point now that we've validated it (more or less)
	mp := &MountPoint{
		UUID:      uuid.New(),
		Path:      path,
		Prefix:    prefix,
		UID:       int(uid),
		GID:       int(gid),
		Store:     true,
		Replicate: true,
		Comments:  make([]string, 0, 0),
		Options:   []string{"defaults"},
	}

	// Add the mount point to the fs table.
	if err := api.Fluid.FS.AddMountPoint(mp); err != nil {
		return http.StatusInternalServerError, nil, err
	}

	// Return the response with the created mount point
	data := make(JSON)
	data["mount"] = mp.String()
	return http.StatusOK, data, nil
}

//===========================================================================
// Helper functions
//===========================================================================

// Helper function to decode the JSON in an HTTP Request
func readRequestJSON(r *http.Request) (JSON, error) {
	// Read the data from the request stream (limit the size to 100 MB)
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 104857600))
	if err != nil {
		return nil, err
	}

	// Attempt to close the body of the request for reading
	if err := r.Body.Close(); err != nil {
		return nil, err
	}

	// Unmarshall the JSON data into a JSON data structure
	var data JSON
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	return data, nil
}
