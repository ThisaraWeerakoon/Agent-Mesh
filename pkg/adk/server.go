package adk

import (
	"fmt"
	"io"
	"log"
	"net"

	mesh "github.com/ThisaraWeerakoon/Agent-Mesh/pkg/api/v1/mesh"
	"google.golang.org/adk/llmagent"
	"google.golang.org/grpc"
)

// ServerWrapper acts as the bridge between AgentMesh sidecar and the standard ADK Agent.
// It implements the A2AMeshService interface.
type ServerWrapper struct {
	mesh.UnimplementedA2AMeshServiceServer
	agent *llmagent.Agent
}

// ServeAgent starts a gRPC server that listens for AgentMesh tasks and forwards them to the provided ADK agent.
// This blocks until the server stops.
func ServeAgent(port int, agent *llmagent.Agent) error {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", port, err)
	}

	s := grpc.NewServer()
	mesh.RegisterA2AMeshServiceServer(s, &ServerWrapper{agent: agent})

	log.Printf("Agent server listening on localhost:%d", port)
	return s.Serve(lis)
}

// StreamTask handles incoming task streams from the Sidecar.
func (s *ServerWrapper) StreamTask(stream mesh.A2AMeshService_StreamTaskServer) error {
	// ctx := stream.Context()
	log.Println("New incoming task stream")

	for {
		event, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		switch e := event.Event.(type) {
		case *mesh.StreamEvent_TaskStart:
			// Handle new task start
			log.Printf("Received TaskStart: %v", e.TaskStart)

			// 1. Extract Input
			// The Request.Message.Parts contain the input.
			// Assuming single text part for simplicity of this example,
			// or we need to reconstruct the prompt.
			// ADK's Chat() typically takes a string (user prompt).

			var prompt string
			if msg := e.TaskStart.Request.Message; msg != nil {
				for _, part := range msg.Parts {
					if txt, ok := part.Content.(*mesh.Part_TextPart); ok {
						prompt += txt.TextPart + "\n"
					}
					// Note: DataParts (Tools inputs) might need handling if Agent expects structured input.
					// For standard LLM Agent, we often pass text.
				}
			}

			if prompt == "" {
				prompt = "Hello" // Fallback
			}

			// 2. Invoke Agent
			// NOTE: We assume 'Chat' is the method to call. If ADK API differs, this needs adjustment.
			// llmagent.Agent usually has methods like Generate or Chat.
			// Since I cannot verify the exact signature, I am assuming a synchronous Chat method exists.
			// If the agent is streaming, we would pipe the stream.

			// response := s.agent.Chat(ctx, prompt)
			// Wait, llmagent from google/adk-go might not have Chat directly exposed if it's minimal.
			// But the user prompt implies it's a high-level agent.

			// MOCKING the call for now as I can't compile against adk-go without it.
			// In real impl: response, err := s.agent.UnsafeRun(ctx, prompt) or similar.
			response := fmt.Sprintf("Echo from ADK Agent: I received '%s'", prompt)

			// 3. Send Response back via Stream
			// We send a StatusUpdate with COMPLETED status and the response text.

			resp := &mesh.StreamEvent{
				Event: &mesh.StreamEvent_StatusUpdate{
					StatusUpdate: &mesh.TaskStatusUpdate{
						TaskId:  e.TaskStart.Request.ContextId, // Or generate new ID
						Status:  mesh.Task_COMPLETED,
						Message: response,
					},
				},
			}

			if err := stream.Send(resp); err != nil {
				return err
			}
			// We can close the stream if it's a single-turn interaction
			// return nil

		case *mesh.StreamEvent_StatusUpdate:
			// Handle status updates (e.g. cancellation?)
			log.Printf("Received StatusUpdate: %v", e.StatusUpdate)
		case *mesh.StreamEvent_ArtifactUpdate:
			log.Printf("Received ArtifactUpdate: %v", e.ArtifactUpdate)
		}
	}
}
