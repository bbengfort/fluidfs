// The configuration file /etc/fstab contains the necessary information to
// automate the process of mounting partitions. In a nutshell, mounting is the
// process where a raw (physical) partition is prepared for access and
// assigned a location on the file system tree (or mount point).

package fluid

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// The regular expression to match an update line
const (
	fstabUpdateLine = `^# FluidFS fstab config last updated: ([\w\d\s\-\+:,]+)$`
	fstabUpdateFmt  = "# FluidFS fstab config last updated: %s\n"
	fstabUpdateDate = "Monday, 02 Jan 2006 at 15:04:05 -0700"
)

//===========================================================================
// FS Table Structs and Interfaces
//===========================================================================

// MountPoint describes a location on disk to mount the FUSE file system and
// to watch for file system operations. The MountPoint options determine how
// to replicate and store data about files, and every FluidFS store can have
// one or more mount points, with multiple users and groups.
//
// MountPoint objects are not created directly but are rather read from lines
// in an fstab file, described by the FSTable object, which maintains and
// updates the fstab file on disk. In a nutshell, mounting is the process
// where a raw (physical) partition is prepared for access and assigned a
// location on the file system tree (or mount point).
type MountPoint struct {
	UUID      uuid.UUID // A device-unique ID for the mount point across replicas
	Path      string    // The path to the location on disk to mount
	Prefix    string    // The bucket or prefix for all names at this mount point
	UID       int       // The default user (local fs) of the mount point
	GID       int       // The default group (local fs) of the mount point
	Store     bool      // Whether or not to store blobs for this mount point
	Replicate bool      // Whether or not to replicate objects in this mount point
	Comments  []string  // Comments in the fstab file preceeding the mount point definition
	Options   []string  // Generic comma separated options for any online tweaking
}

// FSTable maintins information about all MountPoints in the system and is
// managed by the FluidFS daemon through interactions with the fluid command
// line program. The FSTable object manages (reads and updates) an fstab file
// in the configuration directory that should not be modified by the user.
// Note that this is inspired by /etc/fstab in Linux and closely tied to it.
//
// Because fstab stands for File System Table, we also use it in FluidFS to
// describe the mount points for FUSE. Each mount point requires a global
// prefix since the global root, `/` is protected (similar to a bucket in S3).
// The current plan is for the mount points to be uniquely identified via a
// UUID based on the IP address so that they can be identified across
// replicas. Additionally, the mount point should identify user and group ids
// relative to the local file system.
//
// This file is not part of the configuration system because it will need to
// be automatically modified by the system. It resides in it's own separate
// file in the configuration directory. The current plan for the file format
// is as follows:
//
//     [UUID] [Mount Point] [Prefix] [UID] [GID] [Options] [Store] [Replicate]
//
// This format will be parsed and loaded into MountPoint objects. Any lines
// starting with a '#' will be ignored as comments. All whitespace between
// options will be split including tabs, spaces, and multiple spaces.
type FSTable struct {
	Mounts  []*MountPoint // A list of the mount points in the fstab
	Path    string        // The path on disk where the fstab is stored
	Updated time.Time     // The timestamp of the last update to the fstab
}

//===========================================================================
// FSTable Methods
//===========================================================================

// Load an FSTable object from the specified path. The path will be stored
// along with the FSTable object, so this is the primary entry point to the
// creation of an FSTable. An error will be returned if there is a problem
// parsing or reading the file, however no error will be returned if the fstab
// file does not exist, and instead the FSTable will be initialized with no
// mount points (so that it can be saved later during the run time).
func (fstab *FSTable) Load(path string) error {

	// Initialize the fstab object with reasonable defaults.
	// NOTE: this will blow away any data already on the FSTable.
	fstab.Mounts = make([]*MountPoint, 0)
	fstab.Path = path
	fstab.Updated = time.Now()

	// Store comments before each mount point definition
	comments := make([]string, 0)

	// Be able to parse comment update lines
	updateLine := regexp.MustCompile(fstabUpdateLine)

	// If the fstab file does not exist, return without an error.
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}

	// Open the fstab file for reading
	fobj, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("could not open the fstab for reading: %s", err.Error())
	}

	// Ensure the file will be closed at the end
	defer fobj.Close()

	// Create a line scanner to read the fstab file line by line.
	scanner := bufio.NewScanner(fobj)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		// Continue if the line is empty
		if line == "" {
			continue
		}

		// Check if the line is a comment
		if strings.HasPrefix(line, "#") {

			// If it is a comment, check if it's an update line.
			if updateLine.MatchString(line) {

				// We have an update string!
				sub := updateLine.FindStringSubmatch(line)

				// Parse the date if possible
				date, err := time.Parse(fstabUpdateDate, strings.TrimSpace(sub[1]))
				if err != nil {
					return fmt.Errorf("could not parse update line: %s", err.Error())
				}

				// Set the updated time stamp.
				fstab.Updated = date

			} else {

				// Add the comments to the MountPoint comments
				comments = append(comments, line)

			}
		} else {

			// Otherwise, this line is a mount point definition (or better be).
			mp := new(MountPoint)
			if err := mp.Parse(line); err != nil {
				return err
			}

			// Add the comments to the mount point and reset the comments
			mp.Comments = comments
			comments = make([]string, 0)

			// Append the mount point to the mount points
			fstab.Mounts = append(fstab.Mounts, mp)
		}

	}

	// If there is a scanning error, return it
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("could not open the fstab for reading: %s", err.Error())
	}

	return nil
}

