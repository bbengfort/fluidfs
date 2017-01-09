// Implements the JSON Command, Config, and Status API as well as the HTTP
// handlers for the Web interface. Note that the CLI client also uses the web
// API for command, config, and status (C2S) interfaction.

package fluid

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

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
)

//===========================================================================
// Stand-Alone C2S API and Web Server
//===========================================================================

// C2SAPI implements the web server for the command, config, and status JSON
// API that serves both a web interface and the command line client.
type C2SAPI struct {
	Router *mux.Router
	Fluid  *Server
}

// Init the C2SAPI with a hook to the server that the API wraps.
func (api *C2SAPI) Init(fluid *Server) error {
	// Initialize the API
	api.Fluid = fluid
	api.Router = mux.NewRouter().StrictSlash(true)

	// Add handlers and routes
	api.AddHandler(StatusEndpoint, api.StatusHandler)

	// Add the static files service from the binary assets
	api.Router.Handle(RootEndpoint, WebLogger(api.Fluid.Logger, http.FileServer(assetFS())))

	// No errors occurred
	return nil
}

// Run the API at the specified address.
func (api *C2SAPI) Run(addr string) error {

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
	return srv.ListenAndServe()
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
	data["timestamp"] = time.Now().String()
	return http.StatusOK, data, nil
}
