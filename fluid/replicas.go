// Defines replicas and communiations between each replica.

package fluid

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Default port numbers for communication and services between replicas.
const (
	DefaultPort = 4157
)

//===========================================================================
// Replica Type
//===========================================================================

// Replica defines a host on the network and should include information about
// how to contact the Replica. Additional information can include metrics
// about two way communication from the local replica to the remote replica.
// TODO: Be smarter about times and metrics with another nested struct(s)
type Replica struct {
	Precedence uint      `yaml:"precedence"`         // the precedence of the replica for versions
	Name       string    `yaml:"name"`               // a simple hostname for the replica
	Addr       string    `yaml:"addr"`               // the IP address or domain name of the replica
	Port       uint      `yaml:"port,omitempty"`     // default port the replica is listening on
	Leader     bool      `yaml:"leader"`             // whether or not the replica is a leader
	Term       uint64    `yaml:"term,omitempty"`     // the last seen term of the replica
	Epoch      uint64    `yaml:"epoch,omitempty"`    // the last seen epoch of the replica
	Tags       []string  `yaml:"tags"`               // the tags owned by the replica
	Created    time.Time `yaml:"created"`            // timestamp the record was created locally
	Updated    time.Time `yaml:"updated"`            // timestamp the record was updated locally
	LastSeen   time.Time `yaml:"last_seen"`          // the timestamp of the last communication to the replica
	Sent       uint64    `yaml:"sent,omitempty"`     // number of messages sent to the replica
	Recv       uint64    `yaml:"recv,omitempty"`     // number of messages received from the replica
	TLSCert    string    `yaml:"tls_cert,omitempty"` // path on disk to the replica-specific TLS certificate
}

//===========================================================================
// Replica Functions
//===========================================================================

// LocalReplica creates and returns a new replica that defaults to information
// about the localhost. If name is an empty string, then the hostname is used,
// if addr is empty, then a reasonable external IP address is used and so on.
func LocalReplica(name string, addr string, port uint, precedence uint) (*Replica, error) {
	var err error

	// Perform the name check
	if name == "" {
		// Get the hostname of the localhost
		if name, err = os.Hostname(); err != nil {
			return nil, fmt.Errorf("could not discover replica name: %s", err)
		}
	}

	// Perform the addr check
	if addr == "" {
		// Get the external ip of the local host
		if addr, err = ExternalIP(); err != nil {
			return nil, err
		}
	}

	// Perform the port check
	if port == 0 {
		port = DefaultPort
	}

	// Perform the precedence check
	if precedence == 0 {
		// Assign a random integer between 1 and 1000
		precedence = uint(rand.Intn(1000)) + 1
	}

	replica := &Replica{
		Precedence: precedence,
		Name:       name,
		Addr:       addr,
		Port:       port,
	}

	if err = replica.Init(); err != nil {
		return nil, err
	}

	return replica, nil
}

// DefaultLocalReplica returns the local replica with all default values and
// no supplied values (e.g. from a configuration). Essentially a helper.
func DefaultLocalReplica() (*Replica, error) {
	return LocalReplica("", "", 0, 0)
}

//===========================================================================
// Replica Initialization
//===========================================================================

// Init the Replica with default values and return an error if the Replica is
// not configured correctly (e.g. is missing a hostname or an address).
// NOTE that replica should be non-destructive, that is it should not
// overwrite any information about the Replica already stored. This means you
// can use the Init() method to validate a Replica.
func (r *Replica) Init() error {
	prefix := "invalid replica: "

	// Validate the replica has required values
	if r.Precedence == 0 {
		return NewError("no precedence int", ErrInvalidReplica, prefix)
	}

	if r.Name == "" {
		return NewError("no name set", ErrInvalidReplica, prefix)
	}

	if r.Addr == "" {
		return NewError("no ip address or hostname", ErrInvalidReplica, prefix)
	}

	// Add reasonable defaults
	if r.Port == 0 {
		r.Port = DefaultPort
	}

	if r.Created.IsZero() {
		r.Created = time.Now()
	}

	if r.Updated.IsZero() {
		r.Updated = time.Now()
	}

	return nil
}

