package main

import (
	"context"
	"log"
	"os"

	// mesh_adk "github.com/ThisaraWeerakoon/Agent-Mesh/pkg/adk"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/cmd/launcher"
	"google.golang.org/adk/cmd/launcher/full"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/agenttool"
	"google.golang.org/genai"
)

func main() {
	// "summary_agent" is the ID of the agent in the AgentMesh network
	// This creates a tool that, when called, sends a task via AgentMesh.
	// summaryTool, err := mesh_adk.ImportTool("summary_agent")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	ctx := context.Background()
	geminiModel, err := gemini.NewModel(ctx, "gemini-2.5-flash-lite", &genai.ClientConfig{
		APIKey: os.Getenv("GOOGLE_API_KEY"),
	})
	if err != nil {
		log.Fatalf("Failed to create model: %v", err)
	}

	// 2. Create the Summary Agent (using the shared model)
	summaryAgent, err := llmagent.New(llmagent.Config{
		Name:        "summary_agent",
		Instruction: "You are a specialized summarizer. Your goal is to provide concise summaries of any text provided to you.",
		Model:       geminiModel, // Reuse the model here
	})
	if err != nil {
		log.Fatalf("Failed to create summary agent: %v", err)
	}

	// 3. Convert the Summary Agent into a Tool
	// This allows the root agent to call 'summary_agent' as if it were a function.
	summaryTool := agenttool.New(summaryAgent, nil)

	rootAgent, err := llmagent.New(llmagent.Config{
		Name:        "root_agent",
		Instruction: `You are a helpful assistant. Use the 'summary_agent' tool to summarize text.`,
		Tools:       []tool.Tool{summaryTool},
		Model:       geminiModel,
	})
	if err != nil {
		log.Fatal(err)
	}

	config := &launcher.Config{
		AgentLoader: agent.NewSingleLoader(rootAgent),
	}

	l := full.NewLauncher()
	if err := l.Execute(ctx, config, os.Args[1:]); err != nil {
		log.Fatalf("Run failed: %v\n\n%s", err, l.CommandLineSyntax())
	}
}
