// Implements the File API for interacting with FluidFS files and directories.

package fluid

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"bazil.org/fuse"
	"golang.org/x/net/context"

	kvdb "github.com/bbengfort/fluidfs/fluid/db"
)

//===========================================================================
// Files
//===========================================================================

// File implements Node and Handler interfaces for file (data containing)
// objects in MemFs. Data is allocated directly in the file object, and is
// not chunked or broken up until transport.
type File struct {
	Node
	Version  *Version // The lamport scalar version of the file
	Previous *Version // The version that directly preceeded this file
	Blobs    []string // The blobs that make up the file
	data     []byte   // Actual data contained by the File
	dirty    bool     // If data has been written but not flushed
}

// Init the file and create the data array
func (f *File) Init(name string, mode os.FileMode, parent *Dir, memfs *FileSystem) {
	// Init the embedded node.
	f.Node.Init(name, mode, parent, memfs)

	// If the file was created without a fetch, set the Version.
	// NOTE: file is initialized after fetch do not overwrite Version!
	if f.Version == nil {
		// Set the version as root with no previous version.
		// NOTE: the version won't be updated until flush.
		f.Version = NewVersion()
		f.Previous = nil
	}

	// Make the data arrays
	f.Blobs = make([]string, 0, 0)
	f.data = make([]byte, 0, 0)
	f.dirty = false
}

//===========================================================================
// File Methods
//===========================================================================

// GetNode returns a pointer to the embedded Node object
func (f *File) GetNode() *Node {
	return &f.Node
}

// GetType returns the file node type for two-phase lookup
func (f *File) GetType() *NodeType {
	if f.Version.IsRoot() {
		return &NodeType{
			NodeRootType,
			fmt.Sprintf("(%s, ROOT)", f.FluidPath()),
		}
	}

	return &NodeType{
		NodeFileType,
		fmt.Sprintf("(%s, %d, %d)", f.FluidPath(), f.Version.Scalar, f.Version.PID),
	}
}

// Store a file in persistant storage. Files are stored in two places; first,
// the version meta data of the file is stored in a key/value embedded
// database and second, the data that comprises the file is stored in chunked
// blobs stored on disk. The store method saves both meta data and blobs.
func (f *File) Store() error {

	// NOTE: store blobs comes first because it changes the meta data.
	if err := f.StoreBlobs(); err != nil {
		return err
	}

	// NOTE: store meta comes after because it changes the meta data.
	if err := f.StoreMeta(); err != nil {
		return err
	}

	return nil
}

// StoreMeta is a helper function to specifically store the metadata of the
// file. It is called from the Store method, but does not handle blobs.
func (f *File) StoreMeta() error {
	// Don't waste time with the database if the metadata isn't dirty.
	if !f.metadirty {
		return nil
	}

	// Get the node type
	ntype := f.GetType()

	// Marshall the file into bytes data.
	data, err := json.Marshal(f)
	if err != nil {
		return fmt.Errorf("could not marshal file: %s", err)
	}

	// Put the data into versions namespace
	if err := db.Put([]byte(ntype.Key), data, kvdb.VersionsBucket); err != nil {
		return fmt.Errorf("could not store version: %s", err)
	}

	// Marshal the node type
	data, err = json.Marshal(ntype)
	if err != nil {
		return fmt.Errorf("could not marshal node type: %s", err)
	}

	// Put the node type into the global namespace
	if err := db.Put([]byte(f.FluidPath()), data, kvdb.NamesBucket); err != nil {
		return fmt.Errorf("could not store name %s: %s", f.FluidPath(), err)
	}

	// Set the metadirty to false
	f.metadirty = false
	logger.Debug("stored metadata for %s", f.FluidPath())
	return nil
}

// StoreBlobs is a helper function to specifically store the blobs that make
// up the complete file. It is called by the Store method but does not handle
// metadata and can be called independently to force data to disk.
func (f *File) StoreBlobs() error {
	//  Don't waste time with chunking if the file isn't dirty
	if !f.dirty {
		return nil
	}

	// Create the chunker for the given data
	chunker, err := NewDefaultChunker(f.data)
	if err != nil {
		return err
	}

	// Clear out the current blobs and make a new blob array.
	f.Blobs = make([]string, 0, 0)

	// Perform the chunking and store the blobs
	for chunker.Next() {
		blob := chunker.Chunk()                // get the next chunk of the file
		f.Blobs = append(f.Blobs, blob.Hash()) // store the identity of the hash
		err := blob.Save("")                   // store the blob on disk
		if err != nil {
			return err
		}
	}

	// Set the metadirty flag to true as we've modified Blobs.
	f.metadirty = true

	// Set the dirty flag to false
	f.dirty = false
	logger.Debug("stored %d blobs on disk", len(f.Blobs))
	return nil
}

