// Implements Node and Handler methods for directories

package fluid

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"golang.org/x/net/context"

	kvdb "github.com/bbengfort/fluidfs/fluid/db"
)

//===========================================================================
// Dir Type and Constructor
//===========================================================================

// Dir implements Node and Handler interfaces for directories and container
// entities in the file system. Most importantly it references its children.
// Dir objects are currently stored in the Prefixes bucket allowing for faster
// lookup without having to do a key scan.
type Dir struct {
	Node
	Children map[string]string // References to the children of the directory
	entities map[string]Entity // Contents of the directory, pointers to Nodes
}

// Init the directory with the required properties for the directory.
func (d *Dir) Init(name string, mode os.FileMode, parent *Dir, memfs *FileSystem) {
	// Make sure the mode is a directory, then init the node.
	mode = os.ModeDir | mode
	d.Node.Init(name, mode, parent, memfs)

	// Make the entities mapping
	d.Children = make(map[string]string)
	d.entities = make(map[string]Entity)
}

//===========================================================================
// Dir Methods
//===========================================================================

// GetNode returns a pointer to the embedded Node object
func (d *Dir) GetNode() *Node {
	return &d.Node
}

// GetType returns the directory node type for two-phase lookup
func (d *Dir) GetType() *NodeType {
	return &NodeType{
		NodeDirType,
		d.FluidPath(),
	}
}

// Store the directory to persistant storage. Directories are only stored as
// meta information in the key/value embedded database. Version information
// about directories is not replicated nor tracked; directories only exist if
// they contain files, or locally if they were created by a user.
func (d *Dir) Store() error {
	// Don't do work if the meta data isn't dirty
	if !d.metadirty {
		return nil
	}

	// Get the node type
	ntype := d.GetType()

	// Marshall the directory into bytes data.
	data, err := json.Marshal(d)
	if err != nil {
		return fmt.Errorf("could not marshal directory: %s", err)
	}

	// Put the data into prefixes namespace
	if err := db.Put([]byte(ntype.Key), data, kvdb.PrefixesBucket); err != nil {
		return fmt.Errorf("could not store prefix: %s", err)
	}

	// Marshal the node type
	data, err = json.Marshal(ntype)
	if err != nil {
		return fmt.Errorf("could not marshal node type: %s", err)
	}

	// Put the node type into the global namespace
	if err := db.Put([]byte(d.FluidPath()), data, kvdb.NamesBucket); err != nil {
		return fmt.Errorf("could not store name %s: %s", d.FluidPath(), err)
	}

	// Set metadirty to false
	d.metadirty = false
	return nil
}

// Fetch the directory from persistant storage. The directory is populated by
// passing in the path to the directory from the prefix.
func (d *Dir) Fetch(key string) error {
	// Fetch the key from the prefixes bucket
	val, err := db.Get([]byte(key), kvdb.PrefixesBucket)
	if err != nil {
		return err
	}

	// Unmarshall the directory metadata into the struct
	if err := json.Unmarshal(val, &d); err != nil {
		return err
	}

	// Update the directory with current information
	d.expanded = false
	d.metadirty = false
	d.entities = make(map[string]Entity)
	return nil
}

// Expand the directory fully in memory by fetching children.
func (d *Dir) Expand() error {
	// Don't expand if already expanded
	if d.expanded {
		return nil
	}

	for name, fpath := range d.Children {
		// Only expand the child if it is not in the entities map
		if _, ok := d.entities[name]; !ok {
			if err := d.ExpandChild(name, fpath); err != nil {
				// NOTE: specific error is logged in Expandchild
				msg := "could not fully expand expand %s"
				logger.Error(msg, d)
				return err
			}
		}
	}

	// Set expanded to true and return
	d.expanded = true
	return nil
}

// ExpandChild is a directory helper function to expand one child by the full
// fluid path from the database. This function is called by Expand on a per-
// child basis, and can be called individually from the lookup function to
// save memory. ExpandChild first has to lookup the entry in the namespace
// bucket, then retrieve the appropriate entity. This two step process will
// update the entities directly, and this method will only return an error if
// for some reason the operation cannot be completed.
func (d *Dir) ExpandChild(name string, fpath string) error {
	// Fetch the entity from the database
	ent, err := FetchEntity(fpath)
	if err != nil {
		msg := "could not expand %s: %s"
		logger.Error(msg, fpath, err)
		return err
	}

	// Add the parent and filesystem to the child
	node := ent.GetNode()
	node.fs = d.fs
	node.parent = d

	// Add the entity to the entities map
	d.entities[name] = ent
	return nil
}

