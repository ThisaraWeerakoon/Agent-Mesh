package main

import (
	"context"
	"iter"
	"log"

	mesh_adk "github.com/ThisaraWeerakoon/Agent-Mesh/pkg/adk"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model"
	"google.golang.org/genai"
)

// MockModel implements a basic model for testing purposes since we don't have a real LLM connected.
type MockModel struct {
}

func (m *MockModel) Name() string {
	return "mock-server-model"
}

func (m *MockModel) GenerateContent(ctx context.Context, req *model.LLMRequest, stream bool) iter.Seq2[*model.LLMResponse, error] {
	return func(yield func(*model.LLMResponse, error) bool) {
		resp := &model.LLMResponse{
			Content: &genai.Content{
				Parts: []*genai.Part{
					{
						Text: "This is a summarized text from the remote agent.",
					},
				},
			},
		}
		yield(resp, nil)
	}
}

func main() {
	// 1. Create your ADK Agent as usual
	// Using a mock model for the example to be runnable without API keys
	mockModel := &MockModel{}

	summaryAgent, err := llmagent.New(llmagent.Config{
		Name:        "summary_agent",
		Model:       mockModel,
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