// Fetch the file from persistant storage. Files are fetched by retrieving
// meta data from the embedded database by version. Historical version
// information can be passed in order to fetch file archives. Note taht the
// block data making up the file isn't fetched from disk until read.
func (f *File) Fetch(key string) error {
	// Fetch the key from the prefixes bucket
	val, err := db.Get([]byte(key), kvdb.VersionsBucket)
	if err != nil {
		return err
	}

	// Unmarshall the directory metadata into the struct
	if err := json.Unmarshal(val, &f); err != nil {
		return err
	}

	// Set expanded to false
	f.expanded = false
	logger.Info("fetched node %d: %s version %s -> %s", f.ID, f.Name, f.Previous, f.Version)
	return nil
}

// Expand the file fully in memory by loading the data blocks from disk.
func (f *File) Expand() error {
	// If already expanded don't worry about doing extra work.
	if f.expanded {
		return nil
	}

	// Load blobs from disk based on their hash signature
	// TODO: if blob is not on disk, request it from a remote node.
	for _, hash := range f.Blobs {
		blob, err := FindBlob(hash, "")
		if err != nil {
			return err
		}

		f.data = append(f.data, blob.Data()...)
	}

	f.expanded = true
	return nil
}

// Free the file from memory usage by emptying the data array.
func (f *File) Free() error {
	f.data = make([]byte, 0, 0)
	f.expanded = false
	return nil
}

//===========================================================================
// File fuse.Node* Interface
//===========================================================================

// Setattr sets the standard metadata for the receiver.
//
// Note, this is also used to communicate changes in the size of
// the file, outside of Writes.
//
// req.Valid is a bitmask of what fields are actually being set.
// For example, the method should not change the mode of the file
// unless req.Valid.Mode() is true.
//
// https://godoc.org/bazil.org/fuse/fs#NodeSetattrer
func (f *File) Setattr(ctx context.Context, req *fuse.SetattrRequest, resp *fuse.SetattrResponse) error {
	if f.IsArchive() || f.fs.readonly {
		return fuse.EPERM
	}

	// If size is set, this represents a truncation for a file (for a dir?)
	if req.Valid.Size() {
		f.fs.Lock() // Only lock if we're going to change the size.

		f.Attrs.Size = req.Size
		f.Attrs.Blocks = Blocks(f.Attrs.Size)
		f.data = f.data[:req.Size] // If size > len(f.data) then panic!
		logger.Debug("truncate size from %d to %d on file %d", f.Attrs.Size, req.Size, f.ID)

		f.fs.Unlock() // Must unlock before Node.Setattr is called!
	}

	// Now use the embedded Node's Setattr method.
	return f.Node.Setattr(ctx, req, resp)
}

// Fsync must be defined or edting with vim or emacs fails.
// Implements NodeFsyncer, which has no associated documentation.
//
// https://godoc.org/bazil.org/fuse/fs#NodeFsyncer
func (f *File) Fsync(ctx context.Context, req *fuse.FsyncRequest) error {
	f.fs.Lock()
	defer f.fs.Unlock()

	logger.Debug("fsync on file %d", f.ID)
	return nil
}

//===========================================================================
// File fuse.Handle* Interface
//===========================================================================

// Flush is called each time the file or directory is closed. Because there
// can be multiple file descriptors referring to a single opened file, Flush
// can be called multiple times.
//
// Because this is an in-memory system, Flush is basically ignored.
//
// https://godoc.org/bazil.org/fuse/fs#HandleFlusher
func (f *File) Flush(ctx context.Context, req *fuse.FlushRequest) error {
	logger.Info("flush file %d (dirty: %t, contains %d bytes with size %d)", f.ID, f.dirty, len(f.data), f.Attrs.Size)

	if f.IsArchive() || f.fs.readonly {
		return fuse.EPERM
	}

	f.fs.Lock()
	defer f.fs.Unlock()

	if !f.dirty {
		return nil
	}

	f.Attrs.Atime = time.Now()
	f.Attrs.Mtime = f.Attrs.Atime

	// Update the version information on the file
	// TODO: what if there is a name conflict at root?
	logger.Warn("before version bump node %d: %s is %s -> %s", f.ID, f.Name, f.Previous, f.Version)
	f.Previous = f.Version
	f.Version = f.Version.Next(local.Precedence)
	logger.Warn("updated %s version from %s to %s", f.Name, f.Previous, f.Version)

	// NOTE: Store both the data and the meta data to disk.
	if err := f.Store(); err != nil {
		msg := "could not write blobs to disk for %s: %s"
		logger.Error(msg, f, err)
		return fuse.EIO
	}

	return nil
}