// Free the directory from memory usage by removing the child pointers.
func (d *Dir) Free() error {

	for name, ent := range d.entities {
		if ent.IsDir() {
			// First call free on all child directories
			if err := ent.Free(); err != nil {
				return err
			}
		}

		// Delete the entity from the mapping
		delete(d.entities, name)
	}

	// Mark expanded to false and return
	d.expanded = false
	return nil
}

//===========================================================================
// Dir fuse.Node* Interface
//===========================================================================

// Create creates and opens a new file in the receiver, which must be a Dir.
// NOTE: the interface docmentation says create a directory, but the docs
// for fuse.CreateRequest say create and open a file (not a directory).
//
// https://godoc.org/bazil.org/fuse/fs#NodeCreater
func (d *Dir) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
	if d.IsArchive() || d.fs.readonly {
		return nil, nil, fuse.EPERM
	}

	d.fs.Lock()
	defer d.fs.Unlock()

	// Update the directory Atime
	d.Attrs.Atime = time.Now()

	// Create the file
	f := new(File)
	f.Init(req.Name, req.Mode, d, d.fs)

	// Set the file's UID and GID to that of the caller
	f.Attrs.Uid = req.Header.Uid
	f.Attrs.Gid = req.Header.Gid

	// Add the file to the directory
	d.Children[f.Name] = f.FluidPath()
	d.entities[f.Name] = f

	// Update the directory Mtime and mark metadirty
	d.Attrs.Mtime = time.Now()
	d.metadirty = true

	// Update the file system state
	d.fs.nfiles++

	// Log the file creation and return the file, which is both node and handle.
	logger.Info("create %q in %q, mode %v", f.Name, d.Path(), req.Mode)
	return f, f, nil
}

// Link creates a new directory entry in the receiver based on an
// existing Node. Receiver must be a directory.
//
// A LinkRequest is a request to create a hard link and contains the old node
// ID and the NewName (a string), the old node is supplied to the server.
//
// NOTE: the 'name' is not modified, so potential debugging problem.
// NOTE: Hard links are not currently stored in the metadata storage.
// NOTE: target.Attrs.Nlink-- never occurs even on hard link removal.
//
// The ln utility creates a new directory entry (linked file) which has the
// same modes as the original file.  It is useful for maintaining multiple
// copies of a file in many places at once without using up storage for the
// "copies"; instead, a link "points" to the original copy.  There are two
// types of links; hard links and symbolic links.  How a link "points" to a
// file is one of the differences between a hard and symbolic link.
//
// By default, ln makes hard links.  A hard link to a file is
// indistinguishable from the original directory entry; any changes to a file
// are effectively independent of the name used to reference the file. Hard
// links may not normally refer to directories and may not span file systems.
//
// https://godoc.org/bazil.org/fuse/fs#NodeLinker
func (d *Dir) Link(ctx context.Context, req *fuse.LinkRequest, old fs.Node) (fs.Node, error) {
	if d.IsArchive() || d.fs.readonly {
		return nil, fuse.EPERM
	}

	d.fs.Lock()
	defer d.fs.Unlock()

	// Create the target from the reference
	target := old.(*File)
	target.Attrs.Nlink++

	// Add the target to to the parent directory and return
	d.entities[req.NewName] = target
	logger.Info("link %q to %q", filepath.Join(d.Path(), req.NewName), target.Path())
	return target, nil
}

