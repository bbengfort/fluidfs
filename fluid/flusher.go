// Handler for metadata flushing

package fluid

import (
	"fmt"
	"time"
)

// Flusher is a goroutine that runs in the background and will flush metadata
// to disk if its dirty, thus preventing interruptions to the file system.
func Flusher(delay int64, echan chan error) {

	// Create the ticker with the delay in millisecond
	ticker := time.NewTicker(time.Millisecond * time.Duration(delay))

	// For every tick in the ticker flush all file systems
	for range ticker.C {

		for _, fs := range fstab.FuseFS {
			err := fs.Flush()
			if err != nil {
				err = fmt.Errorf("error flushing filesystem at %s: %s", fs.root.FluidPath(), err)
				echan <- err
				return
			}
		}

	}
}

// Flush the meta data of a file system by traversing the directory structure
// and looking for metadirty entities.
func (fs *FileSystem) Flush() error {
	fs.Lock()
	defer fs.Unlock()

	var flushes uint64

	// Store the metadata associated with each entity
	err := fs.root.Traverse(func(e Entity) error {
		if e.IsDir() {
			d := e.(*Dir)
			if d.metadirty {
				flushes++
			}
		} else {
			f := e.(*File)
			if f.metadirty {
				flushes++
			}
		}

		return e.Store()
	})

	if err == nil && flushes > 0 {
		logger.Info("flushed %d entities at %s", flushes, fs.root.FluidPath())
	}

	return err
}

// Traverse applies the entity mapper to all the children of the directory and
// returns any errors it encounters along the way. This will implement a
// depth-first traversal and application of the entity mapper.
func (d *Dir) Traverse(em EntityMapper) error {
	// First apply the entity mapper to the children
	for _, child := range d.entities {
		err := child.Traverse(em)
		if err != nil {
			return err
		}
	}

	// Then apply the entity mapper to self
	return em(d)
}

// Traverse applies the entity mapper to the file and returns any errors that
// occur when applying the function.
func (f *File) Traverse(em EntityMapper) error {
	return em(f)
}