// ReadAll the data from a file. Implements HandleReadAller which has no
// associated documentation.
//
// Note that if ReadAll is implemented it supersedes Read() and should only
// be implemented as a convenience for applications that do not need to do
// offset reads.
//
// https://godoc.org/bazil.org/fuse/fs#HandleReadAller
// NOTE: Do not implement
// func (f *File) ReadAll(ctx context.Context) ([]byte, error) {
// 	f.fs.Lock()
// 	defer f.fs.Unlock()
//
// 	// Set the access time on the file.
// 	f.Attrs.Atime = time.Now()
//
// 	// Return the data with no error.
// 	logger.Debug("read all file %d", f.ID)
// 	return f.data, nil
// }

// Read requests to read data from the handle.
//
// There is a page cache in the kernel that normally submits only page-aligned
// reads spanning one or more pages. However, you should not rely on this. To
// see individual requests as submitted by the file system clients, set
// OpenDirectIO.
//
// NOTE: that reads beyond the size of the file as reported by Attr are not
// even attempted (except in OpenDirectIO mode).
//
// https://godoc.org/bazil.org/fuse/fs#HandleReader
func (f *File) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	f.fs.Lock()
	defer f.fs.Unlock()

	if !f.expanded {
		if err := f.Expand(); err != nil {
			// NOTE: error logged in the Expand method
			return fuse.EIO
		}
	}

	// Find the end of the data slice to return.
	to := uint64(req.Offset) + uint64(req.Size)
	if to > f.Attrs.Size {
		to = f.Attrs.Size
	}

	// Set the access time on the file.
	f.Attrs.Atime = time.Now()

	// Set the data on the response object.
	resp.Data = f.data[req.Offset:to]

	logger.Debug("read %d bytes from offset %d in file %d", req.Size, req.Offset, f.ID)
	return nil
}

// Readlink reads a symbolic link.
//
// Symbolic links are currently implemented as files whose data is a pointer
// to the linked object, hence why the Readlink function is implemented here.
//
// https://godoc.org/bazil.org/fuse/fs#NodeReadlinker
func (f *File) Readlink(ctx context.Context, req *fuse.ReadlinkRequest) (string, error) {
	f.fs.Lock()
	defer f.fs.Unlock()

	if !f.expanded {
		if err := f.Expand(); err != nil {
			// NOTE: error logged in the Expand method
			return "", fuse.EIO
		}
	}

	ln := string(f.data)
	logger.Debug("read link from %q to %q", f.Path(), ln)
	return ln, nil
}

// Release the handle to the file. No associated documentation.
//
// https://godoc.org/bazil.org/fuse/fs#HandleReleaser
// func (f *File) Release(ctx context.Context, req *fuse.ReleaseRequest) error {
// 	logger.Debug("release handle on file %d", f.ID)
// 	return nil
// }

// Write requests to write data into the handle at the given offset.
// Store the amount of data written in resp.Size.
//
// There is a writeback page cache in the kernel that normally submits only
// page-aligned writes spanning one or more pages. However, you should not
// rely on this. To see individual requests as submitted by the file system
// clients, set OpenDirectIO.
//
// Writes that grow the file are expected to update the file size (as seen
// through Attr). Note that file size changes are communicated also through
// Setattr.
//
// https://godoc.org/bazil.org/fuse/fs#HandleWriter
func (f *File) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	if f.IsArchive() || f.fs.readonly {
		return fuse.EPERM
	}

	f.fs.Lock()
	defer f.fs.Unlock()

	if !f.expanded {
		if err := f.Expand(); err != nil {
			// NOTE: error logged in the Expand method
			return fuse.EIO
		}
	}

	olen := uint64(len(f.data))   // original data length
	wlen := uint64(len(req.Data)) // data write length
	off := uint64(req.Offset)     // offset of the write
	lim := off + wlen             // The final length of the data

	// Ensure the original size is the same as the set size (debugging)
	if olen != f.Attrs.Size {
		msg := "bad size match: %d vs %d"
		logger.Error(msg, olen, f.Attrs.Size)
	}

	// If the amount of data being written is greater than the amount of data
	// currently being stored, allocate a new array with sufficient size and
	// copy the original data to that buffer.
	if lim > olen {
		buf := make([]byte, lim)

		var to uint64
		if off < olen {
			to = off
		} else {
			to = olen
		}

		copy(buf[0:to], f.data[0:to])
		f.data = buf

		// Update the size attributes of the file
		f.Attrs.Size = lim
		f.Attrs.Blocks = Blocks(f.Attrs.Size)

		// Update the file system state
		f.fs.nbytes += lim - olen
	}

	// Copy the data from the request into our data buffer
	copy(f.data[off:lim], req.Data[:])

	// Set the attributes on the response
	resp.Size = int(wlen)

	// Mark the file as dirty
	f.dirty = true

	logger.Debug("wrote %d bytes offset by %d to file %d", wlen, off, f.ID)
	return nil
}
