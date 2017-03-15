// Defines replicas and communiations between each replica.

package fluid

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
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
// Replica Methods
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
