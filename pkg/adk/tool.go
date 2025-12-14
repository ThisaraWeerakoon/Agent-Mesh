package adk

import (
	"fmt"
	"os"

	"github.com/ThisaraWeerakoon/Agent-Mesh/pkg/api/v1/mesh"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/structpb"
)

// ImportTool creates a new tool that proxies calls to the specified remote agent.
func ImportTool(agentID string) (tool.Tool, error) {
	port := os.Getenv("AGENTMESH_SIDECAR_PORT")
	if port == "" {
		port = "50052" // Default sidecar local port
	}

	// Handler function that executes the tool logic
	handler := func(ctx tool.Context, input map[string]any) (string, error) {
		// Connect to the local Sidecar
		conn, err := grpc.NewClient(fmt.Sprintf("127.0.0.1:%s", port), grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return "", fmt.Errorf("failed to connect to local sidecar: %w", err)
		}
		defer conn.Close()

		client := mesh.NewA2AMeshServiceClient(conn)

		// Convert input to Protobuf Struct
		inputStruct, err := structpb.NewStruct(input)
		if err != nil {
			return "", fmt.Errorf("failed to convert input to struct: %w", err)
		}

		// Prepare Metadata for Routing
		md := metadata.Pairs("x-target-skill", agentID)
		outCtx := metadata.NewOutgoingContext(ctx, md)

		// Start Stream
		stream, err := client.StreamTask(outCtx)
		if err != nil {
			return "", fmt.Errorf("failed to start stream: %w", err)
		}

		// Construct and Send TaskStart
		req := &mesh.TaskSendRequest{
			TargetAgentId: agentID,
			Message: &mesh.Message{
				Role: "user",
				Parts: []*mesh.Part{
					{
						Content: &mesh.Part_DataPart{
							DataPart: inputStruct,
						},
					},
				},
			},
		}

		err = stream.Send(&mesh.StreamEvent{
			Event: &mesh.StreamEvent_TaskStart{
				TaskStart: &mesh.TaskStart{
					Request: req,
				},
			},
		})
		if err != nil {
			return "", fmt.Errorf("failed to send task start: %w", err)
		}

		// Close send direction to indicate we are done sending
		// Note: If we wanted multi-turn, we wouldn't close strictly here, but for a tool call it's req-resp.
		stream.CloseSend()

		// Receive Response
		var outputText string
		for {
			event, err := stream.Recv()
			if err != nil {
				return "", fmt.Errorf("stream recv error: %w", err)
			}

			// Handle different event types if needed.
			// For now, looking for status updates with content?
			// Wait, the MockAgent sent TaskStatusUpdate with Message string.
			// But for real data, we usually want Artifacts or proper Messages.
			// Let's accumulate text from StatusUpdate.Message or potentially new fields.

			if status := event.GetStatusUpdate(); status != nil {
				if status.Message != "" {
					outputText += status.Message + "\n"
				}
				if status.Status == mesh.Task_COMPLETED || status.Status == mesh.Task_FAILED {
					break
				}
			}
			// Handle Artifacts?
			if art := event.GetArtifactUpdate(); art != nil {
				// Append artifact info or content?
				outputText += fmt.Sprintf("[Artifact: %s]\n", art.Artifact.Name)
			}
		}

		if outputText == "" {
			return "Task completed (no output)", nil
		}

		return outputText, nil
	}

	return functiontool.New(functiontool.Config{
		Name:        agentID,
		Description: fmt.Sprintf("Remote agent tool: %s", agentID),
	}, handler)
}