// Mkdir creates (but not opens) a directory in the given directory.
//
// https://godoc.org/bazil.org/fuse/fs#NodeMkdirer
func (d *Dir) Mkdir(ctx context.Context, req *fuse.MkdirRequest) (fs.Node, error) {
	if d.IsArchive() || d.fs.readonly {
		return nil, fuse.EPERM
	}

	d.fs.Lock()
	defer d.fs.Unlock()

	// Update the directory Atime
	d.Attrs.Atime = time.Now()

	// TODO: Allow for the creation of archive directories

	// Create the child directory
	c := new(Dir)
	c.Init(req.Name, req.Mode, d, d.fs)

	// Set the directory's UID and GID to that of the caller
	c.Attrs.Uid = req.Header.Uid
	c.Attrs.Gid = req.Header.Gid

	// Add the directory to the directory
	d.entities[c.Name] = c
	d.Children[c.Name] = c.FluidPath()

	// Update the directory Mtime and mark metadirty
	d.Attrs.Mtime = time.Now()
	d.metadirty = true

	// Update the file system state
	d.fs.ndirs++

	// Log the directory creation and return the dir node
	logger.Info("mkdir %q in %q, mode %v", c.Name, d.Path(), req.Mode)
	return c, nil
}

// Mknode I assume creates but not opens a node and returns it.
//
// https://godoc.org/bazil.org/fuse/fs#NodeMknoder
// TODO: Implement
// func (d *Dir) Mknod(ctx context.Context, req *fuse.MknodRequest) (fs.Node, error) {
//     return nil, nil
// }

// Remove removes the entry with the given name from the receiver, which must
// be a directory.  The entry to be removed may correspond to a file (unlink)
// or to a directory (rmdir).
//
// https://godoc.org/bazil.org/fuse/fs#NodeRemover
func (d *Dir) Remove(ctx context.Context, req *fuse.RemoveRequest) error {
	if d.IsArchive() || d.fs.readonly {
		return fuse.EPERM
	}

	d.fs.Lock()
	defer d.fs.Unlock()

	// Update the directory Atime
	d.Attrs.Atime = time.Now()

	var ent Entity
	var ok bool

	// Get the node from the directory by name.
	if ent, ok = d.entities[req.Name]; !ok {
		logger.Debug("(error) could not find node to remove named %q in %q", req.Name, d.Path())
		return fuse.EEXIST
	}

	// Do not remove a directory that contains files.
	if ent.IsDir() && len(ent.(*Dir).entities) > 0 {
		logger.Debug("(error) will not remove non-empty directory %q in %q", req.Name, d.Path())
		return fuse.EIO
	}

	// Delete the entry from the directory
	delete(d.Children, req.Name)
	delete(d.entities, req.Name)

	// TODO: Delete the entry from meta storage?

	// Update the directory Mtime and mark metadata as dirty
	d.Attrs.Mtime = time.Now()
	d.metadirty = true

	// Update the file system state
	// TODO: decrement the number of links
	if ent.IsDir() {
		d.fs.ndirs--
	} else {
		d.fs.nfiles--
	}

	// Log the directory removal and return no error
	logger.Info("removed %q from %q", req.Name, d.Path())
	return nil
}

// Rename a file in a directory. NOTE: There is no documentation on this.
// Implemented to move the entry by name from the dir to the newDir.
//
// https://godoc.org/bazil.org/fuse/fs#NodeRenamer
func (d *Dir) Rename(ctx context.Context, req *fuse.RenameRequest, newDir fs.Node) error {
	if d.IsArchive() || d.fs.readonly {
		return fuse.EPERM
	}

	d.fs.Lock()
	defer d.fs.Unlock()

	// Update the directory Atime
	d.Attrs.Atime = time.Now()

	var dst *Dir
	var ok bool
	var ent Entity
	var node *Node

	// Convert newDir to an actual Dir object
	if dst, ok = newDir.(*Dir); !ok {
		logger.Debug("(error) could not convert %q to a directory", newDir)
		return fuse.EEXIST
	}

	// Update the dst directory Atime
	dst.Attrs.Atime = time.Now()

	// Get the child entity from the directory
	if ent, ok = d.entities[req.OldName]; !ok {
		logger.Debug("(error) could not find %q in %q to move", req.OldName, d.Path())
		return fuse.EEXIST
	}

	// Get the node from the entity and update attrs.
	node = ent.GetNode()
	node.Name = req.NewName
	node.Attrs.Mtime = time.Now()
	node.metadirty = true

	// Add the entity to the new directory and update.
	dst.entities[req.NewName] = ent
	dst.Children[req.NewName] = node.FluidPath()
	dst.Attrs.Mtime = time.Now()
	dst.metadirty = true

	// Delete the entity from the old directory and update.
	delete(dst.entities, req.OldName)
	delete(dst.Children, req.OldName)
	d.Attrs.Mtime = time.Now()
	d.metadirty = true

	logger.Info("moved %q from %q to %q", req.OldName, d.Path(), ent.Path())
	return nil
}

