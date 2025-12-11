package main
import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"github.com/ThisaraWeerakoon/Agent-Mesh/pkg/sidecar"
)
func main() {
	cfg := sidecar.Config{
		AgentID:      "agent-1",
		RegistryURL:  "localhost:50051", // Adjust to Registry gRPC port
		LocalPort:    50052,
		ExternalPort: 50053,
		AppPort:      50054, // Port where your mock agent runs
		CertFile:     "certs/server-cert.pem",
		KeyFile:      "certs/server-key.pem",
		CAFile:       "certs/ca-cert.pem",
	}
	srv, err := sidecar.NewServer(cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Handle SIGINT/SIGTERM
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()
	log.Println("Starting Sidecar...")
	if err := srv.Run(ctx); err != nil {
		log.Fatalf("Sidecar error: %v", err)
	}
}