// Handler for blob replication with anti-entropy.

package fluid

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"

	pb "github.com/bbengfort/fluidfs/fluid/rpc"
	"golang.org/x/net/context"
)

//===========================================================================
// BlobAEService Type
//===========================================================================

// BlobAEService implements anti-entropy replication of blobs across the
// network, referencing the global hosts for network connectivity and the
// global config for the definition of the anti-entropy delay.
//
// The BlobAEService maintains an in-memory data structure that keeps track
// of the state of blobs in the stoage directory. This data structure is built
// when initialized, and should be updated when blobs are saved to disk. An
// update channel is used to synchronize chunkers and updates to the blob
// storage tree.
//
// Anti-entropy runs on a ticker, and after each anti-entropy delay, the
// service will select a random neighbor from the network (defined by hosts)
// to perform pairwise anti-entropy with.
type BlobAEService struct {
	Delay  time.Duration // anti-entropy delay to set the ticker
	Update chan string   // recieves file paths to update its internal directory
	ticker *time.Ticker  // channel to wait on for anti-entropy sessions
	btree  *BlobTree     // pseudo-merkle tree for blob anti-entropy
}

// RunAntiEntropy starts both the server and the routine ae client.
func RunAntiEntropy(echan chan<- error) {
	// Start the anti-entropy server
	lis, err := net.Listen("tcp", local.GetAddr())
	if err != nil {
		msg := fmt.Sprintf("could not listen on %s", local.GetAddr())
		logger.Error(msg)
		echan <- err
		return
	}

	// Create and register the blob talk server
	server := grpc.NewServer()
	pb.RegisterInterReplicaServer(server, &AEServer{})

	// Run the anti-entropy service
	go AntiEntropy(config.AntiEntropyDelay, echan)

	// Serve on the connection
	if err := server.Serve(lis); err != nil {
		echan <- err
	}

}

// AntiEntropy is a goroutine that runs in the background and will perform
// push based anti-entropy to replicate blobs across the network.
func AntiEntropy(delay int64, echan chan<- error) {

	// Create the anti-entropy ticker with the delay in milliseconds
	ticker := time.NewTicker(time.Millisecond * time.Duration(delay))

	// For every tick in the ticker, conduct pairwise anti-entropy
	for range ticker.C {

		// Select a random host from the network
		remote := hosts.Random(true)
		if remote == nil {
			continue
		}

		send := func() (*pb.PushReply, error) {
			// Create the PushRequest
			request := &pb.PushRequest{1, 4096}

			// Dial the server
			conn, err := grpc.Dial(remote.GetAddr(), grpc.WithInsecure())
			if err != nil {
				return nil, err
			}
			defer conn.Close()

			// Create the client and send
			client := pb.NewInterReplicaClient(conn)
			return client.PushHandler(context.Background(), request)
		}

		// Send the request and get the reply
		reply, err := send()
		if err != nil {
			msg := fmt.Sprintf("could not send push request: %s", err)
			logger.Error(msg)
			echan <- err
			break
		}

		logger.Info("received reply from %s: %s", remote, reply)
	}

}

// AEServer implements the server for the AntiEntropy Client
type AEServer struct{}

// PushHandler replies to a PushRequest from a client that randomly selects a host.
func (s *AEServer) PushHandler(ctx context.Context, request *pb.PushRequest) (*pb.PushReply, error) {
	return &pb.PushReply{true}, nil
}

//===========================================================================
// BlobTree Type (Pseudo Merkle Tree)
//===========================================================================

// BlobTree implements a pseudo-merkle tree. Instead of each node maintaining
// a hash of itself and the contents below it, it simply contains the count of
// the number of leaf nodes (files). BlobTrees can then be used similarly to
// Merkle trees for fast comparison of changes; though collisions can occur if
// two leaf nodes have two different files added to it. This type of tree is
// used for very fast comparison during anti-entropy and the tree should be
// maintained during the operation of FluidFS.
type BlobTree struct {
	sync.Mutex
	Parent   *BlobTree            `json:"-"`        // Link to parent node to traverse to the root
	Name     string               `json:"name"`     // Path component or base path if root node
	Count    uint64               `json:"count"`    // Number of leaf nodes under this branch
	Children map[string]*BlobTree `json:"children"` // Link to children to traverse to leaves
}

