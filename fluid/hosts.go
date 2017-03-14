// Provides support for defining and loading replica hosts on the network.

package fluid

import (
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"time"

	yaml "gopkg.in/yaml.v2"
)

//===========================================================================
// Host Type
//===========================================================================

// Hosts is a collection of replicas that can be loaded and saved to disk in
// YAML format (along with some meta comments). Hosts provides the ability to
// lookup replicas by hostname and to quickly fetch network-level information
// such as who the leaders of the various quorums are.
type Hosts struct {
	Replicas map[string]*Replica // mapping of hostname to replica
	Path     string              // path on disk where the hosts are stored
	Updated  time.Time           // timestamp of the last update to the hosts table
	names    []string            // list of hostnames (replica keys) for random selection.
}

//===========================================================================
// Hosts Serialization Methods
//===========================================================================

// Load the Replica collection from a hosts file on disk. The path will be
// stored alongside the Hosts object, so that updates can be quickly saved
// back to the host file (in case of changes). An error is returned if there
// is a problem parsing or reading the hosts file. However, no error will be
// returned if the file does not exist, and will simply be initialized empty.
func (h *Hosts) Load(path string) error {
	// Initialize the Hosts object with reasonable defaults.
	// NOTE: this will blow away any data already on the Hosts
	h.Replicas = make(map[string]*Replica)
	h.Path = path
	h.Updated = time.Now()

	// If the hosts file does not exist, return without an error.
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}

	// Read the data from the hosts file
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("could not read hosts: %s", err)
	}

	// Unmarshal the YAML data
	replicas := make([]*Replica, 0)
	if err := yaml.Unmarshal(data, &replicas); err != nil {
		return fmt.Errorf("could not unmarshal hosts: %s", err)
	}

	// Make the names slice to store keys
	h.names = make([]string, 0, len(replicas))

	// Load the Replicas into the mapping for quick lookup by Hostname.
	for _, replica := range replicas {
		// Make sure the Replica is correctly initialized
		if err := replica.Init(); err != nil {
			return err
		}

		// Add the replica to the mapping
		if err := h.Put(replica, false); err != nil {
			return err
		}
	}

	// Clear the replicas for deletion and return
	replicas = nil
	return nil
}

// Save the hosts table to the specified path on disk. If the path is empty
// ("") then this method will save the hosts to the stored Path in order to
// update the underling hosts file that was loaded.
func (h *Hosts) Save(path string) error {
	// Use the original path if none passed in
	if path == "" {
		if h.Path == "" {
			return Errors("specify a path to save the hosts file to")
		}

		path = h.Path
	}

	// Create the header comment for the hosts file
	tstmp := h.Updated.Format(time.RFC1123)
	data := []byte(fmt.Sprintf("# FluidFS replicas: last updated %s\n\n", tstmp))

	// Create the replicas list to marshal
	replicas := make([]*Replica, 0, len(h.Replicas))
	for _, replica := range h.Replicas {
		replicas = append(replicas, replica)
	}

	// Marshal the replicas slice and append to data
	rdata, err := yaml.Marshal(replicas)
	if err != nil {
		return fmt.Errorf("could not marshal hosts: %s", err)
	}
	data = append(data, rdata...)

	// Write the hosts file to disk.
	if err := ioutil.WriteFile(path, data, ModeBlob); err != nil {
		return fmt.Errorf("could not write hosts to disk: %s", err)
	}

	return nil
}

// Local returns the locally configured host by utilizing the global config
// object as well as other details in both the hosts file and system.
// This method can return a variety of errors if it cannot create or define
// the localhost or access the system in a number of important ways. This
// method will also modify the hosts file with information about the new
// local replica so that it can be added to the network on demand.
//
// NOTE: Run after Load() or it will blow away the hosts file.
func (h *Hosts) Local() (*Replica, error) {
	// If we don't have a configuration we can't do anything.
	if config == nil {
		return nil, errors.New("no local configuration has been loaded")
	}

	// If we've already loaded the local from disk, return it.
	if h.Has(config.Name) {
		return h.Get(config.Name)
	}

	// Otherwise create a default local replica with the given name.
	replica, err := LocalReplica(config.Name, "", 0, 0)
	if err != nil {
		return nil, err
	}

	// Add it to the hosts and save to disk.
	if err := h.Put(replica, true); err != nil {
		return nil, err
	}

	return replica, nil
}

//===========================================================================
// Hosts Collection Methods
//===========================================================================

// IsEmpty returns true if there are no associated replicas or hosts. The
// system uses this to assign a default network set, the current host.
func (h *Hosts) IsEmpty() bool {
	return len(h.Replicas) == 0
}

// Random returns a random replica. Useful for anti-entropy replica selection.
// If neighbor is true, then this function will not return the local host.
func (h *Hosts) Random(neighbor bool) *Replica {
	if len(h.names) < 2 {
		return nil
	}

	idx := rand.Intn(len(h.names))
	replica := h.Replicas[h.names[idx]]

	if neighbor && replica == local {
		return h.Random(neighbor)
	}

	return replica
}

// Has returns true if the replica name is part of the hosts collection.
func (h *Hosts) Has(name string) bool {
	_, ok := h.Replicas[name]
	return ok
}

// Get returns a replica by name, or an error if no replica by that name.
func (h *Hosts) Get(name string) (*Replica, error) {
	var err error

	replica, ok := h.Replicas[name]
	if !ok {
		err = fmt.Errorf("no replica named '%s'", name)
	}

	return replica, err
}

// Put a replica into the hosts list, and optionally save the updated hosts
// to disk (to ensure they're created correctly next time). Returns an error
// if the replica isn't valid or a replica with the same name or precedence
// value already exists. This method should be used to add hosts.
func (h *Hosts) Put(replica *Replica, save bool) error {
	// Validate the replica
	if err := replica.Init(); err != nil {
		return err
	}

	// Ensure a replica with the name isn't already in the hosts
	// Use replica.Update() to change information about the replica.
	if h.Has(replica.Name) {
		return fmt.Errorf("replica named '%s' already exists", replica.Name)
	}

	// Ensure a replica with the same precedence isn't already in the hosts
	for _, remote := range h.Replicas {
		if replica.Precedence == remote.Precedence {
			return fmt.Errorf("replca with precedence %d (%s) already exists", replica.Precedence, remote)
		}
	}

	// Add the replica to the hosts mapping
	h.Replicas[replica.Name] = replica

	// Add the name to the names slice
	h.names = append(h.names, replica.Name)

	// Save the hosts to disk if required
	if save {
		return h.Save("")
	}

	return nil
}

// Delete a replica by name and optionally save the updated hosts to disk.
// Does not return an error if the name doesn't exist, but also won't save.
func (h *Hosts) Delete(name string, save bool) error {
	// Don't do anything if we don't have the name.
	if !h.Has(name) {
		return nil
	}

	// Delete the replica from the map
	delete(h.Replicas, name)

	// Delete the replica name from the names slice
	h.names = Remove(name, h.names)

	// Save if needed
	if save {
		return h.Save("")
	}

	return nil
}
