package main

import (
	"bufio"
	"context"
	"fmt"
	"iter"
	"log"
	"os"
	"strings"

	mesh_adk "github.com/ThisaraWeerakoon/Agent-Mesh/pkg/adk"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"
	"google.golang.org/adk/tool"
	"google.golang.org/genai"
)

// MockModel implements a basic model that calls the tool.
type MockModel struct {
}

func (m *MockModel) Name() string {
	return "mock-model"
}

func (m *MockModel) GenerateContent(ctx context.Context, req *model.LLMRequest, stream bool) iter.Seq2[*model.LLMResponse, error] {
	return func(yield func(*model.LLMResponse, error) bool) {
		// Mock Logic:
		// 1. Check if the last message is a FunctionResponse (result from tool).
		//    If so, we should "summarize" it for the user (return text).
		// 2. Otherwise (User prompt), we should call the tool.

		var lastPart *genai.Part
		if len(req.Contents) > 0 {
			lastMsg := req.Contents[len(req.Contents)-1]
			if len(lastMsg.Parts) > 0 {
				lastPart = lastMsg.Parts[len(lastMsg.Parts)-1]
			}
		}

		if lastPart != nil && lastPart.FunctionResponse != nil {
			// Step 2: Tool has returned data. Generate final text response.
			resp := &model.LLMResponse{
				Content: &genai.Content{
					Parts: []*genai.Part{
						{
							Text: fmt.Sprintf("Here is the summary from the agent: %v", lastPart.FunctionResponse.Response),
						},
					},
				},
			}
			yield(resp, nil)
			return
		}

		// Step 1: User prompt. Call the tool.
		// Assume the tool name is "summary_agent"
		resp := &model.LLMResponse{
			Content: &genai.Content{
				Parts: []*genai.Part{
					{
						FunctionCall: &genai.FunctionCall{
							Name: "summary_agent",
							Args: map[string]interface{}{
								"text": "Please summarize this request.",
							},
						},
					},
				},
			},
		}
		yield(resp, nil)
	}
}

func main() {
	// "summary_agent" is the ID of the agent in the AgentMesh network
	// This creates a tool that, when called, sends a task via AgentMesh.
	summaryTool, err := mesh_adk.ImportTool("summary_agent")
	if err != nil {
		panic(err)
	}

	// Mock model to trigger the tool
	mockModel := &MockModel{}

	rootAgent, err := llmagent.New(llmagent.Config{
		Name:        "root_agent",
		Instruction: `You are a helpful assistant. Use the 'summary_agent' tool to summarize text.`,
		Tools:       []tool.Tool{summaryTool},
		Model:       mockModel,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Create Runner
	// We use an in-memory session service for this example.
	sessService := session.InMemoryService()
	agentRunner, err := runner.New(runner.Config{
		AppName:        "adk-client-example",
		Agent:          rootAgent,
		SessionService: sessService,
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Root Agent initialized. Type your request below (type 'exit' to quit).")
	scanner := bufio.NewScanner(os.Stdin)

	userID := "user-example"
	sessionID := "session-example"

	_, err = sessService.Create(context.Background(), &session.CreateRequest{
		AppName:   "adk-client-example",
		UserID:    userID,
		SessionID: sessionID,
	})
	if err != nil {
		log.Fatal(err)
	}

	for {
		fmt.Print(">> ")
		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "exit" {
			fmt.Println("Goodbye!")
			break
		}
		if input == "" {
			continue
		}

		msg := &genai.Content{
			Parts: []*genai.Part{
				{Text: input},
			},
		}

		// Run the agent using the runner.
		// The runner handles conversation history, tool calls, and event generation.
		for event, err := range agentRunner.Run(context.Background(), userID, sessionID, msg, agent.RunConfig{}) {
			if err != nil {
				log.Printf("Error during run: %v\n", err)
				break
			}
			// Look for model responses (Text)
			if event.Content != nil {
				for _, part := range event.Content.Parts {
					if part.Text != "" {
						fmt.Println("Agent:", part.Text)
					}
				}
			}
		}
	}
}