//===========================================================================
// Replica RPC Server Methods
//===========================================================================

// Serve creates a grpc.Server with the correct security credentials as
// specified by the configuration. Only the local replica should call serve.
//
// NOTE: This method is currently serving as a stub to show  how the replica
// can be used to create server connections.
// TODO: Make the replica a server during init and have serve listen with the
// error channel for registering complaints (create the lis connection in
// Serve and pass to the various sub serve options).
// TODO: Registration funcationality for services.
func (r *Replica) Serve() (*grpc.Server, error) {
	if config == nil || config.Security == nil {
		return nil, Errorc("the default security configuration isn't initialized", ErrUninitialized)
	}

	if config.Security.Insecure {
		return r.ServeInsecure()
	}

	if config.Security.VerifyClient {
		return r.ServeMutualTLS()
	}

	return r.ServeTLS()
}

// ServeInsecure creates a grpc.Server with no TLS credentials.
func (r *Replica) ServeInsecure() (*grpc.Server, error) {
	srv := grpc.NewServer()
	return srv, nil
}

// ServeTLS creates a grpc.Server with server-side TLS credentials.
func (r *Replica) ServeTLS() (*grpc.Server, error) {
	// Get the key and certificate from the configuration
	crt := config.Security.Cert
	key := config.Security.Key

	// Load the TLS credentials from disk.
	creds, err := credentials.NewServerTLSFromFile(crt, key)
	if err != nil {
		return nil, ParsingError("could not load TLS cert and key from disk", err)
	}

	// Create the grpc server with the credentials
	srv := grpc.NewServer(grpc.Creds(creds))
	return srv, nil
}

// ServeMutualTLS creates a grpc.Server with client verification and TLS
// credentials, both verififed with a certificate authority.
func (r *Replica) ServeMutualTLS() (*grpc.Server, error) {
	// Get the key, certificate, and CA from the configuration
	crt := config.Security.Cert
	key := config.Security.Key
	caf := config.Security.CA

	// Load the certificaters from disk
	certificate, err := tls.LoadX509KeyPair(crt, key)
	if err != nil {
		return nil, ParsingError("could not load TLS cert and key from disk", err)
	}

	// Create a certificate pool from the certificate authority
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(caf)
	if err != nil {
		return nil, ParsingError("could not read ca certificate", err)
	}

	// Append the client certificates from the ca
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		return nil, ParsingError("failed to append client certs from CA", nil)
	}

	// Create the TLS credentials
	creds := credentials.NewTLS(&tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{certificate},
		ClientCAs:    certPool,
	})

	// Create the grpc server with the credentials
	srv := grpc.NewServer(grpc.Creds(creds))
	return srv, nil
}

//===========================================================================
// Replica RPC Dial Methods
//===========================================================================

// Dial creates a grpc connection with the correct security credentials as
// specified by the configuration. All remote replicas will be dialed.
//
// NOTE: This method is currently serving as a stub to show  how the replica
// can be used to create client connections.
// TODO: Make the replica client during init and have dial connect at runtime.
// TODO: Client construction funcationality for services.
func (r *Replica) Dial() (*grpc.ClientConn, error) {
	if config == nil || config.Security == nil {
		return nil, Errorc("the default security configuration isn't initialized", ErrUninitialized)
	}

	if config.Security.Insecure {
		return r.DialInsecure()
	}

	if config.Security.VerifyClient {
		return r.DialMutualTLS()
	}

	return r.DialTLS()

}

// DialInsecure creates a client connection with no security credentials.
func (r *Replica) DialInsecure() (*grpc.ClientConn, error) {
	conn, err := grpc.Dial(r.GetAddr(), grpc.WithInsecure())
	if err != nil {
		return nil, NetworkError("could not connect to %s", err, r.GetAddr())
	}

	return conn, nil
}

