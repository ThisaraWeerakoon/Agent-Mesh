package adk

import (
	"fmt"
	"io"
	"os"

	"github.com/ThisaraWeerakoon/Agent-Mesh/pkg/api/v1/mesh"
	"github.com/google/jsonschema-go/jsonschema"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/structpb"
)

// RemoteTool creates a new tool that proxies calls to the specified remote agent.
// It replicates the behavior of agenttool by using a default "request" string parameter.
func RemoteTool(name, description, targetSkill string) (tool.Tool, error) {
	port := os.Getenv("AGENTMESH_SIDECAR_PORT")
	if port == "" {
		port = "50052" // Default sidecar local port
	}

	// Handler function that executes the tool logic
	handler := func(ctx tool.Context, input map[string]any) (map[string]any, error) {
		// Extract 'request' from input, mirroring agenttool behavior
		req, ok := input["request"].(string)
		if !ok {
			// Fallback: try to find any string argument or convert input to string
			if len(input) > 0 {
				for _, v := range input {
					if s, ok := v.(string); ok {
						req = s
						break
					}
				}
			}
		}
		if req == "" {
			return nil, fmt.Errorf("missing string argument 'request' for tool %s", name)
		}

		// Connect to the local Sidecar
		conn, err := grpc.NewClient(fmt.Sprintf("127.0.0.1:%s", port), grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, fmt.Errorf("failed to connect to local sidecar: %w", err)
		}
		defer conn.Close()

		client := mesh.NewA2AMeshServiceClient(conn)

		// Convert input to Protobuf Struct.
		// Note: We are wrapping the original input map to preserve full context if needed,
		// but primarily sending the text 'req' as the prompt message.
		// However, the mesh protocol expects structured input (DataPart) or text (TextPart).
		// Let's iterate on this: Send 'req' as the user message text.

		inputStruct, err := structpb.NewStruct(input)
		if err != nil {
			return nil, fmt.Errorf("failed to convert input to struct: %w", err)
		}

		// Prepare Metadata for Routing
		md := metadata.Pairs("x-target-skill", targetSkill)
		outCtx := metadata.NewOutgoingContext(ctx, md)

		// Start Stream
		stream, err := client.StreamTask(outCtx)
		if err != nil {
			return nil, fmt.Errorf("failed to start stream: %w", err)
		}

		// Construct and Send TaskStart
		// We send the 'req' (instruction) as the TextPart of the message.
		reqMsg := &mesh.TaskSendRequest{
			TargetAgentId: targetSkill,
			Message: &mesh.Message{
				Role: "user",
				Parts: []*mesh.Part{
					{
						Content: &mesh.Part_TextPart{
							TextPart: req,
						},
					},
					// We also include the full raw input as DataPart just in case the remote agent needs other fields.
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
					Request: reqMsg,
				},
			},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to send task start: %w", err)
		}

		// Close send direction to indicate we are done sending
		stream.CloseSend()

		// Receive Response
		var outputText string
		for {
			event, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, fmt.Errorf("stream recv error: %w", err)
			}

			if status := event.GetStatusUpdate(); status != nil {
				if status.Message != "" {
					outputText += status.Message + "\n"
				}
				if status.Status == mesh.Task_COMPLETED || status.Status == mesh.Task_FAILED {
					break
				}
			}
			if art := event.GetArtifactUpdate(); art != nil {
				outputText += fmt.Sprintf("[Artifact: %s]\n", art.Artifact.Name)
			}
		}

		return map[string]any{"result": outputText}, nil
	}

	// Replicate agenttool's default schema: {"request": "string"}
	defaultSchema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"request": {
				Type:        "string",
				Description: "The request or instruction for the agent.",
			},
		},
		Required: []string{"request"},
	}

	return functiontool.New(functiontool.Config{
		Name:        name,
		Description: description,
		InputSchema: defaultSchema,
	}, handler)
}
