// Andler for blob replication with anti-entropy.

package fluid

import (
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"

	pb "github.com/bbengfort/fluidfs/fluid/comms"
	"golang.org/x/net/context"
)

// RunAntiEntropy starts both the server and the routine ae client.
func RunAntiEntropy(echan chan error) {
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
	pb.RegisterBlobTalkServer(server, &AEServer{})

	// Run the anti-entropy service
	go AntiEntropy(config.AntiEntropyDelay, echan)

	// Serve on the connection
	if err := server.Serve(lis); err != nil {
		echan <- err
	}

}

// AntiEntropy is a goroutine that runs in the background and will perform
// push based anti-entropy to replicate blobs across the network.
func AntiEntropy(delay int64, echan chan error) {

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
			conn, err := grpc.Dial(remote.GetAddr())
			if err != nil {
				return nil, err
			}
			defer conn.Close()

			// Create the client and send
			client := pb.NewBlobTalkClient(conn)
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