// Lookup looks up a specific entry in the receiver,
// which must be a directory.  Lookup should return a Node
// corresponding to the entry.  If the name does not exist in
// the directory, Lookup should return ENOENT.
//
// Lookup need not to handle the names "." and "..".
//
// https://godoc.org/bazil.org/fuse/fs#NodeStringLookuper
// NOTE: implemented NodeStringLookuper rather than NodeRequestLookuper
// https://godoc.org/bazil.org/fuse/fs#NodeRequestLookuper
func (d *Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {

	d.fs.Lock()
	defer d.fs.Unlock()

	// Update the directory Atime
	d.Attrs.Atime = time.Now()

	// We could expand the entire dirctory here, but instead we'll just
	// expand the child node being looked up. This will hopefully save on
	// memory for one-off lookups and will make later expansion faster.
	// First we ensure that this name is in the Children, otherwise we
	// return error; no such file or directory.
	if fpath, ok := d.Children[name]; ok {
		logger.Debug("lookup %s in %s", name, d.Path())

		// Check and see if the entity is in the entities map, if not, expand.
		if _, ok := d.entities[name]; !ok {
			if err := d.ExpandChild(name, fpath); err != nil {
				// NOTE: error is logged in expand child
				return nil, fuse.EIO
			}
		}

		// Get the entity and return the appropriate node
		ent := d.entities[name]
		if ent.IsDir() {
			return ent.(*Dir), nil
		}

		return ent.(*File), nil

	}

	logger.Debug("(error) couldn't lookup %s in %s", name, d.Path())
	return nil, fuse.ENOENT
}

// Symlink creates a new symbolic link in the receiver, which must be a directory.
// TODO is the above true about directories?
//
// https://godoc.org/bazil.org/fuse/fs#NodeSymlinker
func (d *Dir) Symlink(ctx context.Context, req *fuse.SymlinkRequest) (fs.Node, error) {
	if d.IsArchive() || d.fs.readonly {
		return nil, fuse.EPERM
	}

	d.fs.Lock()
	defer d.fs.Unlock()

	// Update the directory Atime
	d.Attrs.Atime = time.Now()

	// Create the new symlink
	ln := new(File)
	ln.Init(req.NewName, os.ModeSymlink|0777, d, d.fs)

	// Set the file's UID and GID to that of the caller
	ln.Attrs.Uid = req.Header.Uid
	ln.Attrs.Gid = req.Header.Gid

	// Add the link data to the file
	ln.data = []byte(req.Target)
	ln.dirty = true
	ln.Attrs.Size = uint64(len(ln.data))

	// Add the symlink to the directory
	d.Children[req.NewName] = ln.FluidPath()
	d.entities[req.NewName] = ln

	// Update the directory Mtime and metadirty
	d.Attrs.Mtime = time.Now()
	d.metadirty = true

	// Update the file system state
	d.fs.nfiles++

	// Log the symlink creation and return the file, which is both node and handle.
	logger.Info("create symlink from %s to %s", ln.Path(), req.Target)
	return ln, nil
}

//===========================================================================
// Dir fuse.Handle* Interface
//===========================================================================

// ReadDirAll reads the entire directory contents and returns a list of fuse
// Dirent objects - which specify the internal contents of the directory.
//
// https://godoc.org/bazil.org/fuse/fs#HandleReadDirAller
func (d *Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	contents := make([]fuse.Dirent, 0, len(d.Children))

	d.fs.Lock()
	defer d.fs.Unlock()

	// Set the access time
	d.Attrs.Atime = time.Now()

	// If not expanded, expand the directory
	if !d.expanded {
		if err := d.Expand(); err != nil {
			// NOTE: error is logged in the Expand() method.
			return nil, fuse.EIO
		}
	}

	// Create the Dirent response
	for _, entity := range d.entities {
		node := entity.GetNode()
		dirent := fuse.Dirent{
			Inode: node.Attrs.Inode,
			Type:  node.FuseType(),
			Name:  node.Name,
		}

		contents = append(contents, dirent)
	}

	logger.Debug("read all for directory %s", d.Path())
	return contents, nil
}
