package sidecar

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"

	mesh "github.com/ThisaraWeerakoon/Agent-Mesh/pkg/api/v1/mesh"
	registry "github.com/ThisaraWeerakoon/Agent-Mesh/pkg/api/v1/registry"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Server implements the Sidecar Proxy.
type Server struct {
	mesh.UnimplementedA2AMeshServiceServer
	config         Config
	registryClient registry.RegistryServiceClient
}

// NewServer creates a new Sidecar Server.
func NewServer(cfg Config) (*Server, error) {
	// Connect to Registry
	// Note: Assuming insecure for registry connection for now as per prompt instructions focusing on sidecar-sidecar mTLS.
	// In a real scenario, this might also use TLS.
	conn, err := grpc.Dial(cfg.RegistryURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to registry: %w", err)
	}

	return &Server{
		config:         cfg,
		registryClient: registry.NewRegistryServiceClient(conn),
	}, nil
}

// Run starts the two concurrent listeners.
func (s *Server) Run(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)

	// 1. Local Listener (Plaintext, localhost)
	g.Go(func() error {
		lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", s.config.LocalPort))
		if err != nil {
			return fmt.Errorf("failed to listen on local port: %w", err)
		}
		log.Printf("Local listener started on localhost:%d", s.config.LocalPort)

		grpcServer := grpc.NewServer()
		mesh.RegisterA2AMeshServiceServer(grpcServer, s)

		// Graceful shutdown on context cancellation
		go func() {
			<-ctx.Done()
			grpcServer.GracefulStop()
		}()

		if err := grpcServer.Serve(lis); err != nil {
			return fmt.Errorf("local server failed: %w", err)
		}
		return nil
	})

	// 2. External Listener (mTLS, 0.0.0.0)
	g.Go(func() error {
		creds, err := loadTLSCredentials(s.config.CAFile, s.config.CertFile, s.config.KeyFile)
		if err != nil {
			return fmt.Errorf("failed to load TLS credentials: %w", err)
		}

		lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", s.config.ExternalPort))
		if err != nil {
			return fmt.Errorf("failed to listen on external port: %w", err)
		}
		log.Printf("External listener started on 0.0.0.0:%d", s.config.ExternalPort)

		grpcServer := grpc.NewServer(
			grpc.Creds(creds),
			grpc.UnaryInterceptor(logPeerIdentityInterceptor),
			grpc.StreamInterceptor(streamLogPeerIdentityInterceptor),
		)
		mesh.RegisterA2AMeshServiceServer(grpcServer, s)

		// Graceful shutdown on context cancellation
		go func() {
			<-ctx.Done()
			grpcServer.GracefulStop()
		}()

		if err := grpcServer.Serve(lis); err != nil {
			return fmt.Errorf("external server failed: %w", err)
		}
		return nil
	})

	return g.Wait()
}

// StreamTask handles the bidirectional streaming of tasks.
// It implements the logic for both Outbound (Local -> Remote) and Inbound (Remote -> Local) requests.
func (s *Server) StreamTask(stream mesh.A2AMeshService_StreamTaskServer) error {
	p, ok := peer.FromContext(stream.Context())
	if !ok {
		return fmt.Errorf("failed to get peer info")
	}

	// Check if connection has TLS info. If yes, it's from External Listener (mTLS).
	_, hasTLS := p.AuthInfo.(credentials.TLSInfo)

	if hasTLS {
		// Case B: Inbound Request (Remote Sidecar -> Local Agent)
		return s.handleInbound(stream)
	} else {
		// Case A: Outbound Request (Local Agent -> Remote Agent)
		return s.handleOutbound(stream)
	}
}

