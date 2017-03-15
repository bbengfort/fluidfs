// Implements FUSE handlers for file system in user space interaction.

package fluid

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/bbengfort/sequence"

	"bazil.org/fuse"
	fusefs "bazil.org/fuse/fs"
)

const minBlockSize = uint64(512)

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
// FileSystem Handling
//===========================================================================

// FileSystem implements the fuse.FS* interfaces.
// TODO: How to store the statistics in the database?
type FileSystem struct {
	sync.Mutex                    // A file system can be locked
	Conn       *fuse.Conn         // A connection to the FUSE server
	Sequence   *sequence.Sequence // iNode sequence object
	root       *Dir               // The root of the file system
	mount      *MountPoint        // The location and options of this mount point
	nfiles     uint64             // The number of files in the file system
	ndirs      uint64             // The number of directories in the file system
	nbytes     uint64             // The amount of data in the file system
	readonly   bool               // If the file system is readonly or not
}

// Init a file system with the replica server and the specified mount point.
func (fs *FileSystem) Init(mp *MountPoint) error {

	// Local storage of pointers to system resources
	fs.mount = mp

	// Handle the Sequence initialization
	fs.Sequence, _ = sequence.New()

	// Fetch the root node from the database
	// TODO: store more of the file system information in the database
	entity, err := FetchEntity(filepath.Join("/", fs.mount.Prefix))
	if err != nil {
		// Could not find prefix in the database -- create a new one
		fs.root = new(Dir)
		fs.root.Init("/", 0755, nil, fs)
		if err := fs.root.Store(); err != nil {
			return err
		}
	} else {
		var ok bool
		fs.root, ok = entity.(*Dir)
		if !ok {
			return fmt.Errorf("could not load root prefix %s from database", fs.mount.Prefix)
		}

		// Add the file system to the root node.
		fs.root.fs = fs
	}

	return nil
}

// Run connects to FUSE, mounts the mount point and Serves the FUSE FS.
func (fs *FileSystem) Run(echan chan<- error) {
	var err error

	// Unmount the FS in case it was mounted with errors
	fuse.Unmount(fs.mount.Path)

	// Mount the FS with the specified options.
	if fs.Conn, err = fuse.Mount(
		fs.mount.Path, fs.mount.MountOptions()...,
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

	if err := fuse.Unmount(fs.mount.Path); err != nil {
		return err
	}

	return nil
}

//===========================================================================
// FileSystem implements the fuse.FS* interfaces
//===========================================================================

// Root is called to obtain the Node for the file system root.
func (fs *FileSystem) Root() (fusefs.Node, error) {
	return fs.root, nil
}

// Destroy is called when the file system is shutting down.
//
// Linux only sends this request for block device backed (fuseblk)
// filesystems, to allow them to flush writes to disk before the
// unmount completes.
func (fs *FileSystem) Destroy() {
	logger.Info("fluidfs://%s mounted at %s is being destroyed", fs.mount.Prefix, fs.mount.Path)
}

// GenerateInode is called to pick a dynamic inode number when it
// would otherwise be 0.
//
// Not all filesystems bother tracking inodes, but FUSE requires
// the inode to be set, and fewer duplicates in general makes UNIX
// tools work better.
//
// Operations where the nodes may return 0 inodes include Getattr,
// Setattr and ReadDir.
//
// If FS does not implement FSInodeGenerator, GenerateDynamicInode
// is used.
//
// Implementing this is useful to e.g. constrain the range of
// inode values used for dynamic inodes.
func (fs *FileSystem) GenerateInode(parentInode uint64, name string) uint64 {
	// Just return the default mechanism for now.
	return fusefs.GenerateDynamicInode(parentInode, name)
}

// Statfs is called to obtain file system metadata.
// It should write that data to resp.
// func (fs *FileSystem) Statfs(ctx context.Context, req *fuse.StatfsRequest, resp *fuse.StatfsResponse) error {
//     return nil
// }
