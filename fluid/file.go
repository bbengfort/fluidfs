// Implements the File API for interacting with FluidFS files and directories.

package fluid

import (
	"os"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"golang.org/x/net/context"
)

//===========================================================================
// FluidFS Server
//===========================================================================

// Dir implements both Node and Handle for the root directory
type Dir struct{}

// Attr sets attributes on the directory
func (Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = 1
	a.Mode = os.ModeDir | 0555
	return nil
}

// Lookup reads the contents of the directory
func (Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	if name == "hello" {
		return File{}, nil
	}
	return nil, fuse.ENOENT
}

var dirDirs = []fuse.Dirent{
	{Inode: 2, Name: "hello", Type: fuse.DT_File},
}

// ReadDirAll returns all the contents of the directory
func (Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	return dirDirs, nil
}

// File implements both Node and Handle for the hello file.
type File struct{}

const greeting = "hello, world\n"

// Attr sets the attributes of a file
func (File) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = 2
	a.Mode = 0444
	a.Size = uint64(len(greeting))
	return nil
}

// ReadAll returns the data from the file
func (File) ReadAll(ctx context.Context) ([]byte, error) {
	return []byte(greeting), nil
}