// handleOutbound handles requests from the Local Agent intended for a Remote Agent.
func (s *Server) handleOutbound(stream mesh.A2AMeshService_StreamTaskServer) error {
	ctx := stream.Context()
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return fmt.Errorf("missing metadata")
	}

	// 1. Discovery: Get TargetSkill from metadata
	targetSkills := md.Get("x-target-skill")
	if len(targetSkills) == 0 {
		return fmt.Errorf("missing x-target-skill metadata")
	}
	targetSkill := targetSkills[0]

	// 2. Registry Lookup
	// Note: ListAgentsRequest might need to be updated to support filtering by skill if not already.
	// Assuming ListAgents returns all and we filter, or it supports filtering.
	listResp, err := s.registryClient.ListAgents(ctx, &registry.ListAgentsRequest{}) // TODO: Add filter if available
	if err != nil {
		return fmt.Errorf("failed to list agents: %w", err)
	}

	var targetAgent *registry.AgentCard
	for _, agent := range listResp.Agents {
		// Simple matching logic: check if agent has the skill.
		// Assuming AgentCard has a list of skills or similar.
		// If not, we might just pick the first one for now as a placeholder.
		// The prompt says "Pick the first available agent".
		targetAgent = agent.AgentCard
		break
	}

	if targetAgent == nil {
		return fmt.Errorf("no agent found for skill: %s", targetSkill)
	}

	// 3. Dial Remote Sidecar (mTLS)
	// We need the remote sidecar's address. Assuming it's in the AgentCard.
	// AgentCard usually has `address` or `endpoints`.
	// Let's assume `targetAgent.Address` is the host:port.
	// Wait, AgentCard definition in registry.proto:
	// message AgentCard { ... repeated AgentInterface supported_interfaces = 9; ... }
	// message AgentInterface { string protocol_binding = 1; string url = 2; }
	// We need to find the interface with protocol_binding="grpc" (or similar).

	var remoteAddr string
	for _, iface := range targetAgent.SupportedInterfaces {
		if iface.ProtocolBinding == "grpc" { // Assuming "grpc" is the binding name
			remoteAddr = iface.Url
			break
		}
	}

	if remoteAddr == "" {
		// Fallback or error
		return fmt.Errorf("no grpc interface found for agent")
	}

	creds, err := loadTLSCredentials(s.config.CAFile, s.config.CertFile, s.config.KeyFile)
	if err != nil {
		return fmt.Errorf("failed to load TLS credentials for outbound: %w", err)
	}

	conn, err := grpc.Dial(remoteAddr, grpc.WithTransportCredentials(creds))
	if err != nil {
		return fmt.Errorf("failed to dial remote sidecar: %w", err)
	}
	defer conn.Close()

	remoteClient := mesh.NewA2AMeshServiceClient(conn)

	// 4. Forwarding: Create bidirectional stream
	// We need to forward metadata as well.
	outCtx := metadata.NewOutgoingContext(ctx, md)
	remoteStream, err := remoteClient.StreamTask(outCtx)
	if err != nil {
		return fmt.Errorf("failed to start remote stream: %w", err)
	}

	// Pipe streams
	errChan := make(chan error, 2)

	// Local -> Remote
	go func() {
		for {
			msg, err := stream.Recv()
			if err == io.EOF {
				remoteStream.CloseSend()
				return
			}
			if err != nil {
				errChan <- err
				return
			}
			if err := remoteStream.Send(msg); err != nil {
				errChan <- err
				return
			}
		}
	}()

	// Remote -> Local
	go func() {
		for {
			msg, err := remoteStream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				errChan <- err
				return
			}
			if err := stream.Send(msg); err != nil {
				errChan <- err
				return
			}
		}
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// handleInbound handles requests from a Remote Sidecar intended for the Local Agent.
func (s *Server) handleInbound(stream mesh.A2AMeshService_StreamTaskServer) error {
	// Forward to Local Agent running on AppPort.
	// Connect to Local Agent (Plaintext)
	conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", s.config.AppPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to local agent: %w", err)
	}
	defer conn.Close()

	localClient := mesh.NewA2AMeshServiceClient(conn)

	ctx := stream.Context()
	// Forward metadata? Yes.
	md, _ := metadata.FromIncomingContext(ctx)
	outCtx := metadata.NewOutgoingContext(ctx, md)

	localStream, err := localClient.StreamTask(outCtx)
	if err != nil {
		return fmt.Errorf("failed to start local stream: %w", err)
	}

	// Pipe streams
	errChan := make(chan error, 2)

	// Remote -> Local
	go func() {
		for {
			msg, err := stream.Recv()
			if err == io.EOF {
				localStream.CloseSend()
				return
			}
			if err != nil {
				errChan <- err
				return
			}
			if err := localStream.Send(msg); err != nil {
				errChan <- err
				return
			}
		}
	}()

	// Local -> Remote
	go func() {
		for {
			msg, err := localStream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				errChan <- err
				return
			}
			if err := stream.Send(msg); err != nil {
				errChan <- err
				return
			}
		}
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *Server) SendTask(ctx context.Context, req *mesh.TaskSendRequest) (*mesh.Task, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *Server) GetTask(ctx context.Context, req *mesh.GetTaskRequest) (*mesh.Task, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *Server) CancelTask(ctx context.Context, req *mesh.CancelTaskRequest) (*emptypb.Empty, error) {
	return nil, fmt.Errorf("not implemented")
}
