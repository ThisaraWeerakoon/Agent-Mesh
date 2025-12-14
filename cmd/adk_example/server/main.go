package main

import (
	"context"
	"log"
	"os"

	mesh_adk "github.com/ThisaraWeerakoon/Agent-Mesh/pkg/adk"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/genai"
)



func main() {

	ctx := context.Background()
	geminiModel, err := gemini.NewModel(ctx, "gemini-2.5-flash-lite", &genai.ClientConfig{
		APIKey: os.Getenv("GOOGLE_API_KEY"),
	})
	if err != nil {
		log.Fatalf("Failed to create model: %v", err)
	}

	summaryAgent, err := llmagent.New(llmagent.Config{
		Name:        "summary_agent",
		Model:       geminiModel,
		Description: "Agent to summarize text.",
		Instruction: `You are an expert summarizer.`,
	})
	if err != nil {
		log.Fatal(err)
	}

	// 2. Start the AgentMesh Server
	// This starts a gRPC server on port 50054 (or whatever your Sidecar is configured to talk to)
	log.Println("Starting Agent Server on port 50054...")
	if err := mesh_adk.ServeAgent(50054, summaryAgent); err != nil {
		log.Fatal(err)
	}
}
