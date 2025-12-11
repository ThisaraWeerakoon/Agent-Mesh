package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	mesh "github.com/ThisaraWeerakoon/Agent-Mesh/pkg/api/v1/mesh"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// MockAgentServer implements the A2AMeshServiceServer interface.
type MockAgentServer struct {
	mesh.UnimplementedA2AMeshServiceServer
}

// StreamTask handles incoming tasks from the Sidecar.
func (s *MockAgentServer) StreamTask(stream mesh.A2AMeshService_StreamTaskServer) error {
	ctx := stream.Context()
	md, _ := metadata.FromIncomingContext(ctx)
	log.Printf("Mock Agent received connection. Metadata: %v", md)

	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Printf("Error receiving message: %v", err)
			return err
		}

		log.Printf("Mock Agent received message: %v", msg)

		// Echo back a response
		response := &mesh.StreamEvent{
			Event: &mesh.StreamEvent_StatusUpdate{
				StatusUpdate: &mesh.TaskStatusUpdate{
					TaskId:  "mock-task-id",
					Status:  mesh.Task_WORKING,
					Message: "Mock Agent received your message!",
				},
			},
		}

		if err := stream.Send(response); err != nil {
			log.Printf("Error sending response: %v", err)
			return err
		}
	}
}

func main() {
	port := 50054
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		log.Fatalf("Failed to listen on port %d: %v", port, err)
	}

	grpcServer := grpc.NewServer()
	mesh.RegisterA2AMeshServiceServer(grpcServer, &MockAgentServer{})

	log.Printf("Mock Agent listening on localhost:%d", port)

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("Shutting down Mock Agent...")
		grpcServer.GracefulStop()
	}()

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
