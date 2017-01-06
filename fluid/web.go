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
// FluidFS Interface to the C2S API
//===========================================================================

// RunC2SAPI creates an HTTP listener for the web interface as well as the
// JSON API that is used for command and control.
func (s *Server) RunC2SAPI() error {
	// Find the address of the API and log the information
	addr := s.PID.Addr()
	s.Logger.Info("starting C2S API and web interface at http://%s/", addr)

	// Create the API and run it.
	api := new(C2SAPI)
	api.Init(s)
	return api.Run(addr)

}

//===========================================================================
// Stand-Alone C2S API and Web Server
//===========================================================================

// APIHandler is a function type that defines how C2SAPI handlers are to be
// implemented. Functions of t his type are passed to the AddHandler() method
// of the API so that they can be routed on.
type APIHandler func(r *http.Request) (int, interface{}, error)

// C2SAPI implements the web server for the command, config, and status JSON
// API that serves both a web interface and the command line client.
type C2SAPI struct {
	Router *mux.Router
	Fluid  *Server
}

// Init the C2SAPI with a hook to the server that the API wraps.
func (api *C2SAPI) Init(fluid *Server) error {
	api.Fluid = fluid
	api.Router = mux.NewRouter().StrictSlash(true)
	api.AddHandler("/status", api.StatusHandler)
	api.Router.Handle("/", WebLogger(api.Fluid.Logger, http.FileServer(assetFS())))
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
			data = make(map[string]string)
			data.(map[string]string)["code"] = strconv.Itoa(code)
			data.(map[string]string)["error"] = err.Error()
		}

		// Otherwise respond to the request
		w.Header().Set("Content-Type", "application/json;charset=UTF-8")
		w.WriteHeader(code)

		// Write the responde data as JSON to the stream
		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	handler := WebLogger(api.Fluid.Logger, outer)
	api.Router.Handle(path, handler)
}

// StatusHandler returns the status command information.
func (api *C2SAPI) StatusHandler(r *http.Request) (int, interface{}, error) {
	data := make(map[string]string)
	data["status"] = "ok"
	data["timestamp"] = time.Now().String()
	return http.StatusOK, data, nil
}