// DialTLS creates a client connection with the specified server credentials
// defaulting to the credentials in the global config if no specific one is
// provided.
func (r *Replica) DialTLS() (*grpc.ClientConn, error) {

	// Find the right certificate
	var cert string
	if r.TLSCert != "" {
		cert = r.TLSCert
	} else {
		cert = config.Security.Cert
	}

	// Create the client TLS credentials
	// TODO: do we need to specify a server name override?
	creds, err := credentials.NewClientTLSFromFile(cert, "")
	if err != nil {
		return nil, ParsingError("could not load tls certifcate for %s", err, r.Name)
	}

	conn, err := grpc.Dial(r.GetAddr(), grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, NetworkError("could not connect to %s", err, r.GetAddr())
	}

	return conn, nil
}

// DialMutualTLS creates a client connection with client verification for
// mutual TLS using a certificate authority.
func (r *Replica) DialMutualTLS() (*grpc.ClientConn, error) {
	// Get the key, certificate, and CA from the configuration
	crt := config.Security.Cert
	key := config.Security.Key
	caf := config.Security.CA

	// Load the certificaters from disk
	certificate, err := tls.LoadX509KeyPair(crt, key)
	if err != nil {
		return nil, ParsingError("could not load TLS cert and key from disk", err)
	}

	// Create a certificate pool from the certificate authority
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(caf)
	if err != nil {
		return nil, ParsingError("could not read ca certificate", err)
	}

	// Append the client certificates from the ca
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		return nil, ParsingError("failed to append client certs from CA", nil)
	}

	// Create TLS credentials
	creds := credentials.NewTLS(&tls.Config{
		ServerName:   r.Addr, // NOTE: this is required!
		Certificates: []tls.Certificate{certificate},
		RootCAs:      certPool,
	})

	// Create a connection with the TLS credentials
	conn, err := grpc.Dial(r.GetAddr(), grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, NetworkError("could not connect to %s", err, r.GetAddr())
	}

	return conn, nil
}

//===========================================================================
// Replica Update Methods
//===========================================================================

// Update a replica by key/value pair - used to ensure that all modifications
// to the Replica are tracked via the updated timestamp.
func (r *Replica) Update(field string, value interface{}) error {
	var ok bool
	field = strings.ToLower(field)

	switch field {
	case "precedence":
		r.Precedence, ok = value.(uint)
	case "name":
		r.Name, ok = value.(string)
	case "addr":
		r.Addr, ok = value.(string)
	case "port":
		r.Port, ok = value.(uint)
	case "leader":
		r.Leader, ok = value.(bool)
	case "term":
		r.Term, ok = value.(uint64)
	case "epoch":
		r.Epoch, ok = value.(uint64)
	default:
		ok = false
	}

	if !ok {
		return fmt.Errorf("could not update field %s to %v", field, value)
	}

	r.Updated = time.Now()
	return nil
}

// UpdateLastSeen allows error-free setting of the last seen timestamp.
func (r *Replica) UpdateLastSeen(ts time.Time) {
	if ts.IsZero() {
		ts = time.Now()
	}

	r.LastSeen = ts
	r.Updated = time.Now()
}

// UpdateSent allows error-free incrementing of the sent messages count.
func (r *Replica) UpdateSent(amt uint64) {
	r.Sent += amt
	r.Updated = time.Now()
}

// UpdateRecv allows error-free incrementing of the recv messages count.
func (r *Replica) UpdateRecv(amt uint64) {
	r.Recv += amt
	r.Updated = time.Now()
}

// GetAddr returns a complete host/addr with the port number.
func (r *Replica) GetAddr() string {
	return fmt.Sprintf("%s:%d", r.Addr, r.Port)
}

// String returns the host/address representation of the replica.
func (r *Replica) String() string {
	return fmt.Sprintf("%s@%s:%d", r.Name, r.Addr, r.Port)
}
