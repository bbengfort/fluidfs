// Implements FUSE handlers for file system in user space interaction.

package fluid

import (
	"fmt"

	"bazil.org/fuse"
	fusefs "bazil.org/fuse/fs"
)

//===========================================================================
// Helper Functions and Initializations
//===========================================================================

// DefaultMountOptions creates a new mapping of mount options to default FUSE
// mount option values. These options can then be overrided by specific
// MountPoint configurations or whose values can be passed directly to FUSE.
func DefaultMountOptions() map[string]fuse.MountOption {
	opts := make(map[string]fuse.MountOption)
	opts["fsname"] = fuse.FSName("fluidfs")
	opts["subtype"] = fuse.Subtype("fluidfs")
	opts["local"] = fuse.LocalVolume()
	return opts
}

//===========================================================================
// FileSystem implements the fuse.FS* interfaces
//===========================================================================

// FileSystem implements the fuse.FS* interfaces.
type FileSystem struct {
	MountPoint *MountPoint // The location and options of this mount point
	Conn       *fuse.Conn  // A connection to the FUSE server
}

// Run connects to FUSE, mounts the mount point and Serves the FUSE FS.
func (fs *FileSystem) Run(echan chan error) {
	var err error

	// Unmount the FS in case it was mounted with errors
	// fuse.Unmount(fs.MountPoint.Path)

	// Mount the FS with the specified options.
	if fs.Conn, err = fuse.Mount(
		fs.MountPoint.Path, fs.MountPoint.MountOptions()...,
	); err != nil {
		echan <- fmt.Errorf("could not run FS: %s", err.Error())
		return
	}

	// Ensure that the connection is closed when done.
	defer fs.Conn.Close()

	// Serve the file system.
	if err = fusefs.Serve(fs.Conn, fs); err != nil {
		echan <- fmt.Errorf("could not run FS: %s", err.Error())
		return
	}

	// Check if the mount process has an error to report.
	<-fs.Conn.Ready
	if fs.Conn.MountError != nil {
		echan <- fmt.Errorf("could not run FS: %s", err.Error())
		return
	}
}

// Shutdown the connection to FUSE and unmount the mount point.
func (fs *FileSystem) Shutdown() error {
	if fs.Conn == nil {
		return nil
	}

	// Currently: do not close because it is deferred in the Run() method.
	// if err := fs.Conn.Close(); err != nil {
	// 	return err
	// }

	if err := fuse.Unmount(fs.MountPoint.Path); err != nil {
		return err
	}

	return nil
}

// Root returns the root directory node.
func (fs *FileSystem) Root() (fusefs.Node, error) {
	return Dir{}, nil
}
