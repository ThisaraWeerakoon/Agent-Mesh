package main

import (
	"log"
	"os"

	"github.com/ThisaraWeerakoon/Agent-Mesh/internal/adapters/handler/http"
	"github.com/ThisaraWeerakoon/Agent-Mesh/internal/adapters/repository/memory"
	"github.com/ThisaraWeerakoon/Agent-Mesh/internal/core/services"
)

func main() {
	// Configuration
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	// Dependencies
	repo := memory.NewRegistryRepository()
	service := services.NewRegistryService(repo)
	handler := http.NewRegistryHandler(service)

	// Router
	router := http.SetupRouter(handler)

	// Server
	log.Printf("Starting A2A Registry Server on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
