package main

import (
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	grpcHandler "github.com/ThisaraWeerakoon/Agent-Mesh/internal/adapters/handler/grpc"
	"github.com/ThisaraWeerakoon/Agent-Mesh/internal/adapters/handler/http"
	"github.com/ThisaraWeerakoon/Agent-Mesh/internal/adapters/repository/memory"
	"github.com/ThisaraWeerakoon/Agent-Mesh/internal/core/services"
	pb "github.com/ThisaraWeerakoon/Agent-Mesh/pkg/api/v1/registry"
)

func main() {
	// 1. Initialize Adapters
	repo := memory.NewRegistryRepository()

	// 2. Initialize Service
	service := services.NewRegistryService(repo)

	// 3. Initialize Handlers
	httpH := http.NewRegistryHandler(service)
	grpcH := grpcHandler.NewRegistryServer(service)

	// 4. Run Servers
	go runGRPCServer(grpcH)
	runHTTPServer(httpH)
}

func runHTTPServer(handler *http.RegistryHandler) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	r := http.SetupRouter(handler)

	log.Printf("Starting A2A Registry HTTP Server on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}

func runGRPCServer(handler *grpcHandler.RegistryServer) {
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "50051"
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterRegistryServiceServer(s, handler)
	reflection.Register(s) // Enable reflection for grpcurl

	log.Printf("Starting A2A Registry gRPC Server on port %s", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