// Init a BlobTree from a root path on disk. This method walks the directory
// structure at the supplied root path and adds all files (skipping hidden
// files and directories) to the BlobTree. This is the primary way to
// construct a BlobTree, then the tree can be updated as files are added.
func (t *BlobTree) Init(root string, skipHidden bool) error {
	// Initialize the tree as the root node.
	t.Parent = nil
	t.Name = root
	t.Count = 0
	t.Children = make(map[string]*BlobTree)

	// Create the WalkFunc to walk the directory structure and build the tree.
	visit := func(path string, f os.FileInfo, err error) error {
		if f.IsDir() {
			if skipHidden {
				if IsHidden(f) {
					// If it's a hidden directory, don't walk its structure
					return filepath.SkipDir
				}

				// Otherwise continue with the walk
				return nil
			}

		}

		// If the file is hidden return without adding to the tree.
		if skipHidden && IsHidden(f) {
			return nil
		}

		// Add the file with its complete path from the root down.
		return t.AddFile(path)
	}

	// Execute the file path walk with the walk function
	return filepath.Walk(root, visit)
}

//===========================================================================
// BlobTree Manipulation Methods
//===========================================================================

// AddChild creates a new child (directory) on the current tree with the
// given name, which should be the name of the directory so that the path is
// correctly constructed from the root.
func (t *BlobTree) AddChild(name string) {
	t.Children[name] = &BlobTree{
		Parent:   t,
		Name:     name,
		Count:    0,
		Children: make(map[string]*BlobTree),
	}
}

// AddFile takes a complete path (relative to the root of the tree) and adds
// the file by creating subdirectories on the tree and incrementing the count
// of each directory on the way down. This method should only be called on the
// root node of the tree (otherwise an error is returned).
func (t *BlobTree) AddFile(path string) error {
	// Lock the entire tree during an AddFile
	// TODO: improve performance by only locking the node being modified
	t.Lock()
	defer t.Unlock()

	if !t.IsRoot() {
		return Errorc("AddFile() should be called on the root of the tree", ErrBlobStorage)
	}

	// Get the path relative to the root node.
	relpath, err := filepath.Rel(t.Name, path)
	if err != nil {
		return WrapError("could find path relative to blob tree root", ErrBlobStorage, "", err)
	}

	dname, _ := filepath.Split(relpath)
	parts := strings.Split(dname, string(filepath.Separator))
	return t.addFile(parts)
}

// Private addFile methods performs the recursion from the setup in AddFile.
// Because AddFile must be called from root, this is top to bottom traversal.
func (t *BlobTree) addFile(names []string) error {
	// Increment the files count on the node.
	t.Count++

	// Recurse on children
	if len(names) > 0 {
		// Get the name of the child
		name := names[0]

		// If the name is zero, we're done
		if name == "" {
			return nil
		}

		// Create the child directory if it doesn't exist
		if _, ok := t.Children[name]; !ok {
			t.AddChild(name)
		}

		// Recurse the addFile on the children, less the current name.
		return t.Children[name].addFile(names[1:])
	}

	// Otherwise we're done!
	return nil
}

//===========================================================================
// BlobTree Helper Methods
//===========================================================================

// IsRoot returns true if the Tree does not have a parent
func (t *BlobTree) IsRoot() bool {
	return t.Parent == nil
}

// Path returns the path of the directory by scanning up the tree.
func (t *BlobTree) Path() string {
	if t.Parent == nil {
		return t.Name
	}

	return filepath.Join(t.Parent.Path(), t.Name)
}

// String returns the path of the tree and its count.
func (t *BlobTree) String() string {
	return fmt.Sprintf("%d files under %s", t.Count, t.Path())
}