// Save an FSTable to the specified path on disk. If path is empty ("") then
// this method will save the fstab to the stored Path in order to update the
// underlying fstab file that was loaded.
func (fstab *FSTable) Save(path string) error {
	// Get the Path if it doesn't exist
	if path == "" {
		path = fstab.Path
	}

	// Update the updated timestamp
	fstab.Updated = time.Now()

	output := fmt.Sprintf(fstabUpdateFmt, fstab.Updated.Format(fstabUpdateDate))

	for _, mp := range fstab.Mounts {
		output += fmt.Sprintf("\n%s\n%s\n", strings.Join(mp.Comments, "\n"), mp.String())
	}

	return ioutil.WriteFile(path, []byte(output), ModeBlob)
}

// AddMountPoint attempts to add, save, and mount a MountPoint to the file
// system table. This may require a consensus decision or other communication
// and may return an error. All MountPoints shoudl be added through this
// method.
func (fstab *FSTable) AddMountPoint(mp *MountPoint) error {
	// Verify MountPoint uniqueness
	for _, cmp := range fstab.Mounts {
		if mp.UUID == cmp.UUID {
			return fmt.Errorf("mount point with uuid '%s' already exists", cmp.UUID.String())
		}

		if mp.Path == cmp.Path {
			return fmt.Errorf("mount point with path '%s' already exists", cmp.Path)
		}

		if mp.Prefix == cmp.Prefix {
			return fmt.Errorf("mount point with prefix '%s' already exists", cmp.Prefix)
		}
	}

	// Append the mount point to the mounts list.
	fstab.Mounts = append(fstab.Mounts, mp)

	// Save the fstab file to disk
	if err := fstab.Save(""); err != nil {
		return err
	}

	return nil
}

// Status returns a string that updates the user about the current status of
// FSTable mount points, indicating number and health.
func (fstab *FSTable) Status() string {
	return fmt.Sprintf("fs has %d mount points", len(fstab.Mounts))
}

//===========================================================================
// MountPoint Methods
//===========================================================================

// Parse a mount point definition line from an fstab file. Currently the line
// format respects white space delimiters (tab, space, and multi-space):
//
//     [UUID] [Mount Point] [Prefix] [UID] [GID] [Options] [Store] [Replicate]
//
// This populates the mount point object and will overwrite any data already
// stored in the object. The FSTable object passes lines from the fstab file
// to mount point objects in order to instantiate them.
func (mp *MountPoint) Parse(line string) error {
	var err error
	fields := strings.Fields(line)

	if len(fields) != 8 {
		return errors.New("could not parse mount point: not enough fields")
	}

	// Parse the UUID
	if mp.UUID, err = uuid.Parse(fields[0]); err != nil {
		return fmt.Errorf("could not parse UUID field: %s", err.Error())
	}

	// Set the mount point path and prefix strings
	mp.Path = fields[1]
	mp.Prefix = fields[2]

	// Parse the UID and GID integers
	if mp.UID, err = strconv.Atoi(fields[3]); err != nil {
		return fmt.Errorf("could not parse UID field: %s", err.Error())
	}

	if mp.GID, err = strconv.Atoi(fields[4]); err != nil {
		return fmt.Errorf("could not parse GID field: %s", err.Error())
	}

	// Split the options on comma and store.
	mp.Options = strings.Split(fields[5], ",")

	// Parse the Store and Replicate Boolean values
	if mp.Store, err = strconv.ParseBool(fields[6]); err != nil {
		return fmt.Errorf("could not parse Store field: %s", err.Error())
	}

	if mp.Replicate, err = strconv.ParseBool(fields[7]); err != nil {
		return fmt.Errorf("could not parse Replicate field: %s", err.Error())
	}

	return nil
}

// String returns a string representation of the MountPoint as defined by the
// line definition for the fstab file. It is used to write the mount point to
// the fstab file when the FSTable is saved.
func (mp *MountPoint) String() string {
	if mp.Options == nil || len(mp.Options) == 0 {
		mp.Options = []string{"defaults"}
	}

	fields := []string{
		mp.UUID.String(),
		mp.Path,
		mp.Prefix,
		strconv.Itoa(mp.UID),
		strconv.Itoa(mp.GID),
		strings.Join(mp.Options, ","),
		strconv.FormatBool(mp.Store),
		strconv.FormatBool(mp.Replicate),
	}

	return strings.Join(fields, " ")
}
