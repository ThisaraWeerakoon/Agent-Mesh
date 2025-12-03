package main

import (
	"log"
	"os"

	"github.com/jenish2917/a2a-registry-go/internal/adapters/handler/http"
	"github.com/jenish2917/a2a-registry-go/internal/adapters/repository/memory"
	"github.com/jenish2917/a2a-registry-go/internal/core/services"
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
